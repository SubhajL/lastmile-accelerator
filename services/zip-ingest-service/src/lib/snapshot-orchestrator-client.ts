type CreateSnapshotReq = {
  mode: 'C';
  sourceRef: {
    zip: {
      filename: string;
      sizeBytes: number;
      sha256: string;
    };
  };
};

type SnapshotRes = {
  snapshotId: string;
};

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null;
}

export async function createSnapshotViaOrchestrator(args: {
  baseUrl: string;
  projectId: string;
  body: CreateSnapshotReq;
}): Promise<SnapshotRes> {
  const res = await fetch(`${args.baseUrl}/v1/projects/${args.projectId}/snapshots`, {
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
