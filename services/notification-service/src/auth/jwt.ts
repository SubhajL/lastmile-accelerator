import fp from 'fastify-plugin';
import type { FastifyInstance, FastifyReply, FastifyRequest } from 'fastify';
import fastifyJwt from '@fastify/jwt';

type JwtOpts = { secret: string };

declare module 'fastify' {
  interface FastifyRequest {
    user?: { roles?: string[]; [k: string]: unknown };
  }
}

export const jwtPlugin = fp(async function (app: FastifyInstance, opts: JwtOpts) {
  await app.register(fastifyJwt, { secret: opts.secret });
});

declare module 'fastify' {
  interface FastifyInstance {}
}

export function requireRoleGuard(role: string) {
  return async function (req: FastifyRequest, reply: FastifyReply) {
    try {
      const decoded = await req.jwtVerify<{ roles?: string[] }>();
      req.user = decoded as { roles?: string[] };
      const roles = Array.isArray(decoded?.roles) ? decoded!.roles! : [];
      if (!roles.includes(role)) {
        return reply.code(403).send({ error: 'forbidden' });
      }
    } catch {
      return reply.code(401).send({ error: 'unauthorized' });
    }
  };
}
