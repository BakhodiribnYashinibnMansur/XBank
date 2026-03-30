# KYC & AML

## KYC (Know Your Customer)
Hujjat turlari: PASSPORT, ID_CARD, DRIVER_LICENSE

### Status Flow
```
PENDING → APPROVED / REJECTED
```

### KYC Enforcement
- Account yaratish: KYC talab qilinadi
- Transfer > $500: KYC talab qilinadi
- Card chiqarish: KYC talab qilinadi
- `kyc_required` middleware orqali

### KYC Model
```go
type KYCVerification struct {
    DocumentType    DocumentType
    DocumentNumber  EncryptedString  // AES-256-GCM (application-level, key Vault da)
    FrontImageHash  string           // SHA-256 (integrity check)
    BackImageHash   string           // SHA-256
    EncryptedDEK    []byte           // RSA-4096(DEK, KYC KEK public) — per-document
    EncryptionNonce []byte           // AES GCM nonce
    EncryptionKeyID string           // qaysi KEK ishlatilgan
    Status          KYCStatus
    RejectionReason string
    VerifiedAt      *time.Time
    ExpiresAt       time.Time
}
```

### Envelope Encryption (Hujjat fayllar uchun)
```
SHIFRLASH:
  1. Random DEK generatsiya (har bir hujjat uchun yangi)
  2. document → AES-256-GCM(document, DEK, nonce) → encrypted_doc → S3/MinIO
  3. DEK → RSA_Encrypt(DEK, kyc_KEK_public) → encrypted_dek → DB
  4. SHA-256(document) → file_hash (integrity check)

DESHIFRLASH:
  1. DB: encrypted_dek → RSA_Decrypt(kyc_KEK_private) → DEK
  2. S3: encrypted_doc → AES_Decrypt(encrypted_doc, DEK, nonce) → document
  3. SHA-256(document) == file_hash tekshirish

KEK ROTATE:
  Faqat DEK lar re-wrap — hujjatlar qayta shifrlanmaydi
```

DB buzilsa yoki S3 buzilsa — KEK private faqat Vault da, ma'lumotlar xavfsiz.
Batafsil: [Encryption & PKI](../security/encryption.md#kyc-document-encryption-envelope)

## AML (Anti-Money Laundering) — FATF Compliance

### Real-time Risk Scoring (0-100)
```
score = w1*amount + w2*velocity + w3*account_age + w4*pattern + w5*country

LOW (0-30):     auto approve
MEDIUM (30-70): proceed + review queue
HIGH (70-100):  BLOCK + admin review
```

### AML Flags
| Flag | Trigger |
|---|---|
| LARGE_AMOUNT | > $10,000 |
| RAPID_SUCCESSION | 5+ transfers/hour |
| SUSPICIOUS_PATTERN | Round amounts, same recipient |
| NEW_ACCOUNT | Account < 7 days |
| HIGH_RISK_COUNTRY | Sanctioned countries |

### FATF Requirements
- **CDD (Customer Due Diligence)**: KYC hujjat tekshirish
- **STR (Suspicious Transaction Report)**: risk > 70 → admin review
- **Record keeping**: 7 yil saqlash
- **Threshold reporting**: > $10,000 → avtomatik flag

## API Endpoints

| Method | Endpoint | Middleware |
|---|---|---|
| POST | `/api/v1/kyc/submit` | Session |
| GET | `/api/v1/kyc/status` | Session |
| GET | `/admin/aml/reviews` | Admin+IPWhitelist |
| POST | `/admin/aml/reviews/{id}/approve` | Admin+IPWhitelist |
| POST | `/admin/kyc/{id}/approve` | Admin+IPWhitelist |
