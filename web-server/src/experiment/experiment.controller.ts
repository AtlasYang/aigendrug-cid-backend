import {
  Body,
  Controller,
  Delete,
  forwardRef,
  Get,
  Inject,
  Param,
  Post,
} from '@nestjs/common';
import { ExperimentService } from './experiment.service';
import { KafkaService } from 'src/kafka/kafka.service';
import { ApiBody, ApiOperation, ApiResponse } from '@nestjs/swagger';
import {
  ExperimentBatchCreateDto,
  ExperimentBatchCreateSchema,
  ExperimentCreateDto,
  ExperimentCreateSchema,
  ExperimentSchema,
} from './experiment.interface';

@Controller('experiment')
export class ExperimentController {
  constructor(
    private readonly experimentService: ExperimentService,
    @Inject(forwardRef(() => KafkaService)) private kafkaService: KafkaService,
  ) {}

  @ApiOperation({
    summary: 'Get all experiments',
    description: 'Returns a list of all experiments',
    tags: ['experiment'],
  })
  @ApiResponse({
    status: 200,
    description: 'List of all experiments',
    schema: { type: 'array', items: ExperimentSchema },
  })
  @Get()
  async getExperiments() {
    return this.experimentService.getExperiments();
  }

  @ApiOperation({
    summary: 'Get experiments by job ID',
    description: 'Returns a list of experiments by job ID',
    tags: ['experiment'],
  })
  @ApiResponse({
    status: 200,
    description: 'List of experiments by job ID',
    schema: { type: 'array', items: ExperimentSchema },
  })
  @Get('job/:jobId')
  async getExperimentsByJobId(@Param('jobId') jobId: number) {
    return this.experimentService.getExperimentsByJobId(jobId);
  }

  @ApiOperation({
    summary: 'Get experiment by ID',
    description: 'Returns an experiment by its ID',
    tags: ['experiment'],
  })
  @ApiResponse({
    status: 200,
    description: 'Experiment by ID',
    schema: ExperimentSchema,
  })
  @Get(':experimentId')
  async getExperimentById(@Param('experimentId') experimentId: number) {
    return this.experimentService.getExperimentById(experimentId);
  }

  @ApiOperation({
    summary: 'Create experiment',
    description: 'Creates a new experiment and sends a message to Kafka',
    tags: ['experiment'],
  })
  @ApiBody({ schema: ExperimentCreateSchema })
  @ApiResponse({
    status: 201,
    description: 'Experiment created',
    schema: {
      type: 'object',
      properties: {
        success: { type: 'boolean' },
        status: { type: 'string' },
      },
    },
  })
  @Post()
  async createExperiment(@Body() experimentData: ExperimentCreateDto) {
    const experiment =
      await this.experimentService.createExperiment(experimentData);

    const kafkaMessage = {
      job_id: experimentData.job_id,
      experiment_id: experiment.id,
      protein_data: experimentData.ligand_smiles,
      target_value: experimentData.measured_value,
    };

    if (experimentData.type === 0) {
      try {
        await this.kafkaService.sendMessage('ModelTrainRequest', kafkaMessage);
      } catch (e) {
        await this.experimentService.updateExperimentTrainingStatus(
          experiment.id,
          3,
        );
        return {
          success: false,
          status: 'Model training request failed: ' + e,
        };
      }
    } else if (experimentData.type === 1) {
      try {
        await this.kafkaService.sendMessage(
          'ModelInferenceRequest',
          kafkaMessage,
        );
      } catch (e) {
        await this.experimentService.updateExperimentTrainingStatus(
          experiment.id,
          3,
        );
        return {
          success: false,
          status: 'Model inference request failed: ' + e,
        };
      }
    } else {
      await this.experimentService.updateExperimentTrainingStatus(
        experiment.id,
        3,
      );
      return {
        success: false,
        status: 'Invalid experiment type',
      };
    }

    await this.experimentService.updateExperimentTrainingStatus(
      experiment.id,
      1,
    );

    return { success: true, status: 'Experiment created' };
  }

  @ApiOperation({
    summary: 'Create experiments in batch',
    description: 'Creates multiple experiments in batch',
    tags: ['experiment'],
  })
  @ApiBody({ schema: ExperimentBatchCreateSchema })
  @ApiResponse({
    status: 201,
    description: 'Experiments created',
    schema: {
      type: 'object',
      properties: {
        success: { type: 'boolean' },
        status: { type: 'string' },
      },
    },
  })
  @Post('batch')
  async createExperimentsBatch(
    @Body() experimentData: ExperimentBatchCreateDto,
  ) {}

  @ApiOperation({
    summary: 'Delete experiment by ID',
    description: 'Deletes an experiment by its ID',
    tags: ['experiment'],
  })
  @ApiResponse({
    status: 200,
    description: 'Experiment deleted',
  })
  @Delete(':experimentId')
  async deleteExperiment(@Param('experimentId') experimentId: number) {
    return this.experimentService.deleteExperiment(experimentId);
  }
}
