package errors

import "errors"

type publicError struct {
	err error
	msg string
}

func Public(err error, msg string) error {
	return publicError{err, msg}
}

func (pe publicError) Error() string {
	return pe.err.Error()
}
func (pe publicError) Public() string {
	return pe.msg
}
func (pe publicError) Unwarp() error {
	return pe.err
}

var (
	As = errors.As
	Is = errors.Is
)

// 这个模块主要用于区分面向内部的错误err和面向用户的错误msg
