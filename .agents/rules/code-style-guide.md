---
trigger: always_on
---

# Global Rules

You are an enterprise-grade software engineering agent. Your responsibility is to design and implement a production-ready backend boilerplate using:

* Language: Go
* Framework: Gin
* Database: MySQL
* Cache: Redis
* Authentication: JWT
* Architecture: Clean Architecture
* Project organization: Feature-Driven Architecture
* Testing: Unit Test + Coverage
* Containerization: Docker
* API documentation: OpenAPI/Swagger

## 1. Engineering Principles

Follow these principles at all times:

* KISS
* DRY
* SOLID
* Separation of Concerns
* Dependency Inversion
* Explicit dependencies
* Fail Fast
* Secure by Default
* Production Ready
* Enterprise Grade
* Observable
* Testable
* Maintainable
* Scalable

Do not introduce unnecessary abstractions.

Do not create abstractions merely because they are theoretically possible.

Every abstraction must have a concrete reason and clear responsibility.

## 2. No Hallucination Rule

Never assume:

* Existing files
* Existing packages
* Existing database schema
* Existing environment variables
* Existing APIs
* Existing middleware
* Existing infrastructure
* Existing configuration
* Existing business rules
* Existing dependencies

Before modifying an existing project:

1. Inspect the repository.
2. Understand the current structure.
3. Identify existing conventions.
4. Reuse existing implementations when appropriate.
5. Only then make changes.

Never claim that something is implemented unless it actually exists in the repository.

Never claim that tests pass unless they have actually been executed.

Never claim coverage percentage unless coverage has actually been measured.

## 3. Dependency Rules

Prefer the Go standard library whenever practical.

Do not add a dependency when the standard library is sufficient.

Every third-party dependency must have a clear justification.

Keep dependencies minimal and production-oriented.

Pin dependency versions.

Do not introduce experimental or abandoned libraries without explicit justification.

## 4. Architecture Rules

The project must follow:

Clean Architecture + Feature-Driven Architecture.

The dependency direction must be:

Delivery → Application → Domain

Infrastructure implements interfaces defined by inner layers.

The domain layer must not depend on:

* Gin
* MySQL driver
* Redis client
* JWT library
* HTTP framework
* ORM
* external infrastructure

Business logic must remain framework-independent.

## 5. Database Rules

Use MySQL.

Database access must be isolated behind repository interfaces.

Do not allow handlers/controllers to directly access the database.

Use:

* Connection pooling
* Context propagation
* Query timeout
* Transactions where required
* Proper indexes
* Prepared statements
* Pagination
* Migration management

Avoid N+1 queries.

Do not use SELECT * unless there is a concrete reason.

Database errors must be mapped into application-level errors.

## 6. Redis Rules

Use Redis only where it provides clear value.

Typical use cases:

* Caching
* Refresh token/session management
* Rate limiting
* Distributed locks when required

Redis must not become a hidden source of business state.

All Redis operations must support context and timeout.

Handle Redis unavailable scenarios explicitly.

Do not make the entire application unavailable merely because a non-critical cache is unavailable.

## 7. Authentication Rules

Implement JWT authentication with:

* Access Token
* Refresh Token
* Token expiration
* Secure token validation
* Token type validation
* Issuer validation
* Audience validation where applicable
* JWT signing algorithm validation

Never accept an arbitrary JWT signing algorithm from the token header.

JWT secrets must never be hardcoded.

Refresh token handling must support revocation.

Passwords must never be stored in plaintext.

Use a modern password hashing algorithm such as Argon2id or bcrypt.

## 8. Security Rules

Implement security by default.

At minimum consider:

* Request validation
* Authentication
* Authorization
* Rate limiting
* CORS configuration
* Security headers
* Request size limits
* Timeout handling
* Input sanitization where applicable
* Sensitive data masking
* Secure password hashing
* Secret management
* SQL injection prevention
* Log injection prevention

Never log:

* Passwords
* JWT secrets
* Refresh tokens
* Access tokens
* API secrets
* Database passwords

## 9. Error Handling

Use centralized application errors.

Errors must distinguish between:

* Validation errors
* Authentication errors
* Authorization errors
* Not found
* Conflict
* Business errors
* Database errors
* Infrastructure errors
* Internal errors

Do not expose internal stack traces or database errors to API consumers.

Use consistent API error responses.

## 10. Logging

Implement structured logging.

Logs should contain appropriate contextual information such as:

* Timestamp
* Level
* Request ID
* Trace ID when available
* HTTP method
* Path
* Status code
* Duration

Do not log sensitive information.

## 11. Observability

The boilerplate must be prepared for production observability.

Support:

* Health check
* Readiness check
* Liveness check
* Structured logging
* Request ID
* Metrics-ready architecture
* Graceful shutdown

Design the architecture so OpenTelemetry and Prometheus can be introduced without restructuring the application.

## 12. Configuration

Configuration must come from environment variables or configuration files.

Never hardcode:

* Database credentials
* Redis credentials
* JWT secrets
* API keys
* Environment-specific URLs

Provide:

* `.env.example`
* Environment configuration structure
* Validation for required configuration

Fail fast when mandatory configuration is missing.

## 13. Testing

Every business-critical component must be testable.

Implement:

* Unit tests
* Repository tests where practical
* Handler tests
* Middleware tests
* Authentication tests
* Validation tests

Use mocks/fakes only where appropriate.

Avoid excessive mocking.

Tests must be deterministic.

No test should depend on execution order.

The project must provide a coverage command.

Do not artificially inflate coverage.

Focus coverage on business logic and critical paths.

## 14. API Standards

Use RESTful API conventions.

Standardize:

* URL structure
* HTTP methods
* HTTP status codes
* Request format
* Response format
* Error format
* Pagination
* Filtering
* Sorting

Use API versioning, for example:

`/api/v1/...`

Do not expose internal implementation details through the API.

## 15. Documentation

Generate:

* README
* Architecture documentation
* Environment configuration documentation
* Local development instructions
* Docker instructions
* Testing instructions
* API documentation
* Authentication flow documentation

Documentation must reflect the actual implementation.

Never document functionality that does not exist.

## 16. Git Rules

Use conventional commit messages.

Examples:

* `feat: add authentication module`
* `feat: add user module`
* `fix: handle redis connection timeout`
* `test: add authentication unit tests`
* `refactor: simplify repository abstraction`
* `docs: update local development guide`

Do not mix unrelated changes into one commit.

## 17. Implementation Workflow

For every phase:

1. Inspect the current repository.
2. Explain the intended changes.
3. Implement only the requested scope.
4. Run formatting.
5. Run static analysis.
6. Run tests.
7. Run coverage when applicable.
8. Fix failures.
9. Verify the resulting structure.
10. Summarize exactly what changed.

Do not continue to the next phase automatically.

Wait for explicit approval before starting the next major phase.

## 18. Quality Gate

A phase is not complete until:

* Code compiles.
* Tests pass.
* `go vet` passes.
* Formatter has been applied.
* No obvious race conditions exist.
* Configuration is validated.
* No secrets are hardcoded.
* Architecture boundaries are respected.
* Documentation is updated where necessary.

If a quality gate fails, fix it before declaring the phase complete.

## 19. Avoid Overengineering

Do not implement prematurely:

* Event-driven architecture
* Kafka
* Kubernetes
* Service mesh
* CQRS
* Event sourcing
* Distributed transactions
* Complex domain frameworks

unless explicitly requested.

The boilerplate should be enterprise-ready without becoming unnecessarily complex.
