# Production-Ready Go REST API Boilerplate

Enterprise-grade Go backend boilerplate architected with **Clean Architecture** and **Feature-Driven Architecture** using **Gin**, **MySQL**, **Redis**, **JWT Authentication**, and **RBAC Authorization**.

---

## Features

- **Architecture**: Clean Architecture + Feature-Driven Vertical Slices (Delivery $\rightarrow$ Application $\rightarrow$ Domain).
- **Framework**: High performance Gin Web Framework.
- **Database**: MySQL with connection pooling, transactions, migrations, and repository pattern.
- **Caching & Sessions**: Redis for high-speed token revocation and RBAC permission caching.
- **Authentication**: JWT Access/Refresh tokens with cryptographic token rotation, UUID JTI, and algorithm enforcement.
- **Authorization & RBAC**: Fine-grained deterministic permission middleware (`RequirePermission("resource:action")`), system role protections, idempotent seeder.
- **Canonical Feature Template**: Reference module (`articles`) demonstrating full CRUD, safe sorting allowlists, filtering, and pagination.
- **Developer Scaffolding**: Instant feature code generation via `make feature NAME=<name>`.
- **Production Hardening**: Slowloris protection (`ReadHeaderTimeout`), strict request body size limits (`413 Payload Too Large`), OWASP security headers (`nosniff`, `DENY`, `strict-origin-when-cross-origin`, strict CSP).
- **Distributed Rate Limiting**: Redis-backed rate limiting with dual tiers (general API tier: 100 req/min, auth tier: 10 req/min) returning standardized `429 Too Many Requests` with retry headers.
- **Idempotency Engine**: Redis-backed `Idempotency-Key` middleware with SHA-256 fingerprinting preventing duplicate side effects from client retries.
- **Ordered Graceful Teardown**: Coordinated shutdown draining in-flight requests before executing MySQL and Redis cleanup handlers.
- **Password Security**: Bcrypt hashing with secure cost factor and constant-time verification.
- **Observability**: Go stdlib `log/slog` structured logging, Request-ID tracing, and health check probes (`/health`, `/health/live`, `/health/ready`).
- **Resilience**: Panic recovery middleware, request timeout, graceful shutdown on `SIGINT`/`SIGTERM`.
- **Developer Experience**: Comprehensive `Makefile`, typed error hierarchy, unified JSON response envelope, multi-stage minimal Docker container, OpenAPI 3.0 specification.

---

## Directory Structure

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
│   │   ├── middleware/              # RequestID, Logger, Recovery, CORS, Timeout, Auth, RequirePermission
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
│           ├── domain/              # Article Entity & Repository Interface
│           ├── application/         # Commands, Queries, DTOs & Service
│           ├── infrastructure/      # MySQL Repository & In-Memory Test Repo
│           └── delivery/            # Handlers, Request/Response & Protected Routes
├── migrations/                      # SQL Schema migrations
├── docs/                            # Guides, Architecture & OpenAPI 3.0 specification
├── Dockerfile                       # Multi-stage production container
├── docker-compose.yml               # Complete stack (API, MySQL 8, Redis 7)
├── Makefile                         # DX tooling
├── .env.example                     # Environment template
└── README.md
```

---

## Quick Start

### 1. Prerequisites
- [Go 1.22+](https://golang.org/dl/)
- [Docker](https://www.docker.com/) & Docker Compose (optional for containerized run)
- [Make](https://www.gnu.org/software/make/) (optional)

### 2. Environment Configuration
Copy `.env.example` to `.env`:
```bash
cp .env.example .env
```

### 3. Run with Docker Compose (Recommended)
Starts API, MySQL 8, and Redis 7 in connected containers with health checks:
```bash
make docker-up
# Or: docker-compose up -d --build
```

### 4. Run Locally
```bash
# Start MySQL & Redis if running locally, then:
make run
# Or: go run ./cmd/api
```

### 5. Scaffold a New Feature
Generate a complete, compiling Clean Architecture module skeleton:
```bash
make feature NAME=product
```

---

## API Endpoints

### Health Module
| Method | Path | Description |
|---|---|---|
| `GET` | `/health` | Overall system health overview |
| `GET` | `/health/live` | Liveness probe (shallow check for k8s/docker) |
| `GET` | `/health/ready` | Readiness probe (deep check: MySQL + Redis) |

### Authentication Module
| Method | Path | Description | Protected |
|---|---|---|:---:|
| `POST` | `/api/v1/auth/register` | Register new user account | No |
| `POST` | `/api/v1/auth/login` | Login and receive token pair | No |
| `POST` | `/api/v1/auth/refresh` | Refresh tokens (with Token Rotation) | No |
| `POST` | `/api/v1/auth/logout` | Revoke active refresh token | No |
| `GET` | `/api/v1/auth/me` | Fetch authenticated user profile | **Yes** (Bearer) |
| `GET` | `/api/v1/auth/me/permissions` | Fetch current user effective permissions | **Yes** (Bearer) |

### RBAC Module
| Method | Path | Description | Required Permission |
|---|---|---|:---:|
| `POST` | `/api/v1/roles` | Create new role | `role:create` |
| `GET` | `/api/v1/roles` | List roles (paginated) | `role:read` |
| `GET` | `/api/v1/roles/:id` | Get role by ID | `role:read` |
| `PUT` | `/api/v1/roles/:id` | Update role | `role:update` |
| `DELETE` | `/api/v1/roles/:id` | Delete custom role (System roles protected) | `role:delete` |
| `PUT` | `/api/v1/roles/:id/permissions` | Assign permissions to role (Transactional) | `role:update` |
| `GET` | `/api/v1/roles/:id/permissions` | Get permissions assigned to role | `role:read` |
| `POST` | `/api/v1/permissions` | Create new permission | `permission:create` |
| `GET` | `/api/v1/permissions` | List permissions (paginated) | `permission:read` |
| `GET` | `/api/v1/permissions/:id` | Get permission by ID | `permission:read` |
| `PUT` | `/api/v1/permissions/:id` | Update permission | `permission:update` |
| `DELETE` | `/api/v1/permissions/:id` | Delete permission | `permission:delete` |
| `PUT` | `/api/v1/users/:id/roles` | Assign roles to user (Transactional) | `user:assign-role` |
| `GET` | `/api/v1/users/:id/roles` | Get roles assigned to user | `role:read` |
| `GET` | `/api/v1/users/:id/permissions` | Get effective permissions for user | `permission:read` |

### Articles Module (Reference Feature)
| Method | Path | Description | Required Permission |
|---|---|---|:---:|
| `GET` | `/api/v1/articles` | List articles (paginated with filters & sort) | Public |
| `GET` | `/api/v1/articles/:id` | Get article by ID | Public |
| `GET` | `/api/v1/articles/slug/:slug` | Get article by slug | Public |
| `POST` | `/api/v1/articles` | Create article | `article:create` |
| `PUT` | `/api/v1/articles/:id` | Update article | `article:update` |
| `DELETE` | `/api/v1/articles/:id` | Delete article | `article:delete` |

---

## Testing & Quality Gates

Run all unit tests:
```bash
make test
```

Run test suite with race detector:
```bash
make test-race
```

Run test suite with code coverage report:
```bash
make test-cover
```

Format and static analysis:
```bash
make fmt
make vet
```

---

## License

MIT
