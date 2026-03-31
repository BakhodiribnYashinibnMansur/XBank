# Encryption & PKI Strategy

## Overview

XBank uses a combination of **asymmetric (public/private key)** and **symmetric (AES)** encryption to protect all sensitive data. Keys are managed through a Key Hierarchy.

## Encryption Standard

| Data | Algorithm | Notes |
|---|---|---|
| Card number (PAN) | **E2EE: ECIES (P-256) + AES-256-GCM** | Encrypted on client, decrypted only in Vault/HSM |
| Card PIN | **E2EE: ECIES + ISO 9564 PIN Block** | Only HSM decrypts, server never sees it |
| Card CVV | **E2EE: ECIES (P-256)** | Only for verification, NEVER stored |
| KYC documents | **E2EE: ECIES + Envelope (AES-256-GCM)** | Encrypted on client, Vault decrypts, DEK per-document |
| KYC document number | **E2EE: ECIES (P-256)** | Encrypted on client, decrypted in Vault |
| Password | **bcrypt** (cost=12) | Hash, cannot be decrypted |
| PIN | **bcrypt** (cost=12) | Hash, cannot be decrypted |
| CVV | **bcrypt** (cost=12) | Hash, NEVER plain text |
| Refresh token | **SHA-256** | One-way hash |
| Device fingerprint | **SHA-256** | One-way hash |
| File integrity | **SHA-256** | Checksum |
| JWT signing | **ES256** (ECDSA P-256) | Asymmetric, key rotation |
| Transfer signing | **ECDSA P-256** (per-user) | Client private key, server public key |
| Auth challenge | **ECDSA P-256** | Challenge-response |
| API traffic | **TLS 1.3** | Transport encryption (on top of E2EE) |
| DB connection | **SSL verify-full** | PostgreSQL SSL |
| Inter-service | **mTLS** | For future microservice use |

## Key Hierarchy

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

KEK = Key Encryption Key (key that encrypts other keys)
DEK = Data Encryption Key (key that encrypts data)
```

### Rules
- **KEK** is stored only in HSM/Vault — NEVER in DB
- **DEK** encrypts data with AES-256-GCM
- **DEK** itself is stored encrypted with KEK in the DB
- When KEK is rotated, only DEKs are re-encrypted (data is not touched)

## JWT Signing (ES256)

### Why ES256 instead of RS256?
| | RS256 | ES256 |
|---|---|---|
| Algorithm | RSA 2048-bit | ECDSA P-256 |
| Key size | 2048-bit | 256-bit |
| Signature size | 256 byte | 64 byte |
| Sign speed | Slow | **10x faster** |
| Verify speed | Fast | Fast |
| NIST approved | Yes | **Yes** |
| JWT size | Large | **Small** |

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
1. Generate new key pair (jwt_2026_q2)
2. New key → ACTIVE, old key → ROTATE_OUT
3. New tokens → signed with new key
4. Old tokens → verified with old key (30 days)
5. After 30 days, old key → RETIRED
```

## Transfer Signing (ECDSA per-user)

### Why ECDSA instead of HMAC?
```
HMAC (symmetric):
  Client and Server know the SAME secret
  If Server is compromised → all users are compromised

ECDSA (asymmetric):
  Client → SIGNS with private key
  Server → VERIFIES with public key
  Server NEVER sees the private key!
```

### Client Key Registration Flow

```
┌─ 1. Client generates keypair (LOCAL, server never sees it) ────────┐
│                                                                        │
│   // On Mobile/Web                                                     │
│   private_key, public_key = ECDSA_Generate(P-256)                     │
│                                                                        │
│   private_key → Secure storage:                                       │
│                 iOS: Secure Enclave (hardware)                        │
│                 Android: StrongBox / TEE (hardware)                   │
│                 Web: SubtleCrypto (non-exportable CryptoKey)          │
│                                                                        │
│   public_key → sent to server                                         │
│                (public, safe — useless even if known)                  │
└────────────────────────────────────────────────────────────────────────┘

┌─ 2. Send public key via authenticated session ──────────────────────┐
│                                                                        │
│   Only a LOGGED IN user can send it (JWT token required)              │
│                                                                        │
│   POST /api/v1/auth/signing-keys                                      │
│   Headers:                                                             │
│     Authorization: Bearer <JWT>          ← identifies who is sending  │
│     X-Device-Fingerprint: "abc123..."    ← which device               │
│   Body:                                                                │
│   {                                                                    │
│     "public_key": "-----BEGIN PUBLIC KEY-----\nMFkw...",              │
│     "algorithm": "ES256",                                             │
│     "device_name": "iPhone 15 Pro"                                    │
│   }                                                                    │
└────────────────────────────────────────────────────────────────────────┘

┌─ 3. Server stores the public key ────────────────────────────────────┐
│                                                                        │
│   Server:                                                              │
│     1. Gets user_id from JWT (who?)                                   │
│     2. Verifies device fingerprint (which device?)                    │
│     3. Stores public_key in DB:                                       │
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

┌─ 4. Later: signing a transfer ──────────────────────────────────────┐
│                                                                        │
│   Client:                                                              │
│     payload = "idemp_key|from|to|amount|currency|timestamp"           │
│     signature = ECDSA_Sign(private_key, SHA256(payload))              │
│                            ^^^^^^^^^^^                                │
│                            only the client knows it                   │
│                                                                        │
│   POST /api/v1/transfers                                              │
│   Headers:                                                             │
│     Authorization: Bearer <JWT>                                       │
│     X-Signature: base64(signature)                                    │
│     X-Signing-Key-ID: "key-uuid"                                     │
│                                                                        │
│   Server:                                                              │
│     public_key = fetched from DB (by key-uuid)                        │
│     valid = ECDSA_Verify(public_key, SHA256(payload), signature)     │
│     valid == true  → transfer proceeds                                │
│     valid == false → 403 Forbidden                                    │
└────────────────────────────────────────────────────────────────────────┘
```

### Why Is This Trustworthy?

```
Question: What if MITM replaces the public key?

Answer: IMPOSSIBLE, because:

  1. TLS 1.3 — transport is protected, MITM cannot intercept
  2. JWT — only a logged-in user can send it
  3. Device fingerprint — only from this device
  4. 2FA — 2FA is required for new devices

  Even if TLS were broken:
    MITM sent their own public key → server stored it
    But MITM does NOT KNOW the client's private key
    → MITM cannot sign transfers themselves (if they sign
      with their own private key, it won't MATCH the
      public key stored on the server)
```

### Go Implementation
```go
import (
    "crypto/ecdsa"
    "crypto/elliptic"
    "crypto/rand"
    "crypto/sha256"
)

// Server: verify signature
func VerifyTransferSignature(publicKey *ecdsa.PublicKey, payload []byte, sig []byte) bool {
    hash := sha256.Sum256(payload)
    return ecdsa.VerifyASN1(publicKey, hash[:], sig)
}
```

## Card Encryption (E2EE: ECIES + AES-256-GCM)

### Flow
```
ENCRYPTION (client-side — mobile/web):
  1. Get server public key: GET /api/v1/crypto/public-key
  2. ECIES encryption (on client):
     ephemeral_key  = ECDH_Generate(P-256)
     shared_secret  = ECDH(ephemeral_private, server_public)
     derived_key    = HKDF-SHA256(shared_secret, salt, "xbank-e2ee-v1")
     encrypted_pan  = AES-256-GCM(card_number, derived_key, nonce)
  3. POST /api/v1/cards { encrypted_pan: { ciphertext, ephemeral_pub, nonce, salt } }

SERVER (never sees plaintext):
  1. Proxies encrypted_pan to Vault/HSM
  2. Vault: ECIES decrypt → plaintext PAN
  3. Vault: AES-256-GCM(PAN, card_DEK) → re-encrypt for storage
  4. Store in DB: re-encrypted_pan + encrypted_dek + nonce + key_id
  5. Compute pan_last_four and pan_hash (SHA-256) — inside Vault

DECRYPTION (reading a card):
  1. Vault: encrypted_dek → decrypt → DEK
  2. Vault: AES_Decrypt(encrypted_pan, DEK, nonce) → card_number
  3. Plaintext only inside Vault — NEVER leaves to server memory
```

### DB Schema
```sql
-- Encryption columns in the cards table:
encrypted_pan       BYTEA NOT NULL,        -- AES-256-GCM(pan, DEK)
encrypted_dek       BYTEA NOT NULL,        -- RSA(DEK, KEK_public)
encryption_nonce    BYTEA NOT NULL,        -- AES GCM nonce (12 byte)
encryption_key_id   VARCHAR(50) NOT NULL,  -- which KEK was used
pan_last_four       CHAR(4) NOT NULL,      -- for masking
pan_hash            VARCHAR(64) NOT NULL,  -- SHA-256 (for search)
```

### When KEK Is Rotated
```
1. Generate new KEK pair (card_kek_v4)
2. For all existing cards:
   a. Decrypt DEK with old KEK
   b. Re-encrypt DEK with new KEK (re-wrap)
   c. Update encrypted_dek and encryption_key_id
3. The card data ITSELF is not re-encrypted — only the DEK
4. Old KEK → ROTATE_OUT (30 days), then RETIRED
```

## KYC Document Encryption (E2EE: ECIES + Envelope)

### Flow
```
ENCRYPTION (client-side — mobile/web):
  1. Get server public key: GET /api/v1/crypto/public-key
  2. Client encrypts the document with ECIES:
     ephemeral_key = ECDH_Generate(P-256)
     shared_secret = ECDH(ephemeral_private, server_public)
     derived_key   = HKDF-SHA256(shared_secret, salt, "xbank-e2ee-v1")
     encrypted_doc = AES-256-GCM(document, derived_key, nonce)
  3. POST /api/v1/kyc/documents { encrypted_document: { ciphertext, ephemeral_pub, nonce, salt } }

SERVER (never sees plaintext):
  1. Proxies encrypted_doc to Vault
  2. Vault: ECIES decrypt → plaintext document
  3. Vault: Generate random DEK (new for each document)
  4. Vault: AES-256-GCM(document, DEK) → re-encrypted_doc
  5. Vault: RSA_Encrypt(DEK, kyc_KEK) → encrypted_dek
  6. re-encrypted_doc → S3/MinIO, encrypted_dek → DB
  7. file_hash = SHA-256(document) — computed inside Vault

DECRYPTION (viewing a document):
  1. Vault: encrypted_dek → RSA_Decrypt(kyc_KEK_private) → DEK
  2. From S3: re-encrypted_doc → sent to Vault
  3. Vault: AES_Decrypt(re_encrypted_doc, DEK, nonce) → document
  4. Vault: SHA-256(document) == file_hash check (integrity)
  5. Plaintext only inside Vault
```

### Security
```
If DB is compromised     → DEKs are encrypted → documents are safe
If S3 is compromised     → documents are encrypted → safe
If both are compromised  → KEK private is in Vault → STILL safe
```

## Key Rotation Schedule

| Key | Algorithm | Rotation | Grace period | Storage |
|---|---|---|---|---|
| JWT signing | ES256 (P-256) | Every 90 days | 30 days (verify only) | Vault |
| Transfer signing (per-user) | ECDSA P-256 | 365 days or user request | 24 hours | Client: private, DB: public |
| Card KEK | RSA-4096 | Every 365 days | 30 days (decrypt only) | Vault |
| KYC KEK | RSA-4096 | Every 365 days | 30 days (decrypt only) | Vault |
| TLS certificate | X.509 | Every 365 days | Auto-renew (Let's Encrypt) | Vault |

## Encryption Keys DB Schema

```sql
CREATE TABLE encryption_keys (
    id              VARCHAR(50) PRIMARY KEY,  -- 'jwt_2026_q1', 'card_kek_v3'
    purpose         VARCHAR(30) NOT NULL,     -- JWT, CARD_KEK, KYC_KEK
    algorithm       VARCHAR(20) NOT NULL,     -- ES256, RSA-4096
    public_key_pem  TEXT NOT NULL,            -- PEM format
    -- private_key is NEVER STORED IN DB! → Vault / HSM
    key_version     INTEGER NOT NULL DEFAULT 1,
    status          VARCHAR(20) NOT NULL DEFAULT 'ACTIVE',
        -- ACTIVE, ROTATE_OUT, RETIRED
    activated_at    TIMESTAMPTZ NOT NULL,
    rotate_after    TIMESTAMPTZ NOT NULL,     -- when rotation is needed
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

## HSM / Vault Integration

```
HashiCorp Vault:
  ├── secret/xbank/jwt/            → JWT private keys
  ├── secret/xbank/card-kek/       → Card KEK private keys
  ├── secret/xbank/kyc-kek/        → KYC KEK private keys
  ├── transit/xbank/               → Encryption as a Service
  └── pki/xbank/                   → TLS certificates

Access Policy:
  auth-service   → jwt/* (read)
  card-service   → card-kek/* (read)
  kyc-service    → kyc-kek/* (read)
  admin          → * (rotate, create)
```

## PKI Architecture Diagram

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
        │        All services                 │
        │   Get public key from JWKS         │
        │   endpoint and verify JWT          │
        └────────────────────────────────────┘

        ┌──────────┐
        │  Client  │
        │ (Mobile) │
        │          │
        │ ECDSA    │
        │ private  │──→ signs transfers
        │ key      │
        └──────────┘
              │
              │ public key (during registration)
              ▼
        ┌──────────┐
        │ Transfer │
        │ Service  │──→ verifies with public key
        └──────────┘
```

## End-to-End Encryption (E2EE)

<!-- E2EE — sensitive data is encrypted on the client,
     decrypted only in HSM/Vault.
     The application server NEVER sees plaintext.
     This is an additional layer of protection on top of TLS. -->

### ECDH Key Exchange — Key Agreement Protocol

<!-- ECDH (Elliptic Curve Diffie-Hellman) — a method for deriving
     a symmetric key from asymmetric keys.
     The basis of Signal Protocol, TLS 1.3, and WhatsApp E2EE. -->

#### Core Principle

```
Client and Server NEVER SEND the SECRET over the network.
Each one: own private key + other's public key = SAME shared secret.

Mathematics (Elliptic Curve):
  Client:  ephemeral_private × Server_public    = SHARED SECRET
  Server:  server_private    × Ephemeral_public = SHARED SECRET  (THE SAME!)

  Reason:  a × (b × G) = b × (a × G)   ← Mathematical property of Elliptic Curves
           ^^^^^^^^^^^   ^^^^^^^^^^^
           client         server
           computes       computes

  G = Generator point (a predefined point on the P-256 curve)
  × = Elliptic Curve point multiplication (not ordinary multiplication!)
```

#### Full Key Exchange Flow

```
                    CLIENT                                      SERVER
                    ══════                                      ══════

  ┌─ 1. Get the server public key ─────────────────────────────────────┐
  │                                                                         │
  │   GET /api/v1/crypto/public-key  ─────────────────────→                │
  │                                                          From Vault    │
  │                                                          public key    │
  │                                   ←─────────────────────               │
  │                                   {                                    │
  │                                     "key_id": "e2ee_2026_q2",         │
  │                                     "public_key": "PEM...",           │
  │                                     "algorithm": "ECIES-P256"         │
  │                                   }                                    │
  └────────────────────────────────────────────────────────────────────────┘

  ┌─ 2. Client: Generate ephemeral keypair (NEW EACH TIME) ───────────────┐
  │                                                                         │
  │   ephemeral_private = random()                                         │
  │                       ^^^^^^^^                                         │
  │                       Stays on client — NEVER SENT                     │
  │                                                                         │
  │   ephemeral_public = ephemeral_private × G                             │
  │                      ^^^^^^^^^^^^^^^^^^                                │
  │                      Sent to server (public, safe)                     │
  └────────────────────────────────────────────────────────────────────────┘

  ┌─ 3. Client: Compute shared secret ─────────────────────────────────────┐
  │                                                                         │
  │   shared_secret = ephemeral_private × server_public_key                │
  │                   ^^^^^^^^^^^^^^^^^   ^^^^^^^^^^^^^^^^^                 │
  │                   only client         public (from API)                 │
  │                   knows it                                              │
  │                                                                         │
  │   shared_secret is NEVER SENT over the network — it is a computed value│
  └────────────────────────────────────────────────────────────────────────┘

  ┌─ 4. Client: Derive symmetric key (HKDF) ──────────────────────────────┐
  │                                                                         │
  │   salt = random_32_bytes                                               │
  │                                                                         │
  │   aes_key = HKDF-SHA256(                                               │
  │     ikm:  shared_secret,        ← ECDH result                         │
  │     salt: salt,                 ← random (sent to server)              │
  │     info: "xbank-e2ee-v1",     ← context string (hardcoded)           │
  │     len:  32                    ← 256 bit = AES-256 key               │
  │   )                                                                    │
  │                                                                         │
  │   Now encryption with AES-256-GCM is possible!                         │
  └────────────────────────────────────────────────────────────────────────┘

  ┌─ 5. Client → Server: send encrypted data ──────────────────────────────┐
  │                                                                         │
  │   nonce = random_12_bytes                                              │
  │   ciphertext = AES-256-GCM(plaintext, aes_key, nonce)                 │
  │                                                                         │
  │   POST /api/v1/cards  ────────────────────────────→                    │
  │   {                                                                     │
  │     "encrypted_pan": {                                                 │
  │       "ciphertext": base64(ciphertext),           ← encrypted          │
  │       "ephemeral_public_key": base64(eph_pub),    ← only PUBLIC key    │
  │       "nonce": base64(nonce),                     ← AES-GCM nonce     │
  │       "salt": base64(salt),                       ← HKDF salt         │
  │       "key_id": "e2ee_2026_q2"                    ← which server key  │
  │     }                                                                   │
  │   }                                                                     │
  │                                                                         │
  │   SENT: ephemeral_public, ciphertext, nonce, salt                      │
  │   NOT SENT: ephemeral_private, shared_secret, aes_key                  │
  └────────────────────────────────────────────────────────────────────────┘

  ┌─ 6. Server (Vault): Compute THE SAME shared secret ────────────────────┐
  │                                                                         │
  │   Inside Vault (server application never sees it):                     │
  │                                                                         │
  │   shared_secret = server_private × ephemeral_public_key                │
  │                   ^^^^^^^^^^^^^^   ^^^^^^^^^^^^^^^^^^^^                 │
  │                   only Vault       client sent it                       │
  │                   knows it                                              │
  │                                                                         │
  │   aes_key = HKDF-SHA256(shared_secret, salt, "xbank-e2ee-v1")         │
  │                                                                         │
  │   plaintext = AES-256-GCM_Decrypt(ciphertext, aes_key, nonce)         │
  │                                                                         │
  │   Client and Server computed THE SAME aes_key!                         │
  │   Because: eph_priv × srv_pub = srv_priv × eph_pub                    │
  └────────────────────────────────────────────────────────────────────────┘
```

#### What Is Visible / Not Visible on the Network

```
A hacker is sniffing the network (MITM or packet sniffing):

  ┌─────────────────────────────────────────────────────────────────┐
  │  CAN SEE (public/open)            │  CANNOT SEE (secret)       │
  │──────────────────────────────────│─────────────────────────────│
  │  server_public_key               │  server_private_key         │
  │     (openly obtained from API)   │     (only in Vault)         │
  │                                   │                             │
  │  ephemeral_public_key            │  ephemeral_private_key      │
  │     (sent inside the request)    │     (in client memory,      │
  │                                   │      then deleted)          │
  │                                   │                             │
  │  ciphertext                      │  shared_secret              │
  │     (encrypted, useless)         │     (computed, never        │
  │                                   │      passed over network)   │
  │                                   │                             │
  │  nonce, salt                     │  aes_key                    │
  │     (useless without key)        │     (HKDF result)           │
  │                                   │                             │
  │                                   │  plaintext                 │
  │                                   │     (original data)         │
  └─────────────────────────────────────────────────────────────────┘

  The hacker sees two public keys:
    server_public  = srv_priv × G
    ephemeral_public = eph_priv × G

  To compute shared_secret = srv_priv × eph_priv × G
  they need to know either srv_priv OR eph_priv.

  Deriving private key from public key = ECDLP
  (Elliptic Curve Discrete Logarithm Problem)

  For P-256: ~2^128 operations needed
  = BILLIONS of years with today's most powerful supercomputer
```

#### Forward Secrecy

```
A NEW ephemeral keypair is generated for EACH request:

  Request 1: eph_key_A → shared_secret_A → aes_key_A → encrypt(PAN_1)
  Request 2: eph_key_B → shared_secret_B → aes_key_B → encrypt(PAN_2)
  Request 3: eph_key_C → shared_secret_C → aes_key_C → encrypt(KYC_doc)

  eph_key_A, B, C — each is random, NOT RELATED to each other.
  After the request, ephemeral_private_key is DELETED from memory.

Result:
  ┌─────────────────────────────────────────────────────────────┐
  │  Scenario                        │  Security                │
  │────────────────────────────────│──────────────────────────│
  │  If one aes_key is broken        │  Only THAT request is    │
  │                                  │  exposed. Others are     │
  │                                  │  SAFE.                   │
  │                                  │                          │
  │  If server_private_key leaks     │  PREVIOUS requests       │
  │  (Vault is compromised)          │  CANNOT be decrypted.    │
  │                                  │  Because eph_private     │
  │                                  │  keys are already        │
  │                                  │  deleted.                │
  │                                  │                          │
  │  Future requests?                │  Key is rotated →        │
  │                                  │  new server keypair.     │
  └─────────────────────────────────────────────────────────────┘

  This is the core property of Signal Protocol:
  "Compromise one session → does NOT compromise past sessions"
```

#### Comparison with Signal Protocol

```
Signal Protocol (X3DH + Double Ratchet):
  - Between two USERS (peer-to-peer messaging)
  - Identity key + Signed pre-key + One-time pre-key
  - Double Ratchet: new key for each message
  - Purpose: real-time chat encryption

XBank E2EE (ECIES):
  - Between CLIENT → SERVER (client-to-HSM)
  - Server static key + Client ephemeral key
  - New ephemeral key for each request (forward secrecy)
  - Purpose: sensitive financial data encryption

  ┌────────────────────────────────────────────────────────────┐
  │                    │  Signal X3DH        │  XBank ECIES    │
  │────────────────────│─────────────────────│─────────────────│
  │  Parties           │  User ↔ User        │  Client → HSM   │
  │  Key exchange      │  X3DH (3 ECDH)      │  1 ECDH         │
  │  Forward secrecy   │  Double Ratchet     │  Ephemeral      │
  │  Key rotation      │  Every message      │  Every request  │
  │  Complexity        │  High               │  Medium          │
  │  Use case          │  Messaging          │  Financial data  │
  └────────────────────────────────────────────────────────────┘

  Full Signal is not needed for XBank:
  - We don't do user↔user chat
  - We send sensitive data from client→server
  - ECIES = a simplified version of Signal's key exchange part
```

### Why TLS Alone Is Not Enough?

```
TLS only:
  Client ──TLS──→ Load Balancer ──plaintext──→ App Server ──plaintext──→ memory
                        ↑                          ↑                       ↑
                   TLS terminate              plaintext could           memory dump
                   (reverse proxy)            end up in logs            attack

E2EE + TLS:
  Client ──encrypt──→ TLS ──→ App Server ──ciphertext──→ HSM/Vault ──decrypt
                                    ↑                         ↑
                              NO plaintext              only here is it
                              in server memory          decrypted
```

### E2EE Architecture: Client-to-HSM

```
┌──────────────────────────────────────────────────────────────────────┐
│                        CLIENT (Mobile/Web)                          │
│                                                                      │
│  1. Get server public key:                                           │
│     GET /api/v1/crypto/public-key → ECDH public key (P-256)         │
│                                                                      │
│  2. Encrypt the sensitive field (ECIES):                             │
│     ephemeral_key = ECDH_GenerateKeypair()                          │
│     shared_secret = ECDH(ephemeral_private, server_public)          │
│     derived_key   = HKDF-SHA256(shared_secret, salt, info)          │
│     ciphertext    = AES-256-GCM(plaintext, derived_key, nonce)      │
│                                                                      │
│  3. Send request:                                                    │
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
│  Server does NOT OPEN encrypted_pan — keeps it as ciphertext         │
│  Proxies to Vault/HSM:                                               │
│                                                                      │
│  result = vault.Decrypt("transit/xbank/e2ee", ciphertext, nonce,    │
│                          ephemeral_public_key)                       │
│                                                                      │
│  or                                                                  │
│                                                                      │
│  Server stores ciphertext directly in DB                             │
│  (card PAN, KYC document — no processing needed)                     │
└──────────────────────────────────────────────────────────────────────┘
         │
         ▼
┌──────────────────────────────────────────────────────────────────────┐
│                        HSM / VAULT                                   │
│                                                                      │
│  1. Ephemeral public key + server private key → shared_secret        │
│  2. HKDF-SHA256(shared_secret) → derived_key                        │
│  3. AES-256-GCM_Decrypt(ciphertext, derived_key, nonce)             │
│  4. Plaintext → process → result                                    │
│  5. Clear plaintext from memory IMMEDIATELY                          │
│                                                                      │
│  Private key NEVER leaves Vault!                                     │
└──────────────────────────────────────────────────────────────────────┘
```

### ECIES (Elliptic Curve Integrated Encryption Scheme)

<!-- ECIES = ECDH key agreement + HKDF key derivation + AES-GCM encryption
     New ephemeral key for each message → forward secrecy -->

```
ECIES encryption process:

  1. Generate ephemeral keypair:
     ephemeral_private, ephemeral_public = ECDH_Generate(P-256)

  2. Compute shared secret:
     shared_secret = ECDH(ephemeral_private, server_public_key)

  3. Key derivation:
     encryption_key = HKDF-SHA256(
       ikm:  shared_secret,
       salt: random_32_bytes,
       info: "xbank-e2ee-v1",
       len:  32  // AES-256 key
     )

  4. Encryption:
     nonce = random_12_bytes  // AES-GCM nonce
     ciphertext, tag = AES-256-GCM_Encrypt(plaintext, encryption_key, nonce)

  5. Result:
     {
       ciphertext:          ciphertext || tag,   // encrypted data + auth tag
       ephemeral_public_key: ephemeral_public,   // for ECDH
       nonce:               nonce,               // AES-GCM nonce
       salt:                salt,                // HKDF salt
       key_id:              "e2ee_2026_q2"       // which server key was used
     }
```

### Why ECIES?

| Property | RSA-OAEP | ECIES |
|---|---|---|
| Key size | 4096-bit | 256-bit (P-256) |
| Max plaintext | ~446 byte (RSA-4096) | **Unlimited** (hybrid) |
| Forward secrecy | No | Yes (ephemeral key) |
| Performance | Slow | **Fast** |
| Mobile battery | High consumption | **Low** |

### Which Fields Require E2EE

| Field | E2EE | Reason |
|---|---|---|
| Card PAN (16 digits) | **Required** | PCI DSS — plaintext must not be in server memory |
| Card PIN | **Required** | ISO 9564 — only HSM decrypts |
| Card CVV | **Required** | PCI DSS — never stored, only for verification |
| KYC document (passport, selfie) | **Required** | Personal data — GDPR, local legislation |
| KYC document number | **Required** | PII (Personally Identifiable Information) |
| Transfer amount | No | Server needs to check balance — plaintext required |
| Account ID | No | Server needs to route |
| OTP / 2FA code | No | Server verifies, TLS is sufficient |
| Login credentials | SRP or TLS | Password should not reach server as plaintext (bcrypt not on client) |

### PIN Block (ISO 9564)

<!-- PIN is decrypted only in HSM — server NEVER sees plaintext PIN -->

```
PIN encryption (ATM/POS/Mobile):

  Format 0 (ISO 9564-1):
    1. PIN block = PIN length || PIN || padding (F)
       Example: PIN = "1234"
       PIN block = 0x04 1234 FFFFFFFFFF

    2. PAN block = 0000 || PAN[3..14]  (last 12 digits, without check digit)
       PAN = 4000001234567890
       PAN block = 0x0000 000123456789

    3. Clear PIN block = PIN block XOR PAN block

    4. Encrypted PIN block = AES-256-GCM(clear_pin_block, pin_encryption_key)
       or
       3DES_Encrypt(clear_pin_block, ZPK)  // for legacy terminals

For Mobile/Web:
    1. Client → ECIES(PIN, server_e2ee_public_key) → encrypted_pin
    2. Server → sends to Vault (never sees plaintext)
    3. Vault → decrypt → PIN hash (bcrypt) → store
    4. Subsequent verify: bcrypt.Compare() in Vault
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
1. Generate new ECDH P-256 keypair (inside Vault)
   vault write transit/xbank/e2ee/keys type=ecdsa-p256

2. Publish public key via API
   GET /api/v1/crypto/public-key
   → { key_id: "e2ee_2026_q2", public_key: PEM, algorithm: "ECIES-P256" }

3. Client starts encrypting with new key_id

4. Old key → ROTATE_OUT (90 days — decryption continues for old clients)

5. After 90 days → RETIRED (only for archived data)
```

### Go Implementation

```go
import (
    "crypto/ecdh"
    "crypto/rand"
    "crypto/sha256"
    "golang.org/x/crypto/hkdf"
)

// Client side: ECIES encryption
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

// EncryptedPayload — E2EE encrypted data
type EncryptedPayload struct {
    Ciphertext         []byte `json:"ciphertext"`
    EphemeralPublicKey []byte `json:"ephemeral_public_key"`
    Nonce              []byte `json:"nonce"`
    Salt               []byte `json:"salt"`
    KeyID              string `json:"key_id"`
}
```

```go
// Server side: Proxy to Vault (server never sees plaintext)
func (s *CryptoService) DecryptE2EE(ctx context.Context, payload *EncryptedPayload) ([]byte, error) {
    // Server does NOT decrypt itself — sends to Vault
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

### E2EE Request/Response Format

```
Adding a card (with E2EE):

POST /api/v1/cards
{
  "account_id": "acc-uuid",                    ← plaintext (server needs it)
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

### Difference Between E2EE and Existing Encryption

```
Current (server-side encryption):
  Client ──plaintext──→ TLS ──→ Server ──AES-GCM encrypt──→ DB
                                  ↑
                            plaintext exists in
                            server memory!

E2EE (client-side encryption):
  Client ──ECIES encrypt──→ TLS ──→ Server ──ciphertext──→ Vault decrypt
                                      ↑                        ↑
                                NO plaintext            only in Vault
                                in server memory        plaintext exists

Result: Two layers of protection
  1. TLS 1.3     → transport (against MITM)
  2. ECIES (E2EE) → application (against server compromise)
```

## API Endpoints

| Method | Endpoint | Middleware | Description |
|---|---|---|---|
| GET | `/api/v1/crypto/public-key` | Public | E2EE server public key (ECIES P-256) |
| POST | `/api/v1/auth/signing-keys` | Session | Register client public key |
| DELETE | `/api/v1/auth/signing-keys/{id}` | Session+2FA | Revoke signing key |
| GET | `/api/v1/auth/signing-keys` | Session | List of active signing keys |
| GET | `/.well-known/jwks.json` | Public | JWT public keys (JWKS) |
