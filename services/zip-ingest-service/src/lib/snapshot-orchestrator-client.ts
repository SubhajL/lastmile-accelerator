export type ZipSourceRefWithSha256 = {
  filename: string;
  sizeBytes: number;
  sha256: string;
  storageKey?: string;
  uploadId?: string;
};

export type ZipSourceRefCreate = {
  filename: string;
  sizeBytes: number;
  sha256?: string;
  storageKey?: string;
  uploadId?: string;
};

type CreateSnapshotReq = {
  mode: 'C';
  sourceRef: {
    zip: ZipSourceRefCreate;
  };
};

type SnapshotRes = {
  snapshotId: string;
};

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null && !Array.isArray(value);
}

export async function createSnapshotViaOrchestrator(args: {
  baseUrl: string;
  projectId: string;
  body: CreateSnapshotReq;
}): Promise<SnapshotRes> {
  const projectId = encodeURIComponent(args.projectId);
  const res = await fetch(`${args.baseUrl}/v1/projects/${projectId}/snapshots`, {
    method: 'POST',
    headers: { 'content-type': 'application/json' },
    body: JSON.stringify(args.body),
  });

  if (!res.ok) {
    throw new Error(`snapshot-orchestrator error: status=${res.status}`);
  }

  const json = (await res.json()) as unknown;
  const snapshotId = isRecord(json) && typeof json.snapshotId === 'string' ? json.snapshotId : null;
  if (!snapshotId) {
    throw new Error('snapshot-orchestrator returned invalid response');
  }

  return { snapshotId };
}
