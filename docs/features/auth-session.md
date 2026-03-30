# Auth & Session Management

## Token Strategy
```
Access Token:  JWT ES256 (ECDSA P-256), 15 min TTL, stateless
Refresh Token: opaque (random 32 byte), 30 kun, DB'da SHA-256 hash saqlanadi
Device Token:  qurilma fingerprint hash
```

### Nima uchun ES256?
- RS256 dan **10x tez** signing
- **Kichik** key (256-bit vs 2048-bit) va signature (64 byte vs 256 byte)
- NIST approved, bank standarti
- Go `crypto/ecdsa` da native support

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
Barcha servislar shu endpoint dan JWT public key ni olib verify qiladi.
Key rotation: har **90 kunda**, eski key 30 kun davomida faqat verify uchun ishlaydi.
Batafsil: [Encryption & PKI](../security/encryption.md#jwt-signing-es256)

## Session Flow
1. **LOGIN** → parol tekshirish → 2FA (agar yoqilgan) → Session yaratish → Redis + PostgreSQL
2. **REQUEST** → JWT validate → Redis session check → continue/401
3. **REFRESH** → refresh token hash → DB lookup → rotate → yangi tokenlar
4. **LOGOUT** → Redis + PostgreSQL revoke → access token blacklist

## Refresh Token Theft Detection
Agar eski (rotated) refresh token ishlatilsa:
- BARCHA sessionlarni revoke (token theft!)
- Security alert notification yuboriladi

## 2FA/MFA (TOTP — RFC 6238)
- Google Authenticator / Authy compatible
- Setup: `POST /api/v1/auth/2fa/setup` → QR code (otpauth:// URI)
- Verify: `POST /api/v1/auth/2fa/verify` → 6-digit code
- Sensitive operatsiyalar: transfer > $1000, card issue, PIN change
- Backup codes: 10 ta one-time recovery codes
- 3 muvaffaqiyatsiz 2FA = 30 min lock

## RBAC + ABAC

### Roles (RBAC)
| Role | Huquqlar |
|---|---|
| CUSTOMER | O'z hisobini ko'rish, transfer |
| TELLER | Mijoz hisoblarini ko'rish, cash operations |
| MANAGER | Limitlarni o'zgartirish, approval |
| ADMIN | Tizim sozlamalari |
| AUDITOR | Faqat o'qish (barcha ma'lumotlar) |

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
Muhim operatsiyalar uchun (transfer > threshold):
1. Client → `POST /api/v1/auth/challenge` → `{nonce, expires_in: 120s}`
2. Client → `ECDSA_Sign(private_key, SHA256(nonce + payload))` → `X-Signature` header
3. Server → `ECDSA_Verify(user_public_key, ...)` + nonce one-time use
4. Server → execute operation

**Asymmetric:** Server faqat public key ni biladi — private key hech qachon tarmoqda yoki serverda bo'lmaydi.
Batafsil: [Encryption & PKI](../security/encryption.md#transfer-signing-ecdsa-per-user)

## Brute-Force Protection
- 5 failed login = 15 min lock
- 3 failed 2FA = 30 min lock
- 3 failed PIN = karta bloklanadi
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
