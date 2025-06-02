import {
  Controller,
  Delete,
  Get,
  HttpCode,
  HttpException,
  HttpStatus,
  Logger,
  Param,
  Post,
  Req,
  Res,
} from '@nestjs/common';
import { FastifyReply, FastifyRequest } from 'fastify';
import { FileService } from './file.service';
import { MultipartFile } from '@fastify/multipart';
import {
  ApiBadRequestResponse,
  ApiBody,
  ApiConsumes,
  ApiCreatedResponse,
  ApiNoContentResponse,
  ApiNotFoundResponse,
  ApiOkResponse,
  ApiOperation,
  ApiParam,
  ApiTags,
} from '@nestjs/swagger';

@ApiTags('Files')
@Controller('files')
export class FileController {
  private readonly logger = new Logger(FileController.name);

  constructor(private fileService: FileService) {}

  @Post('upload')
  @HttpCode(HttpStatus.CREATED)
  @ApiOperation({ summary: 'Upload one or multiple files' })
  @ApiConsumes('multipart/form-data')
  @ApiBody({
    description: 'File to upload',
    schema: {
      type: 'object',
      properties: {
        file: {
          type: 'string',
          format: 'binary',
        },
      },
    },
  })
  @ApiCreatedResponse({ description: 'File(s) uploaded successfully' })
  @ApiBadRequestResponse({ description: 'Invalid file format or request' })
  async upload(@Req() req: FastifyRequest) {
    const parts = req.parts();
    const result = [];

    for await (const part of parts) {
      if ((part as MultipartFile).file) {
        const filePart = part as MultipartFile;
        this.logger.log(`Uploading file: ${filePart.filename}`);

        const fileId = await this.fileService.upload(
          filePart.filename,
          filePart.file
        );

        result.push({ filename: filePart.filename, fileId });
      }
    }

    this.logger.log(`Uploaded ${result.length} file(s)`);
    return result;
  }

  @Get(':id')
  @ApiOperation({ summary: 'Download a file by ID' })
  @ApiParam({ name: 'id', type: 'string', description: 'GridFS file ID' })
  @ApiOkResponse({ description: 'File streamed successfully' })
  @ApiNotFoundResponse({ description: 'File not found' })
  async getFile(@Param('id') id: string, @Res() res: FastifyReply) {
    this.logger.log(`Streaming file with ID: ${id}`);
    try {
      const stream = this.fileService.streamFile(id);
      res.header('Content-Type', 'application/octet-stream');
      stream.pipe(res.raw);
    } catch (error) {
      this.logger.error(`Failed to stream file: ${error.message}`);
      throw new HttpException('File not found', HttpStatus.NOT_FOUND);
    }
  }

  @Delete(':id')
  @HttpCode(HttpStatus.NO_CONTENT)
  @ApiOperation({ summary: 'Delete a file by ID' })
  @ApiParam({ name: 'id', type: 'string', description: 'GridFS file ID' })
  @ApiNoContentResponse({ description: 'File deleted successfully' })
  @ApiNotFoundResponse({ description: 'File not found' })
  async deleteFile(@Param('id') id: string) {
    this.logger.log(`Deleting file with ID: ${id}`);
    try {
      await this.fileService.deleteFile(id);
    } catch {
      this.logger.warn(`File not found: ${id}`);
      throw new HttpException('File not found', HttpStatus.NOT_FOUND);
    }
  }
}
