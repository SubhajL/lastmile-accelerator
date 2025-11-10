import type { Pool } from 'pg';
import type { CreateScaffoldInput, UpdateScaffoldInput } from '../schemas/scaffolds.js';
import type { ScaffoldsRepo, TestScaffold } from './scaffolds.repo.js';

export class PgScaffoldsRepo implements ScaffoldsRepo {
  constructor(private pool: Pool) {}

  private mapRow(row: any): TestScaffold {
    const createdAt = row.created_at instanceof Date ? row.created_at.toISOString() : String(row.created_at);
    const updatedAt = row.updated_at instanceof Date ? row.updated_at.toISOString() : String(row.updated_at);
    return {
      id: row.id,
      projectId: row.projectId ?? row.project_id,
      type: row.type,
      framework: row.framework,
      language: row.language,
      config: row.config ?? {},
      createdAt,
      updatedAt,
    };
  }

  async create(projectId: string, input: CreateScaffoldInput): Promise<TestScaffold> {
    const { randomUUID } = await import('node:crypto');
    const id = randomUUID();
    const res = await this.pool.query(
      `INSERT INTO test_scaffolds (id, project_id, type, framework, language, config)
       VALUES ($1, $2, $3, $4, $5, $6)
       RETURNING id, project_id, type, framework, language, config, created_at, updated_at`,
      [id, projectId, input.type, input.framework, input.language, input.config ?? {}]
    );
    return this.mapRow(res.rows[0]);
  }

  async getById(id: string): Promise<TestScaffold | null> {
    const res = await this.pool.query(
      `SELECT id, project_id, type, framework, language, config, created_at, updated_at FROM test_scaffolds WHERE id=$1`,
      [id]
    );
    return res.rowCount ? this.mapRow(res.rows[0]) : null;
  }

  async listByProject(projectId: string, limit: number, cursor?: string): Promise<{ items: TestScaffold[]; nextCursor?: string }> {
    let createdAtCursor: Date | null = null;
    let idCursor: string | null = null;
    if (cursor) {
      const c = await this.pool.query<{ created_at: Date; id: string }>(`SELECT created_at, id FROM test_scaffolds WHERE id=$1`, [cursor]);
      if ((c.rowCount ?? 0) > 0) {
        createdAtCursor = c.rows[0].created_at;
        idCursor = c.rows[0].id;
      }
    }

    const params: any[] = [projectId];
    let where = 'project_id = $1';
    if (createdAtCursor && idCursor) {
      params.push(createdAtCursor, idCursor);
      where += ` AND (created_at < $2 OR (created_at = $2 AND id < $3))`;
    }
    params.push(limit + 1);

    const res = await this.pool.query(
      `SELECT id, project_id, type, framework, language, config, created_at, updated_at
       FROM test_scaffolds WHERE ${where}
       ORDER BY created_at DESC, id DESC
       LIMIT $${params.length}`,
      params
    );
    const rows = res.rows;
    const items = rows.slice(0, limit).map((r) => this.mapRow(r));
    const nextCursorId = rows.length > limit ? rows[limit - 1].id : undefined;
    return { items, nextCursor: nextCursorId };
  }

  async update(id: string, patch: UpdateScaffoldInput): Promise<TestScaffold | null> {
    const res = await this.pool.query(
      `UPDATE test_scaffolds SET
        framework = COALESCE($2, framework),
        language = COALESCE($3, language),
        config = COALESCE($4, config),
        updated_at = now()
       WHERE id=$1
       RETURNING id, project_id, type, framework, language, config, created_at, updated_at`,
      [id, patch.framework ?? null, patch.language ?? null, patch.config ?? null]
    );
    return res.rowCount ? this.mapRow(res.rows[0]) : null;
  }

  async delete(id: string): Promise<boolean> {
    const res = await this.pool.query(`DELETE FROM test_scaffolds WHERE id=$1`, [id]);
    return (res.rowCount ?? 0) > 0;
  }
}
