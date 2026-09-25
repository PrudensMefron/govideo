package core

import (
	"math"
	"testing"
)

func TestTrimDurationCountsBackwardsFromEnd(t *testing.T) {
	d, err := TrimDuration(240, 20, 15)
	if err != nil || d != 205 {
		t.Fatalf("got %v, %v; expected 205 seconds", d, err)
	}
	for _, v := range [][3]float64{{0, 0, 0}, {10, 0, 0}, {10, -1, 2}, {10, 2, -1}, {10, 8, 2}, {10, 8, 3}, {10, 9.95, 0}, {10, math.NaN(), 0}, {10, 0, math.Inf(1)}, {math.Inf(1), 1, 0}} {
		if _, err := TrimDuration(v[0], v[1], v[2]); err == nil {
			t.Errorf("accepted %v", v)
		}
	}
}
