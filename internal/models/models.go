package models

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type GenderType string
type RoleType string
type TokenType string

const (
	GenderMale   GenderType = "MALE"
	GenderFemale GenderType = "FEMALE"

	RoleUser  RoleType = "USER"
	RoleAdmin RoleType = "ADMIN"

	AccountToken TokenType = "ACCOUNT"
	ServiceToken TokenType = "SERVICE"
)

type Account struct {
	ID             uuid.UUID   `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	Username       string      `gorm:"type:text;not null;uniqueIndex" json:"username"`
	Email          string      `gorm:"type:text;not null;uniqueIndex" json:"email"`
	HashedPassword string      `gorm:"type:text;not null" json:"-"`
	FirstName      string      `gorm:"type:text;not null" json:"first_name"`
	LastName       string      `gorm:"type:text;not null" json:"last_name"`
	DateOfBirth    *time.Time  `gorm:"type:timestamp" json:"date_of_birth,omitempty"`
	Gender         *GenderType `gorm:"type:text" json:"gender,omitempty"`
	RoleType       RoleType    `gorm:"type:text;default:'USER'" json:"role_type"`
	IsVerified     bool        `gorm:"default:false" json:"is_verified"`
	Avatar         *string     `gorm:"type:text" json:"avatar,omitempty"`
	CreatedAt      time.Time   `gorm:"type:timestamp;default:current_timestamp" json:"created_at"`
	UpdatedAt      time.Time   `gorm:"type:timestamp;default:current_timestamp" json:"updated_at"`
}

type Service struct {
	ID                 uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	OwnerID            uuid.UUID `gorm:"type:uuid" json:"owner_id"`
	ServiceName        string    `gorm:"type:text;not null" json:"service_name"`
	ServiceLogo        *string   `gorm:"type:text" json:"service_logo,omitempty"`
	ServiceDescription *string   `gorm:"type:text" json:"service_description,omitempty"`
	CreatedAt          time.Time `gorm:"type:timestamp;default:current_timestamp" json:"created_at"`
	UpdatedAt          time.Time `gorm:"type:timestamp;default:current_timestamp" json:"updated_at"`

	// Add relationships
	Owner Account        `gorm:"foreignKey:OwnerID"`
	Users []ServicesUser `gorm:"foreignKey:ServiceID"`
}

type ServicesUser struct {
	ID         uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	ServiceID  uuid.UUID `gorm:"type:uuid" json:"service_id"`
	UserID     uuid.UUID `gorm:"type:uuid" json:"user_id"`
	IsVerified bool      `gorm:"default:false" json:"is_verified"`
	CreatedAt  time.Time `gorm:"type:timestamp;default:current_timestamp" json:"created_at"`
	UpdatedAt  time.Time `gorm:"type:timestamp;default:current_timestamp" json:"updated_at"`

	// Add relationships
	User    Account `gorm:"foreignKey:UserID"`
	Service Service `gorm:"foreignKey:ServiceID"`
}

type RefreshToken struct {
	ID           uuid.UUID  `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	UserID       uuid.UUID  `gorm:"type:uuid;not null" json:"user_id"`
	ServiceID    *uuid.UUID `gorm:"type:uuid" json:"service_id,omitempty"` // Made optional
	TokenType    TokenType  `gorm:"type:text;not null" json:"token_type"`
	RefreshToken string     `gorm:"type:text;not null" json:"refresh_token"`
	ExpiresAt    time.Time  `gorm:"type:timestamp;not null" json:"expires_at"`
	CreatedAt    time.Time  `gorm:"type:timestamp;default:current_timestamp" json:"created_at"`
	UpdatedAt    time.Time  `gorm:"type:timestamp;default:current_timestamp" json:"updated_at"`
}

type AuthorizationToken struct {
	UserID    string    `json:"user_id"`
	RoleType  RoleType  `json:"role_type"`
	ServiceID *string   `json:"service_id"`
	TokenType TokenType `json:"token_type"`
	ExpiresAt int64     `json:"expires_at"`
}

// Implement jwt.Claims interface
func (t *AuthorizationToken) GetExpirationTime() (*jwt.NumericDate, error) {
	if t.ExpiresAt == 0 {
		return nil, nil
	}
	return jwt.NewNumericDate(time.Unix(t.ExpiresAt, 0)), nil
}

func (t *AuthorizationToken) GetIssuedAt() (*jwt.NumericDate, error) {
	return nil, nil
}

func (t *AuthorizationToken) GetNotBefore() (*jwt.NumericDate, error) {
	return nil, nil
}

func (t *AuthorizationToken) GetIssuer() (string, error) {
	return "", nil
}

func (t *AuthorizationToken) GetSubject() (string, error) {
	return "", nil
}

func (t *AuthorizationToken) GetAudience() (jwt.ClaimStrings, error) {
	return nil, nil
}

// Validate token claims
func (t *AuthorizationToken) Valid() error {
	if t.UserID == "" || t.TokenType == "" || t.ExpiresAt == 0 {
		return jwt.ErrTokenInvalidClaims
	}
	if t.ExpiresAt < time.Now().Unix() {
		return jwt.ErrTokenExpired
	}
	return nil
}
