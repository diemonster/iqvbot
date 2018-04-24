package slash

import (
	"math/rand"
	"time"
)

// Pacific Standard Time location
var PST *time.Location

func init() {
	rand.Seed(time.Now().UTC().UnixNano())

	p, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		panic(err)
	}

	PST = p
}
