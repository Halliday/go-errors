package errors

import (
	"strconv"
	"strings"
)

type RichError struct {
	Name     string      `json:"name"`
	Code     int         `json:"code,omitempty"`
	Desc     string      `json:"desc,omitempty"`
	Link     string      `json:"link,omitempty"`
	Data     interface{} `json:"data,omitempty"`
	CausedBy error       `json:"causedBy,omitempty"`
}

func NewRich(name string, code int, desc string, link string, data interface{}, causedBy error) *RichError {
	return &RichError{name, code, desc, link, data, causedBy}
}

func Rich(err error) *RichError {
	if err == nil {
		return nil
	}
	if richError, ok := err.(*RichError); ok {
		return richError
	}

	code, name, desc := inspect(err)

	return &RichError{
		Name:     name,
		Code:     code,
		Desc:     desc,
		CausedBy: Rich(Unwrap(err)),
	}
}

var _ NameError = (*RichError)(nil)
var _ DescError = (*RichError)(nil)
var _ CodeError = (*RichError)(nil)
var _ LinkError = (*RichError)(nil)
var _ DataError = (*RichError)(nil)

func (r RichError) ErrorName() string {
	return r.Name
}

func (r RichError) ErrorCode() int {
	return r.Code
}

func (r RichError) ErrorDescription() string {
	return r.Desc
}

func (r RichError) ErrorLink() string {
	return r.Link
}

func (err RichError) ErrorData() interface{} {
	return err.Data
}

func (r RichError) Unwrap() error {
	if r.CausedBy == nil {
		return nil
	}
	return r.CausedBy
}

func (r RichError) Error() string {
	var b strings.Builder
	b.WriteString(strconv.Itoa(r.Code))
	if r.Name != "" {
		b.WriteByte(' ')
		b.WriteString(r.Name)
	}
	if r.Desc != "" {
		b.WriteByte(' ')
		b.WriteString(r.Desc)
	}
	if r.CausedBy != nil {
		b.WriteString(" (")
		b.WriteString(r.CausedBy.Error())
		b.WriteByte(')')
	}
	return b.String()
}
