# Feature Development Guide

This guide describes how to design, scaffold, and implement new business modules in this backend boilerplate following **Clean Architecture** and **Feature-Driven Vertical Slices**.

---

## 1. Architectural Principles & Rules

1. **Dependency Inward Rule**:
   $$\text{Delivery} \longrightarrow \text{Application} \longrightarrow \text{Domain}$$
   - The **Domain** layer has **zero dependencies** on external packages, frameworks, or database drivers.
   - The **Application** layer defines use case commands, queries, and coordinates domain entities and repository interfaces.
   - The **Infrastructure** layer implements the repository contracts defined by the domain.
   - The **Delivery** layer handles HTTP routing, request binding, and JSON serialization.

2. **Explicit DTO Boundaries**:
   - Never expose raw Domain entities or database models directly through HTTP handlers.
   - Handlers map `Request DTO` $\rightarrow$ `Application Command/Query`.
   - Application use cases map `Domain Entity` $\rightarrow$ `Response DTO`.

3. **Safe Database Access & Repositories**:
   - All repository methods must accept `context.Context` as the first argument.
   - Repositories must use domain-oriented method names (e.g. `FindByID`, `FindBySlug`, `List`) instead of database-centric names (`GetRow`, `QuerySQL`).
   - Query filters and sort parameters must use parameterized queries and safe column allowlists.

---

## 2. Standard Feature Directory Layout

```text
internal/modules/<feature_name>/
├── domain/
│   ├── entity/
│   │   ├── <feature>.go             # Domain entity, business invariants, status constructors
│   │   └── <feature>_test.go        # Pure domain unit tests
│   └── repository/
│       └── <feature>_repository.go  # Pure interface contracts
│
├── application/
│   ├── command/
│   │   ├── <feature>_commands.go    # Create/Update/Delete command structs & validation
│   │   └── commands_test.go
│   ├── query/
│   │   ├── <feature>_queries.go     # Query parameter structs & validation
│   │   └── queries_test.go
│   ├── dto/
│   │   └── <feature>_dto.go         # Application Data Transfer Objects & mappers
│   ├── service.go                   # Use Case interface and implementation
│   └── service_test.go              # Application service unit tests
│
├── infrastructure/
│   └── persistence/
│       ├── mysql_<feature>_repository.go      # Production MySQL implementation
│       └── inmemory_<feature>_repository.go   # In-Memory implementation for fast unit tests
│
└── delivery/
    └── http/
        ├── handler/
        │   ├── <feature>_handler.go      # Gin HTTP request handlers
        │   └── <feature>_handler_test.go # Handler unit tests with HTTP mocks
        ├── request/
        │   └── <feature>_request.go      # Gin request binding structs with validation tags
        ├── response/
        │   └── <feature>_response.go     # HTTP JSON response serialization models
        └── routes.go                     # Route registration & permission bindings
```

---

## 3. Scaffolding a New Feature

You can automatically generate a compiling canonical feature skeleton using the Makefile command:

```bash
make feature NAME=product
```

Or via direct Go command:
```bash
go run ./cmd/scaffold -name=product
```

This generates all domain entities, repository contracts, application services, request/response DTOs, handlers, and in-memory test mocks under `internal/modules/product/`.

---

## 4. Step-by-Step Feature Implementation Workflow

### Step 1: Database Migration
Create migration files under `migrations/`:
- `migrations/00000X_create_<feature>_table.up.sql`
- `migrations/00000X_create_<feature>_table.down.sql`

Example:
```sql
CREATE TABLE IF NOT EXISTS products (
    id VARCHAR(36) NOT NULL,
    name VARCHAR(255) NOT NULL,
    price DECIMAL(10, 2) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

### Step 2: Define Domain Entity & Invariants
In `internal/modules/<feature>/domain/entity/<feature>.go`:
- Validate required fields in the constructor `New<Feature>(...)`.
- Define methods on the entity struct for state transitions.

### Step 3: Implement Application Commands, Queries & Service
- Validate input in `command.Validate()` and `query.Validate()`.
- Coordinate repository calls in `application.New<Feature>Service(repo)`.

### Step 4: Implement MySQL Repository
In `internal/modules/<feature>/infrastructure/persistence/mysql_<feature>_repository.go`:
- Always use parameterized queries (`?`).
- Map database errors to centralized application errors (`appErrors.NewNotFoundError`, `appErrors.NewConflictError`, `appErrors.NewInternalError`).

### Step 5: Implement Delivery Handler & Routes
In `internal/modules/<feature>/delivery/http/`:
- Bind JSON using `c.ShouldBindJSON(&req)`.
- Use `internal/shared/pagination.Extract(c)` for pagination.
- Use `internal/shared/queryparam.ExtractSorting(c, allowlist, defaultField, defaultOrder)` for safe sorting.
- Respond with `sharedResponse.OK`, `sharedResponse.Created`, or `sharedResponse.Success`.
- Protect routes with `middleware.AuthMiddleware` and `middleware.RequirePermission`.

### Step 6: Main Wiring & Route Registration
In `cmd/api/main.go`:
- Instantiate MySQL repository: `featureRepo := featurePersistence.NewMySQL<Feature>Repository(mysqlDB.DB)`
- Instantiate service: `featureService := featureApp.New<Feature>Service(featureRepo)`
- Register routes: `featureHTTP.RegisterRoutes(apiV1, jwtService, rbacService, featureService)`
