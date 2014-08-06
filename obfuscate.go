package util

import (
	"encoding/base64"
	"encoding/binary"
	"math/rand"
	"strings"
)

const (
	urlEncoding = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

type IDObfuscator struct {
	enc *base64.Encoding
}

// NewIDObfuscator encodes uint64 values using.
func NewIDObfuscator(key int64) *IDObfuscator {
	r := rand.New(rand.NewSource(key))
	n := r.Perm(len(urlEncoding))
	out := make([]byte, len(urlEncoding))
	for i, v := range n {
		out[i] = urlEncoding[v]
	}
	return &IDObfuscator{
		enc: base64.NewEncoding(string(out)),
	}
}

func (i *IDObfuscator) Encode(v uint64) string {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, v)
	for len(b) > 0 && b[len(b)-1] == 0 {
		b = b[:len(b)-1]
	}
	return strings.TrimRight(i.enc.EncodeToString(b), "=")
}

func (i *IDObfuscator) Decode(s string) (uint64, error) {
	if m := len(s) % 4; m != 0 {
		s += strings.Repeat("=", 4-m)
	}
	ib, err := i.enc.DecodeString(s)
	if err != nil {
		return 0, err
	}
	b := make([]byte, 8)
	copy(b, ib)
	return binary.LittleEndian.Uint64(b), nil
}
