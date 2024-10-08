import { Controller, Get } from '@nestjs/common';
import { AppService } from './app.service';
import { ApiOperation } from '@nestjs/swagger';

@Controller()
export class AppController {
  constructor(private readonly appService: AppService) {}

  @ApiOperation({
    summary: 'Hello Aigendrug!',
    description: 'Developer does not sleep at night',
  })
  @Get()
  getHello(): string {
    return this.appService.getHello();
  }
}
