package apperror

import (
	"errors"
	"fmt"
	"net/http"
)

var (
	ErrRedisDisabled = errors.New("redis is disabled")
)

type Error struct {
	Status  int
	Message string
	Err     error
}

func (e *Error) Error() string {
	if e.Err == nil {
		return e.Message
	}
	return fmt.Sprintf("%s : %v", e.Message, e.Err)
}

func (e *Error) Unwrap() error {
	return e.Err
}

func New(status int, message string) *Error {
	return &Error{
		Status:  status,
		Message: message,
	}
}

func Wrap(err error, status int, message string) *Error {
	return &Error{
		Status:  status,
		Message: message,
		Err:     err,
	}
}

func BadRequest(message string) *Error {
	return New(http.StatusBadRequest, message)
}

func Unauthorized(message string) *Error {
	return New(http.StatusUnauthorized, message)
}

func ServiceUnavailable(message string) *Error {
	return New(http.StatusServiceUnavailable, message)
}

func Conflict(message string) *Error {
	return New(http.StatusConflict, message)
}

func Internal(message string) *Error {
	return New(http.StatusInternalServerError, message)
}
