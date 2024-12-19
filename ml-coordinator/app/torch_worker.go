package app

import (
	"fmt"
	"sync"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

const (
	WorkerStatusIdle   = "idle"
	WorkerStatusLoaded = "loaded"
	WorkerStatusBusy   = "busy"
)

type TorchWorker struct {
	ID               int
	Status           string
	LoadedJobID      int // JobID of the loaded model, -1 if no model is loaded
	DatabaseService  *PostgresService
	WeightCacheLayer *WeightCacheLayer
	PyTorchExecutor  *PyTorchExecutor
}

type TorchWorkerManager struct {
	Workers       []*TorchWorker
	MaxSize       int
	JobQueue      chan *JobRequest
	KafkaProducer *kafka.Producer
	mu            sync.Mutex
}

// type JobRequest struct {
// 	JobID          int
// 	RequestType    string // "inference", "train", "initialize"
// 	ProteinData    string
// 	TargetValue    float64
// 	InitialLigands []InitialLigand // Only used for model initialization
// 	ResultChan     chan float64
// 	ErrChan        chan error
// }

type JobRequest struct {
	JobID      int
	ResultChan chan float64
	ErrChan    chan error
}

func NewTorchWorker(id int, address string) *TorchWorker {
	return &TorchWorker{
		ID:              id,
		Status:          WorkerStatusIdle,
		LoadedJobID:     -1,
		PyTorchExecutor: &PyTorchExecutor{Address: address},
	}
}

func (tw *TorchWorker) Process(jobID int) error {
	fmt.Printf("Processing job %d\n", jobID)

	ed, err := tw.DatabaseService.GetAllExperimentsByJobID(jobID)
	if err != nil {
		return fmt.Errorf("failed to get experiment data: %v", err)
	}

	err = tw.DatabaseService.UpdateExperimentStatueByJobID(jobID, 1)
	if err != nil {
		return fmt.Errorf("failed to update experiment status: %v", err)
	}

	basePath := "/app/csv"
	trainCsvPath := fmt.Sprintf("%s/%d_train.csv", basePath, jobID)
	testCsvPath := fmt.Sprintf("%s/%d_test.csv", basePath, jobID)

	err = LigandDataToCsv(ed.TestedLigands, trainCsvPath)
	if err != nil {
		return fmt.Errorf("failed to write train CSV: %v", err)
	}
	err = LigandDataToCsv(ed.UntestedLigands, testCsvPath)
	if err != nil {
		return fmt.Errorf("failed to write test CSV: %v", err)
	}

	err = tw.PyTorchExecutor.ProcessWithModel(trainCsvPath, testCsvPath)
	if err != nil {
		return fmt.Errorf("failed to process with model: %v", err)
	}

	newLigandData, err := CsvToLigandData(testCsvPath)
	if err != nil {
		return fmt.Errorf("failed to read test CSV: %v", err)
	}

	fmt.Print("New ligand data:\n")
	fmt.Printf("SMILES, StdValue\n")

	for _, ligand := range newLigandData {
		fmt.Printf("SMILES: %s, StdValue: %f\n", ligand.SMILES, ligand.StdValue)
		err = tw.DatabaseService.UpdatePredictedValueBySMILES(jobID, ligand.SMILES, ligand.StdValue)
		if err != nil {
			return fmt.Errorf("failed to update predicted value: %v", err)
		}
	}

	err = tw.DatabaseService.UpdateExperimentStatueByJobID(jobID, 2)
	if err != nil {
		return fmt.Errorf("failed to update experiment status: %v", err)
	}

	return nil
}

func (tw *TorchWorker) LoadModel(jobID int) error {
	weight, err := tw.WeightCacheLayer.GetWeight(jobID)
	if err != nil {
		return err
	}

	err = tw.PyTorchExecutor.LoadModel(weight.Path)
	if err != nil {
		return err
	}

	tw.LoadedJobID = jobID

	return nil
}

func (tw *TorchWorker) InitializeModel(jobID int, initialLigands []InitialLigand) error {
	err := tw.PyTorchExecutor.InitializeModel(jobID, initialLigands)
	if err != nil {
		return fmt.Errorf("failed to initialize model: %v", err)
	}

	// Upload the model weights to the storage
	err = tw.WeightCacheLayer.UploadWeight(jobID)
	if err != nil {
		return fmt.Errorf("failed to upload model weights: %v", err)
	}

	// Load the model after initialization
	err = tw.LoadModel(jobID)
	if err != nil {
		return fmt.Errorf("failed to load model after initialization: %v", err)
	}

	return nil
}

func (tw *TorchWorker) RunInference(jobID int, proteinData string) (float64, error) {
	if tw.LoadedJobID == -1 {
		err := tw.LoadModel(jobID)
		if err != nil {
			return 0.0, fmt.Errorf("failed to load model for inference: %v", err)
		}
	}

	if tw.LoadedJobID != jobID {
		err := tw.LoadModel(jobID)
		if err != nil {
			return 0.0, fmt.Errorf("failed to load model for inference: %v", err)
		}
	}

	result, err := tw.PyTorchExecutor.RunInference(proteinData)
	if err != nil {
		return 0.0, fmt.Errorf("inference request failed: %v", err)
	}

	return result, nil
}

func (tw *TorchWorker) TrainModel(jobID int, proteinData string, targetValue float64) error {
	if tw.LoadedJobID == -1 {
		err := tw.LoadModel(jobID)
		if err != nil {
			return fmt.Errorf("failed to load model for training: %v", err)
		}
	}

	if tw.LoadedJobID != jobID {
		err := tw.LoadModel(jobID)
		if err != nil {
			return fmt.Errorf("failed to load model for training: %v", err)
		}
	}

	err := tw.PyTorchExecutor.TrainModel(proteinData, targetValue)
	if err != nil {
		return fmt.Errorf("training request failed: %v", err)
	}

	return nil
}

func (tw *TorchWorker) UnloadModel() {
	tw.LoadedJobID = -1
	tw.Status = WorkerStatusIdle
}

func (tw *TorchWorker) Run(jobQueue chan *JobRequest) {
	for {
		fmt.Printf("Worker %d is waiting for job\n", tw.ID)

		job := <-jobQueue

		fmt.Printf("Worker %d received job %d\n", tw.ID, job.JobID)

		tw.Status = WorkerStatusBusy

		fmt.Printf("Worker %d is processing job %d\n", tw.ID, job.JobID)

		err := tw.Process(job.JobID)

		fmt.Printf("Worker %d finished processing job %d\n", tw.ID, job.JobID)

		tw.Status = WorkerStatusIdle

		job.ErrChan <- err

		// if tw.Status == WorkerStatusIdle || tw.Status == WorkerStatusLoaded {
		// 	tw.Status = WorkerStatusBusy

		// 	switch job.RequestType {
		// 	case "initialize":
		// 		err := tw.InitializeModel(job.JobID, job.InitialLigands)
		// 		tw.Status = WorkerStatusLoaded

		// 		job.ErrChan <- err

		// 	case "inference":
		// 		result, err := tw.RunInference(job.JobID, job.ProteinData)

		// 		tw.Status = WorkerStatusLoaded

		// 		job.ResultChan <- result
		// 		job.ErrChan <- err

		// 	case "train":
		// 		err := tw.TrainModel(job.JobID, job.ProteinData, job.TargetValue)
		// 		tw.Status = WorkerStatusLoaded

		// 		job.ErrChan <- err
		// 	}
		// } else {
		// 	jobQueue <- job
		// }
	}
}

func NewTorchWorkerManager(maxSize int) *TorchWorkerManager {
	return &TorchWorkerManager{
		Workers:  make([]*TorchWorker, 0),
		MaxSize:  maxSize,
		JobQueue: make(chan *JobRequest, 100),
		mu:       sync.Mutex{},
	}
}

func (twm *TorchWorkerManager) AddWorker(worker *TorchWorker) {
	twm.mu.Lock()
	defer twm.mu.Unlock()
	twm.Workers = append(twm.Workers, worker)
}

func (twm *TorchWorkerManager) EnqueueJob(job *JobRequest) {
	twm.mu.Lock()
	defer twm.mu.Unlock()

	twm.JobQueue <- job
}
