import { Prop, Schema, SchemaFactory } from '@nestjs/mongoose';
import { Document } from 'mongoose';
import { Role, Status } from '@codeed/types';

export type AccountDocument = Account & Document;

@Schema({ timestamps: true })
export class Account {
  @Prop({ required: true })
  firstName: string;

  @Prop({ required: true })
  lastName: string;

  @Prop({ enum: Role, required: true })
  role: Role;

  @Prop({ enum: Status, default: Status.Active })
  status: Status;

  @Prop()
  photo?: string;

  @Prop()
  deletedAt?: Date;

  @Prop()
  createdAt?: Date;

  @Prop()
  updatedAt?: Date;
}

export const AccountSchema = SchemaFactory.createForClass(Account);
