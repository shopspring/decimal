package decimal

import (
	"testing"
)

func TestMsgPack(t *testing.T) {
	for _, x := range testTable {
		s := x.short
		 // limit to 31 digits
		if len(s) > 31 {
			s = s[:31]
			if s[30] == '.' {
				s = s[:30]
			}
		}

		// Prepare Test Decimal Data
		amount, err := NewFromString(s)
		if err != nil{
			t.Error(err)
		}

		// MarshalMsg
		var b []byte
	 	out, err := amount.MarshalMsg(b)
		if err != nil{
			t.Errorf("error marshalMsg %s: %v", s, err)
		}

		// check msg type
		typ := out[0] & 0xe0
		if typ != 0xa0 {
			t.Errorf("error marshalMsg, expected type = %b, got %b", 0xa0, typ)
		}

		// check msg len
		sz := int(out[0] & 0x1f)
		if sz != len(s) {
			t.Errorf("error marshalMsg, expected size = %d, got %d", len(s), sz)
		}

		// UnmarshalMsg
		var unmarshalAmount Decimal
		_, err = unmarshalAmount.UnmarshalMsg(out)
		if err != nil{
			t.Errorf("error unmarshalMsg %s: %v", s, err)
		}else if !unmarshalAmount.Equal(amount) {
				t.Errorf("expected %s, got %s (%s, %d)",
					amount.String(), unmarshalAmount.String(),
					unmarshalAmount.value.String(), unmarshalAmount.exp)
		}
	}
}