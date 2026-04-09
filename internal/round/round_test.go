package round_test

import (
	"testing"

	"github.com/goscho/fw-pegel-dispatcher/internal/round"
)

func TestFloat32(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		in   float32
		prec int
		want float32
	}{
		{"port1 two decimals", 0.123, 2, 0.12},
		{"port2 half up to int", 0.5, 0, 1},
		{"port2 half up", 0.6, 0, 1},
		{"port2 down", 0.4, 0, 0},
		{"webio sample", 0.209, 2, 0.21},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := round.Float32(tc.in, tc.prec)
			if got != tc.want {
				t.Fatalf("Float32(%v, %d) = %v, want %v", tc.in, tc.prec, got, tc.want)
			}
		})
	}
}
