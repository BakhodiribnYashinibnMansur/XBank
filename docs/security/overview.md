# Security Overview

## Encryption
- **At rest**: AES-256-GCM (card data, KYC docs, TOTP secrets)
- **In transit**: TLS 1.2+ (production: TLS 1.3)
- **Key management**: Environment variable, key rotation support
- **DB level**: pgcrypto extension

## Hashing (one-way)
- **Passwords**: bcrypt, cost=12
- **CVV/PIN**: bcrypt (never plain)
- **Refresh token**: SHA-256
- **Sensitive fields**: masked in logs (`****1234`)

## Rate Limiting (Sliding Window, Redis)
```
Global:         10,000 req/sec
Per-IP:         100 req/min
/auth/login:    5 req/min
/transfer:      10 req/min
/auth/2fa:      3 req/min
```

## OWASP Top 10 Protection
| Threat | Protection |
|---|---|
| SQL Injection | Parameterized queries ($1) |
| XSS | JSON only + CSP headers |
| Broken Auth | JWT + 2FA + session management |
| Sensitive Data | AES-256, bcrypt, TLS, masking |
| Security Misconfig | Helmet, CORS, env-based config |
| CSRF | CSRF token middleware |
| Broken Access | RLS + RBAC/ABAC middleware |

## Sensitive Data Protection
```
Never logged:
  password, card_number, cvv, pin, totp_secret, refresh_token
Only masked in logs: ****1234, ***@***.com
card_number is never returned in full in responses
```

## IP Whitelisting
Admin panel is only accessible from designated IPs: `ADMIN_WHITELIST_IPS=192.168.1.0/24`

## Circuit Breaker
For external services: 5 failures → 30s open → half-open → test → close/open
