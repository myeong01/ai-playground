package url

import "fmt"

type Error struct {
	errType string
	msg     string
}

var (
	ErrBadRequest = Error{errType: ErrTypeBadRequest}
)

const (
	ErrTypeBadRequest = "bad request"
)

func (e Error) WithMsg(msg string) Error {
	e.msg = msg
	return e
}

func (e Error) Error() string {
	return fmt.Sprintf("{\"type\":\"%s\",\"msg\":\"%s\"}", e.errType, e.msg)
}

func IsBadRequestError(err error) bool {
	e, valid := err.(Error)
	if !valid {
		return false
	}
	return e.errType == ErrTypeBadRequest
}
