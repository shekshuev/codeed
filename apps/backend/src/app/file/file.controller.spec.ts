import { Test, TestingModule } from '@nestjs/testing';
import { FileController } from './file.controller';
import { FileService } from './file.service';
import { FastifyReply, FastifyRequest } from 'fastify';
import { Readable, Writable } from 'stream';
import { ObjectId } from 'mongodb';

describe('FileController', () => {
  let controller: FileController;

  const mockFileService = {
    upload: jest.fn(),
    streamFile: jest.fn(),
    deleteFile: jest.fn(),
  };

  beforeEach(async () => {
    const module: TestingModule = await Test.createTestingModule({
      controllers: [FileController],
      providers: [
        {
          provide: FileService,
          useValue: mockFileService,
        },
      ],
    }).compile();

    controller = module.get<FileController>(FileController);
  });

  afterEach(() => {
    jest.clearAllMocks();
  });

  it('should upload a single file and return fileId', async () => {
    const fileId = new ObjectId();
    const fakePart = {
      file: Readable.from(['file content']),
      filename: 'test.txt',
    };

    const req = {
      parts: async function* () {
        yield fakePart;
      },
    } as unknown as FastifyRequest;

    mockFileService.upload.mockResolvedValue(fileId);

    const result = await controller.upload(req);
    expect(mockFileService.upload).toHaveBeenCalledWith(
      'test.txt',
      fakePart.file
    );
    expect(result).toEqual([{ filename: 'test.txt', fileId }]);
  });

  it('should return empty array if no file uploaded', async () => {
    const req = {
      // eslint-disable-next-line require-yield
      parts: async function* () {
        return;
      },
    } as unknown as FastifyRequest;

    const result = await controller.upload(req);
    expect(result).toEqual([]);
  });

  it('should stream file successfully', async () => {
    const mockStream = Readable.from(['file content']);
    mockFileService.streamFile.mockReturnValue(mockStream);

    const res = {
      header: jest.fn(),
      raw: new Writable({
        write(chunk, _encoding, callback) {
          callback();
        },
      }),
    } as unknown as FastifyReply;

    await controller.getFile('507f191e810c19729de860ea', res);
    expect(mockFileService.streamFile).toHaveBeenCalled();
    expect(res.header).toHaveBeenCalledWith(
      'Content-Type',
      'application/octet-stream'
    );
  });

  it('should return 404 if stream fails', async () => {
    mockFileService.streamFile.mockImplementation(() => {
      throw new Error('not found');
    });

    const res = {
      header: jest.fn(),
      raw: new Writable({
        write(_chunk, _encoding, callback) {
          callback();
        },
      }),
    } as unknown as FastifyReply;

    await expect(() => controller.getFile('bad_id', res)).rejects.toThrow(
      'File not found'
    );
  });

  it('should delete file successfully', async () => {
    const id = '507f191e810c19729de860ea';
    await controller.deleteFile(id);
    expect(mockFileService.deleteFile).toHaveBeenCalledWith(id);
  });

  it('should return 404 if delete fails', async () => {
    mockFileService.deleteFile.mockRejectedValue(new Error('not found'));
    await expect(() => controller.deleteFile('bad_id')).rejects.toThrow(
      'File not found'
    );
  });
});
