import { Injectable, Logger } from '@nestjs/common';
import { InjectConnection } from '@nestjs/mongoose';
import { Connection } from 'mongoose';
import { GridFSBucket, ObjectId } from 'mongodb';

/**
 * FileService handles upload, streaming, and deletion of files via MongoDB GridFS.
 */
@Injectable()
export class FileService {
  private readonly logger = new Logger(FileService.name);
  private bucket: GridFSBucket;

  constructor(@InjectConnection() private readonly connection: Connection) {
    this.bucket = new GridFSBucket(this.connection.db);
    this.logger.log('GridFS bucket initialized');
  }

  /**
   * Uploads a file stream to GridFS.
   * @param filename - Original file name
   * @param stream - Readable stream from the file
   * @returns ObjectId of the uploaded file
   */
  upload(filename: string, stream: NodeJS.ReadableStream): Promise<ObjectId> {
    this.logger.log(`Starting upload: ${filename}`);

    return new Promise<ObjectId>((resolve, reject) => {
      const uploadStream = this.bucket.openUploadStream(filename);

      stream
        .pipe(uploadStream)
        .on('error', (err) => {
          this.logger.error(`Upload failed: ${filename}`, err);
          reject(err);
        })
        .on('finish', () => {
          this.logger.log(
            `Upload finished: ${filename} (id: ${uploadStream.id})`
          );
          resolve(uploadStream.id);
        });
    });
  }

  /**
   * Streams a file from GridFS by its ID.
   * @param fileId - ObjectId as string
   * @returns Readable stream of the file
   */
  streamFile(fileId: string): NodeJS.ReadableStream {
    this.logger.log(`Streaming file with ID: ${fileId}`);
    return this.bucket.openDownloadStream(new ObjectId(fileId));
  }

  /**
   * Deletes a file from GridFS by its ID.
   * @param fileId - ObjectId as string
   */
  async deleteFile(fileId: string): Promise<void> {
    this.logger.log(`Deleting file with ID: ${fileId}`);
    await this.bucket.delete(new ObjectId(fileId));
    this.logger.log(`File deleted: ${fileId}`);
  }
}
