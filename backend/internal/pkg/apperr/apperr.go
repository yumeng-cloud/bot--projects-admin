package apperr

import "net/http"

type AppError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	StatusCode int    `json:"-"`
}

func (e *AppError) Error() string { return e.Message }

func New(status int, code, msg string) *AppError {
	return &AppError{Code: code, Message: msg, StatusCode: status}
}

func BadRequest(msg string) *AppError  { return New(http.StatusBadRequest, "bad_request", msg) }
func Unauthorized(msg string) *AppError { return New(http.StatusUnauthorized, "unauthorized", msg) }
func Forbidden(msg string) *AppError   { return New(http.StatusForbidden, "forbidden", msg) }
func NotFound(msg string) *AppError    { return New(http.StatusNotFound, "not_found", msg) }
func Conflict(msg string) *AppError    { return New(http.StatusConflict, "conflict", msg) }
func VersionConflict(msg string) *AppError {
	return New(http.StatusConflict, "version_conflict", msg)
}
func TooMany(msg string) *AppError  { return New(http.StatusTooManyRequests, "rate_limited", msg) }
func Internal(msg string) *AppError { return New(http.StatusInternalServerError, "internal", msg) }
