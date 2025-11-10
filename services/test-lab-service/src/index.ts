import { createApp } from './app.js';

const port = Number(process.env.SERVICE_PORT || 7202);

(async () => {
  try {
    const app = await createApp();
    await app.listen({ port, host: '0.0.0.0' });
    // eslint-disable-next-line no-console
    console.log(`service listening on :${port}`);
  } catch (err) {
    // eslint-disable-next-line no-console
    console.error(err);
    process.exit(1);
  }
})();
