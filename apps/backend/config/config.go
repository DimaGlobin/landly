package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config is the main application configuration structure
type Config struct {
	App           AppConfig           `mapstructure:"app"`
	Server        ServerConfig        `mapstructure:"server"`
	Auth          AuthConfig          `mapstructure:"auth"`
	Database      DatabaseConfig      `mapstructure:"database"`
	Storage       StorageConfig       `mapstructure:"storage"`
	AI            AIConfig            `mapstructure:"ai"`
	Render        RenderConfig        `mapstructure:"render"`
	Logging       LoggingConfig       `mapstructure:"logging"`
	Observability ObservabilityConfig `mapstructure:"observability"`
}

type AppConfig struct {
	Env     string `mapstructure:"env"`
	Name    string `mapstructure:"name"`
	Version string `mapstructure:"version"`
	BaseURL string `mapstructure:"base_url"`
}

type ServerConfig struct {
	HTTP HTTPConfig `mapstructure:"http"`
	CORS CORSConfig `mapstructure:"cors"`
}

type HTTPConfig struct {
	Addr         string        `mapstructure:"addr"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
}

type CORSConfig struct {
	AllowedOrigins []string `mapstructure:"allowed_origins"`
	AllowedMethods []string `mapstructure:"allowed_methods"`
	AllowedHeaders []string `mapstructure:"allowed_headers"`
}

type AuthConfig struct {
	JWT JWTConfig `mapstructure:"jwt"`
}

type JWTConfig struct {
	Secret          string        `mapstructure:"secret"`
	AccessTokenTTL  time.Duration `mapstructure:"access_token_ttl"`
	RefreshTokenTTL time.Duration `mapstructure:"refresh_token_ttl"`
}

type DatabaseConfig struct {
	Postgres PostgresConfig `mapstructure:"postgres"`
	Redis    RedisConfig    `mapstructure:"redis"`
}

type PostgresConfig struct {
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	User            string        `mapstructure:"user"`
	Password        string        `mapstructure:"password"`
	DBName          string        `mapstructure:"dbname"`
	SSLMode         string        `mapstructure:"sslmode"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
	DSN             string        // Computed field
}

type RedisConfig struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
	PoolSize int    `mapstructure:"pool_size"`
}

type StorageConfig struct {
	S3  S3Config  `mapstructure:"s3"`
	CDN CDNConfig `mapstructure:"cdn"`
}

type S3Config struct {
	Endpoint  string `mapstructure:"endpoint"`
	AccessKey string `mapstructure:"access_key"`
	SecretKey string `mapstructure:"secret_key"`
	Bucket    string `mapstructure:"bucket"`
	UseSSL    bool   `mapstructure:"use_ssl"`
	Region    string `mapstructure:"region"`
}

type CDNConfig struct {
	BaseURL string `mapstructure:"base_url"`
	Enabled bool   `mapstructure:"enabled"`
}

type AIConfig struct {
	Provider  string          `mapstructure:"provider"`
	OpenAI    OpenAIConfig    `mapstructure:"openai"`
	Anthropic AnthropicConfig `mapstructure:"anthropic"`
}

type OpenAIConfig struct {
	APIKey      string  `mapstructure:"api_key"`
	Model       string  `mapstructure:"model"`
	MaxTokens   int     `mapstructure:"max_tokens"`
	Temperature float64 `mapstructure:"temperature"`
}

type AnthropicConfig struct {
	APIKey    string `mapstructure:"api_key"`
	Model     string `mapstructure:"model"`
	MaxTokens int    `mapstructure:"max_tokens"`
}

type RenderConfig struct {
	TmpDir       string        `mapstructure:"tmp_dir"`
	CleanupAfter time.Duration `mapstructure:"cleanup_after"`
}

type LoggingConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
	Output string `mapstructure:"output"`
}

type ObservabilityConfig struct {
	Metrics MetricsConfig `mapstructure:"metrics"`
	Tracing TracingConfig `mapstructure:"tracing"`
}

type MetricsConfig struct {
	Enabled bool `mapstructure:"enabled"`
	Port    int  `mapstructure:"port"`
}

type TracingConfig struct {
	Enabled    bool    `mapstructure:"enabled"`
	Endpoint   string  `mapstructure:"endpoint"`
	SampleRate float64 `mapstructure:"sample_rate"`
}

// Load loads configuration from YAML files and environment variables.
// Priority order (lowest to highest):
// 1. Built-in defaults
// 2. config.yml (optional)
// 3. config.local.yml (optional)
// 4. Environment variables (LANDLY_*)
// 5. PORT env var (for serverless containers)
//
// For production/serverless: config.yml is NOT included in the Docker image.
// All configuration must come from environment variables.
func Load() (*Config, error) {
	v := viper.New()

	// Track config sources for logging
	var configSources []string

	// 1. Set defaults (allows running without config.yml)
	setDefaults(v)
	configSources = append(configSources, "defaults")

	// 2. Configure env vars BEFORE reading files (so they can override)
	// Format: LANDLY_DATABASE_POSTGRES_HOST
	v.SetEnvPrefix("LANDLY")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()
	configSources = append(configSources, "env:LANDLY_*")

	// 3. Try to load config.yml (optional, not present in production images)
	configPath := findConfigPath()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(configPath)
	v.AddConfigPath(".")

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// config.yml not found - this is normal for serverless/production
			// Continue with defaults + env vars
			fmt.Fprintf(os.Stderr, "[config] source: %s (config.yml not found)\n", strings.Join(configSources, " -> "))
		} else {
			// File found but failed to read/parse
			return nil, fmt.Errorf("failed to parse config.yml: %w", err)
		}
	} else {
		configSources = append(configSources, "file:"+v.ConfigFileUsed())
		fmt.Fprintf(os.Stderr, "[config] source: %s\n", strings.Join(configSources, " -> "))
	}

	// 4. Merge config.local.yml if exists (for local development)
	v.SetConfigName("config.local")
	if err := v.MergeInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to parse config.local.yml: %w", err)
		}
	} else {
		fmt.Fprintf(os.Stderr, "[config] merged: config.local.yml\n")
	}

	// 5. Unmarshal into struct
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Build Postgres DSN from individual fields
	cfg.Database.Postgres.DSN = buildPostgresDSN(&cfg.Database.Postgres)

	if cfg.App.BaseURL == "" {
		cfg.App.BaseURL = "http://localhost:8080"
	}

	// 6. Override server address with PORT env var (for serverless containers)
	// PORT has the highest priority for listen address
	if port := os.Getenv("PORT"); port != "" {
		cfg.Server.HTTP.Addr = ":" + port
		fmt.Fprintf(os.Stderr, "[config] PORT=%s override -> listening on %s\n", port, cfg.Server.HTTP.Addr)
	}

	// 7. Validate required fields
	missingVars, validationErr := validateConfig(&cfg)

	if validationErr != nil {
		// Validation errors (e.g., JWT secret too short) are always fatal
		return nil, fmt.Errorf("config validation failed: %w", validationErr)
	}

	if len(missingVars) > 0 {
		// Log missing vars (without values for security)
		fmt.Fprintf(os.Stderr, "[config] missing required env vars: %s\n", strings.Join(missingVars, ", "))

		if IsBootstrapMode() {
			// Bootstrap mode: log warning but continue startup
			fmt.Fprintf(os.Stderr, "[config] LANDLY_BOOTSTRAP_MODE=1: skipping validation, app may not function correctly\n")
		} else {
			// Strict mode: return error
			return nil, fmt.Errorf("missing required environment variables: %s (set LANDLY_BOOTSTRAP_MODE=1 to skip)", strings.Join(missingVars, ", "))
		}
	}

	return &cfg, nil
}

// IsBootstrapMode checks if bootstrap mode is enabled.
// In bootstrap mode, the app starts without required env vars (for health checks).
// Activated via LANDLY_BOOTSTRAP_MODE=1 or LANDLY_BOOTSTRAP_MODE=true
func IsBootstrapMode() bool {
	val := strings.ToLower(os.Getenv("LANDLY_BOOTSTRAP_MODE"))
	return val == "1" || val == "true" || val == "yes"
}

// setDefaults sets default values for all config fields.
// This allows the app to run with env vars only, without config.yml.
func setDefaults(v *viper.Viper) {
	// App defaults
	v.SetDefault("app.env", "production")
	v.SetDefault("app.name", "landly")
	v.SetDefault("app.version", "1.0.0")
	v.SetDefault("app.base_url", "")

	// Server defaults
	v.SetDefault("server.http.addr", ":8080")
	v.SetDefault("server.http.read_timeout", "30s")
	v.SetDefault("server.http.write_timeout", "30s")
	v.SetDefault("server.cors.allowed_origins", []string{"*"})
	v.SetDefault("server.cors.allowed_methods", []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"})
	v.SetDefault("server.cors.allowed_headers", []string{"Authorization", "Content-Type"})

	// Auth defaults (secret MUST be provided via env)
	v.SetDefault("auth.jwt.secret", "")
	v.SetDefault("auth.jwt.access_token_ttl", "15m")
	v.SetDefault("auth.jwt.refresh_token_ttl", "168h")

	// Database defaults
	v.SetDefault("database.postgres.host", "")
	v.SetDefault("database.postgres.port", 5432)
	v.SetDefault("database.postgres.user", "landly")
	v.SetDefault("database.postgres.password", "")
	v.SetDefault("database.postgres.dbname", "landly")
	v.SetDefault("database.postgres.sslmode", "require")
	v.SetDefault("database.postgres.max_open_conns", 25)
	v.SetDefault("database.postgres.max_idle_conns", 5)
	v.SetDefault("database.postgres.conn_max_lifetime", "5m")

	v.SetDefault("database.redis.addr", "")
	v.SetDefault("database.redis.password", "")
	v.SetDefault("database.redis.db", 0)
	v.SetDefault("database.redis.pool_size", 10)

	// Storage defaults
	v.SetDefault("storage.s3.endpoint", "")
	v.SetDefault("storage.s3.access_key", "")
	v.SetDefault("storage.s3.secret_key", "")
	v.SetDefault("storage.s3.bucket", "")
	v.SetDefault("storage.s3.use_ssl", true)
	v.SetDefault("storage.s3.region", "ru-central1")

	v.SetDefault("storage.cdn.base_url", "")
	v.SetDefault("storage.cdn.enabled", false)

	// AI defaults
	v.SetDefault("ai.provider", "mock")
	v.SetDefault("ai.openai.api_key", "")
	v.SetDefault("ai.openai.model", "gpt-4")
	v.SetDefault("ai.openai.max_tokens", 4000)
	v.SetDefault("ai.openai.temperature", 0.7)
	v.SetDefault("ai.anthropic.api_key", "")
	v.SetDefault("ai.anthropic.model", "claude-3-opus-20240229")
	v.SetDefault("ai.anthropic.max_tokens", 4000)

	// Render defaults
	v.SetDefault("render.tmp_dir", "/tmp/landly")
	v.SetDefault("render.cleanup_after", "1h")

	// Logging defaults
	v.SetDefault("logging.level", "info")
	v.SetDefault("logging.format", "json")
	v.SetDefault("logging.output", "stdout")

	// Observability defaults
	v.SetDefault("observability.metrics.enabled", false)
	v.SetDefault("observability.metrics.port", 9090)
	v.SetDefault("observability.tracing.enabled", false)
	v.SetDefault("observability.tracing.endpoint", "")
	v.SetDefault("observability.tracing.sample_rate", 0.1)
}

// findConfigPath searches for config.yml in project root
func findConfigPath() string {
	// Check current directory
	if _, err := os.Stat("config.yml"); err == nil {
		return "."
	}

	// Check parent directories (for running from apps/backend)
	dir, _ := os.Getwd()
	for i := 0; i < 3; i++ {
		dir = filepath.Dir(dir)
		configFile := filepath.Join(dir, "config.yml")
		if _, err := os.Stat(configFile); err == nil {
			return dir
		}
	}

	return "."
}

// buildPostgresDSN builds DSN string for PostgreSQL connection
func buildPostgresDSN(cfg *PostgresConfig) string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.DBName,
		cfg.SSLMode,
	)
}

// validateConfig checks required fields.
// Returns:
//   - []string: list of missing env vars (may be empty)
//   - error: critical validation error (e.g., JWT secret too short)
//
// Critical errors block startup even in bootstrap mode.
// Missing env vars block startup only in strict mode (without bootstrap).
func validateConfig(cfg *Config) ([]string, error) {
	var missingEnvVars []string

	// JWT Secret - check presence and validity
	if cfg.Auth.JWT.Secret == "" {
		missingEnvVars = append(missingEnvVars, "LANDLY_AUTH_JWT_SECRET")
	} else if len(cfg.Auth.JWT.Secret) < 32 {
		// Critical error: secret is set but too short
		return nil, fmt.Errorf("LANDLY_AUTH_JWT_SECRET must be at least 32 characters (got %d)", len(cfg.Auth.JWT.Secret))
	} else if strings.EqualFold(cfg.App.Env, "production") && strings.Contains(cfg.Auth.JWT.Secret, "dev-secret") {
		// Critical error: dev-secret in production
		return nil, fmt.Errorf("LANDLY_AUTH_JWT_SECRET contains 'dev-secret' which is not allowed in production")
	}

	// Apply defaults for TTL
	if cfg.Auth.JWT.AccessTokenTTL <= 0 {
		cfg.Auth.JWT.AccessTokenTTL = 15 * time.Minute
	}
	if cfg.Auth.JWT.RefreshTokenTTL <= 0 {
		cfg.Auth.JWT.RefreshTokenTTL = 7 * 24 * time.Hour
	}

	// Database - required
	if cfg.Database.Postgres.Host == "" {
		missingEnvVars = append(missingEnvVars, "LANDLY_DATABASE_POSTGRES_HOST")
	}

	// S3 Bucket - required
	if cfg.Storage.S3.Bucket == "" {
		missingEnvVars = append(missingEnvVars, "LANDLY_STORAGE_S3_BUCKET")
	}

	return missingEnvVars, nil
}

// Deprecated: legacy methods for backward compatibility
// TODO: remove after migrating all code

func (c *Config) GetHTTPAddr() string {
	return c.Server.HTTP.Addr
}

func (c *Config) GetJWTSecret() string {
	return c.Auth.JWT.Secret
}

func (c *Config) GetPostgresDSN() string {
	return c.Database.Postgres.DSN
}

func (c *Config) GetRedisAddr() string {
	return c.Database.Redis.Addr
}

func (c *Config) GetS3Endpoint() string {
	return c.Storage.S3.Endpoint
}

func (c *Config) GetS3AccessKey() string {
	return c.Storage.S3.AccessKey
}

func (c *Config) GetS3SecretKey() string {
	return c.Storage.S3.SecretKey
}

func (c *Config) GetS3Bucket() string {
	return c.Storage.S3.Bucket
}

func (c *Config) GetS3UseSSL() bool {
	return c.Storage.S3.UseSSL
}

func (c *Config) GetCDNBase() string {
	return c.Storage.CDN.BaseURL
}

func (c *Config) GetTmpDir() string {
	return c.Render.TmpDir
}

func (c *Config) GetAIProvider() string {
	return c.AI.Provider
}
