# Transfers

## Double-Entry Bookkeeping
Every transfer = 2 ledger entries (debit + credit). `SUM(all entries) = 0` ALWAYS.

```sql
-- From A to B, 100,000 so'm:
INSERT INTO ledger_entries (transfer_id, account_id, side, amount_minor)
VALUES
  ('txn_001', 'account_A', 'DEBIT',  100000),
  ('txn_001', 'account_B', 'CREDIT', 100000);
```

## Transfer State Machine

```
INITIATED → PENDING → PROCESSING → COMPLETED
                                 → FAILED
                                 → REVERSED (only from COMPLETED)
```

Invalid transitions are restricted:
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

If any step fails → compensate previous steps.

## Idempotency
```
Client → POST /transfer {idempotency_key: "abc-123", amount: 50000}
Server:
  1. Does key exist in DB? → YES → return previous result
  2. NO → execute transaction, save result
```
If the network drops and the client retries — funds will NOT be deducted TWICE.

## Transaction Signing (ECDSA per-user)

Each user signs the transfer request with their own **ECDSA P-256 private key**.
The server only verifies with the **public key** — the private key is never on the server.

### Flow
```
1. Client → creates payload:
   payload = "idempotency_key|from|to|amount|currency|timestamp"

2. Client → signs with private key:
   signature = ECDSA_Sign(private_key, SHA256(payload))

3. Client → sends request:
   Header: X-Signature: base64(signature)
   Header: X-Signing-Key-ID: "key_uuid"

4. Server → retrieves public key from user_signing_keys
5. Server → ECDSA_Verify(public_key, SHA256(payload), signature)
```

### Why not HMAC?
- **HMAC**: shared secret — if server is compromised, all users are compromised
- **ECDSA**: server only knows public key — private key stays on the client

Details: [Encryption & PKI](../security/encryption.md#transfer-signing-ecdsa-per-user)

## Transaction Reversal
The original transaction is NOT DELETED — a reverse transaction is created:
```
A → B: $100 (original, COMPLETED)
B → A: $100 (reversal, reference = original_id)
Ledger: 4 entries (2 original + 2 reversal)
```

## Scheduled Transactions
```sql
frequency: 'ONCE', 'DAILY', 'WEEKLY', 'MONTHLY'
next_run_at: next execution time
```
pg_cron or app-level scheduler checks every minute.

## Isolation Level: SERIALIZABLE
```sql
BEGIN ISOLATION LEVEL SERIALIZABLE;
  SELECT * FROM accounts WHERE id = $1 FOR UPDATE;
  -- balance check, debit, credit
COMMIT;
```

## Deadlock Prevention
Lock accounts in UUID order:
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

*2FA only for transfers > $1000
