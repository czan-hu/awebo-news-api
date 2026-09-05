package exceptions

import "net/http"

// AppError is the canonical error shape returned to clients, matching the
// `Error` schema in docs/openapi.yaml: { statusCode, message, errors? }.
type AppError struct {
	StatusCode int               `json:"statusCode"`
	Message    string            `json:"message"`
	Errors     map[string]string `json:"errors,omitempty"`
}

func (e *AppError) Error() string {
	return e.Message
}

func New(statusCode int, message string) *AppError {
	return &AppError{StatusCode: statusCode, Message: message}
}

func BadRequest(message string) *AppError {
	if message == "" {
		message = "Проверьте правильность заполнения полей."
	}
	return New(http.StatusBadRequest, message)
}

func ValidationError(message string, errs map[string]string) *AppError {
	if message == "" {
		message = "Проверьте правильность заполнения полей."
	}
	return &AppError{StatusCode: http.StatusBadRequest, Message: message, Errors: errs}
}

func Unauthorized(message string) *AppError {
	if message == "" {
		message = "Требуется вход."
	}
	return New(http.StatusUnauthorized, message)
}

func Forbidden(message string) *AppError {
	if message == "" {
		message = "Недостаточно прав."
	}
	return New(http.StatusForbidden, message)
}

func NotFound(message string) *AppError {
	if message == "" {
		message = "Не найдено."
	}
	return New(http.StatusNotFound, message)
}

func Conflict(message string) *AppError {
	if message == "" {
		message = "Конфликт данных."
	}
	return New(http.StatusConflict, message)
}

func TooManyRequests(message string) *AppError {
	if message == "" {
		message = "Слишком много запросов, попробуйте позже."
	}
	return New(http.StatusTooManyRequests, message)
}

func Internal(message string) *AppError {
	if message == "" {
		message = "Что-то пошло не так."
	}
	return New(http.StatusInternalServerError, message)
}
