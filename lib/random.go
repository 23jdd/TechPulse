package lib

import (
	"fmt"
	"math/rand"
)

func Generate() string {
	var state string
	for i := 0; i < 6; i++ {
		state += fmt.Sprintf("%d", rand.Intn(10))
	}
	return state
}
