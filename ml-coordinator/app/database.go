package app

import (
	"context"
	"fmt"
	"os"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPostgresClient(c *context.Context) *pgxpool.Pool {
	connHost := os.Getenv("POSTGRES_CONNECTION")
	dbpool, err := pgxpool.New(*c, connHost)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}

	return dbpool
}

type PostgresService struct {
	ctx            *context.Context
	PostgresClient *pgxpool.Pool
}

func NewPostgresService(ctx *context.Context) *PostgresService {
	return &PostgresService{
		ctx:            ctx,
		PostgresClient: NewPostgresClient(ctx),
	}
}

func (ps *PostgresService) GetAllExperimentsByJobID(jobID int) (*ExperimentData, error) {
	var testedligands []LigandData
	err := pgxscan.Select(*ps.ctx, ps.PostgresClient, &testedligands, "SELECT ligand_smiles as smiles, measured_value as std_value FROM experiment WHERE job_id = $1 AND type = 0", jobID)
	if err != nil {
		return nil, err
	}

	var untestedligands []LigandData
	err = pgxscan.Select(*ps.ctx, ps.PostgresClient, &untestedligands, "SELECT ligand_smiles as smiles FROM experiment WHERE job_id = $1 AND type = 1", jobID)
	if err != nil {
		return nil, err
	}

	return &ExperimentData{
		JobID:           jobID,
		TestedLigands:   testedligands,
		UntestedLigands: untestedligands,
	}, nil
}

func (ps *PostgresService) UpdateExperimentStatueByJobID(jobID int, status int) error {
	_, err := ps.PostgresClient.Exec(*ps.ctx, "UPDATE experiment SET training_status = $1 WHERE job_id = $2 AND type = 1", status, jobID)
	if err != nil {
		return err
	}

	return nil
}

func (ps *PostgresService) UpdatePredictedValueBySMILES(jobID int, smiles string, value float64) error {
	_, err := ps.PostgresClient.Exec(*ps.ctx, "UPDATE experiment SET predicted_value = $1 WHERE job_id = $2 AND ligand_smiles = $3", value, jobID, smiles)
	if err != nil {
		return err
	}

	return nil
}
