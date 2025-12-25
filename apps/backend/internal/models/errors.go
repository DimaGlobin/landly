package domain

import "fmt"

// Error базовый тип для доменных ошибок
type Error struct {
	Code    string
	Message string
	Err     error
}

func (e *Error) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *Error) Unwrap() error {
	return e.Err
}

func (e *Error) HTTPStatus() int {
	switch e.Code {
	case "NOT_FOUND", "PROJECT_NOT_FOUND", "USER_NOT_FOUND":
		return 404
	case "ALREADY_EXISTS", "CONFLICT", "USER_ALREADY_EXISTS":
		return 409
	case "INVALID_INPUT", "BAD_REQUEST", "VALIDATION_ERROR":
		return 400
	case "UNAUTHORIZED", "INVALID_TOKEN", "TOKEN_EXPIRED", "INVALID_CREDENTIALS":
		return 401
	case "FORBIDDEN":
		return 403
	case "INTERNAL_ERROR", "GENERATION_FAILED", "RENDER_FAILED", "PUBLISH_FAILED":
		return 500
	default:
		return 500
	}
}

func (e *Error) WithMessage(msg string) *Error {
	return &Error{
		Code:    e.Code,
		Message: msg,
		Err:     e.Err,
	}
}

func (e *Error) WithError(err error) *Error {
	return &Error{
		Code:    e.Code,
		Message: e.Message,
		Err:     err,
	}
}

// Predefined errors with standardized codes for i18n on frontend
var (
	// Authentication errors
	ErrUnauthorized = &Error{
		Code:    "UNAUTHORIZED",
		Message: "Authorization required",
	}

	ErrInvalidToken = &Error{
		Code:    "INVALID_TOKEN",
		Message: "Invalid or malformed token",
	}

	ErrTokenExpired = &Error{
		Code:    "TOKEN_EXPIRED",
		Message: "Token has expired",
	}

	ErrInvalidCredentials = &Error{
		Code:    "INVALID_CREDENTIALS",
		Message: "Invalid email or password",
	}

	ErrForbidden = &Error{
		Code:    "FORBIDDEN",
		Message: "Access denied",
	}

	// Resource errors
	ErrNotFound = &Error{
		Code:    "NOT_FOUND",
		Message: "Resource not found",
	}

	ErrProjectNotFound = &Error{
		Code:    "PROJECT_NOT_FOUND",
		Message: "Project not found",
	}

	ErrUserNotFound = &Error{
		Code:    "USER_NOT_FOUND",
		Message: "User not found",
	}

	ErrAlreadyExists = &Error{
		Code:    "ALREADY_EXISTS",
		Message: "Resource already exists",
	}

	ErrUserAlreadyExists = &Error{
		Code:    "USER_ALREADY_EXISTS",
		Message: "User with this email already exists",
	}

	ErrConflict = &Error{
		Code:    "CONFLICT",
		Message: "Resource conflict",
	}

	// Validation errors
	ErrValidationError = &Error{
		Code:    "VALIDATION_ERROR",
		Message: "Validation error",
	}

	ErrInvalidInput = &Error{
		Code:    "INVALID_INPUT",
		Message: "Invalid input",
	}

	ErrBadRequest = &Error{
		Code:    "BAD_REQUEST",
		Message: "Bad request",
	}

	// Operation errors
	ErrInternal = &Error{
		Code:    "INTERNAL_ERROR",
		Message: "Internal server error",
	}

	ErrGenerationFailed = &Error{
		Code:    "GENERATION_FAILED",
		Message: "AI generation failed",
	}

	ErrRenderFailed = &Error{
		Code:    "RENDER_FAILED",
		Message: "Rendering failed",
	}

	ErrPublishFailed = &Error{
		Code:    "PUBLISH_FAILED",
		Message: "Publishing failed",
	}
)
