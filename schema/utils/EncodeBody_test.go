package utils

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"io"
	"testing"

	"github.com/andybalholm/brotli"
	"github.com/stretchr/testify/assert"
)

func TestParseAcceptEncoding(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		header   string
		expected map[string]float64
	}{
		{
			name:   "Single gzip encoding without quality",
			header: "gzip",
			expected: map[string]float64{
				"gzip": 1.0, // default quality
			},
		},
		{
			name:   "Multiple encodings with and without quality",
			header: "gzip, deflate;q=0.5, br;q=99",
			expected: map[string]float64{
				"gzip":    1.0, // default quality
				"deflate": 0.5,
				"br":      99.0,
			},
		},
		{
			name:     "Empty Accept-Encoding header",
			header:   "",
			expected: map[string]float64{},
		},
		{
			name:   "Invalid quality value, defaults to 1.0",
			header: "gzip;q=invalid, deflate",
			expected: map[string]float64{
				"gzip":    1.0, // default quality
				"deflate": 1.0, // default quality
			},
		},
		{
			name:   "Single br encoding without quality",
			header: "br",
			expected: map[string]float64{
				"br": 1.0, // default quality
			},
		},
		{
			name:   "Single br encoding with quality",
			header: "br;q=5",
			expected: map[string]float64{
				"br": 5.0,
			},
		},
		{
			name:   "Invalid encoding type",
			header: "invalid;q=2.0",
			expected: map[string]float64{
				"invalid": 2.0,
			},
		},
		{
			name:   "Invalid encoding and quality",
			header: "invalid;q=number",
			expected: map[string]float64{
				"invalid": 1.0,
			},
		},
		{
			name:   "Mixed valid and invalid encoding types",
			header: "gzip, invalid;q=0.5, br;q=99",
			expected: map[string]float64{
				"gzip":    1.0, // default quality
				"invalid": 0.5,
				"br":      99.0,
			},
		},
		{
			name:   "Quality value of 0",
			header: "gzip;q=0",
			expected: map[string]float64{
				"gzip": 0.0,
			},
		},
		{
			name:   "Quality value greater than 1",
			header: "gzip;q=1.5",
			expected: map[string]float64{
				"gzip": 1.5,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseAcceptEncoding(&tt.header)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestChooseEncoding(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		header   string
		expected string
	}{
		{
			name:     "Prefer br when both br, gzip, and deflate are available",
			header:   "gzip, deflate, br",
			expected: "br",
		},
		{
			name:     "Prefer gzip when both gzip and deflate are available",
			header:   "gzip, deflate",
			expected: "gzip",
		},
		{
			name:     "Choose deflate when gzip is not available",
			header:   "deflate;q=1.0",
			expected: "deflate",
		},
		{
			name:     "Return empty when no acceptable encodings are provided",
			header:   "br",
			expected: "br",
		},
		{
			name:     "Return empty when quality is 0",
			header:   "gzip;q=0, deflate;q=0",
			expected: "",
		},
		{
			name:     "Ignore encoding with quality 0 and choose available",
			header:   "gzip;q=0, deflate;q=1.0",
			expected: "deflate",
		},
		{
			name:     "Handle empty Accept-Encoding header",
			header:   "",
			expected: "",
		},
		{
			name:     "Handle invalid quality value, default to gzip",
			header:   "gzip;q=invalid, deflate;q=0.8",
			expected: "gzip",
		},
		{
			name:     "Handle invalid encoding type",
			header:   "invalid;q=1.0",
			expected: "",
		},
		{
			name:     "Mixed valid and invalid encoding types",
			header:   "gzip, invalid;q=0.5, br;q=99",
			expected: "br",
		},
		{
			name:     "Quality value greater than 1",
			header:   "gzip;q=1.5, deflate;q=0.8",
			expected: "gzip",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := chooseEncoding(&tt.header)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Helper function to decompress gzip data
func gzipDecompress(data []byte) ([]byte, error) {
	reader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	decompressed, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	return decompressed, nil
}

func TestGzipCompress(t *testing.T) {
	tests := []struct {
		name    string
		input   []byte
		wantErr bool
	}{
		{
			name:    "Non-empty input",
			input:   []byte("Hello, world!"),
			wantErr: false,
		},
		{
			name:    "Empty input",
			input:   []byte(""),
			wantErr: false,
		},
		{
			name:    "Space case",
			input:   []byte(" "),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, encoding, err := gzipCompress(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("gzipCompress() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Decompress the output for validation
			decompressed, err := gzipDecompress(got)
			assert.NoError(t, err)
			assert.Equal(t, tt.input, decompressed)
			assert.Equal(t, "gzip", encoding)
		})
	}
}

// Helper function to decompress deflate data
func flateDecompress(data []byte) ([]byte, error) {
	reader := flate.NewReader(bytes.NewReader(data))
	defer reader.Close()
	decompressed, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	return decompressed, nil
}

func TestDeflateCompress(t *testing.T) {
	tests := []struct {
		name    string
		input   []byte
		wantErr bool
	}{
		{
			name:    "Non-empty input",
			input:   []byte("Hello, world!"),
			wantErr: false,
		},
		{
			name:    "Empty input",
			input:   []byte(""),
			wantErr: false,
		},
		{
			name:    "Space case",
			input:   []byte(" "),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, encoding, err := deflateCompress(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("deflateCompress() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// If no error, verify the output by decompressing
			if !tt.wantErr {
				decompressed, err := flateDecompress(got)
				assert.NoError(t, err)
				assert.Equal(t, tt.input, decompressed)
				assert.Equal(t, "deflate", encoding)
			}
		})
	}
}

func brotliDecompress(data []byte) ([]byte, error) {
	reader := brotli.NewReader(bytes.NewReader(data))
	decompressed, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	return decompressed, nil
}

func TestBrotliCompress(t *testing.T) {
	tests := []struct {
		name    string
		input   []byte
		wantErr bool
	}{
		{
			name:    "Non-empty input",
			input:   []byte("Hello, world!"),
			wantErr: false,
		},
		{
			name:    "Empty input",
			input:   []byte(""),
			wantErr: false,
		},
		{
			name:    "Space case",
			input:   []byte(" "),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, encoding, err := brotliCompress(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("brotliCompress() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// If no error, verify the output by decompressing
			if !tt.wantErr {
				decompressed, err := brotliDecompress(got)
				assert.NoError(t, err)
				assert.Equal(t, tt.input, decompressed)
				assert.Equal(t, "br", encoding)
			}
		})
	}
}

func TestEncodeBody(t *testing.T) {
	const bodyText = "Hello, world!"

	tests := []struct {
		name                 string
		body                 []byte
		acceptEncodingHeader string
		wantErr              bool
		expectedOutput       []byte // nil when we expect an error
		expectedEncoding     string // empty when we expect an error
	}{
		{
			name:                 "Encode with gzip",
			body:                 []byte(bodyText),
			acceptEncodingHeader: "gzip",
			wantErr:              false,
			expectedOutput:       []byte(bodyText),
			expectedEncoding:     "gzip",
		},
		{
			name:                 "Encode with deflate",
			body:                 []byte(bodyText),
			acceptEncodingHeader: "deflate",
			wantErr:              false,
			expectedOutput:       []byte(bodyText),
			expectedEncoding:     "deflate",
		},
		{
			name:                 "No encoding",
			body:                 []byte(bodyText),
			acceptEncodingHeader: "",
			wantErr:              false,
			expectedOutput:       []byte(bodyText),
			expectedEncoding:     "",
		},
		{
			name:                 "identity encoding",
			body:                 []byte(bodyText),
			acceptEncodingHeader: "identity",
			wantErr:              false,
			expectedOutput:       []byte(bodyText),
			expectedEncoding:     "",
		},
		{
			name:                 "Brotli encoding",
			body:                 []byte(bodyText),
			acceptEncodingHeader: "br",
			wantErr:              false,
			expectedOutput:       []byte(bodyText),
			expectedEncoding:     "br",
		},
		{
			name:                 "Empty body with gzip",
			body:                 []byte(""),
			acceptEncodingHeader: "gzip",
			wantErr:              false,
			expectedOutput:       []byte(""),
			expectedEncoding:     "gzip",
		},
		{
			name:                 "Empty body with no encoding",
			body:                 []byte(""),
			acceptEncodingHeader: "",
			wantErr:              false,
			expectedOutput:       []byte(""),
			expectedEncoding:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, encoding, err := EncodeBody(tt.body, tt.acceptEncodingHeader)
			if (err != nil) != tt.wantErr {
				t.Errorf("EncodeBody() test = %s output = %s encoding = %s error = %v, wantErr %v", tt.name, output, encoding, err, tt.wantErr)
				return
			}
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.Equal(t, tt.expectedEncoding, encoding)

				var decompressed []byte
				var decompressErr error

				switch tt.acceptEncodingHeader {
				case "gzip":
					assert.Equal(t, "gzip", encoding)
					decompressed, decompressErr = gzipDecompress(output)
					assert.NoError(t, decompressErr)
					assert.Equal(t, tt.expectedOutput, decompressed)
				case "deflate":
					assert.Equal(t, "deflate", encoding)
					decompressed, decompressErr = flateDecompress(output)
					assert.NoError(t, decompressErr)
					assert.Equal(t, tt.expectedOutput, decompressed)
				case "br":
					assert.Equal(t, "br", encoding)
					decompressed, decompressErr = brotliDecompress(output)
					assert.NoError(t, decompressErr)
					assert.Equal(t, tt.expectedOutput, decompressed)
				default:
					assert.Equal(t, "", encoding)
					assert.Equal(t, tt.expectedOutput, output)
				}
			}
		})
	}
}
