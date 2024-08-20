package key

import (
	"encoding/hex"

	"golang.org/x/crypto/blake2b"
)

// Key is an object that stores the key data (in bytes), and also a pre-computed checksum of the key
type BLAKE2 struct {
	keyBytes []byte // keyBytes is the raw key data
	keyHash  []byte // keyHash is the pre-computed checksum of the keyBytes
}

// compute runs a hash on the keyBytes and stores the result
func (k *BLAKE2) compute() {
	hash := blake2b.Sum512(k.keyBytes)
	k.keyHash = hash[:]
}

// Get returns the hashed key data
func (k *BLAKE2) Get() []byte {
	if len(k.keyHash) == 0 {
		k.compute()
	}

	return k.keyHash
}

// String returns the hashed key data as a string
func (k *BLAKE2) String() string {
	return hex.EncodeToString(k.Get())
}

// NewKey creates a new Key object with the given key data
func NewBLAKE2Key(keyBytes []byte) *BLAKE2 {
	k := &BLAKE2{
		keyBytes: keyBytes,
	}
	k.compute()
	return k
}

// NewKeyStr creates a new Key object with the given key string data
func NewBLAKE2KeyStr(key string) *BLAKE2 {
	return NewBLAKE2Key([]byte(key))
}
