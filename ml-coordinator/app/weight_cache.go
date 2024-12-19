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

func (wp *WeightProvider) GetAGCheckpoint(jobID int) (*LoadedWeight, error) {
	AGBucket := os.Getenv("WEIGHT_BUCKET")

	storageService := storage.NewMinioService(wp.ctx, wp.storageClient)
	agKey := AGKey(jobID)
	agPath := AGPath(jobID)

	agBytes, err := storageService.GetObject(AGBucket, agKey)
	if err != nil {
		return nil, err
	}

	err = ExtractTar(agBytes, "/app/weights")
	if err != nil {
		return nil, err
	}

	return &LoadedWeight{JobID: jobID, Path: agPath, Synced: true}, nil
}

func (wcl *WeightCacheLayer) GetWeight(jobID int) (*LoadedWeight, error) {
	wcl.mu.RLock()
	if weight, ok := wcl.Weights[jobID]; ok {
		wcl.updateLRU(jobID)
		wcl.mu.RUnlock()
		return weight, nil
	}
	wcl.mu.RUnlock()

	weight, err := wcl.Provider.GetAGCheckpoint(jobID)
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

func (wcl *WeightCacheLayer) UploadWeight(jobID int) error {
	WeightBucket := os.Getenv("WEIGHT_BUCKET")
	storageService := storage.NewMinioService(wcl.Provider.ctx, wcl.Provider.storageClient)
	agPath := AGPath(jobID)

	agTarBytes, err := ArchiveTar(agPath)
	if err != nil {
		return err
	}

	err = storageService.PutObject(WeightBucket, AGKey(jobID), agTarBytes)
	if err != nil {
		return err
	}

	weight := &LoadedWeight{JobID: jobID, Path: agPath, Synced: true}
	wcl.AddWeight(jobID, weight)

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
	weight, err := wcl.Provider.GetAGCheckpoint(jobID)
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
