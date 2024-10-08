import { Controller, Post, Body } from '@nestjs/common';
import { KafkaService } from './kafka.service';
import { ApiOperation } from '@nestjs/swagger';

@Controller('kafka')
export class KafkaController {
  constructor(private readonly kafkaService: KafkaService) {}

  @ApiOperation({
    summary: 'Send inference message to Kafka (Do not implement)',
    description: 'Send a message to Kafka for model inference',
    tags: ['kafka'],
  })
  @Post('send/inference')
  async sendMessage(
    @Body()
    message: {
      jobId: number;
      experimentId: number;
      proteinData: number[];
    },
  ) {
    const kafkaMessage = {
      job_id: message.jobId,
      experiment_id: message.experimentId,
      protein_data: message.proteinData,
    };
    try {
      await this.kafkaService.sendMessage(
        'ModelInferenceRequest',
        kafkaMessage,
      );
    } catch (e) {
      return { sucess: false, status: 'Model inference request failed: ' + e };
    }
    return { sucess: true, status: 'Model inference request sent' };
  }

  @ApiOperation({
    summary: 'Send training message to Kafka (Do not implement)',
    description: 'Send a message to Kafka for model training',
    tags: ['kafka'],
  })
  @Post('send/train')
  async sendMessageTraining(
    @Body()
    message: {
      jobId: number;
      experimentId: number;
      proteinData: number[];
      targetValue: number;
    },
  ) {
    const kafkaMessage = {
      job_id: message.jobId,
      experiment_id: message.experimentId,
      protein_data: message.proteinData,
      target_value: message.targetValue,
    };
    try {
      await this.kafkaService.sendMessage('ModelTrainRequest', kafkaMessage);
    } catch (e) {
      return { sucess: false, status: 'Model training request failed: ' + e };
    }
    return { sucess: true, status: 'Model training request sent' };
  }
}
