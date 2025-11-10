import type { TemplateBundle } from './store.js';

export interface ValidateOptions {
  allowMissingText?: boolean;
}

export function validateTemplateBundle(bundle: TemplateBundle | null | undefined, opts: ValidateOptions = {}): asserts bundle is Required<Pick<TemplateBundle, 'subject' | 'html'>> & Pick<TemplateBundle, 'text'> {
  if (!bundle) throw new Error('template not found');
  const { subject, html, text } = bundle as any;
  if (typeof subject !== 'string' || subject.trim() === '') throw new Error('invalid template: missing subject');
  if (typeof html !== 'string' || html.trim() === '') throw new Error('invalid template: missing html');
  if (!opts.allowMissingText) {
    if (text !== undefined && typeof text !== 'string') throw new Error('invalid template: text must be string when provided');
  }
}
