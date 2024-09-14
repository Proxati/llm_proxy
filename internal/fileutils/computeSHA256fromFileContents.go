package fileutils

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
)

// ComputeSHA256FromFileContents computes the SHA-256 checksum of the file at the given path
func ComputeSHA256FromFileContents(commandPath string) (string, error) {
	// Open the file
	file, err := os.Open(commandPath)
	if err != nil {
		return "", fmt.Errorf("unable to open file: %w", err)
	}
	defer file.Close()

	// Create a new SHA-256 hash
	hash := sha256.New()

	// Copy the file's contents into the hash
	if _, err := io.Copy(hash, file); err != nil {
		return "", fmt.Errorf("unable to compute hash: %w", err)
	}

	// Get the final hash sum
	sum := hash.Sum(nil)

	// Convert the hash sum to a hexadecimal string
	return hex.EncodeToString(sum), nil
}
