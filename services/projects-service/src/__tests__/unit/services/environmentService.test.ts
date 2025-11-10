import { describe, it, expect, vi, beforeEach } from 'vitest';
import * as db from '../../../db';
import * as events from '../../../events/eventPublisher';
import {
  listEnvironments,
  createEnvironment,
  updateEnvironment,
  deleteEnvironment,
  setIngestionModes,
} from '../../../services/environmentService';
import { NotFoundError } from '../../../utils/errors';

function mockProjectExists(ok: boolean) {
  const spy = vi.spyOn(db, 'query');
  spy.mockResolvedValueOnce({ rows: ok ? [{ id: 'p1', tenant_id: 't1' }] : [] } as any);
  return spy;
}

describe('services/environmentService.ts', () => {
  beforeEach(() => { vi.restoreAllMocks(); });

  it('listEnvironments validates project tenant match', async () => {
    const q = mockProjectExists(true);
    q.mockResolvedValueOnce({ rows: [{ id: 'e1', name: 'dev' }] } as any);
    const rows = await listEnvironments('t1', 'p1');
    expect(rows[0].id).toBe('e1');

    mockProjectExists(false);
    await expect(listEnvironments('t2', 'pX')).rejects.toBeInstanceOf(NotFoundError);
  });

  it('createEnvironment inserts and publishes created', async () => {
    const spy = mockProjectExists(true);
    spy.mockResolvedValueOnce({ rows: [{ id: 'e1', name: 'qa' }] } as any);
    const ev = vi.spyOn(events, 'publishEnvironmentEvent').mockResolvedValue();

    const row = await createEnvironment('t1', 'p1', { name: 'qa' });
    expect(row.id).toBe('e1');
    expect(ev).toHaveBeenCalledWith('created', expect.objectContaining({ tenantId: 't1', projectId: 'p1' }), undefined);
  });

  it('updateEnvironment updates config and publishes updated', async () => {
    const spy = mockProjectExists(true);
    spy
      .mockResolvedValueOnce({ rows: [{ id: 'e1', name: 'dev', config: { a: 1 } }] } as any)
      .mockResolvedValueOnce({ rows: [{ id: 'e1', name: 'dev', config: { a: 2 } }] } as any);
    const ev = vi.spyOn(events, 'publishEnvironmentEvent').mockResolvedValue();

    const row = await updateEnvironment('t1', 'p1', 'e1', { config: { a: 2 } });
    expect(row.config.a).toBe(2);
    expect(ev).toHaveBeenCalledWith('updated', expect.objectContaining({ environmentId: 'e1', projectId: 'p1', tenantId: 't1' }), undefined);
  });

  it('deleteEnvironment blocks when last env', async () => {
    mockProjectExists(true);
    vi.spyOn(db, 'transaction').mockImplementation(async (fn: any) => {
      const client = { query: vi.fn().mockResolvedValueOnce({ rows: [{ cnt: '1' }] }) } as any;
      return fn(client);
    });

    await expect(deleteEnvironment('t1', 'p1', 'e1')).rejects.toThrow('last environment');

    mockProjectExists(true);
    vi.spyOn(db, 'transaction').mockImplementation(async (fn: any) => {
      const client = { query: vi.fn()
        .mockResolvedValueOnce({ rows: [{ cnt: '2' }] })
        .mockResolvedValueOnce({ rows: [] }) } as any;
      return fn(client);
    });
    const ev = vi.spyOn(events, 'publishEnvironmentEvent').mockResolvedValue();
    await deleteEnvironment('t1', 'p1', 'e1');
    expect(ev).toHaveBeenCalledWith('deleted', expect.objectContaining({ environmentId: 'e1', projectId: 'p1', tenantId: 't1' }), undefined);
  });

  it('setIngestionModes updates and publishes', async () => {
    mockProjectExists(true);
    vi.spyOn(db, 'transaction').mockImplementation(async (fn: any) => {
      const client = { query: vi.fn().mockResolvedValue({ rows: [] }) } as any;
      return fn(client);
    });
    const ev = vi.spyOn(events, 'publishEnvironmentEvent').mockResolvedValue();

    await setIngestionModes('t1', 'p1', { modes: ['A', 'B'], defaultMode: 'B' });
    expect(ev).toHaveBeenCalledWith('updated', expect.objectContaining({ projectId: 'p1', tenantId: 't1', ingestionModes: ['A', 'B'], defaultMode: 'B' }), undefined);
  });
});
