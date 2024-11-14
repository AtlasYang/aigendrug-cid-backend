import { Inject, Injectable } from '@nestjs/common';
import { Client } from 'pg';
import { PG_CONNECTION } from 'src/constants';
import { Experiment, ExperimentCreateDto } from './experiment.interface';

@Injectable()
export class ExperimentService {
  constructor(@Inject(PG_CONNECTION) private readonly pgConnection: Client) {}

  async getExperiments() {
    const res = await this.pgConnection.query('SELECT * FROM experiment');
    return res.rows.map((row: Experiment) => {
      return {
        ...row,
        predicted_value: Number(row.predicted_value),
        measured_value: Number(row.measured_value),
      };
    }) as Experiment[];
  }

  async getExperimentsByJobId(jobId: number) {
    const res = await this.pgConnection.query(
      'SELECT * FROM experiment WHERE job_id = $1',
      [jobId],
    );
    return res.rows.map((row: Experiment) => {
      return {
        ...row,
        predicted_value: Number(row.predicted_value),
        measured_value: Number(row.measured_value),
      };
    }) as Experiment[];
  }

  async getExperimentById(experimentId: number) {
    const res = await this.pgConnection.query(
      'SELECT * FROM experiment WHERE id = $1',
      [experimentId],
    );
    return {
      ...res.rows[0],
      predicted_value: Number(res.rows[0].predicted_value),
      measured_value: Number(res.rows[0].measured_value),
    } as Experiment;
  }

  async createExperiment(experimentData: ExperimentCreateDto) {
    const res = await this.pgConnection.query(
      'INSERT INTO experiment (type, name, ligand_smiles, measured_value, training_status, job_id) VALUES ($1, $2, $3, $4, $5, $6) RETURNING *',
      [
        experimentData.type,
        experimentData.name,
        experimentData.ligand_smiles,
        experimentData.measured_value,
        0,
        experimentData.job_id,
      ],
    );

    return res.rows[0] as Experiment;
  }

  async updateExperimentPredictedValue(
    experimentId: number,
    predictedValue: number,
  ) {
    const res = await this.pgConnection.query(
      'UPDATE experiment SET predicted_value = $1 WHERE id = $2 RETURNING *',
      [predictedValue, experimentId],
    );

    return res.rows[0] as Experiment;
  }

  async updateExperimentTrainingStatus(
    experimentId: number,
    trainingStatus: number,
  ) {
    const res = await this.pgConnection.query(
      'UPDATE experiment SET training_status = $1 WHERE id = $2 RETURNING *',
      [trainingStatus, experimentId],
    );

    return res.rows[0] as Experiment;
  }

  async deleteExperiment(experimentId: number) {
    await this.pgConnection.query('DELETE FROM experiment WHERE id = $1', [
      experimentId,
    ]);
  }
}
