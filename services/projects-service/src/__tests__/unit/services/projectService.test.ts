import { describe, it, expect, vi, beforeEach } from 'vitest';
import * as db from '../../../db';
import * as events from '../../../events/eventPublisher';
import { ValidationError, NotFoundError } from '../../../utils/errors';
import { parseCreateProject } from '../../../utils/validators';
import { listProjects, getProject, createProject, updateProject, deleteProject } from '../../../services/projectService';

describe('services/projectService.ts', () => {
  beforeEach(() => {
    vi.restoreAllMocks();
  });

  it('listProjects returns non-deleted projects for tenant ordered by created_at desc', async () => {
    const rows = [
      { id: 'p2', tenant_id: 't1', name: 'B', created_at: '2024-02-01', deleted_at: null },
      { id: 'p1', tenant_id: 't1', name: 'A', created_at: '2024-01-01', deleted_at: null },
    ];
    const q = vi.spyOn(db, 'query').mockResolvedValue({ rows } as any);

    const res = await listProjects('t1');
    expect(q).toHaveBeenCalled();
    const [sql, params] = q.mock.calls[0];
    expect(String(sql).toLowerCase()).toContain('from projects');
    expect(params).toContain('t1');
    expect(res.map(r => r.id)).toEqual(['p2', 'p1']);
  });

  it('getProject returns project for correct tenant or 404', async () => {
    const q = vi.spyOn(db, 'query');
    q.mockResolvedValueOnce({ rows: [{ id: 'p1', tenant_id: 't1', name: 'X' }] } as any);

    const p = await getProject('t1', 'p1');
    expect(p.id).toBe('p1');

    q.mockResolvedValueOnce({ rows: [] } as any);
    await expect(getProject('t2', 'missing')).rejects.toBeInstanceOf(NotFoundError);
  });

  it('createProject inserts project, default envs, owner member, publishes event', async () => {
    const trx = vi.spyOn(db, 'transaction').mockImplementation(async (fn: any) => {
      const client = { query: vi.fn().mockResolvedValue({ rows: [{ id: 'new-project-id' }] }) } as any;
      return fn(client);
    });
    const ev = vi.spyOn(events, 'publishProjectEvent').mockResolvedValue();

    const input = parseCreateProject({ name: 'New Project', description: 'Test' });
    const res = await createProject('t1', input, 'u1');

    expect(trx).toHaveBeenCalled();
    expect(ev).toHaveBeenCalledWith('created', expect.objectContaining({ tenantId: 't1' }), undefined);
    expect(res.id).toBeDefined();
  });

  it('updateProject updates fields and publishes project.updated', async () => {
    vi.spyOn(db, 'query')
      .mockResolvedValueOnce({ rows: [{ id: 'p1', tenant_id: 't1', name: 'Old' }] } as any) // existing
      .mockResolvedValueOnce({ rows: [{ id: 'p1', tenant_id: 't1', name: 'New' }] } as any); // updated

    const ev = vi.spyOn(events, 'publishProjectEvent').mockResolvedValue();

    const res = await updateProject('t1', 'p1', { name: 'New' });
    expect(res.name).toBe('New');
    expect(ev).toHaveBeenCalledWith('updated', expect.objectContaining({ projectId: 'p1', tenantId: 't1' }), undefined);
  });

  it('deleteProject soft-deletes and publishes project.deleted', async () => {
    vi.spyOn(db, 'query')
      .mockResolvedValueOnce({ rows: [{ id: 'p1', tenant_id: 't1' }] } as any) // existing
      .mockResolvedValueOnce({ rows: [] } as any); // delete ok
    const ev = vi.spyOn(events, 'publishProjectEvent').mockResolvedValue();

    await deleteProject('t1', 'p1');
    expect(ev).toHaveBeenCalledWith('deleted', expect.objectContaining({ projectId: 'p1', tenantId: 't1' }), undefined);
  });
});