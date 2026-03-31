# Beneficiaries — Transfer Recipients

## Model
```go
type Beneficiary struct {
    AggregateRoot
    UserID          uuid.UUID
    Name            string
    AccountNumber   string           // internal or IBAN
    BankName        string
    BankCode        string           // BIC/SWIFT
    Currency        Currency
    BeneficiaryType BeneficiaryType  // INTERNAL, EXTERNAL, INTERNATIONAL
    IsVerified      bool
    IsActive        bool             // for soft delete (false = deleted)
}
```

## Beneficiary Types

| Type | Description | Validation |
|---|---|---|
| `INTERNAL` | Account within XBank | Check account number existence |
| `EXTERNAL` | Other local bank | Bank code (MFO) + account number format |
| `INTERNATIONAL` | International transfer | IBAN format + SWIFT/BIC code validation |

## Validation Rules

<!-- When adding a beneficiary, the account number and bank code are validated -->
```
IBAN validation (for INTERNATIONAL):
  1. Length check (depends on country, e.g.: DE=22, GB=22, UZ=23)
  2. Only digits and uppercase letters
  3. Mod97 algorithm (ISO 13616) — IBAN checksum
  4. Country code (first 2 letters) existence

SWIFT/BIC validation:
  1. Length: 8 or 11 characters
  2. Format: AAAA BB CC (DDD) — bank, country, location, (branch)

Internal account validation:
  1. Check account existence within XBank
  2. Account must be in ACTIVE status
  3. Cannot add own account as beneficiary
```

## Limits

```
Per user:
  - Maximum 50 beneficiaries (active)
  - Maximum 5 new beneficiaries per day
  - New beneficiary + large amount transfer (within 24 hours) = fraud flag
```

## Soft Delete

<!-- Financial data is NEVER hard deleted.
     Deleted beneficiaries have is_active=false. -->
```
Deletion: is_active = false, deleted_at = NOW()
Result:
  - Beneficiary does not appear in the list
  - But the reference is preserved in old transfer history
  - DELETE is recorded in the audit log
```

## API

| Method | Endpoint | Middleware | Description |
|---|---|---|---|
| POST | `/api/v1/beneficiaries` | Session | Add new beneficiary |
| GET | `/api/v1/beneficiaries` | Session | User's beneficiaries |
| GET | `/api/v1/beneficiaries/{id}` | Session | Single beneficiary details |
| PUT | `/api/v1/beneficiaries/{id}` | Session | Update beneficiary (name, bank details) |
| DELETE | `/api/v1/beneficiaries/{id}` | Session | Soft delete (is_active=false) |

## Fraud Integration

<!-- Fraud checks related to new beneficiaries -->
```
Fraud rules:
  - New beneficiary + large amount transfer within 24 hours = MEDIUM risk
  - 3+ new beneficiaries per day = FLAG
  - New beneficiary from new device = FLAG + 2FA mandatory
```
