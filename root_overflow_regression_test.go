package decimal

import (
	"math"
	"testing"
)

func TestRootRoundScaleOverflow(t *testing.T) {
	for _, tc := range []struct {
		name   string
		n      int64
		places int32
	}{
		{"positive wrap", math.MaxInt64, 1},
		{"positive wrap multiple times", math.MaxInt64, 3},
		{"minimum int64 wrap", 1 << 62, 1},
		{"negative wrap", math.MaxInt64, -3},
		{"result exponent overflow", 1, math.MinInt32},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, value := range []int64{-1, 0, 1} {
				d := NewFromInt(value)
				if got, err := d.RootRound(tc.n, tc.places); err == nil {
					t.Errorf("RootRound(%d, %d, %d) returned coefficient %s exponent %d without a range error", value, tc.n, tc.places, got.Coefficient(), got.Exponent())
				}
				if !d.Equal(NewFromInt(value)) || d.Exponent() != 0 {
					t.Errorf("RootRound changed its receiver: coefficient %s exponent %d", d.Coefficient(), d.Exponent())
				}
			}
		})
	}
}

func TestRootRoundLargeNegativeScale(t *testing.T) {
	// These scales fit int64 and truncate to zero before extracting the root.
	// In particular, negating the exact MinInt64 scale must not overflow.
	for _, tc := range []struct {
		n      int64
		places int32
		exp    int32
	}{
		{1 << 62, -3, 0},        // k == MinInt64
		{math.MaxInt64, -2, -1}, // k == MinInt64 after adding d.exp
		{math.MaxInt64, -2, 0},
		{math.MaxInt64, -1, 0},
	} {
		for _, value := range []int64{0, 1} {
			got, err := New(value, tc.exp).RootRound(tc.n, tc.places)
			if err != nil || !got.IsZero() || got.Exponent() != -tc.places {
				t.Errorf("RootRound(%de%d, %d, %d): coefficient %s exponent %d, error %v", value, tc.exp, tc.n, tc.places, got.Coefficient(), got.Exponent(), err)
			}
		}
	}
}
