package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type PyTorchExecutor struct {
	Address string
}

func (pte *PyTorchExecutor) InitializeModel(jobID int, initialLigands []InitialLigand) error {
	data := map[string]interface{}{
		"initial_ligands": initialLigands,
	}
	jsonData, _ := json.Marshal(data)

	resp, err := http.Post(fmt.Sprintf("http://%s/initialize/%d", pte.Address, jobID), "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to initialize model: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to initialize model, server responded with status: %v", resp.StatusCode)
	}

	return nil
}

func (pte *PyTorchExecutor) ProcessWithModel(trainDataPath string, testDataPath string) error {
	data := map[string]string{
		"train_csv_path": trainDataPath,
		"test_csv_path":  testDataPath,
	}
	jsonData, _ := json.Marshal(data)

	resp, err := http.Post(fmt.Sprintf("http://%s/process", pte.Address), "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to process data: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to process data, server responded with status: %v", resp.StatusCode)
	}

	return nil
}

func (pte *PyTorchExecutor) LoadModel(weightPath string) error {
	data := map[string]string{"weight_path": weightPath}
	jsonData, _ := json.Marshal(data)

	resp, err := http.Post(fmt.Sprintf("http://%s/load", pte.Address), "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to load model: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to load model, server responded with status: %v", resp.StatusCode)
	}

	return nil
}

func (pte *PyTorchExecutor) RunInference(proteinData string) (float64, error) {
	data := map[string]string{"protein_data": proteinData}
	jsonData, _ := json.Marshal(data)

	resp, err := http.Post(fmt.Sprintf("http://%s/inference", pte.Address), "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return 0.0, fmt.Errorf("inference request failed: %v", err)
	}
	defer resp.Body.Close()

	var result map[string]float64
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return 0.0, fmt.Errorf("failed to decode inference result: %v", err)
	}

	return result["result"], nil
}

func (pte *PyTorchExecutor) TrainModel(proteinData string, targetValue float64) error {
	data := map[string]interface{}{
		"protein_data": proteinData,
		"target_value": targetValue,
	}
	jsonData, _ := json.Marshal(data)

	resp, err := http.Post(fmt.Sprintf("http://%s/train", pte.Address), "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("training request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("training failed, server responded with status: %v", resp.StatusCode)
	}

	return nil
}
