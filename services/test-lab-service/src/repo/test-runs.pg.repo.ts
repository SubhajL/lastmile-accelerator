import type { Pool } from 'pg';

export interface TestRunRecord {
  id: string;
  projectId: string;
  snapshotId?: string | null;
  type: string;
  status: string;
  startedAt?: string | null;
  finishedAt?: string | null;
  results?: Record<string, unknown> | null;
  artifacts?: Record<string, unknown> | null;
  createdAt: string;
}

export interface CreateTestRunInput {
  type: string;
  snapshotId?: string;
  status?: string; // default 'queued'
}

export interface ListRunsOptions {
  limit: number;
  cursor?: string;
  status?: string;
  type?: string;
  from?: Date;
  to?: Date;
}

export class PgTestRunsRepo {
  constructor(private pool: Pool) {}

  private mapRow(row: any): TestRunRecord {
    const iso = (d: any) => (d instanceof Date ? d.toISOString() : d ? String(d) : null);
    return {
      id: row.id,
      projectId: row.project_id,
      snapshotId: row.snapshot_id,
      type: row.type,
      status: row.status,
      startedAt: iso(row.started_at),
      finishedAt: iso(row.finished_at),
      results: row.results,
      artifacts: row.artifacts,
      createdAt: iso(row.created_at)!,
    };
  }

  async create(projectId: string, input: CreateTestRunInput): Promise<TestRunRecord> {
    const { randomUUID } = await import('node:crypto');
    const id = randomUUID();
    const status = input.status ?? 'queued';
    const res = await this.pool.query(
      `INSERT INTO test_runs (id, project_id, snapshot_id, type, status, created_at)
       VALUES ($1, $2, $3, $4, $5, now())
       RETURNING *`,
      [id, projectId, input.snapshotId ?? null, input.type, status]
    );
    return this.mapRow(res.rows[0]);
  }

  async getById(id: string): Promise<TestRunRecord | null> {
    const res = await this.pool.query(`SELECT * FROM test_runs WHERE id=$1`, [id]);
    return res.rowCount ? this.mapRow(res.rows[0]) : null;
  }

  async listByProject(projectId: string, opts: ListRunsOptions): Promise<{ items: TestRunRecord[]; nextCursor?: string }> {
    const params: any[] = [projectId];
    let where = 'project_id = $1';
    let idx = params.length + 1;
    if (opts.status) { params.push(opts.status); where += ` AND status = $${idx++}`; }
    if (opts.type) { params.push(opts.type); where += ` AND type = $${idx++}`; }
    if (opts.from) { params.push(opts.from); where += ` AND created_at >= $${idx++}`; }
    if (opts.to) { params.push(opts.to); where += ` AND created_at <= $${idx++}`; }

    if (opts.cursor) {
      const c = await this.pool.query(`SELECT created_at, id FROM test_runs WHERE id=$1`, [opts.cursor]);
      if ((c.rowCount ?? 0) > 0) {
        params.push(c.rows[0].created_at, c.rows[0].id);
        where += ` AND (created_at < $${idx} OR (created_at = $${idx} AND id < $${idx + 1}))`;
        idx += 2;
      }
    }
    params.push(opts.limit + 1);

    const res = await this.pool.query(`SELECT * FROM test_runs WHERE ${where} ORDER BY created_at DESC, id DESC LIMIT $${params.length}`, params);
    const rows = res.rows;
    const items = rows.slice(0, opts.limit).map(r => this.mapRow(r));
    const nextCursor = rows.length > opts.limit ? rows[opts.limit - 1].id : undefined;
    return { items, nextCursor };
  }

  async updateStatus(id: string, status: string, fields?: { startedAt?: Date; finishedAt?: Date; results?: any; artifacts?: any }): Promise<TestRunRecord | null> {
    const results = fields?.results === undefined ? null : JSON.stringify(fields.results);
    const artifacts = fields?.artifacts === undefined ? null : JSON.stringify(fields.artifacts);
    const res = await this.pool.query(
      `UPDATE test_runs SET
         status = $2,
         started_at = COALESCE($3, started_at),
         finished_at = COALESCE($4, finished_at),
         results = COALESCE($5::jsonb, results),
         artifacts = COALESCE($6::jsonb, artifacts)
       WHERE id=$1
       RETURNING *`,
      [id, status, fields?.startedAt ?? null, fields?.finishedAt ?? null, results, artifacts]
    );
    return res.rowCount ? this.mapRow(res.rows[0]) : null;
  }

  async delete(id: string): Promise<boolean> {
    const res = await this.pool.query(`DELETE FROM test_runs WHERE id=$1`, [id]);
    return (res.rowCount ?? 0) > 0;
  }
}
