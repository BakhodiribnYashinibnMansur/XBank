package apperror

// ErrorSeverity indicates how critical an error is for monitoring and alerting.
type ErrorSeverity string

const (
	SeverityCritical ErrorSeverity = "CRITICAL"
	SeverityHigh     ErrorSeverity = "HIGH"
	SeverityMedium   ErrorSeverity = "MEDIUM"
	SeverityLow      ErrorSeverity = "LOW"
	SeverityInfo     ErrorSeverity = "INFO"
)

// ErrorCategory classifies an error by its domain for filtering and dashboards.
type ErrorCategory string

const (
	CategoryValidation ErrorCategory = "VALIDATION"
	CategorySecurity   ErrorCategory = "SECURITY"
	CategoryData       ErrorCategory = "DATA"
	CategoryBusiness   ErrorCategory = "BUSINESS"
	CategorySystem     ErrorCategory = "SYSTEM"
	CategoryExternal   ErrorCategory = "EXTERNAL"
)
