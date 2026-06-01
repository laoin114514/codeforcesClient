package codeforcessdk

import (
	"crypto/sha512"
	"encoding/hex"
)

// ============================哈希编码器===============================================
type HashEncoder struct {
}

func NewHashEncoder() *HashEncoder {
	return &HashEncoder{}
}
func (h *HashEncoder) Hash512(data string) string {
	byteData := []byte(data)
	hash := sha512.Sum512(byteData)
	hashHex := hex.EncodeToString(hash[:])
	return hashHex
}
func (h *HashEncoder) Hash512_224(data string) string {
	byteData := []byte(data)
	hash := sha512.Sum512_224(byteData)
	hashHex := hex.EncodeToString(hash[:])
	return hashHex
}
func (h *HashEncoder) Hash512_256(data string) string {
	byteData := []byte(data)
	hash := sha512.Sum512_256(byteData)
	hashHex := hex.EncodeToString(hash[:])
	return hashHex
}
