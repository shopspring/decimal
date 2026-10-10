//go:build go1.13
// +build go1.13

package decimal

import (
	"fmt"
	"math"
	"math/big"
	"strconv"
)

// newFromFloat converts val to a Decimal holding the shortest decimal digit
// string that round-trips back to val when parsed at the given bit size.
//
// strconv already computes exactly that string, so we ask it for the value in
// 'e' format with precision -1 and read the digits back out. The 'e' layout is
// [-]d[.ddd]e±dd, which is bounded (at most 24 bytes) and trivial to scan,
// unlike 'f' which can run to over 750 bytes for extreme exponents.
func newFromFloat(val float64, bitSize int) Decimal {
	if math.IsNaN(val) || math.IsInf(val, 0) {
		panic(fmt.Sprintf("Cannot create a Decimal from %v", val))
	}

	var buf [32]byte
	b := strconv.AppendFloat(buf[:0], val, 'e', -1, bitSize)

	neg := b[0] == '-'
	if neg {
		b = b[1:]
	}

	// Split the significand from the exponent. b always contains an 'e'.
	var i int
	for b[i] != 'e' {
		i++
	}
	digits, expDigits := b[:i], b[i+1:]

	// The shortest round-trip form has at most 17 significant digits for a
	// float64 and 9 for a float32, so the significand always fits in an int64.
	var mant int64
	nd := 0
	for _, c := range digits {
		if c == '.' {
			continue
		}
		mant = mant*10 + int64(c-'0')
		nd++
	}
	if neg {
		mant = -mant
	}

	esign := 1
	switch expDigits[0] {
	case '-':
		esign = -1
		expDigits = expDigits[1:]
	case '+':
		expDigits = expDigits[1:]
	}
	e := 0
	for _, c := range expDigits {
		e = e*10 + int(c-'0')
	}

	// digits is d.ddd, i.e. the significand scaled by 10^(nd-1).
	return Decimal{value: big.NewInt(mant), exp: int32(esign*e - nd + 1)}
}
