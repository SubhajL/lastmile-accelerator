import { describe, it, expect, vi } from 'vitest';
import { createHandlebarsRenderer } from './renderer.js';
import { promises as fs } from 'fs';
import { tmpdir } from 'os';
import { join } from 'path';

async function mkTemplateDir(name: string, subject: string, html: string, text?: string) {
  const base = await fs.mkdtemp(join(tmpdir(), 'tmpl-'));
  const dir = join(base, name);
  await fs.mkdir(dir);
  await fs.writeFile(join(dir, 'subject.hbs'), subject, 'utf8');
  await fs.writeFile(join(dir, 'html.hbs'), html, 'utf8');
  if (text) await fs.writeFile(join(dir, 'text.hbs'), text, 'utf8');
  return { base, dir };
}

describe('channels/renderer', () => {
  it('renders subject, html, and text when provided', async () => {
    const { base } = await mkTemplateDir(
      'snapshot-ready',
      'Snapshot {{snapshotId}} ready',
      '<p>Project {{projectId}}</p>',
      'Project {{projectId}}'
    );

    const render = createHandlebarsRenderer({ templatesDir: base });
    const out = await render('snapshot-ready', { snapshotId: 's1', projectId: 'p1' });

    expect(out.subject).toBe('Snapshot s1 ready');
    expect(out.html).toBe('<p>Project p1</p>');
    expect(out.text).toBe('Project p1');
  });

  it('falls back to stripped text when text.hbs missing', async () => {
    const { base } = await mkTemplateDir(
      'snapshot-ready',
      'Ready',
      '<p>Hi <strong>there</strong></p>'
    );

    const render = createHandlebarsRenderer({ templatesDir: base });
    const out = await render('snapshot-ready', {});

    expect(out.text).toBe('Hi there');
  });
});
