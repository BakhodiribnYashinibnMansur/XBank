# Transfers

## Double-Entry Bookkeeping
Har bir transfer = 2 ta ledger entry (debit + credit). `SUM(all entries) = 0` DOIMO.

```sql
-- A dan B ga 100,000 so'm:
INSERT INTO ledger_entries (transfer_id, account_id, side, amount_minor)
VALUES
  ('txn_001', 'account_A', 'DEBIT',  100000),
  ('txn_001', 'account_B', 'CREDIT', 100000);
```

## Transfer State Machine

```
INITIATED → PENDING → PROCESSING → COMPLETED
                                 → FAILED
                                 → REVERSED (faqat COMPLETED dan)
```

Noto'g'ri o'tishlar cheklanadi:
```go
func (s TransferStatus) CanTransitionTo(next TransferStatus) bool
```

## Saga Pattern (Orchestrator)

```
Step 1:  Fraud Check         → Compensate: -
Step 2:  AML Screening       → Compensate: -
Step 3:  Lock Source Account  → Compensate: Unlock
Step 4:  Validate Balance     → Compensate: Unlock
Step 5:  Lock Target Account  → Compensate: Unlock both
Step 6:  Debit Source         → Compensate: Refund Source
Step 7:  Credit Target        → Compensate: Reverse Credit + Refund
Step 8:  Create Ledger        → Compensate: Void entries
Step 9:  Complete Transfer    → Compensate: Mark Failed
Step 10: Emit Events          → -
```

Agar har qanday step fail → compensate previous steps.

## Idempotency
```
Client → POST /transfer {idempotency_key: "abc-123", amount: 50000}
Server:
  1. key bazada bormi? → HA → eski natijani qaytar
  2. YO'Q → tranzaksiyani bajar, natijani saqla
```
Agar tarmoq uzilsa va client qayta yubosa — pul IKKI MARTA yechilmaydi.

## Transaction Signing (ECDSA per-user)

Har bir foydalanuvchi o'z **ECDSA P-256 private key** bilan transfer so'rovini imzolaydi.
Server faqat **public key** bilan tekshiradi — private key hech qachon serverda bo'lmaydi.

### Flow
```
1. Client → payload yaratadi:
   payload = "idempotency_key|from|to|amount|currency|timestamp"

2. Client → private key bilan imzolaydi:
   signature = ECDSA_Sign(private_key, SHA256(payload))

3. Client → so'rov yuboradi:
   Header: X-Signature: base64(signature)
   Header: X-Signing-Key-ID: "key_uuid"

4. Server → user_signing_keys dan public key oladi
5. Server → ECDSA_Verify(public_key, SHA256(payload), signature)
```

### Nima uchun HMAC emas?
- **HMAC**: shared secret — server buzilsa barcha userlar compromise
- **ECDSA**: server faqat public key biladi — private key client da qoladi

Batafsil: [Encryption & PKI](../security/encryption.md#transfer-signing-ecdsa-per-user)

## Transaction Reversal
Original tranzaksiya O'CHIRILMAYDI — teskari tranzaksiya yaratiladi:
```
A → B: $100 (original, COMPLETED)
B → A: $100 (reversal, reference = original_id)
Ledger: 4 ta entry (2 original + 2 reversal)
```

## Scheduled Transactions
```sql
frequency: 'ONCE', 'DAILY', 'WEEKLY', 'MONTHLY'
next_run_at: keyingi bajarilish vaqti
```
pg_cron yoki app-level scheduler har daqiqa tekshiradi.

## Isolation Level: SERIALIZABLE
```sql
BEGIN ISOLATION LEVEL SERIALIZABLE;
  SELECT * FROM accounts WHERE id = $1 FOR UPDATE;
  -- balance check, debit, credit
COMMIT;
```

## Deadlock Prevention
Account'larni UUID tartibida lock qilish:
```go
if fromID > toID { lock(toID); lock(fromID) }
else { lock(fromID); lock(toID) }
```

## API Endpoints

| Method | Endpoint | Middleware |
|---|---|---|
| POST | `/api/v1/transfers` | Session+KYC+Idempotency+ECDSA+2FA* |
| GET | `/api/v1/transfers/{id}` | Session |
| GET | `/api/v1/accounts/{id}/transfers` | Session |

*2FA faqat > $1000 transfer uchun
