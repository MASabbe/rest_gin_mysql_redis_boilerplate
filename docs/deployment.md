# Deployment, Migration & Production Delivery Guide

This document defines the operational procedures for deploying, migrating, and maintaining this boilerplate in production environments.

---

## 1. Production Architecture Overview

The application is packaged as a lightweight, multi-stage, non-root Alpine container (`UID 10001`) with externalized configuration and database state.

```text
                                       +-------------------+
                                       |  GitHub Actions   |
                                       +---------+---------+
                                                 |
                                     (1. Build & Push Image)
                                                 v
+-------------------+                  +-------------------+
|   Load Balancer   |  -- (Traffic) -> |  App Container    | (Non-root, Port 8080)
| (Ingress / Envoy) |                  |  (Stateless API)  |
+---------+---------+                  +----+---------+----+
          |                                 |         |
     (Health Check)                         |         |
          v                                 v         v
    /health/ready                    +----------+  +----------+
                                     |  MySQL 8 |  |  Redis 7 |
                                     +----------+  +----------+
```

---

## 2. Deployment Workflow & Ordering

To prevent downtime and ensure transactional integrity, deployments must follow an ordered, deterministic sequence:

```text
1. Trigger Release (Git Tag vX.Y.Z or merge to main)
   ↓
2. GitHub Actions CI runs Quality Gate (fmt, vet, tests, race, coverage)
   ↓
3. Security Automation runs (Secret scan, gosec, Trivy container scan)
   ↓
4. Push Immutable Image Tag (e.g. ghcr.io/org/repo:sha-abc1234 & v1.0.0)
   ↓
5. Run Database Migrations (Explicit, backward-compatible expand phase)
   ↓
6. Deploy Application Container with new immutable image tag
   ↓
7. Orchestrator polls Readiness Probe (GET /health/ready)
   ↓
8. Execute Automated Smoke Test (Health + safe authenticated endpoint)
   ↓
9. Route production traffic to updated containers & drain old instances
```

---

## 3. Database Migration Strategy

### Guiding Principles
- **Never Run Destructive Migrations in App Startup**: Application containers must not execute uncoordinated migrations on boot.
- **Explicit & Versioned**: All schema changes are stored in `/migrations` as versioned pairs:
  - `{version}_{name}.up.sql`
  - `{version}_{name}.down.sql`
- **Reversible & Observable**: Every migration must have an associated down migration.
- **Expand / Contract (Zero-Downtime) Pattern**:
  1. **Phase 1 (Expand)**: Add new columns/tables (nullable or with defaults). Old and new code versions both operate safely.
  2. **Phase 2 (Deploy)**: Deploy new application version reading/writing the new schema.
  3. **Phase 3 (Contract)**: In a subsequent release, remove legacy unused columns/tables after old instances are drained.

### Running Migrations in CI/CD
Using `golang-migrate` CLI:
```bash
# Run pending migrations
migrate -path migrations -database "mysql://${MYSQL_USER}:${MYSQL_PASSWORD}@tcp(${MYSQL_HOST}:${MYSQL_PORT})/${MYSQL_DATABASE}" up

# Check migration status
migrate -path migrations -database "mysql://${MYSQL_USER}:${MYSQL_PASSWORD}@tcp(${MYSQL_HOST}:${MYSQL_PORT})/${MYSQL_DATABASE}" version
```

---

## 4. Rollback Strategy

### Application Rollback
Because images are tagged immutably by commit SHA and semantic version, application rollback is instantaneous:
```bash
# Re-deploy previous stable immutable tag
IMAGE_TAG="sha-previous-known-good"
```

### Database Rollback Rules
> [!WARNING]
> Database rollbacks are NOT automatic. Never run `migrate down` blindly in production without verifying data safety.

1. If schema changes were made following the **Expand/Contract** pattern, the previous application version will continue running without rolling back the database.
2. If a database rollback is required, test the `.down.sql` script in a staging replica first to verify zero data corruption before executing against production.

---

## 5. Smoke Testing & Verification

Immediately following deployment, run automated smoke tests against the new instance:

### 1. Liveness & Readiness Checks
```bash
# 1. Liveness Check (Should return 200 OK immediately)
curl -f -i http://localhost:8080/health/live

# 2. Readiness Check (Verifies MySQL and Redis connections)
curl -f -i http://localhost:8080/health/ready

# 3. Metrics Exposition Check (Verifies Prometheus metrics handler)
curl -f -i http://localhost:8080/metrics
```

### 2. Authenticated Flow Smoke Test
```bash
# Register temporary smoke test user
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"smoketest@example.com","password":"SmokeTestPassword123!"}'

# Verify profile
curl -X GET http://localhost:8080/api/v1/auth/me \
  -H "Authorization: Bearer <ACCESS_TOKEN>"
```

---

## 6. Environment Configuration Catalog

All configuration is externalized via environment variables.

| Variable | Type | Default / Example | Description |
|---|:---:|---|---|
| `APP_NAME` | string | `backend-api` | Application name for metrics and logging |
| `APP_ENV` | string | `production` | Environment (`development`, `staging`, `production`) |
| `LOG_LEVEL` | string | `info` | Logging verbosity (`debug`, `info`, `warn`, `error`) |
| `SERVER_HOST` | string | `0.0.0.0` | HTTP binding host |
| `SERVER_PORT` | int | `8080` | HTTP binding port |
| `SERVER_READ_TIMEOUT` | duration | `10s` | Maximum duration for reading entire request |
| `SERVER_WRITE_TIMEOUT` | duration | `10s` | Maximum duration for writing response |
| `SERVER_IDLE_TIMEOUT` | duration | `60s` | Maximum amount of time to wait for next request |
| `SERVER_MAX_BODY_SIZE` | int64 | `2097152` | Request body limit in bytes (default 2MB) |
| `SERVER_SHUTDOWN_TIMEOUT` | duration | `10s` | Maximum duration to drain active requests |
| `MYSQL_HOST` | string | `mysql.internal` | MySQL database host |
| `MYSQL_PORT` | int | `3306` | MySQL database port |
| `MYSQL_USER` | string | `app_user` | MySQL database username |
| `MYSQL_PASSWORD` | string | *(Secret)* | MySQL database password |
| `MYSQL_DATABASE` | string | `app_db` | MySQL database name |
| `MYSQL_MAX_OPEN_CONNS` | int | `25` | Connection pool max open connections |
| `MYSQL_MAX_IDLE_CONNS` | int | `10` | Connection pool max idle connections |
| `REDIS_HOST` | string | `redis.internal` | Redis cache host |
| `REDIS_PORT` | int | `6379` | Redis cache port |
| `REDIS_PASSWORD` | string | *(Secret)* | Optional Redis authentication password |
| `REDIS_DB` | int | `0` | Redis logical database number |
| `JWT_SECRET` | string | *(Secret, min 32 chars)* | HMAC-SHA256 signing secret |
| `JWT_ACCESS_TOKEN_TTL` | duration | `15m` | Access token lifespan |
| `JWT_REFRESH_TOKEN_TTL` | duration | `168h` (7d) | Refresh token lifespan |
| `JWT_ISSUER` | string | `backend-api` | Expected JWT issuer |
| `JWT_AUDIENCE` | string | `backend-api-users` | Expected JWT audience |

---

## 7. Recommended Branch Protection Rules

To maintain enterprise quality and prevent regression:

- **Target Branches**: `main`, `master`, `release/*`
- **Required Checks**:
  - `Build & Test Quality Gate` (must pass `ci.yml`)
  - `Automated Security Scan` (must pass `security.yml`)
- **Review Requirements**:
  - Require at least 1 approving review from code owners.
  - Dismiss stale pull request approvals when new commits are pushed.
  - Require conversation resolution before merging.
- **Merge Rules**:
  - Require linear history (Squash and Merge or Rebase).
  - Strictly prevent direct pushes (force-push disabled).
