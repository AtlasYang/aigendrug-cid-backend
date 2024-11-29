import { Inject, Injectable } from '@nestjs/common';
import { Client } from 'pg';
import { PG_CONNECTION } from 'src/constants';
import { Job, JobCreateDto, JobInitialLigands } from './job.interface';
import { Readable } from 'stream';
import * as csv from 'csv-parser';

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

  async parseJobInitialLigandsFile(
    file: Express.Multer.File,
  ): Promise<JobInitialLigands[]> {
    return new Promise((resolve, reject) => {
      const initialLigands = [];
      const stream = new Readable();
      stream._read = () => {};
      stream.push(file.buffer);
      stream.push(null);

      const csvParser = csv();

      stream
        .pipe(csvParser)
        .on('data', (row) => {
          initialLigands.push(row);
        })
        .on('end', () => {
          resolve(initialLigands);
        })
        .on('error', (err) => {
          reject(err);
        });
    });
  }
}
