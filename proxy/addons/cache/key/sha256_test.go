package key

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSHA256Interface(t *testing.T) {
	key := NewSHA256KeyStr("test")
	assert.Implements(t, (*Key)(nil), key)
}

func TestSHA256KeyGet(t *testing.T) {
	// Initialize a Key with an empty keySha
	key := NewSHA256KeyStr("test")

	// Call the Get method
	result := key.Get()

	// Check if the key has been computed and returned
	expectedHash := sha256.Sum256([]byte("test"))
	assert.Equal(t, expectedHash[:], result)
	assert.Equal(t, hex.EncodeToString(expectedHash[:]), key.String())
}

func BenchmarkComputeSha256ChecksumSmall(b *testing.B) {
	key := NewSHA256KeyStr(lorem)

	// Run the computeBlakeChecksum method b.N times
	for i := 0; i < b.N; i++ {
		key.compute()
	}
}

func BenchmarkComputeSha256ChecksumBig(b *testing.B) {
	testData := strings.Repeat(lorem, repeat)

	b.ResetTimer()
	key := NewSHA256KeyStr(testData)

	// Run the computeBlakeChecksum method b.N times
	for i := 0; i < b.N; i++ {
		key.compute()
	}
}
