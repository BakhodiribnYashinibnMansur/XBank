# Fraud Detection

## Risk Scoring

<!-- Risk score — a score from 0 to 100.
     5 factors are added with weights (weighted).
     The weights are distributed as follows: -->

```
score = w1*velocity + w2*amount + w3*device + w4*behavior + w5*time

Weights:
  w1 = 0.30  (velocity — speed, most important)
  w2 = 0.25  (amount — sum)
  w3 = 0.20  (device — device)
  w4 = 0.15  (behavior — behavior)
  w5 = 0.10  (time — time)

Result levels:
  LOW (0-30):     → auto approve (automatic approval)
  MEDIUM (30-70): → require OTP/2FA (additional verification)
  HIGH (70-100):  → block + manual review (block + admin review)
```

### Scoring Example
```
User: from a new device, at 3:00 AM, large amount transfer

  velocity:  20/100  (1 transfer/hour — normal)              × 0.30 = 6
  amount:    80/100  ($9,000 — high)                          × 0.25 = 20
  device:    90/100  (new, unknown device)                    × 0.20 = 18
  behavior:  40/100  (normal amount pattern)                  × 0.15 = 6
  time:      70/100  (3:00 AM — unusual)                      × 0.10 = 7

  Total score = 6 + 20 + 18 + 6 + 7 = 57 → MEDIUM → 2FA requested
```

## Real-time Checks

### Velocity Checks

| Check | Threshold | Result |
|---|---|---|
| Transfers per hour | 5+ | FLAG |
| Transfers per day | 20+ | BLOCK |
| Daily amount | > daily_limit | BLOCK |
| Large amount from new device | > $1,000 | HOLD + 2FA |
| Round amount pattern | 3+ consecutive (1000, 2000, 3000) | FLAG |
| Frequent transfers to same beneficiary | 5+ / day | FLAG |

### Behavioral Analysis
- **Time anomaly:** Unusual time (e.g., user always active during the day, but transfers at night)
- **Geo anomaly:** Unusual IP geo-location (e.g., Uzbekistan → Brazil)
- **Amount anomaly:** Unusual amount pattern (e.g., always 100K, suddenly 9M)
- **Beneficiary anomaly:** New beneficiary + large amount within 24 hours

### Device Fingerprinting
```sql
CREATE TABLE device_fingerprints (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users(id),
    device_id   VARCHAR(64) NOT NULL,           -- SHA-256(navigator + screen + timezone + ...)
    ip_address  INET,
    user_agent  TEXT,
    is_trusted  BOOLEAN DEFAULT FALSE,           -- device confirmed by user
    first_seen  TIMESTAMPTZ DEFAULT NOW(),
    last_seen   TIMESTAMPTZ DEFAULT NOW(),
    login_count INTEGER DEFAULT 0                -- number of logins
);

CREATE INDEX idx_device_user ON device_fingerprints (user_id, is_trusted);
CREATE INDEX idx_device_id ON device_fingerprints (device_id);
```

**Device rules:**
- New device → notification + 2FA mandatory (first time)
- User can mark device as "trusted" → 2FA not required next time
- Multiple accounts = 1 device → FLAG (multiple accounts on one device)
- 1 account = 10+ devices → ALERT (too many devices)

## FraudCheck Model
```go
type FraudCheck struct {
    ID          uuid.UUID
    TransferID  uuid.UUID
    UserID      uuid.UUID
    RiskScore   int          // 0-100
    RiskLevel   RiskLevel    // LOW, MEDIUM, HIGH
    Checks      []FraudFlag  // which checks triggered a flag
    DeviceMatch bool         // whether the device matched or not
    Decision    string       // APPROVE, REQUIRE_2FA, BLOCK
    ReviewedBy  *uuid.UUID   // if reviewed by admin
    ReviewedAt  *time.Time
    Notes       string       // admin notes
}
```

## Fraud Case Management

<!-- HIGH risk or BLOCKed transfers are reviewed in the admin panel -->
```
Admin Panel Flow:
  1. Fraud alert received → appears in admin panel
  2. Admin views transfer details:
     - User history
     - Device information
     - Risk score breakdown (which factor gave a high score)
     - Similar fraud cases
  3. Admin makes a decision:
     a. "Approve" → transfer proceeds (risk_level lowered)
     b. "Reject" → transfer is cancelled
     c. "Freeze account" → account FROZEN, all transfers BLOCKed
  4. Recorded in audit log
```

## False Positive Handling

<!-- The fraud system may sometimes incorrectly block a legitimate transfer -->
```
For the user:
  1. Transfer blocked → "Transfer is being reviewed" message
  2. User can contact the bank
  3. Admin can review and approve the transfer

System learning:
  - Admin-approved transfers → affect future scoring
  - Confirmed device → device risk score decreases
  - Confirmed beneficiary → beneficiary risk decreases
```

## API Endpoints

| Method | Endpoint | Middleware | Description |
|---|---|---|---|
| GET | `/admin/fraud/reviews` | Admin+IPWhitelist | Fraud cases pending review |
| POST | `/admin/fraud/reviews/{id}/approve` | Admin+IPWhitelist | Approve a fraud case |
| POST | `/admin/fraud/reviews/{id}/reject` | Admin+IPWhitelist | Reject a fraud case |
| GET | `/admin/fraud/stats` | Admin+IPWhitelist | Fraud statistics |
