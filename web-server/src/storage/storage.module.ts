import { Module } from '@nestjs/common';
import * as Minio from 'minio';
import { MINIO_CONNECTION } from 'src/constants';

const storageProvider = {
  provide: MINIO_CONNECTION,
  useValue: new Minio.Client({
    endPoint: process.env.STORAGE_HOST,
    port: 9000,
    useSSL: false,
    accessKey: process.env.MINIO_ACCESS_KEY_ID,
    secretKey: process.env.MINIO_SECRET_ACCESS_KEY,
  }),
};

@Module({
  providers: [storageProvider],
  exports: [storageProvider],
})
export class StorageModule {}
