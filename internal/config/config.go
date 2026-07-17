package config

import (
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	AppEnv  string
	AppPort int

	DBHost     string
	DBPort     int
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	JWTSecret     string
	JWTAccessTTL  time.Duration
	JWTRefreshTTL time.Duration

	LogLevel  string
	AdminUser string
	AdminPass string
}

func Load(path string) (*Config, error) {
	viper.SetConfigFile(path)
	viper.SetConfigType("env")

	viper.AutomaticEnv()

	viper.SetDefault("APP_ENV", "development")
	viper.SetDefault("APP_PORT", 8080)
	viper.SetDefault("DB_PORT", 5432)
	viper.SetDefault("DB_SSLMODE", "disable")
	viper.SetDefault("JWT_ACCESS_TTL", "15m")
	viper.SetDefault("JWT_REFRESH_TTL", "72h")
	viper.SetDefault("LOG_LEVEL", "debug")
	viper.SetDefault("ADMIN_USER", "admin")
	viper.SetDefault("ADMIN_PASS", "admin")

	if err := viper.ReadInConfig(); err != nil {
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	accessTTL, err := time.ParseDuration(viper.GetString("JWT_ACCESS_TTL"))
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_ACCESS_TTL: %w", err)
	}

	refreshTTL, err := time.ParseDuration(viper.GetString("JWT_REFRESH_TTL"))
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_REFRESH_TTL: %w", err)
	}

	cfg := &Config{
		AppEnv:        viper.GetString("APP_ENV"),
		AppPort:       viper.GetInt("APP_PORT"),
		DBHost:        viper.GetString("DB_HOST"),
		DBPort:        viper.GetInt("DB_PORT"),
		DBUser:        viper.GetString("DB_USER"),
		DBPassword:    viper.GetString("DB_PASSWORD"),
		DBName:        viper.GetString("DB_NAME"),
		DBSSLMode:     viper.GetString("DB_SSLMODE"),
		JWTSecret:     viper.GetString("JWT_SECRET"),
		JWTAccessTTL:  accessTTL,
		JWTRefreshTTL: refreshTTL,
		LogLevel:      viper.GetString("LOG_LEVEL"),
		AdminUser:     viper.GetString("ADMIN_USER"),
		AdminPass:     viper.GetString("ADMIN_PASS"),
	}

	return cfg, nil
}

func (c *Config) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.DBHost, c.DBPort, c.DBUser, url.QueryEscape(c.DBPassword), c.DBName, c.DBSSLMode,
	)
}
