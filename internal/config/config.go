package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	AppEnv string
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

	LogLevel string
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

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	viper.BindEnv("APP_ENV")
	viper.BindEnv("APP_PORT")
	viper.BindEnv("DB_HOST")
	viper.BindEnv("DB_PORT")
	viper.BindEnv("DB_USER")
	viper.BindEnv("DB_PASSWORD")
	viper.BindEnv("DB_NAME")
	viper.BindEnv("DB_SSLMODE")
	viper.BindEnv("JWT_SECRET")
	viper.BindEnv("JWT_ACCESS_TTL")
	viper.BindEnv("JWT_REFRESH_TTL")
	viper.BindEnv("LOG_LEVEL")

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
	}

	return cfg, nil
}

func (c *Config) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName, c.DBSSLMode,
	)
}
