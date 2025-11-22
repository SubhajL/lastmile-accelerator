import fp from 'fastify-plugin';
import type { FastifyInstance, FastifyReply, FastifyRequest } from 'fastify';
import fastifyJwt from '@fastify/jwt';

type JwtOpts = { secret: string };

declare module '@fastify/jwt' {
  interface FastifyJWT {
    payload: { roles?: string[]; [k: string]: unknown };
    user: { roles?: string[]; [k: string]: unknown };
  }
}

export const jwtPlugin = fp(async function (app: FastifyInstance, opts: JwtOpts) {
  await app.register(fastifyJwt, { secret: opts.secret });
});

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

export function buildJwtAuth(
  app: FastifyInstance,
  authConfig: { jwtPublicKey?: string; jwtJwksUrl?: string },
): void {
  // Use public key if provided, otherwise use a default secret for development
  const secret = authConfig.jwtPublicKey || process.env.JWT_SECRET || 'development-secret';

  app.register(jwtPlugin, { secret });
}
