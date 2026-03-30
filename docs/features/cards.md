# Cards — PCI DSS Compliant

## Card Aggregate
```go
type Card struct {
    AggregateRoot
    AccountID       uuid.UUID
    UserID          uuid.UUID
    EncryptedPAN    []byte               // AES-256-GCM (DEK bilan, Vault da shifrlangan)
    EncryptedDEK    []byte               // RSA-4096(DEK, Card KEK public)
    EncryptionNonce []byte               // AES GCM nonce (12 byte)
    EncryptionKeyID string               // qaysi KEK ishlatilgan
    MaskedNumber    string               // **** **** **** 1234
    PANHash         string               // SHA-256 (qidiruv uchun)
    CardholderName  string
    ExpiryMonth     int
    ExpiryYear      int
    CVVHash         string               // bcrypt (HECH QACHON plain)
    PINHash         string               // bcrypt
    CardType        CardType             // DEBIT, VIRTUAL
    Status          CardStatus           // INACTIVE, ACTIVE, BLOCKED, EXPIRED, CANCELLED
    DailyLimit      Money
    MonthlyLimit    Money
}
```

## Client → Server: Card Data Yuborish (E2EE)

<!-- PAN, PIN, CVV — HECH QACHON plaintext tarmoqda yoki server memory da bo'lmaydi.
     Client ECIES bilan shifrlaydi → Server ciphertext ni Vault ga proxy qiladi. -->

### 1. E2EE Public Key Olish

```
Client app ochilganda yoki birinchi marta card flow ga kirganda:

GET /api/v1/crypto/public-key
Authorization: Bearer <JWT>

Response:
{
  "key_id": "e2ee_2026_q2",
  "public_key": "-----BEGIN PUBLIC KEY-----\nMFkwEwYHKoZIzj0C...",
  "algorithm": "ECIES-P256"
}

Client bu key ni cache qiladi (key_id bilan birga).
key_id o'zgarganda yangi key oladi.
```

### 2. Karta Yaratish (Issue) — Client Flow

```
┌──────────────────────────────────────────────────────────────────────┐
│                      CLIENT (Mobile/Web)                             │
│                                                                      │
│  User kiritadi:                                                      │
│    PAN:  4000 0012 3456 7890                                        │
│    PIN:  1234                                                        │
│    CVV:  567                                                         │
│                                                                      │
│  ┌─ Luhn validation (client-side) ─────────────────────────────┐    │
│  │  sum = luhn_checksum("4000001234567890")                    │    │
│  │  sum % 10 == 0  → ✅ Valid                                  │    │
│  │  (server ham qayta tekshiradi, lekin tez feedback uchun)    │    │
│  └─────────────────────────────────────────────────────────────┘    │
│                                                                      │
│  ┌─ Har bir sensitive field uchun alohida ECIES ───────────────┐    │
│  │                                                              │    │
│  │  // PAN shifrlash                                            │    │
│  │  eph_key_1 = ECDH_Generate(P-256)      ← YANGI key          │    │
│  │  shared_1  = ECDH(eph_priv_1, server_pub)                   │    │
│  │  aes_key_1 = HKDF(shared_1, salt_1, "xbank-e2ee-v1")       │    │
│  │  enc_pan   = AES-GCM(pan_bytes, aes_key_1, nonce_1)         │    │
│  │                                                              │    │
│  │  // PIN shifrlash                                            │    │
│  │  eph_key_2 = ECDH_Generate(P-256)      ← YANGI key (boshqa)│    │
│  │  shared_2  = ECDH(eph_priv_2, server_pub)                   │    │
│  │  aes_key_2 = HKDF(shared_2, salt_2, "xbank-e2ee-v1")       │    │
│  │  enc_pin   = AES-GCM(pin_bytes, aes_key_2, nonce_2)         │    │
│  │                                                              │    │
│  │  // CVV shifrlash                                            │    │
│  │  eph_key_3 = ECDH_Generate(P-256)      ← YANGI key (boshqa)│    │
│  │  shared_3  = ECDH(eph_priv_3, server_pub)                   │    │
│  │  aes_key_3 = HKDF(shared_3, salt_3, "xbank-e2ee-v1")       │    │
│  │  enc_cvv   = AES-GCM(cvv_bytes, aes_key_3, nonce_3)         │    │
│  │                                                              │    │
│  │  ⚠️ Har bir field ALOHIDA ephemeral key — bitta buzilsa     │    │
│  │     boshqalari xavfsiz (forward secrecy per-field)          │    │
│  └──────────────────────────────────────────────────────────────┘    │
│                                                                      │
│  ┌─ Ephemeral private key larni O'CHIRISH ─────────────────────┐    │
│  │  eph_priv_1 = nil   // memory dan tozalash                  │    │
│  │  eph_priv_2 = nil                                           │    │
│  │  eph_priv_3 = nil                                           │    │
│  │  runtime.GC()       // garbage collect                      │    │
│  └─────────────────────────────────────────────────────────────┘    │
│                                                                      │
│  Plaintext PAN, PIN, CVV ham memory dan O'CHIRILADI.                │
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

### 4. Server Ishlash Flow

```
┌──────────────────────────────────────────────────────────────────────┐
│                      APPLICATION SERVER                              │
│                                                                      │
│  ⚠️ Server PAN, PIN, CVV ni HECH QACHON plaintext ko'rmaydi         │
│                                                                      │
│  1. Request validation (ciphertext mavjudligini tekshirish)          │
│     ├── encrypted_pan bor?    → ✅                                   │
│     ├── encrypted_pin bor?    → ✅                                   │
│     ├── encrypted_cvv bor?    → ✅                                   │
│     ├── key_id valid?         → ✅ (ACTIVE yoki ROTATE_OUT)          │
│     └── cardholder_name, expiry valid? → ✅                          │
│                                                                      │
│  2. Idempotency check                                                │
│     └── X-Idempotency-Key DB da bormi? → bor → cached response      │
│                                                                      │
│  3. Account va User tekshiruv                                        │
│     ├── account_id user ga tegishlimi?                               │
│     ├── KYC status = VERIFIED?                                       │
│     └── Karta limiti oshmaganmi?                                     │
│                                                                      │
│  4. Vault ga yuborish (PAN uchun)                                    │
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
│     │  Vault ichida:                                       │         │
│     │    a. ECIES decrypt → plaintext PAN                  │         │
│     │    b. Luhn validation (server-side)                   │         │
│     │    c. Random DEK generatsiya                         │         │
│     │    d. AES-256-GCM(PAN, DEK) → encrypted_pan_storage  │         │
│     │    e. RSA(DEK, card_KEK) → encrypted_dek             │         │
│     │    f. SHA-256(PAN) → pan_hash                        │         │
│     │    g. PAN[12:16] → last_four ("7890")                │         │
│     │    h. PAN plaintext O'CHIRISH                        │         │
│     │                                                      │         │
│     │  Return:                                             │         │
│     │    encrypted_pan_storage, encrypted_dek, nonce,      │         │
│     │    pan_hash, last_four, key_id                       │         │
│     └──────────────────────────────────────────────────────┘         │
│                                                                      │
│  5. Vault ga yuborish (PIN uchun)                                    │
│     ┌──────────────────────────────────────────────────────┐         │
│     │  vault.DecryptAndHash("transit/xbank/e2ee", {        │         │
│     │    ciphertext: encrypted_pin.ciphertext,             │         │
│     │    ...,                                              │         │
│     │    hash_algorithm: "bcrypt",                         │         │
│     │    bcrypt_cost: 12                                   │         │
│     │  })                                                  │         │
│     │                                                      │         │
│     │  Vault ichida:                                       │         │
│     │    a. ECIES decrypt → plaintext PIN                  │         │
│     │    b. PIN format tekshirish (4-6 raqam)              │         │
│     │    c. bcrypt.Hash(PIN, cost=12) → pin_hash           │         │
│     │    d. PIN plaintext O'CHIRISH                        │         │
│     │                                                      │         │
│     │  Return: pin_hash                                    │         │
│     └──────────────────────────────────────────────────────┘         │
│                                                                      │
│  6. Vault ga yuborish (CVV uchun)                                    │
│     ┌──────────────────────────────────────────────────────┐         │
│     │  vault.DecryptAndHash("transit/xbank/e2ee", {        │         │
│     │    ciphertext: encrypted_cvv.ciphertext,             │         │
│     │    ...,                                              │         │
│     │    hash_algorithm: "bcrypt",                         │         │
│     │    bcrypt_cost: 12                                   │         │
│     │  })                                                  │         │
│     │                                                      │         │
│     │  Vault ichida:                                       │         │
│     │    a. ECIES decrypt → plaintext CVV                  │         │
│     │    b. CVV format tekshirish (3-4 raqam)              │         │
│     │    c. bcrypt.Hash(CVV, cost=12) → cvv_hash           │         │
│     │    d. CVV plaintext O'CHIRISH                        │         │
│     │    ⚠️ CVV HECH QACHON SAQLANMAYDI — faqat hash       │         │
│     │                                                      │         │
│     │  Return: cvv_hash                                    │         │
│     └──────────────────────────────────────────────────────┘         │
│                                                                      │
│  7. DB ga saqlash                                                    │
│     INSERT INTO cards (                                              │
│       account_id, user_id,                                          │
│       encrypted_pan,     ← Vault dan (re-encrypted)                  │
│       encrypted_dek,     ← Vault dan                                 │
│       encryption_nonce,  ← Vault dan                                 │
│       encryption_key_id, ← card_kek_v3                               │
│       masked_number,     ← "**** **** **** 7890"                     │
│       pan_hash,          ← Vault dan (SHA-256)                       │
│       cvv_hash,          ← Vault dan (bcrypt)                        │
│       pin_hash,          ← Vault dan (bcrypt)                        │
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

## PIN Tekshirish Flow (E2EE)

<!-- ATM, POS, yoki Mobile dan PIN kiritilganda -->

```
┌─ Client ─────────────────────────────────────────────────────────────┐
│                                                                       │
│  User PIN kiritadi: 1234                                             │
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
│  1. DB dan pin_hash olish                                            │
│  2. Vault ga yuborish:                                               │
│     vault.DecryptAndCompare({                                        │
│       ciphertext: ...,                                               │
│       pin_hash: "$2a$12$..."   ← DB dan                             │
│     })                                                               │
│                                                                       │
│  Vault ichida:                                                       │
│    a. ECIES decrypt → plaintext PIN                                  │
│    b. bcrypt.Compare(pin_hash, PIN)                                  │
│    c. PIN plaintext O'CHIRISH                                        │
│    d. Return: match = true/false                                     │
│                                                                       │
│  3. match == false → wrong_pin_count++                               │
│     wrong_pin_count >= 3 → Card BLOCKED + alert                     │
│                                                                       │
│  4. match == true → operatsiya davom etadi                           │
└───────────────────────────────────────────────────────────────────────┘
```

## CVV Tekshirish Flow (Online Payment)

```
┌─ Client (E-commerce checkout) ───────────────────────────────────────┐
│                                                                       │
│  User kiritadi: PAN + Expiry + CVV                                   │
│                                                                       │
│  Har birini ALOHIDA ECIES bilan shifrlaydi                           │
│  (PAN, CVV — har biri o'z ephemeral key bilan)                       │
│                                                                       │
│  POST /api/v1/cards/verify-payment                                   │
│  {                                                                    │
│    "encrypted_pan": { ... },         ← E2EE                         │
│    "encrypted_cvv": { ... },         ← E2EE                         │
│    "expiry_month": 12,               ← plaintext (xavfsiz)          │
│    "expiry_year": 2029,              ← plaintext (xavfsiz)          │
│    "amount": 150000,                 ← plaintext (server kerak)      │
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
│  3. DB: pan_hash bo'yicha karta topish                              │
│  4. Vault: ECIES decrypt CVV → CVV plaintext                        │
│  5. Vault: bcrypt.Compare(card.cvv_hash, CVV) → match?             │
│  6. Expiry tekshirish (server-side, plaintext)                       │
│  7. Barcha plaintext O'CHIRISH (Vault ichida)                       │
│                                                                       │
│  match == true  → Payment davom etadi (hold → capture)              │
│  match == false → 400 "Invalid card details"                        │
└───────────────────────────────────────────────────────────────────────┘
```

## Card Data Hayot Sikli (Lifecycle)

```
┌─────────────────────────────────────────────────────────────────┐
│  Ma'lumot     │ Client da   │ Tarmoqda    │ Server da   │ DB da │
│───────────────│─────────────│─────────────│─────────────│───────│
│  PAN          │ Plaintext   │ E2EE        │ Ciphertext  │ AES   │
│  (16 raqam)   │ → ECIES     │ (ciphertext)│ → Vault     │ (DEK) │
│               │ → O'CHIRISH │             │ → O'CHIRISH │       │
│               │             │             │             │       │
│  PIN          │ Plaintext   │ E2EE        │ Ciphertext  │ bcrypt│
│  (4-6 raqam)  │ → ECIES     │ (ciphertext)│ → Vault     │ hash  │
│               │ → O'CHIRISH │             │ → O'CHIRISH │       │
│               │             │             │             │       │
│  CVV          │ Plaintext   │ E2EE        │ Ciphertext  │ bcrypt│
│  (3-4 raqam)  │ → ECIES     │ (ciphertext)│ → Vault     │ hash  │
│               │ → O'CHIRISH │             │ → O'CHIRISH │       │
│               │             │             │             │       │
│  Last 4       │ ❌          │ ❌          │ Vault →     │ Plain │
│               │             │             │ hisoblaydi  │ "7890"│
│               │             │             │             │       │
│  PAN Hash     │ ❌          │ ❌          │ Vault →     │ SHA   │
│               │             │             │ hisoblaydi  │ -256  │
└─────────────────────────────────────────────────────────────────┘

Plaintext mavjud bo'lgan joylar:
  1. Client input vaqtida      → ECIES dan keyin darhol o'chiriladi
  2. Vault/HSM ichida          → operatsiya tugagach darhol o'chiriladi

  ⚠️ Boshqa HECH QAYERDA plaintext bo'lmaydi:
     ❌ Tarmoqda (E2EE)
     ❌ Server memory da (ciphertext → Vault proxy)
     ❌ DB da (AES encrypted yoki bcrypt hash)
     ❌ Logda (sensitive field mask qilinadi)
```

## E2EE Encryption (ECIES + AES-256-GCM)

Har bir karta uchun **unique DEK** (Data Encryption Key) generatsiya qilinadi.
DEK o'zi **Card KEK** (Key Encryption Key) bilan shifrlangan holda DB da saqlanadi.
Card KEK private key faqat **Vault/HSM** da.

```
Client → Server (E2EE):
  1. Client: ECIES_Encrypt(PAN, server_e2ee_public_key) → ciphertext
  2. Server: ciphertext ni Vault ga proxy (plaintext ko'rmaydi)

Vault ichida (re-encrypt for storage):
  3. ECIES decrypt → plaintext PAN
  4. Random DEK generatsiya (AES-256 key)
  5. AES-256-GCM(PAN, DEK, nonce) → encrypted_pan
  6. RSA_Encrypt(DEK, card_KEK_public) → encrypted_dek
  7. DB: encrypted_pan + encrypted_dek + nonce + key_id

Deshifrlash (karta o'qish):
  1. Vault: encrypted_dek → RSA_Decrypt(card_KEK_private) → DEK
  2. Vault: AES_Decrypt(encrypted_pan, DEK, nonce) → PAN
  3. Plaintext faqat Vault ichida

KEK ROTATE:
  Faqat DEK re-wrap (karta ma'lumoti qayta shifrlanmaydi)
```

Batafsil: [Encryption & PKI](../security/encryption.md#end-to-end-encryption-e2ee)

## PCI DSS Compliance
- **Req 3**: Card data **E2EE (ECIES)** + storage **AES-256-GCM** (DEK bilan), unique DEK per-card
- **Req 3.2**: CVV faqat bcrypt hash saqlanadi, plaintext HECH QACHON
- **Req 3.4**: PAN hech qachon plain text saqlanmaydi, server memory da ham emas (E2EE)
- **Req 3.5**: KEK private key faqat HSM/Vault da
- **Req 4**: TLS 1.3 transport + E2EE application-level
- **Req 7**: Foydalanuvchi faqat o'z kartalariga (RLS)
- **Req 8**: Strong auth + 2FA
- **Req 10**: Card data access audit logda

## Tokenizatsiya
Haqiqiy karta raqami o'rniga random token:
```go
type CardToken struct {
    Token     string    // "tok_xxxxxxxxxxxxxxxx"
    LastFour  string    // "1234"
    IsActive  bool
    ExpiresAt time.Time
}
```

## EMV Standart
- Luhn algorithm — card number validation
- Card network detection (Visa, MasterCard, UnionPay)
- 3D Secure result struct (online payment uchun)

## Operatsiyalar
- **Issue** → E2EE → Vault decrypt → AES encrypt PAN (DEK + KEK), bcrypt hash CVV/PIN → INACTIVE
- **Activate** → ACTIVE
- **Block** → BLOCKED (3 wrong PIN → auto block)
- **PlaceHold** → available_balance kamaytirish
- **CaptureHold** → partial yoki to'liq capture
- **ReleaseHold** → hold bekor qilish

## API Endpoints

| Method | Endpoint | Middleware | Tavsif |
|---|---|---|---|
| POST | `/api/v1/cards` | Session+KYC+2FA | Karta yaratish (E2EE: PAN+PIN+CVV) |
| GET | `/api/v1/cards` | Session | Kartalar ro'yxati (masked) |
| GET | `/api/v1/cards/{id}` | Session | Karta ma'lumoti (masked) |
| POST | `/api/v1/cards/{id}/activate` | Session | Kartani faollashtirish |
| POST | `/api/v1/cards/{id}/block` | Session | Kartani bloklash |
| POST | `/api/v1/cards/{id}/verify-pin` | Session | PIN tekshirish (E2EE) |
| PUT | `/api/v1/cards/{id}/pin` | Session+2FA | PIN o'zgartirish (E2EE) |
| PUT | `/api/v1/cards/{id}/limits` | Session | Limit o'zgartirish |
| POST | `/api/v1/cards/{id}/tokenize` | Session | Tokenizatsiya |
| POST | `/api/v1/cards/verify-payment` | Session | Online to'lov (E2EE: PAN+CVV) |
