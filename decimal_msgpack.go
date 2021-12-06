package decimal

import (
	"errors"
)

var (
	errShortBytes = errors.New("msgp: too few bytes left to read object")
)

// MarshalMsg implements msgp.Marshaler
//  Note: limit to 31 digits, if d.IntPart size large than 31, will be lose.
func (d Decimal) MarshalMsg(b []byte) (o []byte, err error) {
	o = require(b, d.Msgsize())
	str := d.String()
	sz := len(str)
	// limit to 31 digits
	// note, if d.IntPart size large than 31, will be lose.
	if sz > 31 {
		sz = 31
		// if last char is '.' then limit to 30 digits
		if str[30] == '.' {
			sz = 30
		}

		str = str[:sz]
	}

	o, n := ensure(b, 1+sz)
	o[n] = byte(0xa0 | sz)
	n++

	return o[:n+copy(o[n:], str)], nil
}

// UnmarshalMsg implements msgp.Unmarshaler
func (d *Decimal) UnmarshalMsg(b []byte) (o []byte, err error) {
	l := len(b)
	if l < 1 {
		return nil, errShortBytes
	}

	sz := int(b[0] & 0x1f)
	if len(b[1:]) < sz {
		err = errShortBytes
		return
	}
	if *d, err = NewFromString(string(b[1 : sz+1])); err == nil {
		o = b[sz:]
	}
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (d Decimal) Msgsize() int {
	return 31
}

// Require ensures that cap(old)-len(old) >= extra.
func require(old []byte, extra int) []byte {
	l := len(old)
	c := cap(old)
	r := l + extra
	if c >= r {
		return old
	} else if l == 0 {
		return make([]byte, 0, extra)
	}
	// the new size is the greater
	// of double the old capacity
	// and the sum of the old length
	// and the number of new bytes
	// necessary.
	c <<= 1
	if c < r {
		c = r
	}
	n := make([]byte, l, c)
	copy(n, old)
	return n
}

// ensure 'sz' extra bytes in 'b' btw len(b) and cap(b)
func ensure(b []byte, sz int) ([]byte, int) {
	l := len(b)
	c := cap(b)
	if c-l < sz {
		o := make([]byte, (2*c)+sz) // exponential growth
		n := copy(o, b)
		return o[:n+sz], n
	}
	return b[:l+sz], l
}
