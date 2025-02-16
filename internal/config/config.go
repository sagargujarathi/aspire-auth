package config

import (
	"os"
	"time"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	JWT      JWTConfig
	Email    EmailConfig
}

type ServerConfig struct {
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

type DatabaseConfig struct {
	URL string
}

type RedisConfig struct {
	Address  string
	Password string
	DB       int
}

type JWTConfig struct {
	AccessTokenSecret  string
	RefreshTokenSecret string
	AccessExpiry       time.Duration
	RefreshExpiry      time.Duration
}

type EmailConfig struct {
	From     string
	Username string
	Password string
	Host     string
	Port     int
}

func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port:         os.Getenv("PORT_ADDRESS"),
			ReadTimeout:  time.Second * 10,
			WriteTimeout: time.Second * 10,
		},
		Database: DatabaseConfig{
			URL: os.Getenv("DB_CONNECTION_URL"),
		},
		Redis: RedisConfig{
			Address:  os.Getenv("REDIS_ADDRESS"),
			Password: os.Getenv("REDIS_PASSWORD"),
			DB:       0,
		},
		JWT: JWTConfig{
			AccessTokenSecret:  os.Getenv("ACCESS_TOKEN_SECRET_KEY"),
			RefreshTokenSecret: os.Getenv("REFRESH_TOKEN_SECRET_KEY"),
			AccessExpiry:       time.Minute * 15,
			RefreshExpiry:      time.Hour * 24 * 7,
		},
		Email: EmailConfig{
			From:     os.Getenv("EMAIL_FROM"),
			Username: os.Getenv("EMAIL_USERNAME"),
			Password: os.Getenv("EMAIL_PASSWORD"),
			Host:     os.Getenv("SMTP_HOST"),
			Port:     587, // Default SMTP port
		},
	}
}
