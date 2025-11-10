import { promises as fs } from 'fs';
import { join } from 'path';
import Handlebars from 'handlebars';
import type { TemplateBundle } from '../templates/store.js';

export interface HandlebarsRendererOptions {
  templatesDir: string;
  loader?: { load: (name: string) => Promise<TemplateBundle | null> };
}

export function createHandlebarsRenderer(opts: HandlebarsRendererOptions) {
  return async function render(templateName: string, payload: Record<string, unknown>) {
    let subjectT: string | undefined;
    let htmlT: string | undefined;
    let textT: string | undefined;

    if (opts.loader) {
      const bundle = await opts.loader.load(templateName);
      if (bundle) {
        subjectT = bundle.subject;
        htmlT = bundle.html;
        textT = bundle.text;
      }
    }

    if (!subjectT || !htmlT) {
      const dir = join(opts.templatesDir, templateName);
      const [s, h, t] = await Promise.all([
        subjectT ? Promise.resolve(subjectT) : fs.readFile(join(dir, 'subject.hbs'), 'utf8'),
        htmlT ? Promise.resolve(htmlT) : fs.readFile(join(dir, 'html.hbs'), 'utf8'),
        textT ? Promise.resolve(textT) : fs.readFile(join(dir, 'text.hbs'), 'utf8').catch(() => undefined)
      ]);
      subjectT = s;
      htmlT = h;
      textT = t;
    }

    const subject = Handlebars.compile(subjectT!)(payload);
    const html = Handlebars.compile(htmlT!)(payload);
    const text = textT ? Handlebars.compile(textT)(payload) : html.replace(/<[^>]+>/g, '').trim();

    return { subject, html, text };
  };
}
