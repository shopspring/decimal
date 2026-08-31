package decimal

// EqualZero returns whether the numbers represented by d equals zero.
func (d Decimal) EqualZero() bool {
	return d.Equal(Zero)
}

// NotZero returns whether d is not zero
func (d Decimal) NotZero() bool {
	return !d.EqualZero()
}

// GreaterThanZero (GT0) returns true when d is greater than zero.
func (d Decimal) GreaterThanZero() bool {
	return d.GreaterThan(Zero)
}

// GreaterThanOrEqualZero (GTE0) returns true when d is greater than or equal to zero.
func (d Decimal) GreaterThanOrEqualZero() bool {
	return d.GreaterThanOrEqual(Zero)
}

// LessThanZero returns true when d is less than zero.
func (d Decimal) LessThanZero() bool {
	return d.LessThan(Zero)
}

// LessThanOrEqualZero returns true when d is less than or equal to zero.
func (d Decimal) LessThanOrEqualZero() bool {
	return d.LessThanOrEqual(Zero)
}
