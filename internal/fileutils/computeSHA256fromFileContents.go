package fileutils

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
)

const BufferSizeBytes = 4096 * 2

// ComputeSHA256FromFileContents computes the SHA-256 checksum of the file at the given path.
func ComputeSHA256FromFileContents(commandPath string) (string, error) {
	return ComputeSHA256FromFileContentsContext(context.Background(), commandPath)
}

// ComputeSHA256FromFileContents computes the SHA-256 checksum of the file at the given path with context support.
func ComputeSHA256FromFileContentsContext(ctx context.Context, filePath string) (string, error) {
	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("unable to open file: %w", err)
	}
	defer file.Close()

	// Create a new SHA-256 hash
	hash := sha256.New()
	buf := make([]byte, BufferSizeBytes)

	for {
		n, err := file.Read(buf)
		if err != nil && err != io.EOF {
			return "", fmt.Errorf("unable to read file: %w", err)
		}
		if n == 0 {
			break
		}
		if _, err := hash.Write(buf[:n]); err != nil {
			return "", fmt.Errorf("unable to write to hash: %w", err)
		}

		select {
		case <-ctx.Done():
			return "", fmt.Errorf("operation canceled: %w", ctx.Err())
		default:
		}
	}

	err = file.Close()
	if err != nil {
		return "", fmt.Errorf("unable to close file: %w", err)
	}

	// Get the final hash sum
	sum := hash.Sum(nil)

	// Convert the hash sum to a hexadecimal string
	return hex.EncodeToString(sum), nil
}
