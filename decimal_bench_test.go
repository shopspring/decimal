package decimal

import (
	"fmt"
	"math"
	"math/big"
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
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewFromFloatWithExponent(in[i], math.MinInt32)
	}
}

func BenchmarkNewFromFloat(b *testing.B) {
	rng := rand.New(rand.NewSource(0xdead1337))
	in := make([]float64, b.N)
	for i := range in {
		in[i] = rng.NormFloat64() * 10e20
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewFromFloat(in[i])
	}
}

func BenchmarkNewFromStringFloat(b *testing.B) {
	rng := rand.New(rand.NewSource(0xdead1337))
	in := make([]string, b.N)
	for i := range in {
		in[i] = strconv.FormatFloat(rng.NormFloat64()*10e20, 'f', -1, 64)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = NewFromString(in[i])
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

func numDigits(b *testing.B, want int, val Decimal) {
	b.Helper()
	for i := 0; i < b.N; i++ {
		if have := val.NumDigits(); have != want {
			b.Fatalf("\nHave: %d\nWant: %d", have, want)
		}
	}
}

func BenchmarkDecimal_NumDigits10(b *testing.B) {
	numDigits(b, 10, New(3478512345, -3))
}

func BenchmarkDecimal_NumDigits100(b *testing.B) {
	s := make([]byte, 102)
	for i := range s {
		s[i] = byte('0' + i%10)
	}
	s[0] = '-'
	s[100] = '.'
	d, err := NewFromString(string(s))
	if err != nil {
		b.Log(d)
		b.Error(err)
	}
	numDigits(b, 100, d)
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

func BenchmarkDecimal_Add_different_precision(b *testing.B) {
	d1 := NewFromFloat(1000.123)
	d2 := NewFromFloat(500).Mul(NewFromFloat(0.12))

	b.ReportAllocs()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		d1.Add(d2)
	}
}

func BenchmarkDecimal_Sub_different_precision(b *testing.B) {
	d1 := NewFromFloat(1000.123)
	d2 := NewFromFloat(500).Mul(NewFromFloat(0.12))

	b.ReportAllocs()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		d1.Sub(d2)
	}
}

func BenchmarkDecimal_Add_same_precision(b *testing.B) {
	d1 := NewFromFloat(1000.123)
	d2 := NewFromFloat(500.123)

	b.ReportAllocs()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		d1.Add(d2)
	}
}

func BenchmarkDecimal_Sub_same_precision(b *testing.B) {
	d1 := NewFromFloat(1000.123)
	d2 := NewFromFloat(500.123)

	b.ReportAllocs()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		d1.Sub(d2)
	}
}

func BenchmarkDecimal_IsInteger(b *testing.B) {
	d := RequireFromString("12.000")

	b.ReportAllocs()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		d.IsInteger()
	}
}

func BenchmarkDecimal_Pow(b *testing.B) {
	d1 := RequireFromString("5.2")
	d2 := RequireFromString("6.3")

	for i := 0; i < b.N; i++ {
		d1.Pow(d2)
	}
}

func BenchmarkDecimal_PowWithPrecision(b *testing.B) {
	d1 := RequireFromString("5.2")
	d2 := RequireFromString("6.3")

	for i := 0; i < b.N; i++ {
		_, _ = d1.PowWithPrecision(d2, 8)
	}
}
func BenchmarkDecimal_PowInt32(b *testing.B) {
	d1 := RequireFromString("5.2")
	d2 := int32(10)

	for i := 0; i < b.N; i++ {
		_, _ = d1.PowInt32(d2)
	}
}

func BenchmarkDecimal_PowBigInt(b *testing.B) {
	d1 := RequireFromString("5.2")
	d2 := big.NewInt(10)

	for i := 0; i < b.N; i++ {
		_, _ = d1.PowBigInt(d2)
	}
}

func BenchmarkDecimal_PowNegativeExponent(b *testing.B) {
	d1 := RequireFromString("1.0041666666666667")
	d2 := RequireFromString("-360")

	for i := 0; i < b.N; i++ {
		d1.Pow(d2)
	}
}

func BenchmarkDecimal_NewFromString(b *testing.B) {
	count := 72
	prices := make([]string, 0, count)
	for i := 1; i <= count; i++ {
		prices = append(prices, fmt.Sprintf("%d.%d", i*100, i))
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, p := range prices {
			d, err := NewFromString(p)
			if err != nil {
				b.Log(d)
				b.Error(err)
			}
		}
	}
}

func BenchmarkDecimal_NewFromString_large_number(b *testing.B) {
	count := 72
	prices := make([]string, 0, count)
	for i := 1; i <= count; i++ {
		prices = append(prices, "9323372036854775807.9223372036854775807")
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, p := range prices {
			d, err := NewFromString(p)
			if err != nil {
				b.Log(d)
				b.Error(err)
			}
		}
	}
}

func BenchmarkDecimal_ExpHullAbraham(b *testing.B) {
	b.ResetTimer()

	d := RequireFromString("30.412346346346")

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = d.ExpHullAbrham(10)
	}
}

func BenchmarkDecimal_ExpTaylor(b *testing.B) {
	b.ResetTimer()

	d := RequireFromString("30.412346346346")

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = d.ExpTaylor(10)
	}
}

func BenchmarkDecimal_Normalize(b *testing.B) {
	for _, c := range []struct {
		name string
		d    Decimal
	}{
		{"NoZeros", RequireFromString("123.456")},
		{"OneZero", RequireFromString("12.50")},
		{"Div", NewFromInt(5).Div(NewFromInt(2))},
		{"Big", RequireFromString("123456789012345678901234567890.5000")},
		{"ManyZeros", RequireFromString("1.5").Round(10000)},
	} {
		b.Run(c.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				c.d.Normalize()
			}
		})
	}
}

func BenchmarkDecimal_Sqrt(b *testing.B) {
	for _, c := range []struct {
		name   string
		d      Decimal
		places int32
	}{
		{"Small", RequireFromString("0.002342"), 16},
		{"Int", NewFromInt(2), 16},
		{"Big", RequireFromString("123456789012345678901234567890.123"), 16},
		{"Places1000", NewFromInt(2), 1000},
	} {
		b.Run(c.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_, _ = c.d.SqrtRound(c.places)
			}
		})
	}
}

func BenchmarkDecimal_UnmarshalJSON(b *testing.B) {
	b.ResetTimer()

	bstr := []byte("1234.56789")

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = (&Decimal{}).UnmarshalJSON(bstr)
	}
}
