package command

import (
	"context"
	"time"

	domainCard "github.com/BakhodiribnYashinibnMansur/XBank/internal/domain/card"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/infrastructure/crypto"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/infrastructure/metrics"
	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/apperror"
)

const tokenTTL = 365 * 24 * time.Hour // 1 year

// TokenService - manages card tokenization
type TokenService struct {
	tokenRepo domainCard.TokenRepository
	cardRepo  domainCard.Repository
	tokenizer *crypto.Tokenizer
}

func NewTokenService(tokenRepo domainCard.TokenRepository, cardRepo domainCard.Repository, tokenizer *crypto.Tokenizer) *TokenService {
	return &TokenService{tokenRepo: tokenRepo, cardRepo: cardRepo, tokenizer: tokenizer}
}

// Tokenize - generate an opaque token for a card's PAN
func (s *TokenService) Tokenize(ctx context.Context, cardID string) (_ *domainCard.Token, err error) {
	defer metrics.ObserveService("TokenService", "Tokenize", time.Now(), &err)

	c, err := s.cardRepo.GetByID(ctx, cardID)
	if err != nil {
		return nil, domainCard.ErrCardNotFound
	}

	if s.tokenizer == nil {
		return nil, apperror.ErrInternal.WithMessage("Tokenization not configured")
	}

	token, encryptedPAN, lastFour, err := s.tokenizer.Tokenize(c.PAN)
	if err != nil {
		return nil, apperror.ErrInternal.Wrap(err)
	}

	now := time.Now()
	t := &domainCard.Token{
		Token:        token,
		CardID:       cardID,
		PANEncrypted: encryptedPAN,
		LastFour:     lastFour,
		ExpiresAt:    now.Add(tokenTTL),
		CreatedAt:    now,
		IsActive:     true,
	}

	if err := s.tokenRepo.Create(ctx, t); err != nil {
		return nil, apperror.ErrInternal.Wrap(err)
	}

	return t, nil
}

// Detokenize - retrieve the real PAN from a token (for authorized systems only)
func (s *TokenService) Detokenize(ctx context.Context, tokenStr string) (_ string, err error) {
	defer metrics.ObserveService("TokenService", "Detokenize", time.Now(), &err)

	t, err := s.tokenRepo.GetByToken(ctx, tokenStr)
	if err != nil || !t.IsActive {
		return "", apperror.ErrNotFound.WithMessage("Token not found or inactive")
	}
	if time.Now().After(t.ExpiresAt) {
		return "", apperror.ErrNotFound.WithMessage("Token has expired")
	}

	pan, err := s.tokenizer.Detokenize(t.PANEncrypted)
	if err != nil {
		return "", apperror.ErrInternal.Wrap(err)
	}

	return pan, nil
}

// ListTokens - list active tokens for a card
func (s *TokenService) ListTokens(ctx context.Context, cardID string) (_ []*domainCard.Token, err error) {
	defer metrics.ObserveService("TokenService", "ListTokens", time.Now(), &err)
	return s.tokenRepo.ListByCardID(ctx, cardID)
}

// RevokeToken - deactivate a token
func (s *TokenService) RevokeToken(ctx context.Context, tokenStr string) (err error) {
	defer metrics.ObserveService("TokenService", "RevokeToken", time.Now(), &err)
	return s.tokenRepo.Deactivate(ctx, tokenStr)
}
