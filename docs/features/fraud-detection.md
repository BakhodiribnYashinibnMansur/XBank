# Fraud Detection

## Risk Scoring

<!-- Risk score — 0 dan 100 gacha bo'lgan ball.
     5 ta omil vaznli (weighted) qo'shiladi.
     Vaznlar quyidagicha taqsimlangan: -->

```
score = w1*velocity + w2*amount + w3*device + w4*behavior + w5*time

Vaznlar (weights):
  w1 = 0.30  (velocity — tezlik, eng muhim)
  w2 = 0.25  (amount — summa)
  w3 = 0.20  (device — qurilma)
  w4 = 0.15  (behavior — xatti-harakat)
  w5 = 0.10  (time — vaqt)

Natija darajalari:
  LOW (0-30):     → auto approve (avtomatik tasdiqlash)
  MEDIUM (30-70): → require OTP/2FA (qo'shimcha tasdiqlash)
  HIGH (70-100):  → block + manual review (bloklash + admin ko'rib chiqish)
```

### Scoring Misol
```
Foydalanuvchi: yangi qurilmadan, tunda 3:00 da, katta summa transfer

  velocity:  20/100  (1 ta transfer/soat — normal)     × 0.30 = 6
  amount:    80/100  ($9,000 — yuqori)                  × 0.25 = 20
  device:    90/100  (yangi, noma'lum qurilma)           × 0.20 = 18
  behavior:  40/100  (oddiy summa pattern)               × 0.15 = 6
  time:      70/100  (tunda 3:00 — g'ayrioddiy)         × 0.10 = 7

  Jami score = 6 + 20 + 18 + 6 + 7 = 57 → MEDIUM → 2FA so'raladi
```

## Real-time Tekshiruvlar

### Velocity Checks (Tezlik tekshiruvlari)

| Tekshiruv | Threshold | Natija |
|---|---|---|
| Transfer soni / soat | 5+ | FLAG |
| Transfer soni / kun | 20+ | BLOCK |
| Kunlik summa | > daily_limit | BLOCK |
| Yangi qurilmadan katta summa | > $1,000 | HOLD + 2FA |
| Round amount pattern | 3+ ketma-ket (1000, 2000, 3000) | FLAG |
| Bir xil beneficiary ga tez-tez | 5+ / kun | FLAG |

### Behavioral Analysis (Xatti-harakat tahlili)
- **Vaqt anomaliya:** Odatiy bo'lmagan vaqt (masalan, foydalanuvchi doim kunduzi, lekin tunda transfer)
- **Geo anomaliya:** Odatiy bo'lmagan IP geo-lokatsiya (masalan, O'zbekiston → Braziliya)
- **Summa anomaliya:** Odatiy bo'lmagan summa pattern (masalan, doim 100K, birdan 9M)
- **Beneficiary anomaliya:** Yangi beneficiary + 24 soat ichida katta summa

### Device Fingerprinting
```sql
CREATE TABLE device_fingerprints (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users(id),
    device_id   VARCHAR(64) NOT NULL,           -- SHA-256(navigator + screen + timezone + ...)
    ip_address  INET,
    user_agent  TEXT,
    is_trusted  BOOLEAN DEFAULT FALSE,           -- foydalanuvchi tasdiqlagan qurilma
    first_seen  TIMESTAMPTZ DEFAULT NOW(),
    last_seen   TIMESTAMPTZ DEFAULT NOW(),
    login_count INTEGER DEFAULT 0                -- necha marta kirgan
);

CREATE INDEX idx_device_user ON device_fingerprints (user_id, is_trusted);
CREATE INDEX idx_device_id ON device_fingerprints (device_id);
```

**Device qoidalari:**
- Yangi device → notification + 2FA mandatory (birinchi marta)
- Foydalanuvchi qurilmani "ishonchli" deb belgilashi mumkin → keyingi safar 2FA shart emas
- Multiple accounts = 1 device → FLAG (bitta qurilmada bir nechta hisob)
- 1 account = 10+ devices → ALERT (juda ko'p qurilma)

## FraudCheck Model
```go
type FraudCheck struct {
    ID          uuid.UUID
    TransferID  uuid.UUID
    UserID      uuid.UUID
    RiskScore   int          // 0-100
    RiskLevel   RiskLevel    // LOW, MEDIUM, HIGH
    Checks      []FraudFlag  // qaysi tekshiruvlar flag berdi
    DeviceMatch bool         // qurilma mos keldi yoki yo'q
    Decision    string       // APPROVE, REQUIRE_2FA, BLOCK
    ReviewedBy  *uuid.UUID   // admin ko'rib chiqqan bo'lsa
    ReviewedAt  *time.Time
    Notes       string       // admin izohi
}
```

## Fraud Case Management

<!-- HIGH risk yoki BLOCK qilingan transfer lar admin panel da ko'rib chiqiladi -->
```
Admin Panel Flow:
  1. Fraud alert keldi → admin panel da ko'rinadi
  2. Admin transfer tafsilotlarini ko'radi:
     - Foydalanuvchi tarixi
     - Device ma'lumoti
     - Risk score taqsimoti (qaysi omil yuqori ball berdi)
     - O'xshash fraud case lar
  3. Admin qaror qabul qiladi:
     a. "Tasdiqlash" → transfer davom etadi (risk_level pastga tushadi)
     b. "Rad etish" → transfer bekor qilinadi
     c. "Account muzlatish" → account FROZEN, barcha transfer lar BLOCK
  4. Audit log ga qayd qilinadi
```

## False Positive Handling

<!-- Fraud tizimi ba'zan to'g'ri transfer ni noto'g'ri bloklashi mumkin -->
```
Foydalanuvchi uchun:
  1. Transfer bloklandi → "Transfer tekshirilmoqda" xabari
  2. Foydalanuvchi bank ga murojaat qilishi mumkin
  3. Admin transfer ni tekshirib, tasdiqlashi mumkin

Tizim o'rganishi:
  - Admin tasdiqlagan transfer lar → future scoring ga ta'sir
  - Tasdiqlangan qurilma → device risk score pasayadi
  - Tasdiqlangan beneficiary → beneficiary risk pasayadi
```

## API Endpoints

| Method | Endpoint | Middleware | Tavsif |
|---|---|---|---|
| GET | `/admin/fraud/reviews` | Admin+IPWhitelist | Ko'rib chiqish kerak bo'lgan fraud case lar |
| POST | `/admin/fraud/reviews/{id}/approve` | Admin+IPWhitelist | Fraud case ni tasdiqlash |
| POST | `/admin/fraud/reviews/{id}/reject` | Admin+IPWhitelist | Fraud case ni rad etish |
| GET | `/admin/fraud/stats` | Admin+IPWhitelist | Fraud statistikasi |
