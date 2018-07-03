// Package naclh is a collection of helpers for the nacl package
package naclh

import "encoding/base64"

// KeyToString returns a base64 encoded string representation of nacl key
func KeyToString(key *[32]byte) string {
	t := make([]byte, 32)
	for i, b := range *key {
		t[i] = b
	}
	return base64.StdEncoding.EncodeToString(t)
}
