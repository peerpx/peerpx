package core

import (
	"crypto/sha256"

	"github.com/shengdoushi/base58"
)

// GetHash is mainly used to get hash (universal & unique ID) of a photo
// Base58(sha256([]byte))
func GetHash(data []byte) (string, error) {
	h := sha256.New()
	if _, err := h.Write(data); err != nil {
		return "", err
	}
	return base58.Encode(h.Sum(nil), base58.IPFSAlphabet), nil
}
