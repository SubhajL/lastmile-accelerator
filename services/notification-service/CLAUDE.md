# Notification Service - Multi-Channel Notification Delivery

**Technology:** Node.js/TypeScript, Fastify
**Ports:** REST: 7902, gRPC: 50122
**Parent Context:** This extends [../../CLAUDE.md](../../CLAUDE.md) and [../CLAUDE.md](../CLAUDE.md)

## Service Purpose

The Notification Service provides multi-channel notification delivery for the LMA platform. It manages notification templates (email, SMS, Slack, webhooks), routes notifications based on user preferences and priority, tracks delivery status and retry logic for failed notifications, and provides a unified API for all platform services to send notifications. The service supports template rendering with Handlebars, rate limiting per channel, and preference management for opt-in/opt-out.

## Development Commands

### From This Directory
```bash
# Node service commands
pnpm install          # Install dependencies
pnpm dev              # Hot-reload with tsx watch
pnpm test             # Run tests with Vitest
pnpm test:coverage    # Generate coverage report
pnpm typecheck        # Type checking without building
pnpm lint             # ESLint with max-warnings=0
pnpm build            # Build for production
pnpm start            # Run production build
```

### From Root (using Turbo)
```bash
bunx turbo run dev --filter=notification-service
bunx turbo run test --filter=notification-service
bunx turbo run build --filter=notification-service
```

### Pre-PR Checklist
```bash
# Run all quality gates
pnpm typecheck && pnpm lint && pnpm test && pnpm build

# Or from root with turbo
bunx turbo run typecheck lint test build --filter=notification-service
```

## Architecture

### Directory Structure
```
notification-service/
├── src/
│   ├── bootstrap/               # Service initialization
│   │   ├── app.ts              # Fastify app factory
│   │   └── telemetry.ts        # OpenTelemetry setup
│   ├── channels/                # Notification channels
│   │   ├── email.channel.ts    # Email via Nodemailer
│   │   ├── slack.channel.ts    # Slack webhook delivery
│   │   ├── sms.channel.ts      # SMS via Twilio/SNS
│   │   └── webhook.channel.ts  # Custom webhook delivery
│   ├── consumers/               # NATS event consumers
│   │   ├── notification.consumer.ts
│   │   └── preference.consumer.ts
│   ├── db/                      # Database layer
│   │   ├── client.ts           # PostgreSQL connection
│   │   └── migrations/         # SQL migrations
│   ├── events/                  # Event publishing
│   │   └── publisher.ts
│   ├── metrics/                 # Prometheus metrics
│   │   └── collector.ts
│   ├── notifications/           # Core notification logic
│   │   ├── service.ts          # Notification orchestration
│   │   ├── router.ts           # Routing based on preferences
│   │   └── retry.ts            # Retry logic for failures
│   ├── prefs/                   # User preferences
│   │   ├── repository.ts
│   │   └── service.ts
│   ├── recipients/              # Recipient management
│   │   └── resolver.ts
│   ├── redis/                   # Redis client
│   │   └── client.ts
│   ├── routing/                 # Routing logic
│   │   └── rules.ts
│   ├── templates/               # Handlebars templates
│   │   ├── email/
│   │   │   ├── welcome.hbs
│   │   │   ├── alert.hbs
│   │   │   └── report.hbs
│   │   ├── slack/
│   │   └── sms/
│   ├── config.ts                # Configuration management
│   ├── index.ts                 # Entry point
│   └── types.ts                 # TypeScript types
├── dist/                        # Build output
├── package.json                 # Dependencies and scripts
├── tsconfig.json                # TypeScript configuration
├── vitest.config.ts             # Test configuration
├── Makefile                     # Build shortcuts
├── CONTEXT.md                   # Service metadata
└── AGENTS.md                    # AI development shortcuts
```

### Key Components

**Notification Router:**
- File: `src/notifications/router.ts` - Routes notifications to appropriate channels
- Pattern: Checks user preferences, priority, and channel availability
- Example: High-priority alerts bypass user opt-out preferences

**Channel Handlers:**
- Files: `src/channels/*.channel.ts` - Implement delivery for each channel
- Pattern: Each channel has send(), validate(), and retry() methods
- Example: `email.channel.ts` uses Nodemailer with SMTP configuration

**Template Renderer:**
- File: `src/templates/` - Handlebars templates for each notification type
- Pattern: Templates organized by channel with partials support
- Example: `email/welcome.hbs` renders personalized welcome emails

**Preference Manager:**
- File: `src/prefs/service.ts` - Manages user notification preferences
- Pattern: Opt-in/opt-out per channel, quiet hours, priority overrides
- Example: User can disable non-critical Slack notifications

**Retry Logic:**
- File: `src/notifications/retry.ts` - Exponential backoff for failed deliveries
- Pattern: Retries with increasing delays, max 5 attempts, DLQ for permanent failures
- Example: Email delivery failure retries at 1m, 5m, 15m, 1h, 4h

### Dependencies

**Core:**
- `fastify` 4.27.0 - Fast web framework for REST API
- `@fastify/cors` ^9.0.1 - CORS support
- `@fastify/helmet` ^11.1.1 - Security headers
- `@fastify/jwt` ^7.2.4 - JWT authentication
- `pg` ^8.11.3 - PostgreSQL client for notification logs
- `ioredis` ^5.3.2 - Redis for rate limiting and caching
- `nats` ^2.19.0 - NATS messaging for event-driven workflows

**Notification Channels:**
- `nodemailer` ^6.9.7 - Email delivery via SMTP
- `handlebars` ^4.7.8 - Template rendering
- `@slack/webhook` (via custom client) - Slack webhook integration
- `twilio` (future) or AWS SNS - SMS delivery

**Security & Secrets:**
- `node-vault` ^0.10.2 - HashiCorp Vault client for API keys

**Observability:**
- `@opentelemetry/api` ^1.7.0 - OpenTelemetry tracing
- `@opentelemetry/sdk-node` ^0.45.1 - OTel SDK
- `@opentelemetry/instrumentation-fastify` ^0.31.0 - Fastify tracing
- `@opentelemetry/instrumentation-http` ^0.45.1 - HTTP tracing
- `pino` (via Fastify) - Structured logging
- Prometheus metrics (via prom-client or custom)

**Testing:**
- `vitest` ^1.0.4 - Fast unit test framework
- `@vitest/coverage-v8` ^1.0.4 - Code coverage
- `@types/nodemailer` ^6.4.14 - Type definitions

## Code Organization Patterns

### Notification Routing
✅ **DO:** Use router to determine channels based on preferences
```typescript
// src/notifications/router.ts
export async function routeNotification(
  notification: Notification,
  userId: string
): Promise<Channel[]> {
  const prefs = await prefsService.getPreferences(userId);
  return selectChannels(notification, prefs);
}
```
❌ **DON'T:** Hardcode channel selection in business logic

### Channel Implementation
✅ **DO:** Implement consistent interface for all channels
```typescript
// src/channels/email.channel.ts
export class EmailChannel implements NotificationChannel {
  async send(recipient: string, template: string, data: any): Promise<void> {
    const html = renderTemplate(template, data);
    await this.mailer.sendMail({ to: recipient, html });
  }
}
```
❌ **DON'T:** Skip error handling or retry logic in channels

### Template Rendering
✅ **DO:** Use Handlebars with partials for reusability
```typescript
// src/templates/renderer.ts
import Handlebars from 'handlebars';
export function renderTemplate(name: string, data: any): string {
  const template = Handlebars.compile(templates[name]);
  return template(data);
}
```
❌ **DON'T:** Inline HTML in code; always use templates

### Error Handling
✅ **DO:** Track delivery failures and route to DLQ
```typescript
// src/notifications/retry.ts
try {
  await channel.send(notification);
} catch (error) {
  await scheduleRetry(notification, attempt + 1);
  if (attempt >= MAX_RETRIES) {
    await sendToDLQ(notification, error);
  }
}
```
❌ **DON'T:** Silently drop failed notifications

## API Endpoints

### REST API

**Base URL:** `http://localhost:7902`

**Key Endpoints:**
- `GET /healthz` - Health check with dependency status
- `GET /metrics` - Prometheus metrics
- `POST /api/v1/notifications` - Send notification via API
- `GET /api/v1/notifications/{id}` - Get notification delivery status
- `POST /api/v1/notifications/bulk` - Send batch notifications
- `GET /api/v1/preferences/{userId}` - Get user notification preferences
- `PUT /api/v1/preferences/{userId}` - Update notification preferences
- `POST /api/v1/templates` - Create/update notification template
- `GET /api/v1/templates/{id}` - Get template by ID
- `GET /api/v1/channels/{channel}/status` - Check channel health

### gRPC API

**Port:** 50122

**Services:**
- `NotificationService` - Notification delivery and management
  - `SendNotification` - Send single notification
  - `SendBulk` - Send batch notifications
  - `GetDeliveryStatus` - Check delivery status
  - `UpdatePreferences` - Manage user preferences

**Proto Files:** `api/*.proto` (if present)

## Database Schema

**Tables:**

**`notifications`** - Notification delivery log
- Columns: `id`, `user_id`, `type`, `channel`, `status`, `content`, `sent_at`, `delivered_at`, `error`
- Indexes: `idx_user_id`, `idx_status`, `idx_sent_at`
- Purpose: Track all notifications sent and their delivery status

**`notification_preferences`** - User preferences
- Columns: `user_id`, `channel`, `enabled`, `quiet_hours_start`, `quiet_hours_end`, `priority_threshold`
- Indexes: `idx_user_id`, `idx_channel`
- Purpose: Store per-user, per-channel notification settings

**`templates`** - Notification templates
- Columns: `id`, `name`, `channel`, `subject`, `body`, `version`, `created_at`, `updated_at`
- Indexes: `idx_name`, `idx_channel`
- Purpose: Versioned templates for consistent messaging

**`delivery_failures`** - Failed delivery attempts
- Columns: `id`, `notification_id`, `channel`, `attempt`, `error`, `retry_at`, `created_at`
- Indexes: `idx_notification_id`, `idx_retry_at`
- Purpose: Track failures for retry and monitoring

**Migrations:**
- Location: `src/db/migrations/`
- Tool: Custom migration scripts
- Commands: `pnpm db:migrate:up`, `pnpm db:migrate:down`

## Event Handling

**Published Events:**
- `notification.sent` - When notification successfully delivered
  - Payload: `{notification_id, user_id, channel, sent_at}`
- `notification.failed` - When delivery fails after all retries
  - Payload: `{notification_id, user_id, channel, error, attempt}`
- `preference.updated` - When user updates notification preferences
  - Payload: `{user_id, channel, enabled, updated_at}`

**Subscribed Events:**
- `user.created` - Initialize default notification preferences
- `project.alert` - Send high-priority project alerts
- `test-run.failed` - Notify on test failures
- `migration.failed` - Alert on database migration issues
- `vulnerability.detected` - Security alert notifications

## Testing Strategy

### Unit Tests
- Location: `src/**/__tests__/*.test.ts`
- Coverage: Target >80%
- Mock: Email/SMS/Slack clients, database, Redis
- Example: Test template rendering, routing logic, retry schedules

### Integration Tests
- Location: `src/__tests__/integration/`
- Setup: Use test SMTP server (Ethereal), mock Slack webhooks
- Pattern: Test full notification flow from API -> channel delivery

### Running Tests
```bash
# All tests
pnpm test

# Coverage report
pnpm test:coverage

# Specific test file
pnpm test notification.service.test.ts
```

## Configuration

### Environment Variables
```bash
# Service-specific config
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=notifications@example.com
SMTP_FROM=LMA Notifications <noreply@example.com>
SMTP_SECURE=true

SLACK_WEBHOOK_URL=https://hooks.slack.com/services/xxx
TWILIO_ACCOUNT_SID=ACxxx
TWILIO_AUTH_TOKEN=xxx
SMS_FROM=+1234567890

MAX_RETRY_ATTEMPTS=5
RETRY_BACKOFF_MS=60000  # 1 minute initial
RATE_LIMIT_PER_MINUTE=100
ENABLE_QUIET_HOURS=true

# Standard config (inherited from ../CLAUDE.md)
SERVICE_NAME=notification-service
SERVICE_PORT=7902
GRPC_PORT=50122
ENV=dev
DATABASE_URL=postgresql://user:pass@localhost:5432/notifications
REDIS_URL=redis://localhost:6379
NATS_URL=nats://localhost:4222
VAULT_ADDR=http://vault:8200
OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4318
LOG_LEVEL=info
```

### Secrets
- Stored in: Vault at `secret/notification-service/`
- Accessed via: `node-vault` SDK or secrets-env-service
- Keys: SMTP credentials, Twilio API keys, Slack webhook URLs

## Quick Find Commands

### Find Code
```bash
# Find notification sending logic
rg -n "sendNotification|send\(" services/notification-service/src/

# Find channel implementations
rg -n "class.*Channel|export.*Channel" services/notification-service/src/channels/

# Find template rendering
rg -n "Handlebars|renderTemplate" services/notification-service/src/

# Find event subscribers
rg -n "subscribe.*\." services/notification-service/src/consumers/

# Find preference management
rg -n "getPreferences|updatePreferences" services/notification-service/src/prefs/
```

### Find Dependencies
```bash
# Check what depends on this service
rg -n "notification-service" --glob "docker-compose*.yml" --glob "*.yaml"

# Find all notification triggers
rg -n "publish.*notification\." services/
```

## Common Gotchas

- **SMTP Authentication:** Many providers require app-specific passwords; regular passwords may fail; use OAuth2 for Gmail
- **Rate Limiting:** Email providers have strict rate limits; implement exponential backoff and respect limits to avoid IP blacklisting
- **Quiet Hours Timezone:** User quiet hours must respect user timezone; store timezone in preferences or infer from user profile
- **Template Caching:** Template compilation is expensive; cache compiled Handlebars templates in memory or Redis
- **Webhook Failures:** External webhooks may timeout or fail silently; implement timeout and retry logic
- **Large Attachments:** Email attachments increase delivery time and failure rate; use links to cloud storage instead
- **HTML Email Rendering:** Email clients have inconsistent HTML support; test templates across major clients (Gmail, Outlook, Apple Mail)

## Related Services

- **projects-service:** Sends project-related notifications (deployment success/failure, alerts)
- **test-lab-service:** Triggers notifications on test run completion
- **db-guardian-service:** Alerts on migration validation failures
- **dep-governance-service:** Sends security alerts for detected vulnerabilities
- **observability-service:** Aggregates notification metrics and delivery rates

## Useful Links

- Service Catalog: `../../service_catalog.yaml`
- CONTEXT.md: Service purpose, SLOs, ownership details
- AGENTS.md: AI-assisted development shortcuts
- CI Workflow: `../../.github/workflows/ci-notification-service.yml`
- Email Template Guide: `docs/template-guide.md` (if exists)
- Nodemailer Docs: https://nodemailer.com/
