package command

import (
	"context"
	"encoding/json"
	"time"

	session "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/iam/generic/session/domain"
	user "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/iam/generic/user/domain"
	infraAuth "github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/security/jwt"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/metrics"
	infraRedis "github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/db/redis"
	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/apperror"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = apperror.ErrInvalidCredentials
	ErrAccountLocked      = apperror.Forbidden(3010, "Account locked due to too many failed attempts")
)

type Service struct {
	userRepo     user.Repository
	sessionRepo  session.Repository    // DB (audit)
	jwtService   *infraAuth.JWTService
	totpService  *infraAuth.TOTPService   // nil = TOTP disabled globally
	sessionCache *infraRedis.SessionCache  // Redis (active check) — nil = use DB only
	loginLimiter *infraRedis.LoginLimiter  // brute-force — nil = disabled
}

func NewService(
	userRepo user.Repository,
	sessionRepo session.Repository,
	jwtService *infraAuth.JWTService,
	totpService *infraAuth.TOTPService,
	sessionCache *infraRedis.SessionCache,
	loginLimiter *infraRedis.LoginLimiter,
) *Service {
	return &Service{
		userRepo:     userRepo,
		sessionRepo:  sessionRepo,
		jwtService:   jwtService,
		totpService:  totpService,
		sessionCache: sessionCache,
		loginLimiter: loginLimiter,
	}
}

type LoginResult struct {
	AccessToken  string
	RefreshToken string
	User         *user.User
	TOTPRequired bool   // true = client must call /auth/totp/verify
}

// Login - user sign-in
func (s *Service) Login(ctx context.Context, email, password, userAgent, ipAddress string) (_ *LoginResult, err error) {
	defer metrics.ObserveService("AuthService", "Login", time.Now(), &err)

	// 0. Brute-force check
	if s.loginLimiter != nil {
		locked, _ := s.loginLimiter.IsLocked(ctx, email)
		if locked {
			metrics.LoginsTotal.WithLabelValues("locked").Inc()
			return nil, ErrAccountLocked
		}
	}

	// 1. Find user
	u, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		s.recordLoginFailure(ctx, email)
		return nil, ErrInvalidCredentials
	}

	// 2. Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password)); err != nil {
		s.recordLoginFailure(ctx, email)
		return nil, ErrInvalidCredentials
	}

	// 2.5. If TOTP is enabled, don't issue tokens yet — require TOTP verification
	if u.TOTPEnabled {
		if s.loginLimiter != nil {
			s.loginLimiter.ResetAttempts(ctx, email)
		}
		metrics.LoginsTotal.WithLabelValues("totp_pending").Inc()
		return &LoginResult{
			User:         u,
			TOTPRequired: true,
		}, nil
	}

	// 3. Generate tokens
	tokenPair, err := s.jwtService.GenerateTokenPair(u.ID, u.Email, string(u.Role), ipAddress)
	if err != nil {
		return nil, err
	}

	// 4. Save session to DB (audit)
	refreshTokenHash := infraAuth.HashToken(tokenPair.RefreshToken)
	expiresAt := time.Now().Add(s.jwtService.RefreshTokenTTL())

	sess, err := session.NewSession(u.ID, refreshTokenHash, userAgent, ipAddress, expiresAt)
	if err != nil {
		return nil, err
	}
	if err := s.sessionRepo.Create(ctx, sess); err != nil {
		return nil, err
	}

	// 5. Cache session in Redis (active check)
	s.cacheSession(ctx, refreshTokenHash, sess, s.jwtService.RefreshTokenTTL())

	// 6. Reset login attempts on success
	if s.loginLimiter != nil {
		s.loginLimiter.ResetAttempts(ctx, email)
	}

	metrics.LoginsTotal.WithLabelValues("success").Inc()
	return &LoginResult{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		User:         u,
	}, nil
}

// LoginWithTOTP - complete login after TOTP verification.
// Called when Login returns TOTPRequired=true.
func (s *Service) LoginWithTOTP(ctx context.Context, email, code, userAgent, ipAddress string) (_ *LoginResult, err error) {
	defer metrics.ObserveService("AuthService", "LoginWithTOTP", time.Now(), &err)

	u, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	if !u.TOTPEnabled || u.TOTPSecret == "" {
		return nil, apperror.ErrTOTPNotEnabled
	}

	// Validate TOTP code
	if s.totpService == nil || !s.totpService.Validate(code, u.TOTPSecret) {
		metrics.LoginsTotal.WithLabelValues("totp_failed").Inc()
		return nil, apperror.ErrTOTPInvalidCode
	}

	// TOTP valid — issue tokens
	tokenPair, err := s.jwtService.GenerateTokenPair(u.ID, u.Email, string(u.Role), ipAddress)
	if err != nil {
		return nil, err
	}

	refreshTokenHash := infraAuth.HashToken(tokenPair.RefreshToken)
	expiresAt := time.Now().Add(s.jwtService.RefreshTokenTTL())

	sess, err := session.NewSession(u.ID, refreshTokenHash, userAgent, ipAddress, expiresAt)
	if err != nil {
		return nil, err
	}
	if err := s.sessionRepo.Create(ctx, sess); err != nil {
		return nil, err
	}
	s.cacheSession(ctx, refreshTokenHash, sess, s.jwtService.RefreshTokenTTL())

	metrics.LoginsTotal.WithLabelValues("totp_success").Inc()
	return &LoginResult{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		User:         u,
	}, nil
}

// SetupTOTP - generate a TOTP secret for a user (before enabling)
func (s *Service) SetupTOTP(ctx context.Context, userID string) (secret, url string, err error) {
	defer metrics.ObserveService("AuthService", "SetupTOTP", time.Now(), &err)

	if s.totpService == nil {
		return "", "", apperror.ErrInternal.WithMessage("TOTP not configured")
	}

	u, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return "", "", err
	}
	if u.TOTPEnabled {
		return "", "", apperror.ErrTOTPAlreadyEnabled
	}

	secret, url, err = s.totpService.GenerateSecret(u.Email)
	if err != nil {
		return "", "", apperror.ErrInternal.Wrap(err)
	}

	// Save secret but don't enable yet — user must verify first
	if err := s.userRepo.UpdateTOTP(ctx, userID, secret, false, nil); err != nil {
		return "", "", apperror.ErrInternal.Wrap(err)
	}

	return secret, url, nil
}

// VerifyAndEnableTOTP - verify a TOTP code and enable 2FA
func (s *Service) VerifyAndEnableTOTP(ctx context.Context, userID, code string) (err error) {
	defer metrics.ObserveService("AuthService", "VerifyAndEnableTOTP", time.Now(), &err)

	u, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	if u.TOTPEnabled {
		return apperror.ErrTOTPAlreadyEnabled
	}
	if u.TOTPSecret == "" {
		return apperror.ErrBadRequest.WithMessage("Call /auth/totp/setup first")
	}

	if s.totpService == nil || !s.totpService.Validate(code, u.TOTPSecret) {
		return apperror.ErrTOTPInvalidCode
	}

	now := time.Now()
	return s.userRepo.UpdateTOTP(ctx, userID, u.TOTPSecret, true, &now)
}

// DisableTOTP - turn off 2FA (requires password confirmation)
func (s *Service) DisableTOTP(ctx context.Context, userID, password string) (err error) {
	defer metrics.ObserveService("AuthService", "DisableTOTP", time.Now(), &err)

	u, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password)); err != nil {
		return ErrInvalidCredentials
	}

	return s.userRepo.UpdateTOTP(ctx, userID, "", false, nil)
}

// Refresh - obtain new token pair
func (s *Service) Refresh(ctx context.Context, refreshToken, userAgent, ipAddress string) (_ *LoginResult, err error) {
	defer metrics.ObserveService("AuthService", "Refresh", time.Now(), &err)

	refreshTokenHash := infraAuth.HashToken(refreshToken)

	// 1. Check Redis first (fast), fallback to DB
	sess, err := s.findSession(ctx, refreshTokenHash)
	if err != nil {
		return nil, session.ErrSessionNotFound
	}

	// 2. Check expiration
	if sess.IsExpired() {
		s.deleteSession(ctx, sess.ID, refreshTokenHash)
		return nil, session.ErrSessionExpired
	}

	// 3. Delete old session (rotation)
	s.deleteSession(ctx, sess.ID, refreshTokenHash)

	// 4. Get user
	u, err := s.userRepo.GetByID(ctx, sess.UserID)
	if err != nil {
		return nil, err
	}

	// 5. New tokens
	tokenPair, err := s.jwtService.GenerateTokenPair(u.ID, u.Email, string(u.Role), ipAddress)
	if err != nil {
		return nil, err
	}

	// 6. New session
	newHash := infraAuth.HashToken(tokenPair.RefreshToken)
	expiresAt := time.Now().Add(s.jwtService.RefreshTokenTTL())

	newSess, err := session.NewSession(u.ID, newHash, userAgent, ipAddress, expiresAt)
	if err != nil {
		return nil, err
	}
	if err := s.sessionRepo.Create(ctx, newSess); err != nil {
		return nil, err
	}
	s.cacheSession(ctx, newHash, newSess, s.jwtService.RefreshTokenTTL())

	return &LoginResult{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		User:         u,
	}, nil
}

// Logout - end session + blacklist access token
func (s *Service) Logout(ctx context.Context, refreshToken string) (err error) {
	defer metrics.ObserveService("AuthService", "Logout", time.Now(), &err)

	refreshTokenHash := infraAuth.HashToken(refreshToken)
	sess, err := s.sessionRepo.GetByRefreshToken(ctx, refreshTokenHash)
	if err != nil {
		return nil
	}
	s.deleteSession(ctx, sess.ID, refreshTokenHash)
	return nil
}

// LogoutAll - end all sessions
func (s *Service) LogoutAll(ctx context.Context, userID string) (err error) {
	defer metrics.ObserveService("AuthService", "LogoutAll", time.Now(), &err)
	return s.sessionRepo.DeleteAllByUserID(ctx, userID)
}

// BlacklistToken - add access token to blacklist (called on logout)
func (s *Service) BlacklistToken(ctx context.Context, tokenJTI string, remainingTTL time.Duration) {
	if s.sessionCache != nil {
		s.sessionCache.BlacklistToken(ctx, tokenJTI, remainingTTL)
	}
}

// IsTokenBlacklisted - check if access token was revoked
func (s *Service) IsTokenBlacklisted(ctx context.Context, tokenJTI string) bool {
	if s.sessionCache == nil {
		return false
	}
	blacklisted, _ := s.sessionCache.IsBlacklisted(ctx, tokenJTI)
	return blacklisted
}

// --- Internal helpers ---

func (s *Service) recordLoginFailure(ctx context.Context, email string) {
	metrics.LoginsTotal.WithLabelValues("failed").Inc()
	if s.loginLimiter != nil {
		s.loginLimiter.RecordFailure(ctx, email)
	}
}

// cacheSession - save session to Redis
func (s *Service) cacheSession(ctx context.Context, tokenHash string, sess *session.Session, ttl time.Duration) {
	if s.sessionCache == nil {
		return
	}
	data, err := json.Marshal(sess)
	if err != nil {
		return
	}
	s.sessionCache.SetSession(ctx, tokenHash, string(data), ttl)
}

// findSession - check Redis first, fallback to DB
func (s *Service) findSession(ctx context.Context, refreshTokenHash string) (*session.Session, error) {
	// Try Redis
	if s.sessionCache != nil {
		cached, err := s.sessionCache.GetSession(ctx, refreshTokenHash)
		if err == nil && cached != "" {
			var sess session.Session
			if json.Unmarshal([]byte(cached), &sess) == nil {
				return &sess, nil
			}
		}
	}

	// Fallback to DB
	return s.sessionRepo.GetByRefreshToken(ctx, refreshTokenHash)
}

// deleteSession - remove from both Redis and DB
func (s *Service) deleteSession(ctx context.Context, sessionID, tokenHash string) {
	if s.sessionCache != nil {
		s.sessionCache.DeleteSession(ctx, tokenHash)
	}
	s.sessionRepo.DeleteByID(ctx, sessionID)
}
