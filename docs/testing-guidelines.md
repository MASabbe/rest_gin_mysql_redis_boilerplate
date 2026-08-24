# Testing Guidelines

This document establishes the testing strategy, standards, and patterns for this boilerplate.

---

## 1. Testing Strategy & Pyramid

1. **Domain Unit Tests (`*_test.go` in `domain/entity`)**:
   - Test pure business validation, invariant enforcement, state transitions, and edge cases.
   - Zero mocking required. Fast execution.
2. **Application Use Case Tests (`service_test.go` in `application`)**:
   - Test use case execution, validation flows, error propagation, and business logic.
   - Tested against thread-safe in-memory repository doubles located in `infrastructure/persistence/inmemory_*_repository.go`.
3. **Delivery & Handler Tests (`*_handler_test.go` in `delivery/http/handler`)**:
   - Use `gin.CreateTestContext` and `httptest.NewRecorder` to verify HTTP status codes, headers, parameter extraction, and JSON response bodies.
   - Use mock token/authorization services to test authentication and permission enforcement.
4. **Middleware Tests (`middleware_test.go` in `internal/shared/middleware`)**:
   - Verify request ID propagation, logger context, panic recovery, CORS headers, and timeout cancellation.

---

## 2. In-Memory Repository Pattern for Unit Tests

Avoid complex mocking frameworks (like `gomock` or string-matching sqlmock) when testing application services. Instead, use clean, thread-safe in-memory repository implementations that implement the domain repository interfaces:

```go
// Example in application/service_test.go
repo := persistence.NewInMemoryArticleRepository()
svc := application.NewArticleService(repo)

// Run deterministic use case assertions
article, err := svc.CreateArticle(ctx, command.CreateArticleCommand{...})
assert.NoError(t, err)
```

---

## 3. Test Execution Commands

Run all tests:
```bash
make test
# Or: go test -v ./...
```

Run test suite with race detector:
```bash
make test-race
# Or: go test -v -race ./...
```

Run test coverage report:
```bash
make test-cover
# Or: go test -v -coverprofile=coverage.out ./... && go tool cover -func=coverage.out
```
