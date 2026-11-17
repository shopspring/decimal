package decimal

import (
	"math"
	"math/big"
	"testing"
)

func TestScientificNotationExponentWidth(t *testing.T) {
	for _, tc := range []struct {
		name        string
		coefficient string
		exponent    int32
		want        string
	}{
		{"positive_boundary", "12", math.MaxInt32, "1.2E2147483648"},
		{"negative_boundary", "-12", math.MaxInt32, "-1.2E2147483648"},
		{"larger_coefficient", "123456789012345678901", math.MaxInt32, "1.23456789012345678901E2147483667"},
		{"trailing_zeros", "100", math.MaxInt32, "1.00E2147483649"},
		{"single_digit_boundary", "7", math.MaxInt32, "7E2147483647"},
		{"below_boundary", "12", math.MaxInt32 - 1, "1.2E2147483647"},
		{"minimum_exponent", "12", math.MinInt32, "1.2E-2147483647"},
		{"ordinary", "1234", -2, "1.234E1"},
		{"zero", "0", math.MaxInt32, "0"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			coefficient, ok := new(big.Int).SetString(tc.coefficient, 10)
			if !ok {
				t.Fatal("invalid test coefficient")
			}
			d := NewFromBigInt(coefficient, tc.exponent)
			if got := d.ScientificNotationString(); got != tc.want {
				t.Errorf("ScientificNotationString() = %q, want %q", got, tc.want)
			}
			if d.Coefficient().String() != tc.coefficient || d.Exponent() != tc.exponent {
				t.Fatal("formatting changed the input")
			}
		})
	}
}

func TestStringScientificNotationExponentWidth(t *testing.T) {
	old := UseScientificNotation
	UseScientificNotation = true
	defer func() { UseScientificNotation = old }()
	for _, tc := range []struct {
		name  string
		value int64
		want  string
	}{
		{"positive_boundary", 12, "1.2E2147483648"},
		{"negative_boundary", -12, "-1.2E2147483648"},
		{"single_digit_boundary", 7, "7E2147483647"},
		{"zero", 0, "0"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			d := New(tc.value, math.MaxInt32)
			if got := d.String(); got != tc.want {
				t.Errorf("String() = %q, want %q", got, tc.want)
			}
			text, err := d.MarshalText()
			if err != nil || string(text) != tc.want {
				t.Errorf("MarshalText() = %q, %v; want %q, nil", text, err, tc.want)
			}
			value, err := d.Value()
			if err != nil || value != tc.want {
				t.Errorf("Value() = %q, %v; want %q, nil", value, err, tc.want)
			}
		})
	}
}
