# Cards — PCI DSS Compliant

## Card Aggregate
```go
type Card struct {
    AggregateRoot
    AccountID       uuid.UUID
    UserID          uuid.UUID
    EncryptedPAN    []byte               // AES-256-GCM (encrypted with DEK, encrypted in Vault)
    EncryptedDEK    []byte               // RSA-4096(DEK, Card KEK public)
    EncryptionNonce []byte               // AES GCM nonce (12 byte)
    EncryptionKeyID string               // which KEK was used
    MaskedNumber    string               // **** **** **** 1234
    PANHash         string               // SHA-256 (for lookup)
    CardholderName  string
    ExpiryMonth     int
    ExpiryYear      int
    CVVHash         string               // bcrypt (NEVER plain)
    PINHash         string               // bcrypt
    CardType        CardType             // DEBIT, VIRTUAL
    Status          CardStatus           // INACTIVE, ACTIVE, BLOCKED, EXPIRED, CANCELLED
    DailyLimit      Money
    MonthlyLimit    Money
}
```

## Client → Server: Sending Card Data (E2EE)

<!-- PAN, PIN, CVV — NEVER in plaintext on the network or in server memory.
     Client encrypts with ECIES → Server proxies ciphertext to Vault. -->

### 1. Obtaining E2EE Public Key

```
When the client app opens or the first time entering the card flow:

GET /api/v1/crypto/public-key
Authorization: Bearer <JWT>

Response:
{
  "key_id": "e2ee_2026_q2",
  "public_key": "-----BEGIN PUBLIC KEY-----\nMFkwEwYHKoZIzj0C...",
  "algorithm": "ECIES-P256"
}

Client caches this key (along with key_id).
When key_id changes, it fetches the new key.
```

### 2. Card Creation (Issue) — Client Flow

```
┌──────────────────────────────────────────────────────────────────────┐
│                      CLIENT (Mobile/Web)                             │
│                                                                      │
│  User enters:                                                        │
│    PAN:  4000 0012 3456 7890                                        │
│    PIN:  1234                                                        │
│    CVV:  567                                                         │
│                                                                      │
│  ┌─ Luhn validation (client-side) ─────────────────────────────┐    │
│  │  sum = luhn_checksum("4000001234567890")                    │    │
│  │  sum % 10 == 0  → ✅ Valid                                  │    │
│  │  (server also re-validates, but for fast feedback)          │    │
│  └─────────────────────────────────────────────────────────────┘    │
│                                                                      │
│  ┌─ Separate ECIES for each sensitive field ───────────────────┐    │
│  │                                                              │    │
│  │  // Encrypt PAN                                              │    │
│  │  eph_key_1 = ECDH_Generate(P-256)      ← NEW key            │    │
│  │  shared_1  = ECDH(eph_priv_1, server_pub)                   │    │
│  │  aes_key_1 = HKDF(shared_1, salt_1, "xbank-e2ee-v1")       │    │
│  │  enc_pan   = AES-GCM(pan_bytes, aes_key_1, nonce_1)         │    │
│  │                                                              │    │
│  │  // Encrypt PIN                                              │    │
│  │  eph_key_2 = ECDH_Generate(P-256)      ← NEW key (different)│    │
│  │  shared_2  = ECDH(eph_priv_2, server_pub)                   │    │
│  │  aes_key_2 = HKDF(shared_2, salt_2, "xbank-e2ee-v1")       │    │
│  │  enc_pin   = AES-GCM(pin_bytes, aes_key_2, nonce_2)         │    │
│  │                                                              │    │
│  │  // Encrypt CVV                                              │    │
│  │  eph_key_3 = ECDH_Generate(P-256)      ← NEW key (different)│    │
│  │  shared_3  = ECDH(eph_priv_3, server_pub)                   │    │
│  │  aes_key_3 = HKDF(shared_3, salt_3, "xbank-e2ee-v1")       │    │
│  │  enc_cvv   = AES-GCM(cvv_bytes, aes_key_3, nonce_3)         │    │
│  │                                                              │    │
│  │  ⚠️ Each field has a SEPARATE ephemeral key — if one is     │    │
│  │     compromised, the others remain safe (forward secrecy     │    │
│  │     per-field)                                               │    │
│  └──────────────────────────────────────────────────────────────┘    │
│                                                                      │
│  ┌─ DELETE ephemeral private keys ─────────────────────────────┐    │
│  │  eph_priv_1 = nil   // clear from memory                   │    │
│  │  eph_priv_2 = nil                                           │    │
│  │  eph_priv_3 = nil                                           │    │
│  │  runtime.GC()       // garbage collect                      │    │
│  └─────────────────────────────────────────────────────────────┘    │
│                                                                      │
│  Plaintext PAN, PIN, CVV are also DELETED from memory.              │
└──────────────────────────────────────────────────────────────────────┘
```

### 3. HTTP Request

```
POST /api/v1/cards
Headers:
  Authorization: Bearer <JWT>
  Content-Type: application/json
  X-Device-Fingerprint: "sha256:abc123..."
  X-Idempotency-Key: "idemp-uuid"

Body:
{
  "account_id": "acc-uuid",
  "cardholder_name": "BAKHODIR YASHINI",
  "expiry_month": 12,
  "expiry_year": 2029,
  "card_type": "DEBIT",

  "encrypted_pan": {
    "ciphertext": "base64(AES-GCM encrypted PAN)...",
    "ephemeral_public_key": "base64(eph_pub_1)...",
    "nonce": "base64(nonce_1)...",
    "salt": "base64(salt_1)...",
    "key_id": "e2ee_2026_q2"
  },

  "encrypted_pin": {
    "ciphertext": "base64(AES-GCM encrypted PIN)...",
    "ephemeral_public_key": "base64(eph_pub_2)...",
    "nonce": "base64(nonce_2)...",
    "salt": "base64(salt_2)...",
    "key_id": "e2ee_2026_q2"
  },

  "encrypted_cvv": {
    "ciphertext": "base64(AES-GCM encrypted CVV)...",
    "ephemeral_public_key": "base64(eph_pub_3)...",
    "nonce": "base64(nonce_3)...",
    "salt": "base64(salt_3)...",
    "key_id": "e2ee_2026_q2"
  }
}
```

### 4. Server Processing Flow

```
┌──────────────────────────────────────────────────────────────────────┐
│                      APPLICATION SERVER                              │
│                                                                      │
│  ⚠️ Server NEVER sees PAN, PIN, CVV in plaintext                    │
│                                                                      │
│  1. Request validation (check ciphertext presence)                   │
│     ├── encrypted_pan present?    → ✅                               │
│     ├── encrypted_pin present?    → ✅                               │
│     ├── encrypted_cvv present?    → ✅                               │
│     ├── key_id valid?             → ✅ (ACTIVE or ROTATE_OUT)        │
│     └── cardholder_name, expiry valid? → ✅                          │
│                                                                      │
│  2. Idempotency check                                                │
│     └── X-Idempotency-Key exists in DB? → yes → cached response     │
│                                                                      │
│  3. Account and User verification                                    │
│     ├── Does account_id belong to the user?                          │
│     ├── KYC status = VERIFIED?                                       │
│     └── Card limit not exceeded?                                     │
│                                                                      │
│  4. Send to Vault (for PAN)                                          │
│     ┌──────────────────────────────────────────────────────┐         │
│     │  vault.DecryptAndReEncrypt("transit/xbank/e2ee", {   │         │
│     │    ciphertext:          encrypted_pan.ciphertext,    │         │
│     │    ephemeral_public_key: encrypted_pan.eph_pub,      │         │
│     │    nonce:               encrypted_pan.nonce,         │         │
│     │    salt:                encrypted_pan.salt,          │         │
│     │    key_id:              encrypted_pan.key_id,        │         │
│     │    re_encrypt_with:     "card_kek_v3"               │         │
│     │  })                                                  │         │
│     │                                                      │         │
│     │  Inside Vault:                                       │         │
│     │    a. ECIES decrypt → plaintext PAN                  │         │
│     │    b. Luhn validation (server-side)                   │         │
│     │    c. Generate random DEK                            │         │
│     │    d. AES-256-GCM(PAN, DEK) → encrypted_pan_storage  │         │
│     │    e. RSA(DEK, card_KEK) → encrypted_dek             │         │
│     │    f. SHA-256(PAN) → pan_hash                        │         │
│     │    g. PAN[12:16] → last_four ("7890")                │         │
│     │    h. DELETE PAN plaintext                           │         │
│     │                                                      │         │
│     │  Return:                                             │         │
│     │    encrypted_pan_storage, encrypted_dek, nonce,      │         │
│     │    pan_hash, last_four, key_id                       │         │
│     └──────────────────────────────────────────────────────┘         │
│                                                                      │
│  5. Send to Vault (for PIN)                                          │
│     ┌──────────────────────────────────────────────────────┐         │
│     │  vault.DecryptAndHash("transit/xbank/e2ee", {        │         │
│     │    ciphertext: encrypted_pin.ciphertext,             │         │
│     │    ...,                                              │         │
│     │    hash_algorithm: "bcrypt",                         │         │
│     │    bcrypt_cost: 12                                   │         │
│     │  })                                                  │         │
│     │                                                      │         │
│     │  Inside Vault:                                       │         │
│     │    a. ECIES decrypt → plaintext PIN                  │         │
│     │    b. PIN format check (4-6 digits)                  │         │
│     │    c. bcrypt.Hash(PIN, cost=12) → pin_hash           │         │
│     │    d. DELETE PIN plaintext                           │         │
│     │                                                      │         │
│     │  Return: pin_hash                                    │         │
│     └──────────────────────────────────────────────────────┘         │
│                                                                      │
│  6. Send to Vault (for CVV)                                          │
│     ┌──────────────────────────────────────────────────────┐         │
│     │  vault.DecryptAndHash("transit/xbank/e2ee", {        │         │
│     │    ciphertext: encrypted_cvv.ciphertext,             │         │
│     │    ...,                                              │         │
│     │    hash_algorithm: "bcrypt",                         │         │
│     │    bcrypt_cost: 12                                   │         │
│     │  })                                                  │         │
│     │                                                      │         │
│     │  Inside Vault:                                       │         │
│     │    a. ECIES decrypt → plaintext CVV                  │         │
│     │    b. CVV format check (3-4 digits)                  │         │
│     │    c. bcrypt.Hash(CVV, cost=12) → cvv_hash           │         │
│     │    d. DELETE CVV plaintext                           │         │
│     │    ⚠️ CVV is NEVER STORED — only the hash            │         │
│     │                                                      │         │
│     │  Return: cvv_hash                                    │         │
│     └──────────────────────────────────────────────────────┘         │
│                                                                      │
│  7. Save to DB                                                       │
│     INSERT INTO cards (                                              │
│       account_id, user_id,                                          │
│       encrypted_pan,     ← from Vault (re-encrypted)                 │
│       encrypted_dek,     ← from Vault                                │
│       encryption_nonce,  ← from Vault                                │
│       encryption_key_id, ← card_kek_v3                               │
│       masked_number,     ← "**** **** **** 7890"                     │
│       pan_hash,          ← from Vault (SHA-256)                      │
│       cvv_hash,          ← from Vault (bcrypt)                       │
│       pin_hash,          ← from Vault (bcrypt)                       │
│       cardholder_name, expiry_month, expiry_year,                    │
│       card_type, status = 'INACTIVE'                                 │
│     )                                                                │
│                                                                      │
│  8. Response                                                         │
└──────────────────────────────────────────────────────────────────────┘
```

### 5. Response

```json
{
  "status": "success",
  "data": {
    "card_id": "card-uuid",
    "masked_number": "**** **** **** 7890",
    "cardholder_name": "BAKHODIR YASHINI",
    "expiry_month": 12,
    "expiry_year": 2029,
    "card_type": "DEBIT",
    "status": "INACTIVE",
    "daily_limit": { "amount": 5000000, "currency": "UZS" },
    "monthly_limit": { "amount": 50000000, "currency": "UZS" }
  }
}
```

## PIN Verification Flow (E2EE)

<!-- When PIN is entered from ATM, POS, or Mobile -->

```
┌─ Client ─────────────────────────────────────────────────────────────┐
│                                                                       │
│  User enters PIN: 1234                                               │
│                                                                       │
│  eph_key = ECDH_Generate(P-256)                                      │
│  encrypted_pin = ECIES_Encrypt("1234", server_public_key, eph_key)   │
│                                                                       │
│  POST /api/v1/cards/{id}/verify-pin                                  │
│  {                                                                    │
│    "encrypted_pin": {                                                │
│      "ciphertext": "base64...",                                      │
│      "ephemeral_public_key": "base64...",                            │
│      "nonce": "base64...",                                           │
│      "salt": "base64...",                                            │
│      "key_id": "e2ee_2026_q2"                                       │
│    }                                                                  │
│  }                                                                    │
└───────────────────────────────────────────────────────────────────────┘
         │
         ▼
┌─ Server ─────────────────────────────────────────────────────────────┐
│                                                                       │
│  1. Retrieve pin_hash from DB                                        │
│  2. Send to Vault:                                                   │
│     vault.DecryptAndCompare({                                        │
│       ciphertext: ...,                                               │
│       pin_hash: "$2a$12$..."   ← from DB                            │
│     })                                                               │
│                                                                       │
│  Inside Vault:                                                       │
│    a. ECIES decrypt → plaintext PIN                                  │
│    b. bcrypt.Compare(pin_hash, PIN)                                  │
│    c. DELETE PIN plaintext                                           │
│    d. Return: match = true/false                                     │
│                                                                       │
│  3. match == false → wrong_pin_count++                               │
│     wrong_pin_count >= 3 → Card BLOCKED + alert                     │
│                                                                       │
│  4. match == true → operation continues                              │
└───────────────────────────────────────────────────────────────────────┘
```

## CVV Verification Flow (Online Payment)

```
┌─ Client (E-commerce checkout) ───────────────────────────────────────┐
│                                                                       │
│  User enters: PAN + Expiry + CVV                                     │
│                                                                       │
│  Encrypts each with SEPARATE ECIES                                   │
│  (PAN, CVV — each with its own ephemeral key)                        │
│                                                                       │
│  POST /api/v1/cards/verify-payment                                   │
│  {                                                                    │
│    "encrypted_pan": { ... },         ← E2EE                         │
│    "encrypted_cvv": { ... },         ← E2EE                         │
│    "expiry_month": 12,               ← plaintext (safe)             │
│    "expiry_year": 2029,              ← plaintext (safe)             │
│    "amount": 150000,                 ← plaintext (server needs it)   │
│    "currency": "UZS",                                                │
│    "merchant_id": "merch-uuid"                                       │
│  }                                                                    │
└───────────────────────────────────────────────────────────────────────┘
         │
         ▼
┌─ Server ─────────────────────────────────────────────────────────────┐
│                                                                       │
│  1. Vault: ECIES decrypt PAN → PAN plaintext                        │
│  2. Vault: SHA-256(PAN) → pan_hash_computed                         │
│  3. DB: find card by pan_hash                                        │
│  4. Vault: ECIES decrypt CVV → CVV plaintext                        │
│  5. Vault: bcrypt.Compare(card.cvv_hash, CVV) → match?             │
│  6. Expiry check (server-side, plaintext)                            │
│  7. DELETE all plaintext (inside Vault)                              │
│                                                                       │
│  match == true  → Payment continues (hold → capture)                │
│  match == false → 400 "Invalid card details"                        │
└───────────────────────────────────────────────────────────────────────┘
```

## Card Data Lifecycle

```
┌─────────────────────────────────────────────────────────────────┐
│  Data          │ On Client  │ On Network  │ On Server   │ In DB │
│────────────────│────────────│─────────────│─────────────│───────│
│  PAN           │ Plaintext  │ E2EE        │ Ciphertext  │ AES   │
│  (16 digits)   │ → ECIES    │ (ciphertext)│ → Vault     │ (DEK) │
│                │ → DELETE   │             │ → DELETE    │       │
│                │            │             │             │       │
│  PIN           │ Plaintext  │ E2EE        │ Ciphertext  │ bcrypt│
│  (4-6 digits)  │ → ECIES    │ (ciphertext)│ → Vault     │ hash  │
│                │ → DELETE   │             │ → DELETE    │       │
│                │            │             │             │       │
│  CVV           │ Plaintext  │ E2EE        │ Ciphertext  │ bcrypt│
│  (3-4 digits)  │ → ECIES    │ (ciphertext)│ → Vault     │ hash  │
│                │ → DELETE   │             │ → DELETE    │       │
│                │            │             │             │       │
│  Last 4        │ ❌          │ ❌          │ Vault →     │ Plain │
│                │            │             │ computes    │ "7890"│
│                │            │             │             │       │
│  PAN Hash      │ ❌          │ ❌          │ Vault →     │ SHA   │
│                │            │             │ computes    │ -256  │
└─────────────────────────────────────────────────────────────────┘

Places where plaintext exists:
  1. During client input      → deleted immediately after ECIES
  2. Inside Vault/HSM         → deleted immediately after operation

  ⚠️ Plaintext exists NOWHERE else:
     ❌ On the network (E2EE)
     ❌ In server memory (ciphertext → Vault proxy)
     ❌ In DB (AES encrypted or bcrypt hash)
     ❌ In logs (sensitive fields are masked)
```

## E2EE Encryption (ECIES + AES-256-GCM)

A **unique DEK** (Data Encryption Key) is generated for each card.
The DEK itself is stored in the DB encrypted with the **Card KEK** (Key Encryption Key).
The Card KEK private key is only in **Vault/HSM**.

```
Client → Server (E2EE):
  1. Client: ECIES_Encrypt(PAN, server_e2ee_public_key) → ciphertext
  2. Server: proxies ciphertext to Vault (does not see plaintext)

Inside Vault (re-encrypt for storage):
  3. ECIES decrypt → plaintext PAN
  4. Generate random DEK (AES-256 key)
  5. AES-256-GCM(PAN, DEK, nonce) → encrypted_pan
  6. RSA_Encrypt(DEK, card_KEK_public) → encrypted_dek
  7. DB: encrypted_pan + encrypted_dek + nonce + key_id

Decryption (reading a card):
  1. Vault: encrypted_dek → RSA_Decrypt(card_KEK_private) → DEK
  2. Vault: AES_Decrypt(encrypted_pan, DEK, nonce) → PAN
  3. Plaintext only inside Vault

KEK ROTATE:
  Only DEK re-wrap (card data is not re-encrypted)
```

Details: [Encryption & PKI](../security/encryption.md#end-to-end-encryption-e2ee)

## PCI DSS Compliance
- **Req 3**: Card data **E2EE (ECIES)** + storage **AES-256-GCM** (with DEK), unique DEK per-card
- **Req 3.2**: CVV stored only as bcrypt hash, plaintext NEVER
- **Req 3.4**: PAN is never stored in plain text, not even in server memory (E2EE)
- **Req 3.5**: KEK private key only in HSM/Vault
- **Req 4**: TLS 1.3 transport + E2EE application-level
- **Req 7**: User can only access their own cards (RLS)
- **Req 8**: Strong auth + 2FA
- **Req 10**: Card data access in audit log

## Tokenization
A random token instead of the real card number:
```go
type CardToken struct {
    Token     string    // "tok_xxxxxxxxxxxxxxxx"
    LastFour  string    // "1234"
    IsActive  bool
    ExpiresAt time.Time
}
```

## EMV Standard
- Luhn algorithm — card number validation
- Card network detection (Visa, MasterCard, UnionPay)
- 3D Secure result struct (for online payment)

## Operations
- **Issue** → E2EE → Vault decrypt → AES encrypt PAN (DEK + KEK), bcrypt hash CVV/PIN → INACTIVE
- **Activate** → ACTIVE
- **Block** → BLOCKED (3 wrong PIN → auto block)
- **PlaceHold** → reduce available_balance
- **CaptureHold** → partial or full capture
- **ReleaseHold** → cancel hold

## API Endpoints

| Method | Endpoint | Middleware | Description |
|---|---|---|---|
| POST | `/api/v1/cards` | Session+KYC+2FA | Create card (E2EE: PAN+PIN+CVV) |
| GET | `/api/v1/cards` | Session | List cards (masked) |
| GET | `/api/v1/cards/{id}` | Session | Card details (masked) |
| POST | `/api/v1/cards/{id}/activate` | Session | Activate card |
| POST | `/api/v1/cards/{id}/block` | Session | Block card |
| POST | `/api/v1/cards/{id}/verify-pin` | Session | Verify PIN (E2EE) |
| PUT | `/api/v1/cards/{id}/pin` | Session+2FA | Change PIN (E2EE) |
| PUT | `/api/v1/cards/{id}/limits` | Session | Change limits |
| POST | `/api/v1/cards/{id}/tokenize` | Session | Tokenization |
| POST | `/api/v1/cards/verify-payment` | Session | Online payment (E2EE: PAN+CVV) |
