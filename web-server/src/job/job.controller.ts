import { Body, Controller, Delete, Get, Param, Post } from '@nestjs/common';
import { JobService } from './job.service';
import { ApiBody, ApiOperation, ApiResponse } from '@nestjs/swagger';
import { JobCreateDto, JobCreateSchema, JobSchema } from './job.interface';

@Controller('job')
export class JobController {
  constructor(private readonly jobService: JobService) {}

  @ApiOperation({
    summary: 'Get all jobs',
    description: 'Returns a list of all jobs',
    tags: ['job'],
  })
  @ApiResponse({
    status: 200,
    description: 'List of all jobs',
    schema: { type: 'array', items: JobSchema },
  })
  @Get()
  async getJobs() {
    return this.jobService.getJobs();
  }

  @ApiOperation({
    summary: 'Get job by ID',
    description: 'Returns a job by its ID',
    tags: ['job'],
  })
  @ApiResponse({
    status: 200,
    description: 'Job by ID',
    schema: JobSchema,
  })
  @Get(':jobId')
  async getJobById(@Param('jobId') jobId: number) {
    return this.jobService.getJobById(jobId);
  }

  @ApiOperation({
    summary: 'Create job',
    description: 'Creates a new job',
    tags: ['job'],
  })
  @ApiBody({ schema: JobCreateSchema })
  @ApiResponse({
    status: 201,
    description: 'Job created',
    schema: JobSchema,
  })
  @Post()
  async createJob(@Body() jobData: JobCreateDto) {
    return this.jobService.createJob(jobData);
  }

  @ApiOperation({
    summary: 'Delete job by ID',
    description: 'Deletes a job by its ID',
    tags: ['job'],
  })
  @ApiResponse({
    status: 200,
    description: 'Job deleted',
  })
  @Delete(':jobId')
  async deleteJob(@Param('jobId') jobId: number) {
    return this.jobService.deleteJob(jobId);
  }
}
