import { ApiProperty, PartialType } from '@nestjs/swagger';
import { IsEnum, IsOptional, IsString } from 'class-validator';
import {
  CreateAccountDto as CreateAccountDtoType,
  ReadAccountDto as ReadAccountDtoType,
  Role,
  Status,
  UpdateAccountDto as UpdateAccountDtoType,
} from '@codeed/types';

export class CreateAccountDto implements CreateAccountDtoType {
  @ApiProperty({
    example: 'John',
    description: 'First name of the user',
  })
  @IsString()
  firstName: string;

  @ApiProperty({
    example: 'Doe',
    description: 'Last name of the user',
  })
  @IsString()
  lastName: string;

  @ApiProperty({
    enum: Role,
    example: Role.Student,
    description: 'User role in the system',
  })
  @IsEnum(Role)
  role: Role;

  @ApiProperty({
    example: '664f04ae712f6a932b4b01b1',
    description: 'ID of uploaded avatar file',
    required: false,
  })
  @IsOptional()
  @IsString()
  photo?: string;
}

export class UpdateAccountDto
  extends PartialType(CreateAccountDto)
  implements UpdateAccountDtoType
{
  @ApiProperty({
    enum: Status,
    example: Status.Active,
    description: 'Current status of the user',
    required: false,
  })
  @IsEnum(Status)
  @IsOptional()
  status?: Status;
}

export class ReadAccountDto implements ReadAccountDtoType {
  @ApiProperty({
    example: '664f0392712f6a932b4b01a2',
    description: 'Unique ID of the account',
  })
  id: string;

  @ApiProperty({
    example: 'John',
    description: 'First name of the user',
  })
  firstName: string;

  @ApiProperty({
    example: 'Doe',
    description: 'Last name of the user',
  })
  lastName: string;

  @ApiProperty({
    enum: Role,
    example: Role.Teacher,
    description: 'User role in the system',
  })
  role: Role;

  @ApiProperty({
    enum: Status,
    example: Status.Blocked,
    description: 'User status (active or blocked)',
  })
  status: Status;

  @ApiProperty({
    example: '664f04ae712f6a932b4b01b1',
    description: 'Avatar file ID (if uploaded)',
    required: false,
  })
  photo?: string;

  @ApiProperty({
    example: null,
    description: 'Soft deletion timestamp (null if active)',
    required: false,
  })
  deletedAt?: string | null;

  @ApiProperty({
    example: '2024-06-01T15:23:45.000Z',
    description: 'Date the account was created',
  })
  createdAt?: string;

  @ApiProperty({
    example: '2024-06-03T10:12:30.000Z',
    description: 'Date the account was last updated',
  })
  updatedAt?: string;
}
