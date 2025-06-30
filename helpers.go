package decimal

// Float returns the decimal value as a float64
// float is a helper function for float64() and doesnot return bool exact
func (d Decimal) Float() float64 {
	f, _ := d.Float64()
	return f
}

// NotEqual returns true when d is not equal to d2
func (d Decimal) NotEqual(d2 Decimal) bool {
	return !d.Equal(d2)
}

// Max returns the maximum value between d and d2
func (d Decimal) Max(d2 Decimal) Decimal {
	return Max(d, d2)
}

// Min returns the minimum value between d and d2
func (d Decimal) Min(d2 Decimal) Decimal {
	return Min(d, d2)
}

//NewFromInt returns a decimal with the value of int v
func NewFromInt(v int) Decimal {
	return New(int64(v), 0)
}

//NewFromInt32 returns a decimal with the value of int v
func NewFromInt32(v int) Decimal {
	return New(int64(v), 0)
}

//NewFromInt64 returns a decimal with the value of int v
func NewFromInt64(v int64) Decimal {
	return New(v, 0)
}
