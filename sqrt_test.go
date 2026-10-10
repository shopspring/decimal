package decimal

import (
	"testing"
)

func TestSqrt(t *testing.T) {
	tables := []struct {
		x Decimal
		n Decimal
	}{
		{One, One},
		{Four, Two},
		{Sixteen, Four},
		{Two, NewFromFloat(1.4142135623730951)},
	}

	for _, table := range tables {
		result := table.x.Sqrt()
		if result.NotEqual(table.n) {
			t.Errorf("Sqrt of (%v) was incorrect, got: %v, want: %v.", table.x.String(), result.String(), table.n.String())
		}
	}
}
