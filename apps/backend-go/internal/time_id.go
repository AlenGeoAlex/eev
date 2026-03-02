package internal

import (
	"crypto/rand"
	"math/big"
	"time"
)

const charset = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"

func base62Encode(num int64, length int) string {
	result := make([]byte, length)
	for i := length - 1; i >= 0; i-- {
		result[i] = charset[num%36]
		num /= 36
	}
	return string(result)
}

func randomBase62(length int) string {
	b := make([]byte, length)
	for i := range b {
		n, _ := rand.Int(rand.Reader, big.NewInt(36))
		b[i] = charset[n.Int64()]
	}
	return string(b)
}

func MicroTimeID() string {
	hourBucket := time.Now().Unix() / 3600
	timePart := base62Encode(hourBucket%3844, 2)
	randomPart := randomBase62(4)

	return timePart + randomPart
}
