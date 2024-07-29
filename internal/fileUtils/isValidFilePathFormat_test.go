package fileUtils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsValidFilePathFormat(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{"Valid Unix Path", "/valid/path/to/file.txt", true},
		{"Valid Windows Path", "C:\\valid\\path\\to\\file.txt", true},
		{"Invalid Path with Special Characters", "invalid<>path", false},
		{"Invalid Path with Pipe", "another/invalid|path", false},
		{"Valid Path with Dots", "valid/path/with/dots.ext", true},
		{"Mixed Slashes Path", "C:/mixed\\slashes/path.txt", false},
		{"Spaces Paths", "some random words with spaces", true},
		{"Empty Path", "", false},
		{"Root Path", "/", true},
		{"Windows Root Path", "C:\\", true},
		{"Valid Path with Spaces", "/valid/path with spaces/file.txt", true},
		{"Valid Path with Underscores", "/valid/path_with_underscores/file.txt", true},
		{"Valid Path with Hyphens", "/valid/path-with-hyphens/file.txt", true},
		{"Invalid Path with Question Mark", "invalid?path", false},
		{"Invalid Path with Asterisk", "invalid*path", false},
		{"Valid Path with Numbers", "/valid/path123/file.txt", true},
		{"Valid Path with Parentheses", "/valid/path(with)/file.txt", true},
		{"Valid Path with Plus Sign", "/valid/path+with+plus/file.txt", true},
		{"Invalid Path with Colon", "invalid:path", false},
		{"Invalid Path with Double Quote", "invalid\"path", false},
		{"Http URI", "http://example.com", false},
		{"Https URI", "https://example.com", false},
		{"gRPC URI", "grpc://example.com", false},
		{"file URI", "file://example.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidFilePathFormat(tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}
