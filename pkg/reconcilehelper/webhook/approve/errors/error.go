package errors

import (
	"fmt"
	"net/http"
)

type Error struct {
	t   string
	msg string
}

const (
	ErrTypeBadRequest          = "badRequest"
	ErrTypeUnauthenticated     = "unauthenticated"
	ErrTypeForbidden           = "forbidden"
	ErrTypeInternalServerError = "internalServerError"
)

func (e *Error) Error() string {
	return fmt.Sprintf("error [%s] : [%s]", e.t, e.msg)
}

func (e *Error) HTTPStatusCode() int {
	if e == nil {
		return http.StatusOK
	}
	switch e.t {
	case ErrTypeBadRequest:
		return http.StatusBadRequest
	case ErrTypeUnauthenticated:
		return http.StatusUnauthorized
	case ErrTypeForbidden:
		return http.StatusForbidden
	default:
		return http.StatusInternalServerError
	}
}

func BadRequest(msg string) *Error {
	return &Error{
		t:   ErrTypeBadRequest,
		msg: msg,
	}
}

func Unauthenticated(msg string) *Error {
	return &Error{
		t:   ErrTypeUnauthenticated,
		msg: msg,
	}
}

func Forbidden(msg string) *Error {
	return &Error{
		t:   ErrTypeForbidden,
		msg: msg,
	}
}

func InternalServerError(msg string) *Error {
	return &Error{
		t:   ErrTypeInternalServerError,
		msg: msg,
	}
}
