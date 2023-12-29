package decimal

import "testing"

func TestConstApproximation(t *testing.T) {
	for _, testCase := range []struct {
		Const                 string
		Precision             int32
		ExpectedApproximation string
	}{
		{"2.3025850929940456840179914546", 0, "2"},
		{"2.3025850929940456840179914546", 1, "2.3"},
		{"2.3025850929940456840179914546", 3, "2.302"},
		{"2.3025850929940456840179914546", 5, "2.302585"},
		{"2.3025850929940456840179914546", 10, "2.302585092994045"},
		{"2.3025850929940456840179914546", 100, "2.3025850929940456840179914546"},
		{"2.3025850929940456840179914546", -1, "2"},
		{"2.3025850929940456840179914546", -5, "2"},
		{"3.14159265359", 0, "3"},
		{"3.14159265359", 1, "3.1"},
		{"3.14159265359", 2, "3.141"},
		{"3.14159265359", 4, "3.1415926"},
		{"3.14159265359", 13, "3.14159265359"},
	} {
		ca := newConstApproximation(testCase.Const)
		expected, _ := NewFromString(testCase.ExpectedApproximation)

		approximation := ca.withPrecision(testCase.Precision)

		if approximation.Cmp(expected) != 0 {
			t.Errorf("expected approximation %s, got %s - for const with %s precision %d", testCase.ExpectedApproximation, approximation.String(), testCase.Const, testCase.Precision)
		}
	}
}
