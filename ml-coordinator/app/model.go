package app

const (
	TopicModelInferenceRequest = "ModelInferenceRequest"
	TopicModelTrainRequest     = "ModelTrainRequest"
)

type ModelInferenceRequest struct {
	JobID        int    `json:"job_id"`
	ExperimentID int    `json:"experiment_id"`
	ProteinData  string `json:"protein_data"`
}

type ModelTrainRequest struct {
	JobID        int     `json:"job_id"`
	ExperimentID int     `json:"experiment_id"`
	ProteinData  string  `json:"protein_data"`
	TargetValue  float64 `json:"target_value"`
}
