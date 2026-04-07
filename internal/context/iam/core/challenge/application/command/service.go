package command

import (
	"context"
	"time"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/domain/challenge"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/domain/user"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/infrastructure/metrics"
	infraRedis "github.com/BakhodiribnYashinibnMansur/XBank/internal/infrastructure/redis"
	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/apperror"
	"golang.org/x/crypto/bcrypt"
)

const maxPendingChallenges = 5

type Service struct {
	challengeRepo challenge.Repository
	userRepo      user.Repository
	cache         *infraRedis.ChallengeCache // nil = DB only
}

func NewService(
	challengeRepo challenge.Repository,
	userRepo user.Repository,
	cache *infraRedis.ChallengeCache,
) *Service {
	return &Service{
		challengeRepo: challengeRepo,
		userRepo:      userRepo,
		cache:         cache,
	}
}

// Request - create a new challenge for a user.
// Returns the challenge ID so the client can verify it.
func (s *Service) Request(ctx context.Context, userID string, method challenge.Method, action, metadata string) (_ *challenge.Challenge, err error) {
	defer metrics.ObserveService("ChallengeService", "Request", time.Now(), &err)

	// Rate limit: max 5 pending challenges per user
	count, err := s.challengeRepo.CountPendingByUser(ctx, userID)
	if err != nil {
		return nil, apperror.ErrInternal.Wrap(err)
	}
	if count >= maxPendingChallenges {
		return nil, apperror.ErrTooManyRequest.WithMessage("Too many pending challenges")
	}

	ch, err := challenge.NewChallenge(userID, method, action, metadata)
	if err != nil {
		return nil, apperror.ErrBadRequest.Wrap(err)
	}

	if err := s.challengeRepo.Create(ctx, ch); err != nil {
		return nil, apperror.ErrInternal.Wrap(err)
	}

	return ch, nil
}

// Verify - verify a challenge with the user's password.
// Returns the challenge with token set on success.
func (s *Service) Verify(ctx context.Context, challengeID, userID, password string) (_ *challenge.Challenge, err error) {
	defer metrics.ObserveService("ChallengeService", "Verify", time.Now(), &err)

	ch, err := s.challengeRepo.GetByID(ctx, challengeID)
	if err != nil {
		return nil, apperror.ErrChallengeNotFound
	}

	// Ensure the challenge belongs to this user
	if ch.UserID != userID {
		return nil, apperror.ErrForbidden
	}

	// Check if already expired/used/failed
	if ch.Status != challenge.StatusPending {
		return nil, apperror.ErrChallengeAlreadyUsed
	}
	if time.Now().After(ch.ExpiresAt) {
		ch.Status = challenge.StatusExpired
		s.challengeRepo.Update(ctx, ch)
		return nil, apperror.ErrChallengeExpired
	}

	// Verify password
	u, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, apperror.ErrUserNotFound
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password)); err != nil {
		ch.Fail()
		s.challengeRepo.Update(ctx, ch)
		return nil, apperror.ErrInvalidCredentials
	}

	// Mark as verified and generate token
	if err := ch.Verify(); err != nil {
		return nil, apperror.ErrChallengeExpired
	}

	if err := s.challengeRepo.Update(ctx, ch); err != nil {
		return nil, apperror.ErrInternal.Wrap(err)
	}

	// Cache token in Redis for fast middleware lookup
	if s.cache != nil {
		s.cache.SetToken(ctx, ch.Token, ch.UserID, challenge.TokenTTL)
	}

	return ch, nil
}

// ValidateToken - check if a challenge token is valid for the given user.
// Used by RequireChallenge middleware.
func (s *Service) ValidateToken(ctx context.Context, token, userID string) (err error) {
	defer metrics.ObserveService("ChallengeService", "ValidateToken", time.Now(), &err)

	// Fast path: check Redis cache
	if s.cache != nil {
		cachedUserID, err := s.cache.GetUserByToken(ctx, token)
		if err == nil && cachedUserID == userID {
			return nil
		}
	}

	// Slow path: check DB
	ch, err := s.challengeRepo.GetByToken(ctx, token)
	if err != nil {
		return apperror.ErrChallengeTokenInvalid
	}

	if !ch.IsTokenValid(token) || ch.UserID != userID {
		return apperror.ErrChallengeTokenInvalid
	}

	return nil
}
