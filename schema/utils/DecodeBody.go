package utils

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"fmt"
	"io"

	"github.com/andybalholm/brotli"
)

// DecodeBody decompresses a byte array (response body) based on the content encoding
func DecodeBody(body []byte, contentEncoding string) (decodedBody []byte, err error) {
	switch contentEncoding {
	case gzipEncoding:
		reader, err := gzip.NewReader(bytes.NewReader(body))
		if err != nil {
			return nil, fmt.Errorf("gzip decompress error: %v", err)
		}
		defer reader.Close()
		decodedBody, err = io.ReadAll(reader)
		if err != nil {
			return nil, fmt.Errorf("gzip reader error: %v", err)
		}
	case deflateEncoding:
		reader := flate.NewReader(bytes.NewReader(body))
		defer reader.Close()
		decodedBody, err = io.ReadAll(reader)
		if err != nil {
			return nil, fmt.Errorf("deflate reader error: %v", err)
		}
	case brotliEncoding:
		reader := brotli.NewReader(bytes.NewReader(body))
		decodedBody, err = io.ReadAll(reader)
		if err != nil {
			return nil, fmt.Errorf("brotli reader error: %v", err)
		}
	case "", identityEncoding:
		// no encoding, do nothing
		return body, nil
	default:
		return nil, fmt.Errorf("unsupported encoding: %s", contentEncoding)
	}

	return decodedBody, nil
}
