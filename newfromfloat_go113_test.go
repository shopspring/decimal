//go:build go1.13
// +build go1.13

package decimal

import (
	"math"
	"math/rand"
	"strconv"
	"testing"
)

// TestNewFromFloatShortestRoundTrip pins the contract of NewFromFloat and
// NewFromFloat32: the result is the shortest decimal that parses back to the
// original float. It covers the exponent and significand shapes a
// shortest-representation conversion has to get right, namely subnormals,
// three-digit exponents, single-digit significands and the range extremes.
//
// It is gated on Go 1.13 because before that strconv itself could return a
// shortest form that was not the nearest one (golang/go#29491), which would make
// the reference value below wrong rather than the conversion under test.
func TestNewFromFloatShortestRoundTrip(t *testing.T) {
	f64 := []float64{
		1, -1, 0.1, -0.1, 0.5, 2.5, 100, 123.456,
		1e15, 1e16, 1e17, 1e21, 1e22, 1e23, 9007199254740992,
		1e-100, 1e100, 1e-300, 1e300,
		math.MaxFloat64, -math.MaxFloat64,
		math.SmallestNonzeroFloat64, -math.SmallestNonzeroFloat64,
		math.Pi, math.E,
	}
	// Every power of two, which sweeps the whole exponent range including the
	// subnormals, plus both neighbours of every value collected so far.
	for e := -1074; e <= 1023; e++ {
		f64 = append(f64, math.Ldexp(1, e))
	}
	for _, f := range append([]float64{}, f64...) {
		f64 = append(f64, math.Nextafter(f, math.Inf(1)), math.Nextafter(f, math.Inf(-1)))
	}

	for _, f := range f64 {
		if f == 0 || math.IsInf(f, 0) {
			continue
		}
		want, err := NewFromString(strconv.FormatFloat(f, 'f', -1, 64))
		if err != nil {
			t.Fatalf("NewFromString(%v): %v", f, err)
		}
		got := NewFromFloat(f)
		if !got.Equal(want) {
			t.Errorf("NewFromFloat(%v) = %s (%s, %d), want %s (%s, %d)",
				f, got, got.value, got.exp, want, want.value, want.exp)
		}
		if back, _ := got.Float64(); back != f {
			t.Errorf("NewFromFloat(%v).Float64() = %v, does not round-trip", f, back)
		}
	}

	f32 := []float32{
		1, -1, 0.1, -0.1, 100, 123.456,
		math.MaxFloat32, -math.MaxFloat32,
		math.SmallestNonzeroFloat32, -math.SmallestNonzeroFloat32,
	}
	for e := -149; e <= 127; e++ {
		f32 = append(f32, float32(math.Ldexp(1, e)))
	}
	for _, f := range f32 {
		if f == 0 {
			continue
		}
		want, err := NewFromString(strconv.FormatFloat(float64(f), 'f', -1, 32))
		if err != nil {
			t.Fatalf("NewFromString(%v): %v", f, err)
		}
		got := NewFromFloat32(f)
		if !got.Equal(want) {
			t.Errorf("NewFromFloat32(%v) = %s (%s, %d), want %s (%s, %d)",
				f, got, got.value, got.exp, want, want.value, want.exp)
		}
	}
}

// TestNewFromFloatRandomRoundTrip sweeps random bit patterns, which puts a
// meaningful share of subnormals and extreme exponents through the conversion.
func TestNewFromFloatRandomRoundTrip(t *testing.T) {
	rng := rand.New(rand.NewSource(0xdead1337))
	for i := 0; i < 200000; i++ {
		f := math.Float64frombits(rng.Uint64())
		if f == 0 || math.IsNaN(f) || math.IsInf(f, 0) {
			continue
		}
		want, err := NewFromString(strconv.FormatFloat(f, 'f', -1, 64))
		if err != nil {
			t.Fatalf("NewFromString(%v): %v", f, err)
		}
		if got := NewFromFloat(f); !got.Equal(want) {
			t.Fatalf("NewFromFloat(%v) = %s (%s, %d), want %s (%s, %d)",
				f, got, got.value, got.exp, want, want.value, want.exp)
		}
	}
	for i := 0; i < 200000; i++ {
		f := math.Float32frombits(rng.Uint32())
		if f == 0 || math.IsNaN(float64(f)) || math.IsInf(float64(f), 0) {
			continue
		}
		want, err := NewFromString(strconv.FormatFloat(float64(f), 'f', -1, 32))
		if err != nil {
			t.Fatalf("NewFromString(%v): %v", f, err)
		}
		if got := NewFromFloat32(f); !got.Equal(want) {
			t.Fatalf("NewFromFloat32(%v) = %s (%s, %d), want %s (%s, %d)",
				f, got, got.value, got.exp, want, want.value, want.exp)
		}
	}
}
