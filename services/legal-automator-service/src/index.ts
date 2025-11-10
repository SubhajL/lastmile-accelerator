import Fastify from 'fastify';

const app = Fastify();
app.get('/healthz', async () => 'ok');

const port = Number(process.env.SERVICE_PORT || 7503);
app.listen({ port, host: '0.0.0.0' }).then(() => {
  console.log(`service listening on :${port}`);
}).catch(err => {
  console.error(err);
  process.exit(1);
});
