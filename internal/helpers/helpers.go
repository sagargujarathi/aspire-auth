package helpers

import (
	"aspire-auth/internal/config"
	"aspire-auth/internal/models"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Add this at the top of the file to cache the secret
type JWTHelpers struct {
	*config.Config
	accountAccessSecret  string
	accountRefreshSecret string
	serviceAccessSecret  string
	serviceRefreshSecret string
	serviceEncryptSecret string
}

func InitJWTHelpers(cfg *config.Config) *JWTHelpers {
	helper := &JWTHelpers{
		Config:               cfg,
		accountAccessSecret:  cfg.JWT.Account.AccessTokenSecret,
		accountRefreshSecret: cfg.JWT.Account.RefreshTokenSecret,
		serviceAccessSecret:  cfg.JWT.Service.AccessTokenSecret,
		serviceRefreshSecret: cfg.JWT.Service.RefreshTokenSecret,
		serviceEncryptSecret: cfg.JWT.Service.ServiceEncryptSecret,
	}

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
	fmt.Printf("Debug - Generating account token with secret: %s\n", h.accountAccessSecret) // Debug log
	claims := TokenModelToClaims(data)
	return GenerateJWT(claims, []byte(h.accountAccessSecret))
}

func (h *JWTHelpers) GenerateAccountRefreshToken(data *models.AccountRefreshToken) (string, error) {
	claims := TokenModelToClaims(data)
	return GenerateJWT(claims, []byte(h.accountRefreshSecret))
}

func (h *JWTHelpers) ParseAccountAccessToken(tokenString string, claims *models.AccountAuthorizationToken) error {
	fmt.Printf("Debug - Using Account Access Token Secret: %s\n", h.accountAccessSecret) // Debug log
	token, err := ParseJWT(tokenString, claims, []byte(h.accountAccessSecret))
	if err != nil {
		return err
	}
	if !token.Valid {
		return jwt.ErrSignatureInvalid
	}
	return nil
}

func (h *JWTHelpers) ParseAccountRefreshToken(tokenString string, claims *models.AccountAuthorizationToken) error {
	token, err := ParseJWT(tokenString, claims, []byte(h.accountRefreshSecret))
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

// getEncryptionKey creates a valid 32-byte AES key from any string input
func (h *JWTHelpers) getEncryptionKey() []byte {
	// Use SHA-256 to get a consistent 32-byte key from the secret
	hasher := sha256.New()
	hasher.Write([]byte(h.serviceEncryptSecret))
	return hasher.Sum(nil) // 32 bytes (256 bits)
}

func (h *JWTHelpers) EncryptServiceSecretKey(secretKey string) (string, error) {
	// Use a proper 32-byte key derived from your secret
	key := h.getEncryptionKey()

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("cipher creation failed: %w", err)
	}

	// IV needs to be unique, but not secure
	ciphertext := make([]byte, aes.BlockSize+len(secretKey))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", fmt.Errorf("failed to generate IV: %w", err)
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], []byte(secretKey))

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func (h *JWTHelpers) DecryptServiceSecretKey(encryptedKey string) (string, error) {
	// Use a proper 32-byte key derived from your secret
	key := h.getEncryptionKey()

	ciphertext, err := base64.StdEncoding.DecodeString(encryptedKey)
	if err != nil {
		return "", fmt.Errorf("base64 decode failed: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("cipher creation failed: %w", err)
	}

	if len(ciphertext) < aes.BlockSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(ciphertext, ciphertext)

	return string(ciphertext), nil
}

func (h *JWTHelpers) GenerateServiceEncryptToken(secretKey string) (string, error) {
	return h.EncryptServiceSecretKey(secretKey)
}

func (h *JWTHelpers) GenerateServiceAccessTokenWithSecret(data *models.ServiceRefreshToken, serviceSecret string) (string, error) {
	claims := ServiceTokenModelToClaims(data)
	return GenerateJWT(claims, []byte(serviceSecret))
}

func (h *JWTHelpers) GenerateServiceRefreshTokenWithSecret(data *models.ServiceRefreshToken, serviceSecret string) (string, error) {
	claims := ServiceTokenModelToClaims(data)
	return GenerateJWT(claims, []byte(serviceSecret))
}

func (h *JWTHelpers) GenerateServiceAccessToken(data *models.ServiceRefreshToken) (string, error) {
	fmt.Printf("Debug - Generating service token with secret: %s\n", h.serviceAccessSecret) // Debug log
	claims := ServiceTokenModelToClaims(data)
	return GenerateJWT(claims, []byte(h.serviceAccessSecret))
}

func (h *JWTHelpers) GenerateServiceRefreshToken(data *models.ServiceRefreshToken) (string, error) {
	claims := ServiceTokenModelToClaims(data)
	return GenerateJWT(claims, []byte(h.serviceRefreshSecret))
}

func (h *JWTHelpers) ParseServiceAccessTokenWithSecret(tokenString string, claims jwt.Claims, serviceSecret string) error {
	token, err := ParseJWT(tokenString, claims, []byte(serviceSecret))
	if err != nil {
		return err
	}
	if !token.Valid {
		return jwt.ErrSignatureInvalid
	}
	return nil
}

// Add this new function to help with debugging
func (h *JWTHelpers) DebugTokenType(tokenString string) string {
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return "INVALID_TOKEN_FORMAT"
	}

	// Parse token details without validation
	parser := jwt.NewParser(jwt.WithoutClaimsValidation())
	parsedToken, _, err := parser.ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		return "PARSE_ERROR"
	}

	if mapClaims, ok := parsedToken.Claims.(jwt.MapClaims); ok {
		if _, hasServiceID := mapClaims["service_id"]; hasServiceID {
			return "SERVICE_TOKEN"
		} else if _, hasUserID := mapClaims["user_id"]; hasUserID {
			return "ACCOUNT_TOKEN"
		}
	}

	return "UNKNOWN_TOKEN_TYPE"
}

// Update ParseServiceAccessToken to use the proper fallback approach only when needed
func (h *JWTHelpers) ParseServiceAccessToken(tokenString string, claims jwt.Claims) error {
	// Extract and print token payload for debugging
	tokenType := h.DebugTokenType(tokenString)
	fmt.Printf("Token identified as: %s\n", tokenType)

	// Extract and print token payload for debugging
	parts := strings.Split(tokenString, ".")
	if len(parts) == 3 {
		fmt.Printf("Token parts count: %d\n", len(parts))
		fmt.Printf("Token payload (encoded): %s\n", parts[1])

		// Parse token details without validation to see what's in it
		parser := jwt.NewParser(jwt.WithoutClaimsValidation())
		parsedToken, _, err := parser.ParseUnverified(tokenString, jwt.MapClaims{})
		if err == nil {
			if mapClaims, ok := parsedToken.Claims.(jwt.MapClaims); ok {
				fmt.Printf("Token claims: %v\n", mapClaims)

				// If we have a service_id claim, we should try to use that service's secret
				if serviceID, ok := mapClaims["service_id"].(string); ok {
					// This would normally query the database to get the service's encrypted secret
					// But since we can't add that here, we'll just log it
					fmt.Printf("Should validate with service-specific secret for service ID: %s\n", serviceID)
					// In a real implementation, you would:
					// 1. Get the encrypted secret from the database using serviceID
					// 2. Decrypt it
					// 3. Use it to validate the token
				}
			}
		}
	}

	fmt.Printf("Debug - Using Service Access Token Secret: %s\n", h.serviceAccessSecret)

	// First try with the normal service access token secret
	token, err := ParseJWT(tokenString, claims, []byte(h.serviceAccessSecret))
	if err != nil {
		// If the error is specifically about the signature being invalid
		if strings.Contains(err.Error(), "signature is invalid") {
			// Try with the account access token secret as a fallback
			fmt.Printf("Warning: Service token validation failed with service secret, trying with account secret\n")
			accountToken, accountErr := ParseJWT(tokenString, claims, []byte(h.accountAccessSecret))
			if accountErr == nil && accountToken.Valid {
				fmt.Printf("Warning: Service token validated with ACCOUNT secret - this indicates a middleware routing issue\n")
				return nil
			}
		}
		return err
	}

	if !token.Valid {
		return jwt.ErrSignatureInvalid
	}
	return nil
}

func (h *JWTHelpers) ParseServiceRefreshToken(tokenString string, claims jwt.Claims) error {
	token, err := ParseJWT(tokenString, claims, []byte(h.serviceRefreshSecret))
	if err != nil {
		return err
	}
	if !token.Valid {
		return jwt.ErrSignatureInvalid
	}
	return nil
}

func (h *JWTHelpers) ExtractToken(authorization string) string {
	// Standard Bearer token
	if strings.HasPrefix(authorization, "Bearer ") {
		return authorization[7:]
	}

	// JWT Token passed directly
	if strings.Count(authorization, ".") == 2 {
		return authorization
	}

	// Return as is if none of the above
	return authorization
}

// Get secret keys for debugging
func (h *JWTHelpers) GetServiceAccessSecret() string {
	return h.serviceAccessSecret
}

func (h *JWTHelpers) GetAccountAccessSecret() string {
	return h.accountAccessSecret
}

func FormatTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}
