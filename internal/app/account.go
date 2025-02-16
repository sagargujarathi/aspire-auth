package app

import (
	"aspire-auth/internal/domain/account"
	"aspire-auth/internal/models"
	"aspire-auth/internal/request"
	"aspire-auth/internal/utils"
	"aspire-auth/internal/validation"
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

type AccountUseCase interface {
	Create(ctx context.Context, req *request.CreateAccountRequest) (*models.Account, string, error)
	Update(ctx context.Context, id string, req *request.UpdateAccountRequest) error
	Delete(ctx context.Context, id string) error
	Verify(ctx context.Context, accountID, otp string) error
}

type accountService struct {
	repo      account.Repository
	validator validation.Validator
	redis     *redis.Client
}

func NewAccountUseCase(repo account.Repository, validator validation.Validator, redis *redis.Client) AccountUseCase {
	return &accountService{
		repo:      repo,
		validator: validator,
		redis:     redis,
	}
}

func (s *accountService) Create(ctx context.Context, req *request.CreateAccountRequest) (*models.Account, string, error) {
	if err := s.validator.ValidateCreateAccount(req); err != nil {
		return nil, "", err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", err
	}

	account := &models.Account{
		Username:       req.Username,
		Email:          req.Email,
		HashedPassword: string(hashedPassword),
		FirstName:      req.FirstName,
		LastName:       req.LastName,
		RoleType:       models.RoleUser,
	}

	if err := s.repo.Create(account); err != nil {
		return nil, "", err
	}

	// Generate OTP
	otp := fmt.Sprintf("%06d", rand.Intn(999999))
	redisKey := fmt.Sprintf("otp:%s", account.ID.String())

	// Store OTP in Redis
	err = s.redis.Set(ctx, redisKey, otp, 15*time.Minute).Err()
	if err != nil {
		return account, "", err
	}

	return account, otp, nil
}

func (s *accountService) Verify(ctx context.Context, accountID, otp string) error {
	redisKey := fmt.Sprintf("otp:%s", accountID)

	storedOTP, err := s.redis.Get(ctx, redisKey).Result()
	if err != nil {
		return utils.ErrInvalidToken
	}

	if storedOTP != otp {
		return utils.ErrInvalidToken
	}

	if err := s.repo.Verify(accountID); err != nil {
		return err
	}

	s.redis.Del(ctx, redisKey)
	return nil
}

func (s *accountService) Update(ctx context.Context, id string, req *request.UpdateAccountRequest) error {
	if err := s.validator.ValidateUpdateAccount(req); err != nil {
		return err
	}

	updates := make(map[string]interface{})
	if req.Username != "" {
		updates["username"] = req.Username
	}
	if req.FirstName != "" {
		updates["first_name"] = req.FirstName
	}
	if req.LastName != "" {
		updates["last_name"] = req.LastName
	}
	if !req.DateOfBirth.IsZero() {
		updates["date_of_birth"] = req.DateOfBirth
	}

	return s.repo.Update(id, updates)
}

func (s *accountService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(id)
}
