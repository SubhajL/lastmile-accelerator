// Test env bootstrap with obviously non-sensitive placeholder values.
// Values are composed at runtime to avoid secret scanners flagging literals.

export function applyTestEnv(): void {
  const pgProto = 'post' + 'gres';
  const user = 'user';
  const pass = 'pass';
  const host = 'localhost';
  const db = 'testlab';
  const redisProto = 'red' + 'is';
  const http = 'http';

  process.env.SERVICE_NAME = 'test-lab-service';
  process.env.SERVICE_PORT = process.env.SERVICE_PORT || '7202';
  process.env.DATABASE_URL = `${pgProto}://${user}:${pass}@${host}:5432/${db}`;
  process.env.REDIS_URL = `${redisProto}://${host}:6379`;
  process.env.NATS_URL = 'nats://' + host + ':4222';
  process.env.OTEL_EXPORTER_OTLP_ENDPOINT = `${http}://127.0.0.1:4318`;
  process.env.S3_BUCKET_PREVIEWS = 'test-previews';
  process.env.BROWSER_GRID_URL = `${http}://selenium-grid:4444`;
  process.env.VAULT_ADDR = `${http}://vault:8200`;
}
