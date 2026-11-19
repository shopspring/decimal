package decimal

import (
	"math"
	"math/big"
	"math/rand"
	"reflect"
	"testing"
)

func TestCmpExponentBounds(t *testing.T) {
	// Exhaust the near-boundary cases where either conservative bit bound
	// might overlap. The oracle is rational comparison, not decimal rescaling.
	for a := int64(1); a <= 64; a++ {
		for b := int64(1); b <= 64; b++ {
			for delta := int32(-5); delta <= 5; delta++ {
				for _, sign := range []int64{-1, 1} {
					left, right := New(sign*a, delta), New(sign*b, 0)
					want := left.Rat().Cmp(right.Rat())
					if got := left.Cmp(right); got != want {
						t.Fatalf("%d e%d cmp %d: got%d want%d", sign*a, delta, sign*b, got, want)
					}
					if got := right.Cmp(left); got != -want {
						t.Fatalf("reverse %d e%d cmp %d: got%d want%d", sign*a, delta, sign*b, got, -want)
					}
				}
			}
		}
	}
}

func TestCmpExponentBoundsLargeCoefficients(t *testing.T) {
	rng := rand.New(rand.NewSource(20261118))
	for i := 0; i < 10000; i++ {
		a := new(big.Int).Rand(rng, new(big.Int).Lsh(big.NewInt(1), uint(rng.Intn(1024)+1)))
		b := new(big.Int).Rand(rng, new(big.Int).Lsh(big.NewInt(1), uint(rng.Intn(1024)+1)))
		if rng.Intn(2) == 0 {
			a.Neg(a)
		}
		if rng.Intn(2) == 0 {
			b.Neg(b)
		}
		left, right := NewFromBigInt(a, int32(rng.Intn(301)-150)), NewFromBigInt(b, int32(rng.Intn(301)-150))
		beforeLeft, beforeRight := left.Copy(), right.Copy()
		want := left.Rat().Cmp(right.Rat())
		if got := left.Cmp(right); got != want {
			t.Fatalf("case%d: got%d want%d", i, got, want)
		}
		if !reflect.DeepEqual(left, beforeLeft) || !reflect.DeepEqual(right, beforeRight) {
			t.Fatal("Cmp mutated operand")
		}
	}
}

func TestCmpExponentBoundsFullExponentRange(t *testing.T) {
	for _, sign := range []int64{-1, 1} {
		for _, pair := range [][2]int32{{math.MaxInt32, math.MinInt32}, {math.MaxInt32, 0}, {0, math.MinInt32}, {math.MinInt32 + 1, math.MinInt32}} {
			high, low := New(sign, pair[0]), New(sign, pair[1])
			if got := high.Cmp(low); got != int(sign) {
				t.Fatalf("exp%v sign%d: got%d", pair, sign, got)
			}
			if got := low.Cmp(high); got != -int(sign) {
				t.Fatalf("reverse exp%v sign%d: got%d", pair, sign, got)
			}
		}
	}
}

func TestCmpExponentBoundsNoScalingAllocation(t *testing.T) {
	// Both values are valid under the default decode exponent limit. Comparing
	// clearly separated magnitudes should not materialize a power of ten.
	high, low := New(1, 10000), New(1, 0)
	allocations := testing.AllocsPerRun(100, func() {
		if high.Cmp(low) != 1 {
			t.Fatal("incorrect comparison")
		}
	})
	if allocations != 0 {
		t.Fatalf("disjoint magnitudes allocated %v times", allocations)
	}
}

func BenchmarkCmpExponentBounds(b *testing.B) {
	for _, c := range []struct {
		name string
		a, b Decimal
	}{
		{"same_scale", New(12345, -2), New(12346, -2)},
		{"overlap", New(12345, -2), New(123456, -3)},
		{"disjoint", New(12345, -2), New(2345, -3)},
		{"negative", New(-12345, -2), New(-2345, -3)},
		{"zero", New(12345, -2), Zero},
		{"gap100", New(1, 100), New(1, 0)},
		{"gap1000", New(1, 1000), New(1, 0)},
		{"gap10000", New(1, 10000), New(1, 0)},
	} {
		b.Run(c.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				if c.a.Cmp(c.b) == 2 {
					b.Fatal("invalid comparison")
				}
			}
		})
	}
}
