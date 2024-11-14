package app

import (
	"fmt"
	"sync"
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
	WeightCacheLayer *WeightCacheLayer
	PyTorchExecutor  *PyTorchExecutor
}

type TorchWorkerManager struct {
	Workers  []*TorchWorker
	MaxSize  int
	JobQueue chan *JobRequest
	mu       sync.Mutex
}

type JobRequest struct {
	JobID       int
	RequestType string // "inference" or "train"
	ProteinData string
	TargetValue float64
	ResultChan  chan float64
	ErrChan     chan error
}

func NewTorchWorker(id int, address string) *TorchWorker {
	return &TorchWorker{
		ID:              id,
		Status:          WorkerStatusIdle,
		LoadedJobID:     -1,
		PyTorchExecutor: &PyTorchExecutor{Address: address},
	}
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
		job := <-jobQueue

		if tw.Status == WorkerStatusIdle || tw.Status == WorkerStatusLoaded {
			tw.Status = WorkerStatusBusy

			switch job.RequestType {
			case "inference":
				result, err := tw.RunInference(job.JobID, job.ProteinData)

				tw.Status = WorkerStatusLoaded

				job.ResultChan <- result
				job.ErrChan <- err
			case "train":
				err := tw.TrainModel(job.JobID, job.ProteinData, job.TargetValue)
				tw.Status = WorkerStatusLoaded

				job.ErrChan <- err
			}
		} else {
			jobQueue <- job
		}
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
