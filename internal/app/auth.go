package app

import (
	"aspire-auth/internal/config"
	"aspire-auth/internal/domain/account"
	"aspire-auth/internal/models"
	"aspire-auth/internal/utils"
	"context"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthUseCase struct {
	accountRepo account.Repository
	config      *config.Config
}

func NewAuthUseCase(repo account.Repository, cfg *config.Config) *AuthUseCase {
	return &AuthUseCase{
		accountRepo: repo,
		config:      cfg,
	}
}

func (uc *AuthUseCase) Login(ctx context.Context, email, password string) (*models.Account, string, string, error) {
	account, err := uc.accountRepo.FindByEmail(email)
	if err != nil {
		return nil, "", "", utils.ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(account.HashedPassword), []byte(password)); err != nil {
		return nil, "", "", utils.ErrInvalidCredentials
	}

	if !account.IsVerified {
		return nil, "", "", utils.ErrAccountNotVerified
	}

	accessToken, err := uc.generateAccessToken(account)
	if err != nil {
		return nil, "", "", err
	}

	refreshToken, err := uc.generateRefreshToken(account)
	if err != nil {
		return nil, "", "", err
	}

	return account, accessToken, refreshToken, nil
}

func (uc *AuthUseCase) RefreshToken(ctx context.Context, refreshToken string) (string, string, error) {
	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(refreshToken, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(uc.config.JWT.RefreshTokenSecret), nil
	})

	if err != nil || !token.Valid {
		return "", "", utils.ErrInvalidToken
	}

	account, err := uc.accountRepo.FindByID(claims["id"].(string))
	if err != nil {
		return "", "", utils.ErrInvalidToken
	}

	newAccess, err := uc.generateAccessToken(account)
	if err != nil {
		return "", "", err
	}

	newRefresh, err := uc.generateRefreshToken(account)
	if err != nil {
		return "", "", err
	}

	return newAccess, newRefresh, nil
}

func (uc *AuthUseCase) generateAccessToken(account *models.Account) (string, error) {
	claims := jwt.MapClaims{
		"id":        account.ID.String(),
		"role_type": account.RoleType,
		"exp":       time.Now().Add(uc.config.JWT.AccessExpiry).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(uc.config.JWT.AccessTokenSecret))
}

func (uc *AuthUseCase) generateRefreshToken(account *models.Account) (string, error) {
	claims := jwt.MapClaims{
		"id":        account.ID.String(),
		"role_type": account.RoleType,
		"exp":       time.Now().Add(uc.config.JWT.RefreshExpiry).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(uc.config.JWT.RefreshTokenSecret))
}
