package format

import (
	"strings"
)

// FormattedStringSlice is a custom type that implements the flag.Value interface
// which creates a nice formatted string representation of a slice of strings.
type FormattedStringSlice []string

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

func (f *FormattedStringSlice) Set(value string) error {
	*f = strings.Split(value, ",")
	return nil
}

func (f *FormattedStringSlice) Type() string {
	return "strings"
}
