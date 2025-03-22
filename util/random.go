package util

import (
	"math/rand"
	"strings"
	"time"
)

var rng *rand.Rand

func init() {
	source := rand.NewSource(time.Now().UnixNano())
	rng = rand.New(source)
}

const alphabet = "abcdefghijklmnopqrstuvwxyz0123456789_"

func RandomInt(min int64, max int64) int64 {
	return min + rng.Int63n(max-min+1)
}

func RandomString(length int) string {
	var sb strings.Builder
	k := len(alphabet)
	for i := 0; i < length; i++ {
		c := alphabet[rng.Intn(k)]
		sb.WriteByte(c)
	}

	return sb.String()
}

func RandomUsername() string {
	return RandomString(10)
}

func RandomEmail() string {
	domains := []string{"example.com", "test.net", "mock.org"}
	domain := domains[rng.Intn(len(domains))]
	local := RandomString(10)
	return local + "@" + domain
}