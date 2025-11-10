import type { FastifyInstance } from 'fastify';
import { query } from '../db';
import { getNats } from '../nats';

export function registerReadyRoute(app: FastifyInstance) {
  app.get('/readyz', async (_req, reply) => {
    const details: any = { db: { healthy: false }, nats: { healthy: false } };

    // Check DB
    try {
      const res = await query('SELECT 1 as ok');
      if (res && (res as any).rows && (res as any).rows.length > 0) {
        details.db.healthy = true;
      }
    } catch (err: any) {
      details.db.error = err?.message || String(err);
    }

    // Check NATS
    try {
      const nc: any = getNats();
      if (nc && typeof nc.flush === 'function') {
        await nc.flush();
      }
      details.nats.healthy = true;
    } catch (err: any) {
      details.nats.error = err?.message || String(err);
    }

    const healthy = details.db.healthy && details.nats.healthy;
    return reply.code(healthy ? 200 : 503).send({ ok: healthy, details });
  });
}