package helpers

import (
	"aspire-auth/internal/config"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTHelpers struct {
	*config.Config
}

func InitJWTHelpers(cfg *config.Config) *JWTHelpers {
	return &JWTHelpers{cfg}
}

// COMMON HELPERS

func GenerateJWT(data *jwt.MapClaims, secretKey []byte) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, data)
	return token.SignedString(secretKey)
}

func ParseJWT(tokenString string, claims jwt.Claims, secretKey []byte) (*jwt.Token, error) {
	return jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})
}

// ACCOUNT HELPERS

func (h *JWTHelpers) GenerateAccountAccessToken(data *jwt.MapClaims) (string, error) {
	return GenerateJWT(data, []byte(h.JWT.Account.AccessTokenSecret))
}

func (h *JWTHelpers) GenerateAccountRefreshToken(data *jwt.MapClaims) (string, error) {
	return GenerateJWT(data, []byte(h.JWT.Account.RefreshTokenSecret))
}

func (h *JWTHelpers) ParseAccountAccessToken(tokenString string, claims jwt.Claims) error {
	token, err := ParseJWT(tokenString, claims, []byte(h.JWT.Account.AccessTokenSecret))
	if err != nil {
		return err
	}
	if !token.Valid {
		return jwt.ErrSignatureInvalid
	}
	return nil
}

func (h *JWTHelpers) ParseAccountRefreshToken(tokenString string, claims jwt.Claims) error {
	token, err := ParseJWT(tokenString, claims, []byte(h.JWT.Account.RefreshTokenSecret))
	if err != nil {
		return err
	}
	if !token.Valid {
		return jwt.ErrSignatureInvalid
	}
	return nil
}

// SERVICE HELPERS

func (h *JWTHelpers) GenerateServiceAccessToken(data *jwt.MapClaims) (string, error) {
	return GenerateJWT(data, []byte(h.JWT.Service.AccessTokenSecret))
}

func (h *JWTHelpers) GenerateServiceRefreshToken(data *jwt.MapClaims) (string, error) {
	return GenerateJWT(data, []byte(h.JWT.Service.RefreshTokenSecret))
}

func (h *JWTHelpers) GenerateServiceEncryptToken(data *jwt.MapClaims) (string, error) {
	return GenerateJWT(data, []byte(h.JWT.Service.ServiceEncryptSecret))
}

func (h *JWTHelpers) ParseServiceEncryptToken(tokenString string, claims jwt.Claims) error {
	token, err := ParseJWT(tokenString, claims, []byte(h.JWT.Service.ServiceEncryptSecret))
	if err != nil {
		return err
	}
	if !token.Valid {
		return jwt.ErrSignatureInvalid
	}
	return nil
}

func (h *JWTHelpers) ParseServiceAccessToken(tokenString string, claims jwt.Claims) error {
	token, err := ParseJWT(tokenString, claims, []byte(h.JWT.Service.AccessTokenSecret))
	if err != nil {
		return err
	}
	if !token.Valid {
		return jwt.ErrSignatureInvalid
	}
	return nil
}

func (h *JWTHelpers) ParseServiceRefreshToken(tokenString string, claims jwt.Claims) error {
	token, err := ParseJWT(tokenString, claims, []byte(h.JWT.Service.RefreshTokenSecret))
	if err != nil {
		return err
	}
	if !token.Valid {
		return jwt.ErrSignatureInvalid
	}
	return nil
}

func (h *JWTHelpers) ExtractToken(authorization string) string {
	return authorization[7:]
}

func FormatTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}
