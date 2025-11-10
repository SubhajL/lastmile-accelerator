import { describe, it, expect, vi } from 'vitest';
import * as db from '../../../db';
import * as svc from '../../../services/environmentService';
import * as events from '../../../events/eventPublisher';

describe('environmentService transactional integrity', () => {
  it('setIngestionModes runs delete+inserts inside one transaction', async () => {
    vi.spyOn(events, 'publishEnvironmentEvent').mockResolvedValue();
    // stub project existence check
    vi.spyOn(db, 'query').mockResolvedValueOnce({ rows: [{ id: 'p1', tenant_id: 't1' }] } as any);

    const client = { query: vi.fn().mockResolvedValue({ rows: [] }) } as any;
    const trx = vi.spyOn(db, 'transaction').mockImplementation(async (fn: any) => fn(client));

    await svc.setIngestionModes('t1', 'p1', { modes: ['A','B'], defaultMode: 'B' });

    expect(trx).toHaveBeenCalledTimes(1);
    expect(client.query).toHaveBeenCalledWith(expect.stringMatching(/DELETE FROM ingestion_modes/), ['p1']);
    // two inserts
    expect(client.query).toHaveBeenCalledWith(expect.stringMatching(/INSERT INTO ingestion_modes/), ['p1', 'A', false]);
    expect(client.query).toHaveBeenCalledWith(expect.stringMatching(/INSERT INTO ingestion_modes/), ['p1', 'B', true]);
  });

  it('deleteEnvironment runs count+delete inside one transaction', async () => {
    vi.spyOn(events, 'publishEnvironmentEvent').mockResolvedValue();
    // stub project existence check
    vi.spyOn(db, 'query').mockResolvedValueOnce({ rows: [{ id: 'p1', tenant_id: 't1' }] } as any);

    const trx = vi.spyOn(db, 'transaction').mockImplementation(async (fn: any) => {
      const txClient = { query: vi.fn()
        .mockResolvedValueOnce({ rows: [{ cnt: '2' }] }) // count
        .mockResolvedValueOnce({ rows: [] }) // delete
      } as any;
      return fn(txClient);
    });

    await svc.deleteEnvironment('t1', 'p1', 'e1');
    expect(trx).toHaveBeenCalledTimes(1);
  });
});