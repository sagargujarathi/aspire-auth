package helpers

import (
	"os"

	"github.com/golang-jwt/jwt/v5"
)

var ACCESS_TOKEN_SECRET_KEY = []byte(os.Getenv("ACCESS_TOKEN_SECRET_KEY"))
var REFRESH_TOKEN_SECRET_KEY = []byte(os.Getenv("REFRESH_TOKEN_SECRET_KEY"))

func GenerateJWT(data *jwt.MapClaims, secretKey []byte) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, data)
	return token.SignedString(secretKey)
}

func ParseJWT(tokenString string, claims jwt.Claims, secretKey []byte) (*jwt.Token, error) {
	return jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})
}

func GenerateAccessToken(data *jwt.MapClaims) (string, error) {
	return GenerateJWT(data, ACCESS_TOKEN_SECRET_KEY)
}

func GenerateRefreshToken(data *jwt.MapClaims) (string, error) {
	return GenerateJWT(data, REFRESH_TOKEN_SECRET_KEY)
}

func ParseAccessToken(tokenString string, claims jwt.Claims) error {
	token, err := ParseJWT(tokenString, claims, ACCESS_TOKEN_SECRET_KEY)
	if err != nil {
		return err
	}
	if !token.Valid {
		return jwt.ErrSignatureInvalid
	}
	return nil
}

func ParseRefreshToken(tokenString string, claims jwt.Claims) error {
	token, err := ParseJWT(tokenString, claims, REFRESH_TOKEN_SECRET_KEY)
	if err != nil {
		return err
	}
	if !token.Valid {
		return jwt.ErrSignatureInvalid
	}
	return nil
}

func ExtractToken(authorization string) string {
	return authorization[7:]
}
