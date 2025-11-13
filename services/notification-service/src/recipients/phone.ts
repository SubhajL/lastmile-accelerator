type JobLike = { payload?: { phone?: string; phones?: Array<string | number> }; userId?: string };

export async function getRecipientPhoneFromJob(job: JobLike): Promise<string> {
  const p = (job?.payload ?? {});
  if (typeof p.phone === 'string') return p.phone;
  if (Array.isArray(p.phones) && p.phones.length > 0) return String(p.phones[0]);
  if (typeof job.userId === 'string' && /^\+?[0-9]{7,15}$/.test(job.userId)) return job.userId;
  throw new Error('No recipient phone found');
}
