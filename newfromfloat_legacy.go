//go:build !go1.13
// +build !go1.13

package decimal

import (
	"fmt"
	"math"
	"math/big"
)

// newFromFloat converts val to a Decimal holding the shortest decimal digit
// string that round-trips back to val when parsed at the given bit size.
//
// This is the pre-Go 1.13 implementation: it computes the shortest form itself,
// using the vendored copy of strconv's multiprecision decimal (decimal-go.go and
// rounding.go), because strconv.FormatFloat could return a shortest form that is
// not the nearest one before Go 1.13 (golang/go#29491).
func newFromFloat(val float64, bitSize int) Decimal {
	if math.IsNaN(val) || math.IsInf(val, 0) {
		panic(fmt.Sprintf("Cannot create a Decimal from %v", val))
	}

	var bits uint64
	var flt *floatInfo
	if bitSize == 32 {
		// XOR is workaround for https://github.com/golang/go/issues/26285
		a := math.Float32bits(float32(val)) ^ 0x80808080
		bits = uint64(a) ^ 0x80808080
		flt = &float32info
	} else {
		bits = math.Float64bits(val)
		flt = &float64info
	}

	exp := int(bits>>flt.mantbits) & (1<<flt.expbits - 1)
	mant := bits & (uint64(1)<<flt.mantbits - 1)

	switch exp {
	case 0:
		// denormalized
		exp++

	default:
		// add implicit top bit
		mant |= uint64(1) << flt.mantbits
	}
	exp += flt.bias

	var d decimal
	d.Assign(mant)
	d.Shift(exp - int(flt.mantbits))
	d.neg = bits>>(flt.expbits+flt.mantbits) != 0

	roundShortest(&d, mant, exp, flt)
	// If less than 19 digits, we can do calculation in an int64.
	if d.nd < 19 {
		tmp := int64(0)
		m := int64(1)
		for i := d.nd - 1; i >= 0; i-- {
			tmp += m * int64(d.d[i]-'0')
			m *= 10
		}
		if d.neg {
			tmp *= -1
		}
		return Decimal{value: big.NewInt(tmp), exp: int32(d.dp) - int32(d.nd)}
	}
	dValue := new(big.Int)
	dValue, ok := dValue.SetString(string(d.d[:d.nd]), 10)
	if ok {
		return Decimal{value: dValue, exp: int32(d.dp) - int32(d.nd)}
	}

	return NewFromFloatWithExponent(val, int32(d.dp)-int32(d.nd))
}
