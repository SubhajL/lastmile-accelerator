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
  if (
    !json ||
    typeof json !== 'object' ||
    !('snapshotId' in json) ||
    typeof (json as any).snapshotId !== 'string'
  ) {
    throw new Error('snapshot-orchestrator returned invalid response');
  }

  return { snapshotId: (json as any).snapshotId };
}
