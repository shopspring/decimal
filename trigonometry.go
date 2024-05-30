// shopspring/decimal/trigonometry.go

package decimal

// Trigonometry functions

// NOTE(davis): still using NewFromFloat(), as they pass unit tests (expected values)
var pi = NewFromFloat(3.14159265358979323846264338327950288419716939937510582097494459)
var pi4a = NewFromFloat(7.85398125648498535156e-1)                             // 0x3fe921fb40000000, Pi/4 split into three parts
var pi4b = NewFromFloat(3.77489470793079817668e-8)                             // 0x3e64442d00000000,
var pi4c = NewFromFloat(2.69515142907905952645e-15)                            // 0x3ce8469898cc5170,
var m4pi = NewFromFloat(1.273239544735162542821171882678754627704620361328125) // 4/pi

// Atan returns the arctangent, in radians, of x.
func (d Decimal) Atan() Decimal {
	if d.Equal(NewFromFloat(0.0)) {
		return d
	}
	if d.GreaterThan(NewFromFloat(0.0)) {
		return d.satan()
	}
	return d.Neg().satan().Neg()
}

var _xatan_P0 = RequireFromString("-8.750608600031904122785e-01")
var _xatan_P1 = RequireFromString("-1.615753718733365076637e+01")
var _xatan_P2 = RequireFromString("-7.500855792314704667340e+01")
var _xatan_P3 = RequireFromString("-1.228866684490136173410e+02")
var _xatan_P4 = RequireFromString("-6.485021904942025371773e+01")
var _xatan_Q0 = RequireFromString("2.485846490142306297962e+01")
var _xatan_Q1 = RequireFromString("1.650270098316988542046e+02")
var _xatan_Q2 = RequireFromString("4.328810604912902668951e+02")
var _xatan_Q3 = RequireFromString("4.853903996359136964868e+02")
var _xatan_Q4 = RequireFromString("1.945506571482613964425e+02")

func (d Decimal) xatan() Decimal {
	z := d.Mul(d)
	b1 := _xatan_P0.Mul(z).Add(_xatan_P1).Mul(z).Add(_xatan_P2).Mul(z).Add(_xatan_P3).Mul(z).Add(_xatan_P4).Mul(z)
	b2 := z.Add(_xatan_Q0).Mul(z).Add(_xatan_Q1).Mul(z).Add(_xatan_Q2).Mul(z).Add(_xatan_Q3).Mul(z).Add(_xatan_Q4)
	z = b1.Div(b2)
	z = d.Mul(z).Add(d)
	return z
}

var _satan_Morebits = NewFromFloat(6.123233995736765886130e-17) // pi/2 = PIO2 + Morebits
var _satan_Tan3pio8 = NewFromFloat(2.41421356237309504880)      // tan(3*pi/8)

// satan reduces its argument (known to be positive)
// to the range [0, 0.66] and calls xatan.
func (d Decimal) satan() Decimal {
	if d.LessThanOrEqual(NewFromFloat(0.66)) {
		return d.xatan()
	}
	if d.GreaterThan(_satan_Tan3pio8) {
		return pi.Div(NewFromFloat(2.0)).Sub(NewFromFloat(1.0).Div(d).xatan()).Add(_satan_Morebits)
	}
	return pi.Div(NewFromFloat(4.0)).Add((d.Sub(NewFromFloat(1.0)).Div(d.Add(NewFromFloat(1.0)))).xatan()).Add(NewFromFloat(0.5).Mul(_satan_Morebits))
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
	if d.Equal(NewFromFloat(0.0)) {
		return d
	}
	// make argument positive but save the sign
	sign := false
	if d.LessThan(NewFromFloat(0.0)) {
		d = d.Neg()
		sign = true
	}

	j := d.Mul(m4pi).IntPart()    // integer part of x/(Pi/4), as integer for tests on the phase angle
	y := NewFromFloat(float64(j)) // integer part of x/(Pi/4), as float

	// map zeros to origin
	if j&1 == 1 {
		j++
		y = y.Add(NewFromFloat(1.0))
	}
	j &= 7 // octant modulo 2Pi radians (360 degrees)
	// reflect in x axis
	if j > 3 {
		sign = !sign
		j -= 4
	}
	z := d.Sub(y.Mul(pi4a)).Sub(y.Mul(pi4b)).Sub(y.Mul(pi4c)) // Extended precision modular arithmetic
	zz := z.Mul(z)

	if j == 1 || j == 2 {
		w := zz.Mul(zz).Mul(_cos[0].Mul(zz).Add(_cos[1]).Mul(zz).Add(_cos[2]).Mul(zz).Add(_cos[3]).Mul(zz).Add(_cos[4]).Mul(zz).Add(_cos[5]))
		y = NewFromFloat(1.0).Sub(NewFromFloat(0.5).Mul(zz)).Add(w)
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
	// make argument positive
	sign := false
	if d.LessThan(NewFromFloat(0.0)) {
		d = d.Neg()
	}

	j := d.Mul(m4pi).IntPart()    // integer part of x/(Pi/4), as integer for tests on the phase angle
	y := NewFromFloat(float64(j)) // integer part of x/(Pi/4), as float

	// map zeros to origin
	if j&1 == 1 {
		j++
		y = y.Add(NewFromFloat(1.0))
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

	z := d.Sub(y.Mul(pi4a)).Sub(y.Mul(pi4b)).Sub(y.Mul(pi4c)) // Extended precision modular arithmetic
	zz := z.Mul(z)

	if j == 1 || j == 2 {
		y = z.Add(z.Mul(zz).Mul(_sin[0].Mul(zz).Add(_sin[1]).Mul(zz).Add(_sin[2]).Mul(zz).Add(_sin[3]).Mul(zz).Add(_sin[4]).Mul(zz).Add(_sin[5])))
	} else {
		w := zz.Mul(zz).Mul(_cos[0].Mul(zz).Add(_cos[1]).Mul(zz).Add(_cos[2]).Mul(zz).Add(_cos[3]).Mul(zz).Add(_cos[4]).Mul(zz).Add(_cos[5]))
		y = NewFromFloat(1.0).Sub(NewFromFloat(0.5).Mul(zz)).Add(w)
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
	NewFromFloat(1.36812963470692954678e+4),  //0x40cab8a5eeb36572
	NewFromFloat(-1.32089234440210967447e+6), //0xc13427bc582abc96
	NewFromFloat(2.50083801823357915839e+7),  //0x4177d98fc2ead8ef
	NewFromFloat(-5.38695755929454629881e+7), //0xc189afe03cbe5a31
}

// Tan returns the tangent of the radian argument x.
func (d Decimal) Tan() Decimal {
	if d.Equal(NewFromFloat(0.0)) {
		return d
	}

	// make argument positive but save the sign
	sign := false
	if d.LessThan(NewFromFloat(0.0)) {
		d = d.Neg()
		sign = true
	}

	j := d.Mul(m4pi).IntPart()    // integer part of x/(Pi/4), as integer for tests on the phase angle
	y := NewFromFloat(float64(j)) // integer part of x/(Pi/4), as float

	// map zeros to origin
	if j&1 == 1 {
		j++
		y = y.Add(NewFromFloat(1.0))
	}

	z := d.Sub(y.Mul(pi4a)).Sub(y.Mul(pi4b)).Sub(y.Mul(pi4c)) // Extended precision modular arithmetic
	zz := z.Mul(z)

	if zz.GreaterThan(NewFromFloat(1e-14)) {
		w := zz.Mul(_tanP[0].Mul(zz).Add(_tanP[1]).Mul(zz).Add(_tanP[2]))
		x := zz.Add(_tanQ[1]).Mul(zz).Add(_tanQ[2]).Mul(zz).Add(_tanQ[3]).Mul(zz).Add(_tanQ[4])
		y = z.Add(z.Mul(w.Div(x)))
	} else {
		y = z
	}
	if j&2 == 2 {
		y = NewFromFloat(-1.0).Div(y)
	}
	if sign {
		y = y.Neg()
	}
	return y
}
