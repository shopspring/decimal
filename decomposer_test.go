package decimal

import (
	"testing"
)

func TestDecomposerRoundTrip(t *testing.T) {
	list := []struct {
		N string // Name.
		S string // String value.
		E bool   // Expect an error.
	}{
		{N: "Normal-1", S: "123.456"},
		{N: "Normal-2", S: "-123.456"},
		{N: "Large-1", S: "9937443000"},
		{N: "Large-2", S: "-9937443000"},
		{N: "AllDecimal-1", S: "0.04828239821"},
		{N: "AllDecimal-2", S: "-0.04828239821"},
		{N: "LargeAndDecimal-1", S: "4378273024.234239278"},
		{N: "LargeAndDecimal-2", S: "-4378273024.234239278"},
		{N: "Zero", S: "0"},
		{N: "One-1", S: "1"},
		{N: "One-2", S: "-1"},
		{N: "LargerThenFloat-1", S: "1234567890123456842"},
		{N: "LargerThenFloat-2", S: "-1234567890123456842"},
	}
	for _, item := range list {
		d, err := NewFromString(item.S)
		if err != nil {
			t.Fatal(err)
		}
		set := &Decimal{}
		err = set.Compose(d.Decompose(nil))
		if err == nil && item.E {
			t.Fatal("expected error, got <nil>")
		}
		if err != nil && !item.E {
			t.Fatalf("unexpected error: %v", err)
		}
		if set.Cmp(d) != 0 {
			t.Fatalf("values incorrect, got %v want %v (%s)", set, d, item.S)
		}
		if set.String() != item.S {
			t.Fatalf("string value incorrect, got %q want %q", set.String(), item.S)
		}
	}
}

func TestDecomposerCompose(t *testing.T) {
	list := []struct {
		N string // Name.
		S string // String value.

		Form byte // Form
		Neg  bool
		Coef []byte // Coefficent
		Exp  int32

		Err bool // Expect an error.
	}{
		{N: "Zero", S: "0", Coef: nil, Exp: 0},
		{N: "Normal-1", S: "123.456", Coef: []byte{0x01, 0xE2, 0x40}, Exp: -3},
		{N: "Neg-1", S: "-123.456", Neg: true, Coef: []byte{0x01, 0xE2, 0x40}, Exp: -3},
		{N: "PosExp-1", S: "123456000", Coef: []byte{0x01, 0xE2, 0x40}, Exp: 3},
		{N: "PosExp-2", S: "-123456000", Neg: true, Coef: []byte{0x01, 0xE2, 0x40}, Exp: 3},
		{N: "AllDec-1", S: "0.123456", Coef: []byte{0x01, 0xE2, 0x40}, Exp: -6},
		{N: "AllDec-2", S: "-0.123456", Neg: true, Coef: []byte{0x01, 0xE2, 0x40}, Exp: -6},
		{N: "NaN-1", S: "NaN", Form: 2, Err: true},
		{N: "NaN-2", S: "-NaN", Form: 2, Neg: true, Err: true},
		{N: "Infinity-1", S: "Infinity", Form: 1, Err: true},
		{N: "Infinity-2", S: "-Infinity", Form: 1, Neg: true, Err: true},
	}

	for _, item := range list {
		d := &Decimal{}
		err := d.Compose(item.Form, item.Neg, item.Coef, item.Exp)
		if err != nil && !item.Err {
			t.Fatalf("unexpected error, got %v", err)
		}
		if item.Err {
			if err == nil {
				t.Fatal("expected error, got <nil>")
			}
			return
		}
		if s := d.String(); s != item.S {
			t.Fatalf("unexpected value, got %q want %q", s, item.S)
		}
	}
}
