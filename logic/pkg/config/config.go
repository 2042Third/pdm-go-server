package config

import (
	"crypto/ed25519"
	"encoding/base64"
	"fmt"
	"github.com/joho/godotenv"
	"os"
	"strconv"
	"time"
)

type Environment string

const (
	Development Environment = "development"
	Production  Environment = "production"
	Testing     Environment = "testing"
)

type DatabaseConfig struct {
}

type AuthConfig struct {
	PublicKey  ed25519.PublicKey
	PrivateKey ed25519.PrivateKey
}

type Config struct {
	Env           Environment
	Server        ServerConfig
	StaticContent StaticContentConfig
	Redis         RedisConfig
	Database      DatabaseConfig
	Auth          AuthConfig
	Email         EmailConfig
	Logging       LogConfig
	Metrics       MetricsConfig
}

func (c Config) GetEnv(env string) interface{} {
	switch env {
	default:
		return "TODO" // TODO: Implement environment checking
	}
}

type RedisConfig struct {
	Address  string        // Redis server address
	Password string        // Redis password
	DB       int           // Redis database number
	Timeout  time.Duration // Operation timeout
}

type StaticContentConfig struct {
	InternalPath   string
	StatusPassword string
}

type ServerConfig struct {
	Port         string
	Environment  string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

type LogConfig struct {
	Level string
	File  string
	JSON  bool
}

type MetricsConfig struct {
	Enabled bool
	Path    string
}

type EmailConfig struct {
	ApiKey string
}

func LoadConfig() (*Config, error) {
	// Determine environment
	env := getEnvOrDefault("APP_ENV", string(Development))

	// Load appropriate .env file
	envFile := fmt.Sprintf(".env.%s", env)
	if err := godotenv.Load(envFile); err != nil {
		// Fallback to default .env
		if err := godotenv.Load(); err != nil {
			return nil, fmt.Errorf("error loading .env file: %w", err)
		}
	}

	privateKey, publicKey, err := loadKeys()
	if err != nil {
		//log.Fatalf("Failed to load keys: %v", err)
	}

	return &Config{
		Env: Environment(env),
		Server: ServerConfig{
			Port:         getEnvOrDefault("PORT", "8080"),
			Environment:  env,
			ReadTimeout:  getDurationOrDefault("SERVER_READ_TIMEOUT", 30*time.Second),
			WriteTimeout: getDurationOrDefault("SERVER_WRITE_TIMEOUT", 30*time.Second),
		},
		StaticContent: StaticContentConfig{
			InternalPath:   getEnvOrDefault("INTERNAL_PATH", ""),
			StatusPassword: getEnvOrDefault("STATUS_PASSWORD", ""),
		},
		Logging: LogConfig{
			Level: getEnvOrDefault("LOG_LEVEL", "info"),
			File:  getEnvOrDefault("LOG_FILE", ""),
			JSON:  getBoolOrDefault("LOG_JSON", true),
		},
		Metrics: MetricsConfig{
			Enabled: getBoolOrDefault("METRICS_ENABLED", true),
			Path:    getEnvOrDefault("METRICS_PATH", "/metrics"),
		},
		Auth: AuthConfig{
			PublicKey:  publicKey,
			PrivateKey: privateKey,
		},
		Email: EmailConfig{
			ApiKey: getEnvOrDefault("EMAIL_API_KEY", ""),
		},
		Redis: RedisConfig{
			Address:  os.Getenv("REDIS_URL"),
			Password: os.Getenv("REDIS_PASSWORD"),
			DB:       getIntOrDefault("REDIS_DB", 0),
		},
	}, nil
}

func loadKeys() (ed25519.PrivateKey, ed25519.PublicKey, error) {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		return nil, nil, fmt.Errorf("error loading .env file: %w", err)
	}

	// Get encoded keys from environment
	privKeyStr := os.Getenv("JWT_PRIVATE_KEY")
	pubKeyStr := os.Getenv("JWT_PUBLIC_KEY")

	if privKeyStr == "" || pubKeyStr == "" {
		return nil, nil, fmt.Errorf("JWT keys not found in environment")
	}

	// Decode private key
	privKey, err := base64.StdEncoding.DecodeString(privKeyStr)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decode private key: %w", err)
	}

	// Decode public key
	pubKey, err := base64.StdEncoding.DecodeString(pubKeyStr)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decode public key: %w", err)
	}

	// Verify key sizes
	if len(privKey) != ed25519.PrivateKeySize {
		return nil, nil, fmt.Errorf("invalid private key size")
	}
	if len(pubKey) != ed25519.PublicKeySize {
		return nil, nil, fmt.Errorf("invalid public key size")
	}

	return ed25519.PrivateKey(privKey), ed25519.PublicKey(pubKey), nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getDurationOrDefault(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

func getBoolOrDefault(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if b, err := strconv.ParseBool(value); err == nil {
			return b
		}
	}
	return defaultValue
}

func getIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if i, err := strconv.Atoi(value); err == nil {
			return i
		}
	}
	return defaultValue
}
