import { randomUUID } from 'crypto';

export function requestIdMiddleware() {
  return async function (request: any, reply: any) {
    const incoming = request?.headers?.['x-request-id'];
    const id = typeof incoming === 'string' && incoming.length > 0 ? incoming : randomUUID();
    request.id = id;
    if (typeof reply?.header === 'function') {
      reply.header('x-request-id', id);
    }
  };
}