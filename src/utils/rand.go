package utils

import (
	"fmt"
	"math/rand"
)

var chars = []rune("1234567890abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

var length = 16

func randStr(prefix string) string {
	b := make([]rune, length)

	for i := 0; i < length; i++ {
		b[i] = chars[rand.Intn(len(chars))]
	}

	return fmt.Sprintf("%s%s", prefix, string(b))
}

func NewSessID() string {
	return randStr("sess:")
}

// TODO: was uint32 in reference, why?
func NewClientID() string {
	return randStr("client:")
}
