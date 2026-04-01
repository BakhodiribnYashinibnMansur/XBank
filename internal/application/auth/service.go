package auth

import (
	"context"
	"time"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/domain/session"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/domain/user"
	infraAuth "github.com/BakhodiribnYashinibnMansur/XBank/internal/infrastructure/auth"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/infrastructure/metrics"
	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/apperror"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = apperror.ErrInvalidCredentials
)

// Service - authentication use cases
type Service struct {
	userRepo    user.Repository
	sessionRepo session.Repository
	jwtService  *infraAuth.JWTService
}

func NewService(userRepo user.Repository, sessionRepo session.Repository, jwtService *infraAuth.JWTService) *Service {
	return &Service{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		jwtService:  jwtService,
	}
}

// LoginResult - result of a login operation
type LoginResult struct {
	AccessToken  string
	RefreshToken string
	User         *user.User
}

// Login - user sign-in
func (s *Service) Login(ctx context.Context, email, password, userAgent, ipAddress string) (*LoginResult, error) {
	// 1. Find the user
	u, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		metrics.LoginsTotal.WithLabelValues("failed").Inc()
		return nil, ErrInvalidCredentials // Security: don't reveal "user not found"
	}

	// 2. Verify the password
	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password)); err != nil {
		metrics.LoginsTotal.WithLabelValues("failed").Inc()
		return nil, ErrInvalidCredentials // Security: don't reveal "wrong password"
	}

	// 3. Generate token pair
	tokenPair, err := s.jwtService.GenerateTokenPair(u.ID, u.Email, string(u.Role), ipAddress)
	if err != nil {
		return nil, err
	}

	// 4. Create session (refresh token is stored as a HASH)
	refreshTokenHash := infraAuth.HashToken(tokenPair.RefreshToken)
	expiresAt := time.Now().Add(s.jwtService.RefreshTokenTTL())

	sess, err := session.NewSession(u.ID, refreshTokenHash, userAgent, ipAddress, expiresAt)
	if err != nil {
		return nil, err
	}

	if err := s.sessionRepo.Create(ctx, sess); err != nil {
		return nil, err
	}

	metrics.LoginsTotal.WithLabelValues("success").Inc()
	return &LoginResult{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		User:         u,
	}, nil
}

// Refresh - obtain a new token pair (via refresh token)
func (s *Service) Refresh(ctx context.Context, refreshToken, userAgent, ipAddress string) (*LoginResult, error) {
	// 1. Find the session by refresh token hash
	refreshTokenHash := infraAuth.HashToken(refreshToken)
	sess, err := s.sessionRepo.GetByRefreshToken(ctx, refreshTokenHash)
	if err != nil {
		return nil, session.ErrSessionNotFound
	}

	// 2. Check session expiration
	if sess.IsExpired() {
		s.sessionRepo.DeleteByID(ctx, sess.ID)
		return nil, session.ErrSessionExpired
	}

	// 3. Delete the old session (rotation - new token each time)
	s.sessionRepo.DeleteByID(ctx, sess.ID)

	// 4. Get user data
	u, err := s.userRepo.GetByID(ctx, sess.UserID)
	if err != nil {
		return nil, err
	}

	// 5. Generate new token pair
	tokenPair, err := s.jwtService.GenerateTokenPair(u.ID, u.Email, string(u.Role), ipAddress)
	if err != nil {
		return nil, err
	}

	// 6. Create new session
	newHash := infraAuth.HashToken(tokenPair.RefreshToken)
	expiresAt := time.Now().Add(s.jwtService.RefreshTokenTTL())

	newSess, err := session.NewSession(u.ID, newHash, userAgent, ipAddress, expiresAt)
	if err != nil {
		return nil, err
	}

	if err := s.sessionRepo.Create(ctx, newSess); err != nil {
		return nil, err
	}

	return &LoginResult{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		User:         u,
	}, nil
}

// Logout - end a single session
func (s *Service) Logout(ctx context.Context, refreshToken string) error {
	refreshTokenHash := infraAuth.HashToken(refreshToken)
	sess, err := s.sessionRepo.GetByRefreshToken(ctx, refreshTokenHash)
	if err != nil {
		return nil // Don't return an error even if session is not found
	}
	return s.sessionRepo.DeleteByID(ctx, sess.ID)
}

// LogoutAll - end all sessions
func (s *Service) LogoutAll(ctx context.Context, userID string) error {
	return s.sessionRepo.DeleteAllByUserID(ctx, userID)
}
