import { randomUUID } from 'node:crypto';
import type { CreateScaffoldInput, UpdateScaffoldInput } from '../schemas/scaffolds.js';

export interface TestScaffold {
  id: string;
  projectId: string;
  type: CreateScaffoldInput['type'];
  framework: CreateScaffoldInput['framework'];
  language: CreateScaffoldInput['language'];
  config: Record<string, unknown>;
  createdAt: string; // ISO
  updatedAt: string; // ISO
}

export interface ScaffoldsRepo {
  create(projectId: string, input: CreateScaffoldInput): Promise<TestScaffold>;
  getById(id: string): Promise<TestScaffold | null>;
  listByProject(projectId: string, limit: number, cursor?: string): Promise<{ items: TestScaffold[]; nextCursor?: string }>; 
  update(id: string, patch: UpdateScaffoldInput): Promise<TestScaffold | null>;
  delete(id: string): Promise<boolean>;
}

export class InMemoryScaffoldsRepo implements ScaffoldsRepo {
  private store = new Map<string, TestScaffold>();
  private byProject = new Map<string, string[]>();

  async create(projectId: string, input: CreateScaffoldInput): Promise<TestScaffold> {
    const now = new Date().toISOString();
    const rec: TestScaffold = {
      id: randomUUID(),
      projectId,
      type: input.type,
      framework: input.framework,
      language: input.language,
      config: input.config ?? {},
      createdAt: now,
      updatedAt: now,
    };
    this.store.set(rec.id, rec);
    const arr = this.byProject.get(projectId) ?? [];
    arr.unshift(rec.id);
    this.byProject.set(projectId, arr);
    return rec;
  }

  async getById(id: string): Promise<TestScaffold | null> {
    return this.store.get(id) ?? null;
  }

  async listByProject(projectId: string, limit: number, cursor?: string): Promise<{ items: TestScaffold[]; nextCursor?: string }> {
    const ids = this.byProject.get(projectId) ?? [];
    let start = 0;
    if (cursor) {
      const idx = ids.indexOf(cursor);
      start = idx >= 0 ? idx + 1 : 0;
    }
    const slice = ids.slice(start, start + limit);
    const items = slice.map((id) => this.store.get(id)!).filter(Boolean);
    const nextCursor = ids.length > start + limit ? slice[slice.length - 1] : undefined;
    return { items, nextCursor };
  }

  async update(id: string, patch: UpdateScaffoldInput): Promise<TestScaffold | null> {
    const cur = this.store.get(id);
    if (!cur) return null;
    const updated: TestScaffold = {
      ...cur,
      framework: patch.framework ?? cur.framework,
      language: patch.language ?? cur.language,
      config: patch.config ?? cur.config,
      updatedAt: new Date().toISOString(),
    };
    this.store.set(id, updated);
    return updated;
  }

  async delete(id: string): Promise<boolean> {
    const cur = this.store.get(id);
    if (!cur) return false;
    this.store.delete(id);
    const arr = this.byProject.get(cur.projectId) ?? [];
    this.byProject.set(cur.projectId, arr.filter((x) => x !== id));
    return true;
  }
}
