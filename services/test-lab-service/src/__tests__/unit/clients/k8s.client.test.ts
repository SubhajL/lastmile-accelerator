import { describe, it, expect, vi } from 'vitest';
import type { BatchV1Api } from '@kubernetes/client-node';
import { createJob, getJobStatus } from '../../../clients/k8s.js';

function makeBatch(overrides: Partial<BatchV1Api> = {}): BatchV1Api {
  return {
    createNamespacedJob: vi.fn().mockResolvedValue({ body: { metadata: { name: 'job-1', uid: 'uid-1' } } }),
    readNamespacedJob: vi.fn().mockResolvedValue({ body: { status: { succeeded: 1 } } }),
    // @ts-ignore
    ...overrides,
  } as BatchV1Api;
}

describe('k8s client', () => {
  it('createJob uses namespace and returns metadata', async () => {
    const batch = makeBatch();
    const res = await createJob(batch, 'default', { apiVersion: 'batch/v1', kind: 'Job', metadata: { name: 'job-1' }, spec: {} } as any);
    expect(res).toEqual({ name: 'job-1', uid: 'uid-1' });
    expect((batch.createNamespacedJob as any).mock.calls[0][0]).toBe('default');
  });

  it('getJobStatus normalizes statuses', async () => {
    let batch = makeBatch();
    expect(await getJobStatus(batch, 'ns', 'name')).toBe('succeeded');

    batch = makeBatch({ readNamespacedJob: vi.fn().mockResolvedValue({ body: { status: { failed: 1 } } }) as any });
    expect(await getJobStatus(batch, 'ns', 'name')).toBe('failed');

    batch = makeBatch({ readNamespacedJob: vi.fn().mockResolvedValue({ body: { status: { active: 1 } } }) as any });
    expect(await getJobStatus(batch, 'ns', 'name')).toBe('running');

    batch = makeBatch({ readNamespacedJob: vi.fn().mockResolvedValue({ body: { status: {} } }) as any });
    expect(await getJobStatus(batch, 'ns', 'name')).toBe('pending');
  });
});
