package app

import (
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/admin/generic/featureflag"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/admin/supporting/integration"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/admin/supporting/sitesetting"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/admin/supporting/statistics"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/banking/core/account"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/banking/core/card"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/banking/core/ledger"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/banking/core/transfer"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/banking/supporting/beneficiary"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/banking/supporting/exchange"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/banking/supporting/fraud"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/banking/supporting/kyc"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/banking/supporting/reconciliation"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/content/generic/notification"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/content/generic/translation"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/content/supporting/announcement"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/iam/core/challenge"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/iam/generic/authz"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/iam/generic/session"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/iam/generic/user"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/iam/supporting/audit"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/iam/supporting/contact"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/ops/generic/metric"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/ops/generic/ratelimit"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/ops/generic/systemerror"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/ops/supporting/errorcode"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/ops/supporting/iprule"
	appKernel "github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/application"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/domain"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/config"
	infraRedis "github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/db/redis"
	infraCrypto "github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/security/crypto"
	infraAuth "github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/security/jwt"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DDDBoundedContexts holds all Bounded Context instances.
type DDDBoundedContexts struct {
	// Banking — Core
	Account  *account.BoundedContext
	Card     *card.BoundedContext
	Transfer *transfer.BoundedContext
	Ledger   *ledger.BoundedContext

	// Banking — Supporting
	Beneficiary    *beneficiary.BoundedContext
	Exchange       *exchange.BoundedContext
	KYC            *kyc.BoundedContext
	Fraud          *fraud.BoundedContext
	Reconciliation *reconciliation.BoundedContext

	// IAM — Core
	Challenge *challenge.BoundedContext

	// IAM — Generic
	User    *user.BoundedContext
	Session *session.BoundedContext
	Authz   *authz.BoundedContext

	// IAM — Supporting
	Contact *contact.BoundedContext
	Audit   *audit.BoundedContext

	// Admin — Generic
	FeatureFlag *featureflag.BoundedContext

	// Admin — Supporting
	SiteSetting *sitesetting.BoundedContext
	Statistics  *statistics.BoundedContext
	Integration *integration.BoundedContext

	// Content — Generic
	Notification *notification.BoundedContext
	Translation  *translation.BoundedContext

	// Content — Supporting
	Announcement *announcement.BoundedContext

	// Ops — Generic
	SystemError *systemerror.BoundedContext
	RateLimit   *ratelimit.BoundedContext
	Metric      *metric.BoundedContext

	// Ops — Supporting
	ErrorCode *errorcode.BoundedContext
	IPRule    *iprule.BoundedContext
}

// NewDDDBoundedContexts instantiates all Bounded Contexts with their dependencies.
func NewDDDBoundedContexts(
	pool *pgxpool.Pool,
	txManager domain.TxManager,
	publisher domain.EventPublisher,
	eventBus appKernel.EventBus,
	cfg *config.Config,
	jwtService *infraAuth.JWTService,
	totpService *infraAuth.TOTPService,
	sessionCache *infraRedis.SessionCache,
	loginLimiter *infraRedis.LoginLimiter,
	challengeCache *infraRedis.ChallengeCache,
	cardEncryptor *infraCrypto.AESEncryptor,
	tokenizer *infraCrypto.Tokenizer,
	auditLog domain.AuditLog,
) *DDDBoundedContexts {
	// IAM — Generic
	userBC := user.NewBoundedContext(pool)
	sessionBC := session.NewBoundedContext(pool, userBC.AuthPort, jwtService, totpService, sessionCache, loginLimiter)

	// IAM — Core
	challengeBC := challenge.NewBoundedContext(pool, userBC.AuthPort, challengeCache)

	// IAM — Generic (authz)
	authzBC := authz.NewBoundedContext(pool)

	// IAM — Supporting
	contactBC := contact.NewBoundedContext(pool)
	auditBC := audit.NewBoundedContext(pool)

	// Banking — Core
	ledgerBC := ledger.NewBoundedContext(pool)
	accountBC := account.NewBoundedContext(pool, txManager, publisher, cfg.Kafka.Topics, auditLog)
	transferBC := transfer.NewBoundedContext(pool, accountBC.TransferPort, txManager, publisher, cfg.Kafka.Topics)
	cardBC := card.NewBoundedContext(pool, cardEncryptor, tokenizer)

	// Banking — Supporting
	beneficiaryBC := beneficiary.NewBoundedContext(pool)
	exchangeBC := exchange.NewBoundedContext(pool)
	kycBC := kyc.NewBoundedContext(pool)
	fraudBC := fraud.NewBoundedContext(pool)
	reconBC := reconciliation.NewBoundedContext(accountBC.Reader, ledgerBC.Reader)

	// Admin
	featureFlagBC := featureflag.NewBoundedContext(pool, eventBus)
	siteSettingBC := sitesetting.NewBoundedContext(pool, eventBus)
	statisticsBC := statistics.NewBoundedContext(pool)

	// Content
	notificationBC := notification.NewBoundedContext(pool, eventBus)
	translationBC := translation.NewBoundedContext(pool, eventBus)
	announcementBC := announcement.NewBoundedContext(pool, eventBus)

	// Ops
	systemErrorBC := systemerror.NewBoundedContext(pool, eventBus)
	rateLimitBC := ratelimit.NewBoundedContext(pool)
	metricBC := metric.NewBoundedContext(pool)
	errorCodeBC := errorcode.NewBoundedContext(pool)
	ipRuleBC := iprule.NewBoundedContext(pool)

	// Admin — Supporting (integration)
	integrationBC := integration.NewBoundedContext(pool)

	return &DDDBoundedContexts{
		Account:        accountBC,
		Card:           cardBC,
		Transfer:       transferBC,
		Ledger:         ledgerBC,
		Beneficiary:    beneficiaryBC,
		Exchange:       exchangeBC,
		KYC:            kycBC,
		Fraud:          fraudBC,
		Reconciliation: reconBC,
		Challenge:      challengeBC,
		User:           userBC,
		Session:        sessionBC,
		Authz:          authzBC,
		Contact:        contactBC,
		Audit:          auditBC,
		FeatureFlag:    featureFlagBC,
		SiteSetting:    siteSettingBC,
		Statistics:     statisticsBC,
		Notification:   notificationBC,
		Translation:    translationBC,
		Announcement:   announcementBC,
		SystemError:    systemErrorBC,
		RateLimit:      rateLimitBC,
		Metric:         metricBC,
		ErrorCode:      errorCodeBC,
		IPRule:         ipRuleBC,
		Integration:    integrationBC,
	}
}
