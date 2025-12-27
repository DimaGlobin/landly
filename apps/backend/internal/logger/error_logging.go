package logger

import (
	"context"
	"fmt"

	domain "github.com/landly/backend/internal/models"
	"go.uber.org/zap"
)

// LogError логирует ошибку с полным контекстом для внутреннего использования
// Используется для логирования всех ошибок, включая внутренние детали
func LogError(ctx context.Context, msg string, err error, fields ...zap.Field) {
	log := WithContext(ctx)
	
	allFields := []zap.Field{
		zap.String("error_type", fmt.Sprintf("%T", err)),
		zap.Error(err),
		zap.String("error_message", err.Error()),
	}
	
	// Добавляем детали для domain.Error
	if domainErr, ok := err.(*domain.Error); ok {
		allFields = append(allFields,
			zap.String("error_code", domainErr.Code),
			zap.String("domain_message", domainErr.Message),
			zap.Int("http_status", domainErr.HTTPStatus()),
		)
		// Если есть вложенная ошибка, логируем её тоже
		if domainErr.Err != nil {
			allFields = append(allFields,
				zap.String("wrapped_error_type", fmt.Sprintf("%T", domainErr.Err)),
				zap.String("wrapped_error_message", domainErr.Err.Error()),
			)
		}
	}
	
	allFields = append(allFields, fields...)
	log.Error(msg, allFields...)
}

// LogInternalError логирует внутреннюю ошибку с максимальным контекстом
// Используется для ошибок, которые не должны попадать клиенту
func LogInternalError(ctx context.Context, msg string, err error, fields ...zap.Field) {
	log := WithContext(ctx)
	
	allFields := []zap.Field{
		zap.String("error_type", fmt.Sprintf("%T", err)),
		zap.Error(err),
		zap.String("error_message", err.Error()),
		zap.Bool("internal_error", true),
	}
	
	allFields = append(allFields, fields...)
	log.Error(msg, allFields...)
}

// LogDomainError логирует доменную ошибку с контекстом
// Используется для ошибок, которые могут быть отправлены клиенту
func LogDomainError(ctx context.Context, msg string, err *domain.Error, fields ...zap.Field) {
	log := WithContext(ctx)
	
	allFields := []zap.Field{
		zap.String("error_code", err.Code),
		zap.String("error_message", err.Message),
		zap.Int("http_status", err.HTTPStatus()),
	}
	
	// Если есть вложенная ошибка, логируем её для внутреннего использования
	if err.Err != nil {
		allFields = append(allFields,
			zap.String("wrapped_error_type", fmt.Sprintf("%T", err.Err)),
			zap.String("wrapped_error_message", err.Err.Error()),
			zap.Error(err.Err),
		)
	}
	
	allFields = append(allFields, fields...)
	log.Error(msg, allFields...)
}

// LogSQLError логирует SQL ошибку с максимальным контекстом
// Используется в repositories для логирования SQL ошибок
func LogSQLError(ctx context.Context, operation string, err error, sqlQuery string, params map[string]interface{}, fields ...zap.Field) {
	log := WithContext(ctx)
	
	allFields := []zap.Field{
		zap.String("operation", operation),
		zap.String("error_type", fmt.Sprintf("%T", err)),
		zap.Error(err),
		zap.String("error_message", err.Error()),
		zap.String("sql_query", sqlQuery),
		zap.Bool("sql_error", true),
	}
	
	if len(params) > 0 {
		allFields = append(allFields, zap.Any("sql_params", params))
	}
	
	allFields = append(allFields, fields...)
	log.Error("SQL error", allFields...)
}

