# Encryption & PKI Strategy

## Umumiy ko'rinish

XBank barcha sensitive ma'lumotlarni himoya qilish uchun **asymmetric (public/private key)** va **symmetric (AES)** encryption kombinatsiyasini ishlatadi. Kalitlar ierarxiyasi (Key Hierarchy) orqali boshqariladi.

## Encryption standarti

| Ma'lumot | Algorithm | Izoh |
|---|---|---|
| Karta raqami (PAN) | **Hybrid: ECDH + AES-256-GCM** | Har bir karta uchun unique DEK |
| KYC hujjatlar | **Envelope: RSA-4096 + AES-256-GCM** | DEK per-document, KEK rotate |
| KYC hujjat raqami | **AES-256-GCM** | Application-level, key Vault da |
| Parol | **bcrypt** (cost=12) | Hash, deshifrlash mumkin emas |
| PIN | **bcrypt** (cost=12) | Hash, deshifrlash mumkin emas |
| CVV | **bcrypt** (cost=12) | Hash, HECH QACHON plain text |
| Refresh token | **SHA-256** | One-way hash |
| Device fingerprint | **SHA-256** | One-way hash |
| Fayl integrity | **SHA-256** | Checksum |
| JWT signing | **ES256** (ECDSA P-256) | Asymmetric, key rotation |
| Transfer signing | **ECDSA P-256** (per-user) | Client private key, server public key |
| Auth challenge | **ECDSA P-256** | Challenge-response |
| API trafik | **TLS 1.3** | Transport encryption |
| DB connection | **SSL verify-full** | PostgreSQL SSL |
| Servislar arasi | **mTLS** | Kelajakda microservice uchun |

## Key Hierarchy (Kalit ierarxiyasi)

```
                ┌──────────────────────────┐
                │     HSM / Vault (KEK)    │
                │                          │
                │  Root KEK (master key)   │
                │          │               │
                │    ┌─────┼─────┐         │
                │    ▼     ▼     ▼         │
                │  JWT   Card   KYC        │
                │  KEK   KEK    KEK        │
                └────┬─────┬─────┬─────────┘
                     │     │     │
          ┌──────────┘     │     └──────────┐
          ▼                ▼                ▼
    ┌──────────┐    ┌──────────┐    ┌──────────┐
    │ JWT Keys │    │ Card DEK │    │ KYC DEK  │
    │ (ES256)  │    │ (per-card│    │(per-doc) │
    │          │    │  AES-256)│    │ AES-256) │
    └──────────┘    └──────────┘    └──────────┘

KEK = Key Encryption Key (kalitni shifrlash kaliti)
DEK = Data Encryption Key (ma'lumotni shifrlash kaliti)
```

### Qoidalar
- **KEK** faqat HSM/Vault da saqlanadi — DB da HECH QACHON
- **DEK** AES-256-GCM bilan ma'lumotni shifrlaydi
- **DEK** o'zi KEK bilan shifrlangan holda DB da saqlanadi
- KEK rotate bo'lganda faqat DEK lar qayta shifrlanadi (ma'lumotga tegmaydi)

## JWT Signing (ES256)

### Nima uchun RS256 emas, ES256?
| | RS256 | ES256 |
|---|---|---|
| Algorithm | RSA 2048-bit | ECDSA P-256 |
| Key size | 2048-bit | 256-bit |
| Signature size | 256 byte | 64 byte |
| Sign tezligi | Sekin | **10x tez** |
| Verify tezligi | Tez | Tez |
| NIST approved | Ha | **Ha** |
| JWT hajmi | Katta | **Kichik** |

### JWT Header
```json
{
  "alg": "ES256",
  "typ": "JWT",
  "kid": "jwt_2026_q2"
}
```

### JWKS Endpoint
```
GET /.well-known/jwks.json
```
```json
{
  "keys": [
    {
      "kty": "EC",
      "crv": "P-256",
      "kid": "jwt_2026_q2",
      "use": "sig",
      "x": "base64url...",
      "y": "base64url..."
    },
    {
      "kty": "EC",
      "crv": "P-256",
      "kid": "jwt_2026_q1",
      "use": "sig",
      "x": "base64url...",
      "y": "base64url...",
      "_status": "rotate_out"
    }
  ]
}
```

### Key Rotation Flow
```
1. Yangi key pair generatsiya (jwt_2026_q2)
2. Yangi key → ACTIVE, eski key → ROTATE_OUT
3. Yangi tokenlar → yangi key bilan sign
4. Eski tokenlar → eski key bilan verify (30 kun)
5. 30 kundan keyin eski key → RETIRED
```

## Transfer Signing (ECDSA per-user)

### Nima uchun HMAC emas, ECDSA?
```
HMAC (symmetric):
  Client va Server BIR XIL secret ni biladi
  Server buzilsa → barcha userlar compromise

ECDSA (asymmetric):
  Client → Private key bilan IMZOLAYDI
  Server → Public key bilan TEKSHIRADI
  Server HECH QACHON private key ni ko'rmaydi!
```

### Client Key Registration Flow
```
1. Ro'yxatdan o'tganda yoki yangi qurilmadan:
   Client → ECDSA P-256 key pair generatsiya (local)
   Client → POST /api/v1/auth/signing-keys
            { "public_key": "PEM...", "algorithm": "ES256" }

2. Server → public key ni DB ga saqlaydi (user_signing_keys)

3. Har bir transfer da:
   payload = "idempotency_key|from|to|amount|currency|timestamp"
   signature = ECDSA_Sign(private_key, SHA256(payload))
   Header: X-Signature: base64(signature)
   Header: X-Signing-Key-ID: "key_uuid"

4. Server → public key bilan verify:
   valid = ECDSA_Verify(user_public_key, SHA256(payload), signature)
```

### Go Implementation
```go
import (
    "crypto/ecdsa"
    "crypto/elliptic"
    "crypto/rand"
    "crypto/sha256"
)

// Server: signature tekshirish
func VerifyTransferSignature(publicKey *ecdsa.PublicKey, payload []byte, sig []byte) bool {
    hash := sha256.Sum256(payload)
    return ecdsa.VerifyASN1(publicKey, hash[:], sig)
}
```

## Card Encryption (Hybrid: ECDH + AES-256-GCM)

### Flow
```
SHIFRLASH (karta yaratish):
  1. Random DEK generatsiya (AES-256 key, har bir karta uchun yangi)
  2. card_number → AES-256-GCM(card_number, DEK, nonce) → encrypted_pan
  3. DEK → RSA_Encrypt(DEK, card_KEK_public) → encrypted_dek
  4. DB ga saqlash: encrypted_pan + encrypted_dek + nonce + key_id

DESHIFRLASH (karta o'qish):
  1. encrypted_dek → RSA_Decrypt(encrypted_dek, card_KEK_private) → DEK
     (card_KEK_private faqat Vault/HSM da!)
  2. encrypted_pan → AES_Decrypt(encrypted_pan, DEK, nonce) → card_number
```

### DB Schema
```sql
-- cards jadvalidagi encryption ustunlari:
encrypted_pan       BYTEA NOT NULL,        -- AES-256-GCM(pan, DEK)
encrypted_dek       BYTEA NOT NULL,        -- RSA(DEK, KEK_public)
encryption_nonce    BYTEA NOT NULL,        -- AES GCM nonce (12 byte)
encryption_key_id   VARCHAR(50) NOT NULL,  -- qaysi KEK ishlatilgan
pan_last_four       CHAR(4) NOT NULL,      -- masking uchun
pan_hash            VARCHAR(64) NOT NULL,  -- SHA-256 (qidiruv uchun)
```

### KEK Rotate bo'lganda
```
1. Yangi KEK pair generatsiya (card_kek_v4)
2. Barcha mavjud kartalar uchun:
   a. Eski KEK bilan DEK ni decrypt
   b. Yangi KEK bilan DEK ni re-encrypt (re-wrap)
   c. encrypted_dek va encryption_key_id yangilash
3. Karta ma'lumotining O'ZI qayta shifrlanmaydi — faqat DEK
4. Eski KEK → ROTATE_OUT (30 kun), keyin RETIRED
```

## KYC Document Encryption (Envelope)

### Flow
```
SHIFRLASH (hujjat yuklash):
  1. Random DEK generatsiya (har bir hujjat uchun yangi)
  2. document → AES-256-GCM(document, DEK, nonce) → encrypted_doc → S3/MinIO
  3. DEK → RSA_Encrypt(DEK, kyc_KEK_public) → encrypted_dek → DB
  4. DB ga: encrypted_dek + nonce + key_id + file_hash(SHA-256)

DESHIFRLASH (hujjat ko'rish):
  1. DB dan: encrypted_dek → RSA_Decrypt(kyc_KEK_private) → DEK
  2. S3 dan: encrypted_doc → AES_Decrypt(encrypted_doc, DEK, nonce) → document
  3. SHA-256(document) == file_hash tekshirish (integrity)
```

### Xavfsizlik
```
DB buzilsa     → DEK lar shifrlangan → hujjatlar xavfsiz
S3 buzilsa     → hujjatlar shifrlangan → xavfsiz
Ikkalasi ham   → KEK private Vault da → BARIBIR xavfsiz
```

## Key Rotation Jadvali

| Key | Algorithm | Rotation | Grace period | Saqlash |
|---|---|---|---|---|
| JWT signing | ES256 (P-256) | Har 90 kunda | 30 kun (verify only) | Vault |
| Transfer signing (per-user) | ECDSA P-256 | 365 kun yoki user so'rovi | 24 soat | Client: private, DB: public |
| Card KEK | RSA-4096 | 365 kunda | 30 kun (decrypt only) | Vault |
| KYC KEK | RSA-4096 | 365 kunda | 30 kun (decrypt only) | Vault |
| TLS sertifikat | X.509 | 365 kunda | Auto-renew (Let's Encrypt) | Vault |

## Encryption Keys DB Schema

```sql
CREATE TABLE encryption_keys (
    id              VARCHAR(50) PRIMARY KEY,  -- 'jwt_2026_q1', 'card_kek_v3'
    purpose         VARCHAR(30) NOT NULL,     -- JWT, CARD_KEK, KYC_KEK
    algorithm       VARCHAR(20) NOT NULL,     -- ES256, RSA-4096
    public_key_pem  TEXT NOT NULL,            -- PEM format
    -- private_key DB DA SAQLANMAYDI! → Vault / HSM
    key_version     INTEGER NOT NULL DEFAULT 1,
    status          VARCHAR(20) NOT NULL DEFAULT 'ACTIVE',
        -- ACTIVE, ROTATE_OUT, RETIRED
    activated_at    TIMESTAMPTZ NOT NULL,
    rotate_after    TIMESTAMPTZ NOT NULL,     -- qachon rotate qilish kerak
    retired_at      TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_enc_keys_purpose ON encryption_keys (purpose, status);

CREATE TABLE user_signing_keys (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users(id),
    device_id       VARCHAR(255) NOT NULL,
    algorithm       VARCHAR(20) NOT NULL DEFAULT 'ES256',
    public_key_pem  TEXT NOT NULL,
    status          VARCHAR(20) NOT NULL DEFAULT 'ACTIVE',
        -- ACTIVE, REVOKED
    activated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at      TIMESTAMPTZ NOT NULL,
    revoked_at      TIMESTAMPTZ,
    last_used_at    TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    UNIQUE(user_id, device_id, status)
);

CREATE INDEX idx_user_sign_keys ON user_signing_keys (user_id, status)
    WHERE status = 'ACTIVE';
```

## HSM / Vault integratsiya

```
HashiCorp Vault:
  ├── secret/xbank/jwt/            → JWT private keys
  ├── secret/xbank/card-kek/       → Card KEK private keys
  ├── secret/xbank/kyc-kek/        → KYC KEK private keys
  ├── transit/xbank/               → Encryption as a Service
  └── pki/xbank/                   → TLS sertifikatlar

Access Policy:
  auth-service   → jwt/* (read)
  card-service   → card-kek/* (read)
  kyc-service    → kyc-kek/* (read)
  admin          → * (rotate, create)
```

## PKI Arxitektura diagramma

```
                    ┌─────────────────────┐
                    │   HSM / Vault       │
                    │                     │
                    │  JWT Private Key    │
                    │  Card KEK Private   │
                    │  KYC KEK Private    │
                    └────────┬────────────┘
                             │
              ┌──────────────┼──────────────┐
              │              │              │
              ▼              ▼              ▼
        ┌──────────┐  ┌──────────┐  ┌──────────┐
        │   Auth   │  │   Card   │  │   KYC    │
        │ Service  │  │ Service  │  │ Service  │
        │          │  │          │  │          │
        │ JWT sign │  │ Hybrid   │  │ Envelope │
        │(private) │  │ encrypt  │  │ encrypt  │
        └──────────┘  └──────────┘  └──────────┘
              ▲
              │ verify (public key)
        ┌─────┴──────────────────────────────┐
        │        Barcha servislar             │
        │   JWKS endpoint dan public key     │
        │   olib JWT ni verify qiladi        │
        └────────────────────────────────────┘

        ┌──────────┐
        │  Client  │
        │ (Mobile) │
        │          │
        │ ECDSA    │
        │ private  │──→ transfer imzolaydi
        │ key      │
        └──────────┘
              │
              │ public key (ro'yxatdan o'tganda)
              ▼
        ┌──────────┐
        │ Transfer │
        │ Service  │──→ public key bilan verify
        └──────────┘
```

## API Endpoints

| Method | Endpoint | Middleware | Tavsif |
|---|---|---|---|
| POST | `/api/v1/auth/signing-keys` | Session | Client public key ro'yxatdan o'tkazish |
| DELETE | `/api/v1/auth/signing-keys/{id}` | Session+2FA | Signing key revoke |
| GET | `/api/v1/auth/signing-keys` | Session | Faol signing keys ro'yxati |
| GET | `/.well-known/jwks.json` | Public | JWT public keys (JWKS) |
