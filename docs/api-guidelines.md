# API Design Guidelines

This document standardizes RESTful API design conventions, status code usage, response wrappers, pagination, rate limiting, and idempotency for this boilerplate.

---

## 1. URL & Endpoint Conventions

- Use lowercase, hyphen-separated kebab-case for resource paths (e.g. `/api/v1/user-roles`).
- Use plural nouns for resource collections (e.g. `/api/v1/articles`, `/api/v1/roles`).
- Standard CRUD URI mapping:

| Action | HTTP Method | URI Pattern | Status Code |
|---|---|---|---|
| Create resource | `POST` | `/api/v1/articles` | `201 Created` |
| List resources | `GET` | `/api/v1/articles` | `200 OK` |
| Get resource by ID | `GET` | `/api/v1/articles/:id` | `200 OK` |
| Update resource | `PUT` | `/api/v1/articles/:id` | `200 OK` |
| Delete resource | `DELETE` | `/api/v1/articles/:id` | `200 OK` |

---

## 2. Standard JSON Response Envelope

All API responses strictly conform to the standardized JSON envelope.

### Single Entity Response (`200 OK` / `201 Created`)
```json
{
  "success": true,
  "message": "Article retrieved successfully",
  "data": {
    "id": "123e4567-e89b-12d3-a456-426614174000",
    "title": "Clean Architecture in Go",
    "slug": "clean-architecture-in-go",
    "status": "published",
    "created_at": "2026-08-24T12:00:00Z",
    "updated_at": "2026-08-24T12:00:00Z"
  },
  "meta": null
}
```

### Paginated Collection Response (`200 OK`)
```json
{
  "success": true,
  "message": "Articles retrieved successfully",
  "data": [
    {
      "id": "123e4567-e89b-12d3-a456-426614174000",
      "title": "Clean Architecture in Go",
      "slug": "clean-architecture-in-go",
      "status": "published"
    }
  ],
  "meta": {
    "page": 1,
    "page_size": 20,
    "total_count": 100,
    "total_pages": 5
  }
}
```

### Error Response (`4xx` / `5xx`)
```json
{
  "success": false,
  "message": "Validation failed",
  "errors": [
    {
      "field": "title",
      "message": "title length must be 3-255 characters"
    }
  ],
  "data": null,
  "meta": null
}
```

---

## 3. Rate Limiting & Headers

All API endpoints are protected by Redis-backed distributed rate limiting:
- **General API Tier**: Default 100 requests per minute per IP / Authenticated User.
- **Authentication Tier**: Strict 10 requests per minute per IP on `/api/v1/auth/*` to prevent brute force.
- **Response Headers**:
  - `X-RateLimit-Limit`: Maximum permitted request quota in the current window.
  - `X-RateLimit-Remaining`: Number of requests remaining in the current window.
  - `X-RateLimit-Reset`: Unix timestamp when the current rate limit window resets.
  - `Retry-After`: (On HTTP 429) Number of seconds to wait before retrying.

---

## 4. Idempotency Support (`Idempotency-Key`)

Clients can safely retry state-mutating requests (`POST`, `PUT`, `PATCH`, `DELETE`) by providing a unique UUID in the `Idempotency-Key` header:
- **Behavior**:
  - If identical request payload is repeated within 24h $\rightarrow$ returns cached response directly with `X-Idempotent-Replay: true`.
  - If the same key is reused with a different payload body $\rightarrow$ returns `422 Unprocessable Entity` (`IDEMPOTENCY_CONFLICT`).
  - If duplicate concurrent request arrives while the first is processing $\rightarrow$ returns `409 Conflict` (`IDEMPOTENCY_IN_PROGRESS`).

---

## 5. Pagination Standards

- Query parameters: `?page=1&page_size=20`
- Defaults:
  - Default `page`: `1`
  - Default `page_size`: `20`
  - Maximum `page_size`: `100` (Enforced to prevent memory exhaustion)
- Implemented via `internal/shared/pagination.Extract(c)`.

---

## 6. Safe Sorting & Filtering

- Query parameters: `?sort=created_at&order=desc`
- Handlers validate sort parameters against explicit column allowlists.
- Implemented via `internal/shared/queryparam.ExtractSorting(c, allowlist, defaultField, defaultOrder)`.
