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
	case "NOT_FOUND":
		return 404
	case "ALREADY_EXISTS", "CONFLICT":
		return 409
	case "INVALID_INPUT", "BAD_REQUEST":
		return 400
	case "UNAUTHORIZED":
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

// Предопределённые ошибки
var (
	ErrNotFound = &Error{
		Code:    "NOT_FOUND",
		Message: "resource not found",
	}

	ErrAlreadyExists = &Error{
		Code:    "ALREADY_EXISTS",
		Message: "resource already exists",
	}

	ErrConflict = &Error{
		Code:    "CONFLICT",
		Message: "resource conflict",
	}

	ErrInvalidInput = &Error{
		Code:    "INVALID_INPUT",
		Message: "invalid input",
	}

	ErrBadRequest = &Error{
		Code:    "BAD_REQUEST",
		Message: "bad request",
	}

	ErrUnauthorized = &Error{
		Code:    "UNAUTHORIZED",
		Message: "unauthorized",
	}

	ErrForbidden = &Error{
		Code:    "FORBIDDEN",
		Message: "forbidden",
	}

	ErrInternal = &Error{
		Code:    "INTERNAL_ERROR",
		Message: "internal server error",
	}

	ErrGenerationFailed = &Error{
		Code:    "GENERATION_FAILED",
		Message: "AI generation failed",
	}

	ErrRenderFailed = &Error{
		Code:    "RENDER_FAILED",
		Message: "rendering failed",
	}

	ErrPublishFailed = &Error{
		Code:    "PUBLISH_FAILED",
		Message: "publishing failed",
	}
)
