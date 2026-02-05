package domain

import "errors"

var (
	ErrRetriable    = errors.New("retriable error")
	ErrNonRetriable = errors.New("non-retriable error")
)

type AppError struct {
	Err        error
	Retriable  bool
	HTTPStatus int
	Message    string
}

func (e *AppError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return e.Err.Error()
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func NewValidationError(err error, message string) *AppError {
	return &AppError{
		Err:        err,
		Retriable:  false,
		HTTPStatus: 400,
		Message:    message,
	}
}

func NewRetriableError(err error, message string) *AppError {
	return &AppError{
		Err:        err,
		Retriable:  true,
		HTTPStatus: 500,
		Message:    message,
	}
}

func NewNonRetriableError(err error, message string) *AppError {
	return &AppError{
		Err:        err,
		Retriable:  false,
		HTTPStatus: 500,
		Message:    message,
	}
}

func IsRetriable(err error) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Retriable
	}
	return false
}

func HTTPStatus(err error) int {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.HTTPStatus
	}
	return 500
}
