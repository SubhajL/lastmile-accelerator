import { isAppError } from '../utils/errors';
import { getLogger } from '../logger';

function safeLogger() {
  try {
    return getLogger();
  } catch {
    return { error: (...args: any[]) => { try { console.error(...args); } catch {} } } as any;
  }
}

export async function errorHandler(error: any, request: any, reply: any) {
  const logger = safeLogger();

  if (isAppError(error)) {
    logger.error({ err: error, reqId: request?.id }, 'handled app error');
    return reply
      .status(error.statusCode)
      .send({ code: error.code, message: error.message, details: error.details });
  }

  logger.error({ err: error, reqId: request?.id }, 'unhandled error');
  return reply
    .status(500)
    .send({ code: 'INTERNAL_SERVER_ERROR', message: error?.message ?? 'Internal Server Error' });
}