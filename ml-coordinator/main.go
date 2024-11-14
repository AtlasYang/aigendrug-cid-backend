package main

import (
	"aigendrug-cid/ml-coordinator/app"
	"aigendrug-cid/ml-coordinator/storage"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	ctx := context.Background()

	minioClient := storage.NewMinioClient()
	workerCount, err := strconv.Atoi(os.Getenv("WORKER_COUNT"))
	if err != nil {
		log.Fatalf("Invalid WORKER_COUNT: %v", err)
	}
	torchWorkerCount, err := strconv.Atoi(os.Getenv("TORCH_WORKER_COUNT"))
	if err != nil {
		log.Fatalf("Invalid TORCH_WORKER_COUNT: %v", err)
	}
	weightCacheSize, err := strconv.Atoi(os.Getenv("WEIGHT_CACHE_SIZE"))
	if err != nil {
		log.Fatalf("Invalid WEIGHT_CACHE_SIZE: %v", err)
	}

	workerManager := app.NewTorchWorkerManager(torchWorkerCount)
	weightCacheLayer := app.NewWeightCacheLayer(&ctx, minioClient, weightCacheSize)

	addresses := []string{"torch-worker1:5000", "torch-worker2:5000", "torch-worker3:5000"}

	for i := 0; i < torchWorkerCount; i++ {
		address := addresses[i]
		worker := app.NewTorchWorker(i, address)
		worker.WeightCacheLayer = weightCacheLayer
		worker.PyTorchExecutor = &app.PyTorchExecutor{Address: address}
		workerManager.AddWorker(worker)

		go worker.Run(workerManager.JobQueue)
	}

	kafkaBroker := os.Getenv("KAFKA_BROKER_HOST")
	consumerGroup := os.Getenv("KAFKA_CONSUMER_GROUP")
	kafkaTopics := strings.Split(os.Getenv("KAFKA_TOPICS"), ",")
	if err != nil {
		log.Fatalf("Invalid WORKER_COUNT: %v", err)
	}

	kafkaConsumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": kafkaBroker,
		"group.id":          consumerGroup,
		"auto.offset.reset": "earliest",
	})
	if err != nil {
		log.Fatalf("Failed to create Kafka consumer: %v", err)
	}
	defer kafkaConsumer.Close()

	kafkaConsumer.SubscribeTopics(kafkaTopics, nil)

	msgChan := make(chan *kafka.Message, workerCount)
	var wg sync.WaitGroup

	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go worker(msgChan, workerManager, &wg)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		for {
			select {
			case <-ctx.Done():
				close(msgChan)
				return
			default:
				msg, err := kafkaConsumer.ReadMessage(time.Second)
				if err == nil {
					msgChan <- msg
				} else if !err.(kafka.Error).IsTimeout() {
					log.Printf("Consumer error: %v (%v)\n", err, msg)
				}
			}
		}
	}()

	router.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"PUT", "POST"},
	}))

	server := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	router.GET("/weight-cache-status", func(c *gin.Context) {
		c.JSON(http.StatusOK, weightCacheLayer.Weights)
	})

	router.GET("/torch-worker-status", func(c *gin.Context) {
		c.JSON(http.StatusOK, workerManager.Workers)
	})

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server ListenAndServe: %v", err)
		}
	}()

	<-ctx.Done()
	stop()
	log.Println("Shutting down gracefully...")

	ctxShutDown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctxShutDown); err != nil {
		log.Fatalf("HTTP server Shutdown: %v", err)
	}

	wg.Wait()
	log.Println("Server exited")
}

func worker(msgChan chan *kafka.Message, workerManager *app.TorchWorkerManager, wg *sync.WaitGroup) {
	defer wg.Done()
	for msg := range msgChan {
		if *msg.TopicPartition.Topic == app.TopicModelInferenceRequest {
			var req app.ModelInferenceRequest
			err := json.Unmarshal(msg.Value, &req)
			if err != nil {
				log.Printf("Failed to unmarshal ModelInferenceRequest: %v", err)
				continue
			}

			fmt.Println("Inference request received with jobID:", req.JobID)

			res, err := app.HandleInferenceRequest(req, workerManager)
			if err != nil {
				log.Printf("Failed to handle inference request: %v", err)
			}

			log.Printf("Inference request completed for job %d, result: %f", req.JobID, res)
		} else if *msg.TopicPartition.Topic == app.TopicModelTrainRequest {
			var req app.ModelTrainRequest
			err := json.Unmarshal(msg.Value, &req)
			if err != nil {
				log.Printf("Failed to unmarshal ModelTrainRequest: %v", err)
				continue
			}

			fmt.Println("Train request received with jobID:", req.JobID)

			err = app.HandleTrainRequest(req, workerManager)
			if err != nil {
				log.Printf("Failed to handle train request: %v", err)
			}

			log.Printf("Train request completed for job %d", req.JobID)
		} else {
			log.Printf("Unknown topic: %v", msg.TopicPartition.Topic)
		}
	}
}
