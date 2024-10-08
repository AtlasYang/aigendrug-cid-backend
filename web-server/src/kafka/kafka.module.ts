import { forwardRef, Module } from '@nestjs/common';
import { KafkaService } from './kafka.service';
import { KafkaController } from './kafka.controller';
import { ExperimentModule } from 'src/experiment/experiment.module';

@Module({
  imports: [forwardRef(() => ExperimentModule)],
  providers: [KafkaService],
  controllers: [KafkaController],
  exports: [KafkaService],
})
export class KafkaModule {}
