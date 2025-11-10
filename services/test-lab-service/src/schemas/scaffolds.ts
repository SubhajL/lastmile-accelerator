import { z } from 'zod';

export const ScaffoldType = z.enum(['unit', 'e2e', 'smoke']);
export const Framework = z.enum(['vitest', 'jest', 'playwright', 'cypress', 'pytest', 'go-test']);
export const Language = z.enum(['ts', 'js', 'py', 'go']);

export const CreateScaffoldSchema = z.object({
  type: ScaffoldType,
  framework: Framework,
  language: Language,
  config: z.record(z.any()).default({}),
});

export const UpdateScaffoldSchema = z.object({
  framework: Framework.optional(),
  language: Language.optional(),
  config: z.record(z.any()).optional(),
});

export const ListScaffoldsQuerySchema = z.object({
  cursor: z.string().optional(),
  limit: z
    .string()
    .transform((v) => Number(v))
    .pipe(z.number().int().min(1).max(500))
    .optional()
    .default('50' as any),
});

export type CreateScaffoldInput = z.infer<typeof CreateScaffoldSchema>;
export type UpdateScaffoldInput = z.infer<typeof UpdateScaffoldSchema>;
export type ListScaffoldsQuery = z.infer<typeof ListScaffoldsQuerySchema>;
