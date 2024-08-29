package format

import (
	"strings"
)

// FormattedStringSlice is a custom type that implements the flag.Value interface
// which creates a nice formatted string representation of a slice of strings.
type FormattedStringSlice []string

// String formats the slice of strings into a wrapped string useful for printing to the console.
func (f *FormattedStringSlice) String() string {
	var sb strings.Builder
	sb.WriteString("[")
	lineLength := 1
	for i, item := range *f {
		itemLength := len(item) + 2 // account for ", "
		if lineLength+itemLength > 70 {
			sb.WriteString("\n\t")
			lineLength = 1
		}
		sb.WriteString(item)
		if i < len(*f)-1 {
			sb.WriteString(", ")
			lineLength += itemLength
		}
	}
	sb.WriteString("]")
	return sb.String()
}

// Set splits the input string by commas into a slice of strings and assigns it to the FormattedStringSlice.
func (f *FormattedStringSlice) Set(value string) error {
	*f = strings.Split(value, ",")
	return nil
}

// Type returns the type of the flag value, which is needed by the flag.Value interface.
func (f *FormattedStringSlice) Type() string {
	return "strings"
}
