# Frontend Ishlash Flow (Test UI)

## Tech
- Vanilla HTML/JS + Tailwind CDN
- SPA (Single Page Application)
- Fiber static file serving + Go embed

## Security
```
CSP headers:        Content-Security-Policy: default-src 'self'
CSRF token:         Har bir POST/PUT/DELETE formda
Session timeout:    15 min inactivity → auto logout
PAN masking:        **** **** **** 1234
Clipboard block:    Card number nusxalanmaydi
Screen protection:  Sensitive fields: user-select: none + overlay
localStorage:       Faqat access_token (test uchun)
                    Card number, CVV HECH QACHON saqlanmaydi
```

## UX Standartlari
- WCAG 2.1 accessibility (aria-labels, keyboard nav, color contrast)
- Responsive (mobil + desktop)
- Skeleton loading
- Tushunarli xato xabarlari
- Transaction confirmation modal

## Pages Flow

```
┌─────────────────────────────────────────────┐
│                  LOGIN                       │
│  email + password                            │
│  ├── 2FA yoqilgan? → TOTP code input        │
│  └── Success → Dashboard                    │
└─────────────────────────────────────────────┘
          │
          ▼
┌─────────────────────────────────────────────┐
│              DASHBOARD                       │
│  - Barcha accountlar + balanslar             │
│  - Oxirgi 5 tranzaksiya                     │
│  - Real-time notifications (SSE)             │
│  - Quick transfer button                     │
└─────────────────────────────────────────────┘
    │           │           │           │
    ▼           ▼           ▼           ▼
 Accounts   Transfers    Cards    Beneficiaries
```

### Login Flow
```
1. User email + password kiritadi
2. POST /api/v1/auth/login
3. 2FA yoqilgan? → TOTP code input → POST /api/v1/auth/2fa/verify
4. Response: {access_token, refresh_token}
5. access_token → localStorage (test uchun)
6. Redirect → Dashboard
```

### Transfer Flow (UI)
```
1. Transfer sahifasiga o'tish
2. Beneficiary tanlash (dropdown)
3. Summa va valyuta kiritish
4. Komissiya ko'rsatiladi (transparency)
5. "Tasdiqlash" tugmasi → Confirmation Modal
6. Agar > $1000 → TOTP code so'raladi
7. POST /api/v1/transfers (Idempotency-Key + X-Signature header)
8. Loading spinner
9. Success → "Transfer muvaffaqiyatli!" + notification
10. Balans real-time yangilanadi (SSE event)
```

### Card Issue Flow (UI)
```
1. Cards sahifasi → "Yangi karta" tugmasi
2. Account tanlash, karta turi (DEBIT/VIRTUAL)
3. 2FA tasdiqlash (TOTP code)
4. POST /api/v1/cards
5. Karta yaratildi → masked number ko'rsatiladi
6. "Aktivatsiya" tugmasi → POST /activate
```

### Real-Time Notifications (SSE)
<!-- MUHIM: EventSource API custom headers qo'llab-quvvatlamaydi.
     Shuning uchun auth uchun query parameter yoki cookie-based auth ishlatiladi.
     Agar custom headers kerak bo'lsa, fetch-based SSE polyfill ishlatish mumkin. -->
```javascript
// Variant 1: Query parameter orqali auth (sodda)
const eventSource = new EventSource('/api/v1/notifications/stream?token=' + token);

// Variant 2: Cookie-based auth (xavfsizroq, production uchun tavsiya)
// Server HttpOnly cookie orqali sessiyani tekshiradi
// const eventSource = new EventSource('/api/v1/notifications/stream');

eventSource.onmessage = (event) => {
    const notification = JSON.parse(event.data);
    showToast(notification.title, notification.body);
    updateBalance(); // balansni yangilash
};

// SSE uzilganda qayta ulanish (reconnection)
eventSource.onerror = () => {
    // EventSource avtomatik qayta ulanadi (default: 3s)
    // Agar manual boshqarish kerak bo'lsa:
    console.warn('SSE connection lost, reconnecting...');
};
```

### Auto-Logout (Session Timeout)
```javascript
let inactivityTimer;
const TIMEOUT = 15 * 60 * 1000; // 15 min

function resetTimer() {
    clearTimeout(inactivityTimer);
    inactivityTimer = setTimeout(logout, TIMEOUT);
}

document.addEventListener('mousemove', resetTimer);
document.addEventListener('keypress', resetTimer);
```

### Token Refresh
```javascript
// Access token 15 min → refresh before expiry
setInterval(async () => {
    const response = await fetch('/api/v1/auth/refresh', {
        method: 'POST',
        body: JSON.stringify({ refresh_token })
    });
    // Yangi tokenlarni saqlash
}, 14 * 60 * 1000); // 14 min
```

## API Client (JS)
```javascript
class XBankAPI {
    async register(email, password, firstName, lastName)
    async login(email, password, totpCode)
    async createAccount(currency)
    async transfer(fromId, toId, amount, currency)
    async issueCard(accountId, cardType)
    async getStatement(accountId, from, to)
    async getNotifications()
    // ... barcha endpointlar
}
```
