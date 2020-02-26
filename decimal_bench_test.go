package decimal

import (
	"math"
	"math/rand"
	"sort"
	"strconv"
	"testing"
)

type DecimalSlice []Decimal

func (p DecimalSlice) Len() int           { return len(p) }
func (p DecimalSlice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p DecimalSlice) Less(i, j int) bool { return p[i].Cmp(p[j]) < 0 }

func BenchmarkNewFromFloatWithExponent(b *testing.B) {
	rng := rand.New(rand.NewSource(0xdead1337))
	in := make([]float64, b.N)
	for i := range in {
		in[i] = rng.NormFloat64() * 10e20
	}
	b.ReportAllocs()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		in := rng.NormFloat64() * 10e20
		_ = NewFromFloatWithExponent(in, math.MinInt32)
	}
}

func BenchmarkNewFromFloat(b *testing.B) {
	rng := rand.New(rand.NewSource(0xdead1337))
	in := make([]float64, b.N)
	for i := range in {
		in[i] = rng.NormFloat64() * 10e20
	}
	b.ReportAllocs()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		_ = NewFromFloat(in[i])
	}
}

func BenchmarkNewFromStringFloat(b *testing.B) {
	rng := rand.New(rand.NewSource(0xdead1337))
	in := make([]float64, b.N)
	for i := range in {
		in[i] = rng.NormFloat64() * 10e20
	}
	b.ReportAllocs()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		in := strconv.FormatFloat(in[i], 'f', -1, 64)
		_, _ = NewFromString(in)
	}
}

func Benchmark_FloorFast(b *testing.B) {
	input := New(200, 2)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		input.Floor()
	}
}

func Benchmark_FloorRegular(b *testing.B) {
	input := New(200, -2)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		input.Floor()
	}
}

func Benchmark_DivideOriginal(b *testing.B) {
	tcs := createDivTestCases()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, tc := range tcs {
			d := tc.d
			if sign(tc.d2) == 0 {
				continue
			}
			d2 := tc.d2
			prec := tc.prec
			a := d.DivOld(d2, int(prec))
			if sign(a) > 2 {
				panic("dummy panic")
			}
		}
	}
}

func Benchmark_DivideNew(b *testing.B) {
	tcs := createDivTestCases()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, tc := range tcs {
			d := tc.d
			if sign(tc.d2) == 0 {
				continue
			}
			d2 := tc.d2
			prec := tc.prec
			a := d.DivRound(d2, prec)
			if sign(a) > 2 {
				panic("dummy panic")
			}
		}
	}
}

func BenchmarkDecimal_RoundCash_Five(b *testing.B) {
	const want = "3.50"
	for i := 0; i < b.N; i++ {
		val := New(3478, -3)
		if have := val.StringFixedCash(5); have != want {
			b.Fatalf("\nHave: %q\nWant: %q", have, want)
		}
	}
}

func Benchmark_Cmp(b *testing.B) {
	decimals := DecimalSlice([]Decimal{})
	for i := 0; i < 1000000; i++ {
		decimals = append(decimals, New(int64(i), 0))
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sort.Sort(decimals)
	}
}

func Benchmark_decimal_Decimal_Add_different_precision(b *testing.B) {
	d1 := NewFromFloat(1000.123)
	d2 := NewFromFloat(500).Mul(NewFromFloat(0.12))

	b.ReportAllocs()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		d1.Add(d2)
	}
}

func Benchmark_decimal_Decimal_Sub_different_precision(b *testing.B) {
	d1 := NewFromFloat(1000.123)
	d2 := NewFromFloat(500).Mul(NewFromFloat(0.12))

	b.ReportAllocs()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		d1.Sub(d2)
	}
}

func Benchmark_decimal_Decimal_Add_same_precision(b *testing.B) {
	d1 := NewFromFloat(1000.123)
	d2 := NewFromFloat(500.123)

	b.ReportAllocs()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		d1.Add(d2)
	}
}

func Benchmark_decimal_Decimal_Sub_same_precision(b *testing.B) {
	d1 := NewFromFloat(1000.123)
	d2 := NewFromFloat(500.123)

	b.ReportAllocs()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		d1.Add(d2)
	}
}
