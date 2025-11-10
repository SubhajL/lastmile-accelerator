import { z } from 'zod';

export const TestRunType = z.enum(['unit', 'e2e', 'smoke']);
export const TestRunStatus = z.enum(['queued', 'running', 'passed', 'failed', 'cancelled']);

export const CreateTestRunSchema = z.object({
  type: TestRunType,
  snapshotId: z.string().uuid().optional(),
});

export const ListTestRunsQuerySchema = z.object({
  cursor: z.preprocess((v) => (v === 'undefined' ? undefined : v), z.string().uuid().optional()),
  status: TestRunStatus.optional(),
  limit: z.string().transform(Number).pipe(z.number().int().min(1).max(500)).optional().default('50' as any),
});

export const UpdateTestRunStatusSchema = z.object({
  status: TestRunStatus,
  startedAt: z.coerce.date().optional(),
  finishedAt: z.coerce.date().optional(),
  results: z.any().optional(),
  artifacts: z.any().optional(),
});

export const CreateBrowserTestRunSchema = z.object({
  browser: z.string().min(1),
  os: z.string().optional(),
  viewport: z.string().optional(),
});

export const UpdateBrowserRunStatusSchema = z.object({
  status: TestRunStatus,
  startedAt: z.coerce.date().optional(),
  finishedAt: z.coerce.date().optional(),
  screenshots: z.any().optional(),
  logs: z.any().optional(),
});
