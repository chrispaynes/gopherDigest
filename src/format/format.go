package format

import "strings"

// DelimitedString is a
type DelimitedString struct {
	Prefix    string
	Delimiter string
	Suffix    string
}

// DelimitedCollection is
type DelimitedCollection struct {
	Collection []DelimitedString
	Delimiter  string
}

// NewDelimitedString does
func NewDelimitedString(p, d, s string) DelimitedString {
	return DelimitedString{Prefix: p, Suffix: s, Delimiter: d}
}

// Join concatenates a delimitedString
func (ds DelimitedString) Join() string {
	return strings.Join([]string{ds.Prefix, ds.Suffix}, ds.Delimiter)
}

// Titlecase titlecases a DelimitedString field
func (ds DelimitedString) Titlecase() string {
	return strings.Title(strings.ToLower(ds.Suffix))
}

// SplitToTitlecase does
func SplitToTitlecase(p int, tj TitlecaseJoiner) string {
	var str string

	if _, ok := tj.(DelimitedString); ok {
		str = tj.Titlecase()
	}

	return str
}

// TitlecaseJoiner is the interface implemented by delimited string values
type TitlecaseJoiner interface {
	Titlecase() string
	Join() string
}
