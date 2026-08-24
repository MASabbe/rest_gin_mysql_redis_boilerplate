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
│   └── api/
│       └── main.go                  # Main entry point & DI container
├── internal/
│   ├── shared/                      # Cross-cutting foundational infrastructure
│   │   ├── config/                  # Strongly typed config with validation
│   │   ├── database/                # MySQL connection pool
│   │   ├── redis/                   # Redis client pool
│   │   ├── logger/                  # Structured slog logger with secret sanitization
│   │   ├── middleware/              # RequestID, Logger, Recovery, CORS, Timeout, Auth, RequirePermission
│   │   ├── pagination/              # Unified pagination extracting & metadata
│   │   ├── response/                # Unified JSON response contract
│   │   ├── errors/                  # Strongly typed error hierarchy
│   │   └── httpserver/              # Server lifecycle and graceful shutdown
│   │
│   └── modules/                     # Feature slices
│       ├── health/                  # Health check probes (Liveness, Readiness)
│       ├── auth/                    # Authentication & User Management
│       │   ├── domain/              # Entities, Repository/Service Interfaces
│       │   ├── application/         # Use cases (Register, Login, Refresh, Logout, Me)
│       │   ├── infrastructure/      # MySQL, Redis, JWT, Bcrypt implementations
│       │   └── delivery/            # HTTP Handlers, Request/Response DTOs, Routes
│       └── rbac/                    # Role-Based Access Control (Phase 3)
│           ├── domain/              # Role & Permission Entities, Repositories, Authz Service Interface
│           ├── application/         # Commands, Queries, DTOs, RBAC Application Service
│           ├── infrastructure/      # MySQL Repositories with Transactions, Redis Cache, Idempotent Seeder
│           └── delivery/            # Role, Permission & UserRole Handlers, Routes
├── migrations/                      # SQL Schema migrations
├── deployments/                     # Container & orchestration definitions
├── docs/                            # Architecture & OpenAPI specification
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
- **Permissions**: Atomic authorization rules formatted deterministically as `resource:action` (e.g. `user:read`, `role:create`, `permission:delete`).
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

## 5. API Standard Contract

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

## 6. Centralized Error Hierarchy

| Error Type | HTTP Status | Code | Typical Use Case |
|---|---|---|---|
| `ValidationError` | 400 Bad Request | `VALIDATION_ERROR` | Malformed payload, invalid field constraints |
| `UnauthorizedError` | 401 Unauthorized | `UNAUTHORIZED` | Invalid/expired token, wrong credentials |
| `ForbiddenError` | 403 Forbidden | `FORBIDDEN` | Missing required RBAC permission |
| `NotFoundError` | 404 Not Found | `NOT_FOUND` | User, role, or entity does not exist |
| `ConflictError` | 409 Conflict | `CONFLICT` | Duplicate email, role name, or unique constraint violation |
| `BusinessError` | 422 Unprocessable | Custom code | Business rule violation (e.g. attempting to delete system role) |
| `InfrastructureError`| 503 Service Unavailable | `SERVICE_UNAVAILABLE` | Database/Redis connection down |
| `InternalError` | 500 Internal Error | `INTERNAL_SERVER_ERROR` | Panics, unexpected system failures |
