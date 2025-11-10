import { promises as fsPromises } from 'fs';
import { join } from 'path';
import type { TemplateBundle } from './store.js';
import { validateTemplateBundle } from './validate.js';

export interface TemplateLoaderOptions {
  store: { get: (name: string) => Promise<TemplateBundle | null> };
  fsBaseDir: string;
  fs?: { readFile: (path: string, enc: string) => Promise<string> };
}

export interface TemplateLoader {
  load: (name: string) => Promise<TemplateBundle | null>;
}

export function createTemplateLoader(opts: TemplateLoaderOptions): TemplateLoader {
  const fs = opts.fs ?? fsPromises;

  async function readFsBundle(name: string): Promise<TemplateBundle | null> {
    const dir = join(opts.fsBaseDir, name);
    let subject: string | undefined;
    let html: string | undefined;
    let text: string | undefined;

    try {
      subject = await fs.readFile(join(dir, 'subject.hbs'), 'utf8');
    } catch {
      // ignore
    }
    try {
      html = await fs.readFile(join(dir, 'html.hbs'), 'utf8');
    } catch {
      // ignore
    }
    try {
      text = await fs.readFile(join(dir, 'text.hbs'), 'utf8');
    } catch {
      text = undefined;
    }

    if (!subject && !html && !text) return null;
    const bundle: TemplateBundle = { subject, html, text };
    return bundle;
  }

  return {
    async load(name: string): Promise<TemplateBundle | null> {
      const fromStore = await opts.store.get(name);
      const fromFs = await readFsBundle(name);
      if (!fromStore && !fromFs) return null;

      const merged: TemplateBundle = {
        subject: fromStore?.subject ?? fromFs?.subject,
        html: fromStore?.html ?? fromFs?.html,
        text: fromStore?.text ?? fromFs?.text,
      };

      validateTemplateBundle(merged, { allowMissingText: true });
      return merged;
    }
  };
}
