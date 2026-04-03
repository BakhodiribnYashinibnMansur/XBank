package challenge

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/domain/challenge"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/domain/user"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/infrastructure/mock"
	"golang.org/x/crypto/bcrypt"
)

func setupTest(t *testing.T) (*Service, *mock.UserRepository, string) {
	t.Helper()
	userRepo := mock.NewUserRepository()
	challengeRepo := newMemoryChallengeRepo()

	hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.MinCost)
	u, _ := user.NewUser("test@test.com", string(hash), "Test", "User")
	userRepo.Create(context.Background(), u)

	svc := NewService(challengeRepo, userRepo, nil)
	return svc, userRepo, u.ID
}

func TestRequest_Success(t *testing.T) {
	svc, _, userID := setupTest(t)

	ch, err := svc.Request(context.Background(), userID, challenge.MethodPassword, "transfer", "")
	if err != nil {
		t.Fatal(err)
	}
	if ch.Status != challenge.StatusPending {
		t.Errorf("expected PENDING, got %s", ch.Status)
	}
}

func TestRequest_TooMany(t *testing.T) {
	svc, _, userID := setupTest(t)

	for i := 0; i < maxPendingChallenges; i++ {
		svc.Request(context.Background(), userID, challenge.MethodPassword, "transfer", "")
	}

	_, err := svc.Request(context.Background(), userID, challenge.MethodPassword, "transfer", "")
	if err == nil {
		t.Error("should reject when max pending challenges reached")
	}
}

func TestVerify_Success(t *testing.T) {
	svc, _, userID := setupTest(t)

	ch, _ := svc.Request(context.Background(), userID, challenge.MethodPassword, "transfer", "")

	verified, err := svc.Verify(context.Background(), ch.ID, userID, "password123")
	if err != nil {
		t.Fatal(err)
	}
	if verified.Token == "" {
		t.Error("token should be generated on verify")
	}
}

func TestVerify_WrongPassword(t *testing.T) {
	svc, _, userID := setupTest(t)

	ch, _ := svc.Request(context.Background(), userID, challenge.MethodPassword, "transfer", "")

	_, err := svc.Verify(context.Background(), ch.ID, userID, "wrong-password")
	if err == nil {
		t.Error("should reject wrong password")
	}
}

func TestVerify_WrongUser(t *testing.T) {
	svc, _, userID := setupTest(t)

	ch, _ := svc.Request(context.Background(), userID, challenge.MethodPassword, "transfer", "")

	_, err := svc.Verify(context.Background(), ch.ID, "other-user", "password123")
	if err == nil {
		t.Error("should reject wrong user")
	}
}

func TestValidateToken_Success(t *testing.T) {
	svc, _, userID := setupTest(t)

	ch, _ := svc.Request(context.Background(), userID, challenge.MethodPassword, "transfer", "")
	verified, _ := svc.Verify(context.Background(), ch.ID, userID, "password123")

	err := svc.ValidateToken(context.Background(), verified.Token, userID)
	if err != nil {
		t.Fatalf("valid token should pass: %v", err)
	}
}

func TestValidateToken_Invalid(t *testing.T) {
	svc, _, userID := setupTest(t)

	err := svc.ValidateToken(context.Background(), "invalid-token", userID)
	if err == nil {
		t.Error("invalid token should fail")
	}
}

// --- In-memory challenge repo for testing ---

type memoryChallengeRepo struct {
	challenges map[string]*challenge.Challenge
}

func newMemoryChallengeRepo() *memoryChallengeRepo {
	return &memoryChallengeRepo{challenges: make(map[string]*challenge.Challenge)}
}

func (r *memoryChallengeRepo) Create(_ context.Context, c *challenge.Challenge) error {
	r.challenges[c.ID] = c
	return nil
}

func (r *memoryChallengeRepo) GetByID(_ context.Context, id string) (*challenge.Challenge, error) {
	c, ok := r.challenges[id]
	if !ok {
		return nil, fmt.Errorf("challenge not found")
	}
	return c, nil
}

func (r *memoryChallengeRepo) GetByToken(_ context.Context, token string) (*challenge.Challenge, error) {
	for _, c := range r.challenges {
		if c.Token == token {
			return c, nil
		}
	}
	return nil, fmt.Errorf("not found")
}

func (r *memoryChallengeRepo) Update(_ context.Context, c *challenge.Challenge) error {
	r.challenges[c.ID] = c
	return nil
}

func (r *memoryChallengeRepo) CountPendingByUser(_ context.Context, userID string) (int, error) {
	count := 0
	for _, c := range r.challenges {
		if c.UserID == userID && c.Status == challenge.StatusPending && time.Now().Before(c.ExpiresAt) {
			count++
		}
	}
	return count, nil
}
