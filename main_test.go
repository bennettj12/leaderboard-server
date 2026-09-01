package main

import (
	"fmt"
	"math"
	"math/rand"
	"testing"
)

type vec2 struct {
	x, y int
}

func clampf(val float64, min float64, max float64) float64 {
	if val > min {
		if val < max {
			return val
		}
		return max
	} else {
		return min
	}
}
func randVec2() vec2 {
	return vec2{rand.Intn(10), rand.Intn(10)}
}
func TestMain(t *testing.T) {
	random := rand.Intn(32)
	if random > 32 {
		t.Error("Intn higher than max")
	}
	fmt.Println("R:	", random)
	fmt.Println("PI:	", math.Pi)
	fmt.Println("RV:	", randVec2())

}
