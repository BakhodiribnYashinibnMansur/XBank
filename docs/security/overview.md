# Security Overview

## Encryption
- **At rest**: AES-256-GCM (card data, KYC docs, TOTP secrets)
- **In transit**: TLS 1.2+ (production: TLS 1.3)
- **Key management**: Environment variable, key rotation support
- **DB level**: pgcrypto extension

## Hashing (one-way)
- **Parollar**: bcrypt, cost=12
- **CVV/PIN**: bcrypt (hech qachon plain)
- **Refresh token**: SHA-256
- **Sensitive fields**: logda maskalangan (`****1234`)

## Rate Limiting (Sliding Window, Redis)
```
Global:         10,000 req/sec
Per-IP:         100 req/min
/auth/login:    5 req/min
/transfer:      10 req/min
/auth/2fa:      3 req/min
```

## OWASP Top 10 Himoya
| Threat | Himoya |
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
Hech qachon loglanmaydi:
  password, card_number, cvv, pin, totp_secret, refresh_token
Logda faqat masked: ****1234, ***@***.com
Response'da card_number hech qachon to'liq qaytarilmaydi
```

## IP Whitelisting
Admin panel faqat belgilangan IP'lardan: `ADMIN_WHITELIST_IPS=192.168.1.0/24`

## Circuit Breaker
Tashqi servislar uchun: 5 failures → 30s open → half-open → test → close/open
