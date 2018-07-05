// Package cryptobox is a collection of crypto helpers
package cryptobox

import "encoding/base64"

// KeyToString returns a base64 encoded string representation of nacl key
func KeyToString(key *[32]byte) string {
	t := make([]byte, 32)
	for i, b := range *key {
		t[i] = b
	}
	return base64.StdEncoding.EncodeToString(t)
}
