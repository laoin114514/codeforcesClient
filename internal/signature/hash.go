// Package signature 提供 Codeforces API 请求签名所需的密码学工具。
package signature

import (
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"math/rand"
)

// SHA512Sum 返回 data 的十六进制 SHA-512 哈希值。
func SHA512Sum(data string) string {
	hash := sha512.Sum512([]byte(data))
	return hex.EncodeToString(hash[:])
}

// RandomPrefix 返回 6 位零填充的随机数，
// 作为 Codeforces apiSig 计算中的随机前缀。
func RandomPrefix() string {
	return fmt.Sprintf("%06d", rand.Intn(1000000))
}
