import { Module } from '@nestjs/common';
import { AppController } from './app.controller';
import { AppService } from './app.service';
import { ConfigModule } from '@nestjs/config';
import { DbModule } from './db/db.module';
import { FileModule } from './file/file.module';
import { KafkaModule } from './kafka/kafka.module';
import { JobModule } from './job/job.module';
import { ExperimentModule } from './experiment/experiment.module';

@Module({
  imports: [
    ConfigModule.forRoot({
      isGlobal: true,
    }),
    DbModule,
    // FileModule,
    KafkaModule,
    JobModule,
    ExperimentModule,
  ],
  controllers: [AppController],
  providers: [AppService],
})
export class AppModule {}
