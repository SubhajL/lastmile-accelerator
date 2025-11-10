import type { Pool } from 'pg';

export interface BrowserTestRunRecord {
  id: string;
  testRunId: string;
  browser: string;
  os?: string | null;
  viewport?: string | null;
  status: string;
  screenshots?: any | null;
  logs?: any | null;
  startedAt?: string | null;
  finishedAt?: string | null;
}

export interface CreateBrowserTestRunInput {
  browser: string;
  os?: string;
  viewport?: string;
  status?: string; // default 'queued'
}

export class PgBrowserTestRunsRepo {
  constructor(private pool: Pool) {}

  private mapRow(row: any): BrowserTestRunRecord {
    const iso = (d: any) => (d instanceof Date ? d.toISOString() : d ? String(d) : null);
    return {
      id: row.id,
      testRunId: row.test_run_id,
      browser: row.browser,
      os: row.os,
      viewport: row.viewport,
      status: row.status,
      screenshots: row.screenshots,
      logs: row.logs,
      startedAt: iso(row.started_at),
      finishedAt: iso(row.finished_at),
    };
  }

  async create(testRunId: string, input: CreateBrowserTestRunInput): Promise<BrowserTestRunRecord> {
    const { randomUUID } = await import('node:crypto');
    const id = randomUUID();
    const status = input.status ?? 'queued';
    const res = await this.pool.query(
      `INSERT INTO browser_test_runs (id, test_run_id, browser, os, viewport, status)
       VALUES ($1, $2, $3, $4, $5, $6)
       RETURNING *`,
      [id, testRunId, input.browser, input.os ?? null, input.viewport ?? null, status]
    );
    return this.mapRow(res.rows[0]);
  }

  async getById(id: string): Promise<BrowserTestRunRecord | null> {
    const res = await this.pool.query(`SELECT * FROM browser_test_runs WHERE id=$1`, [id]);
    return res.rowCount ? this.mapRow(res.rows[0]) : null;
  }

  async listByTestRun(testRunId: string): Promise<{ items: BrowserTestRunRecord[] }> {
    const res = await this.pool.query(
      `SELECT * FROM browser_test_runs WHERE test_run_id=$1 ORDER BY id ASC`,
      [testRunId]
    );
    return { items: res.rows.map(r => this.mapRow(r)) };
  }

  async updateStatus(
    id: string,
    status: string,
    fields?: { startedAt?: Date; finishedAt?: Date; screenshots?: any; logs?: any }
  ): Promise<BrowserTestRunRecord | null> {
    const screenshots = fields?.screenshots === undefined ? null : JSON.stringify(fields.screenshots);
    const logs = fields?.logs === undefined ? null : JSON.stringify(fields.logs);
    const res = await this.pool.query(
      `UPDATE browser_test_runs SET
         status = $2,
         started_at = COALESCE($3, started_at),
         finished_at = COALESCE($4, finished_at),
         screenshots = COALESCE($5::jsonb, screenshots),
         logs = COALESCE($6::jsonb, logs)
       WHERE id=$1
       RETURNING *`,
      [id, status, fields?.startedAt ?? null, fields?.finishedAt ?? null, screenshots, logs]
    );
    return res.rowCount ? this.mapRow(res.rows[0]) : null;
  }

  async delete(id: string): Promise<boolean> {
    const res = await this.pool.query(`DELETE FROM browser_test_runs WHERE id=$1`, [id]);
    return (res.rowCount ?? 0) > 0;
  }
}
