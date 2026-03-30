# Beneficiaries — Transfer Qiluvchilar

## Model
```go
type Beneficiary struct {
    AggregateRoot
    UserID          uuid.UUID
    Name            string
    AccountNumber   string           // internal yoki IBAN
    BankName        string
    BankCode        string           // BIC/SWIFT
    Currency        Currency
    BeneficiaryType BeneficiaryType  // INTERNAL, EXTERNAL, INTERNATIONAL
    IsVerified      bool
    IsActive        bool             // soft delete uchun (false = o'chirilgan)
}
```

## Beneficiary Turlari

| Tur | Tavsif | Validatsiya |
|---|---|---|
| `INTERNAL` | XBank ichidagi hisob | Account number mavjudligini tekshirish |
| `EXTERNAL` | Boshqa mahalliy bank | Bank code (MFO) + account number format |
| `INTERNATIONAL` | Xalqaro o'tkazma | IBAN format + SWIFT/BIC code validatsiya |

## Validatsiya Qoidalari

<!-- Beneficiary qo'shishda account raqami va bank kodi tekshiriladi -->
```
IBAN validatsiya (INTERNATIONAL uchun):
  1. Uzunlik tekshirish (davlatga qarab, masalan: DE=22, GB=22, UZ=23)
  2. Faqat raqam va katta harflar
  3. Mod97 algoritmi (ISO 13616) — IBAN checksum
  4. Davlat kodi (dastlabki 2 harf) mavjudligi

SWIFT/BIC validatsiya:
  1. Uzunlik: 8 yoki 11 belgi
  2. Format: AAAA BB CC (DDD) — bank, country, location, (branch)

Internal account validatsiya:
  1. XBank ichida account mavjudligini tekshirish
  2. Account ACTIVE statusda bo'lishi kerak
  3. O'z hisobiga beneficiary qo'shib bo'lmaydi
```

## Limitlar

```
Har bir foydalanuvchi uchun:
  - Maksimum 50 ta beneficiary (faol)
  - Bir kunda maksimum 5 ta yangi beneficiary qo'shish
  - Yangi beneficiary + katta summa transfer (24 soat ichida) = fraud flag
```

## Soft Delete

<!-- Moliyaviy ma'lumotlar HECH QACHON hard delete qilinmaydi.
     O'chirilgan beneficiary is_active=false bo'ladi. -->
```
O'chirish: is_active = false, deleted_at = NOW()
Natija:
  - Beneficiary ro'yxatda ko'rinmaydi
  - Lekin eski transfer tarixida havolasi saqlanadi
  - Audit log da DELETE qayd qilinadi
```

## API

| Method | Endpoint | Middleware | Tavsif |
|---|---|---|---|
| POST | `/api/v1/beneficiaries` | Session | Yangi beneficiary qo'shish |
| GET | `/api/v1/beneficiaries` | Session | Foydalanuvchi beneficiary lari |
| GET | `/api/v1/beneficiaries/{id}` | Session | Bitta beneficiary ma'lumoti |
| PUT | `/api/v1/beneficiaries/{id}` | Session | Beneficiary yangilash (name, bank ma'lumotlari) |
| DELETE | `/api/v1/beneficiaries/{id}` | Session | Soft delete (is_active=false) |

## Fraud Integration

<!-- Yangi beneficiary bilan bog'liq fraud tekshiruvlari -->
```
Fraud qoidalari:
  - Yangi beneficiary + 24 soat ichida katta summa transfer = MEDIUM risk
  - Bir kunda 3+ yangi beneficiary = FLAG
  - Yangi qurilmadan yangi beneficiary = FLAG + 2FA mandatory
```
