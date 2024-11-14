package app

import (
	"aigendrug-cid/ml-coordinator/storage"
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/minio/minio-go/v7"
)

type LoadedWeight struct {
	JobID  int
	Path   string
	Synced bool
}

func (lw *LoadedWeight) String() string {
	return fmt.Sprintf("JobID: %d, Path: %s, Synced: %t", lw.JobID, lw.Path, lw.Synced)
}

type WeightProvider struct {
	ctx           *context.Context
	storageClient *minio.Client
}

type WeightCacheLayer struct {
	Weights  map[int]*LoadedWeight
	Provider WeightProvider
	MaxSize  int
	mu       sync.RWMutex
	order    []int
}

func NewWeightCacheLayer(ctx *context.Context, storageClient *minio.Client, maxSize int) *WeightCacheLayer {
	return &WeightCacheLayer{
		Weights:  make(map[int]*LoadedWeight),
		Provider: WeightProvider{ctx: ctx, storageClient: storageClient},
		MaxSize:  maxSize,
		order:    make([]int, 0, maxSize),
	}
}

func (wp *WeightProvider) GetWeight(jobID int) (*LoadedWeight, error) {
	WeightBucket := os.Getenv("WEIGHT_BUCKET")

	storageService := storage.NewMinioService(wp.ctx, wp.storageClient)
	weightKey := WeightKey(jobID)

	weightBytes, err := storageService.GetObject(WeightBucket, weightKey)
	if err != nil {
		fmt.Printf("Failed to get object for job %d: %s\n", jobID, minio.ToErrorResponse(err).Code)
		if minio.ToErrorResponse(err).Code == "NoSuchKey" {
			defaultWeightKey := "default.pth"
			weightBytes, err = storageService.GetObject(WeightBucket, defaultWeightKey)
			if err != nil {
				return nil, err
			}

			err = storageService.CopyObject(WeightBucket, defaultWeightKey, WeightBucket, weightKey)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	weightPath := fmt.Sprintf("./weights/%s", weightKey)
	err = os.WriteFile(weightPath, weightBytes, 0644)
	if err != nil {
		return nil, err
	}

	return &LoadedWeight{JobID: jobID, Path: weightPath, Synced: true}, nil
}

func (wcl *WeightCacheLayer) GetWeight(jobID int) (*LoadedWeight, error) {
	wcl.mu.RLock()
	if weight, ok := wcl.Weights[jobID]; ok {
		wcl.updateLRU(jobID)
		wcl.mu.RUnlock()
		return weight, nil
	}
	wcl.mu.RUnlock()

	weight, err := wcl.Provider.GetWeight(jobID)
	if err != nil {
		return nil, err
	}

	wcl.mu.Lock()
	defer wcl.mu.Unlock()

	if len(wcl.Weights) >= wcl.MaxSize {
		wcl.evictLRU()
	}

	wcl.Weights[jobID] = weight
	wcl.order = append(wcl.order, jobID)

	return weight, nil
}

func (wcl *WeightCacheLayer) AddWeight(jobID int, weight *LoadedWeight) error {
	wcl.mu.Lock()
	defer wcl.mu.Unlock()

	if len(wcl.Weights) >= wcl.MaxSize {
		wcl.evictLRU()
	}

	wcl.Weights[jobID] = weight
	wcl.order = append(wcl.order, jobID)

	return nil
}

func (wcl *WeightCacheLayer) UploadWeight(jobID int, weightPath string) error {
	WeightBucket := os.Getenv("WEIGHT_BUCKET")
	storageService := storage.NewMinioService(wcl.Provider.ctx, wcl.Provider.storageClient)
	weightKey := WeightKey(jobID)

	weightBytes, err := os.ReadFile(weightPath)
	if err != nil {
		return err
	}

	err = storageService.UpdateObject(WeightBucket, weightKey, weightBytes)
	if err != nil {
		return err
	}

	return nil
}

func (wcl *WeightCacheLayer) UpdateWeight(jobID int, weight *LoadedWeight) error {
	wcl.mu.Lock()
	defer wcl.mu.Unlock()
	wcl.Weights[jobID] = weight
	wcl.updateLRU(jobID)
	return nil
}

func (wcl *WeightCacheLayer) SyncWeight(jobID int) error {
	weight, err := wcl.Provider.GetWeight(jobID)
	if err != nil {
		return err
	}

	wcl.mu.Lock()
	defer wcl.mu.Unlock()
	wcl.Weights[jobID] = weight
	wcl.updateLRU(jobID)

	return nil
}

func (wcl *WeightCacheLayer) RemoveWeight(jobID int) error {
	wcl.mu.Lock()
	defer wcl.mu.Unlock()
	delete(wcl.Weights, jobID)
	wcl.removeFromOrder(jobID)
	return nil
}

func (wcl *WeightCacheLayer) RemoveAllWeights() error {
	wcl.mu.Lock()
	defer wcl.mu.Unlock()
	wcl.Weights = make(map[int]*LoadedWeight)
	wcl.order = make([]int, 0, wcl.MaxSize)
	return nil
}

func (wcl *WeightCacheLayer) evictLRU() {
	if len(wcl.order) == 0 {
		return
	}

	oldestJobID := wcl.order[0]
	wcl.order = wcl.order[1:]
	delete(wcl.Weights, oldestJobID)
}

func (wcl *WeightCacheLayer) updateLRU(jobID int) {
	wcl.removeFromOrder(jobID)
	wcl.order = append(wcl.order, jobID)
}

func (wcl *WeightCacheLayer) removeFromOrder(jobID int) {
	for i, id := range wcl.order {
		if id == jobID {
			wcl.order = append(wcl.order[:i], wcl.order[i+1:]...)
			break
		}
	}
}
