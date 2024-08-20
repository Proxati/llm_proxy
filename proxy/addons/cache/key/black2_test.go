package key

import (
	"encoding/hex"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/blake2b"
)

func TestBLAKE2Interface(t *testing.T) {
	key := NewBLAKE2KeyStr("test")
	assert.Implements(t, (*Key)(nil), key)
}

func TestBLAKE2KeyGet(t *testing.T) {
	// Initialize a Key with an empty keySha
	key := NewBLAKE2KeyStr("test")

	// Call the Get method
	result := key.Get()

	// Check if the keySha has been computed and returned
	expectedHash := blake2b.Sum512([]byte("test"))
	assert.Equal(t, expectedHash[:], result)
	assert.Equal(t, hex.EncodeToString(expectedHash[:]), key.String())
}

func TestComputeBlake2Checksum(t *testing.T) {
	key := NewBLAKE2KeyStr("test")

	// Call the computeBlakeChecksum method
	key.compute()

	// Check if the keySha has been computed correctly
	expectedHash := blake2b.Sum512([]byte("test"))
	assert.Equal(t, expectedHash[:], key.keyHash, "The computeBlakeChecksum method did not compute the expected result")
}

func BenchmarkComputeBlake2ChecksumSmall(b *testing.B) {
	key := NewBLAKE2KeyStr(lorem) // from testdata_test.go

	// Run the computeBlakeChecksum method b.N times
	for i := 0; i < b.N; i++ {
		key.compute()
	}
}

func BenchmarkComputeBlake2ChecksumBig(b *testing.B) {
	testData := strings.Repeat(lorem, repeat)

	b.ResetTimer()
	key := NewBLAKE2KeyStr(testData)

	// Run the computeBlakeChecksum method b.N times
	for i := 0; i < b.N; i++ {
		key.compute()
	}
}
