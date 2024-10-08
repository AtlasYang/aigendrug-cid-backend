import {
  Injectable,
  OnModuleInit,
  OnModuleDestroy,
  Inject,
  forwardRef,
} from '@nestjs/common';
import { Kafka, Consumer, Producer } from 'kafkajs';
import { ExperimentService } from 'src/experiment/experiment.service';

@Injectable()
export class KafkaService implements OnModuleInit, OnModuleDestroy {
  private kafka: Kafka;
  private producer: Producer;
  private consumer: Consumer;

  constructor(
    @Inject(forwardRef(() => ExperimentService))
    private experimentService: ExperimentService,
  ) {
    this.kafka = new Kafka({
      clientId: process.env.KAFKA_CLIENT_ID,
      brokers: [process.env.KAFKA_SERVER],
    });

    this.producer = this.kafka.producer();
    this.consumer = this.kafka.consumer({
      groupId: process.env.KAFKA_GROUP_ID,
    });
  }

  async onModuleInit() {
    await this.producer.connect();
    await this.consumer.connect();
    await this.consumer.subscribe({
      topic: 'ModelInferenceResponse',
      fromBeginning: true,
    });
    await this.consumer.subscribe({
      topic: 'ModelTrainResponse',
      fromBeginning: true,
    });

    await this.consumer.run({
      eachMessage: async ({ topic, partition, message }) => {
        console.log({
          topic,
          partition,
          offset: message.offset,
          value: message.value.toString(),
        });

        const messageValue = JSON.parse(message.value.toString());
        if (topic === 'ModelInferenceResponse') {
          console.log('Received ModelInferenceResponse: ', messageValue);
          await this.experimentService.updateExperimentPredictedValue(
            messageValue.experiment_id,
            messageValue.predicted_value,
          );
          await this.experimentService.updateExperimentTrainingStatus(
            messageValue.experiment_id,
            2,
          );
        } else if (topic === 'ModelTrainResponse') {
          console.log('Received ModelTrainResponse: ', messageValue);
          await this.experimentService.updateExperimentTrainingStatus(
            messageValue.experiment_id,
            2,
          );
        }
      },
    });

    console.log('Kafka Producer and Consumer are ready');
  }

  async sendMessage(topic: string, message: object) {
    await this.producer.send({
      topic,
      messages: [{ value: JSON.stringify(message) }],
    });

    console.log(`Message sent to ${topic}: `, message);
  }

  async onModuleDestroy() {
    await this.producer.disconnect();
    await this.consumer.disconnect();
  }
}
