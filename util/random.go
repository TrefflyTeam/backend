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

const (
	defaultAlphabet = "abcdefghijklmnopqrstuvwxyz0123456789"
	emailAlphabet = "abcdefghijklmnopqrstuvwxyz0123456789"
	usernameAlphabet = "abcdefghijklmnopqrstuvwxyz-абвгдеёжзийклмнопрстуфхцчшщьыъэюя"
)

func RandomInt(min int64, max int64) int64 {
	return min + rng.Int63n(max-min+1)
}

func RandomString(length int) string {
	return randomAlphabetString(length, defaultAlphabet)
}

func RandomUsername() string {
	usernameRunes := []rune(usernameAlphabet)
	n := rand.Intn(10) + 6
	b := make([]rune, n)
	for i := range b {
		b[i] = usernameRunes[rand.Intn(len(usernameRunes))]
	}
	return string(b)
}

func RandomEmail() string {
	domains := []string{"example.com", "test.net", "mock.org"}
	domain := domains[rng.Intn(len(domains))]
	local := randomAlphabetString(10, emailAlphabet)
	return local + "@" + domain
}

func randomAlphabetString(length int, alphabet string) string {
	var sb strings.Builder
	k := len(alphabet)
	for i := 0; i < length; i++ {
		c := alphabet[rng.Intn(k)]
		sb.WriteByte(c)
	}

	return sb.String()
}
