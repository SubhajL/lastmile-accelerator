import type { Pool } from 'pg';

export interface PreviewEnvRecord {
  id: string;
  projectId: string;
  snapshotId?: string | null;
  url: string;
  status: string;
  createdAt: string;
  expiresAt?: string | null;
}

export interface CreatePreviewInput {
  url: string;
  snapshotId?: string;
  status?: string; // default 'creating'
}

export interface ListPreviewsOptions {
  limit: number;
  cursor?: string;
}

export interface UpdatePreviewInput {
  url?: string;
  status?: string;
  expiresAt?: Date;
}

export class PgPreviewEnvsRepo {
  constructor(private pool: Pool) {}

  private mapRow(row: any): PreviewEnvRecord {
    const iso = (d: any) => (d instanceof Date ? d.toISOString() : d ? String(d) : null);
    return {
      id: row.id,
      projectId: row.project_id,
      snapshotId: row.snapshot_id,
      url: row.url,
      status: row.status,
      createdAt: iso(row.created_at)!,
      expiresAt: iso(row.expires_at),
    };
  }

  async create(projectId: string, input: CreatePreviewInput): Promise<PreviewEnvRecord> {
    const { randomUUID } = await import('node:crypto');
    const id = randomUUID();
    const status = input.status ?? 'creating';
    const res = await this.pool.query(
      `INSERT INTO preview_environments (id, project_id, snapshot_id, url, status)
       VALUES ($1, $2, $3, $4, $5)
       RETURNING *`,
      [id, projectId, input.snapshotId ?? null, input.url, status]
    );
    return this.mapRow(res.rows[0]);
  }

  async getById(id: string): Promise<PreviewEnvRecord | null> {
    const res = await this.pool.query(`SELECT * FROM preview_environments WHERE id=$1`, [id]);
    return res.rowCount ? this.mapRow(res.rows[0]) : null;
  }

  async listByProject(projectId: string, opts: ListPreviewsOptions): Promise<{ items: PreviewEnvRecord[]; nextCursor?: string }> {
    const params: any[] = [projectId];
    let where = 'project_id = $1';
    let idx = 2;
    if (opts.cursor) {
      const c = await this.pool.query(`SELECT created_at, id FROM preview_environments WHERE id=$1`, [opts.cursor]);
      if ((c.rowCount ?? 0) > 0) {
        params.push(c.rows[0].created_at, c.rows[0].id);
        where += ` AND (created_at < $${idx} OR (created_at = $${idx} AND id < $${idx + 1}))`;
        idx += 2;
      }
    }
    params.push(opts.limit + 1);

    const res = await this.pool.query(
      `SELECT * FROM preview_environments WHERE ${where} ORDER BY created_at DESC, id DESC LIMIT $${params.length}`,
      params
    );
    const rows = res.rows;
    const items = rows.slice(0, opts.limit).map((r) => this.mapRow(r));
    const nextCursor = rows.length > opts.limit ? rows[opts.limit - 1].id : undefined;
    return { items, nextCursor };
  }

  async update(id: string, patch: UpdatePreviewInput): Promise<PreviewEnvRecord | null> {
    const res = await this.pool.query(
      `UPDATE preview_environments SET
         url = COALESCE($2, url),
         status = COALESCE($3, status),
         expires_at = COALESCE($4, expires_at)
       WHERE id=$1
       RETURNING *`,
      [id, patch.url ?? null, patch.status ?? null, patch.expiresAt ?? null]
    );
    return res.rowCount ? this.mapRow(res.rows[0]) : null;
  }

  async delete(id: string): Promise<boolean> {
    const res = await this.pool.query(`DELETE FROM preview_environments WHERE id=$1`, [id]);
    return (res.rowCount ?? 0) > 0;
  }
}
