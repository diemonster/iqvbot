package slash

import (
	"math/rand"
	"time"
)

var PDT = time.FixedZone("PDT", -25200)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

func randomString(length int) string {
	runes := []rune("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

	b := make([]rune, length)
	for i := range b {
		b[i] = runes[rand.Intn(len(runes))]
	}

	return string(b)
}
