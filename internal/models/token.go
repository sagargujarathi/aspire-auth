package models

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Base claims implementation to avoid code duplication
type baseClaims struct {
	ExpiresAt int64 `json:"exp,omitempty"`
	IssuedAt  int64 `json:"iat,omitempty"`
}

// Base claims methods
func (b *baseClaims) GetExpirationTime() (*jwt.NumericDate, error) {
	if b.ExpiresAt == 0 {
		return nil, nil
	}
	return jwt.NewNumericDate(time.Unix(b.ExpiresAt, 0)), nil
}

func (b *baseClaims) GetIssuedAt() (*jwt.NumericDate, error) {
	if b.IssuedAt == 0 {
		return nil, nil
	}
	return jwt.NewNumericDate(time.Unix(b.IssuedAt, 0)), nil
}

func (b *baseClaims) GetNotBefore() (*jwt.NumericDate, error) {
	return b.GetIssuedAt()
}

func (b *baseClaims) GetIssuer() (string, error) {
	return "aspire-auth", nil
}

func (b *baseClaims) GetSubject() (string, error) {
	return "", nil
}

func (b *baseClaims) GetAudience() (jwt.ClaimStrings, error) {
	return []string{"aspire-services"}, nil
}

func (b *baseClaims) Valid() error {
	// Check if token is expired
	if b.ExpiresAt > 0 && time.Unix(b.ExpiresAt, 0).Before(time.Now()) {
		return jwt.ErrTokenExpired
	}

	// Check if token has a valid issued time
	if b.IssuedAt > 0 {
		// Token cannot be used before it's issued
		if time.Unix(b.IssuedAt, 0).After(time.Now()) {
			return fmt.Errorf("token used before issued time")
		}

		// Token should not be from too far in the past (prevents replay attacks)
		// 30 days is a reasonable maximum token age
		if time.Since(time.Unix(b.IssuedAt, 0)) > 30*24*time.Hour {
			return fmt.Errorf("token is too old")
		}
	}

	return nil
}
