package fileUtils

import "regexp"

// isValidFilePathFormat checks if the given string is formatted like a file path.
func isValidFilePathFormat(path string) bool {
	if path == "" {
		return false
	}
	if path == "/" {
		return true
	}
	var validPathPattern = `^([a-zA-Z]:\\|/)?([^<>:"/\\|?*\n]+[/\\])*([^<>:"/\\|?*\n]+)?$`
	matched, _ := regexp.MatchString(validPathPattern, path)
	return matched
}
