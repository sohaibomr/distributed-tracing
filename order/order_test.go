package main

import (
	"math/rand"
	"testing"
	"time"
)

func TestRandInt(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	got := randomInt()
	if got < 1 || got > 8 {
		t.Errorf("randomInt() = %d; want >= 1 and <= 8", got)
	}
}
