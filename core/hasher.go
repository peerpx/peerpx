package core

import (
	"crypto/sha256"
	"encoding/base64"
)

// GetHash is mainly used to get hash (universal & unique ID) of a photo
// Base64(sha256([]byte))
func GetHash(data []byte) (string, error) {
	h := sha256.New()
	if _, err := h.Write(data); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(h.Sum(nil)), nil
}
