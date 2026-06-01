package signature

import (
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"math/rand"
)

func SHA512Sum(data string) string {
	hash := sha512.Sum512([]byte(data))
	return hex.EncodeToString(hash[:])
}

func RandomPrefix() string {
	return fmt.Sprintf("%06d", rand.Intn(1000000))
}
