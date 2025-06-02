export enum Role {
  Student = 'student',
  Teacher = 'teacher',
  Admin = 'admin',
}

export enum Status {
  Active = 'active',
  Blocked = 'blocked',
}

export interface Account {
  id: string;
  firstName: string;
  lastName: string;
  role: Role;
  status: Status;
  photo?: string;
  deletedAt?: string | null;
  createdAt?: string;
  updatedAt?: string;
}

export interface CreateAccountDto {
  firstName: string;
  lastName: string;
  role: Role;
  photo?: string;
}

export interface UpdateAccountDto extends Partial<CreateAccountDto> {
  status?: Status;
}

export type ReadAccountDto = Account;
