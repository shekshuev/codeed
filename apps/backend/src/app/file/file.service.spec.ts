/* eslint-disable @typescript-eslint/no-explicit-any */
import { Test, TestingModule } from '@nestjs/testing';
import { FileService } from './file.service';
import { getConnectionToken } from '@nestjs/mongoose';
import { GridFSBucket, ObjectId } from 'mongodb';
import { Readable, Writable } from 'stream';

describe('FileService', () => {
  let service: FileService;
  let bucketMock: Partial<GridFSBucket>;
  let mockConnection: any;

  beforeEach(async () => {
    const uploadStreamMock = Object.assign(
      new Writable({
        write(chunk, encoding, callback) {
          callback();
        },
      }),
      {
        id: new ObjectId(),
        on: jest.fn().mockImplementation(function (event, cb) {
          if (event === 'finish') cb();
          return this;
        }),
      }
    );

    const deleteMock = jest.fn().mockResolvedValue(undefined);
    const openUploadStream = jest.fn().mockReturnValue(uploadStreamMock);
    const openDownloadStream = jest
      .fn()
      .mockReturnValue(Readable.from(['test file content']));

    bucketMock = {
      openUploadStream,
      openDownloadStream,
      delete: deleteMock,
    };

    mockConnection = {
      db: {
        collection: jest.fn(),
      },
    };

    const module: TestingModule = await Test.createTestingModule({
      providers: [
        FileService,
        {
          provide: getConnectionToken(),
          useValue: mockConnection,
        },
      ],
    }).compile();

    service = module.get<FileService>(FileService);
    (service as any).bucket = bucketMock;
  });

  it('should upload file and return id', async () => {
    const fakeStream = Readable.from(['some content']);
    const fileId = await service.upload('test.txt', fakeStream);

    expect(bucketMock.openUploadStream).toHaveBeenCalledWith('test.txt');
    expect(fileId).toBeInstanceOf(ObjectId);
  });

  it('should stream file', () => {
    const stream = service.streamFile('507f191e810c19729de860ea');

    expect(bucketMock.openDownloadStream).toHaveBeenCalledWith(
      new ObjectId('507f191e810c19729de860ea')
    );
    expect(stream.readable).toBe(true);
  });

  it('should delete file', async () => {
    const id = '507f191e810c19729de860ea';
    await service.deleteFile(id);

    expect(bucketMock.delete).toHaveBeenCalledWith(new ObjectId(id));
  });
});
