package errors

import (
	"net/http"
)

func inspect(err error) (code int, name string, desc string) {
	code = Code(err)
	name = errorName(err)
	if name == "" {
		name = http.StatusText(code)
		if name == "" {
			name = "unknown"
		}
	}
	desc = errorDesc(err)
	if desc == "" {
		desc = err.Error()
	}
	return code, name, desc
}

func Safe(err error) (*RichError, error) {
	if err == nil {
		return nil, nil
	}

	code, name, desc := inspect(err)
	if code == 0 {
		return nil, err
	}

	var causedBy *RichError
	var unsafe error
	if inner := Unwrap(err); inner != nil {
		causedBy, unsafe = Safe(inner)
	}

	return &RichError{
		Name:     name,
		Code:     code,
		Desc:     desc,
		Link:     errorLink(err),
		Data:     errorData(err),
		CausedBy: causedBy,
	}, unsafe
}

type NameError interface {
	error
	ErrorName() string
}

func errorName(err error) string {
	if c, ok := err.(NameError); ok {
		return c.ErrorName()
	}
	return ""
}

type DescError interface {
	error
	ErrorDescription() string
}

func errorDesc(err error) string {
	if c, ok := err.(DescError); ok {
		return c.ErrorDescription()
	}
	return ""
}

type LinkError interface {
	error
	ErrorLink() string
}

func errorLink(err error) string {
	if c, ok := err.(LinkError); ok {
		return c.ErrorLink()
	}
	return ""
}

type CodeError interface {
	error
	ErrorCode() int
}

// Code returns the status code of the CodeError, if any.
// It returns 0 otherwise.
func Code(err error) int {
	if c, ok := err.(CodeError); ok {
		return c.ErrorCode()
	}
	return 0
}

type DataError interface {
	error
	ErrorData() interface{}
}

func errorData(err error) interface{} {
	if c, ok := err.(DataError); ok {
		return c.ErrorData()
	}
	return nil
}
