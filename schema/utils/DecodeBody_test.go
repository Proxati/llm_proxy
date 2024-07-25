package utils

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"testing"

	"github.com/andybalholm/brotli"
	"github.com/stretchr/testify/assert"
)

func compressGzip(data string) ([]byte, error) {
	var buf bytes.Buffer
	writer := gzip.NewWriter(&buf)
	_, err := writer.Write([]byte(data))
	if err != nil {
		return nil, err
	}
	err = writer.Close()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func compressDeflate(data string) ([]byte, error) {
	var buf bytes.Buffer
	writer, err := flate.NewWriter(&buf, flate.DefaultCompression)
	if err != nil {
		return nil, err
	}
	_, err = writer.Write([]byte(data))
	if err != nil {
		return nil, err
	}
	err = writer.Close()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func compressBrotili(data string) ([]byte, error) {
	var buf bytes.Buffer
	writer := brotli.NewWriter(&buf)
	_, err := writer.Write([]byte(data))
	if err != nil {
		return nil, err
	}
	err = writer.Close()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func TestDecodeBody(t *testing.T) {
	originalString := "This is a test string for compression."

	// Test gzip decompression
	gzipCompressed, err := compressGzip(originalString)
	assert.NoError(t, err)
	decompressed, err := DecodeBody(gzipCompressed, "gzip")
	assert.NoError(t, err)
	assert.Equal(t, originalString, string(decompressed))

	// Test deflate decompression
	deflateCompressed, err := compressDeflate(originalString)
	assert.NoError(t, err)
	decompressed, err = DecodeBody(deflateCompressed, "deflate")
	assert.NoError(t, err)
	assert.Equal(t, originalString, string(decompressed))

	// Test brotli decompression
	brotliCompressed, err := compressBrotili(originalString)
	assert.NoError(t, err)
	decompressed, err = DecodeBody(brotliCompressed, "br")
	assert.NoError(t, err)
	assert.Equal(t, originalString, string(decompressed))

	// Test identity encoding
	decompressed, err = DecodeBody([]byte(originalString), "identity")
	assert.NoError(t, err)
	assert.Equal(t, originalString, string(decompressed))

	// Test empty encoding
	decompressed, err = DecodeBody([]byte(originalString), "")
	assert.NoError(t, err)
	assert.Equal(t, originalString, string(decompressed))

	// Test unsupported encoding
	_, err = DecodeBody([]byte(originalString), "unsupported")
	assert.Error(t, err)
}
