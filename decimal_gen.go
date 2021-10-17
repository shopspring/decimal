package decimal

import (
	"errors"
)

// const (
// 	first3 = 0xe0
// 	last5  = 0x1f
//
// 	mfixstr uint8 = 0xa0
// 	mstr8   uint8 = 0xd9
// 	mstr16  uint8 = 0xda
// 	mstr32  uint8 = 0xdb
//
// 	stringPrefixSize = 5
// )

var (
	errShortBytes = errors.New("msgp: too few bytes left to read object")
	//bigEndian     = binary.BigEndian
)


// MarshalMsg implements msgp.Marshaler
func (z Decimal) MarshalMsg(b []byte) (o []byte, err error) {
	o = require(b, z.Msgsize())
	str := z.String()
	sz := len(str)
	if sz > 30 {
		sz = 30
		if str[29] == '.' {
			sz = 29
		}

		str = str[:sz]
	}

	o, n := ensure(b, 1+sz)
	o[n] = byte(0xa0 | sz)
	n++

	return o[:n+copy(o[n:], str)], nil
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *Decimal) UnmarshalMsg(b []byte) (o []byte, err error) {
	l := len(b)
	if l < 1 {
		return nil, errShortBytes
	}

	sz := int(b[0] & 0x1f)
	if len(b[1:]) < sz {
		err = errShortBytes
		return
	}
	if *z, err = NewFromString(string(b[1 : sz+1])); err == nil {
		o = b[sz:]
	}
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z Decimal) Msgsize() int {
	return 31
}
//
// // MarshalMsg implements msgp.Marshaler
// func (z Decimal) MarshalMsg(b []byte) (o []byte, err error) {
// 	o = require(b, z.Msgsize())
// 	o = appendString(o, z.String())
// 	return
// }
//
// // UnmarshalMsg implements msgp.Unmarshaler
// func (z *Decimal) UnmarshalMsg(b []byte) (o []byte, err error) {
// 	v, bts, err := readString(b)
// 	if err != nil {
// 		return
// 	}
// 	d, err := NewFromString(string(v))
// 	if err == nil {
// 		*z = d
// 		o = bts
// 	}
//
// 	return
// }
//
// // Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
// func (z Decimal) Msgsize() (s int) {
// 	s = stringPrefixSize + len(z.String())
// 	return
// }

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

// // AppendString appends a string as a MessagePack 'str' to the slice
// func appendString(b []byte, s string) []byte {
// 	sz := len(s)
// 	var n int
// 	var o []byte
// 	switch {
// 	case sz <= 31:
// 		o, n = ensure(b, 1+sz)
// 		o[n] = wfixstr(uint8(sz))
// 		n++
// 	case sz <= math.MaxUint8:
// 		o, n = ensure(b, 2+sz)
// 		prefixu8(o[n:], mstr8, uint8(sz))
// 		n += 2
// 	case sz <= math.MaxUint16:
// 		o, n = ensure(b, 3+sz)
// 		prefixu16(o[n:], mstr16, uint16(sz))
// 		n += 3
// 	default:
// 		o, n = ensure(b, 5+sz)
// 		prefixu32(o[n:], mstr32, uint32(sz))
// 		n += 5
// 	}
// 	return o[:n+copy(o[n:], s)]
// }
//
// func wfixstr(u uint8) byte {
// 	return (u & last5) | mfixstr
// }
//
// // write prefix and uint8
// func prefixu8(b []byte, pre byte, sz uint8) {
// 	b[0] = pre
// 	b[1] = sz
// }
//
// // write prefix and big-endian uint16
// func prefixu16(b []byte, pre byte, sz uint16) {
// 	b[0] = pre
// 	b[1] = byte(sz >> 8)
// 	b[2] = byte(sz)
// }
//
// // write prefix and big-endian uint32
// func prefixu32(b []byte, pre byte, sz uint32) {
// 	b[0] = pre
// 	b[1] = byte(sz >> 24)
// 	b[2] = byte(sz >> 16)
// 	b[3] = byte(sz >> 8)
// 	b[4] = byte(sz)
// }
//
// // readString reads a messagepack string field
// // without copying. The returned []byte points
// // to the same memory as the input slice.
// // Possible errors:
// // - ErrShortBytes (b not long enough)
// func readString(b []byte) (v []byte, o []byte, err error) {
// 	l := len(b)
// 	if l < 1 {
// 		return nil, nil, ErrShortBytes
// 	}
//
// 	lead := b[0]
// 	var read int
//
// 	if isfixstr(lead) {
// 		read = int(rfixstr(lead))
// 		b = b[1:]
// 	} else {
// 		switch lead {
// 		case mstr8:
// 			if l < 2 {
// 				err = ErrShortBytes
// 				return
// 			}
// 			read = int(b[1])
// 			b = b[2:]
//
// 		case mstr16:
// 			if l < 3 {
// 				err = ErrShortBytes
// 				return
// 			}
// 			read = int(bigEndian.Uint16(b[1:]))
// 			b = b[3:]
//
// 		case mstr32:
// 			if l < 5 {
// 				err = ErrShortBytes
// 				return
// 			}
// 			read = int(bigEndian.Uint32(b[1:]))
// 			b = b[5:]
//
// 		default:
// 			err = fmt.Errorf(`msgp: attempted to decode type 'str' with method for %d`, lead)
// 			return
// 		}
// 	}
//
// 	if len(b) < read {
// 		err = ErrShortBytes
// 		return
// 	}
//
// 	v = b[0:read]
// 	o = b[read:]
// 	return
// }
//
// func isfixstr(b byte) bool {
// 	return b&first3 == mfixstr
// }
//
// func rfixstr(b byte) uint8 {
// 	return b & last5
// }
