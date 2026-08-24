# Architecture & System Design Documentation

This document outlines the architecture, principles, security practices, and patterns used in this boilerplate.

---

## 1. Architectural Style

The application combines **Clean Architecture** (by Robert C. Martin) with **Feature-Driven Architecture (Vertical Slices)**.

### Dependency Rule
Dependencies only point inward:

$$\text{Delivery (HTTP/Handlers)} \longrightarrow \text{Application (Use Cases/DTOs)} \longrightarrow \text{Domain (Entities/Interfaces)}$$

- **Domain Layer (`internal/modules/{feature}/domain`)**:
  - Contains core entities, business invariants, and repository/service interfaces.
  - Zero external dependencies (no Gin, no MySQL driver, no Redis, no HTTP/ORM libraries).
- **Application Layer (`internal/modules/{feature}/application`)**:
  - Implements use cases (commands & queries), coordinates transactions, translates domain models to DTOs.
  - Depends only on the Domain layer and shared primitives.
- **Infrastructure Layer (`internal/modules/{feature}/infrastructure` & `internal/shared`)**:
  - Implements interfaces defined in the Domain or Application layers.
  - Adapts external technology: MySQL (`database/sql`), Redis (`go-redis/v9`), JWT (`golang-jwt/jwt/v5`), Bcrypt password hashing.
- **Delivery Layer (`internal/modules/{feature}/delivery`)**:
  - Exposes HTTP endpoints via Gin handlers, validates request bodies, and serializes standardized JSON responses.

---

## 2. Directory Layout

```text
.
├── cmd/
│   ├── api/
│   │   └── main.go                  # Main entry point & DI container
│   └── scaffold/
│       └── main.go                  # Feature scaffolding CLI generator
├── internal/
│   ├── shared/                      # Cross-cutting foundational infrastructure
│   │   ├── config/                  # Strongly typed config with validation
│   │   ├── database/                # MySQL connection pool
│   │   ├── redis/                   # Redis client pool
│   │   ├── logger/                  # Structured slog logger with secret sanitization
│   │   ├── metrics/                 # Prometheus metrics registry & collectors (Phase 6)
│   │   ├── tracer/                  # W3C TraceContext & OpenTelemetry readiness (Phase 6)
│   │   ├── middleware/              # Hardening, Auth, Metrics, RateLimiter, Idempotency, etc.
│   │   ├── pagination/              # Unified pagination extracting & metadata
│   │   ├── queryparam/              # Safe query parameter sorting with column allowlists
│   │   ├── response/                # Unified JSON response contract
│   │   ├── errors/                  # Strongly typed error hierarchy
│   │   └── httpserver/              # Server lifecycle and graceful shutdown
│   │
│   └── modules/                     # Feature slices
│       ├── health/                  # Health check probes (Liveness, Readiness)
│       ├── auth/                    # Authentication & User Management
│       ├── rbac/                    # Role-Based Access Control (Phase 3)
│       └── article/                 # Canonical Reference Feature (Phase 4)
├── migrations/                      # SQL Schema migrations
├── deployments/                     # Container & orchestration definitions
├── docs/                            # Guides, Architecture & OpenAPI specification
├── Dockerfile                       # Multi-stage production container
├── docker-compose.yml               # Complete stack (API, MySQL 8, Redis 7)
├── Makefile                         # DX tooling
└── README.md
```

---

## 3. Authentication & Security Design

### JWT Token Lifecycle & Rotation
1. **Login / Register**: Issues a short-lived **Access Token** (default 15 minutes) and a longer-lived **Refresh Token** (default 7 days) containing a unique `jti` (UUID).
2. **Redis Storage**: The `jti` is stored in Redis under `auth:refresh:{token_id}` with expiration matching the token's lifetime.
3. **Token Rotation on Refresh**:
   - Client sends refresh token to `POST /api/v1/auth/refresh`.
   - Refresh token is verified cryptographically and validated against Redis.
   - The old refresh token is **immediately revoked** in Redis.
   - A brand new token pair (Access + Refresh) is generated and stored in Redis.
   - If an expired or revoked token is reused, the request is rejected immediately.
4. **Logout**:
   - Client calls `POST /api/v1/auth/logout` with their refresh token.
   - Token is revoked from Redis.

### Cryptographic Protections
- **Algorithm Enforcement**: Validates explicitly against `HS256` to prevent `none`-algorithm and algorithm substitution attacks.
- **Claims Verification**: Validates Issuer (`iss`), Audience (`aud`), and Expiration (`exp`) with clock skew tolerance.
- **Password Security**: Passwords are saved with `bcrypt` (cost 12), never logged, never returned in API responses.

---

## 4. Role-Based Access Control (RBAC) Design

### Conceptual Model
$$\text{User} \longleftrightarrow \text{UserRoles} \longleftrightarrow \text{Role} \longleftrightarrow \text{RolePermissions} \longleftrightarrow \text{Permission}$$

- **Roles**: Logical groupings of permissions (e.g. `admin`, `editor`, `viewer`). Marked with `is_system = true` for immutable application roles.
- **Permissions**: Atomic authorization rules formatted deterministically as `resource:action` (e.g. `user:read`, `role:create`, `permission:delete`, `article:create`).
- **Separation of Concerns**:
  - **Authentication** answers: *Who are you?* (Handled by `AuthMiddleware`).
  - **Authorization** answers: *Are you allowed to perform this operation?* (Handled by `RequirePermission`).

### Redis Permission Caching & Invalidation
- **Cache Key**: `rbac:user:{user_id}:permissions`
- **Cache-Aside Flow**:
  1. `RequirePermission` queries `AuthorizationService.HasPermission(ctx, userID, permission)`.
  2. If cached in Redis $\rightarrow$ instant authorization decision.
  3. If cache miss $\rightarrow$ queries MySQL join table, populates Redis cache with TTL (1 hour).
  4. If Redis is down/unavailable $\rightarrow$ graceful fallback to MySQL without blocking authorization.
- **Explicit Invalidation**:
  - Updating role permissions $\rightarrow$ invalidates all users assigned to that role.
  - Assigning roles to a user $\rightarrow$ invalidates target user cache.
  - Updating/deleting permissions $\rightarrow$ invalidates all affected user caches.

---

## 5. Production Hardening & Resilience (Phase 5)

### HTTP Server Protections
- **ReadHeaderTimeout**: Set to 5 seconds to mitigate Slowloris denial-of-service attacks.
- **MaxBodySize**: Enforced via `http.MaxBytesReader` (default 2MB) responding with `413 Payload Too Large`.
- **Security Headers**: Injects `X-Content-Type-Options: nosniff`, `X-Frame-Options: DENY`, `Referrer-Policy: strict-origin-when-cross-origin`, and `Content-Security-Policy: default-src 'none'; frame-ancestors 'none'`.

### Ordered Graceful Shutdown
1. `SIGINT`/`SIGTERM` received $\rightarrow$ HTTP server stops accepting incoming connections.
2. Active HTTP requests drain gracefully up to `ShutdownTimeout` (default 10s).
3. MySQL connection pool is closed cleanly.
4. Redis client connections are closed cleanly.

### Distributed Rate Limiting
- Backed by Redis sliding 1-minute window counters.
- **Auth Tier**: 10 requests/minute on `/api/v1/auth/*`.
- **General Tier**: 100 requests/minute on `/api/v1/*`.
- Headers: `X-RateLimit-Limit`, `X-RateLimit-Remaining`, `X-RateLimit-Reset`, `Retry-After`.

### Idempotency Engine (`Idempotency-Key`)
- Mutation requests (`POST`, `PUT`, `PATCH`, `DELETE`) with `Idempotency-Key` are fingerprinted via SHA-256 (`method + path + body`).
- Prevents duplicate side-effects from network retries by returning cached responses (`X-Idempotent-Replay: true`).
- Detects payload tampering with identical keys (`422 Unprocessable Entity`).

### User Activity Tracking & Write Throttling
- **`last_login_at`**: Updated explicitly in UTC upon successful user authentication (login). Best-effort failure semantics ensure audit update failures do not prevent successful login.
- **`last_activity_at`**: Recorded on authenticated business requests.
- **Redis Write Throttling**: Uses atomic `SetNX` on key `user:activity:{user_id}` with TTL `USER_ACTIVITY_UPDATE_INTERVAL` (default 5m). Database writes are only triggered on key acquisition/expiration, preventing high-frequency write storms on MySQL.
- **Fallback**: If Redis is unavailable, an in-process bounded cache throttles database updates, failing open safely.

---

## 6. Observability & Operational Readiness (Phase 6)

### Structured Logging (`log/slog`)
- Log level configurable via `LOG_LEVEL` (`DEBUG`, `INFO`, `WARN`, `ERROR`).
- In production (`APP_ENV=production`), output is structured JSON.
- Contextual logging via `logger.WithContext(ctx)` embeds `request_id`, `trace_id`, and `user_id`.
- Request logs record: `timestamp`, `level`, `request_id`, `trace_id`, `method`, `path`, `status`, `latency`, `ip`, `user_agent`.
- Sensitive data masking automatically scrubs `password`, `token`, `secret`, `authorization`, `db_password`, `apikey`.

### Prometheus Metrics (`/metrics`)
All metrics enforce strict low-cardinality label sets:
- `app_http_requests_total{method, path, status}` (path uses template e.g. `/api/v1/articles/:id`)
- `app_http_request_duration_seconds{method, path, status}` (Histogram)
- `app_http_errors_total{method, path, error_type}`
- `app_auth_failures_total{reason}`
- `app_rbac_denials_total{permission}`
- `app_ratelimit_rejections_total{tier}`
- `app_cache_hits_total{cache}` & `app_cache_misses_total{cache}`
- `app_db_connections_open`, `app_db_connections_in_use`, `app_db_connections_idle`, `app_db_connections_wait_count`

### Distributed Tracing & W3C Correlation
- Parses standard W3C `traceparent` (`00-{trace_id}-{span_id}-{flags}`) or `X-Trace-ID` headers.
- Generates 128-bit trace IDs and propagates them across context, database, cache, and HTTP response headers.
- OpenTelemetry boundary interface (`Tracer`, `Span`) ready for OTel SDK drop-in without modifying business logic.

---

## 7. API Standard Contract

### Success Response
```json
{
  "success": true,
  "message": "Operation description",
  "data": {},
  "meta": null
}
```

### Error Response
```json
{
  "success": false,
  "message": "Validation failed",
  "errors": [
    {
      "field": "email",
      "message": "invalid email format"
    }
  ],
  "data": null,
  "meta": null
}
```

---

## 8. Centralized Error Hierarchy

| Error Type | HTTP Status | Code | Typical Use Case |
|---|---|---|---|
| `ValidationError` | 400 Bad Request | `VALIDATION_ERROR` | Malformed payload, invalid field constraints |
| `UnauthorizedError` | 401 Unauthorized | `UNAUTHORIZED` | Invalid/expired token, wrong credentials |
| `ForbiddenError` | 403 Forbidden | `FORBIDDEN` | Missing required RBAC permission or author ownership |
| `NotFoundError` | 404 Not Found | `NOT_FOUND` | User, role, or entity does not exist |
| `ConflictError` | 409 Conflict | `CONFLICT` | Duplicate email, unique slug, or concurrent idempotency in progress |
| `PayloadTooLarge` | 413 Payload Too Large | `PAYLOAD_TOO_LARGE` | Request body exceeds maximum allowed size |
| `BusinessError` | 422 Unprocessable | Custom code | Business rule violation or idempotency payload mismatch |
| `RateLimitExceeded` | 429 Too Many Requests | `RATE_LIMIT_EXCEEDED` | Request rate exceeded within 1-minute window |
| `InfrastructureError`| 503 Service Unavailable | `SERVICE_UNAVAILABLE` | Database/Redis connection down |
| `InternalError` | 500 Internal Error | `INTERNAL_SERVER_ERROR` | Panics, unexpected system failures |
