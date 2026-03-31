# KYC & AML

## KYC (Know Your Customer)
Document types: PASSPORT, ID_CARD, DRIVER_LICENSE

### Status Flow
```
PENDING → APPROVED / REJECTED
```

### KYC Enforcement
- Account creation: KYC required
- Transfer > $500: KYC required
- Card issuance: KYC required
- Enforced via `kyc_required` middleware

### KYC Model
```go
type KYCVerification struct {
    DocumentType    DocumentType
    DocumentNumber  EncryptedString  // AES-256-GCM (application-level, key in Vault)
    FrontImageHash  string           // SHA-256 (integrity check)
    BackImageHash   string           // SHA-256
    EncryptedDEK    []byte           // RSA-4096(DEK, KYC KEK public) — per-document
    EncryptionNonce []byte           // AES GCM nonce
    EncryptionKeyID string           // which KEK was used
    Status          KYCStatus
    RejectionReason string
    VerifiedAt      *time.Time
    ExpiresAt       time.Time
}
```

### Envelope Encryption (For document files)
```
ENCRYPTION:
  1. Generate random DEK (new for each document)
  2. document → AES-256-GCM(document, DEK, nonce) → encrypted_doc → S3/MinIO
  3. DEK → RSA_Encrypt(DEK, kyc_KEK_public) → encrypted_dek → DB
  4. SHA-256(document) → file_hash (integrity check)

DECRYPTION:
  1. DB: encrypted_dek → RSA_Decrypt(kyc_KEK_private) → DEK
  2. S3: encrypted_doc → AES_Decrypt(encrypted_doc, DEK, nonce) → document
  3. Verify SHA-256(document) == file_hash

KEK ROTATE:
  Only DEK re-wrap — documents are not re-encrypted
```

If the DB is compromised or S3 is compromised — the KEK private key is only in Vault, data remains safe.
Details: [Encryption & PKI](../security/encryption.md#kyc-document-encryption-envelope)

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
- **CDD (Customer Due Diligence)**: KYC document verification
- **STR (Suspicious Transaction Report)**: risk > 70 → admin review
- **Record keeping**: retain for 7 years
- **Threshold reporting**: > $10,000 → automatic flag

## API Endpoints

| Method | Endpoint | Middleware |
|---|---|---|
| POST | `/api/v1/kyc/submit` | Session |
| GET | `/api/v1/kyc/status` | Session |
| GET | `/admin/aml/reviews` | Admin+IPWhitelist |
| POST | `/admin/aml/reviews/{id}/approve` | Admin+IPWhitelist |
| POST | `/admin/kyc/{id}/approve` | Admin+IPWhitelist |
