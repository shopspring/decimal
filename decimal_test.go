package decimal

import (
	"database/sql/driver"
	"encoding/json"
	"encoding/xml"
	"math"
	"math/big"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"
)

type testEnt struct {
	float   float64
	short   string
	exact   string
	inexact string
}

var testTable = []*testEnt{
	{3.141592653589793, "3.141592653589793", "", "3.14159265358979300000000000000000000000000000000000004"},
	{3, "3", "", "3.0000000000000000000000002"},
	{1234567890123456, "1234567890123456", "", "1234567890123456.00000000000000002"},
	{1234567890123456000, "1234567890123456000", "", "1234567890123456000.0000000000000008"},
	{1234.567890123456, "1234.567890123456", "", "1234.5678901234560000000000000009"},
	{.1234567890123456, "0.1234567890123456", "", "0.12345678901234560000000000006"},
	{0, "0", "", "0.000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000001"},
	{.1111111111111110, "0.111111111111111", "", "0.111111111111111000000000000000009"},
	{.1111111111111111, "0.1111111111111111", "", "0.111111111111111100000000000000000000023423545644534234"},
	{.1111111111111119, "0.1111111111111119", "", "0.111111111111111900000000000000000000000000000000000134123984192834"},
	{.000000000000000001, "0.000000000000000001", "", "0.00000000000000000100000000000000000000000000000000012341234"},
	{.000000000000000002, "0.000000000000000002", "", "0.0000000000000000020000000000000000000012341234123"},
	{.000000000000000003, "0.000000000000000003", "", "0.00000000000000000299999999999999999999999900000000000123412341234"},
	{.000000000000000005, "0.000000000000000005", "", "0.00000000000000000500000000000000000023412341234"},
	{.000000000000000008, "0.000000000000000008", "", "0.0000000000000000080000000000000000001241234432"},
	{.1000000000000001, "0.1000000000000001", "", "0.10000000000000010000000000000012341234"},
	{.1000000000000002, "0.1000000000000002", "", "0.10000000000000020000000000001234123412"},
	{.1000000000000003, "0.1000000000000003", "", "0.1000000000000003000000000000001234123412"},
	{.1000000000000005, "0.1000000000000005", "", "0.1000000000000005000000000000000006441234"},
	{.1000000000000008, "0.1000000000000008", "", "0.100000000000000800000000000000000009999999999999999999999999999"},
	{1e25, "10000000000000000000000000", "", "10000000000000000000000000.00000000000000000098798978"},
	{math.MaxInt64, strconv.FormatInt(math.MaxInt64, 10), "", strconv.FormatInt(math.MaxInt64, 10)},
	{1.29067116156722e-309, "0", "", "0.000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000001290671161567218558822290567835270536800098852722416870074139002112543896676308448335063375297788379444685193974290737962187240854947838776604607190387984577130572928111657710645015086812756013489109884753559084166516937690932698276436869274093950997935137476803610007959500457935217950764794724766740819156974617155861568214427828145972181876775307023388139991104942469299524961281641158436752347582767153796914843896176260096039358494077706152272661453132497761307744086665088096215425146090058519888494342944692629602847826300550628670375451325582843627504604013541465361435761965354140678551369499812124085312128659002910905639984075064968459581691226705666561364681985266583563078466180095375402399087817404368974165082030458595596655868575908243656158447265625000000000000000000000000000000000000004440000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"},
}

var testTableScientificNotation = map[string]string{
	"1e9":        "1000000000",
	"2.41E-3":    "0.00241",
	"24.2E-4":    "0.00242",
	"243E-5":     "0.00243",
	"1e-5":       "0.00001",
	"245E3":      "245000",
	"1.2345E-1":  "0.12345",
	"0e5":        "0",
	"0e-5":       "0",
	"123.456e0":  "123.456",
	"123.456e2":  "12345.6",
	"123.456e10": "1234560000000",
}

func init() {
	for _, s := range testTable {
		s.exact = strconv.FormatFloat(s.float, 'f', 1500, 64)
		if strings.ContainsRune(s.exact, '.') {
			s.exact = strings.TrimRight(s.exact, "0")
			s.exact = strings.TrimRight(s.exact, ".")
		}
	}

	// add negatives
	withNeg := testTable[:]
	for _, s := range testTable {
		if s.float > 0 && s.short != "0" && s.exact != "0" {
			withNeg = append(withNeg, &testEnt{-s.float, "-" + s.short, "-" + s.exact, "-" + s.inexact})
		}
	}
	testTable = withNeg

	for e, s := range testTableScientificNotation {
		if string(e[0]) != "-" && s != "0" {
			testTableScientificNotation["-"+e] = "-" + s
		}
	}
}

func TestNewFromFloat(t *testing.T) {
	for _, x := range testTable {
		s := x.exact
		d := NewFromFloat(x.float)
		if d.String() != s {
			t.Errorf("expected %s, got %s (%s, %d)",
				s, d.String(),
				d.value.String(), d.exp)
		}
	}

	shouldPanicOn := []float64{
		math.NaN(),
		math.Inf(1),
		math.Inf(-1),
	}

	for _, n := range shouldPanicOn {
		var d Decimal
		if !didPanic(func() { d = NewFromFloat(n) }) {
			t.Fatalf("Expected panic when creating a Decimal from %v, got %v instead", n, d.String())
		}
	}
}

func TestNewFromString(t *testing.T) {
	for _, x := range testTable {
		s := x.short
		d, err := NewFromString(s)
		if err != nil {
			t.Errorf("error while parsing %s", s)
		} else if d.String() != s {
			t.Errorf("expected %s, got %s (%s, %d)",
				s, d.String(),
				d.value.String(), d.exp)
		}
	}

	for _, x := range testTable {
		s := x.exact
		d, err := NewFromString(s)
		if err != nil {
			t.Errorf("error while parsing %s", s)
		} else if d.String() != s {
			t.Errorf("expected %s, got %s (%s, %d)",
				s, d.String(),
				d.value.String(), d.exp)
		}
	}

	for e, s := range testTableScientificNotation {
		d, err := NewFromString(e)
		if err != nil {
			t.Errorf("error while parsing %s", e)
		} else if d.String() != s {
			t.Errorf("expected %s, got %s (%s, %d)",
				s, d.String(),
				d.value.String(), d.exp)
		}
	}
}

func TestFloat64(t *testing.T) {
	for _, x := range testTable {
		s := x.exact
		d, err := NewFromString(s)
		if err != nil {
			t.Errorf("error while parsing %s", s)
		} else if f, exact := d.Float64(); !exact || f != x.float {
			t.Errorf("cannot represent exactly %s", s)
		}
		s = x.inexact
		d, err = NewFromString(s)
		if err != nil {
			t.Errorf("error while parsing %s", s)
		} else if f, exact := d.Float64(); exact || f != x.float {
			t.Errorf("%s should be represented inexactly", s)
		}
	}
}

func TestNewFromStringErrs(t *testing.T) {
	tests := []string{
		"",
		"qwert",
		"-",
		".",
		"-.",
		".-",
		"234-.56",
		"234-56",
		"2-",
		"..",
		"2..",
		"..2",
		".5.2",
		"8..2",
		"8.1.",
		"1e",
		"1-e",
		"1e9e",
		"1ee9",
		"1ee",
		"1eE",
		"1e-",
		"1e-.",
		"1e1.2",
		"123.456e1.3",
		"1e-1.2",
		"123.456e-1.3",
		"123.456Easdf",
		"123.456e" + strconv.FormatInt(math.MinInt64, 10),
		"123.456e" + strconv.FormatInt(math.MinInt32, 10),
	}

	for _, s := range tests {
		_, err := NewFromString(s)

		if err == nil {
			t.Errorf("error expected when parsing %s", s)
		}
	}
}

func TestNewFromStringDeepEquals(t *testing.T) {
	type StrCmp struct {
		str1     string
		str2     string
		expected bool
	}
	tests := []StrCmp{
		{"1", "1", true},
		{"10", "10.0", true},
		{"1.1", "1.10", true},
		{"1.001", "1.01", false},
	}

	for _, cmp := range tests {
		d1, err1 := NewFromString(cmp.str1)
		d2, err2 := NewFromString(cmp.str2)

		if err1 != nil || err2 != nil {
			t.Errorf("error parsing strings to decimals")
		}

		if reflect.DeepEqual(d1, d2) != cmp.expected {
			t.Errorf("comparison result is different from expected results for %s and %s",
				cmp.str1, cmp.str2)
		}
	}
}

func TestRequireFromString(t *testing.T) {
	s := "1.23"
	defer func() {
		err := recover()
		if err != nil {
			t.Errorf("error while parsing %s", s)
		}
	}()

	d := RequireFromString(s)
	if d.String() != s {
		t.Errorf("expected %s, got %s (%s, %d)",
			s, d.String(),
			d.value.String(), d.exp)
	}
}

func TestRequireFromStringErrs(t *testing.T) {
	s := "qwert"
	var d Decimal
	var err interface{}

	func(d Decimal) {
		defer func() {
			err = recover()
		}()

		d = RequireFromString(s)
	}(d)

	if err == nil {
		t.Errorf("panic expected when parsing %s", s)
	}
}

func TestNewFromFloatWithExponent(t *testing.T) {
	type Inp struct {
		float float64
		exp   int32
	}
	// some tests are taken from here https://www.cockroachlabs.com/blog/rounding-implementations-in-go/
	tests := map[Inp]string{
		Inp{123.4, -3}:                 "123.4",
		Inp{123.4, -1}:                 "123.4",
		Inp{123.412345, 1}:             "120",
		Inp{123.412345, 0}:             "123",
		Inp{123.412345, -5}:            "123.41235",
		Inp{123.412345, -6}:            "123.412345",
		Inp{123.412345, -7}:            "123.412345",
		Inp{123.412345, -28}:           "123.4123450000000019599610823207",
		Inp{1230000000, 3}:             "1230000000",
		Inp{123.9999999999999999, -7}:  "124",
		Inp{123.8989898999999999, -7}:  "123.8989899",
		Inp{0.49999999999999994, 0}:    "0",
		Inp{0.5, 0}:                    "1",
		Inp{0., -1000}:                 "0",
		Inp{0.5000000000000001, 0}:     "1",
		Inp{1.390671161567e-309, 0}:    "0",
		Inp{4.503599627370497e+15, 0}:  "4503599627370497",
		Inp{4.503599627370497e+60, 0}:  "4503599627370497110902645731364739935039854989106233267453952",
		Inp{4.503599627370497e+60, 1}:  "4503599627370497110902645731364739935039854989106233267453950",
		Inp{4.503599627370497e+60, -1}: "4503599627370497110902645731364739935039854989106233267453952",
		Inp{50, 2}:                     "100",
		Inp{49, 2}:                     "0",
		Inp{50, 3}:                     "0",
		// subnormals
		Inp{1.390671161567e-309, -2000}: "0.000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000001390671161567000864431395448332752540137009987788957394095829635554502771758698872408926974382819387852542087331897381878220271350970912568035007740861074263206736245957501456549756342151614772544950978154339064833880234531754156635411349342950306987480369774780312897442981323940546749863054846093718407237782253156822124910364044261653195961209878120072488178603782495270845071470243842997312255994555557251870400944414666445871039673491570643357351279578519863428540219295076767898526278029257129758694673164251056158277568765100904638511604478844087596428177947970563689475826736810456067108202083804368114484417399279328807983736233036662284338182105684628835292230438999173947056675615385756827890872955322265625",
		Inp{1.390671161567e-309, -862}:  "0.0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000013906711615670008644313954483327525401370099877889573940958296355545027717586988724089269743828193878525420873318973818782202713509709125680350077408610742632067362459575014565497563421516147725449509781543390648338802345317541566354113493429503069874803697747803128974429813239405467498630548460937184072377822531568221249103640442616531959612098781200724881786037824952708450714702438429973122559945555572518704009444146664458710396734915706433573512795785198634285402192950767678985262780292571297586946731642510561582775687651009046385116044788440876",
		Inp{1.390671161567e-309, -863}:  "0.0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000013906711615670008644313954483327525401370099877889573940958296355545027717586988724089269743828193878525420873318973818782202713509709125680350077408610742632067362459575014565497563421516147725449509781543390648338802345317541566354113493429503069874803697747803128974429813239405467498630548460937184072377822531568221249103640442616531959612098781200724881786037824952708450714702438429973122559945555572518704009444146664458710396734915706433573512795785198634285402192950767678985262780292571297586946731642510561582775687651009046385116044788440876",
	}

	// add negatives
	for p, s := range tests {
		if p.float > 0 {
			if s != "0" {
				tests[Inp{-p.float, p.exp}] = "-" + s
			} else {
				tests[Inp{-p.float, p.exp}] = "0"
			}
		}
	}

	for input, s := range tests {
		d := NewFromFloatWithExponent(input.float, input.exp)
		if d.String() != s {
			t.Errorf("expected %s, got %s (%s, %d)",
				s, d.String(),
				d.value.String(), d.exp)
		}
	}

	shouldPanicOn := []float64{
		math.NaN(),
		math.Inf(1),
		math.Inf(-1),
	}

	for _, n := range shouldPanicOn {
		var d Decimal
		if !didPanic(func() { d = NewFromFloatWithExponent(n, 0) }) {
			t.Fatalf("Expected panic when creating a Decimal from %v, got %v instead", n, d.String())
		}
	}
}

func TestNewFromBigIntWithExponent(t *testing.T) {
	type Inp struct {
		val *big.Int
		exp int32
	}
	tests := map[Inp]string{
		Inp{big.NewInt(123412345), -3}: "123412.345",
		Inp{big.NewInt(2234), -1}:      "223.4",
		Inp{big.NewInt(323412345), 1}:  "3234123450",
		Inp{big.NewInt(423412345), 0}:  "423412345",
		Inp{big.NewInt(52341235), -5}:  "523.41235",
		Inp{big.NewInt(623412345), -6}: "623.412345",
		Inp{big.NewInt(723412345), -7}: "72.3412345",
	}

	// add negatives
	for p, s := range tests {
		if p.val.Cmp(Zero.value) > 0 {
			tests[Inp{p.val.Neg(p.val), p.exp}] = "-" + s
		}
	}

	for input, s := range tests {
		d := NewFromBigInt(input.val, input.exp)
		if d.String() != s {
			t.Errorf("expected %s, got %s (%s, %d)",
				s, d.String(),
				d.value.String(), d.exp)
		}
	}
}

func TestJSON(t *testing.T) {
	for _, x := range testTable {
		s := x.short
		var doc struct {
			Amount Decimal `json:"amount"`
		}
		docStr := `{"amount":"` + s + `"}`
		docStrNumber := `{"amount":` + s + `}`
		err := json.Unmarshal([]byte(docStr), &doc)
		if err != nil {
			t.Errorf("error unmarshaling %s: %v", docStr, err)
		} else if doc.Amount.String() != s {
			t.Errorf("expected %s, got %s (%s, %d)",
				s, doc.Amount.String(),
				doc.Amount.value.String(), doc.Amount.exp)
		}

		out, err := json.Marshal(&doc)
		if err != nil {
			t.Errorf("error marshaling %+v: %v", doc, err)
		} else if string(out) != docStr {
			t.Errorf("expected %s, got %s", docStr, string(out))
		}

		// make sure unquoted marshalling works too
		MarshalJSONWithoutQuotes = true
		out, err = json.Marshal(&doc)
		if err != nil {
			t.Errorf("error marshaling %+v: %v", doc, err)
		} else if string(out) != docStrNumber {
			t.Errorf("expected %s, got %s", docStrNumber, string(out))
		}
		MarshalJSONWithoutQuotes = false
	}
}

func TestUnmarshalJSONNull(t *testing.T) {
	var doc struct {
		Amount Decimal `json:"amount"`
	}
	docStr := `{"amount": null}`
	err := json.Unmarshal([]byte(docStr), &doc)
	if err != nil {
		t.Errorf("error unmarshaling %s: %v", docStr, err)
	} else if !doc.Amount.Equal(Zero) {
		t.Errorf("expected Zero, got %s (%s, %d)",
			doc.Amount.String(),
			doc.Amount.value.String(), doc.Amount.exp)
	}
}

func TestBadJSON(t *testing.T) {
	for _, testCase := range []string{
		"]o_o[",
		"{",
		`{"amount":""`,
		`{"amount":""}`,
		`{"amount":"nope"}`,
		`0.333`,
	} {
		var doc struct {
			Amount Decimal `json:"amount"`
		}
		err := json.Unmarshal([]byte(testCase), &doc)
		if err == nil {
			t.Errorf("expected error, got %+v", doc)
		}
	}
}

func TestNullDecimalJSON(t *testing.T) {
	for _, x := range testTable {
		s := x.short
		var doc struct {
			Amount NullDecimal `json:"amount"`
		}
		docStr := `{"amount":"` + s + `"}`
		docStrNumber := `{"amount":` + s + `}`
		err := json.Unmarshal([]byte(docStr), &doc)
		if err != nil {
			t.Errorf("error unmarshaling %s: %v", docStr, err)
		} else {
			if !doc.Amount.Valid {
				t.Errorf("expected %s to be valid (not NULL), got Valid = false", s)
			}
			if doc.Amount.Decimal.String() != s {
				t.Errorf("expected %s, got %s (%s, %d)",
					s, doc.Amount.Decimal.String(),
					doc.Amount.Decimal.value.String(), doc.Amount.Decimal.exp)
			}
		}

		out, err := json.Marshal(&doc)
		if err != nil {
			t.Errorf("error marshaling %+v: %v", doc, err)
		} else if string(out) != docStr {
			t.Errorf("expected %s, got %s", docStr, string(out))
		}

		// make sure unquoted marshalling works too
		MarshalJSONWithoutQuotes = true
		out, err = json.Marshal(&doc)
		if err != nil {
			t.Errorf("error marshaling %+v: %v", doc, err)
		} else if string(out) != docStrNumber {
			t.Errorf("expected %s, got %s", docStrNumber, string(out))
		}
		MarshalJSONWithoutQuotes = false
	}

	var doc struct {
		Amount NullDecimal `json:"amount"`
	}
	docStr := `{"amount": null}`
	err := json.Unmarshal([]byte(docStr), &doc)
	if err != nil {
		t.Errorf("error unmarshaling %s: %v", docStr, err)
	} else if doc.Amount.Valid {
		t.Errorf("expected null value to have Valid = false, got Valid = true and Decimal = %s (%s, %d)",
			doc.Amount.Decimal.String(),
			doc.Amount.Decimal.value.String(), doc.Amount.Decimal.exp)
	}

	expected := `{"amount":null}`
	out, err := json.Marshal(&doc)
	if err != nil {
		t.Errorf("error marshaling %+v: %v", doc, err)
	} else if string(out) != expected {
		t.Errorf("expected %s, got %s", expected, string(out))
	}

	// make sure unquoted marshalling works too
	MarshalJSONWithoutQuotes = true
	expectedUnquoted := `{"amount":null}`
	out, err = json.Marshal(&doc)
	if err != nil {
		t.Errorf("error marshaling %+v: %v", doc, err)
	} else if string(out) != expectedUnquoted {
		t.Errorf("expected %s, got %s", expectedUnquoted, string(out))
	}
	MarshalJSONWithoutQuotes = false
}

func TestNullDecimalBadJSON(t *testing.T) {
	for _, testCase := range []string{
		"]o_o[",
		"{",
		`{"amount":""`,
		`{"amount":""}`,
		`{"amount":"nope"}`,
		`{"amount":nope}`,
		`0.333`,
	} {
		var doc struct {
			Amount NullDecimal `json:"amount"`
		}
		err := json.Unmarshal([]byte(testCase), &doc)
		if err == nil {
			t.Errorf("expected error, got %+v", doc)
		}
	}
}

func TestXML(t *testing.T) {
	for _, x := range testTable {
		s := x.short
		var doc struct {
			XMLName xml.Name `xml:"account"`
			Amount  Decimal  `xml:"amount"`
		}
		docStr := `<account><amount>` + s + `</amount></account>`
		err := xml.Unmarshal([]byte(docStr), &doc)
		if err != nil {
			t.Errorf("error unmarshaling %s: %v", docStr, err)
		} else if doc.Amount.String() != s {
			t.Errorf("expected %s, got %s (%s, %d)",
				s, doc.Amount.String(),
				doc.Amount.value.String(), doc.Amount.exp)
		}

		out, err := xml.Marshal(&doc)
		if err != nil {
			t.Errorf("error marshaling %+v: %v", doc, err)
		} else if string(out) != docStr {
			t.Errorf("expected %s, got %s", docStr, string(out))
		}
	}
}

func TestBadXML(t *testing.T) {
	for _, testCase := range []string{
		"o_o",
		"<abc",
		"<account><amount>7",
		`<html><body></body></html>`,
		`<account><amount></amount></account>`,
		`<account><amount>nope</amount></account>`,
		`0.333`,
	} {
		var doc struct {
			XMLName xml.Name `xml:"account"`
			Amount  Decimal  `xml:"amount"`
		}
		err := xml.Unmarshal([]byte(testCase), &doc)
		if err == nil {
			t.Errorf("expected error, got %+v", doc)
		}
	}
}

func TestDecimal_rescale(t *testing.T) {
	type Inp struct {
		int     int64
		exp     int32
		rescale int32
	}
	tests := map[Inp]string{
		Inp{1234, -3, -5}: "1.234",
		Inp{1234, -3, 0}:  "1",
		Inp{1234, 3, 0}:   "1234000",
		Inp{1234, -4, -4}: "0.1234",
	}

	// add negatives
	for p, s := range tests {
		if p.int > 0 {
			tests[Inp{-p.int, p.exp, p.rescale}] = "-" + s
		}
	}

	for input, s := range tests {
		d := New(input.int, input.exp).rescale(input.rescale)

		if d.String() != s {
			t.Errorf("expected %s, got %s (%s, %d)",
				s, d.String(),
				d.value.String(), d.exp)
		}

		// test StringScaled
		s2 := New(input.int, input.exp).StringScaled(input.rescale)
		if s2 != s {
			t.Errorf("expected %s, got %s", s, s2)
		}
	}
}

func TestDecimal_Floor(t *testing.T) {
	assertFloor := func(input, expected Decimal) {
		got := input.Floor()
		if !got.Equal(expected) {
			t.Errorf("Floor(%s): got %s, expected %s", input, got, expected)
		}
	}
	type testDataString struct {
		input    string
		expected string
	}
	testsWithStrings := []testDataString{
		{"1.999", "1"},
		{"1", "1"},
		{"1.01", "1"},
		{"0", "0"},
		{"0.9", "0"},
		{"0.1", "0"},
		{"-0.9", "-1"},
		{"-0.1", "-1"},
		{"-1.00", "-1"},
		{"-1.01", "-2"},
		{"-1.999", "-2"},
	}
	for _, test := range testsWithStrings {
		expected, _ := NewFromString(test.expected)
		input, _ := NewFromString(test.input)
		assertFloor(input, expected)
	}

	type testDataDecimal struct {
		input    Decimal
		expected string
	}
	testsWithDecimals := []testDataDecimal{
		{New(100, -1), "10"},
		{New(10, 0), "10"},
		{New(1, 1), "10"},
		{New(1999, -3), "1"},
		{New(101, -2), "1"},
		{New(1, 0), "1"},
		{New(0, 0), "0"},
		{New(9, -1), "0"},
		{New(1, -1), "0"},
		{New(-1, -1), "-1"},
		{New(-9, -1), "-1"},
		{New(-1, 0), "-1"},
		{New(-101, -2), "-2"},
		{New(-1999, -3), "-2"},
	}
	for _, test := range testsWithDecimals {
		expected, _ := NewFromString(test.expected)
		assertFloor(test.input, expected)
	}
}

func Benchmark_FloorFast(b *testing.B) {
	input := New(200, 2)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		input.Floor()
	}
}

func Benchmark_FloorRegular(b *testing.B) {
	input := New(200, -2)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		input.Floor()
	}
}

func TestDecimal_Ceil(t *testing.T) {
	assertCeil := func(input, expected Decimal) {
		got := input.Ceil()
		if !got.Equal(expected) {
			t.Errorf("Ceil(%s): got %s, expected %s", input, got, expected)
		}
	}
	type testDataString struct {
		input    string
		expected string
	}
	testsWithStrings := []testDataString{
		{"1.999", "2"},
		{"1", "1"},
		{"1.01", "2"},
		{"0", "0"},
		{"0.9", "1"},
		{"0.1", "1"},
		{"-0.9", "0"},
		{"-0.1", "0"},
		{"-1.00", "-1"},
		{"-1.01", "-1"},
		{"-1.999", "-1"},
	}
	for _, test := range testsWithStrings {
		expected, _ := NewFromString(test.expected)
		input, _ := NewFromString(test.input)
		assertCeil(input, expected)
	}

	type testDataDecimal struct {
		input    Decimal
		expected string
	}
	testsWithDecimals := []testDataDecimal{
		{New(100, -1), "10"},
		{New(10, 0), "10"},
		{New(1, 1), "10"},
		{New(1999, -3), "2"},
		{New(101, -2), "2"},
		{New(1, 0), "1"},
		{New(0, 0), "0"},
		{New(9, -1), "1"},
		{New(1, -1), "1"},
		{New(-1, -1), "0"},
		{New(-9, -1), "0"},
		{New(-1, 0), "-1"},
		{New(-101, -2), "-1"},
		{New(-1999, -3), "-1"},
	}
	for _, test := range testsWithDecimals {
		expected, _ := NewFromString(test.expected)
		assertCeil(test.input, expected)
	}
}

func TestDecimal_RoundAndStringFixed(t *testing.T) {
	type testData struct {
		input         string
		places        int32
		expected      string
		expectedFixed string
	}
	tests := []testData{
		{"1.454", 0, "1", ""},
		{"1.454", 1, "1.5", ""},
		{"1.454", 2, "1.45", ""},
		{"1.454", 3, "1.454", ""},
		{"1.454", 4, "1.454", "1.4540"},
		{"1.454", 5, "1.454", "1.45400"},
		{"1.554", 0, "2", ""},
		{"1.554", 1, "1.6", ""},
		{"1.554", 2, "1.55", ""},
		{"0.554", 0, "1", ""},
		{"0.454", 0, "0", ""},
		{"0.454", 5, "0.454", "0.45400"},
		{"0", 0, "0", ""},
		{"0", 1, "0", "0.0"},
		{"0", 2, "0", "0.00"},
		{"0", -1, "0", ""},
		{"5", 2, "5", "5.00"},
		{"5", 1, "5", "5.0"},
		{"5", 0, "5", ""},
		{"500", 2, "500", "500.00"},
		{"545", -1, "550", ""},
		{"545", -2, "500", ""},
		{"545", -3, "1000", ""},
		{"545", -4, "0", ""},
		{"499", -3, "0", ""},
		{"499", -4, "0", ""},
	}

	// add negative number tests
	for _, test := range tests {
		expected := test.expected
		if expected != "0" {
			expected = "-" + expected
		}
		expectedStr := test.expectedFixed
		if strings.ContainsAny(expectedStr, "123456789") && expectedStr != "" {
			expectedStr = "-" + expectedStr
		}
		tests = append(tests,
			testData{"-" + test.input, test.places, expected, expectedStr})
	}

	for _, test := range tests {
		d, err := NewFromString(test.input)
		if err != nil {
			panic(err)
		}

		// test Round
		expected, err := NewFromString(test.expected)
		if err != nil {
			panic(err)
		}
		got := d.Round(test.places)
		if !got.Equal(expected) {
			t.Errorf("Rounding %s to %d places, got %s, expected %s",
				d, test.places, got, expected)
		}

		// test StringFixed
		if test.expectedFixed == "" {
			test.expectedFixed = test.expected
		}
		gotStr := d.StringFixed(test.places)
		if gotStr != test.expectedFixed {
			t.Errorf("(%s).StringFixed(%d): got %s, expected %s",
				d, test.places, gotStr, test.expectedFixed)
		}
	}
}

func TestDecimal_BankRoundAndStringFixed(t *testing.T) {
	type testData struct {
		input         string
		places        int32
		expected      string
		expectedFixed string
	}
	tests := []testData{
		{"1.454", 0, "1", ""},
		{"1.454", 1, "1.5", ""},
		{"1.454", 2, "1.45", ""},
		{"1.454", 3, "1.454", ""},
		{"1.454", 4, "1.454", "1.4540"},
		{"1.454", 5, "1.454", "1.45400"},
		{"1.554", 0, "2", ""},
		{"1.554", 1, "1.6", ""},
		{"1.554", 2, "1.55", ""},
		{"0.554", 0, "1", ""},
		{"0.454", 0, "0", ""},
		{"0.454", 5, "0.454", "0.45400"},
		{"0", 0, "0", ""},
		{"0", 1, "0", "0.0"},
		{"0", 2, "0", "0.00"},
		{"0", -1, "0", ""},
		{"5", 2, "5", "5.00"},
		{"5", 1, "5", "5.0"},
		{"5", 0, "5", ""},
		{"500", 2, "500", "500.00"},
		{"545", -2, "500", ""},
		{"545", -3, "1000", ""},
		{"545", -4, "0", ""},
		{"499", -3, "0", ""},
		{"499", -4, "0", ""},
		{"1.45", 1, "1.4", ""},
		{"1.55", 1, "1.6", ""},
		{"1.65", 1, "1.6", ""},
		{"545", -1, "540", ""},
		{"565", -1, "560", ""},
		{"555", -1, "560", ""},
	}

	// add negative number tests
	for _, test := range tests {
		expected := test.expected
		if expected != "0" {
			expected = "-" + expected
		}
		expectedStr := test.expectedFixed
		if strings.ContainsAny(expectedStr, "123456789") && expectedStr != "" {
			expectedStr = "-" + expectedStr
		}
		tests = append(tests,
			testData{"-" + test.input, test.places, expected, expectedStr})
	}

	for _, test := range tests {
		d, err := NewFromString(test.input)
		if err != nil {
			panic(err)
		}

		// test Round
		expected, err := NewFromString(test.expected)
		if err != nil {
			panic(err)
		}
		got := d.RoundBank(test.places)
		if !got.Equal(expected) {
			t.Errorf("Bank Rounding %s to %d places, got %s, expected %s",
				d, test.places, got, expected)
		}

		// test StringFixed
		if test.expectedFixed == "" {
			test.expectedFixed = test.expected
		}
		gotStr := d.StringFixedBank(test.places)
		if gotStr != test.expectedFixed {
			t.Errorf("(%s).StringFixed(%d): got %s, expected %s",
				d, test.places, gotStr, test.expectedFixed)
		}
	}
}

func TestDecimal_Uninitialized(t *testing.T) {
	a := Decimal{}
	b := Decimal{}

	decs := []Decimal{
		a,
		a.rescale(10),
		a.Abs(),
		a.Add(b),
		a.Sub(b),
		a.Mul(b),
		a.Shift(0),
		a.Div(New(1, -1)),
		a.Round(2),
		a.Floor(),
		a.Ceil(),
		a.Truncate(2),
	}

	for _, d := range decs {
		if d.String() != "0" {
			t.Errorf("expected 0, got %s", d.String())
		}
		if d.StringFixed(3) != "0.000" {
			t.Errorf("expected 0, got %s", d.StringFixed(3))
		}
		if d.StringScaled(-2) != "0" {
			t.Errorf("expected 0, got %s", d.StringScaled(-2))
		}
	}

	if a.Cmp(b) != 0 {
		t.Errorf("a != b")
	}
	if a.Sign() != 0 {
		t.Errorf("a.Sign() != 0")
	}
	if a.Exponent() != 0 {
		t.Errorf("a.Exponent() != 0")
	}
	if a.IntPart() != 0 {
		t.Errorf("a.IntPar() != 0")
	}
	f, _ := a.Float64()
	if f != 0 {
		t.Errorf("a.Float64() != 0")
	}
	if a.Rat().RatString() != "0" {
		t.Errorf("a.Rat() != 0, got %s", a.Rat().RatString())
	}
}

func TestDecimal_Add(t *testing.T) {
	type Inp struct {
		a string
		b string
	}

	inputs := map[Inp]string{
		Inp{"2", "3"}:                     "5",
		Inp{"2454495034", "3451204593"}:   "5905699627",
		Inp{"24544.95034", ".3451204593"}: "24545.2954604593",
		Inp{".1", ".1"}:                   "0.2",
		Inp{".1", "-.1"}:                  "0",
		Inp{"0", "1.001"}:                 "1.001",
	}

	for inp, res := range inputs {
		a, err := NewFromString(inp.a)
		if err != nil {
			t.FailNow()
		}
		b, err := NewFromString(inp.b)
		if err != nil {
			t.FailNow()
		}
		c := a.Add(b)
		if c.String() != res {
			t.Errorf("expected %s, got %s", res, c.String())
		}
	}
}

func TestDecimal_Sub(t *testing.T) {
	type Inp struct {
		a string
		b string
	}

	inputs := map[Inp]string{
		Inp{"2", "3"}:                     "-1",
		Inp{"12", "3"}:                    "9",
		Inp{"-2", "9"}:                    "-11",
		Inp{"2454495034", "3451204593"}:   "-996709559",
		Inp{"24544.95034", ".3451204593"}: "24544.6052195407",
		Inp{".1", "-.1"}:                  "0.2",
		Inp{".1", ".1"}:                   "0",
		Inp{"0", "1.001"}:                 "-1.001",
		Inp{"1.001", "0"}:                 "1.001",
		Inp{"2.3", ".3"}:                  "2",
	}

	for inp, res := range inputs {
		a, err := NewFromString(inp.a)
		if err != nil {
			t.FailNow()
		}
		b, err := NewFromString(inp.b)
		if err != nil {
			t.FailNow()
		}
		c := a.Sub(b)
		if c.String() != res {
			t.Errorf("expected %s, got %s", res, c.String())
		}
	}
}

func TestDecimal_Neg(t *testing.T) {
	inputs := map[string]string{
		"0":     "0",
		"10":    "-10",
		"5.56":  "-5.56",
		"-10":   "10",
		"-5.56": "5.56",
	}

	for inp, res := range inputs {
		a, err := NewFromString(inp)
		if err != nil {
			t.FailNow()
		}
		b := a.Neg()
		if b.String() != res {
			t.Errorf("expected %s, got %s", res, b.String())
		}
	}
}

func TestDecimal_NegFromEmpty(t *testing.T) {
	a := Decimal{}
	b := a.Neg()
	if b.String() != "0" {
		t.Errorf("expected %s, got %s", "0", b)
	}
}

func TestDecimal_Mul(t *testing.T) {
	type Inp struct {
		a string
		b string
	}

	inputs := map[Inp]string{
		Inp{"2", "3"}:                     "6",
		Inp{"2454495034", "3451204593"}:   "8470964534836491162",
		Inp{"24544.95034", ".3451204593"}: "8470.964534836491162",
		Inp{".1", ".1"}:                   "0.01",
		Inp{"0", "1.001"}:                 "0",
	}

	for inp, res := range inputs {
		a, err := NewFromString(inp.a)
		if err != nil {
			t.FailNow()
		}
		b, err := NewFromString(inp.b)
		if err != nil {
			t.FailNow()
		}
		c := a.Mul(b)
		if c.String() != res {
			t.Errorf("expected %s, got %s", res, c.String())
		}
	}

	// positive scale
	c := New(1234, 5).Mul(New(45, -1))
	if c.String() != "555300000" {
		t.Errorf("Expected %s, got %s", "555300000", c.String())
	}
}

func TestDecimal_Shift(t *testing.T) {
	type Inp struct {
		a string
		b int32
	}

	inputs := map[Inp]string{
		Inp{"6", 3}:                         "6000",
		Inp{"10", -2}:                       "0.1",
		Inp{"2.2", 1}:                       "22",
		Inp{"-2.2", -1}:                     "-0.22",
		Inp{"12.88", 5}:                     "1288000",
		Inp{"-10234274355545544493", -3}:    "-10234274355545544.493",
		Inp{"-4612301402398.4753343454", 5}: "-461230140239847533.43454",
	}

	for inp, expectedStr := range inputs {
		num, _ := NewFromString(inp.a)

		got := num.Shift(inp.b)
		expected, _ := NewFromString(expectedStr)
		if !got.Equal(expected) {
			t.Errorf("expected %v when shifting %v by %v, got %v",
				expected, num, inp.b, got)
		}
	}
}

func TestDecimal_Div(t *testing.T) {
	type Inp struct {
		a string
		b string
	}

	inputs := map[Inp]string{
		Inp{"6", "3"}:                            "2",
		Inp{"10", "2"}:                           "5",
		Inp{"2.2", "1.1"}:                        "2",
		Inp{"-2.2", "-1.1"}:                      "2",
		Inp{"12.88", "5.6"}:                      "2.3",
		Inp{"1023427554493", "43432632"}:         "23563.5628642767953828", // rounded
		Inp{"1", "434324545566634"}:              "0.0000000000000023",
		Inp{"1", "3"}:                            "0.3333333333333333",
		Inp{"2", "3"}:                            "0.6666666666666667", // rounded
		Inp{"10000", "3"}:                        "3333.3333333333333333",
		Inp{"10234274355545544493", "-3"}:        "-3411424785181848164.3333333333333333",
		Inp{"-4612301402398.4753343454", "23.5"}: "-196268144782.9138440146978723",
	}

	for inp, expectedStr := range inputs {
		num, err := NewFromString(inp.a)
		if err != nil {
			t.FailNow()
		}
		denom, err := NewFromString(inp.b)
		if err != nil {
			t.FailNow()
		}
		got := num.Div(denom)
		expected, _ := NewFromString(expectedStr)
		if !got.Equal(expected) {
			t.Errorf("expected %v when dividing %v by %v, got %v",
				expected, num, denom, got)
		}
		got2 := num.DivRound(denom, int32(DivisionPrecision))
		if !got2.Equal(expected) {
			t.Errorf("expected %v on DivRound (%v,%v), got %v", expected, num, denom, got2)
		}
	}

	type Inp2 struct {
		n    int64
		exp  int32
		n2   int64
		exp2 int32
	}

	// test code path where exp > 0
	inputs2 := map[Inp2]string{
		Inp2{124, 10, 3, 1}: "41333333333.3333333333333333",
		Inp2{124, 10, 3, 0}: "413333333333.3333333333333333",
		Inp2{124, 10, 6, 1}: "20666666666.6666666666666667",
		Inp2{124, 10, 6, 0}: "206666666666.6666666666666667",
		Inp2{10, 10, 10, 1}: "1000000000",
	}

	for inp, expectedAbs := range inputs2 {
		for i := -1; i <= 1; i += 2 {
			for j := -1; j <= 1; j += 2 {
				n := inp.n * int64(i)
				n2 := inp.n2 * int64(j)
				num := New(n, inp.exp)
				denom := New(n2, inp.exp2)
				expected := expectedAbs
				if i != j {
					expected = "-" + expectedAbs
				}
				got := num.Div(denom)
				if got.String() != expected {
					t.Errorf("expected %s when dividing %v by %v, got %v",
						expected, num, denom, got)
				}
			}
		}
	}
}

func TestDecimal_QuoRem(t *testing.T) {
	type Inp4 struct {
		d   string
		d2  string
		exp int32
		q   string
		r   string
	}
	cases := []Inp4{
		{"10", "1", 0, "10", "0"},
		{"1", "10", 0, "0", "1"},
		{"1", "4", 2, "0.25", "0"},
		{"1", "8", 2, "0.12", "0.04"},
		{"10", "3", 1, "3.3", "0.1"},
		{"100", "3", 1, "33.3", "0.1"},
		{"1000", "10", -3, "0", "1000"},
		{"1e-3", "2e-5", 0, "50", "0"},
		{"1e-3", "2e-3", 1, "0.5", "0"},
		{"4e-3", "0.8", 4, "5e-3", "0"},
		{"4.1e-3", "0.8", 3, "5e-3", "1e-4"},
		{"-4", "-3", 0, "1", "-1"},
		{"-4", "3", 0, "-1", "-1"},
	}

	for _, inp4 := range cases {
		d, _ := NewFromString(inp4.d)
		d2, _ := NewFromString(inp4.d2)
		prec := inp4.exp
		q, r := d.QuoRem(d2, prec)
		expectedQ, _ := NewFromString(inp4.q)
		expectedR, _ := NewFromString(inp4.r)
		if !q.Equal(expectedQ) || !r.Equal(expectedR) {
			t.Errorf("bad QuoRem division %s , %s , %d got %v, %v expected %s , %s",
				inp4.d, inp4.d2, prec, q, r, inp4.q, inp4.r)
		}
		if !d.Equal(d2.Mul(q).Add(r)) {
			t.Errorf("not fitting: d=%v, d2= %v, prec=%d, q=%v, r=%v",
				d, d2, prec, q, r)
		}
		if !q.Equal(q.Truncate(prec)) {
			t.Errorf("quotient wrong precision: d=%v, d2= %v, prec=%d, q=%v, r=%v",
				d, d2, prec, q, r)
		}
		if r.Abs().Cmp(d2.Abs().Mul(New(1, -prec))) >= 0 {
			t.Errorf("remainder too large: d=%v, d2= %v, prec=%d, q=%v, r=%v",
				d, d2, prec, q, r)
		}
		if r.value.Sign()*d.value.Sign() < 0 {
			t.Errorf("signum of divisor and rest do not match: d=%v, d2= %v, prec=%d, q=%v, r=%v",
				d, d2, prec, q, r)
		}
	}
}

type DivTestCase struct {
	d    Decimal
	d2   Decimal
	prec int32
}

func createDivTestCases() []DivTestCase {
	res := make([]DivTestCase, 0)
	var n int32 = 5
	a := []int{1, 2, 3, 6, 7, 10, 100, 14, 5, 400, 0, 1000000, 1000000 + 1, 1000000 - 1}
	for s := -1; s < 2; s = s + 2 { // 2
		for s2 := -1; s2 < 2; s2 = s2 + 2 { // 2
			for e1 := -n; e1 <= n; e1++ { // 2n+1
				for e2 := -n; e2 <= n; e2++ { // 2n+1
					var prec int32
					for prec = -n; prec <= n; prec++ { // 2n+1
						for _, v1 := range a { // 11
							for _, v2 := range a { // 11, even if 0 is skipped
								sign1 := New(int64(s), 0)
								sign2 := New(int64(s2), 0)
								d := sign1.Mul(New(int64(v1), int32(e1)))
								d2 := sign2.Mul(New(int64(v2), int32(e2)))
								res = append(res, DivTestCase{d, d2, prec})
							}
						}
					}
				}
			}
		}
	}
	return res
}

func TestDecimal_QuoRem2(t *testing.T) {
	for _, tc := range createDivTestCases() {
		d := tc.d
		if sign(tc.d2) == 0 {
			continue
		}
		d2 := tc.d2
		prec := tc.prec
		q, r := d.QuoRem(d2, prec)
		// rule 1: d = d2*q +r
		if !d.Equal(d2.Mul(q).Add(r)) {
			t.Errorf("not fitting, d=%v, d2=%v, prec=%d, q=%v, r=%v",
				d, d2, prec, q, r)
		}
		// rule 2: q is integral multiple of 10^(-prec)
		if !q.Equal(q.Truncate(prec)) {
			t.Errorf("quotient wrong precision, d=%v, d2=%v, prec=%d, q=%v, r=%v",
				d, d2, prec, q, r)
		}
		// rule 3: abs(r)<abs(d) * 10^(-prec)
		if r.Abs().Cmp(d2.Abs().Mul(New(1, -prec))) >= 0 {
			t.Errorf("remainder too large, d=%v, d2=%v, prec=%d, q=%v, r=%v",
				d, d2, prec, q, r)
		}
		// rule 4: r and d have the same sign
		if r.value.Sign()*d.value.Sign() < 0 {
			t.Errorf("signum of divisor and rest do not match, "+
				"d=%v, d2=%v, prec=%d, q=%v, r=%v",
				d, d2, prec, q, r)
		}
	}
}

// this is the old Div method from decimal
// Div returns d / d2. If it doesn't divide exactly, the result will have
// DivisionPrecision digits after the decimal point.
func (d Decimal) DivOld(d2 Decimal, prec int) Decimal {
	// NOTE(vadim): division is hard, use Rat to do it
	ratNum := d.Rat()
	ratDenom := d2.Rat()

	quoRat := big.NewRat(0, 1).Quo(ratNum, ratDenom)

	// HACK(vadim): converting from Rat to Decimal inefficiently for now
	ret, err := NewFromString(quoRat.FloatString(prec))
	if err != nil {
		panic(err) // this should never happen
	}
	return ret
}

func Benchmark_DivideOriginal(b *testing.B) {
	tcs := createDivTestCases()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, tc := range tcs {
			d := tc.d
			if sign(tc.d2) == 0 {
				continue
			}
			d2 := tc.d2
			prec := tc.prec
			a := d.DivOld(d2, int(prec))
			if sign(a) > 2 {
				panic("dummy panic")
			}
		}
	}
}

func Benchmark_DivideNew(b *testing.B) {
	tcs := createDivTestCases()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, tc := range tcs {
			d := tc.d
			if sign(tc.d2) == 0 {
				continue
			}
			d2 := tc.d2
			prec := tc.prec
			a := d.DivRound(d2, prec)
			if sign(a) > 2 {
				panic("dummy panic")
			}
		}
	}
}

func sign(d Decimal) int {
	return d.value.Sign()
}

// rules for rounded divide, rounded to integer
// rounded_divide(d,d2) = q
// sign q * sign (d/d2) >= 0
// for d and d2 >0 :
// q is already rounded
// q = d/d2 + r , with r > -0.5 and r <= 0.5
// thus q-d/d2 = r, with r > -0.5 and r <= 0.5
// and d2 q -d = r d2 with r d2 > -d2/2 and r d2 <= d2/2
// and 2 (d2 q -d) = x with x > -d2 and x <= d2
// if we factor in precision then x > -d2 * 10^(-precision) and x <= d2 * 10(-precision)

func TestDecimal_DivRound(t *testing.T) {
	cases := []struct {
		d      string
		d2     string
		prec   int32
		result string
	}{
		{"2", "2", 0, "1"},
		{"1", "2", 0, "1"},
		{"-1", "2", 0, "-1"},
		{"-1", "-2", 0, "1"},
		{"1", "-2", 0, "-1"},
		{"1", "-20", 1, "-0.1"},
		{"1", "-20", 2, "-0.05"},
		{"1", "20.0000000000000000001", 1, "0"},
		{"1", "19.9999999999999999999", 1, "0.1"},
	}
	for _, s := range cases {
		d, _ := NewFromString(s.d)
		d2, _ := NewFromString(s.d2)
		result, _ := NewFromString(s.result)
		prec := s.prec
		q := d.DivRound(d2, prec)
		if sign(q)*sign(d)*sign(d2) < 0 {
			t.Errorf("sign of quotient wrong, got: %v/%v is about %v", d, d2, q)
		}
		x := q.Mul(d2).Abs().Sub(d.Abs()).Mul(New(2, 0))
		if x.Cmp(d2.Abs().Mul(New(1, -prec))) > 0 {
			t.Errorf("wrong rounding, got: %v/%v prec=%d is about %v", d, d2, prec, q)
		}
		if x.Cmp(d2.Abs().Mul(New(-1, -prec))) <= 0 {
			t.Errorf("wrong rounding, got: %v/%v prec=%d is about %v", d, d2, prec, q)
		}
		if !q.Equal(result) {
			t.Errorf("rounded division wrong %s / %s scale %d = %s, got %v", s.d, s.d2, prec, s.result, q)
		}
	}
}

func TestDecimal_DivRound2(t *testing.T) {
	for _, tc := range createDivTestCases() {
		d := tc.d
		if sign(tc.d2) == 0 {
			continue
		}
		d2 := tc.d2
		prec := tc.prec
		q := d.DivRound(d2, prec)
		if sign(q)*sign(d)*sign(d2) < 0 {
			t.Errorf("sign of quotient wrong, got: %v/%v is about %v", d, d2, q)
		}
		x := q.Mul(d2).Abs().Sub(d.Abs()).Mul(New(2, 0))
		if x.Cmp(d2.Abs().Mul(New(1, -prec))) > 0 {
			t.Errorf("wrong rounding, got: %v/%v prec=%d is about %v", d, d2, prec, q)
		}
		if x.Cmp(d2.Abs().Mul(New(-1, -prec))) <= 0 {
			t.Errorf("wrong rounding, got: %v/%v prec=%d is about %v", d, d2, prec, q)
		}
	}
}

func TestDecimal_RoundCash(t *testing.T) {
	tests := []struct {
		d        string
		interval uint8
		result   string
	}{
		{"3.44", 5, "3.45"},
		{"3.43", 5, "3.45"},
		{"3.42", 5, "3.40"},
		{"3.425", 5, "3.45"},
		{"3.47", 5, "3.45"},
		{"3.478", 5, "3.50"},
		{"3.48", 5, "3.50"},
		{"348", 5, "348"},

		{"3.23", 10, "3.20"},
		{"3.33", 10, "3.30"},
		{"3.53", 10, "3.50"},
		{"3.949", 10, "3.90"},
		{"3.95", 10, "4.00"},
		{"395", 10, "395"},

		{"6.42", 15, "6.40"},
		{"6.39", 15, "6.40"},
		{"6.35", 15, "6.30"},
		{"6.36", 15, "6.40"},
		{"6.349", 15, "6.30"},
		{"6.30", 15, "6.30"},
		{"666", 15, "666"},

		{"3.23", 25, "3.25"},
		{"3.33", 25, "3.25"},
		{"3.53", 25, "3.50"},
		{"3.93", 25, "4.00"},
		{"3.41", 25, "3.50"},

		{"3.249", 50, "3.00"},
		{"3.33", 50, "3.50"},
		{"3.749999999", 50, "3.50"},
		{"3.75", 50, "4.00"},
		{"3.93", 50, "4.00"},
		{"393", 50, "393"},

		{"3.249", 100, "3.00"},
		{"3.49999", 100, "3.00"},
		{"3.50", 100, "4.00"},
		{"3.75", 100, "4.00"},
		{"3.93", 100, "4.00"},
		{"393", 100, "393"},
	}
	for i, test := range tests {
		d, _ := NewFromString(test.d)
		haveRounded := d.RoundCash(test.interval)
		result, _ := NewFromString(test.result)

		if !haveRounded.Equal(result) {
			t.Errorf("Index %d: Cash rounding for %q interval %d want %q, have %q", i, test.d, test.interval, test.result, haveRounded)
		}
	}
}

func TestDecimal_RoundCash_Panic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			if have, ok := r.(string); ok {
				const want = "Decimal does not support this Cash rounding interval `231`. Supported: 5, 10, 15, 25, 50, 100"
				if want != have {
					t.Errorf("\nWant: %q\nHave: %q", want, have)
				}
			} else {
				t.Errorf("Panic should contain an error string but got:\n%+v", r)
			}
		} else {
			t.Error("Expecting a panic but got nothing")
		}
	}()
	d, _ := NewFromString("1")
	d.RoundCash(231)
}

func BenchmarkDecimal_RoundCash_Five(b *testing.B) {
	const want = "3.50"
	for i := 0; i < b.N; i++ {
		val := New(3478, -3)
		if have := val.StringFixedCash(5); have != want {
			b.Fatalf("\nHave: %q\nWant: %q", have, want)
		}
	}
}

func BenchmarkDecimal_RoundCash_Fifteen(b *testing.B) {
	const want = "6.30"
	for i := 0; i < b.N; i++ {
		val := New(635, -2)
		if have := val.StringFixedCash(15); have != want {
			b.Fatalf("\nHave: %q\nWant: %q", have, want)
		}
	}
}

func TestDecimal_Mod(t *testing.T) {
	type Inp struct {
		a string
		b string
	}

	inputs := map[Inp]string{
		Inp{"3", "2"}:                     "1",
		Inp{"3451204593", "2454495034"}:   "996709559",
		Inp{"24544.95034", ".3451204593"}: "0.3283950433",
		Inp{".1", ".1"}:                   "0",
		Inp{"0", "1.001"}:                 "0",
		Inp{"-7.5", "2"}:                  "-1.5",
		Inp{"7.5", "-2"}:                  "1.5",
		Inp{"-7.5", "-2"}:                 "-1.5",
	}

	for inp, res := range inputs {
		a, err := NewFromString(inp.a)
		if err != nil {
			t.FailNow()
		}
		b, err := NewFromString(inp.b)
		if err != nil {
			t.FailNow()
		}
		c := a.Mod(b)
		if c.String() != res {
			t.Errorf("expected %s, got %s", res, c.String())
		}
	}
}

func TestDecimal_Overflow(t *testing.T) {
	if !didPanic(func() { New(1, math.MinInt32).Mul(New(1, math.MinInt32)) }) {
		t.Fatalf("should have gotten an overflow panic")
	}
	if !didPanic(func() { New(1, math.MaxInt32).Mul(New(1, math.MaxInt32)) }) {
		t.Fatalf("should have gotten an overflow panic")
	}
}

func TestDecimal_ExtremeValues(t *testing.T) {
	// NOTE(vadim): this test takes pretty much forever
	if testing.Short() {
		t.Skip()
	}

	// NOTE(vadim): Seriously, the numbers involved are so large that this
	// test will take way too long, so mark it as success if it takes over
	// 1 second. The way this test typically fails (integer overflow) is that
	// a wrong result appears quickly, so if it takes a long time then it is
	// probably working properly.
	// Why even bother testing this? Completeness, I guess. -Vadim
	const timeLimit = 1 * time.Second
	test := func(f func()) {
		c := make(chan bool)
		go func() {
			f()
			close(c)
		}()
		select {
		case <-c:
		case <-time.After(timeLimit):
		}
	}

	test(func() {
		got := New(123, math.MinInt32).Floor()
		if !got.Equal(NewFromFloat(0)) {
			t.Errorf("Error: got %s, expected 0", got)
		}
	})
	test(func() {
		got := New(123, math.MinInt32).Ceil()
		if !got.Equal(NewFromFloat(1)) {
			t.Errorf("Error: got %s, expected 1", got)
		}
	})
	test(func() {
		got := New(123, math.MinInt32).Rat().FloatString(10)
		expected := "0.0000000000"
		if got != expected {
			t.Errorf("Error: got %s, expected %s", got, expected)
		}
	})
}

func TestIntPart(t *testing.T) {
	for _, testCase := range []struct {
		Dec     string
		IntPart int64
	}{
		{"0.01", 0},
		{"12.1", 12},
		{"9999.999", 9999},
		{"-32768.01234", -32768},
	} {
		d, err := NewFromString(testCase.Dec)
		if err != nil {
			t.Fatal(err)
		}
		if d.IntPart() != testCase.IntPart {
			t.Errorf("expect %d, got %d", testCase.IntPart, d.IntPart())
		}
	}
}

func TestDecimal_Min(t *testing.T) {
	// the first element in the array is the expected answer, rest are inputs
	testCases := [][]float64{
		{0, 0},
		{1, 1},
		{-1, -1},
		{1, 1, 2},
		{-2, 1, 2, -2},
		{-3, 0, 2, -2, -3},
	}

	for _, test := range testCases {
		expected, input := test[0], test[1:]
		expectedDecimal := NewFromFloat(expected)
		decimalInput := []Decimal{}
		for _, inp := range input {
			d := NewFromFloat(inp)
			decimalInput = append(decimalInput, d)
		}
		got := Min(decimalInput[0], decimalInput[1:]...)
		if !got.Equal(expectedDecimal) {
			t.Errorf("Expected %v, got %v, input=%+v", expectedDecimal, got,
				decimalInput)
		}
	}
}

func TestDecimal_Max(t *testing.T) {
	// the first element in the array is the expected answer, rest are inputs
	testCases := [][]float64{
		{0, 0},
		{1, 1},
		{-1, -1},
		{2, 1, 2},
		{2, 1, 2, -2},
		{3, 0, 3, -2},
		{-2, -3, -2},
	}

	for _, test := range testCases {
		expected, input := test[0], test[1:]
		expectedDecimal := NewFromFloat(expected)
		decimalInput := []Decimal{}
		for _, inp := range input {
			d := NewFromFloat(inp)
			decimalInput = append(decimalInput, d)
		}
		got := Max(decimalInput[0], decimalInput[1:]...)
		if !got.Equal(expectedDecimal) {
			t.Errorf("Expected %v, got %v, input=%+v", expectedDecimal, got,
				decimalInput)
		}
	}
}

func TestDecimal_Scan(t *testing.T) {
	// test the Scan method that implements the
	// sql.Scanner interface
	// check for the for different type of values
	// that are possible to be received from the database
	// drivers

	// in normal operations the db driver (sqlite at least)
	// will return an int64 if you specified a numeric format
	a := Decimal{}
	dbvalue := float64(54.33)
	expected := NewFromFloat(dbvalue)

	err := a.Scan(dbvalue)
	if err != nil {
		// Scan failed... no need to test result value
		t.Errorf("a.Scan(54.33) failed with message: %s", err)

	} else {
		// Scan succeeded... test resulting values
		if !a.Equal(expected) {
			t.Errorf("%s does not equal to %s", a, expected)
		}
	}

	// apparently MySQL 5.7.16 and returns these as float32 so we need
	// to handle these as well
	dbvalueFloat32 := float32(54.33)
	expected = NewFromFloat(float64(dbvalueFloat32))

	err = a.Scan(dbvalueFloat32)
	if err != nil {
		// Scan failed... no need to test result value
		t.Errorf("a.Scan(54.33) failed with message: %s", err)

	} else {
		// Scan succeeded... test resulting values
		if !a.Equal(expected) {
			t.Errorf("%s does not equal to %s", a, expected)
		}
	}

	// at least SQLite returns an int64 when 0 is stored in the db
	// and you specified a numeric format on the schema
	dbvalueInt := int64(0)
	expected = New(dbvalueInt, 0)

	err = a.Scan(dbvalueInt)
	if err != nil {
		// Scan failed... no need to test result value
		t.Errorf("a.Scan(0) failed with message: %s", err)

	} else {
		// Scan succeeded... test resulting values
		if !a.Equal(expected) {
			t.Errorf("%s does not equal to %s", a, expected)
		}
	}

	// in case you specified a varchar in your SQL schema,
	// the database driver will return byte slice []byte
	valueStr := "535.666"
	dbvalueStr := []byte(valueStr)
	expected, err = NewFromString(valueStr)
	if err != nil {
		t.Fatal(err)
	}

	err = a.Scan(dbvalueStr)
	if err != nil {
		// Scan failed... no need to test result value
		t.Errorf("a.Scan('535.666') failed with message: %s", err)

	} else {
		// Scan succeeded... test resulting values
		if !a.Equal(expected) {
			t.Errorf("%s does not equal to %s", a, expected)
		}
	}

	// lib/pq can also return strings
	expected, err = NewFromString(valueStr)
	if err != nil {
		t.Fatal(err)
	}

	err = a.Scan(valueStr)
	if err != nil {
		// Scan failed... no need to test result value
		t.Errorf("a.Scan('535.666') failed with message: %s", err)
	} else {
		// Scan succeeded... test resulting values
		if !a.Equal(expected) {
			t.Errorf("%s does not equal to %s", a, expected)
		}
	}

	type foo struct{}
	err = a.Scan(foo{})
	if err == nil {
		t.Errorf("a.Scan(Foo{}) should have thrown an error but did not")
	}
}

func TestDecimal_Value(t *testing.T) {
	// Make sure this does implement the database/sql's driver.Valuer interface
	var d Decimal
	if _, ok := interface{}(d).(driver.Valuer); !ok {
		t.Error("Decimal does not implement driver.Valuer")
	}

	// check that normal case is handled appropriately
	a := New(1234, -2)
	expected := "12.34"
	value, err := a.Value()
	if err != nil {
		t.Errorf("Decimal(12.34).Value() failed with message: %s", err)
	} else if value.(string) != expected {
		t.Errorf("%s does not equal to %s", a, expected)
	}
}

// old tests after this line

func TestDecimal_Scale(t *testing.T) {
	a := New(1234, -3)
	if a.Exponent() != -3 {
		t.Errorf("error")
	}
}

func TestDecimal_Abs1(t *testing.T) {
	a := New(-1234, -4)
	b := New(1234, -4)

	c := a.Abs()
	if c.Cmp(b) != 0 {
		t.Errorf("error")
	}
}

func TestDecimal_Abs2(t *testing.T) {
	a := New(-1234, -4)
	b := New(1234, -4)

	c := b.Abs()
	if c.Cmp(a) == 0 {
		t.Errorf("error")
	}
}

func TestDecimal_Equalities(t *testing.T) {
	a := New(1234, 3)
	b := New(1234, 3)
	c := New(1234, 4)

	if !a.Equal(b) {
		t.Errorf("%q should equal %q", a, b)
	}
	if a.Equal(c) {
		t.Errorf("%q should not equal %q", a, c)
	}

	// note, this block should be deprecated, here for backwards compatibility
	if !a.Equals(b) {
		t.Errorf("%q should equal %q", a, b)
	}

	if !c.GreaterThan(b) {
		t.Errorf("%q should be greater than  %q", c, b)
	}
	if b.GreaterThan(c) {
		t.Errorf("%q should not be greater than  %q", b, c)
	}
	if !a.GreaterThanOrEqual(b) {
		t.Errorf("%q should be greater or equal %q", a, b)
	}
	if !c.GreaterThanOrEqual(b) {
		t.Errorf("%q should be greater or equal %q", c, b)
	}
	if b.GreaterThanOrEqual(c) {
		t.Errorf("%q should not be greater or equal %q", b, c)
	}
	if !b.LessThan(c) {
		t.Errorf("%q should be less than %q", a, b)
	}
	if c.LessThan(b) {
		t.Errorf("%q should not be less than  %q", a, b)
	}
	if !a.LessThanOrEqual(b) {
		t.Errorf("%q should be less than or equal %q", a, b)
	}
	if !b.LessThanOrEqual(c) {
		t.Errorf("%q should be less than or equal  %q", a, b)
	}
	if c.LessThanOrEqual(b) {
		t.Errorf("%q should not be less than or equal %q", a, b)
	}
}

func TestDecimal_ScalesNotEqual(t *testing.T) {
	a := New(1234, 2)
	b := New(1234, 3)
	if a.Equal(b) {
		t.Errorf("%q should not equal %q", a, b)
	}
}

func TestDecimal_Cmp1(t *testing.T) {
	a := New(123, 3)
	b := New(-1234, 2)

	if a.Cmp(b) != 1 {
		t.Errorf("Error")
	}
}

func TestDecimal_Cmp2(t *testing.T) {
	a := New(123, 3)
	b := New(1234, 2)

	if a.Cmp(b) != -1 {
		t.Errorf("Error")
	}
}

func TestPow(t *testing.T) {
	a := New(4, 0)
	b := New(2, 0)
	x := a.Pow(b)
	if x.String() != "16" {
		t.Errorf("Error, saw %s", x.String())
	}
}

func TestNegativePow(t *testing.T) {
	a := New(4, 0)
	b := New(-2, 0)
	x := a.Pow(b)
	if x.String() != "0.0625" {
		t.Errorf("Error, saw %s", x.String())
	}
}

func TestDecimal_Sign(t *testing.T) {
	if Zero.Sign() != 0 {
		t.Errorf("%q should have sign 0", Zero)
	}

	one := New(1, 0)
	if one.Sign() != 1 {
		t.Errorf("%q should have sign 1", one)
	}

	mone := New(-1, 0)
	if mone.Sign() != -1 {
		t.Errorf("%q should have sign -1", mone)
	}
}

func didPanic(f func()) bool {
	ret := false
	func() {

		defer func() {
			if message := recover(); message != nil {
				ret = true
			}
		}()

		// call the target function
		f()

	}()

	return ret

}

func TestDecimal_Coefficient(t *testing.T) {
	d := New(123, 0)
	co := d.Coefficient()
	if co.Int64() != 123 {
		t.Error("Coefficient should be 123; Got:", co)
	}
	co.Set(big.NewInt(0))
	if d.IntPart() != 123 {
		t.Error("Modifying coefficient modified Decimal; Got:", d)
	}
}

type DecimalSlice []Decimal

func (p DecimalSlice) Len() int           { return len(p) }
func (p DecimalSlice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p DecimalSlice) Less(i, j int) bool { return p[i].Cmp(p[j]) < 0 }
func Benchmark_Cmp(b *testing.B) {
	decimals := DecimalSlice([]Decimal{})
	for i := 0; i < 1000000; i++ {
		decimals = append(decimals, New(int64(i), 0))
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sort.Sort(decimals)
	}
}

func TestNullDecimal_Scan(t *testing.T) {
	// test the Scan method that implements the
	// sql.Scanner interface
	// check for the for different type of values
	// that are possible to be received from the database
	// drivers

	// in normal operations the db driver (sqlite at least)
	// will return an int64 if you specified a numeric format

	// Make sure handles nil values
	a := NullDecimal{}
	var dbvaluePtr interface{}
	err := a.Scan(dbvaluePtr)
	if err != nil {
		// Scan failed... no need to test result value
		t.Errorf("a.Scan(nil) failed with message: %s", err)
	} else {
		if a.Valid {
			t.Errorf("%s is not null", a.Decimal)
		}
	}

	dbvalue := float64(54.33)
	expected := NewFromFloat(dbvalue)

	err = a.Scan(dbvalue)
	if err != nil {
		// Scan failed... no need to test result value
		t.Errorf("a.Scan(54.33) failed with message: %s", err)

	} else {
		// Scan succeeded... test resulting values
		if !a.Valid {
			t.Errorf("%s is null", a.Decimal)
		} else if !a.Decimal.Equals(expected) {
			t.Errorf("%s does not equal to %s", a.Decimal, expected)
		}
	}

	// at least SQLite returns an int64 when 0 is stored in the db
	// and you specified a numeric format on the schema
	dbvalueInt := int64(0)
	expected = New(dbvalueInt, 0)

	err = a.Scan(dbvalueInt)
	if err != nil {
		// Scan failed... no need to test result value
		t.Errorf("a.Scan(0) failed with message: %s", err)

	} else {
		// Scan succeeded... test resulting values
		if !a.Valid {
			t.Errorf("%s is null", a.Decimal)
		} else if !a.Decimal.Equals(expected) {
			t.Errorf("%v does not equal %v", a, expected)
		}
	}

	// in case you specified a varchar in your SQL schema,
	// the database driver will return byte slice []byte
	valueStr := "535.666"
	dbvalueStr := []byte(valueStr)
	expected, err = NewFromString(valueStr)
	if err != nil {
		t.Fatal(err)
	}

	err = a.Scan(dbvalueStr)
	if err != nil {
		// Scan failed... no need to test result value
		t.Errorf("a.Scan('535.666') failed with message: %s", err)

	} else {
		// Scan succeeded... test resulting values
		if !a.Valid {
			t.Errorf("%s is null", a.Decimal)
		} else if !a.Decimal.Equals(expected) {
			t.Errorf("%v does not equal %v", a, expected)
		}
	}

	// lib/pq can also return strings
	expected, err = NewFromString(valueStr)
	if err != nil {
		t.Fatal(err)
	}

	err = a.Scan(valueStr)
	if err != nil {
		// Scan failed... no need to test result value
		t.Errorf("a.Scan('535.666') failed with message: %s", err)
	} else {
		// Scan succeeded... test resulting values
		if !a.Valid {
			t.Errorf("%s is null", a.Decimal)
		} else if !a.Decimal.Equals(expected) {
			t.Errorf("%v does not equal %v", a, expected)
		}
	}
}

func TestNullDecimal_Value(t *testing.T) {
	// Make sure this does implement the database/sql's driver.Valuer interface
	var nullDecimal NullDecimal
	if _, ok := interface{}(nullDecimal).(driver.Valuer); !ok {
		t.Error("NullDecimal does not implement driver.Valuer")
	}

	// check that null is handled appropriately
	value, err := nullDecimal.Value()
	if err != nil {
		t.Errorf("NullDecimal{}.Valid() failed with message: %s", err)
	} else if value != nil {
		t.Errorf("%v is not nil", value)
	}

	// check that normal case is handled appropriately
	a := NullDecimal{Decimal: New(1234, -2), Valid: true}
	expected := "12.34"
	value, err = a.Value()
	if err != nil {
		t.Errorf("NullDecimal(12.34).Value() failed with message: %s", err)
	} else if value.(string) != expected {
		t.Errorf("%v does not equal %v", a, expected)
	}
}

func TestBinary(t *testing.T) {
	for _, y := range testTable {
		x := y.float

		// Create the decimal
		d1 := NewFromFloat(x)

		// Encode to binary
		b, err := d1.MarshalBinary()
		if err != nil {
			t.Errorf("error marshalling %v to binary: %v", d1, err)
		}

		// Restore from binary
		var d2 Decimal
		err = (&d2).UnmarshalBinary(b)
		if err != nil {
			t.Errorf("error unmarshalling from binary: %v", err)
		}

		// The restored decimal should equal the original
		if !d1.Equals(d2) {
			t.Errorf("expected %v when restoring, got %v", d1, d2)
		}
	}
}

func slicesEqual(a, b []byte) bool {
	for i, val := range a {
		if b[i] != val {
			return false
		}
	}
	return true
}

func TestGobEncode(t *testing.T) {
	for _, y := range testTable {
		x := y.float
		d1 := NewFromFloat(x)

		b1, err := d1.GobEncode()
		if err != nil {
			t.Errorf("error encoding %v to binary: %v", d1, err)
		}

		d2 := NewFromFloat(x)

		b2, err := d2.GobEncode()
		if err != nil {
			t.Errorf("error encoding %v to binary: %v", d2, err)
		}

		if !slicesEqual(b1, b2) {
			t.Errorf("something about the gobencode is not working properly \n%v\n%v", b1, b2)
		}

		var d3 Decimal
		err = d3.GobDecode(b1)
		if err != nil {
			t.Errorf("Error gobdecoding %v, got %v", b1, d3)
		}
		var d4 Decimal
		err = d4.GobDecode(b2)
		if err != nil {
			t.Errorf("Error gobdecoding %v, got %v", b2, d4)
		}

		eq := d3.Equal(d4)
		if eq != true {
			t.Errorf("Encoding then decoding mutated Decimal")
		}

		eq = d1.Equal(d3)
		if eq != true {
			t.Errorf("Error gobencoding/decoding %v, got %v", d1, d3)
		}
	}
}

func TestSum(t *testing.T) {
	vals := make([]Decimal, 10)
	var i = int64(0)

	for key := range vals {
		vals[key] = New(i, 0)
		i++
	}

	sum := Sum(vals[0], vals[1:]...)
	if !sum.Equal(New(45, 0)) {
		t.Errorf("Failed to calculate sum, expected %s got %s", New(45, 0), sum)
	}
}

func TestAvg(t *testing.T) {
	vals := make([]Decimal, 10)
	var i = int64(0)

	for key := range vals {
		vals[key] = New(i, 0)
		i++
	}

	avg := Avg(vals[0], vals[1:]...)
	if !avg.Equal(NewFromFloat(4.5)) {
		t.Errorf("Failed to calculate average, expected %s got %s", NewFromFloat(4.5).String(), avg.String())
	}
}

func TestRoundBankAnomaly(t *testing.T) {
	a := New(25, -1)
	b := New(250, -2)

	if !a.Equal(b) {
		t.Errorf("Expected %s to equal %s", a, b)
	}

	expected := New(2, 0)

	aRounded := a.RoundBank(0)
	if !aRounded.Equal(expected) {
		t.Errorf("Expected bank rounding %s to equal %s, but it was %s", a, expected, aRounded)
	}

	bRounded := b.RoundBank(0)
	if !bRounded.Equal(expected) {
		t.Errorf("Expected bank rounding %s to equal %s, but it was %s", b, expected, bRounded)
	}
}

// Trig tests

// For Atan
func TestAtan(t *testing.T) {
	inps := []string{
		"-2.91919191919191919",
		"-1.0",
		"-0.25",
		"0.0",
		"0.33",
		"1.0",
		"5.0",
		"10",
	  "11000020.2407442310156021090304691671842603586882014729198302312846062338790031898128063403419218957424",
	}
	sols := []string{
		"-1.2407643882205801102743706275310603586882014729198302312846062338790031898128063403419218957424163818359375",
		"-0.78539816339744833061616997868383017934410073645991511564230311693950159490640317017096094787120819091796875",
		"-0.24497866312686415",
		"0.0",
		"0.318747560420644443",
		"0.78539816339744833061616997868383017934410073645991511564230311693950159490640317017096094787120819091796875",
		"1.3734007669450159012323399573676603586882014729198302312846062338790031898128063403419218957424163818359375",
		"1.4711276743037346312323399573676603586882014729198302312846062338790031898128063403419218957424163818359375",
		"1.5707962358859730612325945023537403586882014729198302312846062338790031898128063403419218957424163818359375",
	}
	for i, inp := range inps {
		d, err := NewFromString(inp)
		if err != nil {
			t.FailNow()
		}
		s, err := NewFromString(sols[i])
		if err != nil {
			t.FailNow()
		}
		a := d.Atan()
		if !a.Equal(s) {
			t.Errorf("expected %s, got %s", s, a)
		}
	}
}

// For Sin
func TestSin(t *testing.T) {
	inps := []string{
		"-2.91919191919191919",
		"-1.0",
		"-0.25",
		"0.0",
		"0.33",
		"1.0",
		"5.0",
		"10",
	  "11000020.2407442310156021090304691671842603586882014729198302312846062338790031898128063403419218957424",
	}
	sols := []string{
		"-0.220571862520030093613839729898389234591448264443828979962999436149729548845614563657783710150808249493659603883785557750857064742068654401611592366453156979425089038732558977915147138610177101332138819745624841999009731533470874398791630404814623756725761484209848526893621903868344844362592350061614721709283168025028307789990557501123216276655612701350632840149744527732872833304632776220881820168763715880072271900752263175824768749676997835634284780882959485313122780499486873741058283218325040301421620652217699025347852371634749059985101913743575566482455108757945784327770283152556778940395109157571785095552777237471737930299497127855782791647416409676751196562400230573256919649166116220548146722732673811940013822174052471575484241086300786209457059067702404625074214637302180559257156598478224357461349695824437855046828206155188482292930722180491096361781653423287753045189099820938975700063357767823639345982521057442719437079960883170296730778877529726323641962581167118741985973085143794261349309399843621404104514700512532703761239968546841161640494797236796414402075947283426422877298624337798757146020442229118476349906575010431352541089067121029579490866064323880357132858662072455977109093817967157613567800944577216024589669093782841129397681562208313003621389103425798339230823330581188201904296875",
		"-0.84147098480789650683573384227719871264659725213827912307678945963233888758666252541652307542471941569920950873234694687123276222626964952320957348347806244448304658355378064371162688704108432893734339281119107889886243608606771440333499496399246182550579552307436056577989755048676139388257574972355638331020916986327112812449023235994235220681962126402627149122741398037411569225585167209440887381294958909446263040889436515610243685011381400045762457675558838333095494482138227725013381292722259213324814381642325470123171162316406069142833674558352099403808484142312206036567122374357983536120130367436871049536150555321227634134943262101222587962853023526365048081446005272233936756658798560188088496199440605960018665932771208203273875932636591008594752836095901878899636892331191188133444513639520743904232369072391326448163751628578886099960312207261234582748222511007033947148867295201489673608530271710726980420265749989442268185776660104680951516862727780322815011940555693819786827407000470928166813302681223319136916385379500478197064927760223020342764479386095349145229773585152310192696649682836151684354735282894320955021997802761069713870309184654536840910246184967655772931790056052366745119986847633246318705770507591911530119455091063686301656648189606417807689546576464738456252196397075290117772324937902655849801056962584258183367906235526162380299193832940968379238027598232552151102936477400362491607666015625",
		"-0.247403959254522929719728461264852108650160857317149198152655832001777826134603088803487480618059635162353515625",
		"0.0",
		"0.324043028394868345912501559420824735488967340443112210114701957327566252120565337691004970110952854156494140625",
		"0.84147098480789650683573384227719871264659725213827912307678945963233888758666252541652307542471941569920950873234694687123276222626964952320957348347806244448304658355378064371162688704108432893734339281119107889886243608606771440333499496399246182550579552307436056577989755048676139388257574972355638331020916986327112812449023235994235220681962126402627149122741398037411569225585167209440887381294958909446263040889436515610243685011381400045762457675558838333095494482138227725013381292722259213324814381642325470123171162316406069142833674558352099403808484142312206036567122374357983536120130367436871049536150555321227634134943262101222587962853023526365048081446005272233936756658798560188088496199440605960018665932771208203273875932636591008594752836095901878899636892331191188133444513639520743904232369072391326448163751628578886099960312207261234582748222511007033947148867295201489673608530271710726980420265749989442268185776660104680951516862727780322815011940555693819786827407000470928166813302681223319136916385379500478197064927760223020342764479386095349145229773585152310192696649682836151684354735282894320955021997802761069713870309184654536840910246184967655772931790056052366745119986847633246318705770507591911530119455091063686301656648189606417807689546576464738456252196397075290117772324937902655849801056962584258183367906235526162380299193832940968379238027598232552151102936477400362491607666015625",
		"-0.95892427466313846886683534084169138601699178595056142681129038893400410281163045903621198874295374290149806823897075576057724999482959874362888984630786877750800707948164478439888413437709227541296453939066231042449153560817956529124771991719434464325257307490209688817348630841990228088502580219431623697760216390920129365244393334429239217501291865772417470981461975427863913628459487634361457463896191397206119089801538318057594728767278995563353925328463456560840888895608182514323808823801838724482524462741712473496938668730262585034295941891365544300178534659369917661430555201335114597830077556398746790054730682231578093617522006106659977088453875512527401574137870617360741477467052899927465250688663724908651769559466919777914497651416655110015287957761809374118954830175592272643086681859204507636705728518901196924686915642366269243409467335313632803457424138328976320449572075303129501941586734619627126950347146441488301272523732890971103193432131967093584975094064352688636321088866275878367912478080039614015674090467373532503838989934725811696669260106530065305130564789502177378344025691707116401561519233199329283661803416619394138904068356512359009119169667571681552281997620458915816124425995894145459910451225779660911316891499301549117918834162954131295015995478902576153796243561887919093281708146016725119431774666779045011119428333353937254827947830587875729623490751729608660980375134386122226715087890625",
		"-0.544021110889369814404798546234261521735237410341301440744913863554325413910729487222404754912397132389280869582562649416317020824255899391102716253549602359933406667420857908751554064982667505778980864397665237869392769251295144996601379178780771501957858736340861984904014191524939793604688756271913886434778180316203812548425041682948461830667971245994025528380063656691687013885731759172341309655658172703937945366976128597657101230629806658897303549468649147494313422814604383947907342285887634159176609268080528703630505118954169027929142559898709956536517735787676487117591429732597764129997800544321069045676840705169697409709450991200259966568075515292000715246707423119024660809599616729869855242609248565503156233847151639381625642548896051440565640966333400046266460293523862957282305141303657282623793830073310394349418290939283661501995774676808852008306303431767521000117474347792119119745767325061425517820206720145795639782171641765977980488850129569363740173287329545019826050789666603219553652178298235868990551128361160456166647394642535572127727629420370720856130359966967599668830582910424823037763854828905673063487188976311335956440851988180468345725600705105113853947651574770698730911647817190648000646872041955859376392859382055300138695355787127336060458870914402496055117808282375335693359375",
		"-0.564291758073920672156060501108780948637502927903402695682970758082173347334144085493099932047342878232107294208163423635351694839181128599190761981992082190183414367145761958978263205360050276797203465120166463337285861568628620441149869080262018002460551105733370384496691552603807079787064355499979892738941692390909198368605275928720995456124329350110140031040327865030504338978174244479702663590952255309812151905391076362043165507118546398673253957125454990703254702293833362635280840010399956637688760201442320428708562854314419339362472820724298986435162931641340060501848890585327411159080467227885852591584159465243368331681937734746160681439404005233844642476786199607795910618003577395193096223575912506207477927572401934790710176468073033112942732454695154492720511230782735184156027155137618173659003878927697237751139877779082916148874086491265428313018347132700490228131397266369640661232003084568637049642401980172435548737268283623715380443755530399410106957744899992161330205161097513785489406292271887390739430612193060474304401657511461250218562733033893187306639264886516827894130028111468708805823753758224864450015655567869605537758979198887215555526950741196023433812962018677695204317878927956155283357420412564940844662156159296973147878187706628631717213628871121500196750275790691375732421875",
	}
	for i, inp := range inps {
		d, err := NewFromString(inp)
		if err != nil {
			t.FailNow()
		}
		s, err := NewFromString(sols[i])
		if err != nil {
			t.FailNow()
		}
		a := d.Sin()
		if !a.Equal(s) {
			t.Errorf("expected %s, got %s", s, a)
		}
	}
}

// For Cos
func TestCos(t *testing.T) {
	inps := []string{
		"-2.91919191919191919",
		"-1.0",
		"-0.25",
		"0.0",
		"0.33",
		"1.0",
		"5.0",
		"10",
	  "11000020.2407442310156021090304691671842603586882014729198302312846062338790031898128063403419218957424",
	}
	sols := []string{
		"-0.975370726167463436728505866323714237868918798735337852490354630275400590397256757982930518671177656988788392411709345759154371178337418085499048191038650091945778580197329516971729513331012900776507504230803826799226710606177535289143559059224362661001603838745169803182674368143452935332062804519861768577138064573259069358975042122485627388192604101700302462362798592115038066626006203292151971117547825788365029212778673068841445412348366566041687247190622871297208478844658065023427261675103854131453079888921085982507674462174796122088474426882259116347806427809304457623578281770915649130727943036401841073052876646614496797452163682420483896962238944883515469778180754205404941121141599952645904787654598990827384683885441411829970587806867632776521257761008506280475694989732289455150722563701599552835784792787892378475275568451704113581492707873602587027241507884412830033365653914549639245334821262039117311046735253960816076120011217151854994903031629514573221302055487026138480621515317251868170670711382669181677841187033304861609062981031260474272177938456322715827189878292157663041430578246790062347759217920400856872733420989333068628730367636465064551304273677450773043171019346261214827671284261815921602597267030573778706635242318228554033918329328451836597811660816887574069149657366695399451977163530476815743551656054860762997757628606444388219917589048259254358441694421344436705112457275390625",
		"0.5403023058681397185594704958555785318789737039530258806063132922287686731767103380722879951346341700301934154242672285869862285136280035256635149087722026456653060623287636554810379311566025776640284286043366692017812351065106848975433243878434601646053373656175870219143335780037780263780528022552589948882631781788387243759097393407418459341665025578959743680402679565710214589639277870162089735459876789876164945626397889278949094061300698882180972876375063478762441763141045691438279791675986386217937407677843979434037325100615614204173839676230023544297002103390623811623287987178412049715000660987495684907814359780765658960715423795658408220614069561530422742718922893554271137549412369147629593458239490366605797325139239944180051692566520667214675204069476110320173267646327440417945329562672924843423546460029298423629611979148079742086487910832189641973781754022003095306885302971446512196936121388134045074972570593581133444925052909080976768554628154520121141808489601642421723177702083881285314379847744230155051966361158007017591582046619808944751148550841102122653076944135728372456492657514171872121435155305860386442665726120257630353953444221670261220705805623292334551636518107895571396829211893470388375917284217734121333575035514871632238034279250742569584498765995392820238318876135963364504277706146240234375",
		"0.968912421710644784098953478920868172610959295777507211448260716808741793725813506998889579335809685289859771728515625",
		"1.0",
		"0.946042343528386971548655274719127421974169853736868957009626534392384054569270779300182994120405055582523345947265625",
		"0.5403023058681397185594704958555785318789737039530258806063132922287686731767103380722879951346341700301934154242672285869862285136280035256635149087722026456653060623287636554810379311566025776640284286043366692017812351065106848975433243878434601646053373656175870219143335780037780263780528022552589948882631781788387243759097393407418459341665025578959743680402679565710214589639277870162089735459876789876164945626397889278949094061300698882180972876375063478762441763141045691438279791675986386217937407677843979434037325100615614204173839676230023544297002103390623811623287987178412049715000660987495684907814359780765658960715423795658408220614069561530422742718922893554271137549412369147629593458239490366605797325139239944180051692566520667214675204069476110320173267646327440417945329562672924843423546460029298423629611979148079742086487910832189641973781754022003095306885302971446512196936121388134045074972570593581133444925052909080976768554628154520121141808489601642421723177702083881285314379847744230155051966361158007017591582046619808944751148550841102122653076944135728372456492657514171872121435155305860386442665726120257630353953444221670261220705805623292334551636518107895571396829211893470388375917284217734121333575035514871632238034279250742569584498765995392820238318876135963364504277706146240234375",
		"0.2836621854632262640413454705014792126730041372592792257206146645426601396441260356011571632089582618925838153657699769828738164020634023542401271546832973668591965970780399287846132178786932147719489387057249755694634919271032742386467263343003762216059912285589934893404186137482489686474984451054771103493648603655340770275593072267521598532667125806036270513157359920795997662877180316554125323347714021810156977000099765637673158040510070215919734518464013962899531358902754977185164550708058199871247160015886870149163066556093196783440812874006182781633602366172183573725257608628866745555998550377960547686470245458626674802626292796549871502254193653242073557087898962773658255810053531370802115864997325060870572958692356612208396119690852149106422002310275167937817158288794799498840438933404405058238604063789542363766559068475105058515056790385539505205746157593694133334912162302627008416603294000615739181111749007688625782080015509782711717307370897514073199106746929535880890780214210719294387601483817437596151502942142722181977054762243644235893248214844557965863434535805464134779033144520828071538407409414463996182938827134980848798774952211899634360773590821173224015934994691863733763099257382589344052735164638723090648652823314675360473802118354232333173901927332659182734442282480813446454703807830810546875",
		"-0.839071529076452452453116298055375283024545196907005751226679269494945423257632785713674884440909521085910953206821682885350514489569383484553680293874464258013074629595233233440423524040185080138431677219644537006650333440982266180848432613265727672098505530500860099490052714852637696219860623419861733967504942346257426883650461547676693756422205686582533132927538864456011428540560137047577109540901355332714454892042154044099772961016405398585134639556422987015399944634792029504830682794892974110049771680445087591216461375304720930330852114906188605879368500242393092090599991056282929140339515922705640197683216586300345608814191348583643897583052496585038234458432300602598656172228800591628810870703061601079842675931288590090649649722933122710180797353905585641576282107791663819561479911160784057959370842865399560455463184561649904326467913053219927760200406450092385176263660350059186320075841952511620387666555291853410264149054381143798602019026379808898830628940629734269442061546806433026794983473131041596411391519359668360706593453379582689919628495055879290310156211842712529358591580000339115277557733264569390191184923563095676385834461753760385376605508361399123052214852866448517012467160704995276728328831739643710159503245565807479821789258404871375420699263398077037976545179716664243275062643380243567701961405078734621807138136709079831010972563517559541512724763379083015024662017822265625",
		"-0.825575442809343611824362325239811764094510857054278582264067433862269037693599066167126550050055810893236466513641579446123546053932400340779013543022627343780304207889697484852104716170210387967098979368968667582761587933736919053415297219989844530337025837550322738303417999408570541023209479831053490074672524731593393200211612747267787376420912055466715589243670137116181283653627378091767001266729750523670904813955028404725364549436654773973161631448021217524149606880882398338428340541423336476442278764927060929088442662022809429880029452211569992113848964950861453209500601571720208371076364637312918805739974730928391130012733244321591317884640789514132100142038718855057870855384682083191451172141637084656463668180074273357513986650562276924927490092054490207521493248546634803683905992004777812863248355102626794030220223795127136641375381651724046389855784239897808572076152799212203148992979849462537797025824681362178656342679605636982892165755312758468261428421302359039383473990693745647491195856344807310687765672840471336376933282820265780965117564858639226791165667469629053398278240657443123596015059287606479740468263567953594727359402416208210264162078048896806628308236533843407536450642386891796397818074083632855343175466156403711148552219179067451933509186842662133841059218230787130271284099320325244795528905606899665072032213644123281103167139222988313296269780039438046514987945556640625",
	}
	for i, inp := range inps {
		d, err := NewFromString(inp)
		if err != nil {
			t.FailNow()
		}
		s, err := NewFromString(sols[i])
		if err != nil {
			t.FailNow()
		}
		a := d.Cos()
		if !a.Equal(s) {
			t.Errorf("expected %s, got %s", s, a)
		}
	}
}

// For Tan
func TestTan(t *testing.T) {
	inps := []string{
		"-2.91919191919191919",
		"-1.0",
		"-0.25",
		"0.0",
		"0.33",
		"1.0",
		"5.0",
		"10",
	  "11000020.2407442310156021090304691671842603586882014729198302312846062338790031898128063403419218957424",
	}
	sols := []string{
		"0.226141565050579195128504091376587961550743024795950198353192839504353407742431869564825319685041904449462890625",
		"-1.5574077246549022",
		"-0.255341921221036275",
		"0.0",
		"0.342524867530038963",
		"1.5574077246549022",
		"-3.3805150062465857",
		"0.648360827459086657992079596346674836000458368506527382600293171822426231687419573290753760375082492828369140625",
		"0.683513254892486971935899897806108130234891903543492515192842067427622421368508870073710248504975849182264766875",
	}
	for i, inp := range inps {
		d, err := NewFromString(inp)
		if err != nil {
			t.FailNow()
		}
		s, err := NewFromString(sols[i])
		if err != nil {
			t.FailNow()
		}
		a := d.Tan()
		if !a.Equal(s) {
			t.Errorf("expected %s, got %s", s, a)
		}
	}
}
