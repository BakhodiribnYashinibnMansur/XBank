# Xalqaro Fintech Standartlari

## Standartlar Xaritasi

| Standart | Soha | XBank'da qanday |
|---|---|---|
| **PCI DSS** | Karta xavfsizligi | AES-256, tokenizatsiya, masking, audit |
| **PSD2/SCA** | EU Open Banking | 2FA + dynamic linking |
| **ISO 20022** | Moliyaviy xabarlar | Transfer struct formatlari |
| **ISO 8583** | Karta tranzaksiyalari | CardTransaction format |
| **ISO 4217** | Valyuta kodlari | Currency VO (code + exponent) |
| **ISO 27001** | Axborot xavfsizligi | Encryption, RBAC, audit, key rotation |
| **FATF** | AML/CFT | KYC/CDD, risk scoring, STR, 7yr retention |
| **SOC 2** | Service controls | Security, availability, integrity, privacy |
| **EMV** | Karta standarti | Luhn, card network detection |
| **3D Secure** | Online to'lov | 3DS result struct |
| **Basel III** | Bank regulatsiya | Limitlar, risk exposure |
| **SWIFT** | Xalqaro o'tkazmalar | BIC, IBAN, reference |
| **OWASP Top 10** | Web xavfsizlik | Barcha 10 threat himoyalangan |
| **GDPR** | Data protection | Anonymization, export, consent |
| **WCAG 2.1** | Accessibility | Frontend a11y |

## PCI DSS Talablari (Qisqacha)

<!-- PCI DSS — karta ma'lumotlarini himoya qilish standarti.
     12 ta asosiy talab (requirement) bor. -->

| Req | Talab | XBank implementatsiyasi |
|---|---|---|
| 1 | Tarmoq himoyasi | Docker network isolation, firewall rules |
| 2 | Default parollarni o'zgartirish | Barcha default credentials o'zgartirilgan |
| 3 | Saqlangan karta ma'lumotlarini himoyalash | **Hybrid encryption** (RSA + AES-256-GCM), unique DEK per-card |
| 4 | Tarmoq orqali uzatiladigan ma'lumotni shifrlash | **TLS 1.3**, SSL verify-full (DB) |
| 5 | Antivirus va zararli dastur himoyasi | Server-level security, container scanning |
| 6 | Xavfsiz tizim va dasturlar | OWASP Top 10 himoya, dependency scanning |
| 7 | Faqat zarur kishilarga kirish | **RLS + RBAC/ABAC**, least privilege DB users |
| 8 | Har bir kishiga unique ID | UUID based auth, **2FA/TOTP** |
| 9 | Fizik kirish cheklash | Cloud/server level (infra) |
| 10 | Monitoring va audit | **Immutable audit log**, Prometheus, Grafana |
| 11 | Xavfsizlik testlari | Integration + Security testlar |
| 12 | Xavfsizlik siyosati | Dokumentatsiya, key rotation jadvali |

## GDPR Compliance

<!-- GDPR — EU foydalanuvchi ma'lumotlarini himoya qilish qonuni -->
```
Data Retention (Saqlash muddati):
  - Audit log: 7 yil (FATF talabi)
  - KYC hujjatlar: 7 yil (keyin o'chiriladi)
  - Notification: 90 kun (keyin arxiv)
  - Session ma'lumotlari: 30 kun (keyin tozalanadi)
  - Tranzaksiya tarixi: 7 yil

Data Subject Rights (Foydalanuvchi huquqlari):
  - Right to Access: GET /api/v1/users/me/data-export
    → Barcha shaxsiy ma'lumotlarni JSON/CSV da yuklab olish
  - Right to Erasure: Moliyaviy ma'lumotlar GDPR bo'yicha ham o'chirilmaydi
    (regulyator saqlash talabi ustunlik oladi)
    Shaxsiy ma'lumotlar anonymizatsiya qilinadi:
      name → "REDACTED"
      email → hash
      phone → "***"
  - Right to Rectification: PUT /api/v1/users/me
    → Shaxsiy ma'lumotlarni yangilash (audit log bilan)

Data Deletion Jarayoni (Saqlash muddati tugaganda):
  1. pg_cron → har hafta expired data ni tekshirish
  2. Shaxsiy ma'lumotlar → anonymizatsiya (hard delete emas)
  3. Fayl saqlash (S3) → encrypted file o'chirish
  4. Audit log → HECH QACHON o'chirilmaydi (7 yil shart)
```

## Audit Evidence Saqlash

<!-- Compliance audit uchun barcha dalillar quyidagi joylarda saqlanadi -->
```
Dalil manbalar:
  1. audit_log jadvali — kim, qachon, nima qildi (immutable, 7 yil)
  2. event_store jadvali — barcha holat o'zgarishlari (immutable)
  3. reconciliation_runs — kunlik tekshiruv natijalari
  4. encryption_keys — key rotation tarixi
  5. Prometheus/Grafana — tizim monitoring tarixi
  6. Git tarixi — kod o'zgarishlari

Audit hisoboti:
  - Admin panel → "Compliance Reports" bo'limi
  - Oyllik avtomatik hisobot generatsiyasi
  - Regulyator so'rovi bo'yicha maxsus hisobot
```
