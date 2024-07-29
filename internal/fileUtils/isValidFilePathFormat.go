package fileUtils

import "regexp"

// isValidFilePathFormat checks if the given string is formatted like a file path.
func isValidFilePathFormat1(path string) bool {
	// Define a regular expression pattern for file paths
	var filePathPattern = `^([a-zA-Z]:\\|/)?([^<>:"/\\|?*\n]+/)*([^<>:"/\\|?*\n]+)$`
	re := regexp.MustCompile(filePathPattern)
	return re.MatchString(path)
}

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
