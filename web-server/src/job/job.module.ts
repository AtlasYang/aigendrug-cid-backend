import { Module } from '@nestjs/common';
import { JobController } from './job.controller';
import { JobService } from './job.service';
import { DbModule } from 'src/db/db.module';
import { FileModule } from 'src/file/file.module';
import { KafkaModule } from 'src/kafka/kafka.module';
import { ExperimentModule } from 'src/experiment/experiment.module';

@Module({
  imports: [DbModule, FileModule, KafkaModule, ExperimentModule],
  controllers: [JobController],
  providers: [JobService],
})
export class JobModule {}
