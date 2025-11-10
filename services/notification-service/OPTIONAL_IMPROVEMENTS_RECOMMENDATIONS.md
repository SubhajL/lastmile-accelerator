# Notification Service - Optional Improvements Recommendations

Based on current implementation status (Tasks 1-4 completed).

---

## Current State

The notification-service has:
- ✅ Complete configuration management with SMTP settings
- ✅ Database schema with `notification_logs`, `delivery_attempts`
- ✅ OpenTelemetry integration with metrics endpoint
- ✅ Worker bootstrap with placeholder email transporter
- ✅ nodemailer dependency installed

**Placeholders in `worker.ts`:**
```typescript
transporter: { sendMail: async () => ({}) } as any,  // Line 28, 46
resolveTo: async (job: any) => job.userId             // Line 31, 48
metrics: { increment: () => {} }                       // Line 40
```

---

## Recommendations by Priority

### 🟢 **Priority 1: Wire Real Nodemailer Transporter** 
**Recommendation:** ✅ **Do it now**

**Why:**
- SMTP config is already complete in `config.ts` (lines 130-146)
- nodemailer is already installed
- MailHog is running in devstack (`lma-mailhog` on port 1030/8030)
- **Without this, no emails will actually send** - it's currently a no-op

**Implementation Effort:** 🟢 **15-20 minutes**

**What to do:**
```typescript
// src/channels/email-transporter.ts
import nodemailer from 'nodemailer';
import type { SmtpConfig } from '../types.js';

export function createEmailTransporter(config: SmtpConfig) {
  return nodemailer.createTransport({
    host: config.host,
    port: config.port,
    secure: config.secure,
    auth: config.user && config.password ? {
      user: config.user,
      pass: config.password
    } : undefined
  });
}
```

Then update `worker.ts`:
```typescript
import { createEmailTransporter } from './channels/email-transporter.js';

const transporter = createEmailTransporter(cfg.smtp);

// Replace placeholders on lines 28 and 46 with real transporter
```

**Test with devstack:**
```bash
# .env.local already has correct values
SMTP_HOST=localhost
SMTP_PORT=1025
SMTP_FROM=dev@lma.local
# User/password can be empty for MailHog

# View emails at http://localhost:8030
```

---

### 🟡 **Priority 2: Real User Resolver**
**Recommendation:** ⚠️ **Depends on user service integration**

**Why:**
- Currently returns `job.userId` which is just a string ID
- Need to resolve to actual email address
- Requires integration with user/identity service (doesn't exist yet?)

**Block Factors:**
1. No user service in the monorepo yet
2. User data schema not defined
3. Unknown if users are in Keycloak, Postgres, or external system

**Implementation Options:**

#### Option A: Database Lookup (if user emails in Postgres)
```typescript
// src/users/resolver.ts
export function createUserResolver(db: Pool) {
  return async (userId: string): Promise<string> => {
    const result = await db.query(
      'SELECT email FROM users WHERE id = $1',
      [userId]
    );
    if (!result.rows[0]) {
      throw new Error(`User not found: ${userId}`);
    }
    return result.rows[0].email;
  };
}
```

#### Option B: Keycloak API
```typescript
// src/users/keycloak-resolver.ts
export function createKeycloakUserResolver(keycloakUrl: string, token: string) {
  return async (userId: string): Promise<string> => {
    const resp = await fetch(
      `${keycloakUrl}/admin/realms/lma/users/${userId}`,
      { headers: { Authorization: `Bearer ${token}` }}
    );
    const user = await resp.json();
    return user.email;
  };
}
```

#### Option C: User Service HTTP Call
```typescript
// src/users/service-resolver.ts
export function createUserServiceResolver(userServiceUrl: string) {
  return async (userId: string): Promise<string> => {
    const resp = await fetch(`${userServiceUrl}/api/v1/users/${userId}`);
    const user = await resp.json();
    return user.email;
  };
}
```

**Current Recommendation:** 🟡 **Wait until user management is clarified**

For now, you can:
1. **Accept email directly in notification jobs** instead of userId
2. Add a TODO comment to implement proper resolution later
3. Add validation to ensure `job.userId` is an email address format

```typescript
// Temporary workaround
resolveTo: async (job: any) => {
  // TODO: Resolve userId to email via user service/Keycloak
  // For now, accept email directly in job payload
  if (job.recipientEmail) return job.recipientEmail;
  if (job.userId.includes('@')) return job.userId; // Assume email format
  throw new Error('Cannot resolve user email - user service not integrated');
}
```

---

### 🟢 **Priority 3: Telemetry Metrics**
**Recommendation:** ✅ **Do it now** (after nodemailer transporter)

**Why:**
- Metrics endpoint already exists at `/metrics`
- OTel is fully configured
- Essential for production monitoring
- Low implementation effort

**Implementation Effort:** 🟢 **20-30 minutes**

**What to do:**
```typescript
// src/metrics/counters.ts
import { metrics } from '@opentelemetry/api';

const meter = metrics.getMeter('notification-service');

export const notificationCounters = {
  sent: meter.createCounter('notifications_sent_total', {
    description: 'Total notifications successfully sent',
  }),
  
  failed: meter.createCounter('notifications_failed_total', {
    description: 'Total notifications that failed delivery',
  }),
  
  dlq: meter.createCounter('notifications_dlq_total', {
    description: 'Total notifications moved to dead letter queue',
  }),
  
  retries: meter.createCounter('notifications_retries_total', {
    description: 'Total notification retry attempts',
  }),
};

// Helper functions
export function incrementSent(channel: string, userId?: string) {
  notificationCounters.sent.add(1, { channel, user_id: userId || 'unknown' });
}

export function incrementFailed(channel: string, reason: string) {
  notificationCounters.failed.add(1, { channel, reason });
}

export function incrementDLQ(channel: string) {
  notificationCounters.dlq.add(1, { channel });
}
```

Update `worker.ts`:
```typescript
import { incrementSent, incrementFailed, incrementDLQ } from './metrics/counters.js';

const runtime = createRuntime({
  // ... other config
  metrics: {
    increment: (metric: string, labels?: Record<string, string>) => {
      if (metric === 'notify_sent') incrementSent(labels?.channel || 'email', labels?.user_id);
      if (metric === 'notify_failed') incrementFailed(labels?.channel || 'email', labels?.reason || 'unknown');
      if (metric === 'notify_dlq') incrementDLQ(labels?.channel || 'email');
    }
  }
});
```

**Verify metrics:**
```bash
curl http://localhost:7902/metrics | grep notifications_
```

---

### 🔵 **Priority 4: Worker Health Endpoint**
**Recommendation:** ✅ **Nice to have, but not urgent**

**Why:**
- Main app already has `/healthz` endpoint
- Worker runs in same process (see `worker.ts` bootstrap)
- Kubernetes liveness/readiness probes can use main app endpoint

**Implementation Effort:** 🟢 **10 minutes**

**What to do (if desired):**
```typescript
// src/app.ts - Add worker health check
app.get('/healthz/worker', async (request, reply) => {
  const checks = {
    redis: await checkRedisConnection(redis),
    nats: await checkNatsConnection(nc),
    queue: await checkQueueHealth(queue),
  };
  
  const healthy = Object.values(checks).every(v => v === true);
  
  return reply.status(healthy ? 200 : 503).send({
    status: healthy ? 'healthy' : 'unhealthy',
    checks,
    timestamp: new Date().toISOString()
  });
});
```

**Current Recommendation:** 🔵 **Skip for now**
- Existing `/healthz` checks database (most critical dependency)
- NATS/Redis failures will show up in metrics/logs anyway

---

## Implementation Order

### Immediate (Now)
1. ✅ **Wire nodemailer transporter** (15-20 min)
   - Creates actual email functionality
   - Can test with MailHog immediately
   - Blocks any real notification testing

2. ✅ **Add telemetry metrics** (20-30 min)
   - Essential for production monitoring
   - Works alongside nodemailer implementation

### Near-Term (Next Sprint)
3. 🟡 **Resolve user email strategy**
   - Requires architectural decision
   - Depends on user service/Keycloak integration
   - Use workaround (accept email in job) until then

### Optional (Lower Priority)
4. 🔵 **Worker health endpoint** - Skip unless specifically needed

---

## Quick Win: Complete Implementation

Here's a 30-minute implementation that covers priorities 1 & 3:

```bash
cd services/notification-service

# 1. Create email transporter
cat > src/channels/email-transporter.ts << 'EOF'
import nodemailer from 'nodemailer';
import type { SmtpConfig } from '../types.js';

export function createEmailTransporter(config: SmtpConfig) {
  return nodemailer.createTransport({
    host: config.host,
    port: config.port,
    secure: config.secure,
    auth: config.user && config.password ? {
      user: config.user,
      pass: config.password
    } : undefined
  });
}
EOF

# 2. Create metrics
cat > src/metrics/counters.ts << 'EOF'
import { metrics } from '@opentelemetry/api';

const meter = metrics.getMeter('notification-service');

export const notificationCounters = {
  sent: meter.createCounter('notifications_sent_total'),
  failed: meter.createCounter('notifications_failed_total'),
  dlq: meter.createCounter('notifications_dlq_total'),
};

export function incrementSent(channel: string) {
  notificationCounters.sent.add(1, { channel });
}

export function incrementFailed(channel: string, reason: string) {
  notificationCounters.failed.add(1, { channel, reason });
}

export function incrementDLQ(channel: string) {
  notificationCounters.dlq.add(1, { channel });
}
EOF

# 3. Update worker.ts (manual edit needed)
# Replace placeholder transporter and metrics
```

---

## Testing Plan

After implementing:

```bash
# 1. Start devstack (already running)
docker ps | grep mailhog

# 2. Set env vars
export SERVICE_NAME=notification-service
export SERVICE_PORT=7902
export SMTP_HOST=localhost
export SMTP_PORT=1025
export SMTP_FROM=dev@lma.local
export SMTP_USER=
export SMTP_PASSWORD=

# 3. Start service
pnpm dev

# 4. Check metrics endpoint
curl http://localhost:7902/metrics | grep notifications_

# 5. Trigger test notification (via NATS or API)
# (Implementation depends on your notification job structure)

# 6. Check MailHog UI
open http://localhost:8030
```

---

## Summary

| Item | Priority | Effort | Do Now? | Blocker |
|------|----------|--------|---------|---------|
| Nodemailer transporter | 🟢 High | 15-20 min | ✅ Yes | None |
| Telemetry metrics | 🟢 High | 20-30 min | ✅ Yes | None |
| User email resolver | 🟡 Medium | 30+ min | ⚠️ Wait | Need user service architecture |
| Worker health endpoint | 🔵 Low | 10 min | ❌ No | None |

**Total time for immediate wins:** ~40 minutes

**Recommendation:** Implement nodemailer + metrics now, defer user resolver until user management architecture is defined.
