# Frontend Operation Flow (Test UI)

## Tech
- Vanilla HTML/JS + Tailwind CDN
- SPA (Single Page Application)
- Fiber static file serving + Go embed

## Security
```
CSP headers:        Content-Security-Policy: default-src 'self'
CSRF token:         Included in every POST/PUT/DELETE form
Session timeout:    15 min inactivity → auto logout
PAN masking:        **** **** **** 1234
Clipboard block:    Card number cannot be copied
Screen protection:  Sensitive fields: user-select: none + overlay
localStorage:       Only access_token (for testing)
                    Card number, CVV are NEVER stored
```

## UX Standards
- WCAG 2.1 accessibility (aria-labels, keyboard nav, color contrast)
- Responsive (mobile + desktop)
- Skeleton loading
- Clear error messages
- Transaction confirmation modal

## Pages Flow

```
┌─────────────────────────────────────────────┐
│                  LOGIN                       │
│  email + password                            │
│  ├── 2FA enabled? → TOTP code input         │
│  └── Success → Dashboard                    │
└─────────────────────────────────────────────┘
          │
          ▼
┌─────────────────────────────────────────────┐
│              DASHBOARD                       │
│  - All accounts + balances                   │
│  - Last 5 transactions                       │
│  - Real-time notifications (SSE)             │
│  - Quick transfer button                     │
└─────────────────────────────────────────────┘
    │           │           │           │
    ▼           ▼           ▼           ▼
 Accounts   Transfers    Cards    Beneficiaries
```

### Login Flow
```
1. User enters email + password
2. POST /api/v1/auth/login
3. 2FA enabled? → TOTP code input → POST /api/v1/auth/2fa/verify
4. Response: {access_token, refresh_token}
5. access_token → localStorage (for testing)
6. Redirect → Dashboard
```

### Transfer Flow (UI)
```
1. Navigate to Transfer page
2. Select beneficiary (dropdown)
3. Enter amount and currency
4. Commission is displayed (transparency)
5. "Confirm" button → Confirmation Modal
6. If > $1000 → TOTP code is requested
7. POST /api/v1/transfers (Idempotency-Key + X-Signature header)
8. Loading spinner
9. Success → "Transfer successful!" + notification
10. Balance updates in real-time (SSE event)
```

### Card Issue Flow (UI)
```
1. Cards page → "New card" button
2. Select account, card type (DEBIT/VIRTUAL)
3. 2FA confirmation (TOTP code)
4. POST /api/v1/cards
5. Card created → masked number is displayed
6. "Activate" button → POST /activate
```

### Real-Time Notifications (SSE)
<!-- IMPORTANT: EventSource API does not support custom headers.
     Therefore, query parameter or cookie-based auth is used for authentication.
     If custom headers are needed, a fetch-based SSE polyfill can be used. -->
```javascript
// Option 1: Auth via query parameter (simple)
const eventSource = new EventSource('/api/v1/notifications/stream?token=' + token);

// Option 2: Cookie-based auth (more secure, recommended for production)
// Server verifies session via HttpOnly cookie
// const eventSource = new EventSource('/api/v1/notifications/stream');

eventSource.onmessage = (event) => {
    const notification = JSON.parse(event.data);
    showToast(notification.title, notification.body);
    updateBalance(); // refresh balance
};

// Reconnection when SSE disconnects
eventSource.onerror = () => {
    // EventSource automatically reconnects (default: 3s)
    // If manual control is needed:
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
    // Save new tokens
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
    // ... all endpoints
}
```
