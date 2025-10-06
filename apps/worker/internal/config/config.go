package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ilyakaznacheev/cleanenv"
)

// Config represents the application configuration.
type Config struct {
	App       App       `yaml:"app"`
	DB        DB        `yaml:"db"`
	Job       Job       `yaml:"job"`
	Logging   Logging   `yaml:"logging"`
	OTEL      OTEL      `yaml:"otel"`
	Redis     Redis     `yaml:"redis"`
	Replicate Replicate `yaml:"replicate"`
	S3        S3        `yaml:"s3"`
}

type App struct {
	Env string `yaml:"env" env:"APP_ENV" env-default:"dev"`
}

type DB struct {
	PGDatabase string `yaml:"pgdatabase" env:"PGDATABASE" env-default:"virtualstaging"`
	PGHost     string `yaml:"pghost" env:"PGHOST" env-default:"localhost"`
	PGPassword string `yaml:"pgpassword" env:"PGPASSWORD" env-default:"postgres"`
	PGPort     int    `yaml:"pgport" env:"PGPORT" env-default:"5432"`
	PGUser     string `yaml:"pguser" env:"PGUSER" env-default:"postgres"`
	PGSSLMode  string `yaml:"pgsslmode" env:"PGSSLMODE" env-default:"disable"`
}

type Job struct {
	QueueName         string `yaml:"queue_name" env:"JOB_QUEUE_NAME" env-default:"default"`
	WorkerConcurrency int    `yaml:"worker_concurrency" env:"WORKER_CONCURRENCY" env-default:"5"`
}

type Logging struct {
	Level string `yaml:"level" env:"LOG_LEVEL" env-default:"info"`
}

type OTEL struct {
	ExporterOTLPEndpoint string `yaml:"exporter_otlp_endpoint" env:"OTEL_EXPORTER_OTLP_ENDPOINT"`
}

type Redis struct {
	Addr string `yaml:"addr" env:"REDIS_ADDR"`
}

type Replicate struct {
	APIToken     string `yaml:"api_token" env:"REPLICATE_API_TOKEN"`
	ModelVersion string `yaml:"model_version" env:"REPLICATE_MODEL_VERSION" env-default:"qwen/qwen-image-edit"`
}

type S3 struct {
	AccessKey      string `yaml:"access_key" env:"S3_ACCESS_KEY"`
	BucketName     string `yaml:"bucket_name" env:"S3_BUCKET_NAME" env-default:"virtual-staging"`
	Endpoint       string `yaml:"endpoint" env:"S3_ENDPOINT"`
	PublicEndpoint string `yaml:"public_endpoint" env:"S3_PUBLIC_ENDPOINT"`
	Region         string `yaml:"region" env:"S3_REGION" env-default:"us-west-1"`
	SecretKey      string `yaml:"secret_key" env:"S3_SECRET_KEY"`
	UsePathStyle   bool   `yaml:"use_path_style" env:"S3_USE_PATH_STYLE"`
}

// Load loads configuration from YAML files based on APP_ENV.
// It loads config/shared.yml first, then overlays config/{env}.yml,
// then apps/worker/secrets.yml (if present).
// Environment variables take precedence over YAML values.
func Load() (*Config, error) {
	cfg := &Config{}

	// Determine environment
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "dev"
	}

	// Determine config directory (project root /config)
	configDir := os.Getenv("CONFIG_DIR")
	if configDir == "" {
		// Default to config/ relative to project root
		configDir = "config"
	}

	// Load shared config first
	sharedPath := filepath.Join(configDir, "shared.yml")
	if _, err := os.Stat(sharedPath); err == nil {
		if err := cleanenv.ReadConfig(sharedPath, cfg); err != nil {
			return nil, fmt.Errorf("failed to read shared config: %w", err)
		}
	}

	// Load environment-specific config
	envPath := filepath.Join(configDir, fmt.Sprintf("%s.yml", env))
	if _, err := os.Stat(envPath); err == nil {
		if err := cleanenv.ReadConfig(envPath, cfg); err != nil {
			return nil, fmt.Errorf("failed to read %s config: %w", env, err)
		}
	}

	// Load app-specific secrets (if present)
	secretsPath := "secrets.yml"
	if _, err := os.Stat(secretsPath); err == nil {
		if err := cleanenv.ReadConfig(secretsPath, cfg); err != nil {
			return nil, fmt.Errorf("failed to read secrets: %w", err)
		}
	}

	// Read environment variables (these override YAML values)
	if err := cleanenv.ReadEnv(cfg); err != nil {
		return nil, fmt.Errorf("failed to read environment variables: %w", err)
	}

	return cfg, nil
}

// DatabaseURL constructs and returns the full PostgreSQL connection URL.
func (c *Config) DatabaseURL() string {
	// Check if DATABASE_URL is set in environment
	if dbURL := os.Getenv("DATABASE_URL"); dbURL != "" {
		return dbURL
	}

	// Construct from individual components
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.DB.PGUser, c.DB.PGPassword, c.DB.PGHost, c.DB.PGPort, c.DB.PGDatabase, c.DB.PGSSLMode)
}

// S3Bucket returns the S3 bucket name, supporting legacy S3_BUCKET env var.
func (c *Config) S3Bucket() string {
	// Support legacy S3_BUCKET environment variable
	if bucket := os.Getenv("S3_BUCKET"); bucket != "" {
		return bucket
	}
	return c.S3.BucketName
}
