import { describe, it, expect, vi } from 'vitest';
import { createHandlebarsRenderer } from '../channels/renderer.js';
import { promises as fs } from 'fs';
import { join } from 'path';

describe('templates rendering', () => {
  it('renders publish-failed template with subject/html/text', async () => {
    const tmplDir = 'src/templates';
    const render = createHandlebarsRenderer({ templatesDir: tmplDir });

    // Ensure files exist (we just created them in repo)
    await fs.stat(join(tmplDir, 'publish-failed/subject.hbs'));

    const out = await render('publish-failed', { snapshotId: 's1', error: 'boom' });
    expect(out.subject).toContain('Publish failed');
    expect(out.html).toContain('boom');
    expect(out.text).toContain('boom');
  });
});
