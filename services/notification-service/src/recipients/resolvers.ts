import type { NotificationJob } from '../consumers/types.js';

export async function getRecipientEmailFromJob(job: NotificationJob): Promise<string> {
  const p = (job.payload || {}) as any;
  if (typeof p.email === 'string') return p.email;
  if (Array.isArray(p.emails) && p.emails.length > 0) return String(p.emails[0]);
  if (typeof job.userId === 'string' && /@/.test(job.userId)) return job.userId;
  throw new Error('No recipient email found');
}

export function createDefaultRecipientsResolver() {
  return async (evt: any): Promise<string[]> => {
    if (Array.isArray(evt?.recipients) && evt.recipients.length > 0) return evt.recipients as string[];
    if (typeof evt?.email === 'string') return [evt.email as string];
    if (Array.isArray(evt?.emails) && evt.emails.length > 0) return evt.emails as string[];
    return [];
  };
}
