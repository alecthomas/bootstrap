package util

import (
	"github.com/stretchrcom/testify/assert"

	"testing"
)

func TestIDObfuscator(t *testing.T) {
	o := NewIDObfuscator(0x12345789)
	id, err := o.Decode(o.Encode(123123123))
	assert.NoError(t, err)
	assert.Equal(t, uint64(123123123), id)
}
