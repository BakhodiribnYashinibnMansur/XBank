# Auth & Session Management

## Token Strategy
```
Access Token:  JWT ES256 (ECDSA P-256), 15 min TTL, stateless
Refresh Token: opaque (random 32 byte), 30 days, SHA-256 hash stored in DB
Device Token:  device fingerprint hash
```

### Why ES256?
- **10x faster** signing than RS256
- **Smaller** key (256-bit vs 2048-bit) and signature (64 byte vs 256 byte)
- NIST approved, banking standard
- Native support in Go `crypto/ecdsa`

## JWT Header + Claims
```json
{
  "alg": "ES256",
  "typ": "JWT",
  "kid": "jwt_2026_q2"
}
```
```json
{
  "sub": "user_uuid",
  "iss": "xbank.uz",
  "sid": "session_uuid",
  "scope": ["transfer:create", "account:read"],
  "device_id": "d_abc123",
  "ip_hash": "sha256...",
  "risk_level": "low",
  "role": "customer"
}
```

## JWKS Endpoint (Public Keys)
```
GET /.well-known/jwks.json
```
All services fetch the JWT public key from this endpoint to verify.
Key rotation: every **90 days**, old key remains valid for verification only for 30 days.
Details: [Encryption & PKI](../security/encryption.md#jwt-signing-es256)

## Session Flow
1. **LOGIN** → password check → 2FA (if enabled) → Create session → Redis + PostgreSQL
2. **REQUEST** → JWT validate → Redis session check → continue/401
3. **REFRESH** → refresh token hash → DB lookup → rotate → new tokens
4. **LOGOUT** → Redis + PostgreSQL revoke → access token blacklist

## Refresh Token Theft Detection
If an old (rotated) refresh token is used:
- REVOKE ALL sessions (token theft!)
- Security alert notification is sent

## 2FA/MFA (TOTP — RFC 6238)
- Google Authenticator / Authy compatible
- Setup: `POST /api/v1/auth/2fa/setup` → QR code (otpauth:// URI)
- Verify: `POST /api/v1/auth/2fa/verify` → 6-digit code
- Sensitive operations: transfer > $1000, card issue, PIN change
- Backup codes: 10 one-time recovery codes
- 3 failed 2FA attempts = 30 min lock

## RBAC + ABAC

### Roles (RBAC)
| Role | Permissions |
|---|---|
| CUSTOMER | View own account, transfer |
| TELLER | View customer accounts, cash operations |
| MANAGER | Modify limits, approval |
| ADMIN | System settings |
| AUDITOR | Read-only (all data) |

### Attribute-Based (ABAC)
```
IF user.role == CUSTOMER
   AND resource.owner == user.id
   AND transaction.amount < user.daily_limit
   AND user.kyc_status == VERIFIED
   AND request.device IN user.trusted_devices
THEN ALLOW
```

## Transaction Signing (ECDSA Challenge-Response)
For critical operations (transfer > threshold):
1. Client → `POST /api/v1/auth/challenge` → `{nonce, expires_in: 120s}`
2. Client → `ECDSA_Sign(private_key, SHA256(nonce + payload))` → `X-Signature` header
3. Server → `ECDSA_Verify(user_public_key, ...)` + nonce one-time use
4. Server → execute operation

**Asymmetric:** The server only knows the public key — the private key is never on the network or on the server.
Details: [Encryption & PKI](../security/encryption.md#transfer-signing-ecdsa-per-user)

## Brute-Force Protection
- 5 failed login = 15 min lock
- 3 failed 2FA = 30 min lock
- 3 failed PIN = card blocked
- Progressive delay: 1s, 2s, 4s, 8s...

## API Endpoints

| Method | Endpoint | Middleware |
|---|---|---|
| POST | `/api/v1/auth/register` | RateLimit, CSRF |
| POST | `/api/v1/auth/login` | RateLimit |
| POST | `/api/v1/auth/refresh` | — |
| POST | `/api/v1/auth/logout` | Session |
| POST | `/api/v1/auth/logout-all` | Session |
| POST | `/api/v1/auth/2fa/setup` | Session |
| POST | `/api/v1/auth/2fa/verify` | Session |
| DELETE | `/api/v1/auth/2fa` | Session+2FA |
| POST | `/api/v1/auth/challenge` | Session |
| POST | `/api/v1/auth/signing-keys` | Session |
| DELETE | `/api/v1/auth/signing-keys/{id}` | Session+2FA |
| GET | `/api/v1/auth/signing-keys` | Session |
| GET | `/.well-known/jwks.json` | Public |
