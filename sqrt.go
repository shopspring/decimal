package decimal

const SqrtMaxIter 100000

// Sqrt returns the square root of d, the result will have
// DivisionPrecision digits after the decimal point.
func (d Decimal) Sqrt() Decimal {
	s, _ := d.SqrtRound(int32(decimal.DivisionPrecision))
	return s
}

// SqrtRound returns the square root of d, the result will have
// precision digits after the decimal point. The bool precise returns whether the precision was reached
func (d Decimal) SqrtRound(precision int32) (Decimal, bool) {
	cutoff := New(1, -precision)
	lo := Zero
	hi := d
	var mid Decimal
	for i := 0; i < SqrtMaxIter; i++ {
		//mid = (lo+hi)/2;
		mid = lo.Add(hi).DivRound(New(2, 0), precision)
		if mid.Mul(mid).Sub(d).Abs().LessThan(cutoff) {
			return mid, true
		}
		if mid.Mul(mid).GreaterThan(d) {
			hi = mid
		} else {
			lo = mid
		}
	}
	return mid, false
}
