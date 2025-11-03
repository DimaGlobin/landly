package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config структура конфигурации приложения
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

// Load загружает конфигурацию из YAML файлов и переменных окружения
func Load() (*Config, error) {
	v := viper.New()

	// Определяем путь к конфигу (корень проекта)
	configPath := findConfigPath()

	// 1. Загружаем базовый config.yml
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(configPath)
	v.AddConfigPath(".")

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config.yml: %w", err)
	}

	// 2. Мержим config.local.yml (если существует)
	v.SetConfigName("config.local")
	if err := v.MergeInConfig(); err != nil {
		// config.local.yml опционален
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to merge config.local.yml: %w", err)
		}
	}

	// 3. Переопределяем через environment variables
	// Формат: LANDLY_DATABASE_POSTGRES_HOST
	v.SetEnvPrefix("LANDLY")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Unmarshal в структуру
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Построить DSN для Postgres
	cfg.Database.Postgres.DSN = buildPostgresDSN(&cfg.Database.Postgres)

	if cfg.App.BaseURL == "" {
		cfg.App.BaseURL = "http://localhost:8080"
	}

	// Валидация
	if err := validateConfig(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// findConfigPath ищет путь к корню проекта с config.yml
func findConfigPath() string {
	// Проверяем текущую директорию
	if _, err := os.Stat("config.yml"); err == nil {
		return "."
	}

	// Проверяем родительские директории (для запуска из apps/backend)
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

// buildPostgresDSN создаёт DSN строку для PostgreSQL
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

// validateConfig проверяет обязательные поля
func validateConfig(cfg *Config) error {
	if cfg.Auth.JWT.Secret == "" {
		return fmt.Errorf("auth.jwt.secret is required")
	}
	if len(cfg.Auth.JWT.Secret) < 32 {
		return fmt.Errorf("auth.jwt.secret must be at least 32 characters")
	}
	if strings.EqualFold(cfg.App.Env, "production") && strings.Contains(cfg.Auth.JWT.Secret, "dev-secret") {
		return fmt.Errorf("auth.jwt.secret must be overridden for production")
	}

	if cfg.Auth.JWT.AccessTokenTTL <= 0 {
		cfg.Auth.JWT.AccessTokenTTL = 15 * time.Minute
	}
	if cfg.Auth.JWT.RefreshTokenTTL <= 0 {
		cfg.Auth.JWT.RefreshTokenTTL = 7 * 24 * time.Hour
	}

	if cfg.Database.Postgres.Host == "" {
		return fmt.Errorf("database.postgres.host is required")
	}

	if cfg.Storage.S3.Bucket == "" {
		return fmt.Errorf("storage.s3.bucket is required")
	}

	return nil
}

// Deprecated: старые поля для обратной совместимости
// TODO: удалить после миграции всего кода

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
