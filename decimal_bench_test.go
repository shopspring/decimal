// +build !go1.2
// +build !go1.3
// +build !go1.4

package decimal

import (
	"math/big"
	"testing"
)

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

func Benchmark_math_big_Float_Sub_different_precision(b *testing.B) {
	d1 := big.NewFloat(1000.123)
	d2 := d1.Mul(big.NewFloat(500), big.NewFloat(0.12))

	b.ReportAllocs()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		d1.Sub(d1, d2)
	}
}

func Benchmark_math_big_Float_Add_different_precision(b *testing.B) {
	d1 := big.NewFloat(1000.123)
	d2 := d1.Mul(big.NewFloat(500), big.NewFloat(0.12))

	b.ReportAllocs()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		d1.Add(d1, d2)
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

func Benchmark_math_big_Float_Add_same_precision(b *testing.B) {
	d1 := big.NewFloat(1000.123)
	d2 := big.NewFloat(500.123)

	b.ReportAllocs()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		d1.Add(d1, d2)
	}
}

func Benchmark_math_big_Float_Sub_same_precision(b *testing.B) {
	d1 := big.NewFloat(1000.123)
	d2 := big.NewFloat(500.123)

	b.ReportAllocs()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		d1.Sub(d1, d2)
	}
}
