package app

var (
	TopicModelProcessRequest    = "ModelProcessRequest"
	TopicModelInitializeRequest = "ModelInitializeRequest"
	TopicModelInferenceRequest  = "ModelInferenceRequest"
	TopicModelTrainRequest      = "ModelTrainRequest"
	TopicModelInferenceResponse = "ModelInferenceResponse"
	TopicModelTrainResponse     = "ModelTrainResponse"
)

type InitialLigand struct {
	Name     string  `json:"name"`
	SMILES   string  `json:"smiles"`
	StdValue float64 `json:"std_value"`
}

type LigandData struct {
	SMILES   string  `json:"smiles"`
	StdValue float64 `json:"std_value"`
}

type ExperimentData struct {
	JobID           int          `json:"job_id"`
	TestedLigands   []LigandData `json:"tested_ligands"`
	UntestedLigands []LigandData `json:"untested_ligands"`
}

type ModelInitializeRequest struct {
	JobID          int             `json:"job_id"`
	InitialLigands []InitialLigand `json:"initial_ligands"`
}

type ModelProcessRequest struct {
	JobID int `json:"job_id"`
}

type ModelProcessResponse struct {
	JobID   int  `json:"job_id"`
	Success bool `json:"success"`
}

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

type ModelInferenceResponse struct {
	JobID        int     `json:"job_id"`
	ExperimentID int     `json:"experiment_id"`
	Result       float64 `json:"result"`
	Success      bool    `json:"success"`
}

type ModelTrainResponse struct {
	JobID        int  `json:"job_id"`
	ExperimentID int  `json:"experiment_id"`
	Success      bool `json:"success"`
}
