package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Server    ServerConfig
	Database  DatabaseConfig
	AWS       AWSConfig
	S3        S3Config
	SQS       SQSConfig
	Cognito   CognitoConfig
	OpenAI    OpenAIConfig
	Search    SearchConfig
	RateLimit RateLimitConfig
	CORS      CORSConfig
	Logging   LoggingConfig
}

type ServerConfig struct {
	Port string
	Env  string
}

type DatabaseConfig struct {
	URL      string
	Host     string
	Port     string
	Name     string
	User     string
	Password string
}

type AWSConfig struct {
	Region          string
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
}

type S3Config struct {
	Bucket           string
	StorageMode      string
	LocalStoragePath string
}

type SQSConfig struct {
	QueueURL string
}

type CognitoConfig struct {
	UserPoolID string
	ClientID   string
	Region     string
}

type OpenAIConfig struct {
	APIKey         string
	EmbeddingModel string
	ChatModel      string
}

type SearchConfig struct {
	RRFK                int
	TopK                int
	SimilarityThreshold float64
}

type RateLimitConfig struct {
	RequestsPerMinute int
}

type CORSConfig struct {
	AllowedOrigins string
}

type LoggingConfig struct {
	Level  string
	Format string
}

func Load() (*Config, error) {
	// Load .env file if it exists (ignore error in production)
	// Try multiple paths to support running from different directories
	_ = godotenv.Load()                // ./. env (current dir)
	_ = godotenv.Load("../../.env")    // when running from cmd/api/
	_ = godotenv.Load("../.env")       // when running from cmd/
	_ = godotenv.Load("backend/.env")  // when running from project root

	cfg := &Config{
		Server: ServerConfig{
			Port: getEnv("PORT", "8080"),
			Env:  getEnv("ENV", "development"),
		},
		Database: DatabaseConfig{
			URL:      getEnv("DATABASE_URL", ""),
			Host:     getEnv("DATABASE_HOST", "localhost"),
			Port:     getEnv("DATABASE_PORT", "5432"),
			Name:     getEnv("DATABASE_NAME", "documind_dev"),
			User:     getEnv("DATABASE_USER", "postgres"),
			Password: getEnv("DATABASE_PASSWORD", "dev_password"),
		},
		AWS: AWSConfig{
			Region:          getEnv("AWS_REGION", "us-east-1"),
			Endpoint:        getEnv("AWS_ENDPOINT", ""),
			AccessKeyID:     getEnv("AWS_ACCESS_KEY_ID", ""),
			SecretAccessKey: getEnv("AWS_SECRET_ACCESS_KEY", ""),
		},
		S3: S3Config{
			Bucket:           getEnv("S3_BUCKET", "documind-dev"),
			StorageMode:      getEnv("STORAGE_MODE", "local"),
			LocalStoragePath: getEnv("LOCAL_STORAGE_PATH", "./storage"),
		},
		SQS: SQSConfig{
			QueueURL: getEnv("SQS_QUEUE_URL", ""),
		},
		Cognito: CognitoConfig{
			UserPoolID: getEnv("COGNITO_USER_POOL_ID", ""),
			ClientID:   getEnv("COGNITO_CLIENT_ID", ""),
			Region:     getEnv("COGNITO_REGION", "us-east-1"),
		},
		OpenAI: OpenAIConfig{
			APIKey:         getEnv("OPENAI_API_KEY", ""),
			EmbeddingModel: getEnv("OPENAI_EMBEDDING_MODEL", "text-embedding-3-small"),
			ChatModel:      getEnv("OPENAI_CHAT_MODEL", "gpt-4"),
		},
		Search: SearchConfig{
			RRFK:                getEnvAsInt("SEARCH_RRF_K", 60),
			TopK:                getEnvAsInt("SEARCH_TOP_K", 10),
			SimilarityThreshold: getEnvAsFloat("SEARCH_SIMILARITY_THRESHOLD", 0.7),
		},
		RateLimit: RateLimitConfig{
			RequestsPerMinute: getEnvAsInt("RATE_LIMIT_REQUESTS_PER_MINUTE", 60),
		},
		CORS: CORSConfig{
			AllowedOrigins: getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:3005"),
		},
		Logging: LoggingConfig{
			Level:  getEnv("LOG_LEVEL", "info"),
			Format: getEnv("LOG_FORMAT", "json"),
		},
	}

	// Build DATABASE_URL if not provided
	if cfg.Database.URL == "" {
		cfg.Database.URL = fmt.Sprintf(
			"postgres://%s:%s@%s:%s/%s?sslmode=disable",
			cfg.Database.User,
			cfg.Database.Password,
			cfg.Database.Host,
			cfg.Database.Port,
			cfg.Database.Name,
		)
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}

func getEnvAsFloat(key string, defaultValue float64) float64 {
	valueStr := getEnv(key, "")
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.ParseFloat(valueStr, 64)
	if err != nil {
		return defaultValue
	}
	return value
}
