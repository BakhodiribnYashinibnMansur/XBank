# Cards — PCI DSS Compliant

## Card Aggregate
```go
type Card struct {
    AggregateRoot
    AccountID       uuid.UUID
    UserID          uuid.UUID
    EncryptedPAN    []byte               // Hybrid: AES-256-GCM (DEK bilan)
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

## Hybrid Encryption (ECDH/RSA + AES-256-GCM)

Har bir karta uchun **unique DEK** (Data Encryption Key) generatsiya qilinadi.
DEK o'zi **Card KEK** (Key Encryption Key) bilan shifrlangan holda DB da saqlanadi.
Card KEK private key faqat **Vault/HSM** da.

```
SHIFRLASH (karta yaratish):
  1. Random DEK generatsiya (AES-256 key)
  2. card_number → AES-256-GCM(card_number, DEK, nonce) → encrypted_pan
  3. DEK → RSA_Encrypt(DEK, card_KEK_public) → encrypted_dek
  4. DB: encrypted_pan + encrypted_dek + nonce + key_id

DESHIFRLASH:
  1. encrypted_dek → RSA_Decrypt(encrypted_dek, card_KEK_private) → DEK
  2. encrypted_pan → AES_Decrypt(encrypted_pan, DEK, nonce) → card_number

KEK ROTATE:
  Faqat DEK re-wrap (karta ma'lumoti qayta shifrlanmaydi)
```

Batafsil: [Encryption & PKI](../security/encryption.md#card-encryption-hybrid-ecdh--aes-256-gcm)

## PCI DSS Compliance
- **Req 3**: Card data **Hybrid encryption** (RSA + AES-256-GCM), unique DEK per-card
- **Req 3.4**: PAN hech qachon plain text saqlanmaydi
- **Req 3.5**: KEK private key faqat HSM/Vault da
- **Req 4**: TLS 1.3 transport
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
- **Issue** → Hybrid encrypt PAN (DEK + KEK), bcrypt hash CVV/PIN → INACTIVE
- **Activate** → ACTIVE
- **Block** → BLOCKED (3 wrong PIN → auto block)
- **PlaceHold** → available_balance kamaytirish
- **CaptureHold** → partial yoki to'liq capture
- **ReleaseHold** → hold bekor qilish

## API Endpoints

| Method | Endpoint | Middleware |
|---|---|---|
| POST | `/api/v1/cards` | Session+KYC+2FA |
| GET | `/api/v1/cards` | Session |
| GET | `/api/v1/cards/{id}` | Session |
| POST | `/api/v1/cards/{id}/activate` | Session |
| POST | `/api/v1/cards/{id}/block` | Session |
| PUT | `/api/v1/cards/{id}/pin` | Session+2FA |
| PUT | `/api/v1/cards/{id}/limits` | Session |
| POST | `/api/v1/cards/{id}/tokenize` | Session |
