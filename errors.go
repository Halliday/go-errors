package errors

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
)

// New returns an error that formats as the given text.
func New(format string, a ...interface{}) error {
	return fmt.Errorf(format, a...)
}

// type StatusError struct {
// 	Text   string
// 	Status int
// }

// type codeError struct {
// 	Message string `json:"msg"`
// 	Code    int    `json:"code"`
// }

// func (err codeError) Error() string {
// 	return err.Message
// }

// func (err codeError) ErrorCode() int {
// 	return err.Code
// }

// func (err StatusError) Error() string {
// 	return err.Text
// }

func NewCode(code int, format string, a ...interface{}) error {
	return &RichError{
		Code: code,
		Desc: fmt.Sprintf(format, a...),
	}
}

// // Status returns an error with a given text and a status code.
// func Status(s int, format string, a ...interface{}) error {
// 	return StatusError{
// 		Status: s,
// 		Text:   fmt.Sprintf(format, a...),
// 	}
// }

func Unwrap(err error) error {
	return errors.Unwrap(err)
}

// ErrBadMethod is a simple "400 Bad Request Method" Error.
var ErrBadMethod = NewCode(400, "bad request method")

// ErrUnauthorized is a simple "401 Unauthorized" Error.
var ErrUnauthorized = NewCode(401, "unauthorized")

// ErrNotFound is a simple "404 Not Found" Error.
var ErrNotFound = NewCode(404, "not found")

// ErrInternal is a simple "500 Internal Server Error".
var ErrInternal = NewCode(500, "internal server error")

type Multi []error

func (multi Multi) Error() string {
	var builder strings.Builder
	for i, err := range multi {
		if i != 0 {
			builder.WriteByte('\n')
		}
		builder.WriteString(err.Error())
	}
	return builder.String()
}

func (multi Multi) Reduce() error {
	if len(multi) == 0 {
		return nil
	}
	if len(multi) == 1 {
		return multi[0]
	}
	return multi
}

func (multi *Multi) Append(errs ...error) {
	*multi = append(*multi, errs...)
}

func JoinNew(err error, format string, a ...interface{}) error {
	return Join(err, fmt.Errorf(format, a...))
}

func Join(errs ...error) error {
	l := 0
	n := 0
	o := 0
	for i, err := range errs {
		if err != nil {
			if multi, ok := err.(Multi); ok {
				l += len(multi)
			} else {
				l++
			}
			n++
			o = i
		}
	}
	if l == 0 {
		return nil
	}
	if n == 1 {
		return errs[o]
	}
	multi := make(Multi, 0, l)
	for _, err := range errs {
		if err != nil {
			if m, ok := err.(Multi); ok {
				multi = append(multi, m...)
			} else {
				multi = append(multi, err)
			}
		}
	}
	return multi
}

type wrapped struct {
	Err      error
	CausedBy error
}

func Wrap(err error, causedBy error) error {
	if causedBy == nil {
		causedBy = Unwrap(err)
	}
	return &wrapped{
		Err:      err,
		CausedBy: causedBy,
	}
}

func (err *wrapped) Error() string {
	return err.Err.Error()
}

func (err *wrapped) ErrorName() string {
	name, _ := ErrorName(err.Err)
	return name
}

func (err *wrapped) ErrorCode() int {
	code := ErrorCode(err.Err)
	if code != 0 {
		return code
	}
	return ErrorCode(err.CausedBy)
}

func (err *wrapped) ErrorDescription() string {
	_, desc := ErrorName(err.Err)
	return desc
}

func (err *wrapped) Unwrap() error {
	return err.CausedBy
}

func ErrorCode(err error) int {
	if c, ok := err.(interface {
		ErrorCode() int
	}); ok {
		return c.ErrorCode()
	}
	err = Unwrap(err)
	if err == nil {
		return 0
	}
	return ErrorCode(err)
}

func ErrorName(err error) (name string, description string) {
	if c, ok := err.(interface {
		ErrorName() string
		ErrorDescription() string
	}); ok {
		return c.ErrorName(), c.ErrorDescription()
	}

	if code := ErrorCode(err); code != 0 {
		name = http.StatusText(code)
		if name == "" {
			name = "unknown"
		}
		description = err.Error()
		return name, description
	}
	return "", ""
}

func Stack(err error) []error {
	s := stack(err)
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
	return s
}

func stack(err error) []error {
	if err == nil {
		return make([]error, 0)
	}
	return append(stack(Unwrap(err)), err)
}
