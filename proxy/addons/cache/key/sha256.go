package key

import (
	"crypto/sha256"
	"encoding/hex"
)

// Key is an object that stores the key data (in bytes), and also a pre-computed checksum of the key
type SHA256 struct {
	keyBytes []byte // keyBytes is the raw key data
	keyHash  []byte // keyHash is the pre-computed checksum of the keyBytes
}

func (k *SHA256) compute() {
	hash := sha256.Sum256(k.keyBytes)
	k.keyHash = hash[:]
}

// Get returns the hashed key data
func (k *SHA256) Get() []byte {
	if len(k.keyHash) == 0 {
		k.compute()
	}

	return k.keyHash
}

// String returns the hashed key data as a string
func (k *SHA256) String() string {
	return hex.EncodeToString(k.Get())
}

// NewKey creates a new Key object with the given key data
func NewSHA256Key(keyBytes []byte) *SHA256 {
	k := &SHA256{
		keyBytes: keyBytes,
	}
	k.compute()
	return k
}

// NewKeyStr creates a new Key object with the given key string data
func NewSHA256KeyStr(key string) *SHA256 {
	return NewSHA256Key([]byte(key))
}
