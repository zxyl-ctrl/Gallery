package rand

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

// 重新封装随机数生成器的原因：好用
func Bytes(n int) ([]byte, error) {
	b := make([]byte, n)
	nRead, err := rand.Read(b)
	if err != nil {
		return nil, fmt.Errorf("bytes: %w", err)
	}
	if nRead < n {
		return nil, fmt.Errorf("bytes: didn't read enough random bytes")
	}
	return b, nil
}

// String returns a random string using crypto/rand.
// n is the number of bytes being used to generate the random string.
func String(n int) (string, error) {
	b, err := Bytes(n)
	if err != nil {
		return "", fmt.Errorf("string: %w", err)
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// const SessionTokenBytes = 32

// func SessionToken() (string, error) {
// 	return String(SessionTokenBytes)
// }

// 为什么使用32字节的数字，1 Byte = 8 bit  256种可能，如果32 Byte 对应的是 256^32 ≈ 1e77 种可能性
// OWASP Foundation建议使用至少128比特或者16字节的session，使用32是因为比至少的好一点
