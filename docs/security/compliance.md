# International Fintech Standards

## Standards Map

| Standard | Area | How XBank implements it |
|---|---|---|
| **PCI DSS** | Card security | AES-256, tokenization, masking, audit |
| **PSD2/SCA** | EU Open Banking | 2FA + dynamic linking |
| **ISO 20022** | Financial messaging | Transfer struct formats |
| **ISO 8583** | Card transactions | CardTransaction format |
| **ISO 4217** | Currency codes | Currency VO (code + exponent) |
| **ISO 27001** | Information security | Encryption, RBAC, audit, key rotation |
| **FATF** | AML/CFT | KYC/CDD, risk scoring, STR, 7yr retention |
| **SOC 2** | Service controls | Security, availability, integrity, privacy |
| **EMV** | Card standard | Luhn, card network detection |
| **3D Secure** | Online payment | 3DS result struct |
| **Basel III** | Banking regulation | Limits, risk exposure |
| **SWIFT** | International transfers | BIC, IBAN, reference |
| **OWASP Top 10** | Web security | All 10 threats protected |
| **GDPR** | Data protection | Anonymization, export, consent |
| **WCAG 2.1** | Accessibility | Frontend a11y |

## PCI DSS Requirements (Summary)

<!-- PCI DSS — a standard for protecting card data.
     It has 12 core requirements. -->

| Req | Requirement | XBank implementation |
|---|---|---|
| 1 | Network protection | Docker network isolation, firewall rules |
| 2 | Change default passwords | All default credentials changed |
| 3 | Protect stored card data | **Hybrid encryption** (RSA + AES-256-GCM), unique DEK per-card |
| 4 | Encrypt data transmitted over the network | **TLS 1.3**, SSL verify-full (DB) |
| 5 | Antivirus and malware protection | Server-level security, container scanning |
| 6 | Secure systems and applications | OWASP Top 10 protection, dependency scanning |
| 7 | Restrict access to only those who need it | **RLS + RBAC/ABAC**, least privilege DB users |
| 8 | Unique ID for each person | UUID based auth, **2FA/TOTP** |
| 9 | Restrict physical access | Cloud/server level (infra) |
| 10 | Monitoring and audit | **Immutable audit log**, Prometheus, Grafana |
| 11 | Security testing | Integration + Security tests |
| 12 | Security policy | Documentation, key rotation schedule |

## GDPR Compliance

<!-- GDPR — EU user data protection law -->
```
Data Retention:
  - Audit log: 7 years (FATF requirement)
  - KYC documents: 7 years (deleted afterwards)
  - Notification: 90 days (then archived)
  - Session data: 30 days (then cleaned)
  - Transaction history: 7 years

Data Subject Rights:
  - Right to Access: GET /api/v1/users/me/data-export
    → Download all personal data in JSON/CSV
  - Right to Erasure: Financial data cannot be deleted even under GDPR
    (regulatory retention requirement takes precedence)
    Personal data is anonymized:
      name → "REDACTED"
      email → hash
      phone → "***"
  - Right to Rectification: PUT /api/v1/users/me
    → Update personal data (with audit log)

Data Deletion Process (when retention period expires):
  1. pg_cron → check for expired data weekly
  2. Personal data → anonymization (not hard delete)
  3. File storage (S3) → delete encrypted files
  4. Audit log → NEVER deleted (7 years required)
```

## Audit Evidence Storage

<!-- All evidence for compliance audits is stored in the following locations -->
```
Evidence sources:
  1. audit_log table — who, when, what was done (immutable, 7 years)
  2. event_store table — all state changes (immutable)
  3. reconciliation_runs — daily check results
  4. encryption_keys — key rotation history
  5. Prometheus/Grafana — system monitoring history
  6. Git history — code changes

Audit reports:
  - Admin panel → "Compliance Reports" section
  - Monthly automatic report generation
  - Custom reports on regulator request
```
