# Encryption & PKI Strategy

## Umumiy ko'rinish

XBank barcha sensitive ma'lumotlarni himoya qilish uchun **asymmetric (public/private key)** va **symmetric (AES)** encryption kombinatsiyasini ishlatadi. Kalitlar ierarxiyasi (Key Hierarchy) orqali boshqariladi.

## Encryption standarti

| Ma'lumot | Algorithm | Izoh |
|---|---|---|
| Karta raqami (PAN) | **E2EE: ECIES (P-256) + AES-256-GCM** | Client da encrypt, faqat Vault/HSM da decrypt |
| Card PIN | **E2EE: ECIES + ISO 9564 PIN Block** | Faqat HSM decrypt qiladi, server ko'rmaydi |
| Card CVV | **E2EE: ECIES (P-256)** | Faqat verify uchun, HECH QACHON saqlanmaydi |
| KYC hujjatlar | **E2EE: ECIES + Envelope (AES-256-GCM)** | Client da encrypt, Vault decrypt, DEK per-document |
| KYC hujjat raqami | **E2EE: ECIES (P-256)** | Client da encrypt, Vault da decrypt |
| Parol | **bcrypt** (cost=12) | Hash, deshifrlash mumkin emas |
| PIN | **bcrypt** (cost=12) | Hash, deshifrlash mumkin emas |
| CVV | **bcrypt** (cost=12) | Hash, HECH QACHON plain text |
| Refresh token | **SHA-256** | One-way hash |
| Device fingerprint | **SHA-256** | One-way hash |
| Fayl integrity | **SHA-256** | Checksum |
| JWT signing | **ES256** (ECDSA P-256) | Asymmetric, key rotation |
| Transfer signing | **ECDSA P-256** (per-user) | Client private key, server public key |
| Auth challenge | **ECDSA P-256** | Challenge-response |
| API trafik | **TLS 1.3** | Transport encryption (E2EE ustiga) |
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
┌─ 1. Client keypair generatsiya qiladi (LOCAL, server ko'rmaydi) ─────┐
│                                                                        │
│   // Mobile/Web da                                                     │
│   private_key, public_key = ECDSA_Generate(P-256)                     │
│                                                                        │
│   private_key → Secure storage:                                       │
│                 iOS: Secure Enclave (hardware)                        │
│                 Android: StrongBox / TEE (hardware)                   │
│                 Web: SubtleCrypto (non-exportable CryptoKey)          │
│                                                                        │
│   public_key → server ga yuboriladi                                   │
│                (ochiq, xavfsiz — bilsa ham foydasiz)                  │
└────────────────────────────────────────────────────────────────────────┘

┌─ 2. Authenticated session orqali public key yuborish ────────────────┐
│                                                                        │
│   ⚠️ Faqat LOGIN qilgan user yuborishi mumkin (JWT token kerak)       │
│                                                                        │
│   POST /api/v1/auth/signing-keys                                      │
│   Headers:                                                             │
│     Authorization: Bearer <JWT>          ← kim yuborayotgani ma'lum   │
│     X-Device-Fingerprint: "abc123..."    ← qaysi qurilma             │
│   Body:                                                                │
│   {                                                                    │
│     "public_key": "-----BEGIN PUBLIC KEY-----\nMFkw...",              │
│     "algorithm": "ES256",                                             │
│     "device_name": "iPhone 15 Pro"                                    │
│   }                                                                    │
└────────────────────────────────────────────────────────────────────────┘

┌─ 3. Server public key ni saqlaydi ───────────────────────────────────┐
│                                                                        │
│   Server:                                                              │
│     1. JWT dan user_id ni oladi (kim?)                                │
│     2. Device fingerprint tekshiradi (qaysi qurilma?)                 │
│     3. public_key ni DB ga saqlaydi:                                  │
│                                                                        │
│        user_signing_keys:                                              │
│        ┌──────────────┬──────────┬────────────┬────────┐              │
│        │ user_id      │ device_id│ public_key │ status │              │
│        ├──────────────┼──────────┼────────────┼────────┤              │
│        │ user-uuid-1  │ iphone15 │ PEM...     │ ACTIVE │              │
│        │ user-uuid-1  │ web-chr  │ PEM...     │ ACTIVE │              │
│        └──────────────┴──────────┴────────────┴────────┘              │
│                                                                        │
│     4. Response: { "key_id": "key-uuid", "status": "ACTIVE" }        │
└────────────────────────────────────────────────────────────────────────┘

┌─ 4. Keyinchalik: transfer imzolash ─────────────────────────────────┐
│                                                                        │
│   Client:                                                              │
│     payload = "idemp_key|from|to|amount|currency|timestamp"           │
│     signature = ECDSA_Sign(private_key, SHA256(payload))              │
│                            ^^^^^^^^^^^                                │
│                            faqat client biladi                        │
│                                                                        │
│   POST /api/v1/transfers                                              │
│   Headers:                                                             │
│     Authorization: Bearer <JWT>                                       │
│     X-Signature: base64(signature)                                    │
│     X-Signing-Key-ID: "key-uuid"                                     │
│                                                                        │
│   Server:                                                              │
│     public_key = DB dan olish (key-uuid bo'yicha)                    │
│     valid = ECDSA_Verify(public_key, SHA256(payload), signature)     │
│     valid == true  → transfer davom etadi                             │
│     valid == false → 403 Forbidden                                    │
└────────────────────────────────────────────────────────────────────────┘
```

### Nima Uchun Bu Ishonchli?

```
Savol: MITM public key ni almashtirsa-chi?

Javob: MUMKIN EMAS, chunki:

  1. TLS 1.3 — transport himoyalangan, MITM o'rtada turolmaydi
  2. JWT — faqat login qilgan user yuborishi mumkin
  3. Device fingerprint — faqat shu qurilmadan
  4. 2FA — yangi qurilma uchun 2FA talab qilinadi

  Hatto TLS buzilsa:
    MITM o'zining public key ni yubordi → server saqladi
    Lekin MITM client ning private key ni BILMAYDI
    → MITM o'zi transfer imzolay olmaydi (o'zining private key bilan
      imzolasa, server da saqlangan public key bilan MATCH bo'lmaydi)
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

## Card Encryption (E2EE: ECIES + AES-256-GCM)

### Flow
```
SHIFRLASH (client tomonida — mobile/web):
  1. Server public key olish: GET /api/v1/crypto/public-key
  2. ECIES shifrlash (client da):
     ephemeral_key  = ECDH_Generate(P-256)
     shared_secret  = ECDH(ephemeral_private, server_public)
     derived_key    = HKDF-SHA256(shared_secret, salt, "xbank-e2ee-v1")
     encrypted_pan  = AES-256-GCM(card_number, derived_key, nonce)
  3. POST /api/v1/cards { encrypted_pan: { ciphertext, ephemeral_pub, nonce, salt } }

SERVER (plaintext ko'rmaydi):
  1. encrypted_pan ni Vault/HSM ga proxy qiladi
  2. Vault: ECIES decrypt → plaintext PAN
  3. Vault: AES-256-GCM(PAN, card_DEK) → re-encrypt for storage
  4. DB ga saqlash: re-encrypted_pan + encrypted_dek + nonce + key_id
  5. pan_last_four va pan_hash (SHA-256) hisoblash — Vault ichida

DESHIFRLASH (karta o'qish):
  1. Vault: encrypted_dek → decrypt → DEK
  2. Vault: AES_Decrypt(encrypted_pan, DEK, nonce) → card_number
  3. Plaintext faqat Vault ichida — server memory ga CHIQMAYDI
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

## KYC Document Encryption (E2EE: ECIES + Envelope)

### Flow
```
SHIFRLASH (client tomonida — mobile/web):
  1. Server public key olish: GET /api/v1/crypto/public-key
  2. Client ECIES bilan hujjatni shifrlaydi:
     ephemeral_key = ECDH_Generate(P-256)
     shared_secret = ECDH(ephemeral_private, server_public)
     derived_key   = HKDF-SHA256(shared_secret, salt, "xbank-e2ee-v1")
     encrypted_doc = AES-256-GCM(document, derived_key, nonce)
  3. POST /api/v1/kyc/documents { encrypted_document: { ciphertext, ephemeral_pub, nonce, salt } }

SERVER (plaintext ko'rmaydi):
  1. encrypted_doc ni Vault ga proxy
  2. Vault: ECIES decrypt → plaintext document
  3. Vault: Random DEK generatsiya (har bir hujjat uchun yangi)
  4. Vault: AES-256-GCM(document, DEK) → re-encrypted_doc
  5. Vault: RSA_Encrypt(DEK, kyc_KEK) → encrypted_dek
  6. re-encrypted_doc → S3/MinIO, encrypted_dek → DB
  7. file_hash = SHA-256(document) — Vault ichida hisoblash

DESHIFRLASH (hujjat ko'rish):
  1. Vault: encrypted_dek → RSA_Decrypt(kyc_KEK_private) → DEK
  2. S3 dan: re-encrypted_doc → Vault ga yuborish
  3. Vault: AES_Decrypt(re_encrypted_doc, DEK, nonce) → document
  4. Vault: SHA-256(document) == file_hash tekshirish (integrity)
  5. Plaintext faqat Vault ichida
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
        │ JWT sign │  │ E2EE     │  │ E2EE     │
        │(private) │  │ ECIES    │  │ ECIES    │
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

## End-to-End Encryption (E2EE)

<!-- E2EE — sensitive ma'lumotlar client da shifrlanadi,
     faqat HSM/Vault da deshifrlanadi.
     Application server HECH QACHON plaintext ko'rmaydi.
     Bu TLS ustiga qo'shimcha himoya qatlami. -->

### ECDH Key Exchange — Kalit Almashish Protokoli

<!-- ECDH (Elliptic Curve Diffie-Hellman) — asimmetrik kalitlardan
     simmetrik kalit hosil qilish usuli.
     Signal Protocol, TLS 1.3, va WhatsApp E2EE ning asosi. -->

#### Asosiy Printsip

```
Client va Server hech qachon SECRET ni tarmoq orqali YUBORMAYDI.
Har biri o'z private key + boshqaning public key = BIR XIL shared secret.

Matematika (Elliptic Curve):
  Client:  ephemeral_private × Server_public    = SHARED SECRET
  Server:  server_private    × Ephemeral_public = SHARED SECRET  (BIR XIL!)

  Sababi:  a × (b × G) = b × (a × G)   ← Elliptic Curve matematik xossasi
           ^^^^^^^^^^^   ^^^^^^^^^^^
           client         server
           hisoblaydi     hisoblaydi

  G = Generator point (P-256 curve da oldindan belgilangan nuqta)
  × = Elliptic Curve point multiplication (oddiy ko'paytirish emas!)
```

#### To'liq Key Exchange Flow

```
                    CLIENT                                      SERVER
                    ══════                                      ══════

  ┌─ 1. Server public key ni olish ────────────────────────────────────────┐
  │                                                                         │
  │   GET /api/v1/crypto/public-key  ─────────────────────→                │
  │                                                          Vault dan     │
  │                                                          public key    │
  │                                   ←─────────────────────               │
  │                                   {                                    │
  │                                     "key_id": "e2ee_2026_q2",         │
  │                                     "public_key": "PEM...",           │
  │                                     "algorithm": "ECIES-P256"         │
  │                                   }                                    │
  └────────────────────────────────────────────────────────────────────────┘

  ┌─ 2. Client: Ephemeral keypair generatsiya (HAR SAFAR YANGI) ───────────┐
  │                                                                         │
  │   ephemeral_private = random()                                         │
  │                       ^^^^^^^^                                         │
  │                       Client da qoladi — HECH QACHON YUBORILMAYDI     │
  │                                                                         │
  │   ephemeral_public = ephemeral_private × G                             │
  │                      ^^^^^^^^^^^^^^^^^^                                │
  │                      Server ga yuboriladi (ochiq, xavfsiz)             │
  └────────────────────────────────────────────────────────────────────────┘

  ┌─ 3. Client: Shared secret hisoblash ───────────────────────────────────┐
  │                                                                         │
  │   shared_secret = ephemeral_private × server_public_key                │
  │                   ^^^^^^^^^^^^^^^^^   ^^^^^^^^^^^^^^^^^                 │
  │                   faqat client        ochiq (API dan olgan)             │
  │                   biladi                                                │
  │                                                                         │
  │   ⚠️ shared_secret TARMOQDA YUBORILAMAYDI — hisoblangan qiymat        │
  └────────────────────────────────────────────────────────────────────────┘

  ┌─ 4. Client: Simmetrik kalit hosil qilish (HKDF) ──────────────────────┐
  │                                                                         │
  │   salt = random_32_bytes                                               │
  │                                                                         │
  │   aes_key = HKDF-SHA256(                                               │
  │     ikm:  shared_secret,        ← ECDH natijasi                       │
  │     salt: salt,                 ← random (server ga yuboriladi)        │
  │     info: "xbank-e2ee-v1",     ← context string (hardcoded)           │
  │     len:  32                    ← 256 bit = AES-256 key               │
  │   )                                                                    │
  │                                                                         │
  │   Endi AES-256-GCM bilan shifrlash mumkin!                             │
  └────────────────────────────────────────────────────────────────────────┘

  ┌─ 5. Client → Server: shifrlangan ma'lumot yuborish ────────────────────┐
  │                                                                         │
  │   nonce = random_12_bytes                                              │
  │   ciphertext = AES-256-GCM(plaintext, aes_key, nonce)                 │
  │                                                                         │
  │   POST /api/v1/cards  ────────────────────────────→                    │
  │   {                                                                     │
  │     "encrypted_pan": {                                                 │
  │       "ciphertext": base64(ciphertext),           ← shifrlangan       │
  │       "ephemeral_public_key": base64(eph_pub),    ← faqat PUBLIC key  │
  │       "nonce": base64(nonce),                     ← AES-GCM nonce     │
  │       "salt": base64(salt),                       ← HKDF salt         │
  │       "key_id": "e2ee_2026_q2"                    ← qaysi server key  │
  │     }                                                                   │
  │   }                                                                     │
  │                                                                         │
  │   ⚠️ Yuborilgan: ephemeral_public, ciphertext, nonce, salt            │
  │   ⚠️ YUBORILMAGAN: ephemeral_private, shared_secret, aes_key          │
  └────────────────────────────────────────────────────────────────────────┘

  ┌─ 6. Server (Vault): BIR XIL shared secret hisoblash ───────────────────┐
  │                                                                         │
  │   Vault ichida (server application ko'rmaydi):                         │
  │                                                                         │
  │   shared_secret = server_private × ephemeral_public_key                │
  │                   ^^^^^^^^^^^^^^   ^^^^^^^^^^^^^^^^^^^^                 │
  │                   faqat Vault      client yubordi                      │
  │                   biladi                                                │
  │                                                                         │
  │   aes_key = HKDF-SHA256(shared_secret, salt, "xbank-e2ee-v1")         │
  │                                                                         │
  │   plaintext = AES-256-GCM_Decrypt(ciphertext, aes_key, nonce)         │
  │                                                                         │
  │   ✅ Client va Server BIR XIL aes_key ni hisobladi!                    │
  │   ✅ Chunki: eph_priv × srv_pub = srv_priv × eph_pub                  │
  └────────────────────────────────────────────────────────────────────────┘
```

#### Tarmoqda Nima Ko'rinadi / Ko'rinmaydi

```
Hacker tarmoqni tinglayapti (MITM yoki packet sniffing):

  ┌─────────────────────────────────────────────────────────────────┐
  │  KO'RA OLADI (public/ochiq)      │  KO'RA OLMAYDI (secret)    │
  │──────────────────────────────────│─────────────────────────────│
  │  ✅ server_public_key            │  ❌ server_private_key      │
  │     (API dan ochiq olinadi)      │     (faqat Vault da)        │
  │                                   │                             │
  │  ✅ ephemeral_public_key         │  ❌ ephemeral_private_key   │
  │     (request ichida yuboriladi)  │     (client memory, keyin   │
  │                                   │      o'chiriladi)           │
  │                                   │                             │
  │  ✅ ciphertext                   │  ❌ shared_secret           │
  │     (shifrlangan, foydasiz)      │     (hisoblangan, tarmoqda  │
  │                                   │      hech qachon o'tmagan)  │
  │                                   │                             │
  │  ✅ nonce, salt                  │  ❌ aes_key                 │
  │     (key siz foydasiz)           │     (HKDF natijasi)         │
  │                                   │                             │
  │                                   │  ❌ plaintext              │
  │                                   │     (original ma'lumot)     │
  └─────────────────────────────────────────────────────────────────┘

  Hacker ikkita public key ni ko'radi:
    server_public  = srv_priv × G
    ephemeral_public = eph_priv × G

  shared_secret = srv_priv × eph_priv × G  ni hisoblash uchun
  srv_priv YOKI eph_priv ni bilish kerak.

  Public key dan private key ni topish = ECDLP
  (Elliptic Curve Discrete Logarithm Problem)

  P-256 uchun: ~2^128 operatsiya kerak
  = Hozirgi eng kuchli superkompyuter bilan MILLIARDLAB yillar
```

#### Forward Secrecy (Oldinga Maxfiylik)

```
Har bir request uchun YANGI ephemeral keypair generatsiya qilinadi:

  Request 1: eph_key_A → shared_secret_A → aes_key_A → encrypt(PAN_1)
  Request 2: eph_key_B → shared_secret_B → aes_key_B → encrypt(PAN_2)
  Request 3: eph_key_C → shared_secret_C → aes_key_C → encrypt(KYC_doc)

  eph_key_A, B, C — har biri random, bir-biriga BOG'LIQ EMAS.
  Request tugagach ephemeral_private_key memory dan O'CHIRILADI.

Natija:
  ┌─────────────────────────────────────────────────────────────┐
  │  Ssenariy                      │  Xavfsizlik              │
  │────────────────────────────────│──────────────────────────│
  │  Bitta aes_key buzilsa         │  Faqat SHU request       │
  │                                │  ochiladi. Boshqalar      │
  │                                │  XAVFSIZ.                 │
  │                                │                           │
  │  server_private_key leak       │  OLDINGI requestlarni     │
  │  bo'lsa (Vault buzilsa)        │  deshifrlab BO'LMAYDI.   │
  │                                │  Chunki eph_private lar   │
  │                                │  allaqachon o'chirilgan.  │
  │                                │                           │
  │  Kelajakdagi requestlar?       │  Key rotate qilinadi →   │
  │                                │  yangi server keypair.    │
  └─────────────────────────────────────────────────────────────┘

  Bu Signal Protocol ning asosiy xossasi:
  "Compromise one session → does NOT compromise past sessions"
```

#### Signal Protocol bilan Taqqoslash

```
Signal Protocol (X3DH + Double Ratchet):
  - Ikki FOYDALANUVCHI o'rtasida (peer-to-peer messaging)
  - Identity key + Signed pre-key + One-time pre-key
  - Double Ratchet: har bir xabar uchun yangi key
  - Maqsad: real-time chat encryption

XBank E2EE (ECIES):
  - CLIENT → SERVER o'rtasida (client-to-HSM)
  - Server static key + Client ephemeral key
  - Har bir request uchun yangi ephemeral key (forward secrecy)
  - Maqsad: sensitive financial data encryption

  ┌────────────────────────────────────────────────────────────┐
  │                    │  Signal X3DH        │  XBank ECIES    │
  │────────────────────│─────────────────────│─────────────────│
  │  Tomonlar          │  User ↔ User        │  Client → HSM   │
  │  Key exchange      │  X3DH (3 ECDH)      │  1 ECDH         │
  │  Forward secrecy   │  ✅ Double Ratchet  │  ✅ Ephemeral   │
  │  Key rotation      │  Har bir xabar      │  Har bir request│
  │  Murakkablik       │  Yuqori             │  O'rtacha        │
  │  Use case          │  Messaging          │  Financial data  │
  └────────────────────────────────────────────────────────────┘

  XBank uchun Signal to'liq kerak emas:
  - Biz user↔user chat qilmaymiz
  - Biz client→server sensitive data yuboramiz
  - ECIES = Signal ning key exchange qismining soddalashtirilgan versiyasi
```

### Nima uchun TLS yetarli emas?

```
Faqat TLS:
  Client ──TLS──→ Load Balancer ──plaintext──→ App Server ──plaintext──→ memory
                        ↑                          ↑                       ↑
                   TLS terminate              plaintext log ga           memory dump
                   (reverse proxy)            tushib qolishi mumkin      attack

E2EE + TLS:
  Client ──encrypt──→ TLS ──→ App Server ──ciphertext──→ HSM/Vault ──decrypt
                                    ↑                         ↑
                              plaintext YO'Q              faqat shu yerda
                              server memory da            deshifrlanadi
```

### E2EE Arxitekturasi: Client-to-HSM

```
┌──────────────────────────────────────────────────────────────────────┐
│                        CLIENT (Mobile/Web)                          │
│                                                                      │
│  1. Server public key ni olish:                                      │
│     GET /api/v1/crypto/public-key → ECDH public key (P-256)         │
│                                                                      │
│  2. Sensitive field ni shifrlash (ECIES):                            │
│     ephemeral_key = ECDH_GenerateKeypair()                          │
│     shared_secret = ECDH(ephemeral_private, server_public)          │
│     derived_key   = HKDF-SHA256(shared_secret, salt, info)          │
│     ciphertext    = AES-256-GCM(plaintext, derived_key, nonce)      │
│                                                                      │
│  3. Request yuborish:                                                │
│     {                                                                │
│       "account_id": "acc-uuid",          ← plaintext (TLS only)     │
│       "amount": 100000,                  ← plaintext (TLS only)     │
│       "encrypted_pan": {                 ← E2EE                     │
│         "ciphertext": "base64...",                                   │
│         "ephemeral_public_key": "base64...",                         │
│         "nonce": "base64...",                                        │
│         "key_id": "e2ee_2026_q2"                                    │
│       }                                                              │
│     }                                                                │
└──────────────────────────────────────────────────────────────────────┘
         │
         │ TLS 1.3 (transport)
         ▼
┌──────────────────────────────────────────────────────────────────────┐
│                      APPLICATION SERVER                              │
│                                                                      │
│  Server encrypted_pan ni OCHMAYDI — ciphertext holida                │
│  Vault/HSM ga proxy qiladi:                                         │
│                                                                      │
│  result = vault.Decrypt("transit/xbank/e2ee", ciphertext, nonce,    │
│                          ephemeral_public_key)                       │
│                                                                      │
│  yoki                                                                │
│                                                                      │
│  Server ciphertext ni to'g'ridan-to'g'ri DB ga saqlaydi             │
│  (karta PAN, KYC hujjat — qayta ishlash kerak emas)                 │
└──────────────────────────────────────────────────────────────────────┘
         │
         ▼
┌──────────────────────────────────────────────────────────────────────┐
│                        HSM / VAULT                                   │
│                                                                      │
│  1. Ephemeral public key + server private key → shared_secret        │
│  2. HKDF-SHA256(shared_secret) → derived_key                        │
│  3. AES-256-GCM_Decrypt(ciphertext, derived_key, nonce)             │
│  4. Plaintext → qayta ishlash → natija                              │
│  5. Plaintext memory dan DARHOL tozalash                            │
│                                                                      │
│  Private key HECH QACHON Vault dan tashqariga chiqmaydi!            │
└──────────────────────────────────────────────────────────────────────┘
```

### ECIES (Elliptic Curve Integrated Encryption Scheme)

<!-- ECIES = ECDH key agreement + HKDF key derivation + AES-GCM encryption
     Har bir xabar uchun yangi ephemeral key → forward secrecy -->

```
ECIES shifrlash jarayoni:

  1. Ephemeral keypair generatsiya:
     ephemeral_private, ephemeral_public = ECDH_Generate(P-256)

  2. Shared secret hisoblash:
     shared_secret = ECDH(ephemeral_private, server_public_key)

  3. Key derivation:
     encryption_key = HKDF-SHA256(
       ikm:  shared_secret,
       salt: random_32_bytes,
       info: "xbank-e2ee-v1",
       len:  32  // AES-256 key
     )

  4. Shifrlash:
     nonce = random_12_bytes  // AES-GCM nonce
     ciphertext, tag = AES-256-GCM_Encrypt(plaintext, encryption_key, nonce)

  5. Natija:
     {
       ciphertext:          ciphertext || tag,   // encrypted data + auth tag
       ephemeral_public_key: ephemeral_public,   // ECDH uchun
       nonce:               nonce,               // AES-GCM nonce
       salt:                salt,                // HKDF salt
       key_id:              "e2ee_2026_q2"       // qaysi server key ishlatilgan
     }
```

### Nima uchun ECIES?

| Xususiyat | RSA-OAEP | ECIES |
|---|---|---|
| Key size | 4096-bit | 256-bit (P-256) |
| Max plaintext | ~446 byte (RSA-4096) | **Cheksiz** (hybrid) |
| Forward secrecy | ❌ Yo'q | ✅ Ha (ephemeral key) |
| Performance | Sekin | **Tez** |
| Mobile battery | Ko'p sarflaydi | **Kam** |

### Qaysi Fieldlar E2EE Talab Qiladi

| Field | E2EE | Sabab |
|---|---|---|
| Card PAN (16 raqam) | ✅ **Majburiy** | PCI DSS — server memory da plaintext bo'lmasligi kerak |
| Card PIN | ✅ **Majburiy** | ISO 9564 — faqat HSM decrypt qiladi |
| Card CVV | ✅ **Majburiy** | PCI DSS — hech qachon saqlanmaydi, faqat tekshirish |
| KYC hujjat (passport, selfie) | ✅ **Majburiy** | Shaxsiy ma'lumot — GDPR, mahalliy qonunchilik |
| KYC hujjat raqami | ✅ **Majburiy** | PII (Personally Identifiable Information) |
| Transfer amount | ❌ Yo'q | Server balance tekshirishi kerak — plaintext zarur |
| Account ID | ❌ Yo'q | Server routing qilishi kerak |
| OTP / 2FA code | ❌ Yo'q | Server verify qiladi, TLS yetarli |
| Login credentials | ⚠️ SRP yoki TLS | Parol server ga plaintext kelmaydi (bcrypt client da emas) |

### PIN Block (ISO 9564)

<!-- PIN faqat HSM da decrypt bo'ladi — server HECH QACHON plaintext PIN ko'rmaydi -->

```
PIN shifrlash (ATM/POS/Mobile):

  Format 0 (ISO 9564-1):
    1. PIN block = PIN length || PIN || padding (F)
       Misol: PIN = "1234"
       PIN block = 0x04 1234 FFFFFFFFFF

    2. PAN block = 0000 || PAN[3..14]  (oxirgi 12 raqam, check digit siz)
       PAN = 4000001234567890
       PAN block = 0x0000 000123456789

    3. Clear PIN block = PIN block XOR PAN block

    4. Encrypted PIN block = AES-256-GCM(clear_pin_block, pin_encryption_key)
       yoki
       3DES_Encrypt(clear_pin_block, ZPK)  // legacy terminal lar uchun

Mobile/Web uchun:
    1. Client → ECIES(PIN, server_e2ee_public_key) → encrypted_pin
    2. Server → Vault ga yuborish (plaintext ko'rmaydi)
    3. Vault → decrypt → PIN hash (bcrypt) → saqlash
    4. Keyingi verify: Vault da bcrypt.Compare()
```

### E2EE Key Management

```
Vault Secrets:
  secret/xbank/e2ee/
    ├── current          → { key_id: "e2ee_2026_q2", private_key: PEM }
    ├── e2ee_2026_q2     → { private_key: PEM, status: ACTIVE }
    ├── e2ee_2026_q1     → { private_key: PEM, status: ROTATE_OUT }
    └── e2ee_2025_q4     → { private_key: PEM, status: RETIRED }
```

### E2EE Key Rotation

```
1. Yangi ECDH P-256 keypair generatsiya (Vault ichida)
   vault write transit/xbank/e2ee/keys type=ecdsa-p256

2. Public key ni API orqali e'lon qilish
   GET /api/v1/crypto/public-key
   → { key_id: "e2ee_2026_q2", public_key: PEM, algorithm: "ECIES-P256" }

3. Client yangi key_id bilan shifrlashni boshlaydi

4. Eski key → ROTATE_OUT (90 kun — eski client lar uchun decrypt davom etadi)

5. 90 kundan keyin → RETIRED (faqat arxivdagi ma'lumotlar uchun)
```

### Go Implementation

```go
import (
    "crypto/ecdh"
    "crypto/rand"
    "crypto/sha256"
    "golang.org/x/crypto/hkdf"
)

// Client side: ECIES shifrlash
func ECIESEncrypt(plaintext []byte, serverPubKey *ecdh.PublicKey) (*EncryptedPayload, error) {
    // 1. Ephemeral keypair
    curve := ecdh.P256()
    ephPriv, err := curve.GenerateKey(rand.Reader)
    if err != nil {
        return nil, err
    }

    // 2. ECDH shared secret
    sharedSecret, err := ephPriv.ECDH(serverPubKey)
    if err != nil {
        return nil, err
    }

    // 3. HKDF key derivation
    salt := make([]byte, 32)
    rand.Read(salt)
    hkdfReader := hkdf.New(sha256.New, sharedSecret, salt, []byte("xbank-e2ee-v1"))
    aesKey := make([]byte, 32) // AES-256
    hkdfReader.Read(aesKey)

    // 4. AES-256-GCM encrypt
    block, _ := aes.NewCipher(aesKey)
    gcm, _ := cipher.NewGCM(block)
    nonce := make([]byte, gcm.NonceSize())
    rand.Read(nonce)
    ciphertext := gcm.Seal(nil, nonce, plaintext, nil)

    return &EncryptedPayload{
        Ciphertext:         ciphertext,
        EphemeralPublicKey: ephPriv.PublicKey().Bytes(),
        Nonce:              nonce,
        Salt:               salt,
        KeyID:              "e2ee_2026_q2",
    }, nil
}

// EncryptedPayload — E2EE shifrlangan ma'lumot
type EncryptedPayload struct {
    Ciphertext         []byte `json:"ciphertext"`
    EphemeralPublicKey []byte `json:"ephemeral_public_key"`
    Nonce              []byte `json:"nonce"`
    Salt               []byte `json:"salt"`
    KeyID              string `json:"key_id"`
}
```

```go
// Server side: Vault ga proxy qilish (server plaintext ko'rmaydi)
func (s *CryptoService) DecryptE2EE(ctx context.Context, payload *EncryptedPayload) ([]byte, error) {
    // Server o'zi decrypt QILMAYDI — Vault ga yuboradi
    result, err := s.vault.Logical().WriteWithContext(ctx,
        "transit/xbank/e2ee/decrypt",
        map[string]interface{}{
            "ciphertext":          base64.StdEncoding.EncodeToString(payload.Ciphertext),
            "ephemeral_public_key": base64.StdEncoding.EncodeToString(payload.EphemeralPublicKey),
            "nonce":               base64.StdEncoding.EncodeToString(payload.Nonce),
            "salt":                base64.StdEncoding.EncodeToString(payload.Salt),
            "key_id":              payload.KeyID,
        },
    )
    if err != nil {
        return nil, fmt.Errorf("vault decrypt failed: %w", err)
    }

    plaintext, _ := base64.StdEncoding.DecodeString(result.Data["plaintext"].(string))
    return plaintext, nil
}
```

### E2EE Request/Response Formati

```
Karta qo'shish (E2EE bilan):

POST /api/v1/cards
{
  "account_id": "acc-uuid",                    ← plaintext (server kerak)
  "cardholder_name": "BAKHODIR YASHINI",       ← plaintext
  "encrypted_pan": {                           ← E2EE
    "ciphertext": "base64...",
    "ephemeral_public_key": "base64...",
    "nonce": "base64...",
    "salt": "base64...",
    "key_id": "e2ee_2026_q2"
  },
  "encrypted_pin": {                           ← E2EE
    "ciphertext": "base64...",
    "ephemeral_public_key": "base64...",
    "nonce": "base64...",
    "salt": "base64...",
    "key_id": "e2ee_2026_q2"
  },
  "encrypted_cvv": {                           ← E2EE
    "ciphertext": "base64...",
    "ephemeral_public_key": "base64...",
    "nonce": "base64...",
    "salt": "base64...",
    "key_id": "e2ee_2026_q2"
  }
}

Response:
{
  "status": "success",
  "data": {
    "card_id": "card-uuid",
    "last_four": "7890",
    "cardholder_name": "BAKHODIR YASHINI",
    "status": "ACTIVE"
  }
}
```

### E2EE + Existing Encryption Orasidagi Farq

```
Hozirgi (server-side encryption):
  Client ──plaintext──→ TLS ──→ Server ──AES-GCM encrypt──→ DB
                                  ↑
                            plaintext server
                            memory da bor!

E2EE (client-side encryption):
  Client ──ECIES encrypt──→ TLS ──→ Server ──ciphertext──→ Vault decrypt
                                      ↑                        ↑
                                plaintext YO'Q           faqat Vault da
                                server memory da         plaintext bor

Natija: Ikki qatlam himoya
  1. TLS 1.3     → transport (MITM dan)
  2. ECIES (E2EE) → application (server compromise dan)
```

## API Endpoints

| Method | Endpoint | Middleware | Tavsif |
|---|---|---|---|
| GET | `/api/v1/crypto/public-key` | Public | E2EE server public key (ECIES P-256) |
| POST | `/api/v1/auth/signing-keys` | Session | Client public key ro'yxatdan o'tkazish |
| DELETE | `/api/v1/auth/signing-keys/{id}` | Session+2FA | Signing key revoke |
| GET | `/api/v1/auth/signing-keys` | Session | Faol signing keys ro'yxati |
| GET | `/.well-known/jwks.json` | Public | JWT public keys (JWKS) |
