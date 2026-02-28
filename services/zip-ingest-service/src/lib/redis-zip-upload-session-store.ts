import { createRedisClient } from '../clients/redis.js';
import type { RedisClientLogger, RedisWrapper } from '../clients/redis.js';

import type { ZipUploadSession, ZipUploadSessionStore } from './zip-upload-session-store.js';

function keyForUploadId(args: { keyPrefix: string; uploadId: string }): string {
  return `${args.keyPrefix}${args.uploadId}`;
}

function parseZipUploadSession(value: unknown): ZipUploadSession | null {
  if (typeof value !== 'object' || value === null) return null;
  const v = value as Record<string, unknown>;

  const projectId = v.projectId;
  const bucket = v.bucket;
  const objectKey = v.objectKey;
  const expiresAtUnixSeconds = v.expiresAtUnixSeconds;

  if (typeof projectId !== 'string' || projectId.length < 1) return null;
  if (typeof bucket !== 'string' || bucket.length < 1) return null;
  if (typeof objectKey !== 'string' || objectKey.length < 1) return null;
  if (typeof expiresAtUnixSeconds !== 'number' || !Number.isFinite(expiresAtUnixSeconds)) return null;

  return { projectId, bucket, objectKey, expiresAtUnixSeconds };
}

export function createRedisZipUploadSessionStore(args: {
  client: RedisWrapper;
  keyPrefix: string;
}): ZipUploadSessionStore {
  return {
    async putSession({ uploadId, session, ttlSeconds }) {
      await args.client.set(
        keyForUploadId({ keyPrefix: args.keyPrefix, uploadId }),
        JSON.stringify(session),
        ttlSeconds,
      );
    },
    async getSession(uploadId) {
      const key = keyForUploadId({ keyPrefix: args.keyPrefix, uploadId });
      const raw = await args.client.get(key);
      if (!raw) return null;
      try {
        const parsed = JSON.parse(raw) as unknown;
        const session = parseZipUploadSession(parsed);
        if (!session) {
          await args.client.del(key);
          return null;
        }
        return session;
      } catch {
        await args.client.del(key);
        return null;
      }
    },
    async deleteSession(uploadId) {
      await args.client.del(keyForUploadId({ keyPrefix: args.keyPrefix, uploadId }));
    },
    async close() {
      await args.client.close();
    },
  };
}

export async function createRedisZipUploadSessionStoreFromUrl(args: {
  redisUrl: string;
  keyPrefix: string;
  logger?: RedisClientLogger;
}): Promise<ZipUploadSessionStore> {
  const client = await createRedisClient(args.redisUrl, { logger: args.logger });
  return createRedisZipUploadSessionStore({ client, keyPrefix: args.keyPrefix });
}
