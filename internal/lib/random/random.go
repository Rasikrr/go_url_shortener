package random

import (
	"math/rand"
	"time"
)

const charSet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func NewRandomString(length int) string {
	rnd := rand.New(rand.NewSource(time.Now().Unix()))

	alias := make([]byte, length)
	for i := range alias {
		alias[i] = charSet[rnd.Intn(len(charSet))]
	}
	return string(alias)
}
