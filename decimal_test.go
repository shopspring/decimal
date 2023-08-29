package xdecimal

import (
	"database/sql/driver"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"math"
	"math/big"
	"math/rand"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"testing/quick"
	"time"
)

func TestMain(m *testing.M) {
	// for test backwards compatibility
	TrimTrailingZeroes = true
	m.Run()
}

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
	{1e25, "10000000000000000000000000", "", ""},
	{1.5e14, "150000000000000", "", ""},
	{1.5e15, "1500000000000000", "", ""},
	{1.5e16, "15000000000000000", "", ""},
	{1.0001e25, "10001000000000000000000000", "", ""},
	{1.0001000000000000033e25, "10001000000000000000000000", "", ""},
	{2e25, "20000000000000000000000000", "", ""},
	{4e25, "40000000000000000000000000", "", ""},
	{8e25, "80000000000000000000000000", "", ""},
	{1e250, "10000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000", "", ""},
	{2e250, "20000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000", "", ""},
	{math.MaxInt64, strconv.FormatFloat(float64(math.MaxInt64), 'f', -1, 64), "", strconv.FormatInt(math.MaxInt64, 10)},
	{1.29067116156722e-309, "0.00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000129067116156722", "", "0.000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000001290671161567218558822290567835270536800098852722416870074139002112543896676308448335063375297788379444685193974290737962187240854947838776604607190387984577130572928111657710645015086812756013489109884753559084166516937690932698276436869274093950997935137476803610007959500457935217950764794724766740819156974617155861568214427828145972181876775307023388139991104942469299524961281641158436752347582767153796914843896176260096039358494077706152272661453132497761307744086665088096215425146090058519888494342944692629602847826300550628670375451325582843627504604013541465361435761965354140678551369499812124085312128659002910905639984075064968459581691226705666561364681985266583563078466180095375402399087817404368974165082030458595596655868575908243656158447265625000000000000000000000000000000000000004440000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"},
	// go Issue 29491.
	{498484681984085570, "498484681984085570", "", ""},
	{5.8339553793802237e+23, "583395537938022370000000", "", ""},
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
	"0.e0":       "0",
	".0e0":       "0",
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
		s := x.short
		d := NewFromFloat(x.float)
		if d.String() != s {
			t.Errorf("expected %s, got %s (float: %v) (%s, %d)",
				s, d.String(), x.float,
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

func TestNewFromFloatRandom(t *testing.T) {
	n := 0
	rng := rand.New(rand.NewSource(0xdead1337))
	for {
		n++
		if n == 10 {
			break
		}
		in := (rng.Float64() - 0.5) * math.MaxFloat64 * 2
		want, err := NewFromString(strconv.FormatFloat(in, 'f', -1, 64))
		if err != nil {
			t.Error(err)
			continue
		}
		got := NewFromFloat(in)
		if !want.Equal(got) {
			t.Errorf("in: %v, expected %s (%s, %d), got %s (%s, %d) ",
				in, want.String(), want.value.String(), want.exp,
				got.String(), got.value.String(), got.exp)
		}
	}
}

func TestNewFromFloatQuick(t *testing.T) {
	err := quick.Check(func(f float64) bool {
		want, werr := NewFromString(strconv.FormatFloat(f, 'f', -1, 64))
		if werr != nil {
			return true
		}
		got := NewFromFloat(f)
		return got.Equal(want)
	}, &quick.Config{})
	if err != nil {
		t.Error(err)
	}
}

func TestNewFromFloat32Random(t *testing.T) {
	n := 0
	rng := rand.New(rand.NewSource(0xdead1337))
	for {
		n++
		if n == 10 {
			break
		}
		in := float32((rng.Float64() - 0.5) * math.MaxFloat32 * 2)
		want, err := NewFromString(strconv.FormatFloat(float64(in), 'f', -1, 32))
		if err != nil {
			t.Error(err)
			continue
		}
		got := NewFromFloat32(in)
		if !want.Equal(got) {
			t.Errorf("in: %v, expected %s (%s, %d), got %s (%s, %d) ",
				in, want.String(), want.value.String(), want.exp,
				got.String(), got.value.String(), got.exp)
		}
	}
}

func TestNewFromFloat32Quick(t *testing.T) {
	err := quick.Check(func(f float32) bool {
		want, werr := NewFromString(strconv.FormatFloat(float64(f), 'f', -1, 32))
		if werr != nil {
			return true
		}
		got := NewFromFloat32(f)
		return got.Equal(want)
	}, &quick.Config{})
	if err != nil {
		t.Error(err)
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

func TestNewFromStringWithTooLongFractional(t *testing.T) {
	_, err := NewFromString("0e" + strconv.FormatInt(math.MaxInt64, 10) + "123")
	if err == nil {
		t.Errorf("must be error")
	}
}

func TestNewFromFormattedString(t *testing.T) {
	for _, testCase := range []struct {
		Formatted string
		Expected  string
		ReplRegex *regexp.Regexp
	}{
		{"$10.99", "10.99", regexp.MustCompile("[$]")},
		{"$ 12.1", "12.1", regexp.MustCompile("[$\\s]")},
		{"$61,690.99", "61690.99", regexp.MustCompile("[$,]")},
		{"1_000_000.00", "1000000.00", regexp.MustCompile("[_]")},
		{"41,410.00", "41410.00", regexp.MustCompile("[,]")},
		{"5200 USD", "5200", regexp.MustCompile("[USD\\s]")},
	} {
		dFormatted, err := NewFromFormattedString(testCase.Formatted, testCase.ReplRegex)
		if err != nil {
			t.Fatal(err)
		}

		dExact, err := NewFromString(testCase.Expected)
		if err != nil {
			t.Fatal(err)
		}

		if !dFormatted.Equal(dExact) {
			t.Errorf("expect %s, got %s", dExact, dFormatted)
		}
	}
}

func TestFloat64(t *testing.T) {
	for _, x := range testTable {
		if x.inexact == "" || x.inexact == "-" {
			continue
		}
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
		"512.99 USD",
		"$99.99",
		"51,850.00",
		"20_000_000.00",
		"$20_000_000.00",
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
		{"1.0", "1.0", true},
		{"10", "10.0", false},
		{"1.1", "1.10", false},
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
	var err any

	func() {
		defer func() {
			err = recover()
		}()

		RequireFromString(s)
	}()

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
	//nolint: gofmt
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

func TestNewFromInt(t *testing.T) {
	tests := map[int64]string{
		0:                   "0",
		1:                   "1",
		323412345:           "323412345",
		9223372036854775807: "9223372036854775807",
	}

	// add negatives
	for p, s := range tests {
		if p > 0 {
			tests[-p] = "-" + s
		}
	}

	for input, s := range tests {
		d := NewFromInt(input)
		if d.String() != s {
			t.Errorf("expected %s, got %s (%s, %d)",
				s, d.String(),
				d.value.String(), d.exp)
		}
	}
}

func TestNewFromInt32(t *testing.T) {
	tests := map[int32]string{
		0:          "0",
		1:          "1",
		323412345:  "323412345",
		2147483647: "2147483647",
	}

	// add negatives
	for p, s := range tests {
		if p > 0 {
			tests[-p] = "-" + s
		}
	}

	for input, s := range tests {
		d := NewFromInt32(input)
		if d.String() != s {
			t.Errorf("expected %s, got %s (%s, %d)",
				s, d.String(),
				d.value.String(), d.exp)
		}
	}
}

func TestNewFromBigIntWithExponent(t *testing.T) {
	type Inp struct {
		val *big.Int
		exp int32
	}
	//nolint: gofmt
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

func TestCopy(t *testing.T) {
	origin := New(1, 0)
	cpy := origin.Copy()

	if cpy.Cmp(origin) != 0 {
		t.Error("expecting copy and origin to be equals, but they are not")
	}

	if origin.value == cpy.value {
		t.Error("expecting copy and origin to have different value pointers")
	}

	//change value
	cpy = cpy.Add(New(1, 0))

	if cpy.Cmp(origin) == 0 {
		t.Error("expecting copy and origin to have different values, but they are equal")
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

		// make sure unquoted marshaling works too
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

		// make sure unquoted marshaling works too
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

	// make sure unquoted marshaling works too
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

func TestNullDecimalXML(t *testing.T) {
	// test valid values
	for _, x := range testTable {
		s := x.short
		var doc struct {
			XMLName xml.Name    `xml:"account"`
			Amount  NullDecimal `xml:"amount"`
		}
		docStr := `<account><amount>` + s + `</amount></account>`
		err := xml.Unmarshal([]byte(docStr), &doc)
		if err != nil {
			t.Errorf("error unmarshaling %s: %v", docStr, err)
		} else if doc.Amount.Decimal.String() != s {
			t.Errorf("expected %s, got %s (%s, %d)",
				s, doc.Amount.Decimal.String(),
				doc.Amount.Decimal.value.String(), doc.Amount.Decimal.exp)
		}

		out, err := xml.Marshal(&doc)
		if err != nil {
			t.Errorf("error marshaling %+v: %v", doc, err)
		} else if string(out) != docStr {
			t.Errorf("expected %s, got %s", docStr, string(out))
		}
	}

	var doc struct {
		XMLName xml.Name    `xml:"account"`
		Amount  NullDecimal `xml:"amount"`
	}

	// test for XML with empty body
	docStr := `<account><amount></amount></account>`
	err := xml.Unmarshal([]byte(docStr), &doc)
	if err != nil {
		t.Errorf("error unmarshaling: %s: %v", docStr, err)
	} else if doc.Amount.Valid {
		t.Errorf("expected null value to have Valid = false, got Valid = true and Decimal = %s (%s, %d)",
			doc.Amount.Decimal.String(),
			doc.Amount.Decimal.value.String(), doc.Amount.Decimal.exp)
	}

	expected := `<account><amount></amount></account>`
	out, err := xml.Marshal(&doc)
	if err != nil {
		t.Errorf("error marshaling %+v: %v", doc, err)
	} else if string(out) != expected {
		t.Errorf("expected %s, got %s", expected, string(out))
	}

	// test for empty XML
	docStr = `<account></account>`
	err = xml.Unmarshal([]byte(docStr), &doc)
	if err != nil {
		t.Errorf("error unmarshaling: %s: %v", docStr, err)
	} else if doc.Amount.Valid {
		t.Errorf("expected null value to have Valid = false, got Valid = true and Decimal = %s (%s, %d)",
			doc.Amount.Decimal.String(),
			doc.Amount.Decimal.value.String(), doc.Amount.Decimal.exp)
	}

	expected = `<account><amount></amount></account>`
	out, err = xml.Marshal(&doc)
	if err != nil {
		t.Errorf("error marshaling %+v: %v", doc, err)
	} else if string(out) != expected {
		t.Errorf("expected %s, got %s", expected, string(out))
	}
}

func TestNullDecimalBadXML(t *testing.T) {
	for _, testCase := range []string{
		"o_o",
		"<abc",
		"<account><amount>7",
		`<html><body></body></html>`,
		`<account><amount>nope</amount></account>`,
		`0.333`,
	} {
		var doc struct {
			XMLName xml.Name    `xml:"account"`
			Amount  NullDecimal `xml:"amount"`
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
	//nolint: gofmt
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
			t.Fatal(err)
		}

		// test Round
		expected, err := NewFromString(test.expected)
		if err != nil {
			t.Fatal(err)
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

func TestDecimal_RoundCeilAndStringFixed(t *testing.T) {
	type testData struct {
		input         string
		places        int32
		expected      string
		expectedFixed string
	}
	tests := []testData{
		{"1.454", 0, "2", ""},
		{"1.454", 1, "1.5", ""},
		{"1.454", 2, "1.46", ""},
		{"1.454", 3, "1.454", ""},
		{"1.454", 4, "1.454", "1.4540"},
		{"1.454", 5, "1.454", "1.45400"},
		{"1.554", 0, "2", ""},
		{"1.554", 1, "1.6", ""},
		{"1.554", 2, "1.56", ""},
		{"0.554", 0, "1", ""},
		{"0.454", 0, "1", ""},
		{"0.454", 5, "0.454", "0.45400"},
		{"0", 0, "0", ""},
		{"0", 1, "0", "0.0"},
		{"0", 2, "0", "0.00"},
		{"0", -1, "0", ""},
		{"5", 2, "5", "5.00"},
		{"5", 1, "5", "5.0"},
		{"5", 0, "5", ""},
		{"500", 2, "500", "500.00"},
		{"500", -2, "500", ""},
		{"545", -1, "550", ""},
		{"545", -2, "600", ""},
		{"545", -3, "1000", ""},
		{"545", -4, "10000", ""},
		{"499", -3, "1000", ""},
		{"499", -4, "10000", ""},
		{"1.1001", 2, "1.11", ""},
		{"-1.1001", 2, "-1.10", ""},
		{"-1.454", 0, "-1", ""},
		{"-1.454", 1, "-1.4", ""},
		{"-1.454", 2, "-1.45", ""},
		{"-1.454", 3, "-1.454", ""},
		{"-1.454", 4, "-1.454", "-1.4540"},
		{"-1.454", 5, "-1.454", "-1.45400"},
		{"-1.554", 0, "-1", ""},
		{"-1.554", 1, "-1.5", ""},
		{"-1.554", 2, "-1.55", ""},
		{"-0.554", 0, "0", ""},
		{"-0.454", 0, "0", ""},
		{"-0.454", 5, "-0.454", "-0.45400"},
		{"-5", 2, "-5", "-5.00"},
		{"-5", 1, "-5", "-5.0"},
		{"-5", 0, "-5", ""},
		{"-500", 2, "-500", "-500.00"},
		{"-500", -2, "-500", ""},
		{"-545", -1, "-540", ""},
		{"-545", -2, "-500", ""},
		{"-545", -3, "0", ""},
		{"-545", -4, "0", ""},
		{"-499", -3, "0", ""},
		{"-499", -4, "0", ""},
	}

	for _, test := range tests {
		d, err := NewFromString(test.input)
		if err != nil {
			t.Fatal(err)
		}

		// test Round
		expected, err := NewFromString(test.expected)
		if err != nil {
			t.Fatal(err)
		}
		got := d.RoundCeil(test.places)
		if !got.Equal(expected) {
			t.Errorf("Rounding ceil %s to %d places, got %s, expected %s",
				d, test.places, got, expected)
		}

		// test StringFixed
		if test.expectedFixed == "" {
			test.expectedFixed = test.expected
		}
		gotStr := got.StringFixed(test.places)
		if gotStr != test.expectedFixed {
			t.Errorf("(%s).StringFixed(%d): got %s, expected %s",
				d, test.places, gotStr, test.expectedFixed)
		}
	}
}

func TestDecimal_RoundFloorAndStringFixed(t *testing.T) {
	type testData struct {
		input         string
		places        int32
		expected      string
		expectedFixed string
	}
	tests := []testData{
		{"1.454", 0, "1", ""},
		{"1.454", 1, "1.4", ""},
		{"1.454", 2, "1.45", ""},
		{"1.454", 3, "1.454", ""},
		{"1.454", 4, "1.454", "1.4540"},
		{"1.454", 5, "1.454", "1.45400"},
		{"1.554", 0, "1", ""},
		{"1.554", 1, "1.5", ""},
		{"1.554", 2, "1.55", ""},
		{"0.554", 0, "0", ""},
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
		{"500", -2, "500", ""},
		{"545", -1, "540", ""},
		{"545", -2, "500", ""},
		{"545", -3, "0", ""},
		{"545", -4, "0", ""},
		{"499", -3, "0", ""},
		{"499", -4, "0", ""},
		{"1.1001", 2, "1.10", ""},
		{"-1.1001", 2, "-1.11", ""},
		{"-1.454", 0, "-2", ""},
		{"-1.454", 1, "-1.5", ""},
		{"-1.454", 2, "-1.46", ""},
		{"-1.454", 3, "-1.454", ""},
		{"-1.454", 4, "-1.454", "-1.4540"},
		{"-1.454", 5, "-1.454", "-1.45400"},
		{"-1.554", 0, "-2", ""},
		{"-1.554", 1, "-1.6", ""},
		{"-1.554", 2, "-1.56", ""},
		{"-0.554", 0, "-1", ""},
		{"-0.454", 0, "-1", ""},
		{"-0.454", 5, "-0.454", "-0.45400"},
		{"-5", 2, "-5", "-5.00"},
		{"-5", 1, "-5", "-5.0"},
		{"-5", 0, "-5", ""},
		{"-500", 2, "-500", "-500.00"},
		{"-500", -2, "-500", ""},
		{"-545", -1, "-550", ""},
		{"-545", -2, "-600", ""},
		{"-545", -3, "-1000", ""},
		{"-545", -4, "-10000", ""},
		{"-499", -3, "-1000", ""},
		{"-499", -4, "-10000", ""},
	}

	for _, test := range tests {
		d, err := NewFromString(test.input)
		if err != nil {
			t.Fatal(err)
		}

		// test Round
		expected, err := NewFromString(test.expected)
		if err != nil {
			t.Fatal(err)
		}
		got := d.RoundFloor(test.places)
		if !got.Equal(expected) {
			t.Errorf("Rounding floor %s to %d places, got %s, expected %s",
				d, test.places, got, expected)
		}

		// test StringFixed
		if test.expectedFixed == "" {
			test.expectedFixed = test.expected
		}
		gotStr := got.StringFixed(test.places)
		if gotStr != test.expectedFixed {
			t.Errorf("(%s).StringFixed(%d): got %s, expected %s",
				d, test.places, gotStr, test.expectedFixed)
		}
	}
}

func TestDecimal_RoundUpAndStringFixed(t *testing.T) {
	type testData struct {
		input         string
		places        int32
		expected      string
		expectedFixed string
	}
	tests := []testData{
		{"1.454", 0, "2", ""},
		{"1.454", 1, "1.5", ""},
		{"1.454", 2, "1.46", ""},
		{"1.454", 3, "1.454", ""},
		{"1.454", 4, "1.454", "1.4540"},
		{"1.454", 5, "1.454", "1.45400"},
		{"1.554", 0, "2", ""},
		{"1.554", 1, "1.6", ""},
		{"1.554", 2, "1.56", ""},
		{"0.554", 0, "1", ""},
		{"0.454", 0, "1", ""},
		{"0.454", 5, "0.454", "0.45400"},
		{"0", 0, "0", ""},
		{"0", 1, "0", "0.0"},
		{"0", 2, "0", "0.00"},
		{"0", -1, "0", ""},
		{"5", 2, "5", "5.00"},
		{"5", 1, "5", "5.0"},
		{"5", 0, "5", ""},
		{"500", 2, "500", "500.00"},
		{"500", -2, "500", ""},
		{"545", -1, "550", ""},
		{"545", -2, "600", ""},
		{"545", -3, "1000", ""},
		{"545", -4, "10000", ""},
		{"499", -3, "1000", ""},
		{"499", -4, "10000", ""},
		{"1.1001", 2, "1.11", ""},
		{"-1.1001", 2, "-1.11", ""},
		{"-1.454", 0, "-2", ""},
		{"-1.454", 1, "-1.5", ""},
		{"-1.454", 2, "-1.46", ""},
		{"-1.454", 3, "-1.454", ""},
		{"-1.454", 4, "-1.454", "-1.4540"},
		{"-1.454", 5, "-1.454", "-1.45400"},
		{"-1.554", 0, "-2", ""},
		{"-1.554", 1, "-1.6", ""},
		{"-1.554", 2, "-1.56", ""},
		{"-0.554", 0, "-1", ""},
		{"-0.454", 0, "-1", ""},
		{"-0.454", 5, "-0.454", "-0.45400"},
		{"-5", 2, "-5", "-5.00"},
		{"-5", 1, "-5", "-5.0"},
		{"-5", 0, "-5", ""},
		{"-500", 2, "-500", "-500.00"},
		{"-500", -2, "-500", ""},
		{"-545", -1, "-550", ""},
		{"-545", -2, "-600", ""},
		{"-545", -3, "-1000", ""},
		{"-545", -4, "-10000", ""},
		{"-499", -3, "-1000", ""},
		{"-499", -4, "-10000", ""},
	}

	for _, test := range tests {
		d, err := NewFromString(test.input)
		if err != nil {
			t.Fatal(err)
		}

		// test Round
		expected, err := NewFromString(test.expected)
		if err != nil {
			t.Fatal(err)
		}
		got := d.RoundUp(test.places)
		if !got.Equal(expected) {
			t.Errorf("Rounding up %s to %d places, got %s, expected %s",
				d, test.places, got, expected)
		}

		// test StringFixed
		if test.expectedFixed == "" {
			test.expectedFixed = test.expected
		}
		gotStr := got.StringFixed(test.places)
		if gotStr != test.expectedFixed {
			t.Errorf("(%s).StringFixed(%d): got %s, expected %s",
				d, test.places, gotStr, test.expectedFixed)
		}
	}
}

func TestDecimal_RoundDownAndStringFixed(t *testing.T) {
	type testData struct {
		input         string
		places        int32
		expected      string
		expectedFixed string
	}
	tests := []testData{
		{"1.454", 0, "1", ""},
		{"1.454", 1, "1.4", ""},
		{"1.454", 2, "1.45", ""},
		{"1.454", 3, "1.454", ""},
		{"1.454", 4, "1.454", "1.4540"},
		{"1.454", 5, "1.454", "1.45400"},
		{"1.554", 0, "1", ""},
		{"1.554", 1, "1.5", ""},
		{"1.554", 2, "1.55", ""},
		{"0.554", 0, "0", ""},
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
		{"500", -2, "500", ""},
		{"545", -1, "540", ""},
		{"545", -2, "500", ""},
		{"545", -3, "0", ""},
		{"545", -4, "0", ""},
		{"499", -3, "0", ""},
		{"499", -4, "0", ""},
		{"1.1001", 2, "1.10", ""},
		{"-1.1001", 2, "-1.10", ""},
		{"-1.454", 0, "-1", ""},
		{"-1.454", 1, "-1.4", ""},
		{"-1.454", 2, "-1.45", ""},
		{"-1.454", 3, "-1.454", ""},
		{"-1.454", 4, "-1.454", "-1.4540"},
		{"-1.454", 5, "-1.454", "-1.45400"},
		{"-1.554", 0, "-1", ""},
		{"-1.554", 1, "-1.5", ""},
		{"-1.554", 2, "-1.55", ""},
		{"-0.554", 0, "0", ""},
		{"-0.454", 0, "0", ""},
		{"-0.454", 5, "-0.454", "-0.45400"},
		{"-5", 2, "-5", "-5.00"},
		{"-5", 1, "-5", "-5.0"},
		{"-5", 0, "-5", ""},
		{"-500", 2, "-500", "-500.00"},
		{"-500", -2, "-500", ""},
		{"-545", -1, "-540", ""},
		{"-545", -2, "-500", ""},
		{"-545", -3, "0", ""},
		{"-545", -4, "0", ""},
		{"-499", -3, "0", ""},
		{"-499", -4, "0", ""},
	}

	for _, test := range tests {
		d, err := NewFromString(test.input)
		if err != nil {
			t.Fatal(err)
		}

		// test Round
		expected, err := NewFromString(test.expected)
		if err != nil {
			t.Fatal(err)
		}
		got := d.RoundDown(test.places)
		if !got.Equal(expected) {
			t.Errorf("Rounding down %s to %d places, got %s, expected %s",
				d, test.places, got, expected)
		}

		// test StringFixed
		if test.expectedFixed == "" {
			test.expectedFixed = test.expected
		}
		gotStr := got.StringFixed(test.places)
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

	//nolint: gofmt
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

	//nolint: gofmt
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

	//nolint: gofmt
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

	//nolint: gofmt
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

	//nolint: gofmt
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
	//nolint: gofmt
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
								d := sign1.Mul(New(int64(v1), e1))
								d2 := sign2.Mul(New(int64(v2), e2))
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
				const want = "Decimal does not support this Cash rounding interval `231`. Supported: 5, 10, 25, 50, 100"
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

func TestDecimal_Mod(t *testing.T) {
	type Inp struct {
		a string
		b string
	}

	//nolint: gofmt
	inputs := map[Inp]string{
		Inp{"3", "2"}:                           "1",
		Inp{"3451204593", "2454495034"}:         "996709559",
		Inp{"9999999999", "1275"}:               "324",
		Inp{"9999999999.9999998", "1275.49"}:    "239.2399998",
		Inp{"24544.95034", "0.3451204593"}:      "0.3283950433",
		Inp{"0.499999999999999999", "0.25"}:     "0.249999999999999999",
		Inp{"0.989512958912895912", "0.000001"}: "0.000000958912895912",
		Inp{"0.1", "0.1"}:                       "0",
		Inp{"0", "1.001"}:                       "0",
		Inp{"-7.5", "2"}:                        "-1.5",
		Inp{"7.5", "-2"}:                        "1.5",
		Inp{"-7.5", "-2"}:                       "-1.5",
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

func TestBigInt(t *testing.T) {
	testCases := []struct {
		Dec       string
		BigIntRep string
	}{
		{"0.0", "0"},
		{"0.00000", "0"},
		{"0.01", "0"},
		{"12.1", "12"},
		{"9999.999", "9999"},
		{"-32768.01234", "-32768"},
		{"-572372.0000000001", "-572372"},
	}

	for _, testCase := range testCases {
		d, err := NewFromString(testCase.Dec)
		if err != nil {
			t.Fatal(err)
		}
		if d.BigInt().String() != testCase.BigIntRep {
			t.Errorf("expect %s, got %s", testCase.BigIntRep, d.BigInt())
		}
	}
}

func TestBigFloat(t *testing.T) {
	testCases := []struct {
		Dec         string
		BigFloatRep string
	}{
		{"0.0", "0"},
		{"0.00000", "0"},
		{"0.01", "0.01"},
		{"12.1", "12.1"},
		{"9999.999", "9999.999"},
		{"-32768.01234", "-32768.01234"},
		{"-572372.0000000001", "-572372"},
		{"512.012345123451234512345", "512.0123451"},
		{"1.010101010101010101010101010101", "1.01010101"},
		{"55555555.555555555555555555555", "55555555.56"},
	}

	for _, testCase := range testCases {
		d, err := NewFromString(testCase.Dec)
		if err != nil {
			t.Fatal(err)
		}
		if d.BigFloat().String() != testCase.BigFloatRep {
			t.Errorf("expect %s, got %s", testCase.BigFloatRep, d.BigFloat())
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
	dbvalue := 54.33
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
	if _, ok := any(d).(driver.Valuer); !ok {
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

func TestDecimal_IsInteger(t *testing.T) {
	for _, testCase := range []struct {
		Dec       string
		IsInteger bool
	}{
		{"0", true},
		{"0.0000", true},
		{"0.01", false},
		{"0.01010101010000", false},
		{"12.0", true},
		{"12.00000000000000", true},
		{"12.10000", false},
		{"9999.0000", true},
		{"99999999.000000000", true},
		{"-656323444.0000000000000", true},
		{"-32768.01234", false},
		{"-32768.0123423562623600000", false},
	} {
		d, err := NewFromString(testCase.Dec)
		if err != nil {
			t.Fatal(err)
		}
		if d.IsInteger() != testCase.IsInteger {
			t.Errorf("expect %t, got %t, for %s", testCase.IsInteger, d.IsInteger(), testCase.Dec)
		}
	}
}

func TestDecimal_ExpHullAbrham(t *testing.T) {
	for _, testCase := range []struct {
		Dec              string
		OverallPrecision uint32
		ExpectedDec      string
	}{
		{"0", 1, "1"},
		{"0.00", 5, "1"},
		{"0.5", 5, "1.6487"},
		{"0.569297", 10, "1.767024397"},
		{"0.569297", 16, "1.76702439654095"},
		{"0.569297", 20, "1.7670243965409496521"},
		{"1.0", 0, "3"},
		{"1.0", 1, "3"},
		{"1.0", 5, "2.7183"},
		{"1.0", 10, "2.718281828"},
		{"3.0", 0, "20"},
		{"3.0", 2, "20"},
		{"5.26", 0, "200"},
		{"5.26", 2, "190"},
		{"5.26", 10, "192.4814913"},
		{"5.2663117716", 2, "190"},
		{"5.2663117716", 10, "193.7002327"},
		{"26.1", 2, "220000000000"},
		{"26.1", 10, "216314672100"},
		{"26.1", 20, "216314672147.05767284"},
		{"50.1591", 10, "6078834887000000000000"},
		{"50.1591", 30, "6078834886623434464595.07937141"},
		{"-0.5", 5, "0.60653"},
		{"-0.569297", 10, "0.5659231429"},
		{"-0.569297", 16, "0.565923142859576"},
		{"-0.569297", 20, "0.56592314285957604443"},
		{"-1.0", 1, "0.4"},
		{"-1.0", 5, "0.36788"},
		{"-3.0", 1, "0"},
		{"-3.0", 2, "0.05"},
		{"-3.0", 10, "0.0497870684"},
		{"-5.26", 2, "0.01"},
		{"-5.26", 10, "0.0051953047"},
		{"-5.2663117716", 2, "0.01"},
		{"-5.2663117716", 10, "0.0051626164"},
		{"-26.1", 2, "0"},
		{"-26.1", 15, "0.000000000004623"},
		{"-50.1591", 10, "0"},
		{"-50.1591", 30, "0.000000000000000000000164505208"},
	} {
		d, _ := NewFromString(testCase.Dec)
		expected, _ := NewFromString(testCase.ExpectedDec)

		exp, err := d.ExpHullAbrham(testCase.OverallPrecision)
		if err != nil {
			t.Fatal(err)
		}

		if exp.Cmp(expected) != 0 {
			t.Errorf("expected %s, got %s, for decimal %s", testCase.ExpectedDec, exp.String(), testCase.Dec)
		}

	}
}

func TestDecimal_ExpTaylor(t *testing.T) {
	for _, testCase := range []struct {
		Dec         string
		Precision   int32
		ExpectedDec string
	}{
		{"0", 1, "1"},
		{"0.00", 5, "1"},
		{"0", -1, "0"},
		{"0.5", 5, "1.64872"},
		{"0.569297", 10, "1.7670243965"},
		{"0.569297", 16, "1.7670243965409497"},
		{"0.569297", 20, "1.76702439654094965215"},
		{"1.0", 0, "3"},
		{"1.0", 1, "2.7"},
		{"1.0", 5, "2.71828"},
		{"1.0", -1, "0"},
		{"1.0", -5, "0"},
		{"3.0", 1, "20.1"},
		{"3.0", 2, "20.09"},
		{"5.26", 0, "192"},
		{"5.26", 2, "192.48"},
		{"5.26", 10, "192.4814912972"},
		{"5.26", -2, "200"},
		{"5.2663117716", 2, "193.70"},
		{"5.2663117716", 10, "193.7002326701"},
		{"26.1", 2, "216314672147.06"},
		{"26.1", 20, "216314672147.05767284062928674083"},
		{"26.1", -2, "216314672100"},
		{"26.1", -10, "220000000000"},
		{"50.1591", 10, "6078834886623434464595.0793714061"},
		{"-0.5", 5, "0.60653"},
		{"-0.569297", 10, "0.5659231429"},
		{"-0.569297", 16, "0.565923142859576"},
		{"-0.569297", 20, "0.56592314285957604443"},
		{"-1.0", 1, "0.4"},
		{"-1.0", 5, "0.36788"},
		{"-1.0", -1, "0"},
		{"-1.0", -5, "0"},
		{"-3.0", 1, "0.1"},
		{"-3.0", 2, "0.05"},
		{"-3.0", 10, "0.0497870684"},
		{"-5.26", 2, "0.01"},
		{"-5.26", 10, "0.0051953047"},
		{"-5.26", -2, "0"},
		{"-5.2663117716", 2, "0.01"},
		{"-5.2663117716", 10, "0.0051626164"},
		{"-26.1", 2, "0"},
		{"-26.1", 15, "0.000000000004623"},
		{"-26.1", -2, "0"},
		{"-26.1", -10, "0"},
		{"-50.1591", 10, "0"},
		{"-50.1591", 30, "0.000000000000000000000164505208"},
	} {
		d, _ := NewFromString(testCase.Dec)
		expected, _ := NewFromString(testCase.ExpectedDec)

		exp, err := d.ExpTaylor(testCase.Precision)
		if err != nil {
			t.Fatal(err)
		}

		if exp.Cmp(expected) != 0 {
			t.Errorf("expected %s, got %s", testCase.ExpectedDec, exp.String())
		}
	}
}

func TestDecimal_NumDigits(t *testing.T) {
	for _, testCase := range []struct {
		Dec               string
		ExpectedNumDigits int
	}{
		{"0", 1},
		{"0.00", 1},
		{"1.0", 2},
		{"3.0", 2},
		{"5.26", 3},
		{"5.2663117716", 11},
		{"3164836416948884.2162426426426267863", 35},
		{"26.1", 3},
		{"529.1591", 7},
		{"-1.0", 2},
		{"-3.0", 2},
		{"-5.26", 3},
		{"-5.2663117716", 11},
		{"-26.1", 3},
		{"", 1},
	} {
		d, _ := NewFromString(testCase.Dec)

		nums := d.NumDigits()
		if nums != testCase.ExpectedNumDigits {
			t.Errorf("expected %d digits for decimal %s, got %d", testCase.ExpectedNumDigits, testCase.Dec, nums)
		}
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

func TestDecimal_CoefficientInt64(t *testing.T) {
	type Inp struct {
		Dec         string
		Coefficient int64
	}

	testCases := []Inp{
		{"1", 1},
		{"1.111", 1111},
		{"1.000000", 1000000},
		{"1.121215125511", 1121215125511},
		{"100000000000000000", 100000000000000000},
		{"9223372036854775807", 9223372036854775807},
		{"10000000000000000000", -8446744073709551616}, // undefined value
	}

	// add negative cases
	for _, tc := range testCases {
		testCases = append(testCases, Inp{"-" + tc.Dec, -tc.Coefficient})
	}

	for _, tc := range testCases {
		d := RequireFromString(tc.Dec)
		coefficient := d.CoefficientInt64()
		if coefficient != tc.Coefficient {
			t.Errorf("expect coefficient %d, got %d, for decimal %s", tc.Coefficient, coefficient, tc.Dec)
		}
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
	var dbvaluePtr any
	err := a.Scan(dbvaluePtr)
	if err != nil {
		// Scan failed... no need to test result value
		t.Errorf("a.Scan(nil) failed with message: %s", err)
	} else {
		if a.Valid {
			t.Errorf("%s is not null", a.Decimal)
		}
	}

	dbvalue := 54.33
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

	// after reuse a.Scan a.Decimal should be null
	err = a.Scan(nil)
	if err != nil {
		// Scan failed... no need to test result value
		t.Errorf("a.Scan(nil) failed with message: %s", err)
	} else {
		if a.Valid {
			t.Errorf("%s is not null", a.Decimal)
		}

		if !a.Decimal.Equal(Zero) {
			t.Errorf("%v does not equal Zero", a)
		}
	}
}

func TestNullDecimal_Value(t *testing.T) {
	// Make sure this does implement the database/sql's driver.Valuer interface
	var nullDecimal NullDecimal
	if _, ok := any(nullDecimal).(driver.Valuer); !ok {
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
			t.Errorf("error marshaling %v to binary: %v", d1, err)
		}

		// Restore from binary
		var d2 Decimal
		err = (&d2).UnmarshalBinary(b)
		if err != nil {
			t.Errorf("error unmarshaling from binary: %v", err)
		}

		// The restored decimal should equal the original
		if !d1.Equals(d2) {
			t.Errorf("expected %v when restoring, got %v", d1, d2)
		}
	}
}

func TestBinary_Zero(t *testing.T) {
	var d1 Decimal

	b, err := d1.MarshalBinary()
	if err != nil {
		t.Fatalf("error marshaling %v to binary: %v", d1, err)
	}

	var d2 Decimal
	err = (&d2).UnmarshalBinary(b)
	if err != nil {
		t.Errorf("error unmarshaling from binary: %v", err)
	}

	if !d1.Equals(d2) {
		t.Errorf("expected %v when restoring, got %v", d1, d2)
	}
}

func TestBinary_DataTooShort(t *testing.T) {
	var d Decimal

	err := d.UnmarshalBinary(nil) // nil slice has length 0
	if err == nil {
		t.Fatalf("expected error, got %v", d)
	}
}

func TestBinary_InvalidValue(t *testing.T) {
	var d Decimal

	err := d.UnmarshalBinary([]byte{0, 0, 0, 0, 'x'}) // valid exponent, invalid value
	if err == nil {
		t.Fatalf("expected error, got %v", d)
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
		"-1.24076438822058001027437062753106",
		"-0.78539816339744833061616997868383",
		"-0.24497866312686415",
		"0.0",
		"0.318747560420644443",
		"0.78539816339744833061616997868383",
		"1.37340076694501580123233995736766",
		"1.47112767430373453123233995736766",
		"1.57079623588597296123259450235374",
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
	sols := []string{"-0.22057186252002995641471297726318877448242875710373383657841216606788849153474483300147427943530288911869356126149550184271061369789963810497434594683859566879253561990821788142048867910104964466745284318577343435957806286762494529983369776697504436326725441516925396488258485248699247367113416543705253919473126183478178486954138205996912770183192357029798618739277146694040778731661407420114923656224752540889120768",
		"-0.841470984807896544828551915928318375739843472469519282898610111931110319333748010828751784005573402229699531838022117989945539661588502120624574802425114599802714611508860519655182175315926637327774878594985045816542706701485174683683726979309922117859910272413672784175028365607893544855897795184024100973080880074046886009375162838756876336134083638363801171409953672944184918309063800980214873465660723218405962257950683415203634506166523593278",
		"-0.2474039592545229296662577977006816864013671875",
		"0",
		"0.3240430283948683457891331120415701894104386268737728",
		"0.841470984807896544828551915928318375739843472469519282898610111931110319333748010828751784005573402229699531838022117989945539661588502120624574802425114599802714611508860519655182175315926637327774878594985045816542706701485174683683726979309922117859910272413672784175028365607893544855897795184024100973080880074046886009375162838756876336134083638363801171409953672944184918309063800980214873465660723218405962257950683415203634506166523593278",
		"-0.958924274663138409032065951037351417114444405831206421994322505831797734568720303321152847999323782235893449831846516332891972309733806145798957570823292783131379570446989311599459252931842975162373777189193072018951049969744350662993214861042908755303566670204873618202680865638534865944483058650517380292320436016362659617294570185140789829574277032406195741535138712427510938542219940873171248862329526140744770994303733112530324791184417282382",
		"-0.54402111088937016772477554483765124109312606762621462357463994520238396180161585438877562935656067241573063207614488370477645194661241525080677431257416988398683714890165970942834453391033857378247849486306346743023618509617104937236345831462093934032592562972419977883837745736210439651143668255744843041350221801750331646628192115694352540293150183983357476391787825596543270240461102629075832777618592034309799936",
		"-0.564291758480422881634770440632390475980828840253516895637281099241819037882007239070203007530085741820184955492382572029153491807930868879341091067301689987699567034024159005627332722089169680203292567574310010066799858914647295684974242359142300929248173166551428537696685165964880390889406578530338963341989826231514301546476672476399906348023294571001061677668735117509440368611093448917120819545826797975989350435900286332895885871219875665471968941335407351099209738417818747252638912592184093301853338763294381446907254104878969784040526201729163408095795934201105630182851806342356035203279670146684553491616847294749721014579109870396804713831114709372638323643327823671187472335866664108658093206409882794958673673978956925250261545083579947618620746006004554405785185537391110314728988164693223775249484198058394348289545771967707968288542718255197272633789792059019367104377340604030147471453833808674013259696102003732963091159662478879760121731138091114134586544668859915547568540172541576138084166990547345181184322550297604278946942918844039406876827936831612756344331500301118652183156052728447906384772901595431751550607818380262138322673253023464533931883787069611052589166000316238423939491520880451263927981787175602294299295744",
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
		"-0.975370726167463467746508538219884948528729295145689640359666742268127382748782064668565276308334226452812521220478854320025773591423493734486361306323829818426063430805234608660356853863442937297855742231573288105774823103008774355455799906250461848079705023428527473474556899228935370709945979509634251305018978306493011197513482210179171510947538040406781879762352211326273272515279567525396877609653501706919545667682725671944948392322552266752",
		"0.54030230586813965874561515067176071767603141150991567490927772778673118786033739102174242337864109186439207498973007363884202112942385976796862442063752663646870430360736682397798633852405003167527051283327366631405990604840629657123985368031838052877290142895506386796217551784101265975360960112885444847880134909594560331781699767647860744559228420471946006511861233129745921297270844542687374552066388998112901504",
		"0.968912421710644784099084544806854121387004852294921875",
		"1",
		"0.9460423435283869715490383692051286742343482760977712222",
		"0.54030230586813965874561515067176071767603141150991567490927772778673118786033739102174242337864109186439207498973007363884202112942385976796862442063752663646870430360736682397798633852405003167527051283327366631405990604840629657123985368031838052877290142895506386796217551784101265975360960112885444847880134909594560331781699767647860744559228420471946006511861233129745921297270844542687374552066388998112901504",
		"0.28366218546322646623291670213892426815646045792775066552057877370468842342090306560693620285882975471913545189522117672866861003904575909174769890684057564495184019705963607555427518763375472432216131070235796577209064861003009894615394882021220247535890708789312783718414424224334988974848162884228012265684775651099758365989567444515619764427493598258393280941942356912304265535918025036942025858493361644535438208",
		"-0.839071529076452222947082170022504835457755803801719612447629165523199043803440231769716865070163209041973184176293170330332317060558438085478980463542480791358920580076809381102480339018809694514100495572097422057215638383077242523713704127605770444906854175870243452753002238589530499630034663296166308443155999957196346563161387705205277189957388653461251461388391745795979375660087266037741360406956289962327970672363315696841378765492754546688",
		"-0.82557544253149396284458404188071504476091346830440347376462206521981928020803354950315062147200396866217255527509254080721982393941347365824137698491042935894213870423296625749297033966815252917361266452901192457318047750698424190124169875103436588397415032138037063155981648677895645409699825582226442363080800781881653440538927704569142007751338851079530521979429507520541625303794665680584709171813053216867014700596866196844144286737568957809383224972108999354839705480223052622003994027222120126949093911643497423100187973906980635670000034664323357488815820848035808846624518774608931622703631130673844138378087837990739103263093532314835289302930152150130664948083902949999427848344301686172490282395687167681679607401220592559832932068966455384902377056623736013617949634746332323529256184776892339963173795176200590119077305668901887229709592836744082027738666294887303249770621722032438202753270710379312736193201366287952361100525126056993039858894987153270630277483613793395809214871734783742285495171911648254647287555645360520115341268930844095156502348405343740866836850201634640011708462641462047870611041595707018966032206807675586825362640000739202116391403514629284000986232673698892843586989003952425039512325844566790376383098534975022847888104706525937115931692008959513984157709954859352131323440787667052399474107219968",
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
		"0.2261415650505790298980791606748881031998682652",
		"-1.5574077246549025",
		"-0.255341921221036275",
		"0.0",
		"0.342524867530038963",
		"1.5574077246549025",
		"-3.3805150062465829",
		"0.6483608274590872485524085572681343280321117494",
		"0.68351325561491170753499935023939368502774607234006019034769919811202010905597996164029250820702097041244539696",
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

func TestNewNullDecimal(t *testing.T) {
	d := NewFromInt(1)
	nd := NewNullDecimal(d)

	if !nd.Valid {
		t.Errorf("expected NullDecimal to be valid")
	}
	if nd.Decimal != d {
		t.Errorf("expected NullDecimal to hold the provided Decimal")
	}
}

func ExampleNewFromFloat32() {
	fmt.Println(NewFromFloat32(123.123123123123).String())
	fmt.Println(NewFromFloat32(.123123123123123).String())
	fmt.Println(NewFromFloat32(-1e13).String())
	// OUTPUT:
	//123.12312
	//0.123123124
	//-10000000000000
}

func ExampleNewFromFloat() {
	fmt.Println(NewFromFloat(123.123123123123).String())
	fmt.Println(NewFromFloat(.123123123123123).String())
	fmt.Println(NewFromFloat(-1e13).String())
	// OUTPUT:
	//123.123123123123
	//0.123123123123123
	//-10000000000000
}
