package core

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHasher(t *testing.T) {
	// get photo
	photoBytes, err := ioutil.ReadFile("../etc/samples/photos/robin.jpg")
	if err != nil {
		panic(err)
	}
	hashed, err := GetHash(photoBytes)
	assert.NoError(t, err)
	assert.Equal(t, "EgJCfVfExg34MA8VtDjR9SmGz8pgGKZbcuBMQCFHhnc=", hashed)
}
