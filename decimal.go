// Package decimal implements an arbitrary precision fixed-point decimal.
//
// The zero-value of a Decimal is 0, as you would expect.
//
// The best way to create a new Decimal is to use decimal.NewFromString, ex:
//
//	n, err := decimal.NewFromString("-123.4567")
//	n.String() // output: "-123.4567"
//
// To use Decimal as part of a struct:
//
//	type StructName struct {
//	    Number Decimal
//	}
//
// Note: This can "only" represent numbers with a maximum of 2^31 digits after the decimal point.
package decimal

import (
	"database/sql/driver"
	"encoding/binary"
	"fmt"
	"math"
	"math/big"
	"math/bits"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

// DivisionPrecision is the number of decimal places in the result when it
// doesn't divide exactly.
//
// Example:
//
//	d1 := decimal.NewFromFloat(2).Div(decimal.NewFromFloat(3))
//	d1.String() // output: "0.6666666666666667"
//	d2 := decimal.NewFromFloat(2).Div(decimal.NewFromFloat(30000))
//	d2.String() // output: "0.0000666666666667"
//	d3 := decimal.NewFromFloat(20000).Div(decimal.NewFromFloat(3))
//	d3.String() // output: "6666.6666666666666667"
//	decimal.DivisionPrecision = 3
//	d4 := decimal.NewFromFloat(2).Div(decimal.NewFromFloat(3))
//	d4.String() // output: "0.667"
var DivisionPrecision = 16

// PowPrecisionNegativeExponent specifies the number of digits after the decimal point in the results of
// Pow, PowInt32 and PowBigInt for negative or non-integer exponents. A result that would round to 0 keeps
// PowPrecisionNegativeExponent significant digits instead (at least one).
// PowWithPrecision is not affected, it takes the precision as an argument.
//
// Example:
//
//	d1, err := decimal.NewFromFloat(15.2).PowInt32(-2)
//	d1.String() // output: "0.0043282548476454"
//
//	decimal.PowPrecisionNegativeExponent = 24
//	d2, err := decimal.NewFromFloat(15.2).PowInt32(-2)
//	d2.String() // output: "0.004328254847645429362881"
var PowPrecisionNegativeExponent = 16

// MarshalJSONWithoutQuotes should be set to true if you want the decimal to
// be JSON marshaled as a number, instead of as a string.
// WARNING: this is dangerous for decimals with many digits, since many JSON
// unmarshallers (ex: Javascript's) will unmarshal JSON numbers to IEEE 754
// double-precision floating point numbers, which means you can potentially
// silently lose precision.
var MarshalJSONWithoutQuotes = false

// TrimTrailingZeros specifies whether trailing zeroes should be trimmed from a string representation of decimal.
// If set to true, trailing zeroes will be truncated (2.00 -> 2, 3.11 -> 3.11, 13.000 -> 13),
// otherwise trailing zeroes will be preserved (2.00 -> 2.00, 3.11 -> 3.11, 13.000 -> 13.000).
// Setting this value to false can be useful for APIs where exact decimal string representation matters.
var TrimTrailingZeros = true

// UseScientificNotation specifies whether scientific notation should be used when a decimal is turned
// into a string that has a "negative" precision.
//
// For example, 1200 rounded to the nearest 100 cannot accurately be shown as "1200" because the last two
// digits are unknown. With this set to true, that number would be expressed as "1.2E3" instead.
var UseScientificNotation = false

// ExpMaxIterations specifies the maximum number of iterations needed to calculate
// precise natural exponent value using ExpHullAbrham method.
var ExpMaxIterations = 1000

// MaxDecodeExponent is the largest exponent magnitude that NewFromString and UnmarshalBinary accept,
// and so UnmarshalJSON, UnmarshalText, Scan, GobDecode, DecodeSpanner and their NullDecimal variants too.
// Operations like String, Add and Float64 take time and memory that grow with the exponent, so without
// this limit a short input like "1e-2000000000" can stall the process or exhaust its memory.
// The default covers the float64 and IEEE 754 decimal128 exponent ranges.
//
// Values created with New, NewFromBigInt or arithmetic are not checked, so they may not decode back from
// their own String or MarshalBinary output. The limit does not bound the cost of Pow or ExpTaylor
// arguments or of very long inputs; validate those separately.
//
// Set it once at program start, before any decoding. Raise it only for trusted input.
var MaxDecodeExponent = 10000

// Zero constant, to make computations faster.
// Zero should never be compared with == or != directly, please use decimal.Equal or decimal.Cmp instead.
var Zero = Decimal{}

var zeroInt = big.NewInt(0)
var oneInt = big.NewInt(1)
var twoInt = big.NewInt(2)
var fourInt = big.NewInt(4)
var fiveInt = big.NewInt(5)
var tenInt = big.NewInt(10)
var twentyInt = big.NewInt(20)

var factorials = []Decimal{New(1, 0)}
var factorialsMutex sync.RWMutex

// Decimal represents a fixed-point decimal. It is immutable.
// number = value * 10 ^ exp
type Decimal struct {
	value *big.Int

	// NOTE(vadim): this must be an int32, because we cast it to float64 during
	// calculations. If exp is 64 bit, we might lose precision.
	// If we cared about being able to represent every possible decimal, we
	// could make exp a *big.Int but it would hurt performance and numbers
	// like that are unrealistic.
	exp int32
}

func (d Decimal) getValue() *big.Int {
	if d.value == nil {
		return zeroInt
	}
	return d.value
}

// New returns a new fixed-point decimal, value * 10 ^ exp.
func New(value int64, exp int32) Decimal {
	return Decimal{
		value: big.NewInt(value),
		exp:   exp,
	}
}

// NewFromInt converts an int64 to Decimal.
//
// Example:
//
//	NewFromInt(123).String() // output: "123"
//	NewFromInt(-10).String() // output: "-10"
func NewFromInt(value int64) Decimal {
	return Decimal{
		value: big.NewInt(value),
		exp:   0,
	}
}

// NewFromInt32 converts an int32 to Decimal.
//
// Example:
//
//	NewFromInt(123).String() // output: "123"
//	NewFromInt(-10).String() // output: "-10"
func NewFromInt32(value int32) Decimal {
	return Decimal{
		value: big.NewInt(int64(value)),
		exp:   0,
	}
}

// NewFromUint64 converts an uint64 to Decimal.
//
// Example:
//
//	NewFromUint64(123).String() // output: "123"
func NewFromUint64(value uint64) Decimal {
	return Decimal{
		value: new(big.Int).SetUint64(value),
		exp:   0,
	}
}

// NewFromBigInt returns a new Decimal from a big.Int, value * 10 ^ exp
func NewFromBigInt(value *big.Int, exp int32) Decimal {
	return Decimal{
		value: new(big.Int).Set(value),
		exp:   exp,
	}
}

// NewFromBigRat returns a new Decimal from a big.Rat. The numerator and
// denominator are divided and rounded to the given precision.
//
// Example:
//
//	d1 := NewFromBigRat(big.NewRat(0, 1), 0)    // output: "0"
//	d2 := NewFromBigRat(big.NewRat(4, 5), 1)    // output: "0.8"
//	d3 := NewFromBigRat(big.NewRat(1000, 3), 3) // output: "333.333"
//	d4 := NewFromBigRat(big.NewRat(2, 7), 4)    // output: "0.2857"
func NewFromBigRat(value *big.Rat, precision int32) Decimal {
	return Decimal{
		value: new(big.Int).Set(value.Num()),
		exp:   0,
	}.DivRound(Decimal{
		value: new(big.Int).Set(value.Denom()),
		exp:   0,
	}, precision)
}

// NewFromString returns a new Decimal from a string representation.
// Trailing zeroes are not trimmed.
//
// Example:
//
//	d, err := NewFromString("-123.45")
//	d2, err := NewFromString(".0001")
//	d3, err := NewFromString("1.47000")
func NewFromString(value string) (Decimal, error) {
	originalInput := value
	var exp int64

	// Check if number is using scientific notation and find dots
	eIndex := -1
	pIndex := -1
	for i, r := range value {
		if r == 'E' || r == 'e' {
			if eIndex > -1 {
				return Decimal{}, fmt.Errorf("can't convert %s to decimal: multiple 'E' characters found", value)
			}
			eIndex = i
			continue
		}

		if r == '.' {
			if pIndex > -1 {
				return Decimal{}, fmt.Errorf("can't convert %s to decimal: too many .s", value)
			}
			pIndex = i
		}
	}

	if eIndex != -1 {
		expInt, err := strconv.ParseInt(value[eIndex+1:], 10, 32)
		if err != nil {
			if e, ok := err.(*strconv.NumError); ok && e.Err == strconv.ErrRange {
				return Decimal{}, fmt.Errorf("can't convert %s to decimal: fractional part too long", value)
			}
			return Decimal{}, fmt.Errorf("can't convert %s to decimal: exponent is not numeric", value)
		}
		value = value[:eIndex]
		exp = expInt
	}

	numLen := len(value)
	if pIndex != -1 {
		if pIndex+1 < len(value) && (value[pIndex+1] == '-' || value[pIndex+1] == '+') {
			// ParseInt and SetString would accept the sign once the point is removed, as in ".-5"
			return Decimal{}, fmt.Errorf("can't convert %s to decimal", value)
		}
		numLen--
		expInt := -len(value[pIndex+1:])
		exp += int64(expInt)
	}

	var dValue *big.Int
	// parsing in an int64 is faster than new(big.Int).SetString so this is just a shortcut for strings we know won't overflow
	if numLen <= 18 {
		parsed64, ok := parseInt64SkipIndex(value, pIndex)
		if !ok {
			return Decimal{}, fmt.Errorf("can't convert %s to decimal", value)
		}
		dValue = big.NewInt(parsed64)
	} else {
		intString := value
		if pIndex != -1 {
			intString = value[:pIndex] + value[pIndex+1:]
		}
		dValue = new(big.Int)
		_, ok := dValue.SetString(intString, 10)
		if !ok {
			return Decimal{}, fmt.Errorf("can't convert %s to decimal", value)
		}
	}

	if exp < math.MinInt32 || exp > math.MaxInt32 {
		// NOTE(vadim): I doubt a string could realistically be this long
		return Decimal{}, fmt.Errorf("can't convert %s to decimal: fractional part too long", originalInput)
	}
	if exp > int64(MaxDecodeExponent) || exp < -int64(MaxDecodeExponent) {
		return Decimal{}, fmt.Errorf("can't convert %s to decimal: exponent %d exceeds MaxDecodeExponent (%d)", originalInput, exp, MaxDecodeExponent)
	}

	return Decimal{
		value: dValue,
		exp:   int32(exp),
	}, nil
}

// parseInt64SkipIndex parses s without the byte at index skip as strconv.ParseInt(s, 10, 64) would,
// s must have at most 18 other bytes so that the result can't overflow.
func parseInt64SkipIndex(s string, skip int) (int64, bool) {
	i, neg := 0, false
	if len(s) > 0 && skip != 0 && (s[0] == '+' || s[0] == '-') {
		i, neg = 1, s[0] == '-'
	}
	var n int64
	digits := 0
	for ; i < len(s); i++ {
		if i == skip {
			continue
		}
		c := s[i]
		if c < '0' || c > '9' {
			return 0, false
		}
		n = n*10 + int64(c-'0')
		digits++
	}
	if neg {
		n = -n
	}
	return n, digits > 0
}

// NewFromFormattedString returns a new Decimal from a formatted string representation.
// The second argument - replRegexp, is a regular expression that is used to find characters that should be
// removed from given decimal string representation. All matched characters will be replaced with an empty string.
//
// Example:
//
//	r := regexp.MustCompile("[$,]")
//	d1, err := NewFromFormattedString("$5,125.99", r)
//
//	r2 := regexp.MustCompile("[_]")
//	d2, err := NewFromFormattedString("1_000_000", r2)
//
//	r3 := regexp.MustCompile("[USD\\s]")
//	d3, err := NewFromFormattedString("5000 USD", r3)
func NewFromFormattedString(value string, replRegexp *regexp.Regexp) (Decimal, error) {
	parsedValue := replRegexp.ReplaceAllString(value, "")
	d, err := NewFromString(parsedValue)
	if err != nil {
		return Decimal{}, err
	}
	return d, nil
}

// RequireFromString returns a new Decimal from a string representation
// or panics if NewFromString had returned an error.
//
// Example:
//
//	d := RequireFromString("-123.45")
//	d2 := RequireFromString(".0001")
func RequireFromString(value string) Decimal {
	dec, err := NewFromString(value)
	if err != nil {
		panic(err)
	}
	return dec
}

// NewFromFloat converts a float64 to Decimal.
//
// The result is the shortest decimal that converts back to the same float64,
// the same number that strconv.FormatFloat(value, 'f', -1, 64) prints.
// This is typically 15 digits, but may be more in some cases.
// See https://www.exploringbinary.com/decimal-precision-of-binary-floating-point-numbers/ for more information.
//
// This is not always the exact value stored in the float. Whole numbers larger
// than 2^53 can change: float64(1<<62) is exactly 4611686018427387904, but
// NewFromFloat returns 4611686018427388000. To convert whole numbers exactly,
// use NewFromFloatWithExponent(value, 0).
//
// For slightly faster conversion, use NewFromFloatWithExponent where you can specify the precision in absolute terms.
//
// NOTE: this will panic on NaN, +/-inf
func NewFromFloat(value float64) Decimal {
	if value == 0 {
		return New(0, 0)
	}
	return newFromFloat(value, math.Float64bits(value), &float64info)
}

// NewFromFloat32 converts a float32 to Decimal.
//
// The result is the shortest decimal that converts back to the same float32.
// This is typically 6-8 digits depending on the input.
// See https://www.exploringbinary.com/decimal-precision-of-binary-floating-point-numbers/ for more information.
//
// As with NewFromFloat, whole numbers larger than 2^24 can change. To convert
// them exactly, use NewFromFloatWithExponent(float64(value), 0).
//
// For slightly faster conversion, use NewFromFloatWithExponent where you can specify the precision in absolute terms.
//
// NOTE: this will panic on NaN, +/-inf
func NewFromFloat32(value float32) Decimal {
	if value == 0 {
		return New(0, 0)
	}
	// XOR is workaround for https://github.com/golang/go/issues/26285
	a := math.Float32bits(value) ^ 0x80808080
	return newFromFloat(float64(value), uint64(a)^0x80808080, &float32info)
}

func newFromFloat(val float64, bits uint64, flt *floatInfo) Decimal {
	if math.IsNaN(val) || math.IsInf(val, 0) {
		panic(fmt.Sprintf("Cannot create a Decimal from %v", val))
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

// NewFromFloatWithExponent converts a float64 to Decimal, with an arbitrary
// number of fractional digits.
//
// Example:
//
//	NewFromFloatWithExponent(123.456, -2).String() // output: "123.46"
func NewFromFloatWithExponent(value float64, exp int32) Decimal {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		panic(fmt.Sprintf("Cannot create a Decimal from %v", value))
	}

	bits := math.Float64bits(value)
	mant := bits & (1<<52 - 1)
	exp2 := int32((bits >> 52) & (1<<11 - 1))
	sign := bits >> 63

	if exp2 == 0 {
		// specials
		if mant == 0 {
			return Decimal{}
		}
		// subnormal
		exp2++
	} else {
		// normal
		mant |= 1 << 52
	}

	exp2 -= 1023 + 52

	// normalizing base-2 values
	for mant&1 == 0 {
		mant = mant >> 1
		exp2++
	}

	// maximum number of fractional base-10 digits to represent 2^N exactly cannot be more than -N if N<0
	if exp < 0 && exp < exp2 {
		if exp2 < 0 {
			exp = exp2
		} else {
			exp = 0
		}
	}

	// representing 10^M * 2^N as 5^M * 2^(M+N)
	exp2 -= exp

	temp := big.NewInt(1)
	dMant := big.NewInt(int64(mant))

	// applying 5^M
	if exp > 0 {
		temp = temp.SetInt64(int64(exp))
		temp = temp.Exp(fiveInt, temp, nil)
	} else if exp < 0 {
		temp = temp.SetInt64(-int64(exp))
		temp = temp.Exp(fiveInt, temp, nil)
		dMant = dMant.Mul(dMant, temp)
		temp = temp.SetUint64(1)
	}

	// applying 2^(M+N)
	if exp2 > 0 {
		dMant = dMant.Lsh(dMant, uint(exp2))
	} else if exp2 < 0 {
		temp = temp.Lsh(temp, uint(-exp2))
	}

	// rounding and downscaling
	if exp > 0 || exp2 < 0 {
		halfDown := new(big.Int).Rsh(temp, 1)
		dMant = dMant.Add(dMant, halfDown)
		dMant = dMant.Quo(dMant, temp)
	}

	if sign == 1 {
		dMant = dMant.Neg(dMant)
	}

	return Decimal{
		value: dMant,
		exp:   exp,
	}
}

// Copy returns a copy of decimal with the same value and exponent, but a different pointer to value.
func (d Decimal) Copy() Decimal {
	return Decimal{
		value: new(big.Int).Set(d.getValue()),
		exp:   d.exp,
	}
}

// rescale returns a rescaled version of the decimal. Returned
// decimal may be less precise if the given exponent is bigger
// than the initial exponent of the Decimal.
// NOTE: this will truncate, NOT round
//
// Example:
//
//	d := New(12345, -4)
//	d2 := d.rescale(-1)
//	d3 := d2.rescale(-4)
//	println(d1)
//	println(d2)
//	println(d3)
//
// Output:
//
//	1.2345
//	1.2
//	1.2000
func (d Decimal) rescale(exp int32) Decimal {
	if d.exp == exp {
		return Decimal{
			new(big.Int).Set(d.getValue()),
			d.exp,
		}
	}

	// NOTE(vadim): must convert exps to float64 before - to prevent overflow
	diff := math.Abs(float64(exp) - float64(d.exp))
	value := new(big.Int)
	if exp > d.exp {
		value.Quo(d.getValue(), pow10(int64(diff)))
		if value.Sign() == 0 && d.getValue().Sign() != 0 {
			// reflect.DeepEqual tells an empty word slice from nil, keep the empty one that
			// dividing a copy of d.value in place used to leave
			value.SetBits([]big.Word{})
		}
	} else {
		value.Mul(d.getValue(), pow10(int64(diff)))
	}

	return Decimal{
		value: value,
		exp:   exp,
	}
}

// Abs returns the absolute value of the decimal.
func (d Decimal) Abs() Decimal {
	if !d.IsNegative() {
		return d
	}
	d2Value := new(big.Int).Abs(d.getValue())
	return Decimal{
		value: d2Value,
		exp:   d.exp,
	}
}

// Add returns d + d2.
func (d Decimal) Add(d2 Decimal) Decimal {
	// rescale returns a new value, so the sum can be stored in it
	if d.exp < d2.exp {
		r := d2.rescale(d.exp)
		r.value.Add(d.getValue(), r.value)
		return r
	} else if d.exp > d2.exp {
		r := d.rescale(d2.exp)
		r.value.Add(r.value, d2.getValue())
		return r
	}

	d3Value := new(big.Int).Add(d.getValue(), d2.getValue())
	return Decimal{
		value: d3Value,
		exp:   d.exp,
	}
}

// Sub returns d - d2.
func (d Decimal) Sub(d2 Decimal) Decimal {
	// rescale returns a new value, so the difference can be stored in it
	if d.exp < d2.exp {
		r := d2.rescale(d.exp)
		r.value.Sub(d.getValue(), r.value)
		return r
	} else if d.exp > d2.exp {
		r := d.rescale(d2.exp)
		r.value.Sub(r.value, d2.getValue())
		return r
	}

	d3Value := new(big.Int).Sub(d.getValue(), d2.getValue())
	return Decimal{
		value: d3Value,
		exp:   d.exp,
	}
}

// Neg returns -d.
func (d Decimal) Neg() Decimal {
	val := new(big.Int).Neg(d.getValue())
	return Decimal{
		value: val,
		exp:   d.exp,
	}
}

// Mul returns d * d2.
func (d Decimal) Mul(d2 Decimal) Decimal {
	expInt64 := int64(d.exp) + int64(d2.exp)
	if expInt64 > math.MaxInt32 || expInt64 < math.MinInt32 {
		// NOTE(vadim): better to panic than give incorrect results, as
		// Decimals are usually used for money
		panic(fmt.Sprintf("exponent %v overflows an int32!", expInt64))
	}

	d3Value := new(big.Int).Mul(d.getValue(), d2.getValue())
	return Decimal{
		value: d3Value,
		exp:   int32(expInt64),
	}
}

// Shift shifts the decimal in base 10.
// It shifts left when shift is positive and right if shift is negative.
// In simpler terms, the given value for shift is added to the exponent
// of the decimal.
// Shift panics if the resulting exponent does not fit in an int32.
func (d Decimal) Shift(shift int32) Decimal {
	exp := int64(d.exp) + int64(shift)
	if exp > math.MaxInt32 || exp < math.MinInt32 {
		panic(fmt.Sprintf("exponent %v overflows an int32!", exp))
	}
	value := d.getValue()
	if value.Sign() == 0 {
		// a copy of zero has a nil word slice, which reflect.DeepEqual tells from an empty one
		value = zeroInt
	}
	return Decimal{
		value: value,
		exp:   int32(exp),
	}
}

// Div returns d / d2. If it doesn't divide exactly, the result will have
// DivisionPrecision digits after the decimal point.
func (d Decimal) Div(d2 Decimal) Decimal {
	return d.DivRound(d2, int32(DivisionPrecision))
}

// QuoRem does division with remainder
// d.QuoRem(d2,precision) returns quotient q and remainder r such that
//
//	d = d2 * q + r, q an integer multiple of 10^(-precision)
//	0 <= r < abs(d2) * 10 ^(-precision) if d>=0
//	0 >= r > -abs(d2) * 10 ^(-precision) if d<0
//
// Note that precision<0 is allowed as input.
func (d Decimal) QuoRem(d2 Decimal, precision int32) (Decimal, Decimal) {
	if d2.getValue().Sign() == 0 {
		panic("decimal division by 0")
	}
	scale := -precision
	e := int64(d.exp) - int64(d2.exp) - int64(scale)
	if e > math.MaxInt32 || e < math.MinInt32 {
		panic("overflow in decimal QuoRem")
	}
	var aa, bb, expo big.Int
	var scalerest int32
	// d = a 10^ea
	// d2 = b 10^eb
	if e < 0 {
		aa = *d.getValue()
		expo.SetInt64(-e)
		bb.Exp(tenInt, &expo, nil)
		bb.Mul(d2.getValue(), &bb)
		scalerest = d.exp
		// now aa = a
		//     bb = b 10^(scale + eb - ea)
	} else {
		expo.SetInt64(e)
		aa.Exp(tenInt, &expo, nil)
		aa.Mul(d.getValue(), &aa)
		bb = *d2.getValue()
		scalerest = scale + d2.exp
		// now aa = a ^ (ea - eb - scale)
		//     bb = b
	}
	var q, r big.Int
	q.QuoRem(&aa, &bb, &r)
	dq := Decimal{value: &q, exp: scale}
	dr := Decimal{value: &r, exp: scalerest}
	return dq, dr
}

// DivRound divides and rounds to a given precision
// i.e. to an integer multiple of 10^(-precision)
//
//	for a positive quotient digit 5 is rounded up, away from 0
//	if the quotient is negative then digit 5 is rounded down, away from 0
//
// Note that precision<0 is allowed as input.
func (d Decimal) DivRound(d2 Decimal, precision int32) Decimal {
	// QuoRem already checks initialization
	q, r := d.QuoRem(d2, precision)
	// the actual rounding decision is based on comparing r*10^precision and d2/2
	// instead compare 2 r 10 ^precision and d2
	var rv2 big.Int
	rv2.Abs(r.getValue())
	rv2.Lsh(&rv2, 1)
	// now rv2 = abs(r.value) * 2
	r2 := Decimal{value: &rv2, exp: r.exp + precision}
	// r2 is now 2 * r * 10 ^ precision
	var c = r2.Cmp(d2.Abs())

	if c < 0 {
		return q
	}

	if d.getValue().Sign()*d2.getValue().Sign() < 0 {
		return q.Sub(New(1, -precision))
	}

	return q.Add(New(1, -precision))
}

// Mod returns d % d2.
func (d Decimal) Mod(d2 Decimal) Decimal {
	_, r := d.QuoRem(d2, 0)
	return r
}

// Pow returns d to the power of d2.
// The result is exact for non-negative integer exponents. For negative or non-integer exponents it is
// rounded half away from zero to PowPrecisionNegativeExponent places after the decimal point. A result
// that would round to 0 keeps PowPrecisionNegativeExponent significant digits instead (at least one).
//
// Pow returns 0 (zero-value of Decimal) instead of error for power operation edge cases, to handle those edge cases use PowWithPrecision.
// Edge cases not handled by Pow:
//   - 0 ** 0 => undefined value
//   - 0 ** y, where y < 0 => infinity
//   - x ** y, where x < 0 and y is non-integer decimal => imaginary value
//   - x ** y, where the magnitude of the result or the precision is too large to represent
//
// Example:
//
//	d1 := decimal.NewFromFloat(4.0)
//	d2 := decimal.NewFromFloat(4.0)
//	res1 := d1.Pow(d2)
//	res1.String() // output: "256"
//
//	d3 := decimal.NewFromFloat(5.0)
//	d4 := decimal.NewFromFloat(5.73)
//	res2 := d3.Pow(d4)
//	res2.String() // output: "10118.0803715950193171"
//
//	d5 := decimal.NewFromInt(10)
//	d6 := decimal.NewFromInt(-18)
//	res3 := d5.Pow(d6)
//	res3.String() // output: "0.000000000000000001"
func (d Decimal) Pow(d2 Decimal) Decimal {
	res, err := d.PowWithPrecision(d2, int32(PowPrecisionNegativeExponent))
	if err != nil {
		return Decimal{}
	}
	return res
}

// PowWithPrecision returns d to the power of d2.
// The result is exact for non-negative integer exponents. For negative or non-integer exponents it is
// rounded half away from zero to precision places after the decimal point. A result that would round
// to 0 keeps precision significant digits instead (at least one).
//
// PowWithPrecision returns error when:
//   - 0 ** 0 => undefined value
//   - 0 ** y, where y < 0 => infinity
//   - x ** y, where x < 0 and y is non-integer decimal => imaginary value
//   - x ** y, where the magnitude of the result or the precision is too large to represent
//
// Example:
//
//	d1 := decimal.NewFromFloat(4.0)
//	d2 := decimal.NewFromFloat(4.0)
//	res1, err := d1.PowWithPrecision(d2, 2)
//	res1.String() // output: "256"
//
//	d3 := decimal.NewFromFloat(5.0)
//	d4 := decimal.NewFromFloat(5.73)
//	res2, err := d3.PowWithPrecision(d4, 5)
//	res2.String() // output: "10118.08037"
//
//	d5 := decimal.NewFromFloat(-3.0)
//	d6 := decimal.NewFromFloat(-6.0)
//	res3, err := d5.PowWithPrecision(d6, 10)
//	res3.String() // output: "0.0013717421"
func (d Decimal) PowWithPrecision(d2 Decimal, precision int32) (Decimal, error) {
	if d.IsZero() {
		if d2.IsZero() {
			return Decimal{}, fmt.Errorf("cannot represent undefined value of 0**0")
		}
		if d2.Sign() < 0 {
			return Decimal{}, fmt.Errorf("cannot represent infinity value of 0 ** y, where y < 0")
		}
		return Decimal{zeroInt, 0}, nil
	}

	if d2.IsInteger() {
		return d.powBigIntWithPrecision(d2.BigInt(), precision)
	}

	if d.Sign() < 0 {
		return Decimal{}, fmt.Errorf("cannot represent imaginary value of x ** y, where x < 0 and y is non-integer decimal")
	}

	return d.powFrac(d2, precision)
}

// PowInt32 returns d to the power of exp, where exp is int32.
// Returns error for 0 ** 0, 0 ** exp where exp < 0, and results whose magnitude is too large to represent.
//
// For negative exponents the result is rounded half away from zero to PowPrecisionNegativeExponent places
// after the decimal point. A result that would round to 0 keeps PowPrecisionNegativeExponent significant
// digits instead (at least one).
//
// Example:
//
//	d1, err := decimal.NewFromFloat(4.0).PowInt32(4)
//	d1.String() // output: "256"
//
//	d2, err := decimal.NewFromFloat(3.13).PowInt32(5)
//	d2.String() // output: "300.4150512793"
func (d Decimal) PowInt32(exp int32) (Decimal, error) {
	return d.PowBigInt(big.NewInt(int64(exp)))
}

// PowBigInt returns d to the power of exp, where exp is big.Int.
// Returns error for 0 ** 0, 0 ** exp where exp < 0, and results whose magnitude is too large to represent.
//
// For negative exponents the result is rounded half away from zero to PowPrecisionNegativeExponent places
// after the decimal point. A result that would round to 0 keeps PowPrecisionNegativeExponent significant
// digits instead (at least one).
//
// Example:
//
//	d1, err := decimal.NewFromFloat(3.0).PowBigInt(big.NewInt(3))
//	d1.String() // output: "27"
//
//	d2, err := decimal.NewFromFloat(629.25).PowBigInt(big.NewInt(5))
//	d2.String() // output: "98654323103449.5673828125"
func (d Decimal) PowBigInt(exp *big.Int) (Decimal, error) {
	return d.powBigIntWithPrecision(exp, int32(PowPrecisionNegativeExponent))
}

var errPowOutOfRange = fmt.Errorf("cannot represent result of power operation, its magnitude or precision is too large")

// powExactBits is the largest estimated cost of computing d^n exactly for negative n.
// Bigger powers are approximated with truncated intermediates, which is cheaper.
const powExactBits = 3000

func (d Decimal) powBigIntWithPrecision(exp *big.Int, precision int32) (Decimal, error) {
	switch {
	case exp.Sign() > 0:
		return d.powExact(exp)
	case d.IsZero() && exp.Sign() == 0:
		return Decimal{}, fmt.Errorf("cannot represent undefined value of 0**0")
	case exp.Sign() == 0:
		return Decimal{oneInt, 0}, nil
	case d.IsZero():
		return Decimal{}, fmt.Errorf("cannot represent infinity value of 0 ** y, where y < 0")
	}
	return d.powNegInt(new(big.Int).Neg(exp), precision)
}

// powExact returns d^n for n > 0.
func (d Decimal) powExact(n *big.Int) (Decimal, error) {
	exp := new(big.Int).Mul(n, big.NewInt(int64(d.exp)))
	if !exp.IsInt64() || exp.Int64() > math.MaxInt32 || exp.Int64() < math.MinInt32 {
		return Decimal{}, errPowOutOfRange
	}
	return Decimal{new(big.Int).Exp(d.getValue(), n, nil), int32(exp.Int64())}, nil
}

// powNegInt returns d^-n for n > 0, rounded as described in PowWithPrecision.
func (d Decimal) powNegInt(n *big.Int, precision int32) (Decimal, error) {
	if d.absIsOne() {
		// only the parity of n matters
		n = big.NewInt(2 - int64(n.Bit(0)))
	}

	// small powers are cheap to compute exactly, and |log10(d^-n)| <= cost keeps them in range
	if n.IsInt64() && n.Int64() <= powExactBits {
		cost := n.Int64() * (int64(d.getValue().BitLen()) + abs64(int64(d.exp)))
		if cost <= powExactBits && powInRange(float64(cost), precision) {
			return d.powNegIntExact(n, precision)
		}
	}

	nf, _ := new(big.Float).SetInt(n).Float64()
	lg := -nf * d.log10Abs()
	if !powInRange(lg, precision) {
		return Decimal{}, errPowOutOfRange
	}

	// the digit shifts of the truncated c^n must fit into int64
	if nd := int64(d.NumDigits()); n.IsInt64() && n.Int64() <= math.MaxInt64/(4*nd) {
		c := new(big.Int).Abs(d.getValue())
		w := powSigDigits(lg, precision) + int64(len(n.String())) + 6
		// each try adds 20 digits, failing 5 times means an exact tie or a value extremely close to one
		for i := 0; i < 5; i, w = i+1, w+20 {
			if res, ok := powNegIntApprox(c, int64(d.exp), n, w, precision); ok {
				if d.Sign() < 0 && n.Bit(0) == 1 {
					res.value.Neg(res.value)
				}
				return res, nil
			}
		}
	}

	// same as for non-integer exponents, which handles any n and exact ties
	res, err := d.Abs().powFrac(Decimal{value: new(big.Int).Neg(n)}, precision)
	if err == nil && d.Sign() < 0 && n.Bit(0) == 1 {
		res.value.Neg(res.value)
	}
	return res, err
}

// powNegIntExact returns d^-n for n > 0, dividing by the exact d^n.
func (d Decimal) powNegIntExact(n *big.Int, precision int32) (Decimal, error) {
	x, err := d.powExact(n)
	if err != nil {
		return Decimal{}, err
	}
	res := New(1, 0).DivRound(x, precision)
	if res.IsZero() {
		// first significant digit of 1/x is at place NumDigits+exp, one earlier when x is a power of ten
		nd := int32(x.NumDigits())
		places := int32(powMinSig(precision)) - 1 + nd + x.exp
		if x.getValue().CmpAbs(pow10(int64(nd)-1)) == 0 {
			places--
		}
		res = New(1, 0).DivRound(x, places)
	}
	return res, nil
}

// powNegIntApprox rounds |1 / (c^n * 10^(e*n))| as described in PowWithPrecision. c^n is approximated
// from below by m*10^s, truncating m to w digits after every multiplication. Each truncation shrinks
// the result by a factor of at most 1-10^(1-w), and the errors add up to less than 5n*10^(1-w),
// lo below allows ten times that.
func powNegIntApprox(c *big.Int, e int64, n *big.Int, w int64, precision int32) (Decimal, bool) {
	cw := new(big.Int).Set(c)
	cs, exact := truncDigits(cw, w)

	m, s := big.NewInt(1), int64(0)
	for i := n.BitLen() - 1; i >= 0; i-- {
		m.Mul(m, m)
		ds, ex := truncDigits(m, w)
		s, exact = 2*s+ds, exact && ex
		if n.Bit(i) == 1 {
			m.Mul(m, cw)
			ds, ex = truncDigits(m, w)
			s, exact = s+cs+ds, exact && ex
		}
	}

	// 10^scale / (m * 10^(s+e*n)) = 10^z / m
	z := int64(Decimal{value: m}.NumDigits()) + w + 2
	scale := z + s + e*n.Int64()
	q, rem := new(big.Int).QuoRem(pow10(z), m, new(big.Int))
	lo, hi := q, q
	if !exact || rem.Sign() != 0 {
		hi = new(big.Int).Add(q, oneInt)
	}
	if !exact {
		lo = new(big.Int).Mul(q, big.NewInt(5))
		lo.Mul(lo, n).Quo(lo, pow10(w-2))
		lo.Sub(q, lo).Sub(lo, oneInt)
	}
	return roundBounds(lo, hi, scale, precision)
}

// truncDigits truncates m > 0 to at least w digits, returning the number of dropped digits
// and whether they were all zeros.
func truncDigits(m *big.Int, w int64) (int64, bool) {
	// m >= 2^(BitLen-1), so m has at least floor((BitLen-1)*log10(2))+1 digits
	drop := int64(float64(m.BitLen()-1)*math.Log10(2)-1e-9) + 1 - w
	if drop <= 0 {
		return 0, true
	}
	var r big.Int
	m.QuoRem(m, pow10(drop), &r)
	return drop, r.Sign() == 0
}

// powFrac returns d^y for d > 0 and non-integer (or negative integer) y, rounded as described in
// PowWithPrecision. It evaluates d^y = 10^q * e^s, where y*ln(d) = q*ln(10) + s and 0 <= s < ln(10),
// in binary fixed-point arithmetic, adding working digits until the error bounds decide the rounding.
func (d Decimal) powFrac(y Decimal, precision int32) (Decimal, error) {
	if d.absIsOne() {
		// 1^y = 1, also for y beyond float64 range
		if !powInRange(0, precision) {
			return Decimal{}, errPowOutOfRange
		}
		k := int64(abs(precision)) + 2
		res, _ := roundBounds(pow10(k), pow10(k), k, precision)
		return res, nil
	}

	yf := y.InexactFloat64()
	lg := yf * d.log10Abs()
	if !powInRange(lg, precision) {
		return Decimal{}, errPowOutOfRange
	}

	k := int64(d.NumDigits()) - 1 + int64(d.exp)
	// covers the working error amplified by |y| and |k|, see powFracFixed
	extra := int64(math.Log10(math.Abs(yf)*(math.Abs(float64(k))+64)+math.Abs(lg)+1)) + 4
	// 9 guard digits make a retry unlikely, each retry adds 20 more
	f := powSigDigits(lg, precision) + 9
	exactTie := false
	for i := 0; ; i, f = i+1, f+20 {
		if res, ok := powFracFixed(d, y, k, f, extra, precision, exactTie); ok {
			return res, nil
		}
		// only a terminating result can sit exactly on a rounding boundary, which no working
		// precision resolves, then the upper bound rounds it away from zero
		if i == 8 {
			exactTie = powIsTerminating(d, y)
		}
	}
}

// powIsTerminating reports whether d^y is a terminating decimal, for d > 0.
func powIsTerminating(d, y Decimal) bool {
	p, q := powRatio(y)
	num, den := powRatio(d)
	if !q.IsInt64() {
		return false
	}
	// d^(p/q) is rational only when num and den are perfect q-th powers
	rn, ok := iroot(num, q.Int64())
	if !ok {
		return false
	}
	rd, ok := iroot(den, q.Int64())
	if !ok {
		return false
	}
	// (rn/rd)^p terminates when its denominator has no prime factors other than 2 and 5
	if p.Sign() < 0 {
		rd = rn
	}
	rd = new(big.Int).Set(rd)
	for _, f := range []*big.Int{twoInt, fiveInt} {
		for r := new(big.Int); rd.Cmp(oneInt) > 0 && r.Rem(rd, f).Sign() == 0; {
			rd.Quo(rd, f)
		}
	}
	return rd.Cmp(oneInt) == 0
}

// powRatio returns d as a fraction p/q in lowest terms, q > 0.
func powRatio(d Decimal) (*big.Int, *big.Int) {
	if d.exp >= 0 {
		return new(big.Int).Mul(d.getValue(), pow10(int64(d.exp))), big.NewInt(1)
	}
	p, q := new(big.Int).Set(d.getValue()), new(big.Int).Set(pow10(-int64(d.exp)))
	g := new(big.Int).GCD(nil, nil, new(big.Int).Abs(p), q)
	return p.Quo(p, g), q.Quo(q, g)
}

// iroot returns the integer k-th root of x >= 1 and whether it is exact.
func iroot(x *big.Int, k int64) (*big.Int, bool) {
	if k == 1 || x.Cmp(oneInt) == 0 {
		return x, true
	}
	if int64(x.BitLen()) <= k {
		// 1 < x < 2^k has no integer k-th root
		return nil, false
	}
	// Newton's method from above converges to the floor of the root
	kb, km1 := big.NewInt(k), big.NewInt(k-1)
	r := new(big.Int).Lsh(oneInt, uint((int64(x.BitLen())+k-1)/k))
	t, u := new(big.Int), new(big.Int)
	for {
		t.Quo(x, t.Exp(r, km1, nil))
		t.Add(t, u.Mul(r, km1)).Quo(t, kb)
		if t.Cmp(r) >= 0 {
			break
		}
		r.Set(t)
	}
	return r, t.Exp(r, kb, nil).Cmp(x) == 0
}

// powFracFixed computes d^y (see powFrac) as an interval of e^s * 10^f with f+extra working digits
// and rounds it. With force the upper bound is used as the exact value.
func powFracFixed(d, y Decimal, k, f, extra int64, precision int32, force bool) (Decimal, bool) {
	b := uint(float64(f+extra)*math.Log2(10)) + 1

	// d = m * 10^k, 1 <= m < 10
	m := new(big.Int).Lsh(d.getValue(), b)
	m.Quo(m, pow10(int64(d.NumDigits())-1))
	t, lnErr := lnFixed(m, b)
	ln10, ln10Err := ln10Fixed(b)
	t.Add(t, new(big.Int).Mul(big.NewInt(k), ln10))
	tErr := big.NewInt(lnErr + 1 + abs64(k)*ln10Err)

	// t = y * ln(d)
	t.Mul(t, y.getValue())
	if y.exp >= 0 {
		t.Mul(t, pow10(int64(y.exp)))
	} else {
		t.Quo(t, pow10(-int64(y.exp)))
	}
	tErr.Mul(tErr, y.Abs().Ceil().getValue()).Add(tErr, oneInt)

	q, s := new(big.Int).DivMod(t, ln10, new(big.Int))
	sErr := new(big.Int).Abs(q)
	sErr.Mul(sErr, big.NewInt(ln10Err)).Add(sErr, tErr)

	// e^s < 11, so an error in s grows at most 11 times
	r, rErr := expFixed(s, b)
	sErr.Mul(sErr, big.NewInt(11)).Add(sErr, big.NewInt(rErr+1))

	// v * 10^(f-q) = e^s * 10^f
	p := pow10(f)
	hi := new(big.Int).Add(r, sErr)
	hi.Mul(hi, p).Rsh(hi, b).Add(hi, oneInt)
	lo := hi
	if !force {
		lo = new(big.Int).Sub(r, sErr)
		lo.Mul(lo, p).Rsh(lo, b)
		if lo.Sign() < 0 {
			lo.SetInt64(0)
		}
	}
	return roundBounds(lo, hi, f-q.Int64(), precision)
}

// lnFixed returns ln(m / 2^b) * 2^b for m / 2^b in [1, 10], with its error bound in ulps.
func lnFixed(m *big.Int, b uint) (*big.Int, int64) {
	// a = math.Log(m) is accurate to ~1e-15, so ln(m) = a + ln(1+z), where z = m*e^-a - 1 is tiny
	// and the series ln(1+z) = z - z^2/2 + z^3/3 - ... gains ~50 bits per term.
	top, sh := new(big.Int), m.BitLen()-60
	if sh > 0 {
		top.Rsh(m, uint(sh))
	} else {
		top.Lsh(m, uint(-sh))
	}
	a := big.NewInt(int64(math.Ldexp(math.Log(math.Ldexp(float64(top.Uint64()), sh-int(b))), 52)))
	if b >= 52 {
		a.Lsh(a, b-52)
	} else {
		a.Rsh(a, 52-b)
	}

	e, eErr := expFixed(new(big.Int).Neg(a), b)
	z := e.Mul(e, m).Rsh(e, b)
	z.Sub(z, new(big.Int).Lsh(oneInt, b))

	sum, p, term, iv := new(big.Int).Set(z), new(big.Int).Set(z), new(big.Int), new(big.Int)
	terms := int64(1)
	for i := int64(2); ; i++ {
		p.Mul(p, z).Rsh(p, b)
		if p.Sign() == 0 {
			break
		}
		term.Quo(p, iv.SetInt64(i))
		if i%2 == 0 {
			sum.Sub(sum, term)
		} else {
			sum.Add(sum, term)
		}
		terms++
	}
	// m <= 10 amplifies the error of e^-a at most 10 times
	return sum.Add(sum, a), 10*eErr + 2*terms + 3
}

// expFixed returns e^(x / 2^b) * 2^b for |x / 2^b| <= 2.4, with its error bound in ulps.
func expFixed(x *big.Int, b uint) (*big.Int, int64) {
	// e^x = (e^(x/2^j))^(2^j): Taylor series of the reduced argument, then j squarings.
	// Guard bits cover the error of up to ~g Taylor terms, amplified up to 11*2^j times by squaring.
	j := uint(math.Sqrt(float64(b))) + 1
	g := b + j + uint(bits.Len(33*(b+j+64)+363)) + 2

	xr := new(big.Int).Lsh(x, g-b)
	xr.Rsh(xr, j)
	sum, term, iv := new(big.Int).Lsh(oneInt, g), new(big.Int).Lsh(oneInt, g), new(big.Int)
	for i := int64(1); ; i++ {
		term.Mul(term, xr).Rsh(term, g)
		term.Quo(term, iv.SetInt64(i))
		if term.Sign() == 0 {
			break
		}
		sum.Add(sum, term)
	}
	for ; j > 0; j-- {
		sum.Mul(sum, sum).Rsh(sum, g)
	}
	return sum.Rsh(sum, g-b), 2
}

var ln10Cache struct {
	once sync.Once
	v    *big.Int
}

const ln10CacheBits = 2048

// ln10Fixed returns ln(10) * 2^b, with its error bound in ulps.
func ln10Fixed(b uint) (*big.Int, int64) {
	if b <= ln10CacheBits {
		ln10Cache.once.Do(func() { ln10Cache.v, _ = ln10FromDigits(ln10CacheBits) })
		return new(big.Int).Rsh(ln10Cache.v, ln10CacheBits-b), 3
	}
	if v, ok := ln10FromDigits(b); ok {
		return v, 2
	}
	return lnFixed(new(big.Int).Lsh(big.NewInt(10), b), b)
}

// ln10FromDigits returns ln(10) * 2^b computed from strLn10, if it has enough digits.
func ln10FromDigits(b uint) (*big.Int, bool) {
	digits := int(float64(b)*math.Log10(2)) + 3
	if digits+2 > len(strLn10) {
		return nil, false
	}
	v, _ := new(big.Int).SetString(strLn10[:1]+strLn10[2:digits+2], 10)
	v.Lsh(v, b)
	return v.Quo(v, pow10(int64(digits))), true
}

// roundBounds rounds v, known only by lo <= v*10^scale <= hi (lo >= 0), half away from zero to
// precision places after the decimal point, or to max(precision, 1) significant digits when that
// would be 0. It reports false when lo and hi do not round to the same value.
func roundBounds(lo, hi *big.Int, scale int64, precision int32) (Decimal, bool) {
	g := scale - int64(precision)
	if g < 1 {
		return Decimal{}, false
	}
	dhi := int64(Decimal{value: hi}.NumDigits())
	// with fewer than g digits hi rounds to 0, skip building 10^g for tiny results
	if g <= dhi {
		if q := roundHalfUp(hi, g); q.Sign() != 0 {
			return Decimal{q, -precision}, q.Cmp(roundHalfUp(lo, g)) == 0
		}
	}

	sig := powMinSig(precision)
	g = dhi - sig
	if g < 2 {
		return Decimal{}, false
	}
	q := roundHalfUp(hi, g)
	res := Decimal{q, int32(g - scale)}
	if int64(Decimal{value: lo}.NumDigits()) == dhi {
		return res, q.Cmp(roundHalfUp(lo, g)) == 0
	}
	// lo < 10^(dhi-1) <= hi, decided only when both round to that power of ten
	return res, q.Cmp(pow10(sig-1)) == 0 && roundHalfUp(lo, g-1).Cmp(pow10(sig)) == 0
}

// roundHalfUp returns x / 10^g rounded half away from zero, for x >= 0 and g >= 1.
func roundHalfUp(x *big.Int, g int64) *big.Int {
	p := pow10(g)
	q, r := new(big.Int).QuoRem(x, p, new(big.Int))
	if r.Lsh(r, 1).Cmp(p) >= 0 {
		q.Add(q, oneInt)
	}
	return q
}

// powSigDigits returns the number of significant digits of a result of magnitude 10^lg rounded to
// precision places, at least powMinSig(precision).
func powSigDigits(lg float64, precision int32) int64 {
	if n := int64(math.Floor(lg)) + 1 + int64(precision); n > powMinSig(precision) {
		return n
	}
	return powMinSig(precision)
}

// powMinSig returns the significant digits kept by results that would round to zero.
func powMinSig(precision int32) int64 {
	if precision < 1 {
		return 1
	}
	return int64(precision)
}

// powInRange reports whether a result of magnitude 10^lg rounded to precision places has an exponent
// that fits into int32.
func powInRange(lg float64, precision int32) bool {
	return math.Abs(lg)+math.Abs(float64(precision)) < math.MaxInt32/2
}

// absIsOne reports whether |d| = 1.
func (d Decimal) absIsOne() bool {
	nd := int64(d.NumDigits())
	return nd-1+int64(d.exp) == 0 && d.getValue().CmpAbs(pow10(nd-1)) == 0
}

// log10Abs approximates log10(|d|) for d != 0.
func (d Decimal) log10Abs() float64 {
	var m big.Float
	e2 := new(big.Float).SetInt(d.getValue()).MantExp(&m)
	mf, _ := m.Float64()
	return math.Log10(math.Abs(mf)) + float64(e2)*math.Log10(2) + float64(d.exp)
}

// pow10Table holds 10^0 ... 10^127, its values must not be modified.
var pow10Table = func() []*big.Int {
	t := make([]*big.Int, 128)
	t[0] = big.NewInt(1)
	for i := 1; i < len(t); i++ {
		t[i] = new(big.Int).Mul(t[i-1], tenInt)
	}
	return t
}()

// pow10Uint64 holds 10^0 ... 10^19.
var pow10Uint64 = func() (t [20]uint64) {
	t[0] = 1
	for i := 1; i < len(t); i++ {
		t[i] = t[i-1] * 10
	}
	return t
}()

// float64Pow10 holds 10^0 ... 10^22, the powers of ten that float64 represents exactly.
var float64Pow10 = [...]float64{1e0, 1e1, 1e2, 1e3, 1e4, 1e5, 1e6, 1e7, 1e8, 1e9, 1e10,
	1e11, 1e12, 1e13, 1e14, 1e15, 1e16, 1e17, 1e18, 1e19, 1e20, 1e21, 1e22}

// pow5Int64 holds 5^0 ... 5^22.
var pow5Int64 = func() (t [23]int64) {
	t[0] = 1
	for i := 1; i < len(t); i++ {
		t[i] = t[i-1] * 5
	}
	return t
}()

// pow10 returns 10^n for n >= 0, the result must not be modified.
func pow10(n int64) *big.Int {
	if n < int64(len(pow10Table)) {
		return pow10Table[n]
	}
	return new(big.Int).Exp(tenInt, big.NewInt(n), nil)
}

func abs64(n int64) int64 {
	if n < 0 {
		return -n
	}
	return n
}

// ExpHullAbrham calculates the natural exponent of decimal (e to the power of d) using Hull-Abraham algorithm.
// OverallPrecision argument specifies the overall precision of the result (integer part + decimal part).
//
// ExpHullAbrham is faster than ExpTaylor for small precision values, but it is much slower for large precision values.
//
// Example:
//
//	NewFromFloat(26.1).ExpHullAbrham(2).String()    // output: "220000000000"
//	NewFromFloat(26.1).ExpHullAbrham(20).String()   // output: "216314672147.05767284"
func (d Decimal) ExpHullAbrham(overallPrecision uint32) (Decimal, error) {
	// Algorithm based on Variable precision exponential function.
	// ACM Transactions on Mathematical Software by T. E. Hull & A. Abrham.
	if d.IsZero() {
		return Decimal{oneInt, 0}, nil
	}

	currentPrecision := overallPrecision

	// Algorithm does not work if currentPrecision * 23 < |x|.
	// Precision is automatically increased in such cases, so the value can be calculated precisely.
	// If newly calculated precision is higher than ExpMaxIterations the currentPrecision will not be changed.
	f := d.Abs().InexactFloat64()
	if ncp := f / 23; ncp > float64(currentPrecision) && ncp < float64(ExpMaxIterations) {
		currentPrecision = uint32(math.Ceil(ncp))
	}

	// fail if abs(d) beyond an over/underflow threshold
	overflowThreshold := New(23*int64(currentPrecision), 0)
	if d.Abs().Cmp(overflowThreshold) > 0 {
		return Decimal{}, fmt.Errorf("over/underflow threshold, exp(x) cannot be calculated precisely")
	}

	// Return 1 if abs(d) small enough; this also avoids later over/underflow
	overflowThreshold2 := New(9, -int32(currentPrecision)-1)
	if d.Abs().Cmp(overflowThreshold2) <= 0 {
		return Decimal{oneInt, 0}, nil
	}

	// t is the smallest integer >= 0 such that the corresponding abs(d/k) < 1
	t := d.exp + int32(d.NumDigits()) // Add d.NumDigits because the paper assumes that d.value [0.1, 1)

	if t < 0 {
		t = 0
	}

	k := New(1, t)                                          // reduction factor
	r := Decimal{new(big.Int).Set(d.getValue()), d.exp - t} // reduced argument
	p := int32(currentPrecision) + t + 2                    // precision for calculating the sum

	// Determine n, the number of therms for calculating sum
	// use first Newton step (1.435p - 1.182) / log10(p/abs(r))
	// for solving appropriate equation, along with directed
	// roundings and simple rational bound for log10(p/abs(r))
	rf := r.Abs().InexactFloat64()
	pf := float64(p)
	nf := math.Ceil((1.453*pf - 1.182) / math.Log10(pf/rf))
	if nf > float64(ExpMaxIterations) || math.IsNaN(nf) {
		return Decimal{}, fmt.Errorf("exact value cannot be calculated in <=ExpMaxIterations iterations")
	}
	n := int64(nf)

	tmp := New(0, 0)
	sum := New(1, 0)
	one := New(1, 0)
	for i := n - 1; i > 0; i-- {
		tmp.value.SetInt64(i)
		sum = sum.Mul(r.DivRound(tmp, p))
		sum = sum.Add(one)
	}

	// res = sum^ki, the same value ki repeated multiplications would give
	ki := k.IntPart()
	expInt64 := int64(sum.exp) * ki
	if expInt64 > math.MaxInt32 || expInt64 < math.MinInt32 {
		panic(fmt.Sprintf("exponent %v overflows an int32!", expInt64))
	}
	res := Decimal{new(big.Int).Exp(sum.getValue(), big.NewInt(ki), nil), int32(expInt64)}

	resNumDigits := int32(res.NumDigits())

	var roundDigits int32
	if resNumDigits > abs(res.exp) {
		roundDigits = int32(currentPrecision) - resNumDigits - res.exp
	} else {
		roundDigits = int32(currentPrecision)
	}

	res = res.Round(roundDigits)

	return res, nil
}

// ExpTaylor calculates the natural exponent of decimal (e to the power of d) using Taylor series expansion.
// Precision argument specifies how precise the result must be (number of digits after decimal point).
// Negative precision is allowed.
//
// ExpTaylor is much faster for large precision values than ExpHullAbrham.
//
// Example:
//
//	d, err := NewFromFloat(26.1).ExpTaylor(2).String()
//	d.String()  // output: "216314672147.06"
//
//	NewFromFloat(26.1).ExpTaylor(20).String()
//	d.String()  // output: "216314672147.05767284062928674083"
//
//	NewFromFloat(26.1).ExpTaylor(-10).String()
//	d.String()  // output: "220000000000"
func (d Decimal) ExpTaylor(precision int32) (Decimal, error) {
	// Note(mwoss): Implementation can be optimized by exclusively using big.Int API only
	if d.IsZero() {
		return Decimal{oneInt, 0}.Round(precision), nil
	}

	var epsilon Decimal
	var divPrecision int32
	if precision < 0 {
		epsilon = New(1, -1)
		divPrecision = 8
	} else {
		epsilon = New(1, -precision-1)
		divPrecision = precision + 1
	}

	decAbs := d.Abs()
	pow := d.Abs()
	factorial := New(1, 0)

	result := New(1, 0)

	for i := int64(1); ; {
		step := pow.DivRound(factorial, divPrecision)
		result = result.Add(step)

		// Stop Taylor series when current step is smaller than epsilon
		if step.Cmp(epsilon) < 0 {
			break
		}

		pow = pow.Mul(decAbs)

		i++

		// Calculate next factorial number or retrieve cached value
		factorialsMutex.RLock()
		if len(factorials) >= int(i) && !factorials[i-1].IsZero() {
			factorial = factorials[i-1]
			factorialsMutex.RUnlock()
		} else {
			prevFactorial := factorials[i-2]
			factorialsMutex.RUnlock()
			factorial = prevFactorial.Mul(New(i, 0))
			factorialsMutex.Lock()
			// Check again in case another goroutine already added it.
			if len(factorials) < int(i) || factorials[i-1].IsZero() {
				factorials = append(factorials, Zero)
				factorials[i-1] = factorial
			}
			factorialsMutex.Unlock()
		}
	}

	if d.Sign() < 0 {
		result = New(1, 0).DivRound(result, precision+1)
	}

	result = result.Round(precision)
	return result, nil
}

// Ln calculates natural logarithm of d.
// Precision argument specifies how precise the result must be (number of digits after decimal point).
// Negative precision is allowed.
//
// Example:
//
//	d1, err := NewFromFloat(13.3).Ln(2)
//	d1.String()  // output: "2.59"
//
//	d2, err := NewFromFloat(579.161).Ln(10)
//	d2.String()  // output: "6.3615805046"
func (d Decimal) Ln(precision int32) (Decimal, error) {
	// Algorithm based on The Use of Iteration Methods for Approximating the Natural Logarithm,
	// James F. Epperson, The American Mathematical Monthly, Vol. 96, No. 9, November 1989, pp. 831-835.
	if d.IsNegative() {
		return Decimal{}, fmt.Errorf("cannot calculate natural logarithm for negative decimals")
	}

	if d.IsZero() {
		return Decimal{}, fmt.Errorf("cannot represent natural logarithm of 0, result: -infinity")
	}

	calcPrecision := precision + 2
	z := d.Copy()

	var comp1, comp3, comp2, comp4, reduceAdjust Decimal
	comp1 = z.Sub(Decimal{oneInt, 0})
	comp3 = Decimal{oneInt, -1}

	// for decimal in range [0.9, 1.1] where ln(d) is close to 0
	usePowerSeries := false

	if comp1.Abs().Cmp(comp3) <= 0 {
		usePowerSeries = true
	} else {
		// reduce input decimal to range [0.1, 1)
		expDelta := int32(z.NumDigits()) + z.exp
		z.exp -= expDelta

		// Input decimal was reduced by factor of 10^expDelta, thus we will need to add
		// ln(10^expDelta) = expDelta * ln(10)
		// to the result to compensate that
		ln10 := ln10.withPrecision(calcPrecision)
		reduceAdjust = NewFromInt32(expDelta)
		reduceAdjust = reduceAdjust.Mul(ln10)

		comp1 = z.Sub(Decimal{oneInt, 0})

		if comp1.Abs().Cmp(comp3) <= 0 {
			usePowerSeries = true
		} else {
			// initial estimate using floats
			zFloat := z.InexactFloat64()
			comp1 = NewFromFloat(math.Log(zFloat))
		}
	}

	epsilon := Decimal{oneInt, -calcPrecision}

	if usePowerSeries {
		// Power Series - https://en.wikipedia.org/wiki/Logarithm#Power_series
		// Calculating n-th term of formula: ln(z+1) = 2 sum [ 1 / (2n+1) * (z / (z+2))^(2n+1) ]
		// until the difference between current and next term is smaller than epsilon.
		// Coverage quite fast for decimals close to 1.0

		// z + 2
		comp2 = comp1.Add(Decimal{twoInt, 0})
		// z / (z + 2)
		comp3 = comp1.DivRound(comp2, calcPrecision)
		// 2 * (z / (z + 2))
		comp1 = comp3.Add(comp3)
		comp2 = comp1.Copy()

		for n := 1; ; n++ {
			// 2 * (z / (z+2))^(2n+1)
			comp2 = comp2.Mul(comp3).Mul(comp3)

			// 1 / (2n+1) * 2 * (z / (z+2))^(2n+1)
			comp4 = NewFromInt(int64(2*n + 1))
			comp4 = comp2.DivRound(comp4, calcPrecision)

			// comp1 = 2 sum [ 1 / (2n+1) * (z / (z+2))^(2n+1) ]
			comp1 = comp1.Add(comp4)

			if comp4.Abs().Cmp(epsilon) <= 0 {
				break
			}
		}
	} else {
		// Halley's Iteration.
		// Calculating n-th term of formula: a_(n+1) = a_n - 2 * (exp(a_n) - z) / (exp(a_n) + z),
		// until the difference between current and next term is smaller than epsilon
		var prevStep Decimal
		maxIters := calcPrecision*2 + 10

		for i := int32(0); i < maxIters; i++ {
			// exp(a_n)
			comp3, _ = comp1.ExpTaylor(calcPrecision)
			// exp(a_n) - z
			comp2 = comp3.Sub(z)
			// 2 * (exp(a_n) - z)
			comp2 = comp2.Add(comp2)
			// exp(a_n) + z
			comp4 = comp3.Add(z)
			// 2 * (exp(a_n) - z) / (exp(a_n) + z)
			comp3 = comp2.DivRound(comp4, calcPrecision)
			// comp1 = a_(n+1) = a_n - 2 * (exp(a_n) - z) / (exp(a_n) + z)
			comp1 = comp1.Sub(comp3)

			if prevStep.Add(comp3).IsZero() {
				// If iteration steps oscillate we should return early and prevent an infinity loop
				// NOTE(mwoss): This should be quite a rare case, returning error is not necessary
				break
			}

			if comp3.Abs().Cmp(epsilon) <= 0 {
				break
			}

			prevStep = comp3
		}
	}

	comp1 = comp1.Add(reduceAdjust)

	return comp1.Round(precision), nil
}

// NumDigits returns the number of digits of the decimal coefficient (d.Value)
func (d Decimal) NumDigits() int {
	v := d.getValue()
	if v.IsInt64() {
		u := uint64(v.Int64())
		if u == 0 {
			return 1
		}
		if v.Sign() < 0 {
			u = -u
		}
		// Not math.Log10, which rounds an exact 1eN like 1e15 down (see #420).
		// bits.Len64(u)*1233>>12 is floor(log10(2^bits)), the digit count or one less.
		n := bits.Len64(u) * 1233 >> 12
		if u >= pow10Uint64[n] {
			n++
		}
		return n
	}

	estimatedNumDigits := int(float64(v.BitLen()) / math.Log2(10))

	// estimatedNumDigits (lg10) may be off by 1, need to verify
	if v.CmpAbs(pow10(int64(estimatedNumDigits))) >= 0 {
		return estimatedNumDigits + 1
	}

	return estimatedNumDigits
}

// IsInteger returns true when decimal can be represented as an integer value, otherwise, it returns false.
func (d Decimal) IsInteger() bool {
	// The most typical case, all decimal with exponent higher or equal 0 can be represented as integer
	if d.exp >= 0 {
		return true
	}
	v := d.getValue()
	if v.Sign() == 0 {
		return true
	}
	if d.exp >= -18 && v.IsInt64() {
		return v.Int64()%int64(pow10Uint64[-d.exp]) == 0
	}
	// When the exponent is negative we have to check every number after the decimal place
	// If all of them are zeroes, we are sure that given decimal can be represented as an integer
	var r big.Int
	q := new(big.Int).Set(v)
	for z := abs(d.exp); z > 0; z-- {
		q.QuoRem(q, tenInt, &r)
		if r.Cmp(zeroInt) != 0 {
			return false
		}
	}
	return true
}

// Abs calculates absolute value of any int32. Used for calculating absolute value of decimal's exponent.
func abs(n int32) int32 {
	if n < 0 {
		return -n
	}
	return n
}

// Cmp compares the numbers represented by d and d2 and returns:
//
//	-1 if d <  d2
//	 0 if d == d2
//	+1 if d >  d2
func (d Decimal) Cmp(d2 Decimal) int {
	if d.exp == d2.exp {
		return d.getValue().Cmp(d2.getValue())
	}
	if s, s2 := d.Sign(), d2.Sign(); s != s2 {
		if s > s2 {
			return 1
		}
		return -1
	} else if s == 0 {
		return 0
	}

	var scaled big.Int
	if d.exp < d2.exp {
		return d.getValue().Cmp(scaled.Mul(d2.getValue(), pow10(int64(d2.exp)-int64(d.exp))))
	}
	return scaled.Mul(d.getValue(), pow10(int64(d.exp)-int64(d2.exp))).Cmp(d2.getValue())
}

// Compare compares the numbers represented by d and d2 and returns:
//
//	-1 if d <  d2
//	 0 if d == d2
//	+1 if d >  d2
func (d Decimal) Compare(d2 Decimal) int {
	return d.Cmp(d2)
}

// Equal returns whether the numbers represented by d and d2 are equal.
func (d Decimal) Equal(d2 Decimal) bool {
	return d.Cmp(d2) == 0
}

// Deprecated: Equals is deprecated, please use Equal method instead.
func (d Decimal) Equals(d2 Decimal) bool {
	return d.Equal(d2)
}

// GreaterThan (GT) returns true when d is greater than d2.
func (d Decimal) GreaterThan(d2 Decimal) bool {
	return d.Cmp(d2) == 1
}

// GreaterThanOrEqual (GTE) returns true when d is greater than or equal to d2.
func (d Decimal) GreaterThanOrEqual(d2 Decimal) bool {
	cmp := d.Cmp(d2)
	return cmp == 1 || cmp == 0
}

// LessThan (LT) returns true when d is less than d2.
func (d Decimal) LessThan(d2 Decimal) bool {
	return d.Cmp(d2) == -1
}

// LessThanOrEqual (LTE) returns true when d is less than or equal to d2.
func (d Decimal) LessThanOrEqual(d2 Decimal) bool {
	cmp := d.Cmp(d2)
	return cmp == -1 || cmp == 0
}

// Sign returns:
//
//	-1 if d <  0
//	 0 if d == 0
//	+1 if d >  0
func (d Decimal) Sign() int {
	return d.getValue().Sign()
}

// IsPositive return
//
//	true if d > 0
//	false if d == 0
//	false if d < 0
func (d Decimal) IsPositive() bool {
	return d.Sign() == 1
}

// IsNegative return
//
//	true if d < 0
//	false if d == 0
//	false if d > 0
func (d Decimal) IsNegative() bool {
	return d.Sign() == -1
}

// IsZero return
//
//	true if d == 0
//	false if d > 0
//	false if d < 0
func (d Decimal) IsZero() bool {
	return d.Sign() == 0
}

// Exponent returns the exponent, or scale component of the decimal.
func (d Decimal) Exponent() int32 {
	return d.exp
}

// Coefficient returns the coefficient of the decimal. It is scaled by 10^Exponent()
func (d Decimal) Coefficient() *big.Int {
	// we copy the coefficient so that mutating the result does not mutate the Decimal.
	return new(big.Int).Set(d.getValue())
}

// CoefficientInt64 returns the coefficient of the decimal as int64. It is scaled by 10^Exponent()
// If coefficient cannot be represented in an int64, the result will be undefined.
func (d Decimal) CoefficientInt64() int64 {
	return d.getValue().Int64()
}

// IntPart returns the integer component of the decimal.
func (d Decimal) IntPart() int64 {
	if v := d.getValue(); d.exp <= 0 && d.exp >= -18 && v.IsInt64() {
		return v.Int64() / int64(pow10Uint64[-d.exp])
	}
	scaledD := d.rescale(0)
	return scaledD.getValue().Int64()
}

// BigInt returns integer component of the decimal as a BigInt.
func (d Decimal) BigInt() *big.Int {
	scaledD := d.rescale(0)
	return scaledD.getValue()
}

// BigFloat returns decimal as BigFloat.
// Be aware that casting decimal to BigFloat might cause a loss of precision.
func (d Decimal) BigFloat() *big.Float {
	f := &big.Float{}
	f.SetString(d.String())
	return f
}

// Rat returns a rational number representation of the decimal.
func (d Decimal) Rat() *big.Rat {
	if d.exp <= 0 {
		// NOTE(vadim): must negate after casting to prevent int32 overflow
		return new(big.Rat).SetFrac(d.getValue(), pow10(-int64(d.exp)))
	}

	num := new(big.Int).Mul(d.getValue(), pow10(int64(d.exp)))
	return new(big.Rat).SetFrac(num, oneInt)
}

// Float64 returns the nearest float64 value for d and a bool indicating
// whether f represents d exactly.
// For more details, see the documentation for big.Rat.Float64
func (d Decimal) Float64() (f float64, exact bool) {
	// Rat() materializes 10^|exp|, which takes forever for huge exponents.
	// Values that far outside the float64 range don't need it: they always
	// round to zero or overflow to infinity.
	sign := d.Sign()
	if sign == 0 {
		return 0, true
	}
	// Both operands are exact here, so the division is correctly rounded like Rat().Float64(),
	// and the result is exact when the 5^k part of 10^k divides the coefficient.
	if v := d.getValue(); d.exp <= 0 && d.exp >= -22 && v.IsInt64() {
		if i := v.Int64(); i >= -1<<53 && i <= 1<<53 {
			return float64(i) / float64Pow10[-d.exp], i%pow5Int64[-d.exp] == 0
		}
	}
	// |d| lies in [10^(magnitude-1), 10^magnitude). float64 spans roughly
	// 1e-324..1e308, so ±400 is safely outside it with margin to spare.
	magnitude := int64(d.NumDigits()) + int64(d.exp)
	if magnitude < -400 {
		return math.Copysign(0, float64(sign)), false
	}
	if magnitude > 400 {
		return math.Inf(sign), false
	}

	return d.Rat().Float64()
}

// InexactFloat64 returns the nearest float64 value for d.
// It doesn't indicate if the returned value represents d exactly.
func (d Decimal) InexactFloat64() float64 {
	f, _ := d.Float64()
	return f
}

// String returns the string representation of the decimal
// with the fixed point.
//
// Example:
//
//	d := New(-12345, -3)
//	println(d.String())
//
// Output:
//
//	-12.345
func (d Decimal) String() string {
	return d.string(TrimTrailingZeros, UseScientificNotation)
}

// StringFixed returns a rounded fixed-point string with places digits after
// the decimal point.
//
// Example:
//
//	NewFromFloat(0).StringFixed(2) // output: "0.00"
//	NewFromFloat(0).StringFixed(0) // output: "0"
//	NewFromFloat(5.45).StringFixed(0) // output: "5"
//	NewFromFloat(5.45).StringFixed(1) // output: "5.5"
//	NewFromFloat(5.45).StringFixed(2) // output: "5.45"
//	NewFromFloat(5.45).StringFixed(3) // output: "5.450"
//	NewFromFloat(545).StringFixed(-1) // output: "540"
//
// Regardless of the UseScientificNotation option, the returned string will never be in scientific notation.
func (d Decimal) StringFixed(places int32) string {
	rounded := d.Round(places)
	return rounded.string(false, false)
}

// StringFixedBank returns a banker rounded fixed-point string with places digits
// after the decimal point.
//
// Example:
//
//	NewFromFloat(0).StringFixedBank(2) // output: "0.00"
//	NewFromFloat(0).StringFixedBank(0) // output: "0"
//	NewFromFloat(5.45).StringFixedBank(0) // output: "5"
//	NewFromFloat(5.45).StringFixedBank(1) // output: "5.4"
//	NewFromFloat(5.45).StringFixedBank(2) // output: "5.45"
//	NewFromFloat(5.45).StringFixedBank(3) // output: "5.450"
//	NewFromFloat(545).StringFixedBank(-1) // output: "540"
//
// Regardless of the UseScientificNotation option, the returned string will never be in scientific notation.
func (d Decimal) StringFixedBank(places int32) string {
	rounded := d.RoundBank(places)
	return rounded.string(false, false)
}

// StringFixedCash returns a Swedish/Cash rounded fixed-point string. For
// more details see the documentation at function RoundCash.
//
// Regardless of the UseScientificNotation option, the returned string will never be in scientific notation.
func (d Decimal) StringFixedCash(interval uint8) string {
	rounded := d.RoundCash(interval)
	return rounded.string(false, false)
}

// Round rounds the decimal to places decimal places.
// If places < 0, it will round the integer part to the nearest 10^(-places).
//
// Example:
//
//	NewFromFloat(5.45).Round(1).String() // output: "5.5"
//	NewFromFloat(545).Round(-1).String() // output: "550" (with UseScientificNotation false, "5.5E2" if true)
func (d Decimal) Round(places int32) Decimal {
	if d.exp == -places {
		return d
	}
	// truncate to places + 1
	ret := d.rescale(-places - 1)

	// add sign(d) * 0.5
	if ret.value.Sign() < 0 {
		ret.value.Sub(ret.value, fiveInt)
	} else {
		ret.value.Add(ret.value, fiveInt)
	}

	// floor for positive numbers, ceil for negative numbers
	_, m := ret.value.DivMod(ret.value, tenInt, new(big.Int))
	ret.exp++
	if ret.value.Sign() < 0 && m.Cmp(zeroInt) != 0 {
		ret.value.Add(ret.value, oneInt)
	}

	return ret
}

// RoundCeil rounds the decimal towards +infinity.
//
// Example:
//
//	NewFromFloat(545).RoundCeil(-2).String()   // output: "600"
//	NewFromFloat(500).RoundCeil(-2).String()   // output: "500"
//	NewFromFloat(1.1001).RoundCeil(2).String() // output: "1.11"
//	NewFromFloat(-1.454).RoundCeil(1).String() // output: "-1.4"
func (d Decimal) RoundCeil(places int32) Decimal {
	if d.exp >= -places {
		return d.rescale(-places)
	}

	rescaled := d.rescale(-places)
	if d.Equal(rescaled) {
		return rescaled
	}

	if d.getValue().Sign() > 0 {
		rescaled.value = new(big.Int).Add(rescaled.getValue(), oneInt)
	}

	return rescaled
}

// RoundFloor rounds the decimal towards -infinity.
//
// Example:
//
//	NewFromFloat(545).RoundFloor(-2).String()   // output: "500"
//	NewFromFloat(-500).RoundFloor(-2).String()   // output: "-500"
//	NewFromFloat(1.1001).RoundFloor(2).String() // output: "1.1"
//	NewFromFloat(-1.454).RoundFloor(1).String() // output: "-1.5"
func (d Decimal) RoundFloor(places int32) Decimal {
	if d.exp >= -places {
		return d.rescale(-places)
	}

	rescaled := d.rescale(-places)
	if d.Equal(rescaled) {
		return rescaled
	}

	if d.getValue().Sign() < 0 {
		rescaled.value = new(big.Int).Sub(rescaled.getValue(), oneInt)
	}

	return rescaled
}

// RoundUp rounds the decimal away from zero.
//
// Example:
//
//	NewFromFloat(545).RoundUp(-2).String()   // output: "600"
//	NewFromFloat(500).RoundUp(-2).String()   // output: "500"
//	NewFromFloat(1.1001).RoundUp(2).String() // output: "1.11"
//	NewFromFloat(-1.454).RoundUp(1).String() // output: "-1.5"
func (d Decimal) RoundUp(places int32) Decimal {
	if d.exp >= -places {
		return d.rescale(-places)
	}

	rescaled := d.rescale(-places)
	if d.Equal(rescaled) {
		return rescaled
	}

	if d.getValue().Sign() > 0 {
		rescaled.value = new(big.Int).Add(rescaled.getValue(), oneInt)
	} else if d.getValue().Sign() < 0 {
		rescaled.value = new(big.Int).Sub(rescaled.getValue(), oneInt)
	}

	return rescaled
}

// RoundDown rounds the decimal towards zero.
//
// Example:
//
//	NewFromFloat(545).RoundDown(-2).String()   // output: "500"
//	NewFromFloat(-500).RoundDown(-2).String()   // output: "-500"
//	NewFromFloat(1.1001).RoundDown(2).String() // output: "1.1"
//	NewFromFloat(-1.454).RoundDown(1).String() // output: "-1.4"
func (d Decimal) RoundDown(places int32) Decimal {
	return d.rescale(-places)
}

// RoundBank rounds the decimal to places decimal places.
// If the final digit to round is equidistant from the nearest two integers the
// rounded value is taken as the even number
//
// If places < 0, it will round the integer part to the nearest 10^(-places).
//
// Examples:
//
//	NewFromFloat(5.45).RoundBank(1).String() // output: "5.4"
//	NewFromFloat(545).RoundBank(-1).String() // output: "540"
//	NewFromFloat(5.46).RoundBank(1).String() // output: "5.5"
//	NewFromFloat(546).RoundBank(-1).String() // output: "550"
//	NewFromFloat(5.55).RoundBank(1).String() // output: "5.6"
//	NewFromFloat(555).RoundBank(-1).String() // output: "560"
func (d Decimal) RoundBank(places int32) Decimal {

	round := d.Round(places)
	remainder := d.Sub(round).Abs()

	half := New(5, -places-1)
	if remainder.Cmp(half) == 0 && round.getValue().Bit(0) != 0 {
		if round.getValue().Sign() < 0 {
			round.value = new(big.Int).Add(round.getValue(), oneInt)
		} else {
			round.value = new(big.Int).Sub(round.getValue(), oneInt)
		}
	}

	return round
}

// RoundCash aka Cash/Penny/öre rounding rounds decimal to a specific
// interval. The amount payable for a cash transaction is rounded to the nearest
// multiple of the minimum currency unit available. The following intervals are
// available: 5, 10, 25, 50 and 100; any other number throws a panic.
//
//	  5:   5 cent rounding 3.43 => 3.45
//	 10:  10 cent rounding 3.45 => 3.50 (5 gets rounded up)
//	 25:  25 cent rounding 3.41 => 3.50
//	 50:  50 cent rounding 3.75 => 4.00
//	100: 100 cent rounding 3.50 => 4.00
//
// For more details: https://en.wikipedia.org/wiki/Cash_rounding
func (d Decimal) RoundCash(interval uint8) Decimal {
	var iVal *big.Int
	switch interval {
	case 5:
		iVal = twentyInt
	case 10:
		iVal = tenInt
	case 25:
		iVal = fourInt
	case 50:
		iVal = twoInt
	case 100:
		iVal = oneInt
	default:
		panic(fmt.Sprintf("Decimal does not support this Cash rounding interval `%d`. Supported: 5, 10, 25, 50, 100", interval))
	}
	dVal := Decimal{
		value: iVal,
	}

	// TODO: optimize those calculations to reduce the high allocations (~29 allocs).
	return d.Mul(dVal).Round(0).Div(dVal).Truncate(2)
}

// Floor returns the nearest integer value less than or equal to d.
func (d Decimal) Floor() Decimal {
	if d.exp >= 0 {
		return d
	}

	// NOTE(vadim): must negate after casting to prevent int32 overflow
	z := new(big.Int).Div(d.getValue(), pow10(-int64(d.exp)))
	return Decimal{value: z, exp: 0}
}

// Ceil returns the nearest integer value greater than or equal to d.
func (d Decimal) Ceil() Decimal {
	if d.exp >= 0 {
		return d
	}

	// NOTE(vadim): must negate after casting to prevent int32 overflow
	z, m := new(big.Int).DivMod(d.getValue(), pow10(-int64(d.exp)), new(big.Int))
	if m.Cmp(zeroInt) != 0 {
		z.Add(z, oneInt)
	}
	return Decimal{value: z, exp: 0}
}

// Truncate truncates off digits from the number, without rounding.
//
// If precision >= 0, it specifies the number of decimal places to keep.
// If precision < 0, it truncates the integer part to the nearest 10^(-precision)
// towards zero.
//
// Example:
//
//	decimal.NewFromString("123.456").Truncate(2).String()  // "123.45"
//	decimal.NewFromString("5432").Truncate(-2).String()    // "5400"
//	decimal.NewFromString("-5432").Truncate(-2).String()   // "-5400"
func (d Decimal) Truncate(precision int32) Decimal {
	if -precision > d.exp {
		return d.rescale(-precision)
	}
	return d
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (d *Decimal) UnmarshalJSON(decimalBytes []byte) error {
	if string(decimalBytes) == "null" {
		return nil
	}

	decimal, err := NewFromString(unquoteIfQuoted(string(decimalBytes)))
	*d = decimal
	if err != nil {
		return fmt.Errorf("error decoding string '%s': %s", string(decimalBytes), err)
	}
	return nil
}

// MarshalJSON implements the json.Marshaler interface.
func (d Decimal) MarshalJSON() ([]byte, error) {
	str := d.String()
	if MarshalJSONWithoutQuotes {
		return []byte(str), nil
	}
	b := make([]byte, 0, len(str)+2)
	b = append(b, '"')
	b = append(b, str...)
	return append(b, '"'), nil
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface. As a string representation
// is already used when encoding to text, this method stores that string as []byte
func (d *Decimal) UnmarshalBinary(data []byte) error {
	// Verify we have at least 4 bytes for the exponent. The GOB encoded value
	// may be empty.
	if len(data) < 4 {
		return fmt.Errorf("error decoding binary %v: expected at least 4 bytes, got %d", data, len(data))
	}

	// Extract the exponent
	exp := int32(binary.BigEndian.Uint32(data[:4]))
	if int64(exp) > int64(MaxDecodeExponent) || int64(exp) < -int64(MaxDecodeExponent) {
		return fmt.Errorf("error decoding binary: exponent %d exceeds MaxDecodeExponent (%d)", exp, MaxDecodeExponent)
	}
	d.exp = exp

	// Extract the value
	d.value = new(big.Int)
	if err := d.value.GobDecode(data[4:]); err != nil {
		return fmt.Errorf("error decoding binary %v: %s", data, err)
	}

	return nil
}

// MarshalBinary implements the encoding.BinaryMarshaler interface.
func (d Decimal) MarshalBinary() (data []byte, err error) {
	// exp is written first, but encode value first to know output size
	var valueData []byte
	if valueData, err = d.getValue().GobEncode(); err != nil {
		return nil, err
	}

	// Write the exponent in front, since it's a fixed size
	expData := make([]byte, 4, len(valueData)+4)
	binary.BigEndian.PutUint32(expData, uint32(d.exp))

	// Return the byte array
	return append(expData, valueData...), nil
}

// Scan implements the sql.Scanner interface for database deserialization.
func (d *Decimal) Scan(value interface{}) error {
	// first try to see if the data is stored in database as a Numeric datatype
	switch v := value.(type) {

	case float32:
		*d = NewFromFloat(float64(v))
		return nil

	case float64:
		// numeric in sqlite3 sends us float64
		*d = NewFromFloat(v)
		return nil

	case int64:
		// at least in sqlite3 when the value is 0 in db, the data is sent
		// to us as an int64 instead of a float64 ...
		*d = New(v, 0)
		return nil

	case uint64:
		// while clickhouse may send 0 in db as uint64
		*d = NewFromUint64(v)
		return nil

	case string:
		var err error
		*d, err = NewFromString(unquoteIfQuoted(v))
		return err

	case []byte:
		var err error
		*d, err = NewFromString(unquoteIfQuoted(string(v)))
		return err

	default:
		return fmt.Errorf("could not convert value '%+v' to any known type", value)
	}
}

// Value implements the driver.Valuer interface for database serialization.
func (d Decimal) Value() (driver.Value, error) {
	return d.String(), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface for XML
// deserialization.
func (d *Decimal) UnmarshalText(text []byte) error {
	str := string(text)

	dec, err := NewFromString(str)
	*d = dec
	if err != nil {
		return fmt.Errorf("error decoding string '%s': %s", str, err)
	}

	return nil
}

// MarshalText implements the encoding.TextMarshaler interface for XML
// serialization.
func (d Decimal) MarshalText() (text []byte, err error) {
	return []byte(d.String()), nil
}

// GobEncode implements the gob.GobEncoder interface for gob serialization.
func (d Decimal) GobEncode() ([]byte, error) {
	return d.MarshalBinary()
}

// GobDecode implements the gob.GobDecoder interface for gob serialization.
func (d *Decimal) GobDecode(data []byte) error {
	return d.UnmarshalBinary(data)
}

// DecodeSpanner decodes a Spanner value into a Decimal
func (d *Decimal) DecodeSpanner(val interface{}) error {
	return d.Scan(val)
}

// EncodeSpanner encodes a Decimal into a Spanner value
func (d Decimal) EncodeSpanner() (interface{}, error) {
	return d.String(), nil
}

// StringScaled first scales the decimal then calls .String() on it.
//
// Deprecated: buggy and unintuitive. Use StringFixed instead.
func (d Decimal) StringScaled(exp int32) string {
	return d.rescale(exp).String()
}

func (d Decimal) string(trimTrailingZeros, useScientificNotation bool) string {
	if d.exp == 0 {
		return d.getValue().String()
	}
	if d.exp >= 0 {
		if useScientificNotation {
			return d.ScientificNotationString()
		} else {
			return d.rescale(0).value.String()
		}
	}

	str := d.getValue().String()
	sign := ""
	if str[0] == '-' {
		sign, str = "-", str[1:]
	}

	var intPart, leadingZeros, fractionalPart string

	// NOTE(vadim): this cast to int will cause bugs if d.exp == INT_MIN
	// and you are on a 32-bit machine. Won't fix this super-edge case.
	dExpInt := int(d.exp)
	if len(str) > -dExpInt {
		intPart = str[:len(str)+dExpInt]
		fractionalPart = str[len(str)+dExpInt:]
	} else {
		intPart = "0"

		num0s := -dExpInt - len(str)
		if num0s <= len(zeros) {
			leadingZeros = zeros[:num0s]
		} else {
			leadingZeros = strings.Repeat("0", num0s)
		}
		fractionalPart = str
	}

	if trimTrailingZeros {
		i := len(fractionalPart) - 1
		for ; i >= 0; i-- {
			if fractionalPart[i] != '0' {
				break
			}
		}
		fractionalPart = fractionalPart[:i+1]
		if fractionalPart == "" {
			leadingZeros = ""
		}
	}

	if len(leadingZeros)+len(fractionalPart) > 0 {
		return sign + intPart + "." + leadingZeros + fractionalPart
	}
	return sign + intPart
}

const zeros = "0000000000000000000000000000000000000000000000000000000000000000"

// ScientificNotationString serializes the decimal into standard scientific notation.
//
// The notation is normalized to have one non-zero digit followed by a decimal point and
// the remaining significant digits followed by "E" and the base-10 exponent.
//
// A zero, which has no significant digits, is simply serialized to "0".
func (d Decimal) ScientificNotationString() string {
	exp := int(d.exp)
	intStr := new(big.Int).Abs(d.getValue()).String()
	if intStr == "0" {
		return intStr
	}
	first := intStr[0]
	var remaining string
	if len(intStr) > 1 {
		remaining = "." + intStr[1:]
		exp = exp + len(intStr) - 1
	}
	number := string(first) + remaining + "E" + strconv.Itoa(exp)
	if d.value.Sign() < 0 {
		return "-" + number
	}
	return number
}

// Min returns the smallest Decimal that was passed in the arguments.
//
// To call this function with an array, you must do:
//
//	Min(arr[0], arr[1:]...)
//
// This makes it harder to accidentally call Min with 0 arguments.
func Min(first Decimal, rest ...Decimal) Decimal {
	ans := first
	for _, item := range rest {
		if item.Cmp(ans) < 0 {
			ans = item
		}
	}
	return ans
}

// Max returns the largest Decimal that was passed in the arguments.
//
// To call this function with an array, you must do:
//
//	Max(arr[0], arr[1:]...)
//
// This makes it harder to accidentally call Max with 0 arguments.
func Max(first Decimal, rest ...Decimal) Decimal {
	ans := first
	for _, item := range rest {
		if item.Cmp(ans) > 0 {
			ans = item
		}
	}
	return ans
}

// Sum returns the combined total of the provided first and rest Decimals
func Sum(first Decimal, rest ...Decimal) Decimal {
	if len(rest) == 0 {
		return first
	}

	// Add returns a new value, so the following items can be added to it in place
	total := first.Add(rest[0])
	last := len(rest) - 1
	if last == 0 {
		return total
	}
	var scaled big.Int
	for _, item := range rest[1:last] {
		switch {
		case item.exp < total.exp:
			total = total.rescale(item.exp)
			total.value.Add(total.value, item.getValue())
		case item.exp > total.exp:
			total.value.Add(total.value, scaled.Mul(item.getValue(), pow10(int64(item.exp)-int64(total.exp))))
		default:
			total.value.Add(total.value, item.getValue())
		}
	}

	// the last Add builds the result the same way as adding one item at a time does
	return total.Add(rest[last])
}

// Avg returns the average value of the provided first and rest Decimals
func Avg(first Decimal, rest ...Decimal) Decimal {
	count := New(int64(len(rest)+1), 0)
	sum := Sum(first, rest...)
	return sum.Div(count)
}

// RescalePair rescales two decimals to common exponential value (minimal exp of both decimals)
func RescalePair(d1 Decimal, d2 Decimal) (Decimal, Decimal) {
	if d1.exp < d2.exp {
		return d1, d2.rescale(d1.exp)
	} else if d1.exp > d2.exp {
		return d1.rescale(d2.exp), d2
	}

	return d1, d2
}

func unquoteIfQuoted(value string) string {
	// If the amount is quoted, strip the quotes
	if len(value) > 2 && value[0] == '"' && value[len(value)-1] == '"' {
		return value[1 : len(value)-1]
	}

	return value
}

// NullDecimal represents a nullable decimal with compatibility for
// scanning null values from the database.
type NullDecimal struct {
	Decimal Decimal
	Valid   bool
}

func NewNullDecimal(d Decimal) NullDecimal {
	return NullDecimal{
		Decimal: d,
		Valid:   true,
	}
}

// Scan implements the sql.Scanner interface for database deserialization.
func (d *NullDecimal) Scan(value interface{}) error {
	if value == nil {
		d.Valid = false
		return nil
	}
	d.Valid = true
	return d.Decimal.Scan(value)
}

// Value implements the driver.Valuer interface for database serialization.
func (d NullDecimal) Value() (driver.Value, error) {
	if !d.Valid {
		return nil, nil
	}
	return d.Decimal.Value()
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (d *NullDecimal) UnmarshalJSON(decimalBytes []byte) error {
	if string(decimalBytes) == "null" {
		d.Valid = false
		return nil
	}
	d.Valid = true
	return d.Decimal.UnmarshalJSON(decimalBytes)
}

// MarshalJSON implements the json.Marshaler interface.
func (d NullDecimal) MarshalJSON() ([]byte, error) {
	if !d.Valid {
		return []byte("null"), nil
	}
	return d.Decimal.MarshalJSON()
}

// UnmarshalText implements the encoding.TextUnmarshaler interface for XML
// deserialization
func (d *NullDecimal) UnmarshalText(text []byte) error {
	str := string(text)

	// check for empty XML or XML without body e.g., <tag></tag>
	if str == "" {
		d.Valid = false
		return nil
	}
	if err := d.Decimal.UnmarshalText(text); err != nil {
		d.Valid = false
		return err
	}
	d.Valid = true
	return nil
}

// MarshalText implements the encoding.TextMarshaler interface for XML
// serialization.
func (d NullDecimal) MarshalText() (text []byte, err error) {
	if !d.Valid {
		return []byte{}, nil
	}
	return d.Decimal.MarshalText()
}

// DecodeSpanner decodes a Spanner value into a Decimal
func (d *NullDecimal) DecodeSpanner(value interface{}) error {
	switch t := value.(type) {
	case nil:
		d.Valid = false
		return nil
	case *string:
		if t == nil {
			d.Valid = false
			return nil
		}
		value = *t
	}
	d.Valid = true

	return d.Decimal.Scan(value)
}

// EncodeSpanner encodes a Decimal into a Spanner value
func (d NullDecimal) EncodeSpanner() (interface{}, error) {
	if !d.Valid {
		return nil, nil
	}
	return d.Decimal.String(), nil
}

// Trig functions

var (
	// Pi/4 split into three parts
	pi4A = NewFromFloat(7.85398125648498535156e-1)                             // 0x3fe921fb40000000
	pi4B = NewFromFloat(3.77489470793079817668e-8)                             // 0x3e64442d00000000
	pi4C = NewFromFloat(2.69515142907905952645e-15)                            // 0x3ce8469898cc5170
	m4PI = NewFromFloat(1.273239544735162542821171882678754627704620361328125) // 4/pi

	atanP0 = NewFromFloat(-8.750608600031904122785e-01)
	atanP1 = NewFromFloat(-1.615753718733365076637e+01)
	atanP2 = NewFromFloat(-7.500855792314704667340e+01)
	atanP3 = NewFromFloat(-1.228866684490136173410e+02)
	atanP4 = NewFromFloat(-6.485021904942025371773e+01)
	atanQ0 = NewFromFloat(2.485846490142306297962e+01)
	atanQ1 = NewFromFloat(1.650270098316988542046e+02)
	atanQ2 = NewFromFloat(4.328810604912902668951e+02)
	atanQ3 = NewFromFloat(4.853903996359136964868e+02)
	atanQ4 = NewFromFloat(1.945506571482613964425e+02)

	atanMorebits = NewFromFloat(6.123233995736765886130e-17) // pi/2 = PIO2 + Morebits
	tan3pio8     = NewFromFloat(2.41421356237309504880)      // tan(3*pi/8)
	piFloat      = NewFromFloat(3.14159265358979323846264338327950288419716939937510582097494459)
)

// Atan returns the arctangent, in radians, of x.
func (d Decimal) Atan() Decimal {
	if d.IsZero() {
		return d
	}
	if d.IsPositive() {
		return d.satan()
	}
	return d.Neg().satan().Neg()
}

func (d Decimal) xatan() Decimal {
	P0, P1, P2, P3, P4 := atanP0, atanP1, atanP2, atanP3, atanP4
	Q0, Q1, Q2, Q3, Q4 := atanQ0, atanQ1, atanQ2, atanQ3, atanQ4
	z := d.Mul(d)
	b1 := P0.Mul(z).Add(P1).Mul(z).Add(P2).Mul(z).Add(P3).Mul(z).Add(P4).Mul(z)
	b2 := z.Add(Q0).Mul(z).Add(Q1).Mul(z).Add(Q2).Mul(z).Add(Q3).Mul(z).Add(Q4)
	z = b1.Div(b2)
	z = d.Mul(z).Add(d)
	return z
}

// satan reduces its argument (known to be positive)
// to the range [0, 0.66] and calls xatan.
func (d Decimal) satan() Decimal {
	Morebits, Tan3pio8, pi := atanMorebits, tan3pio8, piFloat

	if d.LessThanOrEqual(New(66, -2)) {
		return d.xatan()
	}
	if d.GreaterThan(Tan3pio8) {
		return pi.Div(New(2, 0)).Sub(New(1, 0).Div(d).xatan()).Add(Morebits)
	}
	return pi.Div(New(4, 0)).Add((d.Sub(New(1, 0)).Div(d.Add(New(1, 0)))).xatan()).Add(New(5, -1).Mul(Morebits))
}

// sin coefficients
var _sin = [...]Decimal{
	NewFromFloat(1.58962301576546568060e-10), // 0x3de5d8fd1fd19ccd
	NewFromFloat(-2.50507477628578072866e-8), // 0xbe5ae5e5a9291f5d
	NewFromFloat(2.75573136213857245213e-6),  // 0x3ec71de3567d48a1
	NewFromFloat(-1.98412698295895385996e-4), // 0xbf2a01a019bfdf03
	NewFromFloat(8.33333333332211858878e-3),  // 0x3f8111111110f7d0
	NewFromFloat(-1.66666666666666307295e-1), // 0xbfc5555555555548
}

// Sin returns the sine of the radian argument x.
func (d Decimal) Sin() Decimal {
	PI4A, PI4B, PI4C, M4PI := pi4A, pi4B, pi4C, m4PI

	if d.IsZero() {
		return d
	}
	// make argument positive but save the sign
	sign := false
	if d.IsNegative() {
		d = d.Neg()
		sign = true
	}

	j := d.Mul(M4PI).IntPart()    // integer part of x/(Pi/4), as integer for tests on the phase angle
	y := NewFromFloat(float64(j)) // integer part of x/(Pi/4), as float

	// map zeros to origin
	if j&1 == 1 {
		j++
		y = y.Add(New(1, 0))
	}
	j &= 7 // octant modulo 2Pi radians (360 degrees)
	// reflect in x axis
	if j > 3 {
		sign = !sign
		j -= 4
	}
	z := d.Sub(y.Mul(PI4A)).Sub(y.Mul(PI4B)).Sub(y.Mul(PI4C)) // Extended precision modular arithmetic
	zz := z.Mul(z)

	if j == 1 || j == 2 {
		w := zz.Mul(zz).Mul(_cos[0].Mul(zz).Add(_cos[1]).Mul(zz).Add(_cos[2]).Mul(zz).Add(_cos[3]).Mul(zz).Add(_cos[4]).Mul(zz).Add(_cos[5]))
		y = New(1, 0).Sub(New(5, -1).Mul(zz)).Add(w)
	} else {
		y = z.Add(z.Mul(zz).Mul(_sin[0].Mul(zz).Add(_sin[1]).Mul(zz).Add(_sin[2]).Mul(zz).Add(_sin[3]).Mul(zz).Add(_sin[4]).Mul(zz).Add(_sin[5])))
	}
	if sign {
		y = y.Neg()
	}
	return y
}

// cos coefficients
var _cos = [...]Decimal{
	NewFromFloat(-1.13585365213876817300e-11), // 0xbda8fa49a0861a9b
	NewFromFloat(2.08757008419747316778e-9),   // 0x3e21ee9d7b4e3f05
	NewFromFloat(-2.75573141792967388112e-7),  // 0xbe927e4f7eac4bc6
	NewFromFloat(2.48015872888517045348e-5),   // 0x3efa01a019c844f5
	NewFromFloat(-1.38888888888730564116e-3),  // 0xbf56c16c16c14f91
	NewFromFloat(4.16666666666665929218e-2),   // 0x3fa555555555554b
}

// Cos returns the cosine of the radian argument x.
func (d Decimal) Cos() Decimal {

	PI4A, PI4B, PI4C, M4PI := pi4A, pi4B, pi4C, m4PI

	// make argument positive
	sign := false
	if d.IsNegative() {
		d = d.Neg()
	}

	j := d.Mul(M4PI).IntPart()    // integer part of x/(Pi/4), as integer for tests on the phase angle
	y := NewFromFloat(float64(j)) // integer part of x/(Pi/4), as float

	// map zeros to origin
	if j&1 == 1 {
		j++
		y = y.Add(New(1, 0))
	}
	j &= 7 // octant modulo 2Pi radians (360 degrees)
	// reflect in x axis
	if j > 3 {
		sign = !sign
		j -= 4
	}
	if j > 1 {
		sign = !sign
	}

	z := d.Sub(y.Mul(PI4A)).Sub(y.Mul(PI4B)).Sub(y.Mul(PI4C)) // Extended precision modular arithmetic
	zz := z.Mul(z)

	if j == 1 || j == 2 {
		y = z.Add(z.Mul(zz).Mul(_sin[0].Mul(zz).Add(_sin[1]).Mul(zz).Add(_sin[2]).Mul(zz).Add(_sin[3]).Mul(zz).Add(_sin[4]).Mul(zz).Add(_sin[5])))
	} else {
		w := zz.Mul(zz).Mul(_cos[0].Mul(zz).Add(_cos[1]).Mul(zz).Add(_cos[2]).Mul(zz).Add(_cos[3]).Mul(zz).Add(_cos[4]).Mul(zz).Add(_cos[5]))
		y = New(1, 0).Sub(New(5, -1).Mul(zz)).Add(w)
	}
	if sign {
		y = y.Neg()
	}
	return y
}

var _tanP = [...]Decimal{
	NewFromFloat(-1.30936939181383777646e+4), // 0xc0c992d8d24f3f38
	NewFromFloat(1.15351664838587416140e+6),  // 0x413199eca5fc9ddd
	NewFromFloat(-1.79565251976484877988e+7), // 0xc1711fead3299176
}
var _tanQ = [...]Decimal{
	NewFromFloat(1.00000000000000000000e+0),
	NewFromFloat(1.36812963470692954678e+4),  // 0x40cab8a5eeb36572
	NewFromFloat(-1.32089234440210967447e+6), // 0xc13427bc582abc96
	NewFromFloat(2.50083801823357915839e+7),  // 0x4177d98fc2ead8ef
	NewFromFloat(-5.38695755929454629881e+7), // 0xc189afe03cbe5a31
}

// Tan returns the tangent of the radian argument x.
func (d Decimal) Tan() Decimal {

	PI4A, PI4B, PI4C, M4PI := pi4A, pi4B, pi4C, m4PI

	if d.IsZero() {
		return d
	}

	// make argument positive but save the sign
	sign := false
	if d.IsNegative() {
		d = d.Neg()
		sign = true
	}

	j := d.Mul(M4PI).IntPart()    // integer part of x/(Pi/4), as integer for tests on the phase angle
	y := NewFromFloat(float64(j)) // integer part of x/(Pi/4), as float

	// map zeros to origin
	if j&1 == 1 {
		j++
		y = y.Add(New(1, 0))
	}

	z := d.Sub(y.Mul(PI4A)).Sub(y.Mul(PI4B)).Sub(y.Mul(PI4C)) // Extended precision modular arithmetic
	zz := z.Mul(z)

	if zz.GreaterThan(New(1, -14)) {
		w := zz.Mul(_tanP[0].Mul(zz).Add(_tanP[1]).Mul(zz).Add(_tanP[2]))
		x := zz.Add(_tanQ[1]).Mul(zz).Add(_tanQ[2]).Mul(zz).Add(_tanQ[3]).Mul(zz).Add(_tanQ[4])
		y = z.Add(z.Mul(w.Div(x)))
	} else {
		y = z
	}
	if j&2 == 2 {
		y = New(-1, 0).Div(y)
	}
	if sign {
		y = y.Neg()
	}
	return y
}
