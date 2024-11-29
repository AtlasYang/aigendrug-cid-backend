import { Module } from '@nestjs/common';
import { JobController } from './job.controller';
import { JobService } from './job.service';
import { DbModule } from 'src/db/db.module';
import { FileModule } from 'src/file/file.module';

@Module({
  imports: [DbModule, FileModule],
  controllers: [JobController],
  providers: [JobService],
})
export class JobModule {}
