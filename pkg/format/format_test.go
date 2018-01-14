package format

import (
	"reflect"
	"testing"
)

func TestJoin(t *testing.T) {
	tt := []struct {
		name     string
		ds       DelimitedString
		expected string
	}{
		{"Underscore", DelimitedString{Prefix: "ALPHA", Delimiter: "_", Suffix: "CASE"}, "ALPHA_CASE"},
		{"Hyphen", DelimitedString{Prefix: "ALPHA", Delimiter: "-", Suffix: "CASE"}, "ALPHA-CASE"},
		{"Whitespace", DelimitedString{Prefix: "ALPHA", Delimiter: " ", Suffix: "CASE"}, "ALPHA CASE"},
		{"Colon", DelimitedString{Prefix: "ALPHA", Delimiter: ":", Suffix: "CASE"}, "ALPHA:CASE"},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			actual := tc.ds.Join()

			if actual != tc.expected {
				t.Fatalf("join of %v should be %v; got %v", tc.name, tc.expected, actual)
			}
		})
	}
}

func TestTitleCase(t *testing.T) {
	tt := []struct {
		name     string
		ds       DelimitedString
		expected string
	}{
		{"Uppercase", DelimitedString{Prefix: "BRAVO", Delimiter: "_", Suffix: "TEST"}, "Bravo_Test"},
		{"Lowercase", DelimitedString{Prefix: "bravo", Delimiter: "_", Suffix: "test"}, "Bravo_Test"},
		{"Mixedcase", DelimitedString{Prefix: "BrAvO", Delimiter: "_", Suffix: "tEsT"}, "Bravo_Test"},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			actual := tc.ds.Titlecase()

			if actual != tc.expected {
				t.Fatalf("titlecase of %v should be %v, but got %v", tc.name, tc.expected, actual)
			}
		})
	}
}

func TestIndexOfString(t *testing.T) {
	tt := []struct {
		name       string
		word       string
		collection []string
		expected   int
	}{
		{"Single Letter", "B", []string{"A", "B", "C"}, 1},
		{"Multiple Whitespace Strings", " ", []string{" ", "A", "B", " ", "C"}, 0},
		{"Missing String", "Gopher", []string{"A", "B", " ", "C"}, -1},
		{"Single Word", "Gopher", []string{"Alpha", "Bravo", "Gopher", "Charlie"}, 2},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			actual := IndexOfString(tc.word, tc.collection)

			if actual != tc.expected {
				t.Errorf("indexOfString of %v should be %v, but got %v", tc.name, tc.expected, actual)
			}
		})
	}
}

func TestNewDelimitedString(t *testing.T) {
	tt := []struct {
		name      string
		prefix    string
		delimiter string
		suffix    string
		expected  DelimitedString
	}{
		{"AlphaPrefix, _, AlphaSuffix", "AlphaPrefix", "_", "AlphaSuffix", DelimitedString{Prefix: "AlphaPrefix", Delimiter: "_", Suffix: "AlphaSuffix"}},
		{"AlphaPrefix, <space>, AlphaSuffix", "AlphaPrefix", " ", "AlphaSuffix", DelimitedString{Prefix: "AlphaPrefix", Delimiter: " ", Suffix: "AlphaSuffix"}},
		{"<space>, _, AlphaSuffix", " ", "_", "AlphaSuffix", DelimitedString{Prefix: " ", Delimiter: "_", Suffix: "AlphaSuffix"}},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			actual := NewDelimitedString(tc.prefix, tc.delimiter, tc.suffix)

			if !reflect.DeepEqual(actual, tc.expected) {
				t.Errorf("newDelimitedString of %v should be %v, but got %v", tc.name, tc.expected, actual)
			}
		})
	}
}

func TestNewDelimitedCollection(t *testing.T) {
	tt := []struct {
		name      string
		prefix    string
		delimiter string
		suffix    []string
		expected  DelimitedCollection
	}{
		{"AlphaPrefix, _, {AlphaSuffix, BravoSuffix, CharlieSuffix}",
			"AlphaPrefix", "_", []string{"AlphaSuffix", "BravoSuffix", "CharlieSuffix"},
			DelimitedCollection{
				Collection: []DelimitedString{
					{Prefix: "AlphaPrefix", Delimiter: "_", Suffix: "AlphaSuffix"},
					{Prefix: "AlphaPrefix", Delimiter: "_", Suffix: "BravoSuffix"},
					{Prefix: "AlphaPrefix", Delimiter: "_", Suffix: "CharlieSuffix"},
				},
				Delimiter: "_"},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			actual := NewDelimitedCollection(tc.prefix, tc.delimiter, tc.suffix)

			if !reflect.DeepEqual(actual, tc.expected) {
				t.Errorf("newDelimitedString of %v should be %v, but got %v", tc.name, tc.expected, actual)
			}
		})
	}
}
