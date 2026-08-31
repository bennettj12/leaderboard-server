package main

import (
	"fmt"
	"math"
	"math/rand"
	"testing"
)

type geometry interface {
	area() int
	isSquare() bool
}
type rect struct {
	c vec2
}
type vec2 struct {
	x, y int
}
type line struct {
	length int
}

func (l line) area() int {
	return 0
}
func (l line) isSquare() bool {
	return false
}
func (r rect) area() int {
	return r.c.x * r.c.y
}
func (r rect) isSquare() bool {
	return r.c.x == r.c.y
}
func randVec2() vec2 {
	return vec2{rand.Intn(10), rand.Intn(10)}
}
func isBigSquare(g geometry) bool {
	return g.isSquare() && g.area() > 10
}
func TestMain(t *testing.T) {
	random := rand.Intn(32)
	g := rect{vec2{10, 10}}
	if random > 32 {
		t.Error("Intn higher than max")
	}
	fmt.Println("R:	", random)
	fmt.Println("PI:	", math.Pi)
	fmt.Println("RV:	", randVec2())
	fmt.Println("GA:", g.area())

}
