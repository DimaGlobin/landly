package logger

import (
	"context"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type loggerContextKey struct{}

// Config holds logger configuration
type Config struct {
	Service string
	Version string
	Env     string
}

// Logger интерфейс для логгирования
type Logger interface {
	Debug(msg string, fields ...zap.Field)
	Info(msg string, fields ...zap.Field)
	Warn(msg string, fields ...zap.Field)
	Error(msg string, fields ...zap.Field)
	Fatal(msg string, fields ...zap.Field)
	With(fields ...zap.Field) Logger
	WithContext(ctx context.Context) Logger
	Sync() error
	GetZapLogger() *zap.Logger
}

// logger реализация логгера
type logger struct {
	zap *zap.Logger
}

// global config for logger
var globalConfig = Config{
	Service: "landly-backend",
	Version: "unknown",
	Env:     "production",
}

// SetConfig sets global logger configuration
func SetConfig(cfg Config) {
	globalConfig = cfg
}

// New создает новый логгер
func New() Logger {
	return NewWithConfig(globalConfig)
}

// NewWithConfig создает новый логгер с конфигурацией
func NewWithConfig(cfg Config) Logger {
	var zapConfig zap.Config

	env := cfg.Env
	if env == "" {
		env = os.Getenv("LANDLY_APP_ENV")
		if env == "" {
			env = os.Getenv("ENV")
		}
	}

	if env == "development" {
		zapConfig = zap.NewDevelopmentConfig()
		zapConfig.EncoderConfig.TimeKey = "timestamp"
		zapConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	} else {
		zapConfig = zap.NewProductionConfig()
		zapConfig.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
		// Ensure consistent timestamp format in production
		zapConfig.EncoderConfig.TimeKey = "timestamp"
		zapConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	}

	zapLogger, err := zapConfig.Build()
	if err != nil {
		panic("failed to initialize logger: " + err.Error())
	}

	// Add service metadata fields
	zapLogger = zapLogger.With(
		zap.String("service", cfg.Service),
		zap.String("version", cfg.Version),
		zap.String("env", env),
	)

	return &logger{zap: zapLogger}
}

// Debug логирует сообщение уровня DEBUG
func (l *logger) Debug(msg string, fields ...zap.Field) {
	l.zap.Debug(msg, fields...)
}

// Info логирует сообщение уровня INFO
func (l *logger) Info(msg string, fields ...zap.Field) {
	l.zap.Info(msg, fields...)
}

// Warn логирует сообщение уровня WARN
func (l *logger) Warn(msg string, fields ...zap.Field) {
	l.zap.Warn(msg, fields...)
}

// Error логирует сообщение уровня ERROR
func (l *logger) Error(msg string, fields ...zap.Field) {
	l.zap.Error(msg, fields...)
}

// Fatal логирует сообщение уровня FATAL и завершает программу
func (l *logger) Fatal(msg string, fields ...zap.Field) {
	l.zap.Fatal(msg, fields...)
}

// With создает новый логгер с дополнительными полями
func (l *logger) With(fields ...zap.Field) Logger {
	return &logger{zap: l.zap.With(fields...)}
}

// WithContext создает новый логгер с контекстом
func (l *logger) WithContext(ctx context.Context) Logger {
	fields := []zap.Field{}

	if traceID := ctx.Value(traceIDContextKey); traceID != nil {
		if id, ok := traceID.(string); ok {
			fields = append(fields, zap.String("trace_id", id))
		}
	}

	if userID := ctx.Value(userIDContextKey); userID != nil {
		if id, ok := userID.(string); ok {
			fields = append(fields, zap.String("user_id", id))
		}
	}

	loggerFromContext := ctx.Value(loggerContextKey{})
	if loggerFromContext != nil {
		if lgr, ok := loggerFromContext.(Logger); ok {
			return lgr.With(fields...)
		}
	}

	if len(fields) > 0 {
		return &logger{zap: l.zap.With(fields...)}
	}

	return l
}

// Sync синхронизирует логгер
func (l *logger) Sync() error {
	return l.zap.Sync()
}

// GetZapLogger возвращает zap.Logger для совместимости
func (l *logger) GetZapLogger() *zap.Logger {
	return l.zap
}

// GetZapLogger возвращает zap.Logger из глобального логгера
func GetZapLogger() *zap.Logger {
	return Get().GetZapLogger()
}

// Глобальный логгер
var global Logger

// Init инициализирует глобальный логгер
func Init() {
	global = New()
}

// Get возвращает глобальный логгер
func Get() Logger {
	if global == nil {
		global = New()
	}
	return global
}

// Debug логирует сообщение уровня DEBUG через глобальный логгер
func Debug(msg string, fields ...zap.Field) {
	Get().Debug(msg, fields...)
}

// Info логирует сообщение уровня INFO через глобальный логгер
func Info(msg string, fields ...zap.Field) {
	Get().Info(msg, fields...)
}

// Warn логирует сообщение уровня WARN через глобальный логгер
func Warn(msg string, fields ...zap.Field) {
	Get().Warn(msg, fields...)
}

// Error логирует сообщение уровня ERROR через глобальный логгер
func Error(msg string, fields ...zap.Field) {
	Get().Error(msg, fields...)
}

// Fatal логирует сообщение уровня FATAL через глобальный логгер
func Fatal(msg string, fields ...zap.Field) {
	Get().Fatal(msg, fields...)
}

// With создает новый логгер с дополнительными полями через глобальный логгер
func With(fields ...zap.Field) Logger {
	return Get().With(fields...)
}

// WithContext создает новый логгер с контекстом через глобальный логгер
func WithContext(ctx context.Context) Logger {
	return Get().WithContext(ctx)
}

// IntoContext attaches logger to context
func IntoContext(ctx context.Context, log Logger) context.Context {
	return context.WithValue(ctx, loggerContextKey{}, log)
}

// FromContext retrieves logger from context if present
func FromContext(ctx context.Context) Logger {
	if log := ctx.Value(loggerContextKey{}); log != nil {
		if lgr, ok := log.(Logger); ok {
			return lgr
		}
	}
	return Get()
}
