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
  const configuredSecret = authConfig.jwtPublicKey || process.env.JWT_SECRET;
  if (!configuredSecret) {
    if (process.env.NODE_ENV === 'production') {
      throw new Error('JWT secret/public key is required in production');
    }
  }

  // Development fallback to make local tests and dev servers easy to run.
  const secret = configuredSecret ?? 'development-secret';

  app.register(jwtPlugin, { secret });
}
