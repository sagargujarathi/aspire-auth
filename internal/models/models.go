package models

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type GenderType string
type RoleType string

const (
	GenderMale   GenderType = "MALE"
	GenderFemale GenderType = "FEMALE"

	RoleUser  RoleType = "USER"
	RoleAdmin RoleType = "ADMIN"
)

type Account struct {
	ID             uuid.UUID   `gorm:"type:uuid;default:gen_random_uuid()" json:"id"`
	Username       string      `gorm:"type:text;not null" json:"username"`
	Email          string      `gorm:"type:text;not null" json:"email"`
	HashedPassword string      `gorm:"type:text;not null" json:"-"`
	FirstName      string      `gorm:"type:text;not null" json:"first_name"`
	LastName       string      `gorm:"type:text;not null" json:"last_name"`
	DateOfBirth    *time.Time  `gorm:"type:date" json:"date_of_birth,omitempty"`
	Gender         *GenderType `gorm:"type:gender_type" json:"gender,omitempty"`
	RoleType       *RoleType   `gorm:"type:role_type;default:'USER'" json:"role_type,omitempty"`
	IsVerified     bool        `gorm:"default:false" json:"is_verified,omitempty"`
	Avatar         *string     `gorm:"type:text" json:"avatar,omitempty"`
	CreatedAt      time.Time   `gorm:"type:timestamp;default:current_timestamp" json:"created_at"`
	UpdatedAt      time.Time   `gorm:"type:timestamp;default:current_timestamp" json:"updated_at"`
}

type Service struct {
	ID                 uuid.UUID `json:"id" db:"id"`
	OwnerID            uuid.UUID `json:"owner_id" db:"owner_id"`
	ServiceName        string    `json:"service_name" db:"service_name"`
	ServiceLogo        *string   `json:"service_logo,omitempty" db:"service_logo"`
	ServiceDescription *string   `json:"service_description,omitempty" db:"service_description"`
	CreatedAt          time.Time `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time `json:"updated_at" db:"updated_at"`
}

type ServiceUser struct {
	ID         uuid.UUID `json:"id" db:"id"`
	ServiceID  uuid.UUID `json:"service_id" db:"service_id"`
	UserID     uuid.UUID `json:"user_id" db:"user_id"`
	IsVerified bool      `json:"is_verified" db:"is_verified"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
}

type RefreshToken struct {
	ID           uuid.UUID `json:"id" db:"id"`
	UserID       uuid.UUID `json:"user_id" db:"user_id"`
	ServiceID    uuid.UUID `json:"service_id" db:"service_id"`
	RefreshToken string    `json:"refresh_token" db:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

type AuthorizationToken struct {
	ID       string   `json:"id"`
	RoleType RoleType `json:"role_type"`
	Exp      int64    `json:"exp"`
}

// Implement jwt.Claims interface
func (t *AuthorizationToken) GetExpirationTime() (*jwt.NumericDate, error) {
	if t.Exp == 0 {
		return nil, nil
	}
	return jwt.NewNumericDate(time.Unix(t.Exp, 0)), nil
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
	if t.ID == "" {
		return jwt.ErrTokenInvalidClaims
	}
	if t.Exp < time.Now().Unix() {
		return jwt.ErrTokenExpired
	}
	return nil
}
