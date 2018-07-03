package naclh

import (
	"testing"

	"github.com/magiconair/properties/assert"
)

func TestKeyToString(t *testing.T) {
	key := [32]byte{13, 254, 159, 140, 148, 246, 233, 220, 80, 97, 20, 144, 38, 157, 156, 217, 39, 57, 41, 84, 244, 100, 30, 77, 45, 99, 202, 102, 23, 166, 7, 83}
	keyStr := KeyToString(&key)
	assert.Equal(t, keyStr, "Df6fjJT26dxQYRSQJp2c2Sc5KVT0ZB5NLWPKZhemB1M=")
}
