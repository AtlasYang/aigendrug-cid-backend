import { Inject, Injectable } from '@nestjs/common';
import { Client } from 'pg';
import { PG_CONNECTION } from 'src/constants';
import { Job, JobCreateDto } from './job.interface';

@Injectable()
export class JobService {
  constructor(@Inject(PG_CONNECTION) private readonly pgConnection: Client) {}

  async getJobs() {
    const res = await this.pgConnection.query('SELECT * FROM job');
    return res.rows as Job[];
  }

  async getJobById(jobId: number) {
    const res = await this.pgConnection.query(
      'SELECT * FROM job WHERE id = $1',
      [jobId],
    );
    return res.rows[0] as Job;
  }

  async createJob(jobData: JobCreateDto) {
    const res = await this.pgConnection.query(
      'INSERT INTO job (name, target_protein_name) VALUES ($1, $2) RETURNING *',
      [jobData.name, jobData.target_protein_name],
    );

    return res.rows[0] as Job;
  }

  async deleteJob(jobId: number) {
    await this.pgConnection.query('DELETE FROM job WHERE id = $1', [jobId]);
  }
}
