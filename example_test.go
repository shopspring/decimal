package decimal_test

import (
	"fmt"
	"github.com/shopspring/decimal"
	"math/big"
	"regexp"
)

func ExampleNewFromInt() {
	fmt.Println(decimal.NewFromInt(123))
	fmt.Println(decimal.NewFromInt(-10))

	// output:
	// 123
	// -10
}

func ExampleNewFromInt32() {
	fmt.Println(decimal.NewFromInt(123))
	fmt.Println(decimal.NewFromInt(-10))

	// output:
	// 123
	// -10
}

func ExampleNewFromUint64() {
	fmt.Println(decimal.NewFromUint64(123))

	// output:
	// 123
}

func ExampleNewFromBigRat() {
	d1 := decimal.NewFromBigRat(big.NewRat(0, 1), 0)
	d2 := decimal.NewFromBigRat(big.NewRat(4, 5), 1)
	d3 := decimal.NewFromBigRat(big.NewRat(1000, 3), 3)
	d4 := decimal.NewFromBigRat(big.NewRat(2, 7), 4)

	fmt.Println(d1)
	fmt.Println(d2)
	fmt.Println(d3)
	fmt.Println(d4)

	// output:
	// 0
	// 0.8
	// 333.333
	// 0.2857
}

func ExampleNewFromString() {
	d, _ := decimal.NewFromString("-123.45")
	d2, _ := decimal.NewFromString(".0001")
	d3, _ := decimal.NewFromString("1.47000")

	fmt.Println(d)
	fmt.Println(d2)
	fmt.Println(d3)

	// output:
	// -123.45
	// 0.0001
	// 1.47
}

func ExampleNewFromFormattedString() {
	r := regexp.MustCompile("[$,]")
	d1, _ := decimal.NewFromFormattedString("$5,125.99", r)

	r2 := regexp.MustCompile("[_]")
	d2, _ := decimal.NewFromFormattedString("1_000_000", r2)

	r3 := regexp.MustCompile("[USD\\s]")
	d3, _ := decimal.NewFromFormattedString("5000 USD", r3)

	fmt.Println(d1)
	fmt.Println(d2)
	fmt.Println(d3)

	// output:
	// 5125.99
	// 1000000
	// 5000
}

func ExampleRequireFromString() {
	d1 := decimal.RequireFromString("-123.45")
	d2 := decimal.RequireFromString(".0001")

	fmt.Println(d1)
	fmt.Println(d2)

	// output:
	// -123.45
	// 0.0001
}

func ExampleNewFromFloatWithExponent() {
	d1 := decimal.NewFromFloatWithExponent(123.456, -2)
	fmt.Println(d1)

	// output:
	// 123.46
}

func ExampleDecimal_Pow() {
	d1 := decimal.NewFromFloat(4.0)
	d2 := decimal.NewFromFloat(4.0)
	res1 := d1.Pow(d2)
	fmt.Println(res1)

	d3 := decimal.NewFromFloat(5.0)
	d4 := decimal.NewFromFloat(5.73)
	res2 := d3.Pow(d4)
	fmt.Println(res2)

	// output:
	// 256
	// 10118.08037125
}

func ExampleDecimal_PowWithPrecision() {
	d1 := decimal.NewFromFloat(4.0)
	d2 := decimal.NewFromFloat(4.0)
	res1, _ := d1.PowWithPrecision(d2, 2)
	fmt.Println(res1)

	d3 := decimal.NewFromFloat(5.0)
	d4 := decimal.NewFromFloat(5.73)
	res2, _ := d3.PowWithPrecision(d4, 5)
	fmt.Println(res2)

	d5 := decimal.NewFromFloat(-3.0)
	d6 := decimal.NewFromFloat(-6.0)
	res3, _ := d5.PowWithPrecision(d6, 10)
	fmt.Println(res3)

	// output:
	// 256
	// 10118.080371595015625
	// 0.0013717421
}

func ExampleDecimal_PowInt32() {
	d1, _ := decimal.NewFromFloat(4.0).PowInt32(4)
	fmt.Println(d1)

	d2, _ := decimal.NewFromFloat(3.13).PowInt32(5)
	fmt.Println(d2)

	// output:
	// 256
	// 300.4150512793
}

func ExampleDecimal_PowBigInt() {
	d1, _ := decimal.NewFromFloat(3.0).PowBigInt(big.NewInt(3))
	fmt.Println(d1)

	d2, _ := decimal.NewFromFloat(629.25).PowBigInt(big.NewInt(5))
	fmt.Println(d2)

	// output:
	// 27
	// 98654323103449.5673828125
}

func ExampleDecimal_ExpHullAbrham() {
	d1, _ := decimal.NewFromFloat(26.1).ExpHullAbrham(2)
	fmt.Println(d1)

	d2, _ := decimal.NewFromFloat(26.1).ExpHullAbrham(20)
	fmt.Println(d2)

	// output:
	// 220000000000
	// 216314672147.05767284
}

func ExampleDecimal_ExpTaylor() {
	d, _ := decimal.NewFromFloat(26.1).ExpTaylor(2)
	fmt.Println(d)

	d2, _ := decimal.NewFromFloat(26.1).ExpTaylor(20)
	fmt.Println(d2)

	d3, _ := decimal.NewFromFloat(26.1).ExpTaylor(-10)
	fmt.Println(d3)

	// output:
	// 216314672147.06
	// 216314672147.05767284062928674083
	// 220000000000
}

func ExampleDecimal_Ln() {
	d1, _ := decimal.NewFromFloat(13.3).Ln(2)
	fmt.Println(d1)

	d2, _ := decimal.NewFromFloat(579.161).Ln(10)
	fmt.Println(d2)

	// output:
	// 2.59
	// 6.3615805046
}

func ExampleDecimal_StringFixed() {
	d1 := decimal.NewFromFloat(0).StringFixed(2)
	d2 := decimal.NewFromFloat(0).StringFixed(0)
	d3 := decimal.NewFromFloat(5.45).StringFixed(0)
	d4 := decimal.NewFromFloat(5.45).StringFixed(1)
	d5 := decimal.NewFromFloat(5.45).StringFixed(2)
	d6 := decimal.NewFromFloat(5.45).StringFixed(3)
	d7 := decimal.NewFromFloat(545).StringFixed(-1)

	fmt.Println(d1)
	fmt.Println(d2)
	fmt.Println(d3)
	fmt.Println(d4)
	fmt.Println(d5)
	fmt.Println(d6)
	fmt.Println(d7)

	// output:
	// 0.00
	// 0
	// 5
	// 5.5
	// 5.45
	// 5.450
	// 550
}

func ExampleDecimal_StringFixedBank() {
	d1 := decimal.NewFromFloat(0).StringFixedBank(2)
	d2 := decimal.NewFromFloat(0).StringFixedBank(0)
	d3 := decimal.NewFromFloat(5.45).StringFixedBank(0)
	d4 := decimal.NewFromFloat(5.45).StringFixedBank(1)
	d5 := decimal.NewFromFloat(5.45).StringFixedBank(2)
	d6 := decimal.NewFromFloat(5.45).StringFixedBank(3)
	d7 := decimal.NewFromFloat(545).StringFixedBank(-1)

	fmt.Println(d1)
	fmt.Println(d2)
	fmt.Println(d3)
	fmt.Println(d4)
	fmt.Println(d5)
	fmt.Println(d6)
	fmt.Println(d7)

	// output:
	// 0.00
	// 0
	// 5
	// 5.4
	// 5.45
	// 5.450
	// 540
}

func ExampleDecimal_Round() {
	d1 := decimal.NewFromFloat(5.45).Round(1)
	d2 := decimal.NewFromFloat(545).Round(-1)

	fmt.Println(d1)
	fmt.Println(d2)

	// output:
	// 5.5
	// 550
}

func ExampleDecimal_RoundCeil() {
	d1 := decimal.NewFromFloat(545).RoundCeil(-2)
	d2 := decimal.NewFromFloat(500).RoundCeil(-2)
	d3 := decimal.NewFromFloat(1.1001).RoundCeil(2)
	d4 := decimal.NewFromFloat(-1.454).RoundCeil(1)

	fmt.Println(d1)
	fmt.Println(d2)
	fmt.Println(d3)
	fmt.Println(d4)

	// output:
	// 600
	// 500
	// 1.11
	// -1.4
}

func ExampleDecimal_RoundFloor() {
	d1 := decimal.NewFromFloat(545).RoundFloor(-2)
	d2 := decimal.NewFromFloat(500).RoundFloor(-2)
	d3 := decimal.NewFromFloat(1.1001).RoundFloor(2)
	d4 := decimal.NewFromFloat(-1.454).RoundFloor(1)

	fmt.Println(d1)
	fmt.Println(d2)
	fmt.Println(d3)
	fmt.Println(d4)

	// output:
	// 500
	// 500
	// 1.1
	// -1.5
}

func ExampleDecimal_RoundUp() {
	d1 := decimal.NewFromFloat(545).RoundUp(-2)
	d2 := decimal.NewFromFloat(500).RoundUp(-2)
	d3 := decimal.NewFromFloat(1.1001).RoundUp(2)
	d4 := decimal.NewFromFloat(-1.454).RoundUp(1)

	fmt.Println(d1)
	fmt.Println(d2)
	fmt.Println(d3)
	fmt.Println(d4)

	// output:
	// 600
	// 500
	// 1.11
	// -1.5
}

func ExampleDecimal_RoundDown() {
	d1 := decimal.NewFromFloat(545).RoundDown(-2)
	d2 := decimal.NewFromFloat(500).RoundDown(-2)
	d3 := decimal.NewFromFloat(1.1001).RoundDown(2)
	d4 := decimal.NewFromFloat(-1.454).RoundDown(1)

	fmt.Println(d1)
	fmt.Println(d2)
	fmt.Println(d3)
	fmt.Println(d4)

	// output:
	// 500
	// 500
	// 1.1
	// -1.4
}

func ExampleDecimal_RoundBank() {
	d1 := decimal.NewFromFloat(5.45).RoundBank(1)
	d2 := decimal.NewFromFloat(545).RoundBank(-1)
	d3 := decimal.NewFromFloat(5.46).RoundBank(1)
	d4 := decimal.NewFromFloat(546).RoundBank(-1)
	d5 := decimal.NewFromFloat(5.55).RoundBank(1)
	d6 := decimal.NewFromFloat(555).RoundBank(-1)

	fmt.Println(d1)
	fmt.Println(d2)
	fmt.Println(d3)
	fmt.Println(d4)
	fmt.Println(d5)
	fmt.Println(d6)

	// output:
	// 5.4
	// 540
	// 5.5
	// 550
	// 5.6
	// 560
}

func ExampleDecimal_Truncate() {
	d1, _ := decimal.NewFromString("123.456")
	d2 := d1.Truncate(2)

	fmt.Println(d2)

	// output:
	// 123.45
}
