package utils

import (
	"math/rand"
	"strings"
	"time"
)

func random(chars []rune, length int) string {
	rand.Seed(time.Now().UnixNano())
	var b strings.Builder
	for i := 0; i < length; i++ {
		b.WriteRune(chars[rand.Intn(len(chars))])
	}
	return b.String()
}

func GenerateSSID() string {
	return "fork-eat-" + random([]rune("0123456789"), 8)
}

func GeneratePassword() string {
	return random([]rune("0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"), 8)
}
