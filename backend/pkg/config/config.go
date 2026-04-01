package config

import (
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config holds all application configuration loaded from environment variables
// or a configuration file (config.yaml / .env).
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	Storage  StorageConfig
	Auth     AuthConfig
	AI       AIConfig
}

type ServerConfig struct {
	Host         string        `mapstructure:"HOST"`
	Port         string        `mapstructure:"PORT"`
	ReadTimeout  time.Duration `mapstructure:"READ_TIMEOUT"`
	WriteTimeout time.Duration `mapstructure:"WRITE_TIMEOUT"`
	Environment  string        `mapstructure:"ENVIRONMENT"` // "development" | "production"
}

type DatabaseConfig struct {
	DSN             string `mapstructure:"DATABASE_URL"`
	MaxOpenConns    int    `mapstructure:"DB_MAX_OPEN_CONNS"`
	MaxIdleConns    int    `mapstructure:"DB_MAX_IDLE_CONNS"`
	ConnMaxLifetime time.Duration
}

type RedisConfig struct {
	Addr     string `mapstructure:"REDIS_ADDR"`
	Password string `mapstructure:"REDIS_PASSWORD"`
	DB       int    `mapstructure:"REDIS_DB"`
}

type StorageConfig struct {
	Endpoint        string `mapstructure:"S3_ENDPOINT"`
	Region          string `mapstructure:"S3_REGION"`
	Bucket          string `mapstructure:"S3_BUCKET"`
	AccessKeyID     string `mapstructure:"S3_ACCESS_KEY_ID"`
	SecretAccessKey string `mapstructure:"S3_SECRET_ACCESS_KEY"`
	UsePathStyle    bool   `mapstructure:"S3_USE_PATH_STYLE"` // true for MinIO
}

type AuthConfig struct {
	JWTSecret          string        `mapstructure:"JWT_SECRET"`
	AccessTokenExpiry  time.Duration `mapstructure:"JWT_ACCESS_EXPIRY"`
	RefreshTokenExpiry time.Duration `mapstructure:"JWT_REFRESH_EXPIRY"`
	// OIDC
	OIDCIssuer       string `mapstructure:"OIDC_ISSUER"`
	OIDCClientID     string `mapstructure:"OIDC_CLIENT_ID"`
	OIDCClientSecret string `mapstructure:"OIDC_CLIENT_SECRET"`
}

type AIConfig struct {
	OllamaBaseURL string `mapstructure:"OLLAMA_BASE_URL"`
	OllamaModel   string `mapstructure:"OLLAMA_MODEL"`
	OpenAIAPIKey  string `mapstructure:"OPENAI_API_KEY"`
}

// Load reads configuration from environment variables and an optional config file.
func Load() (*Config, error) {
	v := viper.New()

	// Defaults
	v.SetDefault("HOST", "0.0.0.0")
	v.SetDefault("PORT", "8080")
	v.SetDefault("ENVIRONMENT", "development")
	v.SetDefault("READ_TIMEOUT", 30*time.Second)
	v.SetDefault("WRITE_TIMEOUT", 30*time.Second)
	v.SetDefault("DB_MAX_OPEN_CONNS", 25)
	v.SetDefault("DB_MAX_IDLE_CONNS", 5)
	v.SetDefault("REDIS_ADDR", "localhost:6379")
	v.SetDefault("REDIS_DB", 0)
	v.SetDefault("S3_REGION", "us-east-1")
	v.SetDefault("S3_USE_PATH_STYLE", true)
	v.SetDefault("JWT_ACCESS_EXPIRY", 15*time.Minute)
	v.SetDefault("JWT_REFRESH_EXPIRY", 7*24*time.Hour)
	v.SetDefault("OLLAMA_MODEL", "llava")

	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.AddConfigPath("./config")

	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Ignore "config file not found" — env vars alone are sufficient.
	_ = v.ReadInConfig()

	cfg := &Config{}
	if err := v.Unmarshal(cfg); err != nil {
		return nil, err
	}

	// Map flat env-var keys into nested structs.
	cfg.Server = ServerConfig{
		Host:         v.GetString("HOST"),
		Port:         v.GetString("PORT"),
		ReadTimeout:  v.GetDuration("READ_TIMEOUT"),
		WriteTimeout: v.GetDuration("WRITE_TIMEOUT"),
		Environment:  v.GetString("ENVIRONMENT"),
	}
	cfg.Database = DatabaseConfig{
		DSN:             v.GetString("DATABASE_URL"),
		MaxOpenConns:    v.GetInt("DB_MAX_OPEN_CONNS"),
		MaxIdleConns:    v.GetInt("DB_MAX_IDLE_CONNS"),
		ConnMaxLifetime: 5 * time.Minute,
	}
	cfg.Redis = RedisConfig{
		Addr:     v.GetString("REDIS_ADDR"),
		Password: v.GetString("REDIS_PASSWORD"),
		DB:       v.GetInt("REDIS_DB"),
	}
	cfg.Storage = StorageConfig{
		Endpoint:        v.GetString("S3_ENDPOINT"),
		Region:          v.GetString("S3_REGION"),
		Bucket:          v.GetString("S3_BUCKET"),
		AccessKeyID:     v.GetString("S3_ACCESS_KEY_ID"),
		SecretAccessKey: v.GetString("S3_SECRET_ACCESS_KEY"),
		UsePathStyle:    v.GetBool("S3_USE_PATH_STYLE"),
	}
	cfg.Auth = AuthConfig{
		JWTSecret:          v.GetString("JWT_SECRET"),
		AccessTokenExpiry:  v.GetDuration("JWT_ACCESS_EXPIRY"),
		RefreshTokenExpiry: v.GetDuration("JWT_REFRESH_EXPIRY"),
		OIDCIssuer:         v.GetString("OIDC_ISSUER"),
		OIDCClientID:       v.GetString("OIDC_CLIENT_ID"),
		OIDCClientSecret:   v.GetString("OIDC_CLIENT_SECRET"),
	}
	cfg.AI = AIConfig{
		OllamaBaseURL: v.GetString("OLLAMA_BASE_URL"),
		OllamaModel:   v.GetString("OLLAMA_MODEL"),
		OpenAIAPIKey:  v.GetString("OPENAI_API_KEY"),
	}

	return cfg, nil
}
