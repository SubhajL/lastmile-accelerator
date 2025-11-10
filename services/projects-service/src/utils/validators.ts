import { z } from 'zod';
import { ValidationError } from './errors';

const NonEmptyString = z.string().min(1);

const CreateProjectSchema = z.object({
  name: NonEmptyString,
  description: z.string().optional(),
});

const UpdateProjectSchema = z.object({
  name: z.string().min(1).optional(),
  description: z.string().optional(),
}).refine(val => Object.keys(val).length > 0, {
  message: 'At least one field must be provided',
});

const CreateMemberSchema = z.object({
  email: z.string().email(),
  role: z.enum(['owner', 'admin', 'developer', 'viewer']),
});

const CreateEnvironmentSchema = z.object({
  name: NonEmptyString,
  config: z.record(z.any()).optional(),
});

const SetIngestionModesSchema = z.object({
  modes: z.array(z.enum(['A', 'B', 'C'])).min(1).refine(arr => new Set(arr).size === arr.length, {
    message: 'modes must be unique',
  }),
  defaultMode: z.enum(['A', 'B', 'C']),
}).refine(val => val.modes.includes(val.defaultMode), {
  message: 'defaultMode must be included in modes',
});

function wrap<T>(schema: z.ZodType<T>, input: unknown): T {
  const res = schema.safeParse(input);
  if (!res.success) {
    const msg = res.error.issues.map(i => i.message).join('; ');
    throw new ValidationError(`Invalid input: ${msg}`);
  }
  return res.data as T;
}

export function parseCreateProject(input: unknown) {
  return wrap(CreateProjectSchema, input);
}

export function parseUpdateProject(input: unknown) {
  return wrap(UpdateProjectSchema, input);
}

export function parseCreateMember(input: unknown) {
  return wrap(CreateMemberSchema, input);
}

export function parseCreateEnvironment(input: unknown) {
  return wrap(CreateEnvironmentSchema, input);
}

export function parseSetIngestionModes(input: unknown) {
  return wrap(SetIngestionModesSchema, input);
}