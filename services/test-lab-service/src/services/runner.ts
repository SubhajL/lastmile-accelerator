import type { BatchV1Api } from '@kubernetes/client-node';
import type { S3Client } from '@aws-sdk/client-s3';
import type { NatsWrapper } from '../clients/nats.js';
import { createJob, getJobStatus } from '../clients/k8s.js';
import { uploadArtifactsMap, type ArtifactInput } from './artifacts.js';
import { SUBJECTS } from '../events/contracts.js';

export interface TestRunsRepoLike {
  getById(id: string): Promise<{ id: string; projectId: string } | null>;
  updateStatus(id: string, status: string, fields?: { startedAt?: Date; finishedAt?: Date; results?: any; artifacts?: any }): Promise<any>;
}

export function buildTestJobManifest(opts: {
  runId: string;
  projectId: string;
  image: string;
  serviceAccount: string;
  ttlSecondsAfterFinished: number;
  env?: { name: string; value: string }[];
}) {
  const name = `test-run-${opts.runId.slice(0, 8)}`;
  return {
    apiVersion: 'batch/v1',
    kind: 'Job',
    metadata: {
      name,
      labels: {
        app: 'test-lab-service',
        runId: opts.runId,
        projectId: opts.projectId,
      },
    },
    spec: {
      ttlSecondsAfterFinished: opts.ttlSecondsAfterFinished,
      template: {
        metadata: { labels: { app: 'test-lab-service', runId: opts.runId, projectId: opts.projectId } },
        spec: {
          serviceAccountName: opts.serviceAccount,
          restartPolicy: 'Never',
          containers: [
            {
              name: 'runner',
              image: opts.image,
              env: opts.env ?? [],
              command: ['sh', '-lc', 'echo Running tests... && sleep 1'],
            },
          ],
        },
      },
    },
  } as any;
}

export class RunnerOrchestrator {
  constructor(
    private deps: {
      config: {
        namespace: string;
        serviceAccount: string;
        image: string;
        ttlSecondsAfterFinished: number;
        bucketArtifacts: string;
      };
      batch: BatchV1Api;
      s3: S3Client;
      nats: NatsWrapper;
      testRunsRepo: TestRunsRepoLike;
    }
  ) {}

  async handleRunRequested(req: { runId: string; projectId: string; artifacts?: ArtifactInput[] }) {
    const { runId, projectId } = req;
    // Set running
    await this.deps.testRunsRepo.updateStatus(runId, 'running', { startedAt: new Date() });
    await this.deps.nats.publish(SUBJECTS.runStarted, {
      runId,
      projectId,
      status: 'running',
      startedAt: new Date().toISOString(),
    });

    // Create Job
    const manifest = buildTestJobManifest({
      runId,
      projectId,
      image: this.deps.config.image,
      serviceAccount: this.deps.config.serviceAccount,
      ttlSecondsAfterFinished: this.deps.config.ttlSecondsAfterFinished,
      env: [{ name: 'RUN_ID', value: runId }, { name: 'PROJECT_ID', value: projectId }],
    });
    await createJob(this.deps.batch, this.deps.config.namespace, manifest);

    // Poll once (simplified)
    const phase = await getJobStatus(this.deps.batch, this.deps.config.namespace, manifest.metadata.name);

    if (phase === 'succeeded') {
      const uploaded = req.artifacts && req.artifacts.length
        ? await uploadArtifactsMap(this.deps.s3, { bucket: this.deps.config.bucketArtifacts, prefix: `artifacts/${projectId}/${runId}` }, req.artifacts)
        : [];
      await this.deps.testRunsRepo.updateStatus(runId, 'passed', { finishedAt: new Date(), artifacts: uploaded });
      await this.deps.nats.publish(SUBJECTS.runFinished, {
        runId,
        projectId,
        status: 'passed',
        finishedAt: new Date().toISOString(),
        artifacts: uploaded,
      });
    } else if (phase === 'failed') {
      await this.deps.testRunsRepo.updateStatus(runId, 'failed', { finishedAt: new Date() });
      await this.deps.nats.publish(SUBJECTS.runFinished, {
        runId,
        projectId,
        status: 'failed',
        finishedAt: new Date().toISOString(),
      });
    }
  }
}
