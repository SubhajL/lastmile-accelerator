import { z } from 'zod';

export const PreviewStatus = z.enum(['creating', 'ready', 'failed', 'deleted']);

export const CreatePreviewSchema = z.object({
  url: z.string().url(),
  snapshotId: z.string().uuid().optional(),
});

export const ListPreviewsQuerySchema = z.object({
  cursor: z.string().optional(),
  limit: z.string().transform(Number).pipe(z.number().int().min(1).max(500)).optional().default('50' as any),
});

export const UpdatePreviewSchema = z.object({
  url: z.string().url().optional(),
  status: PreviewStatus.optional(),
  expiresAt: z.coerce.date().optional(),
});
