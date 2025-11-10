import type { BatchV1Api, CoreV1Api, V1Job } from '@kubernetes/client-node';

export interface K8sClients {
  batch: BatchV1Api;
  core: CoreV1Api;
}

export async function createK8sClient(): Promise<K8sClients> {
  const mod = await import('@kubernetes/client-node');
  const kc = new mod.KubeConfig();
  kc.loadFromDefault();
  return { batch: kc.makeApiClient(mod.BatchV1Api), core: kc.makeApiClient(mod.CoreV1Api) } as K8sClients;
}

export async function createJob(batch: BatchV1Api, namespace: string, manifest: V1Job): Promise<{ name: string; uid?: string }> {
  const res = await batch.createNamespacedJob(namespace, manifest);
  const meta = (res as any).body.metadata || ({} as any);
  return { name: meta.name!, uid: meta.uid };
}

export type JobPhase = 'pending' | 'running' | 'succeeded' | 'failed';

export async function getJobStatus(batch: BatchV1Api, namespace: string, name: string): Promise<JobPhase> {
  const res = await batch.readNamespacedJob(name, namespace);
  const status = (res as any).body.status || ({} as any);
  if (status.succeeded && status.succeeded > 0) return 'succeeded';
  if (status.failed && status.failed > 0) return 'failed';
  if (status.active && status.active > 0) return 'running';
  return 'pending';
}
