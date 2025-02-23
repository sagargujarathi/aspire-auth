package helpers

import (
	"aspire-auth/internal/config"
	"aspire-auth/internal/models"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Add this at the top of the file to cache the secret
type JWTHelpers struct {
	*config.Config
	cachedAccessSecret  string
	cachedRefreshSecret string
}

func InitJWTHelpers(cfg *config.Config) *JWTHelpers {
	helper := &JWTHelpers{
		Config:              cfg,
		cachedAccessSecret:  cfg.JWT.Account.AccessTokenSecret,
		cachedRefreshSecret: cfg.JWT.Account.RefreshTokenSecret,
	}

	// Debug print the secrets being used
	return helper
}

// COMMON HELPERS

func GenerateJWT(data *jwt.MapClaims, secretKey []byte) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, data)
	return token.SignedString(secretKey)
}

func ParseJWT(tokenString string, claims jwt.Claims, secretKey []byte) (*jwt.Token, error) {
	fmt.Printf("Debug - Secret key used for parsing: %s\n", string(secretKey)) // Debug log
	return jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secretKey, nil
	})
}

// ACCOUNT HELPERS

func TokenModelToClaims(data *models.AccountRefreshToken) *jwt.MapClaims {
	return &jwt.MapClaims{
		"user_id":    data.UserID.String(),
		"role_type":  data.RoleType,
		"expires_at": data.ExpiresAt.Unix(),
	}
}

func (h *JWTHelpers) GenerateAccountAccessToken(data *models.AccountRefreshToken) (string, error) {
	fmt.Printf("Debug - Generating token with secret: %s\n", h.cachedAccessSecret) // Debug log
	claims := TokenModelToClaims(data)
	return GenerateJWT(claims, []byte(h.cachedAccessSecret))
}

func (h *JWTHelpers) GenerateAccountRefreshToken(data *models.AccountRefreshToken) (string, error) {
	claims := TokenModelToClaims(data)
	return GenerateJWT(claims, []byte(h.cachedRefreshSecret))
}

func (h *JWTHelpers) ParseAccountAccessToken(tokenString string, claims *models.AccountAuthorizationToken) error {
	fmt.Printf("Debug - Using Account Access Token Secret: %s\n", h.cachedAccessSecret) // Debug log
	token, err := ParseJWT(tokenString, claims, []byte(h.cachedAccessSecret))
	if err != nil {
		return err
	}
	if !token.Valid {
		return jwt.ErrSignatureInvalid
	}
	return nil
}

func (h *JWTHelpers) ParseAccountRefreshToken(tokenString string, claims *models.AccountAuthorizationToken) error {
	token, err := ParseJWT(tokenString, claims, []byte(h.cachedRefreshSecret))
	if err != nil {
		return err
	}
	if !token.Valid {
		return jwt.ErrSignatureInvalid
	}
	return nil
}

// SERVICE HELPERS

func ServiceTokenModelToClaims(data *models.ServiceRefreshToken) *jwt.MapClaims {
	return &jwt.MapClaims{
		"user_id":    data.UserID.String(),
		"service_id": data.ServiceID.String(),
		"role_type":  data.RoleType,
		"expires_at": data.ExpiresAt.Unix(),
	}
}

func ServiceSecretKeyToClaims(secretKey string) *jwt.MapClaims {
	return &jwt.MapClaims{
		"secret_key": secretKey,
	}
}

func (h *JWTHelpers) GenerateServiceAccessToken(data *models.ServiceRefreshToken) (string, error) {
	claims := ServiceTokenModelToClaims(data)
	return GenerateJWT(claims, []byte(h.JWT.Service.AccessTokenSecret))
}

func (h *JWTHelpers) GenerateServiceRefreshToken(data *models.ServiceRefreshToken) (string, error) {
	claims := ServiceTokenModelToClaims(data)
	return GenerateJWT(claims, []byte(h.JWT.Service.RefreshTokenSecret))
}

func (h *JWTHelpers) GenerateServiceEncryptToken(secretKey string) (string, error) {
	claims := ServiceSecretKeyToClaims(secretKey)
	return GenerateJWT(claims, []byte(h.JWT.Service.ServiceEncryptSecret))
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
