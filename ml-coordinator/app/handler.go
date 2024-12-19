package app

import (
	"fmt"
	"time"
)

func HandleProcessRequest(req ModelProcessRequest, manager *TorchWorkerManager) error {
	job := &JobRequest{
		JobID:   req.JobID,
		ErrChan: make(chan error, 1),
	}

	manager.EnqueueJob(job)

	select {
	case err := <-job.ErrChan:
		return err
	case <-time.After(time.Second * 180):
		return fmt.Errorf("timeout waiting for available worker")
	}
}

// func HandleInitializeRequest(req ModelInitializeRequest, manager *TorchWorkerManager) error {
// 	agReq := &JobRequest{
// 		JobID:          req.JobID,
// 		RequestType:    "initialize",
// 		InitialLigands: req.InitialLigands,
// 		ErrChan:        make(chan error, 1),
// 	}

// 	manager.EnqueueJob(agReq)

// 	select {
// 	case err := <-agReq.ErrChan:
// 		return err
// 	case <-time.After(time.Second * 30):
// 		return fmt.Errorf("timeout waiting for available worker")
// 	}
// }

// func HandleInferenceRequest(req ModelInferenceRequest, manager *TorchWorkerManager) (float64, error) {
// 	job := &JobRequest{
// 		JobID:       req.JobID,
// 		RequestType: "inference",
// 		ProteinData: req.ProteinData,
// 		ResultChan:  make(chan float64, 1),
// 		ErrChan:     make(chan error, 1),
// 	}

// 	fmt.Println("Enqueueing job for inference with jobID:", req.JobID)
// 	manager.EnqueueJob(job)
// 	fmt.Println("Job enqueued for inference with jobID:", req.JobID)

// 	select {
// 	case result := <-job.ResultChan:
// 		return result, nil
// 	case err := <-job.ErrChan:
// 		return 0.0, err
// 	case <-time.After(time.Second * 5):
// 		return 0.0, fmt.Errorf("timeout waiting for available worker")
// 	}
// }

// func HandleTrainRequest(req ModelTrainRequest, manager *TorchWorkerManager) error {
// 	job := &JobRequest{
// 		JobID:       req.JobID,
// 		RequestType: "train",
// 		ProteinData: req.ProteinData,
// 		TargetValue: req.TargetValue,
// 		ErrChan:     make(chan error),
// 	}

// 	manager.EnqueueJob(job)

// 	select {
// 	case err := <-job.ErrChan:
// 		return err
// 	case <-time.After(time.Second * 30):
// 		return fmt.Errorf("timeout waiting for available worker")
// 	}
// }
