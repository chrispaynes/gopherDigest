package format

import (
	"strings"
)

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

// NewDelimitedCollection creates a collection of string with common prefix, delimiter and variable collection of suffixes
func NewDelimitedCollection(prefix, delimiter string, suffixColl []string) DelimitedCollection {
	dsc := DelimitedCollection{Delimiter: delimiter}

	for _, suffix := range suffixColl {
		dsc.Collection = append(dsc.Collection, NewDelimitedString(prefix, delimiter, suffix))
	}

	return dsc
}

// IndexOfStrings returns the index of a string within a slice or -1 if it does not exist
func IndexOfString(dbname string, collection []string) int {
	for i := range collection {
		if collection[i] == dbname {
			return i
		}
	}
	return -1
}
