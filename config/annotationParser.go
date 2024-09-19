package config

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// filterNestedInput filters the options to only include those that start with the prefix
// and then removes the prefix from the keys.
func filterNestedInput(prefix string, options map[string]string) map[string]string {
	nestedOptions := make(map[string]string)
	prefix += "."

	for key, value := range options {
		if strings.HasPrefix(key, prefix) {
			// Strip the prefix
			trimmedKey := strings.TrimPrefix(key, prefix)
			nestedOptions[trimmedKey] = value
		}
	}

	return nestedOptions
}

// parseStructTags parses the options map and sets the values of the fields in the struct v
// Parameters:
// - tagName: the tag name to look for in the struct fields (e.g. "transformer")
// - options: the map of options to parse
// - v: the reflect.Value of the struct to set the values on (e.g. reflect.ValueOf(&Transformer{}))
func parseStructTags(tagName string, options map[string]string, v reflect.Value) error {
	v = reflect.Indirect(v)
	t := v.Type()

	var errs []error
	for key, value := range options {
		if value == "" {
			errs = append(errs, fmt.Errorf("empty value for %s", key))
			continue
		}

		found := false
		for i := 0; i < v.NumField(); i++ {
			field := t.Field(i)
			fieldValue := v.Field(i)

			tagValue, ok := field.Tag.Lookup(tagName)
			if !ok {
				// skip fields without the tagName tag
				continue
			}

			// Check if the key is exactly equal to the tag value
			if key == tagValue {
				found = true
				switch fieldValue.Kind() {
				case reflect.String:
					fieldValue.SetString(value)
				case reflect.Int, reflect.Int64:
					if fieldValue.Type() == reflect.TypeOf(time.Duration(0)) {
						durationValue, err := time.ParseDuration(value)
						if err != nil {
							errs = append(errs, fmt.Errorf("invalid value for %s: %w", key, err))
							continue
						}
						fieldValue.Set(reflect.ValueOf(durationValue))
					} else {
						intValue, err := strconv.ParseInt(value, 10, 64)
						if err != nil {
							errs = append(errs, fmt.Errorf("invalid value for %s: %w", key, err))
							continue
						}
						fieldValue.SetInt(intValue)
					}
				case reflect.Float32, reflect.Float64:
					floatValue, err := strconv.ParseFloat(value, 64)
					if err != nil {
						errs = append(errs, fmt.Errorf("invalid value for %s: %w", key, err))
						continue
					}
					fieldValue.SetFloat(floatValue)
				case reflect.Bool:
					boolValue, err := strconv.ParseBool(value)
					if err != nil {
						errs = append(errs, fmt.Errorf("invalid value for %s: %w", key, err))
						continue
					}
					fieldValue.SetBool(boolValue)
				default:
					errs = append(errs, fmt.Errorf("unsupported field type for %s", key))
				}
				break
			}

			// Handle nested structs
			if strings.HasPrefix(key, tagValue+".") && fieldValue.Kind() == reflect.Struct {
				found = true
				nestedInput := filterNestedInput(tagValue, options)
				nestedErr := parseStructTags(tagName, nestedInput, fieldValue)
				if nestedErr != nil {
					errs = append(errs, fmt.Errorf("error parsing nested struct %s: %w", field.Name, nestedErr))
				}
				break
			}
		}

		if !found {
			errs = append(errs, fmt.Errorf("unrecognized option: %s", key))
		}
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}
