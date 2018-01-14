package format

import (
	"strings"
)

// DelimitedString is a destructured Prefix + Delimiter + Suffix string
type DelimitedString struct {
	Prefix    string
	Delimiter string
	Suffix    string
}

// DelimitedCollection is a collection of DelimitedStrings
type DelimitedCollection struct {
	Collection []DelimitedString
	Delimiter  string
}

// NewDelimitedString creates a new DelimitedString struct
func NewDelimitedString(p, d, s string) DelimitedString {
	return DelimitedString{Prefix: p, Suffix: s, Delimiter: d}
}

// Join concatenates a DelimitedString
func (ds DelimitedString) Join() string {
	return strings.Join([]string{ds.Prefix, ds.Suffix}, ds.Delimiter)
}

// Titlecase titlecases a delimited string's prefix and suffix
func (ds DelimitedString) Titlecase() string {
	return strings.Title(strings.ToLower(ds.Prefix)) +
		ds.Delimiter + strings.Title(strings.ToLower(ds.Suffix))
}

// NewDelimitedCollection creates a collection of a string with common prefix, delimiter and variable number of suffixes
func NewDelimitedCollection(prefix, delimiter string, suffixColl []string) DelimitedCollection {
	dsc := DelimitedCollection{Delimiter: delimiter}

	for _, suffix := range suffixColl {
		dsc.Collection = append(dsc.Collection, NewDelimitedString(prefix, delimiter, suffix))
	}

	return dsc
}

// IndexOfString returns the index of a string within a slice or -1 if the string does not exist
func IndexOfString(dbname string, collection []string) int {
	for i := range collection {
		if collection[i] == dbname {
			return i
		}
	}
	return -1
}
