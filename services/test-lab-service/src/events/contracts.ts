import { z } from 'zod';

// Subjects
export const SUBJECTS = {
  runRequest: 'test-lab.run.request',
  runStarted: 'test-lab.run.started',
  runFinished: 'test-lab.run.finished',
  browserRequest: 'test-lab.browser.request',
  browserStarted: 'test-lab.browser.started',
  browserFinished: 'test-lab.browser.finished',
} as const;

const uuid = () => z.string().uuid();
const iso = () => z.string().refine((s) => !Number.isNaN(Date.parse(s)), 'must be ISO date string');

// Run events
export const zRunRequested = z.object({
  runId: uuid(),
  projectId: uuid(),
  type: z.enum(['unit', 'e2e', 'smoke']).optional(),
  artifacts: z
    .array(
      z.object({ name: z.string(), content: z.string(), contentType: z.string().optional() })
    )
    .optional(),
});
export type RunRequested = z.infer<typeof zRunRequested>;

export const zRunStarted = z.object({
  runId: uuid(),
  projectId: uuid(),
  status: z.literal('running'),
  startedAt: iso(),
});
export type RunStarted = z.infer<typeof zRunStarted>;

export const zRunFinished = z.object({
  runId: uuid(),
  projectId: uuid(),
  status: z.enum(['passed', 'failed', 'cancelled']),
  finishedAt: iso(),
  artifacts: z
    .array(z.object({ bucket: z.string(), key: z.string() }))
    .optional(),
  results: z.any().optional(),
});
export type RunFinished = z.infer<typeof zRunFinished>;

// Browser events
export const zBrowserRequested = z.object({
  browserRunId: uuid(),
  testRunId: uuid(),
  browser: z.string(),
  os: z.string().optional(),
  viewport: z.string().optional(),
});
export type BrowserRequested = z.infer<typeof zBrowserRequested>;

export const zBrowserStarted = z.object({
  browserRunId: uuid(),
  testRunId: uuid(),
  status: z.literal('running'),
  startedAt: iso(),
});
export type BrowserStarted = z.infer<typeof zBrowserStarted>;

export const zBrowserFinished = z.object({
  browserRunId: uuid(),
  testRunId: uuid(),
  status: z.enum(['passed', 'failed', 'cancelled']),
  finishedAt: iso(),
  screenshots: z.array(z.object({ bucket: z.string(), key: z.string() })).optional(),
  logs: z.any().optional(),
});
export type BrowserFinished = z.infer<typeof zBrowserFinished>;
