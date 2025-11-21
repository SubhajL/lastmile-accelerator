import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { buildTestJobManifest, RunnerOrchestrator } from '../../../services/runner.js';
import { SUBJECTS } from '../../../events/contracts.js';
import * as k8sClient from '../../../clients/k8s.js';
import * as artifactsService from '../../../services/artifacts.js';

vi.mock('../../../clients/k8s.js');
vi.mock('../../../services/artifacts.js');

beforeEach(() => {
  vi.mocked(k8sClient.createJob).mockResolvedValue({ name: 'job-1' } as any);
  vi.mocked(k8sClient.getJobStatus).mockResolvedValue('succeeded');
  vi.mocked(artifactsService.uploadArtifactsMap).mockResolvedValue([{ bucket: 'b', key: 'artifacts/run/file.txt' }]);
});

afterEach(() => {
  vi.restoreAllMocks();
  vi.clearAllMocks();
});

describe('Runner service', () => {
  it('buildTestJobManifest sets labels, image, ttl, SA', () => {
    const m = buildTestJobManifest({
      runId: '11111111-1111-1111-1111-111111111111',
      projectId: '11111111-1111-1111-1111-111111111111',
      image: 'runner:latest',
      serviceAccount: 'sa',
      ttlSecondsAfterFinished: 600,
      env: [{ name: 'RUN_ID', value: '11111111-1111-1111-1111-111111111111' }],
    });
    expect(m.apiVersion).toBe('batch/v1');
    expect((m.metadata as any).name).toContain('test-run-');
    expect((m.spec as any).ttlSecondsAfterFinished).toBe(600);
    const container = (m.spec as any).template.spec.containers[0];
    expect(container.image).toBe('runner:latest');
    expect((m.spec as any).template.spec.serviceAccountName).toBe('sa');
  });

  it('orchestrator success flow updates repo, uploads artifacts, publishes events', async () => {
    const testRunsRepo = {
      getById: vi.fn().mockResolvedValue({ id: 'r', projectId: 'p' }),
      updateStatus: vi.fn().mockResolvedValue({ id: 'r', status: 'passed' }),
    };
    const nats = { publish: vi.fn() } as any;
    const batch = {} as any; // not used directly due to mocked createJob/getJobStatus
    const s3 = {} as any;

    const orch = new RunnerOrchestrator({
      config: {
        namespace: 'default',
        serviceAccount: 'sa',
        image: 'runner:img',
        ttlSecondsAfterFinished: 600,
        bucketArtifacts: 'b',
      },
      batch,
      s3,
      nats,
      testRunsRepo: testRunsRepo as any,
    });

    await orch.handleRunRequested({
      runId: '11111111-1111-1111-1111-111111111111',
      projectId: '11111111-1111-1111-1111-111111111111',
      artifacts: [{ name: 'file.txt', content: 'hello', contentType: 'text/plain' }],
    });

    expect(testRunsRepo.updateStatus).toHaveBeenCalledWith(
      '11111111-1111-1111-1111-111111111111',
      'running',
      expect.objectContaining({ startedAt: expect.any(Date) })
    );
    expect(nats.publish).toHaveBeenCalledWith(
      SUBJECTS.runStarted,
      expect.objectContaining({ status: 'running' })
    );
    expect(testRunsRepo.updateStatus).toHaveBeenCalledWith(
      '11111111-1111-1111-1111-111111111111',
      'passed',
      expect.objectContaining({ finishedAt: expect.any(Date) })
    );
    expect(nats.publish).toHaveBeenCalledWith(
      SUBJECTS.runFinished,
      expect.objectContaining({ status: 'passed' })
    );
  });
});
