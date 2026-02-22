import { loadConfig } from './config.js';
import { initTelemetry, shutdownTelemetry } from './telemetry.js';
import { createApp } from './app.js';
import { createDbClient } from './db/client.js';
import { bootstrap as startWorker } from './worker.js';

async function main() {
  const config = loadConfig();

  if (config.worker.enabled) {
    console.log(`Starting ${config.service.name} in WORKER mode (${config.env} environment)`);

    const sdk = initTelemetry(config.observability);
    const worker = await startWorker();

    console.log(`Subscribed to ${config.worker.natsSubjects.length} NATS subjects`);

    // Graceful shutdown
    const shutdown = async () => {
      console.log('SIGTERM received, stopping worker...');
      await worker.stop();
      await shutdownTelemetry(sdk);
      console.log('Worker stopped gracefully');
      process.exit(0);
    };

    process.on('SIGTERM', shutdown);
    process.on('SIGINT', shutdown);
  } else {
    console.log(`Starting ${config.service.name} in HTTP mode (${config.env} environment)`);

    const sdk = initTelemetry(config.observability);
    const dbPool = createDbClient(config.postgres);
    const app = await createApp(config, dbPool);

    const port = config.service.port;
    await app.listen({ port, host: '0.0.0.0' });
    console.log(`${config.service.name} listening on :${port}`);

    // Graceful shutdown
    const shutdown = async () => {
      console.log('Shutting down gracefully...');
      await app.close();
      await dbPool.end();
      await shutdownTelemetry(sdk);
      process.exit(0);
    };

    process.on('SIGTERM', shutdown);
    process.on('SIGINT', shutdown);
  }
}

main().catch((err) => {
  console.error('Failed to start service:', err);
  process.exit(1);
});
