package utils

import (
	"aspire-auth/internal/config"
	"fmt"
	"log"
	"os"
)

// VerifyEnvironmentSetup checks that all required environment variables are set properly
func VerifyEnvironmentSetup() {
	// Check critical environment variables
	requiredVars := []string{
		"PORT_ADDRESS",
		"DB_CONNECTION_URL",
		"ACCOUNT_ACCESS_TOKEN_SECRET_KEY",
		"ACCOUNT_REFRESH_TOKEN_SECRET_KEY",
		"SERVICE_ACCESS_TOKEN_SECRET_KEY",
		"SERVICE_REFRESH_TOKEN_SECRET_KEY",
		"SERVICE_ENCRYPT_SECRET_KEY",
	}

	missingVars := []string{}
	for _, envVar := range requiredVars {
		if os.Getenv(envVar) == "" {
			missingVars = append(missingVars, envVar)
		}
	}

	if len(missingVars) > 0 {
		log.Printf("WARNING: Missing required environment variables: %v", missingVars)
	} else {
		log.Println("All required environment variables are set")
	}

	// Verify service and account secrets are different
	if os.Getenv("SERVICE_ACCESS_TOKEN_SECRET_KEY") == os.Getenv("ACCOUNT_ACCESS_TOKEN_SECRET_KEY") {
		log.Printf("WARNING: Service and account access token secrets are identical. This can cause validation issues.")
	}

	// Check for secret key duplicates
	checkSecretKeyDuplicates()
}

// Check if any secret keys are duplicated
func checkSecretKeyDuplicates() {
	secretKeys := map[string]string{
		"ACCOUNT_ACCESS_TOKEN_SECRET_KEY":  os.Getenv("ACCOUNT_ACCESS_TOKEN_SECRET_KEY"),
		"ACCOUNT_REFRESH_TOKEN_SECRET_KEY": os.Getenv("ACCOUNT_REFRESH_TOKEN_SECRET_KEY"),
		"SERVICE_ACCESS_TOKEN_SECRET_KEY":  os.Getenv("SERVICE_ACCESS_TOKEN_SECRET_KEY"),
		"SERVICE_REFRESH_TOKEN_SECRET_KEY": os.Getenv("SERVICE_REFRESH_TOKEN_SECRET_KEY"),
		"SERVICE_ENCRYPT_SECRET_KEY":       os.Getenv("SERVICE_ENCRYPT_SECRET_KEY"),
	}

	// Map to track which keys have the same value
	valueToKeys := make(map[string][]string)

	for key, value := range secretKeys {
		valueToKeys[value] = append(valueToKeys[value], key)
	}

	// Check for duplicates
	for value, keys := range valueToKeys {
		if len(keys) > 1 && value != "" {
			log.Printf("WARNING: The following keys share the same value: %v", keys)
		}
	}
}

// PrintConfigSummary outputs a summary of the loaded configuration
func PrintConfigSummary(cfg *config.Config) {
	fmt.Println("===== Configuration Summary =====")
	fmt.Printf("Server Port: %s\n", cfg.Server.Port)
	fmt.Printf("Account JWT Secrets: %s... / %s...\n",
		maskString(cfg.JWT.Account.AccessTokenSecret, 8),
		maskString(cfg.JWT.Account.RefreshTokenSecret, 8))
	fmt.Printf("Service JWT Secrets: %s... / %s...\n",
		maskString(cfg.JWT.Service.AccessTokenSecret, 8),
		maskString(cfg.JWT.Service.RefreshTokenSecret, 8))
	fmt.Println("===============================")
}

// maskString masks a string for secure display
func maskString(s string, showChars int) string {
	if len(s) <= showChars {
		return s
	}
	return s[:showChars] + "***"
}
