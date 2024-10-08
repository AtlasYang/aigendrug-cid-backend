import { forwardRef, Module } from '@nestjs/common';
import { ExperimentController } from './experiment.controller';
import { ExperimentService } from './experiment.service';
import { DbModule } from 'src/db/db.module';
import { KafkaModule } from 'src/kafka/kafka.module';

@Module({
  imports: [DbModule, forwardRef(() => KafkaModule)],
  controllers: [ExperimentController],
  providers: [ExperimentService],
  exports: [ExperimentService],
})
export class ExperimentModule {}
