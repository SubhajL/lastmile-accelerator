Absolutely—here’s the **fully updated PRD** with “fix vs. suggest” behavior built-in, repo-ingestion modes first-class, and all changes threaded through architecture, APIs, data, events, SLOs, milestones, and acceptance tests. I’ve preserved your structure and wording while inserting the new pieces.

---

# **0\) Product Name & One-liner**

**Name:** Last-Mile Accelerator (LMA)

**One-liner:** “Ship-ready guardrails for vibe coders—security, tests, telemetry, and deploys that just work.”

---

# **1\) Goals, Non-Goals, Personas**

## **1.1 Goals (v1)**

* Reduce **time-to-first-paid-user** and **mean-time-to-rollback** with opinionated gates.

* Make shipping **safe by default**: security, tests, rate limits, telemetry, legal.

* Keep work **parallelizable**: microservice architecture with stable contracts.

* **NEW:** Make **repo ingestion** frictionless with three customer modes (no jargon):

  * **Connect via Git Provider (Recommended)** — Click-to-Connect

  * **Upload a ZIP (No Git yet)** — Drop-in Zip

  * **Run Checks Inside Your Account (Privacy Mode)** — Agent-in-CI

* **NEW:** Offer **Fix List \+ Auto-Fix** (with **Publish Safely** & one-click **Undo**) so customers can choose **Apply Fixes** or **Keep as Suggestion**.

## **1.2 Non-Goals (v1)**

* Not a general cloud provider.

* Not replacing full SOC/SIEM; we integrate.

* Not building a full marketing suite—only launch essentials.

## **1.3 Personas**

* **Vibe Coder / Solo Founder:** one-click guardrails and checklists.

* **Indie Team Lead:** policy gates, rollbacks, observability.

* **Advisor/Mentor:** readiness scores and failing checks.

**Customer/contract nouns used throughout:** **Customer**, **Customer Code**, **Customer Git Provider Account**, **Safety Checks**, **Fix List**, **Publish Safely**, **Undo**.

---

# **2\) Key Capabilities (mapped to the 20 obstacles)**

* Spec Memory & ADRs (drift control).

* AI Debugger & Trace Explainer (last-20% root cause).

* Secure-by-Default Scaffolds (headers, CSRF, validation).

* Secrets & Env Manager (server-only, drift diff).

* DB Hardening & Index Advisor (least-privilege, migrations, indexes).

* Dependency Governance (SBOM/CVE/license policies).

* Guardrailed CI/CD \+ Instant Rollback (pre-flight, canary, notes).

* Test Harness Factory \+ Cross-Browser Lab (unit/e2e/smoke; Safari/Firefox/iOS/Android).

* AuthZ Matrix & IDOR Scanner (role designer \+ contract tests).

* API Gateway: Rate-Limit/WAF (per-route policies).

* Observability in a Box (OTel traces/logs/metrics, SLOs, alerts).

* Perf & Caching Coach (N+1 detector, cache recipes, queues).

* FinOps Guardrails (cost dashboards, anomalies, forecasts).

* MVP Validator & Scope Freeze (MVV→MVP guard).

* Positioning & Outcome Copy (ICP fit, outcome-first copy).

* Launch Engine (channels, UTM sanity).

* Legal & Policy Automator (Privacy/ToS/cookie \+ consent logs).

* Motivation & Focus Engine (anti-shiny nudges, pods).

* Readiness Scorecard (Technical/Product/Business/Founder Ops).

* Policy-as-Code Publish Gates (merge/publish blockers).

* **NEW:** **Project Sources & Snapshot Ingestion Layer** (unifies all modes into canonical **Snapshot**).

* **NEW:** **Fix List & Patch Engine** (Tiered auto-repair vs. suggestions; **Publish Safely/Undo**).

---

# **3\) System Architecture (microservices)**

## **3.1 Platform Topology**

Monorepo (Turborepo), Envoy Gateway ingress (80/443), Istio mesh (mTLS), NATS JetStream (+Kafka optional), Postgres/Redis/S3/ClickHouse/OpenSearch, OTel→Tempo/Jaeger/Prom/Loki/Alertmanager, OIDC (Keycloak).

External REST+Webhooks, internal gRPC (HTTP/2). Ports: REST **7xxx**, gRPC **50xxx**. Edge WAF+RLS at Envoy.

## **3.2 Microservices (buildable in parallel)**

**Legend:** Stack \= (Lang/Framework \+ DB) • Ports \= REST **7xxx** / gRPC **50xxx** • Secrets via Vault.

### **\[A\] Edge & Identity**

* **api-gateway** — Edge routing, JWT verify, WAF, RLS, TLS. (80/443, gRPC mgmt 50050\)

   ENV: GATEWAY\_ALLOWED\_ORIGINS, RATE\_LIMIT\_REDIS\_URL, OIDC\_ISSUER\_URL, OIDC\_AUDIENCE

* **auth-service** — Tenant & user auth (OIDC), API keys, sessions. (7001/50051)

   ENV: KC\_DB\_URL, KC\_HOSTNAME, EMAIL\_SMTP\_URL

### **\[A2\] NEW — Ingestion & Snapshot Layer**

* **scm-app-service (Mode A: Connect via Git Provider)** — Git app installs, repo selection, webhooks, shallow path-scoped clones. (7051/50041)

   ENV: GITHUB\_APP\_ID, GITHUB\_APP\_PRIVATE\_KEY, GITHUB\_WEBHOOK\_SECRET, SCM\_ALLOWED\_HOSTS, SNAPSHOT\_BUCKET

* **zip-ingest-service (Mode C: Upload a ZIP)** — Secure ZIP upload, malware scan (ClamAV), unpack, create snapshot. (7052/50042)

   ENV: MAX\_UPLOAD\_MB, CLAMAV\_URL, SNAPSHOT\_BUCKET

* **agent-ingest-service (Mode B: Privacy Mode)** — Receive signed results/patch bundles from customer-side agent. (7053/50043)

   ENV: AGENT\_PUBLIC\_KEY\_JWKS, REPORT\_MAX\_MB, RESULTS\_BUCKET

* **snapshot-orchestrator-service (central)** — Normalize any ingestion into canonical **Snapshot** and trigger Safety Checks. (7054/50044)

   ENV: SNAPSHOT\_TTL\_HOURS, PIPELINE\_TOPIC, S3\_SNAPSHOT\_PREFIX

### **\[B\] Product Core**

* **projects-service** — Tenants, projects, envs, memberships. (7002/50052)

* **spec-memory-service** — Requirements, ADRs, drift checks. (7101/50061)

* **ai-debugger-service** — Failure recorder → minimal repro; impact map. (7102/50062)

* **scaffold-secure-service** — Secure templates registry. (7103/—)

* **secrets-env-service** — Vaulting, env parity diff, client leak lint. (7104/50064)

* **db-guardian-service** — Least-priv roles, migration guard, index advisor. (7105/50065)

* **dep-governance-service** — SBOM/CVE/license policies. (7106/50066)

* **NEW: fix-engine-service** — Builds **Fix List**, computes safe code-mods/patches (Tier 1 auto; Tier 2 PR; Tier 3 advisory). (7151/50067)

   Stack: Node (TS) \+ Go workers (tree-sitter/gomod/yarn/pnpm) \+ S3 for patch artifacts.

   ENV: PATCH\_MAX\_FILES, PATCH\_MAX\_LOC, PATCH\_TTL\_DAYS, DIFF\_ENGINE (git|tree)

### **\[C\] Ship-Right**

* **publisher-service** *(renamed from cicd-guardrails-service; customer UI \= “Publish Safely / Undo”)* — Pre-flight checks, canary, instant rollback, release notes, **safe copy/branch** operations. (7201/50071)

   ENV: ROLLBACK\_STRATEGY, CANARY\_HEALTH\_URL, plus provider creds.

* **test-lab-service** — Unit/e2e scaffolds, ephemeral previews, cross-browser grid. (7202/50072)

* **authz-matrix-service** — Role designer, IDOR scan. (7203/50073)

* **rate-limit-service (RLS)** — Per-route policies, Redis store. (7204/50074)

### **\[D\] Run-Right**

* **observability-service** — OTel presets, SLOs, alerts; error inbox. (7301/50081; OTel 4317/4318)

* **perf-coach-service** — Plan analyzer, N+1 detector, cache/queue recipes. (7302/50082)

* **finops-service** — Cost dashboards, anomaly alerts, forecasts; metering. (7401/50091)

### **\[E\] Grow-Right**

* **mvp-validator-service** (7402/50092), **positioning-engine-service** (7501/50101), **launch-engine-service** (7502/50102), **legal-automator-service** (7503/50103), **motivation-engine-service** (7601/50111), **billing-service** (7901/50121), **notification-service** (7902/50122), **webhook-relay-service** (7903/50123)

## **3.3 Port Policy & Cross-Service Access**

Public 80/443; internal REST 7000–7999; internal gRPC 50050–50150; OTel 4317/4318.

Mesh mTLS for east-west; L7 auth via Envoy filters.

---

# **4\) Core APIs (high-level)**

*(Existing examples retained; new endpoints below.)*

## **NEW — Ingestion & Snapshots**

* POST /v1/projects/{id}/connect/scm → start Git-App install; provider URLs.

* POST /v1/projects/{id}/ingest/zip → signed upload URL → returns snapshotId.

* POST /v1/projects/{id}/ingest/agent → accept signed agent report → snapshotId.

* GET /v1/snapshots/{snapshotId} → status & attached reports.

## **NEW — Fix List & Patch/Apply**

* GET /v1/snapshots/{snapshotId}/fix-list → list of proposed fixes {tier, risk, diffs, tests}.

* POST /v1/snapshots/{snapshotId}/fixes/apply → {tierScope: \[1|2\], dryRun: bool, changeBudget: {maxFiles,maxLoc}}

  * **Mode A:** creates safe branch/PR with diffs \+ tests.

  * **Mode B:** returns **Patch Bundle** (signed) for customer CI to apply.

  * **Mode C:** returns **Hardened Bundle** (patched ZIP) \+ **Fix Report**.

* GET /v1/fixes/{fixId}/manifest → **Fix Manifest** (who/when/why/files/tests/locDelta).

## **NEW — Customer-facing Deploy**

* POST /v1/snapshots/{snapshotId}/publish-safely → deploy snapshot with automatic backup.

* POST /v1/publish/{publishId}/undo → one-click rollback.

*(publisher-service keeps the lower-level preflight/canary/rollback APIs for advanced users and internal use.)*

---

# **5\) Data Model (selected entities)**

*(Existing entities retained)*

**NEW (ingestion):**

* **Snapshot(id, projectId, mode\[A|B|C\], sourceRef{repo+sha|zipId|agentRunId}, modulePath, createdAt, storageKey)**

* **ProviderInstall(tenantId, provider, installId, scopes, repos\[\])**

* **AgentRegistration(projectId, publicKey, expiresAt)**

* **ProjectIngestionMode(projectId, defaultMode, allowedModes\[\])**

**NEW (fix/patch/publish):**

* **Fix(id, snapshotId, tier\[1|2|3\], category, severity, filePaths\[\], diffRef, suggestedPatchRef, status\[suggested|applied|rejected\])**

* **FixManifest(id, fixId, summary, risk, tests\[\], files\[\], locDelta, createdBy, createdAt)**

* **PatchBundle(id, snapshotId, mode\[A|B|C\], s3Key, sha256, size, ttlAt)**

* **Publish(id, projectId, snapshotId, releaseId, status, startedAt, finishedAt, undoRef)**

* **FixPolicy(orgId|projectId, yaml, changeBudget{maxFiles,maxLoc}, tierRules)**

---

# **6\) Event Bus (NATS topics & payloads)**

*(Existing topics retained; traceparent everywhere.)*

**NEW (ingestion):**

* ingest.scm.webhook.received

* ingest.zip.uploaded

* ingest.agent.report.received

* snapshot.created {snapshotId, projectId, mode}

* snapshot.ready

**NEW (fix/apply/publish):**

* fixlist.created {snapshotId, count, tierBreakdown}

* fixes.apply.requested {snapshotId, tierScope, dryRun}

* fixes.applied {snapshotId, fixIds\[\], manifestUrl}

* patchbundle.ready {snapshotId, mode, bundleUrl}

* publish.started|healthy|rolledback|failed {publishId, projectId}

---

# **7\) Policy-as-Code**

## **7.1 Publish Gates (unchanged; applies to Snapshots and PRs)**

apiVersion: lma/v1

kind: PublishPolicy

metadata:

  projectId: abc123

spec:

  gates:

    \- name: telemetry

      blocking: true

      rules:

        \- requireErrorMonitoring: true

        \- requireAnalytics: true

    \- name: secrets

      blocking: true

      rules:

        \- noClientSecrets: true

        \- envParity: \["staging","prod"\]

    \- name: security

      blocking: true

      rules:

        \- headersPresent: \["CSP","HSTS","X-Frame-Options"\]

        \- sqlParameterized: true

    \- name: authz

      blocking: true

      rules:

        \- rbacCoverage: "\>=95%"

        \- idorScanClean: true

    \- name: rate-limit

      blocking: true

      rules:

        \- publicPostRoutesLimited: true

    \- name: testing

      blocking: true

      rules:

        \- unitPass: true

        \- e2eSmokePass: true

        \- crossBrowserTopPaths: \["Chrome","Safari","Firefox"\]

    \- name: perf

      blocking: false

      rules:

        \- p95Ms: 1200

    \- name: legal

      blocking: true

      rules:

        \- privacyPolicyPresent: true

        \- cookieConsentEnabled: true

## **7.2** 

## **NEW: Fix Policy (org/project level)**

apiVersion: lma/v1

kind: FixPolicy

metadata:

  scope: org   \# or project

spec:

  tier1:

    autoApply: true

    maxFiles: 25

    maxLoc: 500

  tier2:

    requireApproval: true   \# we open a safe branch/PR or return a patch bundle

    reviewers: \["owners","security-champions"\]

  tier3:

    advisoryOnly: true

  snippetRetention: false     \# code snippets in reports off by default

  dataRegion: "asia-southeast1"

  retentionDays: 90

---

# **8\) Environment Variables (by concern)**

*(Existing groups retained.)*

**NEW (ingestion-specific):**

* **SCM / Mode A:** GITHUB\_WEBHOOK\_SECRET, SCM\_ALLOWED\_HOSTS, SNAPSHOT\_BUCKET

* **ZIP / Mode C:** MAX\_UPLOAD\_MB, CLAMAV\_URL, SNAPSHOT\_BUCKET

* **Agent / Mode B:** AGENT\_PUBLIC\_KEY\_JWKS, REPORT\_MAX\_MB, RESULTS\_BUCKET

**NEW (fix/patch):**

PATCH\_MAX\_FILES, PATCH\_MAX\_LOC, PATCH\_TTL\_DAYS, DIFF\_ENGINE (git|tree)

*All secrets via Vault; never bake into images.*

---

# **9\) SLOs / SLAs / NFRs**

*(Existing SLOs retained.)*

**Ingestion SLOs**

* **Snapshot creation (Mode A/C):** p50 ≤ 30s, p95 ≤ 90s.

* **ZIP processing & malware scan:** p95 ≤ 120s for ≤ 200 MB.

* **Agent report ingest:** p95 ≤ 10s to snapshot-ready.

**Fix/Publish SLOs**

* **Fix List generation:** p50 ≤ 10s for repos ≤ 3k files.

* **Tier-1 apply (dry-run to branch/patch):** p95 ≤ 90s; **change-budget** enforced.

* **Publish Safely canary decision:** ≤ 3 min; **Rollback MTTR:** ≤ 2 min.

**Security & compliance**

mTLS east-west; JWT scopes at edge & mesh; CIS for K8s; cosign image signing; SLSA-L3 target; secrets at rest via KMS; **zero source-code retention by default** (Modes A/C); Mode B sends **results only**; full audit logs for snapshots, fixes, and publishes.

---

# **10\) Deployment & Environments**

Envs: dev → staging → prod.

K8s: 3 node pools (system/compute/batch).

Ingress: Envoy Gateway; cert-manager \+ ACME.

Service discovery: Istio.

Jobs: K8s Jobs for scans/tests; HPA for bursty services.

Storage: Postgres HA (Patroni), Redis HA, S3-compatible object store.

Backups: pgBackRest nightly \+ PITR; object versioning; DR runbook.

---

# **11\) CI/CD Flow**

## **A) Customer-friendly path (UI, no jargon)**

**Connect/Upload/Privacy Mode → Create Snapshot → Safety Checks → Fix List → (Apply Fixes or Keep as Suggestion) → Publish Safely → Undo**

## **B) Advanced (PR-based)**

* **On PR:** lint/type/unit; SBOM/CVE; secrets scan; spec-drift; preview env; cross-browser smoke; gate report comment.

* **On main merge:** build & sign images; canary via publisher-service; auto-promote/rollback; release notes.

---

# **12\) Readiness Scorecard (what users see)**

Technical (Security/Reliability/Perf/Tests), Product (ICP/MVP/Onboarding), Business (Distribution/Legal/Telemetry), Founder Ops (Focus/Momentum).

Weights configurable; failing gates link to **one-click fix** flows.

---

# **13\) Milestones (parallelizable)**

**Priorities (based on typical vibe-platform handoffs):**

**P0:** Mode **A** (Git Provider) \+ Mode **C** (ZIP) \+ Tier-1 fixes.

**P1:** Mode **B** (Privacy/Agent) \+ Tier-2 patch-with-review.

**Sprint 1–2 (Weeks 1–4) – P0**

* scm-app-service (GitHub first), zip-ingest-service, snapshot-orchestrator-service.

* api-gateway, auth-service, projects-service, secrets-env-service.

* **fix-engine-service (Tier-1)**, publisher-service (Publish/Undo).

* Basic Safety Checks (secrets, headers, deps, basic browser smoke). Contracts \+ skeleton UIs.

**Sprint 3–4 (Weeks 5–8) – P0**

* scm-app-service → GitLab/Bitbucket.

* dep-governance, db-guardian, test-lab (previews).

* Observability \+ scorecard stub; legal pack; DMARC/SPF; perf hints.

**Sprint 5–6 (Weeks 9–12) – P1**

* agent-ingest-service \+ agent Action/CLI.

* authz-matrix, observability error inbox, perf-coach.

* mvp-validator, launch engine, billing.

* **fix-engine Tier-2** (PR/patch-bundle workflow, reviewers).

**Hardening (Weeks 13–14)**

* SLOs, alerting, HA, DR runbook; security pen test.

---

# **14\) Risks & Mitigations**

* False positives on gates → soft mode, overrides, appeal notes.

* Browser lab flakiness → isolate tests, retries, artifacts.

* Vendor lock-in → adapters for email/payments/CI; contract-first APIs.

* Cost spikes → HPA autoscale; queue back-pressure; FinOps alerts.

* **Repo access revoked mid-run** → idempotent snapshots, resume/retry, fallback to ZIP/Agent.

* **Large/malicious ZIPs** → size caps, ClamAV, streaming unzip, quarantine bucket.

* **Over-eager auto-fix** → **Fix Policy** with change-budget (max files/LOC) and tiered behavior; dry-run previews always.

---

# **15\) Example** 

# **.env**

#  **Templates**

**Service-agnostic**

ENV=staging

SERVICE\_NAME=publisher

SERVICE\_PORT=7201

OTEL\_EXPORTER\_OTLP\_ENDPOINT=http://otel-collector:4317

LOG\_LEVEL=info

JWT\_PUBLIC\_KEY=https://auth.my-lma.com/.well-known/jwks.json

**publisher-service**

GITHUB\_APP\_ID=...

GITHUB\_PRIVATE\_KEY=@/vault/secrets/github\_app\_key

ROLLBACK\_STRATEGY=bluegreen

CANARY\_HEALTH\_URL=/healthz

**scm-app-service**

GITHUB\_APP\_ID=...

GITHUB\_APP\_PRIVATE\_KEY=@/vault/github\_app.pem

GITHUB\_WEBHOOK\_SECRET=...

SCM\_ALLOWED\_HOSTS=github.com,gitlab.com,bitbucket.org

SNAPSHOT\_BUCKET=lma-snapshots

**zip-ingest-service**

MAX\_UPLOAD\_MB=500

CLAMAV\_URL=tcp://clamav:3310

SNAPSHOT\_BUCKET=lma-snapshots

**agent-ingest-service**

AGENT\_PUBLIC\_KEY\_JWKS=https://api.lma.com/agent/jwks

REPORT\_MAX\_MB=50

RESULTS\_BUCKET=lma-results

**fix-engine-service**

PATCH\_MAX\_FILES=25

PATCH\_MAX\_LOC=500

PATCH\_TTL\_DAYS=7

DIFF\_ENGINE=git

**secrets-env-service / observability-service** (unchanged; see earlier)

---

# **16\) Acceptance Criteria (v1)**

## **Mode A — Connect via Git Provider (Recommended)**

* Customer selects a repo → **Snapshot** created on change → **Safety Checks** run → **Fix List** shown.

* **Apply Fixes** with **Tier-1 auto-repair** creates safe branch and runs verification tests; **Tier-2** offers PR/approval; **Tier-3** advisory only.

* **Publish Safely** deploys with automatic backup; **Undo** reverts instantly.

* Default data posture: **no source retention** (findings/artifacts only).

## **Mode C — Upload a ZIP (No Git yet)**

* ZIP uploaded → malware scan passes → **Snapshot** created → Safety Checks \+ **Fix List**.

* **Apply Fixes** returns **Hardened Bundle** (patched ZIP) \+ **Fix Report**; offer to recreate as PR when Git connects.

* Uploaded archive deleted after snapshot build (configurable TTL).

## **Mode B — Run Checks Inside Your Account (Privacy Mode)**

* Customer adds **Agent** step → platform receives **signed results** → Fix List \+ **Patch Bundle** provided; customer applies inside their CI.

* **Publish Safely/Undo** available via safe-copy flow (or customer manual deploy).

* Platform receives **results only**; no source.

## **Advanced (PR-based teams)**

* Create project → connect repo → PR opened → gate report (secrets, security headers, SBOM/CVE, spec drift, unit tests, preview e2e smoke).

* Merge → canary auto-promote or rollback; release notes posted.

## **Common**

* Readiness scorecard visible; failing gates have **one-click fix** CTAs.

* Legal pack & consent banner generated and installed.

* Error inbox groups errors within 60s of a fault.

* Rate-limits enforced on public POST routes; policy-as-code lives in repo.

* **Fix Policy** YAML respected (tier rules, change-budget, snippet retention, region/retention).

---

This PRD now makes **repo ingestion**, **Fix List \+ Auto-Fix**, and **customer-friendly Publish/Undo** first-class—while keeping the engineering rigor and parallelizable microservices your teams need to start building immediately.

