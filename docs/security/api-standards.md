# API Standards

## Base URL

```
http://localhost:3000/api/v1
```

## Authentication

All protected endpoints require a JWT Bearer token:

```
Authorization: Bearer <access_token>
```

- **Algorithm:** ES256 (ECDSA P-256)
- **Access token TTL:** 15 minutes
- **Refresh token TTL:** 30 days

### HMAC-Protected Endpoints

Transfer and Card operations additionally require HMAC-SHA256 request signing:

```
X-HMAC-Signature: <hmac_sha256(request_body + timestamp)>
X-HMAC-Timestamp: <unix_timestamp>
```

Clock skew tolerance: 5 minutes (configurable).

### Idempotency

Transfer endpoints support idempotency via:

```
X-Idempotency-Key: <unique_uuid>
```

TTL: 24 hours. Duplicate requests return the cached response.

## Response Format

### Success Response

```json
{
  "status": "success",
  "code": 200,
  "data": { ... },
  "meta": {
    "request_id": "550e8400-e29b-41d4-a716-446655440000",
    "correlation_id": "corr-uuid",
    "timestamp": "2026-04-10T10:00:00Z"
  }
}
```

### Error Response

```json
{
  "status": "error",
  "code": 3004,
  "message": "Invalid email or password",
  "data": null,
  "meta": {
    "request_id": "550e8400-e29b-41d4-a716-446655440000",
    "timestamp": "2026-04-10T10:00:00Z"
  }
}
```

### Paginated Response

```json
{
  "status": "success",
  "code": 200,
  "data": {
    "data": [ ... ],
    "pagination": {
      "page": 1,
      "limit": 20,
      "total": 150
    }
  },
  "meta": { ... }
}
```

## Pagination

Query parameters:

| Param | Default | Max | Description |
|-------|---------|-----|-------------|
| `page` | 1 | - | Page number (1-based) |
| `limit` | 20 | 100 | Items per page |

## Error Codes

Error codes are numeric and grouped by category:

| Range | Category | Example |
|-------|----------|---------|
| 1000-1999 | System | `1000` Internal server error |
| 2000-2999 | Validation | `2001` Missing required field |
| 3000-3999 | Authentication | `3004` Invalid credentials |
| 4000-4999 | Business Logic | `4001` Insufficient balance |

## HTTP Status Codes

| Status | Usage |
|--------|-------|
| 200 | Successful GET, PUT, POST (non-create) |
| 201 | Successful resource creation |
| 400 | Validation error, malformed request |
| 401 | Missing or invalid authentication |
| 403 | Insufficient permissions |
| 404 | Resource not found |
| 409 | Conflict (duplicate, state violation) |
| 429 | Rate limit exceeded |
| 500 | Internal server error |

## Rate Limiting

- **Global:** 60 requests per minute per IP
- **Login:** Stricter limits with progressive lockout
- **Custom rules:** Configurable per endpoint via Admin API

Response headers:
```
X-RateLimit-Limit: 60
X-RateLimit-Remaining: 45
X-RateLimit-Reset: 1680000000
```

## Content Type

- **Request:** `application/json`
- **Response:** `application/json`
- **Max body size:** 10 MB

## API Documentation

- **Swagger UI:** `GET /swagger/`
- **ReDoc:** `GET /docs`
- **OpenAPI spec:** `docs/swagger/swagger.yaml` (OpenAPI 3.0.3)

## Route Structure

```
/health              — Liveness probe
/health/live         — Liveness probe
/health/ready        — Readiness probe (dependency checks)
/metrics             — Prometheus metrics
/swagger/*           — Swagger UI
/docs                — ReDoc

/api/v1/auth/*       — Public: login, register, refresh, TOTP
/api/v1/currencies/* — Public: exchange rates

/api/v1/*            — Protected: user, account, transfer, card, etc.

/api/v1/admin/*      — Admin: RBAC + IP whitelist required
```

## Security Headers

Applied via Helmet middleware:

```
X-Content-Type-Options: nosniff
X-Frame-Options: DENY
X-XSS-Protection: 1; mode=block
Referrer-Policy: strict-origin-when-cross-origin
Content-Security-Policy: default-src 'self'
```
