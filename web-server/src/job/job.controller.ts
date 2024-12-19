import {
  Body,
  Controller,
  Delete,
  Get,
  Param,
  Post,
  Res,
  UploadedFile,
  UseInterceptors,
} from '@nestjs/common';
import { JobService } from './job.service';
import {
  ApiBody,
  ApiConsumes,
  ApiOperation,
  ApiResponse,
} from '@nestjs/swagger';
import { JobCreateDto, JobCreateSchema, JobSchema } from './job.interface';
import { FileInterceptor } from '@nestjs/platform-express';
import { FileService } from 'src/file/file.service';
import { KafkaService } from 'src/kafka/kafka.service';
import { ExperimentService } from 'src/experiment/experiment.service';

@Controller('job')
export class JobController {
  constructor(
    private readonly jobService: JobService,
    private readonly experimentService: ExperimentService,
    private readonly fileService: FileService,
    private readonly kafkaService: KafkaService,
  ) {}

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
    summary: 'Upload initial ligand file',
    description: 'Uploads the initial ligand file for a job',
    tags: ['job'],
  })
  @ApiConsumes('multipart/form-data')
  @ApiBody({
    schema: {
      type: 'object',
      properties: {
        file: { type: 'string', format: 'binary' },
      },
      required: ['file'],
    },
  })
  @ApiResponse({
    status: 200,
    description: 'Initial ligand uploaded',
  })
  @ApiResponse({
    status: 400,
    description: 'Invalid file',
  })
  @UseInterceptors(
    FileInterceptor('file', {
      limits: {
        fileSize: 1024 * 1024 * 1024 * 10,
      },
    }),
  )
  @Post(':jobId/upload-initial-ligand')
  async uploadInitialLigand(
    @Param('jobId') jobId: number,
    @UploadedFile() file: Express.Multer.File,
    @Res() res: any,
  ) {
    const initialLigands =
      await this.jobService.parseJobInitialLigandsFile(file);

    console.log(initialLigands);

    if (!initialLigands) {
      return res.status(400).send('Invalid file');
    }

    await this.fileService.uploadJobInitialLigandsFile(jobId, file);

    const kafkaMessage = {
      jobId,
      initialLigands,
    };

    await this.kafkaService.sendMessage('ModelInitializeRequest', kafkaMessage);

    // Save initial ligands to database
    initialLigands.forEach(async (ligand) => {
      await this.experimentService.createExperiment({
        type: 0,
        name: ligand.name,
        ligand_smiles: ligand.smiles,
        measured_value: ligand.std_value,
        job_id: jobId,
      });
    });

    // return this.jobService.uploadInitialLigand(jobId, initialLigands);
    return res.status(200).send('Initial ligand uploaded');
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
