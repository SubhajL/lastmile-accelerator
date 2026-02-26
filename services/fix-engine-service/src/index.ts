import { createApp } from './app.js';

const port = Number(process.env.SERVICE_PORT || 7151);

createApp()
  .then((app) => app.listen({ port, host: '0.0.0.0' }))
  .then(() => {
    console.log(`service listening on :${port}`);
  })
  .catch((err) => {
    console.error(err);
    process.exit(1);
  });
