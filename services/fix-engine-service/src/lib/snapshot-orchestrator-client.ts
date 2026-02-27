function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null && !Array.isArray(value);
}

export async function getSnapshotOrchestratorSnapshot(args: {
  baseUrl: string;
  snapshotId: string;
}): Promise<'ok' | 'not_found' | 'unavailable'> {
  const snapshotId = encodeURIComponent(args.snapshotId);
  const res = await fetch(`${args.baseUrl}/v1/snapshots/${snapshotId}`, {
    method: 'GET',
    headers: { accept: 'application/json' },
  });

  if (res.status === 404) return 'not_found';
  if (!res.ok) return 'unavailable';

  const json = (await res.json()) as unknown;
  const ok = isRecord(json) && typeof json.snapshotId === 'string';
  return ok ? 'ok' : 'unavailable';
}
