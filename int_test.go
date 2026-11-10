package safeint

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"testing"
)

// ---------------------------------------------------------------------------
// 1. Constructor tests
// ---------------------------------------------------------------------------

func TestConstructors(t *testing.T) {
	t.Run("New_int64", func(t *testing.T) {
		v := New[int64](42)
		if v.Val() != 42 {
			t.Fatalf("New[int64](42).Val() = %d, want 42", v.Val())
		}
	})
	t.Run("Zero_int64", func(t *testing.T) {
		v := Zero[int64]()
		if v.Val() != 0 {
			t.Fatalf("Zero[int64]().Val() = %d, want 0", v.Val())
		}
	})
	t.Run("New_uint8_max", func(t *testing.T) {
		v := New[uint8](255)
		if v.Val() != 255 {
			t.Fatalf("New[uint8](255).Val() = %d, want 255", v.Val())
		}
	})
}

// ---------------------------------------------------------------------------
// 2. Checked method tests (*Overflow)
// ---------------------------------------------------------------------------

func TestAddOverflow(t *testing.T) {
	tests := []struct {
		name string
		a, b Int[int8]
		want int8
		ok   bool
	}{
		{"normal", New[int8](10), New[int8](20), 30, true},
		{"max_boundary", New[int8](126), New[int8](1), 127, true},
		{"overflow_signed", New[int8](127), New[int8](1), -128, false},
		{"negative_normal", New[int8](-50), New[int8](-50), -100, true},
		{"neg_overflow", New[int8](-128), New[int8](-1), 127, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := tc.a.AddOverflow(tc.b)
			if ok != tc.ok {
				t.Fatalf("ok = %v, want %v", ok, tc.ok)
			}
			if got.Val() != tc.want {
				t.Fatalf("result = %d, want %d", got.Val(), tc.want)
			}
		})
	}
}

func TestSubOverflow(t *testing.T) {
	tests := []struct {
		name string
		a, b Int[int8]
		want int8
		ok   bool
	}{
		{"normal", New[int8](30), New[int8](10), 20, true},
		{"negative_result", New[int8](-10), New[int8](20), -30, true},
		{"underflow_signed", New[int8](-128), New[int8](1), 127, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := tc.a.SubOverflow(tc.b)
			if ok != tc.ok {
				t.Fatalf("ok = %v, want %v", ok, tc.ok)
			}
			if got.Val() != tc.want {
				t.Fatalf("result = %d, want %d", got.Val(), tc.want)
			}
		})
	}
}

func TestMulOverflow(t *testing.T) {
	tests := []struct {
		name string
		a, b Int[int8]
		want int8
		ok   bool
	}{
		{"normal", New[int8](6), New[int8](7), 42, true},
		{"by_zero", New[int8](127), New[int8](0), 0, true},
		{"negative_normal", New[int8](-5), New[int8](3), -15, true},
		{"min_times_neg1", New[int8](-128), New[int8](-1), -128, false},
		{"overflow_positive", New[int8](64), New[int8](2), -128, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := tc.a.MulOverflow(tc.b)
			if ok != tc.ok {
				t.Fatalf("ok = %v, want %v", ok, tc.ok)
			}
			if got.Val() != tc.want {
				t.Fatalf("result = %d, want %d", got.Val(), tc.want)
			}
		})
	}
}

func TestDivOverflow(t *testing.T) {
	tests := []struct {
		name string
		a, b Int[int8]
		want int8
		ok   bool
	}{
		{"normal", New[int8](42), New[int8](7), 6, true},
		{"negative", New[int8](-42), New[int8](7), -6, true},
		{"min_div_neg1", New[int8](-128), New[int8](-1), 0, false},
		{"div_by_zero", New[int8](1), New[int8](0), 0, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := tc.a.DivOverflow(tc.b)
			if ok != tc.ok {
				t.Fatalf("ok = %v, want %v", ok, tc.ok)
			}
			if got.Val() != tc.want {
				t.Fatalf("result = %d, want %d", got.Val(), tc.want)
			}
		})
	}
}

func TestDivModOverflow(t *testing.T) {
	tests := []struct {
		name  string
		a, b  Int[int8]
		wantQ int8
		wantR int8
		ok    bool
	}{
		{"normal", New[int8](17), New[int8](5), 3, 2, true},
		{"negative_dividend", New[int8](-17), New[int8](5), -3, -2, true},
		{"min_div_neg1", New[int8](-128), New[int8](-1), 0, 0, false},
		{"div_by_zero", New[int8](10), New[int8](0), 0, 0, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			q, r, ok := tc.a.DivModOverflow(tc.b)
			if ok != tc.ok {
				t.Fatalf("ok = %v, want %v", ok, tc.ok)
			}
			if q.Val() != tc.wantQ || r.Val() != tc.wantR {
				t.Fatalf("result = (%d, %d), want (%d, %d)", q.Val(), r.Val(), tc.wantQ, tc.wantR)
			}
		})
	}
}

func TestModOverflow(t *testing.T) {
	tests := []struct {
		name string
		a, b Int[int8]
		want int8
		ok   bool
	}{
		{"normal", New[int8](17), New[int8](5), 2, true},
		{"min_mod_neg1", New[int8](-128), New[int8](-1), 0, true},
		{"mod_by_zero", New[int8](10), New[int8](0), 0, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := tc.a.ModOverflow(tc.b)
			if ok != tc.ok {
				t.Fatalf("ok = %v, want %v", ok, tc.ok)
			}
			if got.Val() != tc.want {
				t.Fatalf("result = %d, want %d", got.Val(), tc.want)
			}
		})
	}
}

func TestNegOverflow(t *testing.T) {
	tests := []struct {
		name string
		a    Int[int8]
		want int8
		ok   bool
	}{
		{"positive", New[int8](42), -42, true},
		{"negative", New[int8](-42), 42, true},
		{"zero", New[int8](0), 0, true},
		{"min_signed", New[int8](-128), -128, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := tc.a.NegOverflow()
			if ok != tc.ok {
				t.Fatalf("ok = %v, want %v", ok, tc.ok)
			}
			if got.Val() != tc.want {
				t.Fatalf("result = %d, want %d", got.Val(), tc.want)
			}
		})
	}
}

func TestAbsOverflow(t *testing.T) {
	tests := []struct {
		name string
		a    Int[int8]
		want int8
		ok   bool
	}{
		{"positive", New[int8](42), 42, true},
		{"negative", New[int8](-42), 42, true},
		{"zero", New[int8](0), 0, true},
		{"min_signed", New[int8](-128), -128, false},
		{"max_signed", New[int8](127), 127, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := tc.a.AbsOverflow()
			if ok != tc.ok {
				t.Fatalf("ok = %v, want %v", ok, tc.ok)
			}
			if got.Val() != tc.want {
				t.Fatalf("result = %d, want %d", got.Val(), tc.want)
			}
		})
	}
}

func TestPowOverflow(t *testing.T) {
	tests := []struct {
		name string
		a    Int[uint8]
		exp  uint
		want uint8
		ok   bool
	}{
		{"2^7", New[uint8](2), 7, 128, true},
		{"2^8_overflow", New[uint8](2), 8, 0, false},
		{"3^6_overflow_wraps", New[uint8](3), 6, 217, false}, // 729 mod 256
		{"3^5", New[uint8](3), 5, 243, true},
		{"anything^0", New[uint8](255), 0, 1, true},
		{"0^0", New[uint8](0), 0, 1, true},
		{"1^100", New[uint8](1), 100, 1, true},
		{"5^2", New[uint8](5), 2, 25, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := tc.a.PowOverflow(tc.exp)
			if ok != tc.ok {
				t.Fatalf("ok = %v, want %v", ok, tc.ok)
			}
			if got.Val() != tc.want {
				t.Fatalf("result = %d, want %d", got.Val(), tc.want)
			}
		})
	}
}

func TestLshOverflow(t *testing.T) {
	tests := []struct {
		name string
		a    Int[int8]
		n    uint
		want int8
		ok   bool
	}{
		{"normal", New[int8](1), 6, 64, true},
		{"overflow_into_sign", New[int8](1), 7, -128, false},
		{"zero_shift", New[int8](127), 0, 127, true},
		{"zero_value", New[int8](0), 10, 0, true},
		{"shift_too_large", New[int8](1), 8, 0, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := tc.a.LshOverflow(tc.n)
			if ok != tc.ok {
				t.Fatalf("ok = %v, want %v", ok, tc.ok)
			}
			// The value on other overflows is not specified (see Lsh).
			if (ok || tc.n >= 8) && got.Val() != tc.want {
				t.Fatalf("result = %d, want %d", got.Val(), tc.want)
			}
		})
	}
}

func TestMulDivOverflow(t *testing.T) {
	tests := []struct {
		name    string
		a, b, c Int[int64]
		want    int64
		ok      bool
	}{
		{
			"maxint64_times2_div2",
			New[int64](math.MaxInt64), New[int64](2), New[int64](2),
			math.MaxInt64, true,
		},
		{
			"normal",
			New[int64](100), New[int64](200), New[int64](10),
			2000, true,
		},
		{
			"div_by_zero",
			New[int64](10), New[int64](20), New[int64](0),
			0, false,
		},
		{
			"negative_result",
			New[int64](-100), New[int64](200), New[int64](10),
			-2000, true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := tc.a.MulDivOverflow(tc.b, tc.c)
			if ok != tc.ok {
				t.Fatalf("ok = %v, want %v", ok, tc.ok)
			}
			if got.Val() != tc.want {
				t.Fatalf("result = %d, want %d", got.Val(), tc.want)
			}
		})
	}
}

func TestMulModOverflow(t *testing.T) {
	tests := []struct {
		name    string
		a, b, c Int[int64]
		want    int64
		ok      bool
	}{
		{
			"7*5_mod_3",
			New[int64](7), New[int64](5), New[int64](3),
			2, true,
		},
		{
			"normal",
			New[int64](100), New[int64](200), New[int64](7),
			1, true, // 100*200=20000, 20000%7=1
		},
		{
			"mod_by_zero",
			New[int64](10), New[int64](20), New[int64](0),
			0, false,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := tc.a.MulModOverflow(tc.b, tc.c)
			if ok != tc.ok {
				t.Fatalf("ok = %v, want %v", ok, tc.ok)
			}
			if got.Val() != tc.want {
				t.Fatalf("result = %d, want %d", got.Val(), tc.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// 3. Wrapping method tests
// ---------------------------------------------------------------------------

func TestWrappingAdd(t *testing.T) {
	tests := []struct {
		name string
		a, b Int[int8]
		want int8
	}{
		{"normal", New[int8](10), New[int8](20), 30},
		{"wraps", New[int8](127), New[int8](1), -128},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.a.Add(tc.b)
			if got.Val() != tc.want {
				t.Fatalf("result = %d, want %d", got.Val(), tc.want)
			}
		})
	}
}

func TestWrappingSub(t *testing.T) {
	tests := []struct {
		name string
		a, b Int[uint8]
		want uint8
	}{
		{"normal", New[uint8](50), New[uint8](30), 20},
		{"wraps", New[uint8](0), New[uint8](1), 255},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.a.Sub(tc.b)
			if got.Val() != tc.want {
				t.Fatalf("result = %d, want %d", got.Val(), tc.want)
			}
		})
	}
}

func TestWrappingMul(t *testing.T) {
	tests := []struct {
		name string
		a, b Int[int8]
		want int8
	}{
		{"normal", New[int8](6), New[int8](7), 42},
		{"wraps", New[int8](64), New[int8](4), 0},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.a.Mul(tc.b)
			if got.Val() != tc.want {
				t.Fatalf("result = %d, want %d", got.Val(), tc.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// 4. Must method tests
// ---------------------------------------------------------------------------

func TestMustAdd(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		var got Int[int8]
		expectNoPanic(t, "MustAdd normal", func() {
			got = New[int8](10).MustAdd(New[int8](20))
		})
		if got.Val() != 30 {
			t.Fatalf("result = %d, want 30", got.Val())
		}
	})
	t.Run("overflow_panics", func(t *testing.T) {
		expectPanic(t, "MustAdd overflow", "safeint: Add overflow", func() {
			New[int8](127).MustAdd(New[int8](1))
		})
	})
}

func TestMustSub(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		var got Int[int8]
		expectNoPanic(t, "MustSub normal", func() {
			got = New[int8](30).MustSub(New[int8](10))
		})
		if got.Val() != 20 {
			t.Fatalf("result = %d, want 20", got.Val())
		}
	})
	t.Run("overflow_panics", func(t *testing.T) {
		expectPanic(t, "MustSub overflow", "safeint: Sub overflow", func() {
			New[uint8](0).MustSub(New[uint8](1))
		})
	})
}

func TestMustMul(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		var got Int[int8]
		expectNoPanic(t, "MustMul normal", func() {
			got = New[int8](6).MustMul(New[int8](7))
		})
		if got.Val() != 42 {
			t.Fatalf("result = %d, want 42", got.Val())
		}
	})
	t.Run("overflow_panics", func(t *testing.T) {
		expectPanic(t, "MustMul overflow", "safeint: Mul overflow", func() {
			New[int8](-128).MustMul(New[int8](-1))
		})
	})
}

func TestMustDiv(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		var got Int[int8]
		expectNoPanic(t, "MustDiv normal", func() {
			got = New[int8](42).MustDiv(New[int8](7))
		})
		if got.Val() != 6 {
			t.Fatalf("result = %d, want 6", got.Val())
		}
	})
	t.Run("div_by_zero_panics", func(t *testing.T) {
		expectPanic(t, "MustDiv div-by-zero", "safeint: Div overflow", func() {
			New[int8](1).MustDiv(New[int8](0))
		})
	})
	t.Run("overflow_panics", func(t *testing.T) {
		expectPanic(t, "MustDiv overflow", "safeint: Div overflow", func() {
			New[int8](-128).MustDiv(New[int8](-1))
		})
	})
}

// ---------------------------------------------------------------------------
// 5. Comparison tests
// ---------------------------------------------------------------------------

// checkCompare checks Cmp, Eq, Lt, Gt, Lte and Gte on (a, b) against want,
// the expected sign of a - b computed without overflow.
func checkCompare[T Integer](t *testing.T, a, b T, want int) {
	t.Helper()
	x, y := New(a), New(b)
	got := []bool{x.Eq(y), x.Lt(y), x.Gt(y), x.Lte(y), x.Gte(y)}
	exp := []bool{want == 0, want < 0, want > 0, want <= 0, want >= 0}
	names := []string{"Eq", "Lt", "Gt", "Lte", "Gte"}
	if c := x.Cmp(y); c != want {
		t.Errorf("%T: Cmp(%d, %d) = %d, want %d", a, a, b, c, want)
	}
	for i := range got {
		if got[i] != exp[i] {
			t.Errorf("%T: %s(%d, %d) = %v, want %v", a, names[i], a, b, got[i], exp[i])
		}
	}
}

// TestCompare includes pairs whose difference overflows T, so an
// implementation that compares by subtraction fails.
func TestCompare(t *testing.T) {
	int64Pairs := []struct {
		a, b int64
		want int
	}{
		{1, 2, -1}, {42, 42, 0}, {10, 5, 1}, {-5, 3, -1}, {3, -5, 1}, {-10, -5, -1}, {-7, -7, 0},
		{math.MinInt64, 1, -1}, {1, math.MinInt64, 1},
		{math.MinInt64, math.MaxInt64, -1}, {math.MaxInt64, math.MinInt64, 1},
		{math.MaxInt64, -1, 1}, {math.MinInt64, math.MinInt64, 0},
	}
	for _, p := range int64Pairs {
		checkCompare(t, p.a, p.b, p.want)
	}
	uint8Pairs := []struct {
		a, b uint8
		want int
	}{
		{0, 255, -1}, {128, 128, 0}, {255, 0, 1}, {1, 2, -1},
	}
	for _, p := range uint8Pairs {
		checkCompare(t, p.a, p.b, p.want)
	}
	uint64Pairs := []struct {
		a, b uint64
		want int
	}{
		{0, math.MaxUint64, -1}, {math.MaxUint64, 0, 1}, {math.MaxUint64, math.MaxUint64, 0}, {1, 1 << 63, -1},
	}
	for _, p := range uint64Pairs {
		checkCompare(t, p.a, p.b, p.want)
	}
}

func TestIsZero(t *testing.T) {
	tests := []struct {
		name string
		a    Int[int64]
		want bool
	}{
		{"zero", New[int64](0), true},
		{"positive", New[int64](1), false},
		{"negative", New[int64](-1), false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.a.IsZero(); got != tc.want {
				t.Fatalf("IsZero = %v, want %v", got, tc.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// 6. String tests
// ---------------------------------------------------------------------------

// stringerInt defines its own String method, which Int[T].String must ignore.
type stringerInt int64

func (stringerInt) String() string { return "custom" }

func TestString(t *testing.T) {
	tests := []struct {
		name string
		val  string
		got  string
	}{
		{"positive_int64", "42", New[int64](42).String()},
		{"negative_int64", "-7", New[int64](-7).String()},
		{"uint8_max", "255", New[uint8](255).String()},
		{"zero", "0", New[int64](0).String()},
		{"uint64_max", "18446744073709551615", New[uint64](math.MaxUint64).String()},
		{"named_stringer_ignored", "25", New[stringerInt](25).String()},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.got != tc.val {
				t.Fatalf("String() = %q, want %q", tc.got, tc.val)
			}
		})
	}
}

// formatterInt defines its own Format method, which Int[T].Format must ignore.
type formatterInt int64

func (formatterInt) Format(f fmt.State, _ rune) { fmt.Fprint(f, "custom") }

func TestFormat(t *testing.T) {
	tests := []struct {
		format string
		arg    interface{}
		want   string
	}{
		{"%d", New[int64](255), "255"},
		{"%x", New[int64](255), "ff"},
		{"%x", New[int64](-255), "-ff"},
		{"%X", New[uint8](255), "FF"},
		{"%#x", New[uint64](255), "0xff"},
		{"%o", New[int8](8), "10"},
		{"%b", New[uint8](5), "101"},
		{"%c", New[int32](65), "A"},
		{"%08d", New[int64](-42), "-0000042"},
		{"%+d", New[int64](5), "+5"},
		{"% d", New[int64](5), " 5"},
		{"%-5d|", New[int32](7), "7    |"},
		{"%.3d", New[int16](7), "007"},
		{"%d", New[uint64](math.MaxUint64), "18446744073709551615"},
		{"%v", New[int64](-7), "-7"},
		{"%+v", New[int64](-7), "-7"},
		{"%#v", New[int64](255), "255"},
		{"%s", New[int64](255), "255"},
		{"%6s", New[int64](255), "   255"},
		{"%q", New[int64](255), `"255"`},
		{"%v", struct{ N Int[uint8] }{New[uint8](9)}, "{9}"},
		{"%v", New[stringerInt](25), "25"},
		{"%x", New[formatterInt](255), "ff"},
		{"%f", New[int8](3), "%!f(int8=3)"},
		{"%f", New[stringerInt](3), "%!f(int64=3)"},
		{"%d", New[formatterInt](255), "255"},
	}
	for _, tc := range tests {
		t.Run(tc.format+"_"+fmt.Sprintf("%T", tc.arg), func(t *testing.T) {
			if got := fmt.Sprintf(tc.format, tc.arg); got != tc.want {
				t.Fatalf("Sprintf(%q) = %q, want %q", tc.format, got, tc.want)
			}
		})
	}
	if got := fmt.Sprint(New[int64](-3)); got != "-3" {
		t.Fatalf("Sprint = %q, want %q", got, "-3")
	}
}

// ---------------------------------------------------------------------------
// 7. ConvertInt tests
// ---------------------------------------------------------------------------

func TestConvertInt(t *testing.T) {
	t.Run("int64_to_int8_fits", func(t *testing.T) {
		got, ok := ConvertInt[int64, int8](New[int64](127))
		if !ok {
			t.Fatalf("ok = false, want true")
		}
		if got.Val() != 127 {
			t.Fatalf("result = %d, want 127", got.Val())
		}
	})
	t.Run("int64_to_int8_overflow", func(t *testing.T) {
		_, ok := ConvertInt[int64, int8](New[int64](128))
		if ok {
			t.Fatalf("ok = true, want false")
		}
	})
	t.Run("uint8_to_int8_overflow", func(t *testing.T) {
		_, ok := ConvertInt[uint8, int8](New[uint8](200))
		if ok {
			t.Fatalf("ok = true, want false")
		}
	})
	t.Run("int8_to_uint8_negative", func(t *testing.T) {
		_, ok := ConvertInt[int8, uint8](New[int8](-1))
		if ok {
			t.Fatalf("ok = true, want false")
		}
	})
	t.Run("int8_to_uint8_fits", func(t *testing.T) {
		got, ok := ConvertInt[int8, uint8](New[int8](100))
		if !ok {
			t.Fatalf("ok = false, want true")
		}
		if got.Val() != 100 {
			t.Fatalf("result = %d, want 100", got.Val())
		}
	})
	t.Run("uint8_to_int8_fits", func(t *testing.T) {
		got, ok := ConvertInt[uint8, int8](New[uint8](127))
		if !ok {
			t.Fatalf("ok = false, want true")
		}
		if got.Val() != 127 {
			t.Fatalf("result = %d, want 127", got.Val())
		}
	})
	t.Run("int64_to_int8_negative_fits", func(t *testing.T) {
		got, ok := ConvertInt[int64, int8](New[int64](-128))
		if !ok {
			t.Fatalf("ok = false, want true")
		}
		if got.Val() != -128 {
			t.Fatalf("result = %d, want -128", got.Val())
		}
	})
	t.Run("int64_to_int8_negative_overflow", func(t *testing.T) {
		_, ok := ConvertInt[int64, int8](New[int64](-129))
		if ok {
			t.Fatalf("ok = true, want false")
		}
	})
}

// ---------------------------------------------------------------------------
// 8. Value semantics test
// ---------------------------------------------------------------------------

func TestValueSemantics(t *testing.T) {
	t.Run("add_does_not_mutate", func(t *testing.T) {
		a := New[int64](10)
		b := a.Add(New[int64](5))
		if a.Val() != 10 {
			t.Fatalf("a was mutated: got %d, want 10", a.Val())
		}
		if b.Val() != 15 {
			t.Fatalf("b = %d, want 15", b.Val())
		}
	})
	t.Run("sub_does_not_mutate", func(t *testing.T) {
		a := New[int64](20)
		b := a.Sub(New[int64](5))
		if a.Val() != 20 {
			t.Fatalf("a was mutated: got %d, want 20", a.Val())
		}
		if b.Val() != 15 {
			t.Fatalf("b = %d, want 15", b.Val())
		}
	})
	t.Run("mul_does_not_mutate", func(t *testing.T) {
		a := New[int64](10)
		b := a.Mul(New[int64](3))
		if a.Val() != 10 {
			t.Fatalf("a was mutated: got %d, want 10", a.Val())
		}
		if b.Val() != 30 {
			t.Fatalf("b = %d, want 30", b.Val())
		}
	})
	t.Run("checked_add_does_not_mutate", func(t *testing.T) {
		a := New[int64](10)
		b, ok := a.AddOverflow(New[int64](5))
		if !ok {
			t.Fatalf("ok = false, want true")
		}
		if a.Val() != 10 {
			t.Fatalf("a was mutated: got %d, want 10", a.Val())
		}
		if b.Val() != 15 {
			t.Fatalf("b = %d, want 15", b.Val())
		}
	})
	t.Run("must_add_does_not_mutate", func(t *testing.T) {
		a := New[int64](10)
		b := a.MustAdd(New[int64](5))
		if a.Val() != 10 {
			t.Fatalf("a was mutated: got %d, want 10", a.Val())
		}
		if b.Val() != 15 {
			t.Fatalf("b = %d, want 15", b.Val())
		}
	})
}

// ---------------------------------------------------------------------------
// JSON serialization tests
// ---------------------------------------------------------------------------

type jsonIndex int64

func (v jsonIndex) MarshalJSON() ([]byte, error) { return json.Marshal(int64(v) + 1) }
func (v *jsonIndex) UnmarshalJSON(data []byte) error {
	var n int64
	if err := json.Unmarshal(data, &n); err != nil {
		return err
	}
	*v = jsonIndex(n - 1)
	return nil
}

func TestJSON_NamedCodecNumericRoundTrip(t *testing.T) {
	original := jsonIndex(1)
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatal(err)
	}
	var bare jsonIndex
	if err := json.Unmarshal(data, &bare); err != nil {
		t.Fatal(err)
	}
	if string(data) != "2" || bare != original {
		t.Fatalf("bare codec: JSON=%s, value=%d", data, bare)
	}
	data, err = json.Marshal(New(original))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "1" {
		t.Fatalf("wrapped JSON = %s, want 1", data)
	}
	var decoded Int[jsonIndex]
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Val() != original {
		t.Fatalf("roundtrip value = %d, want %d", decoded.Val(), original)
	}
}

type jsonQuotedUint uint64

func (v jsonQuotedUint) MarshalJSON() ([]byte, error) {
	return json.Marshal(strconv.FormatUint(uint64(v), 10))
}
func (v *jsonQuotedUint) UnmarshalJSON(data []byte) error {
	var text string
	if err := json.Unmarshal(data, &text); err != nil {
		return err
	}
	n, err := strconv.ParseUint(text, 10, 64)
	if err == nil {
		*v = jsonQuotedUint(n)
	}
	return err
}

func TestJSON_NamedQuotedCodecNumericRoundTrip(t *testing.T) {
	original := jsonQuotedUint(math.MaxUint64)
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatal(err)
	}
	var bare jsonQuotedUint
	if err := json.Unmarshal(data, &bare); err != nil {
		t.Fatal(err)
	}
	if string(data) != `"18446744073709551615"` || bare != original {
		t.Fatalf("bare codec: JSON=%s, value=%d", data, bare)
	}
	data, err = json.Marshal(New(original))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "18446744073709551615" {
		t.Fatalf("wrapped JSON = %s", data)
	}
	var decoded Int[jsonQuotedUint]
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Val() != original {
		t.Fatalf("roundtrip value = %d, want %d", decoded.Val(), original)
	}
}

func checkJSONDecodeState[T Integer](t *testing.T, boundaries []T, invalid []string) {
	t.Helper()
	v := New(T(7))
	// Syntax errors are left out: encoding/json rejects them before it calls
	// UnmarshalJSON, so they cannot detect a regression here.
	for _, data := range append([]string{`"8"`, "true", "{}", "[]", "[1]", `{"x":1}`, "1.5", "1.0", "1e1"}, invalid...) {
		if err := json.Unmarshal([]byte(data), &v); err == nil {
			t.Fatalf("accepted invalid input %s", data)
		}
		if v.Val() != 7 {
			t.Fatalf("input %s changed value to %d", data, v.Val())
		}
	}
	if err := json.Unmarshal([]byte(" \n null \t"), &v); err != nil {
		t.Fatal(err)
	}
	if v.Val() != 7 {
		t.Fatalf("null changed value to %d", v.Val())
	}
	for _, want := range boundaries {
		data, err := json.Marshal(New(want))
		if err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal(data, &v); err != nil {
			t.Fatal(err)
		}
		if v.Val() != want {
			t.Fatalf("input %s decoded to %d, want %d", data, v.Val(), want)
		}
	}
}

func TestJSON_NumericDecodeState(t *testing.T) {
	t.Run("int8", func(t *testing.T) { checkJSONDecodeState(t, []int8{-128, 127}, []string{"-129", "128"}) })
	t.Run("uint8", func(t *testing.T) { checkJSONDecodeState(t, []uint8{0, 255}, []string{"-1", "256"}) })
	t.Run("int32", func(t *testing.T) {
		checkJSONDecodeState(t, []int32{math.MinInt32, -100, 0, 42, math.MaxInt32}, []string{"-2147483649", "2147483648"})
	})
	t.Run("int64", func(t *testing.T) {
		checkJSONDecodeState(t, []int64{math.MinInt64, -12345, math.MaxInt64}, []string{"-9223372036854775809", "9223372036854775808"})
	})
	t.Run("uint64", func(t *testing.T) {
		checkJSONDecodeState(t, []uint64{0, math.MaxInt64 + 1, math.MaxUint64}, []string{"-1", "18446744073709551616"})
	})
	t.Run("named_signed", func(t *testing.T) {
		checkJSONDecodeState(t, []jsonIndex{math.MinInt64, math.MaxInt64}, []string{"-9223372036854775809", "9223372036854775808"})
	})
	t.Run("named_unsigned", func(t *testing.T) {
		checkJSONDecodeState(t, []jsonQuotedUint{0, math.MaxUint64}, []string{"-1", "18446744073709551616"})
	})
}

func TestMarshalJSON(t *testing.T) {
	tests := []struct {
		name string
		val  json.Marshaler
		want string
	}{
		{"int32_positive", New[int32](42), "42"},
		{"int32_negative", New[int32](-1), "-1"},
		{"int32_zero", New[int32](0), "0"},
		{"int32_max", New[int32](math.MaxInt32), "2147483647"},
		{"int32_min", New[int32](math.MinInt32), "-2147483648"},
		{"int8_max", New[int8](127), "127"},
		{"int8_min", New[int8](-128), "-128"},
		{"uint8_zero", New[uint8](0), "0"},
		{"uint8_max", New[uint8](255), "255"},
		{"int16_max", New[int16](math.MaxInt16), "32767"},
		{"int16_min", New[int16](math.MinInt16), "-32768"},
		{"uint16_max", New[uint16](math.MaxUint16), "65535"},
		{"uint32_max", New[uint32](math.MaxUint32), "4294967295"},
		{"int64_max", New[int64](math.MaxInt64), "9223372036854775807"},
		{"int64_min", New[int64](math.MinInt64), "-9223372036854775808"},
		{"uint64_zero", New[uint64](0), "0"},
		{"uint64_max_int64", New[uint64](math.MaxInt64), "9223372036854775807"},
		{"uint64_max_int64_plus1", New[uint64](uint64(math.MaxInt64) + 1), "9223372036854775808"},
		{"uint64_max", New[uint64](math.MaxUint64), "18446744073709551615"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := tc.val.MarshalJSON()
			if err != nil {
				t.Fatalf("MarshalJSON error: %v", err)
			}
			if string(got) != tc.want {
				t.Fatalf("MarshalJSON = %s, want %s", got, tc.want)
			}
		})
	}
}

// TestUnmarshalJSONErrorType verifies that every rejected input reports the
// wrapper type Int[T] in UnmarshalTypeError.Type, regardless of whether the
// rejection came from the intermediate int64/uint64 decode or from the
// narrowing conversion.
func TestUnmarshalJSONErrorType(t *testing.T) {
	check := func(t *testing.T, err error, target interface{}, wantValue string) {
		t.Helper()
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		var ute *json.UnmarshalTypeError
		if !errors.As(err, &ute) {
			t.Fatalf("error type = %T, want *json.UnmarshalTypeError", err)
		}
		if want := reflect.TypeOf(target).Elem(); ute.Type != want {
			t.Fatalf("Type = %v, want %v", ute.Type, want)
		}
		if ute.Value != wantValue {
			t.Fatalf("Value = %q, want %q", ute.Value, wantValue)
		}
	}
	t.Run("uint8_overflow_narrowing", func(t *testing.T) {
		var v Int[uint8]
		check(t, json.Unmarshal([]byte("256"), &v), &v, "number 256")
	})
	t.Run("uint8_negative_intermediate", func(t *testing.T) {
		var v Int[uint8]
		check(t, json.Unmarshal([]byte("-1"), &v), &v, "number -1")
	})
	t.Run("int8_overflow_narrowing", func(t *testing.T) {
		var v Int[int8]
		check(t, json.Unmarshal([]byte("128"), &v), &v, "number 128")
	})
	t.Run("int8_underflow_narrowing", func(t *testing.T) {
		var v Int[int8]
		check(t, json.Unmarshal([]byte("-129"), &v), &v, "number -129")
	})
	t.Run("int64_overflow_intermediate", func(t *testing.T) {
		var v Int[int64]
		check(t, json.Unmarshal([]byte("9223372036854775808"), &v), &v, "number 9223372036854775808")
	})
	t.Run("uint64_overflow_intermediate", func(t *testing.T) {
		var v Int[uint64]
		check(t, json.Unmarshal([]byte("18446744073709551616"), &v), &v, "number 18446744073709551616")
	})
	t.Run("uint64_negative_intermediate", func(t *testing.T) {
		var v Int[uint64]
		check(t, json.Unmarshal([]byte("-1"), &v), &v, "number -1")
	})
	t.Run("string_into_int32", func(t *testing.T) {
		var v Int[int32]
		check(t, json.Unmarshal([]byte(`"hello"`), &v), &v, "string")
	})
	t.Run("float_into_int32", func(t *testing.T) {
		var v Int[int32]
		check(t, json.Unmarshal([]byte("1.5"), &v), &v, "number 1.5")
	})
	t.Run("syntax_error_passes_through", func(t *testing.T) {
		var v Int[int32]
		err := v.UnmarshalJSON([]byte("abc"))
		var se *json.SyntaxError
		if !errors.As(err, &se) {
			t.Fatalf("error type = %T, want *json.SyntaxError", err)
		}
	})
	// Whether encoding/json sets Field on an error returned by an Unmarshaler
	// depends on the Go version (see UnmarshalJSON), so only Type and Value
	// are asserted for nested fields.
	t.Run("nested_struct_field", func(t *testing.T) {
		var s struct {
			Count Int[uint8] `json:"count"`
		}
		check(t, json.Unmarshal([]byte(`{"count": -1}`), &s), &s.Count, "number -1")
	})
}

func TestJSON_StructField(t *testing.T) {
	type ClubMember struct {
		Balance Int[int32] `json:"balance"`
	}

	t.Run("marshal", func(t *testing.T) {
		m := ClubMember{Balance: New[int32](9999)}
		data, err := json.Marshal(m)
		if err != nil {
			t.Fatalf("Marshal error: %v", err)
		}
		want := `{"balance":9999}`
		if string(data) != want {
			t.Fatalf("Marshal = %s, want %s", data, want)
		}
	})
	t.Run("unmarshal", func(t *testing.T) {
		var m ClubMember
		if err := json.Unmarshal([]byte(`{"balance":-500}`), &m); err != nil {
			t.Fatalf("Unmarshal error: %v", err)
		}
		if m.Balance.Val() != -500 {
			t.Fatalf("Balance = %d, want -500", m.Balance.Val())
		}
	})
	t.Run("unmarshal_null_field", func(t *testing.T) {
		var m ClubMember
		if err := json.Unmarshal([]byte(`{"balance":null}`), &m); err != nil {
			t.Fatalf("Unmarshal error: %v", err)
		}
		if m.Balance.Val() != 0 {
			t.Fatalf("Balance = %d, want 0", m.Balance.Val())
		}
	})
	t.Run("unmarshal_missing_field", func(t *testing.T) {
		var m ClubMember
		if err := json.Unmarshal([]byte(`{}`), &m); err != nil {
			t.Fatalf("Unmarshal error: %v", err)
		}
		if m.Balance.Val() != 0 {
			t.Fatalf("Balance = %d, want 0 for missing field", m.Balance.Val())
		}
	})
	t.Run("unmarshal_overflow_field", func(t *testing.T) {
		var m ClubMember
		err := json.Unmarshal([]byte(`{"balance":9999999999999}`), &m)
		if err == nil {
			t.Fatalf("expected error for overflow, got nil")
		}
	})
}

// ---------------------------------------------------------------------------
// SQL driver.Valuer tests
// ---------------------------------------------------------------------------

func TestValue(t *testing.T) {
	tests := []struct {
		name string
		val  driver.Valuer
		want int64
		ok   bool
	}{
		{"int32_positive", New[int32](42), 42, true},
		{"int32_negative", New[int32](-1), -1, true},
		{"int32_zero", New[int32](0), 0, true},
		{"int8_min", New[int8](math.MinInt8), math.MinInt8, true},
		{"int8_max", New[int8](math.MaxInt8), math.MaxInt8, true},
		{"uint8_zero", New[uint8](0), 0, true},
		{"uint8_max", New[uint8](math.MaxUint8), math.MaxUint8, true},
		{"int16_min", New[int16](math.MinInt16), math.MinInt16, true},
		{"int16_max", New[int16](math.MaxInt16), math.MaxInt16, true},
		{"uint16_max", New[uint16](math.MaxUint16), math.MaxUint16, true},
		{"int32_max", New[int32](math.MaxInt32), math.MaxInt32, true},
		{"int32_min", New[int32](math.MinInt32), math.MinInt32, true},
		{"uint32_max", New[uint32](math.MaxUint32), int64(math.MaxUint32), true},
		{"int64_max", New[int64](math.MaxInt64), math.MaxInt64, true},
		{"int64_min", New[int64](math.MinInt64), math.MinInt64, true},
		// uint64 up to MaxInt64 returns int64
		{"uint64_zero", New[uint64](0), 0, true},
		{"uint64_max_int64", New[uint64](math.MaxInt64), math.MaxInt64, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := tc.val.Value()
			if tc.ok {
				if err != nil {
					t.Fatalf("Value() error: %v", err)
				}
				v, ok := got.(int64)
				if !ok {
					t.Fatalf("Value() type = %T, want int64", got)
				}
				if v != tc.want {
					t.Fatalf("Value() = %d, want %d", v, tc.want)
				}
			} else {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
			}
		})
	}
}

func TestValue_Uint64FallbackString(t *testing.T) {
	tests := []struct {
		name string
		val  uint64
		want string
	}{
		{"max_int64_plus1", uint64(math.MaxInt64) + 1, "9223372036854775808"},
		{"max_uint64", math.MaxUint64, "18446744073709551615"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			v := New[uint64](tc.val)
			got, err := v.Value()
			if err != nil {
				t.Fatalf("Value() error: %v", err)
			}
			s, ok := got.(string)
			if !ok {
				t.Fatalf("Value() type = %T, want string", got)
			}
			if s != tc.want {
				t.Fatalf("Value() = %s, want %s", s, tc.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// SQL Scanner tests
// ---------------------------------------------------------------------------

func TestScan_Int64(t *testing.T) {
	t.Run("into_int32", func(t *testing.T) {
		tests := []struct {
			name string
			src  int64
			want int32
			ok   bool
		}{
			{"positive", 42, 42, true},
			{"negative", -100, -100, true},
			{"zero", 0, 0, true},
			{"max_int32", math.MaxInt32, math.MaxInt32, true},
			{"min_int32", math.MinInt32, math.MinInt32, true},
			{"overflow", math.MaxInt32 + 1, 0, false},
			{"underflow", math.MinInt32 - 1, 0, false},
		}
		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				var v Int[int32]
				err := v.Scan(tc.src)
				if tc.ok {
					if err != nil {
						t.Fatalf("Scan error: %v", err)
					}
					if v.Val() != tc.want {
						t.Fatalf("Val = %d, want %d", v.Val(), tc.want)
					}
				} else {
					if err == nil {
						t.Fatalf("expected error, got nil")
					}
				}
			})
		}
	})
	t.Run("into_int8", func(t *testing.T) {
		tests := []struct {
			name string
			src  int64
			want int8
			ok   bool
		}{
			{"min", -128, -128, true},
			{"max", 127, 127, true},
			{"overflow", 128, 0, false},
			{"underflow", -129, 0, false},
		}
		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				var v Int[int8]
				err := v.Scan(tc.src)
				if tc.ok {
					if err != nil {
						t.Fatalf("Scan error: %v", err)
					}
					if v.Val() != tc.want {
						t.Fatalf("Val = %d, want %d", v.Val(), tc.want)
					}
				} else {
					if err == nil {
						t.Fatalf("expected error, got nil")
					}
				}
			})
		}
	})
	t.Run("into_uint8", func(t *testing.T) {
		tests := []struct {
			name string
			src  int64
			want uint8
			ok   bool
		}{
			{"zero", 0, 0, true},
			{"max", 255, 255, true},
			{"overflow", 256, 0, false},
			{"negative", -1, 0, false},
		}
		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				var v Int[uint8]
				err := v.Scan(tc.src)
				if tc.ok {
					if err != nil {
						t.Fatalf("Scan error: %v", err)
					}
					if v.Val() != tc.want {
						t.Fatalf("Val = %d, want %d", v.Val(), tc.want)
					}
				} else {
					if err == nil {
						t.Fatalf("expected error, got nil")
					}
				}
			})
		}
	})
	t.Run("into_int64_identity", func(t *testing.T) {
		tests := []struct {
			name string
			src  int64
		}{
			{"max", math.MaxInt64},
			{"min", math.MinInt64},
			{"zero", 0},
		}
		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				var v Int[int64]
				if err := v.Scan(tc.src); err != nil {
					t.Fatalf("Scan error: %v", err)
				}
				if v.Val() != tc.src {
					t.Fatalf("Val = %d, want %d", v.Val(), tc.src)
				}
			})
		}
	})
	t.Run("into_uint64", func(t *testing.T) {
		tests := []struct {
			name string
			src  int64
			want uint64
			ok   bool
		}{
			{"zero", 0, 0, true},
			{"max_int64", math.MaxInt64, math.MaxInt64, true},
			{"negative", -1, 0, false},
		}
		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				var v Int[uint64]
				err := v.Scan(tc.src)
				if tc.ok {
					if err != nil {
						t.Fatalf("Scan error: %v", err)
					}
					if v.Val() != tc.want {
						t.Fatalf("Val = %d, want %d", v.Val(), tc.want)
					}
				} else {
					if err == nil {
						t.Fatalf("expected error, got nil")
					}
				}
			})
		}
	})
}

func checkScanUint64[T Integer](t *testing.T, src uint64, want T, ok bool) {
	t.Helper()
	v := New(T(7))
	err := v.Scan(src)
	if (err == nil) != ok {
		t.Fatalf("%T.Scan(uint64(%d)): error = %v, want success = %v", v, src, err, ok)
	}
	if !ok {
		want = 7 // Failed scans must preserve the previous value.
	}
	if v.Val() != want {
		t.Fatalf("%T.Scan(uint64(%d)): value = %d, want %d", v, src, v.Val(), want)
	}
}

// TestScan_Uint64 covers drivers that pass uint64 to a Scanner, although it
// is not a standard driver.Value type. go-sql-driver/mysql 1.8 and later does
// this for BIGINT UNSIGNED columns in the text protocol.
func TestScan_Uint64(t *testing.T) {
	checkScanUint64[uint64](t, 0, 0, true)
	checkScanUint64[uint64](t, math.MaxUint64, math.MaxUint64, true)
	checkScanUint64[int64](t, math.MaxInt64, math.MaxInt64, true)
	checkScanUint64[int64](t, math.MaxInt64+1, 0, false)
	checkScanUint64[uint8](t, math.MaxUint8, math.MaxUint8, true)
	checkScanUint64[uint8](t, math.MaxUint8+1, 0, false)
}

func TestScan_Nil(t *testing.T) {
	v := New[int32](99)
	if err := v.Scan(nil); err != nil {
		t.Fatalf("Scan(nil) error: %v", err)
	}
	if v.Val() != 0 {
		t.Fatalf("Val = %d, want 0 after Scan(nil)", v.Val())
	}
}

func TestScan_Float64(t *testing.T) {
	t.Run("whole_number", func(t *testing.T) {
		var v Int[int32]
		if err := v.Scan(float64(42)); err != nil {
			t.Fatalf("Scan error: %v", err)
		}
		if v.Val() != 42 {
			t.Fatalf("Val = %d, want 42", v.Val())
		}
	})
	t.Run("negative_whole", func(t *testing.T) {
		var v Int[int32]
		if err := v.Scan(float64(-7)); err != nil {
			t.Fatalf("Scan error: %v", err)
		}
		if v.Val() != -7 {
			t.Fatalf("Val = %d, want -7", v.Val())
		}
	})
	t.Run("zero", func(t *testing.T) {
		var v Int[int32]
		if err := v.Scan(float64(0)); err != nil {
			t.Fatalf("Scan error: %v", err)
		}
		if v.Val() != 0 {
			t.Fatalf("Val = %d, want 0", v.Val())
		}
	})
	t.Run("negative_zero", func(t *testing.T) {
		var v Int[int32]
		if err := v.Scan(math.Copysign(0, -1)); err != nil {
			t.Fatalf("Scan error: %v", err)
		}
		if v.Val() != 0 {
			t.Fatalf("Val = %d, want 0", v.Val())
		}
	})
	t.Run("fractional_rejected", func(t *testing.T) {
		var v Int[int32]
		err := v.Scan(float64(3.14))
		if err == nil {
			t.Fatalf("expected error for fractional float64, got nil")
		}
	})
	t.Run("tiny_fraction_rejected", func(t *testing.T) {
		var v Int[int32]
		err := v.Scan(1.0000000000001)
		if err == nil {
			t.Fatalf("expected error for near-integer float64, got nil")
		}
	})
	t.Run("overflow_int8", func(t *testing.T) {
		var v Int[int8]
		err := v.Scan(float64(200))
		if err == nil {
			t.Fatalf("expected error for float64 overflow into int8, got nil")
		}
	})
	t.Run("underflow_int8", func(t *testing.T) {
		var v Int[int8]
		err := v.Scan(float64(-200))
		if err == nil {
			t.Fatalf("expected error for float64 underflow into int8, got nil")
		}
	})
	t.Run("negative_into_unsigned", func(t *testing.T) {
		var v Int[uint32]
		err := v.Scan(float64(-1))
		if err == nil {
			t.Fatalf("expected error for negative float64 into unsigned, got nil")
		}
	})
}

func TestScan_Float64Boundaries(t *testing.T) {
	t.Run("signed", func(t *testing.T) {
		for _, tc := range []struct {
			name string
			src  float64
			want int64
			ok   bool
		}{
			{"min", -0x1p63, math.MinInt64, true},
			{"below_min", math.Nextafter(-0x1p63, math.Inf(-1)), 0, false},
			{"below_upper_bound", math.Nextafter(0x1p63, 0), math.MaxInt64 - 1023, true},
			{"upper_bound", 0x1p63, 0, false},
			{"above_upper_bound", math.Nextafter(0x1p63, math.Inf(1)), 0, false},
			{"nan", math.NaN(), 0, false},
			{"positive_infinity", math.Inf(1), 0, false},
			{"negative_infinity", math.Inf(-1), 0, false},
			{"fraction", 1.5, 0, false},
		} {
			t.Run(tc.name, func(t *testing.T) {
				v := New[int64](7)
				err := v.Scan(tc.src)
				if (err == nil) != tc.ok {
					t.Fatalf("Scan(%g): error = %v, want success = %v", tc.src, err, tc.ok)
				}
				want := tc.want
				if !tc.ok {
					want = 7 // Failed scans must preserve the previous value.
				}
				if v.Val() != want {
					t.Fatalf("Scan(%g): value = %d, want %d", tc.src, v.Val(), want)
				}
			})
		}
	})
	t.Run("unsigned", func(t *testing.T) {
		for _, tc := range []struct {
			name string
			src  float64
			want uint64
			ok   bool
		}{
			{"zero", 0, 0, true},
			{"negative_zero", math.Copysign(0, -1), 0, true},
			{"negative", -1, 0, false},
			{"signed_upper_bound", 0x1p63, 1 << 63, true},
			{"below_upper_bound", math.Nextafter(0x1p64, 0), math.MaxUint64 - 2047, true},
			{"upper_bound", 0x1p64, 0, false},
			{"nan", math.NaN(), 0, false},
			{"positive_infinity", math.Inf(1), 0, false},
			{"negative_infinity", math.Inf(-1), 0, false},
			{"fraction", 1.5, 0, false},
		} {
			t.Run(tc.name, func(t *testing.T) {
				v := New[uint64](7)
				err := v.Scan(tc.src)
				if (err == nil) != tc.ok {
					t.Fatalf("Scan(%g): error = %v, want success = %v", tc.src, err, tc.ok)
				}
				want := tc.want
				if !tc.ok {
					want = 7
				}
				if v.Val() != want {
					t.Fatalf("Scan(%g): value = %d, want %d", tc.src, v.Val(), want)
				}
			})
		}
	})
}

func TestScan_ByteSlice(t *testing.T) {
	t.Run("signed_positive", func(t *testing.T) {
		var v Int[int32]
		if err := v.Scan([]byte("12345")); err != nil {
			t.Fatalf("Scan error: %v", err)
		}
		if v.Val() != 12345 {
			t.Fatalf("Val = %d, want 12345", v.Val())
		}
	})
	t.Run("signed_negative", func(t *testing.T) {
		var v Int[int32]
		if err := v.Scan([]byte("-42")); err != nil {
			t.Fatalf("Scan error: %v", err)
		}
		if v.Val() != -42 {
			t.Fatalf("Val = %d, want -42", v.Val())
		}
	})
	t.Run("unsigned", func(t *testing.T) {
		var v Int[uint16]
		if err := v.Scan([]byte("65535")); err != nil {
			t.Fatalf("Scan error: %v", err)
		}
		if v.Val() != 65535 {
			t.Fatalf("Val = %d, want 65535", v.Val())
		}
	})
	t.Run("uint64_max", func(t *testing.T) {
		// go-sql-driver/mysql before 1.8 passes []byte for BIGINT UNSIGNED
		// in the text protocol.
		var v Int[uint64]
		if err := v.Scan([]byte("18446744073709551615")); err != nil {
			t.Fatalf("Scan error: %v", err)
		}
		if v.Val() != math.MaxUint64 {
			t.Fatalf("Val = %d, want MaxUint64", v.Val())
		}
	})
	t.Run("invalid", func(t *testing.T) {
		var v Int[int32]
		err := v.Scan([]byte("abc"))
		if err == nil {
			t.Fatalf("expected error for non-numeric []byte, got nil")
		}
	})
	t.Run("overflow", func(t *testing.T) {
		var v Int[int8]
		err := v.Scan([]byte("999"))
		if err == nil {
			t.Fatalf("expected error for overflow, got nil")
		}
	})
}

func TestScan_String(t *testing.T) {
	t.Run("signed_max", func(t *testing.T) {
		var v Int[int64]
		if err := v.Scan("9223372036854775807"); err != nil {
			t.Fatalf("Scan error: %v", err)
		}
		if v.Val() != math.MaxInt64 {
			t.Fatalf("Val = %d, want MaxInt64", v.Val())
		}
	})
	t.Run("signed_min", func(t *testing.T) {
		var v Int[int64]
		if err := v.Scan("-9223372036854775808"); err != nil {
			t.Fatalf("Scan error: %v", err)
		}
		if v.Val() != math.MinInt64 {
			t.Fatalf("Val = %d, want MinInt64", v.Val())
		}
	})
	t.Run("unsigned_max", func(t *testing.T) {
		var v Int[uint64]
		if err := v.Scan("18446744073709551615"); err != nil {
			t.Fatalf("Scan error: %v", err)
		}
		if v.Val() != math.MaxUint64 {
			t.Fatalf("Val = %d, want MaxUint64", v.Val())
		}
	})
	t.Run("signed_overflow", func(t *testing.T) {
		var v Int[int64]
		err := v.Scan("9223372036854775808") // MaxInt64 + 1
		if err == nil {
			t.Fatalf("expected error for int64 overflow, got nil")
		}
	})
	t.Run("unsigned_overflow", func(t *testing.T) {
		var v Int[uint64]
		err := v.Scan("18446744073709551616") // MaxUint64 + 1
		if err == nil {
			t.Fatalf("expected error for uint64 overflow, got nil")
		}
	})
	t.Run("negative_into_unsigned", func(t *testing.T) {
		var v Int[uint32]
		err := v.Scan("-1")
		if err == nil {
			t.Fatalf("expected error for negative into unsigned, got nil")
		}
	})
	t.Run("int8_boundary", func(t *testing.T) {
		var v Int[int8]
		if err := v.Scan("-128"); err != nil {
			t.Fatalf("Scan error: %v", err)
		}
		if v.Val() != -128 {
			t.Fatalf("Val = %d, want -128", v.Val())
		}
	})
	t.Run("int8_overflow_string", func(t *testing.T) {
		var v Int[int8]
		err := v.Scan("128")
		if err == nil {
			t.Fatalf("expected error for int8 overflow, got nil")
		}
	})
	t.Run("empty_string", func(t *testing.T) {
		var v Int[int32]
		err := v.Scan("")
		if err == nil {
			t.Fatalf("expected error for empty string, got nil")
		}
	})
	t.Run("whitespace", func(t *testing.T) {
		var v Int[int32]
		err := v.Scan(" 42 ")
		if err == nil {
			t.Fatalf("expected error for string with whitespace, got nil")
		}
	})
	t.Run("hex_rejected", func(t *testing.T) {
		var v Int[int32]
		err := v.Scan("0xff")
		if err == nil {
			t.Fatalf("expected error for hex string, got nil")
		}
	})
}

func TestScan_UnsupportedType(t *testing.T) {
	types := []interface{}{true, complex(1, 2), struct{}{}, []int{1}}
	for _, src := range types {
		var v Int[int32]
		err := v.Scan(src)
		if err == nil {
			t.Fatalf("expected error for unsupported type %T, got nil", src)
		}
	}
}

// ---------------------------------------------------------------------------
// SQL round-trip: Value then Scan
// ---------------------------------------------------------------------------

func TestSQL_RoundTrip(t *testing.T) {
	t.Run("int32", func(t *testing.T) {
		tests := []struct {
			name string
			orig Int[int32]
		}{
			{"positive", New[int32](42)},
			{"negative", New[int32](-42)},
			{"zero", New[int32](0)},
			{"max", New[int32](math.MaxInt32)},
			{"min", New[int32](math.MinInt32)},
		}
		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				dv, err := tc.orig.Value()
				if err != nil {
					t.Fatalf("Value error: %v", err)
				}
				var got Int[int32]
				if err := got.Scan(dv); err != nil {
					t.Fatalf("Scan error: %v", err)
				}
				if got.Val() != tc.orig.Val() {
					t.Fatalf("round-trip: got %d, want %d", got.Val(), tc.orig.Val())
				}
			})
		}
	})
	t.Run("int64", func(t *testing.T) {
		for _, val := range []int64{0, math.MaxInt64, math.MinInt64, 1, -1} {
			orig := New[int64](val)
			dv, err := orig.Value()
			if err != nil {
				t.Fatalf("Value(%d) error: %v", val, err)
			}
			var got Int[int64]
			if err := got.Scan(dv); err != nil {
				t.Fatalf("Scan(%d) error: %v", val, err)
			}
			if got.Val() != val {
				t.Fatalf("round-trip: got %d, want %d", got.Val(), val)
			}
		}
	})
	// uint64 > MaxInt64: Value() returns string, Scan() accepts string
	t.Run("uint64_large_via_string", func(t *testing.T) {
		for _, val := range []uint64{uint64(math.MaxInt64) + 1, math.MaxUint64} {
			orig := New[uint64](val)
			dv, err := orig.Value()
			if err != nil {
				t.Fatalf("Value(%d) error: %v", val, err)
			}
			// Value should have returned a string
			s, ok := dv.(string)
			if !ok {
				t.Fatalf("Value(%d) type = %T, want string", val, dv)
			}
			var got Int[uint64]
			if err := got.Scan(s); err != nil {
				t.Fatalf("Scan(%q) error: %v", s, err)
			}
			if got.Val() != val {
				t.Fatalf("round-trip: got %d, want %d", got.Val(), val)
			}
		}
	})
	// uint64 <= MaxInt64: Value() returns int64, Scan() accepts int64
	t.Run("uint64_small_via_int64", func(t *testing.T) {
		for _, val := range []uint64{0, 1, uint64(math.MaxInt64)} {
			orig := New[uint64](val)
			dv, err := orig.Value()
			if err != nil {
				t.Fatalf("Value(%d) error: %v", val, err)
			}
			if _, ok := dv.(int64); !ok {
				t.Fatalf("Value(%d) type = %T, want int64", val, dv)
			}
			var got Int[uint64]
			if err := got.Scan(dv); err != nil {
				t.Fatalf("Scan(%v) error: %v", dv, err)
			}
			if got.Val() != val {
				t.Fatalf("round-trip: got %d, want %d", got.Val(), val)
			}
		}
	})
}

// TestUnmarshalJSONFastPathAllocs guards the UnmarshalJSON comment that the
// canonical-integer fast path does not allocate.
func TestUnmarshalJSONFastPathAllocs(t *testing.T) {
	data := []byte("-1234567")
	var v Int[int64]
	allocs := testing.AllocsPerRun(100, func() {
		if err := v.UnmarshalJSON(data); err != nil {
			t.Fatal(err)
		}
	})
	if allocs != 0 {
		t.Fatalf("UnmarshalJSON allocated %v times per call, want 0", allocs)
	}
}

func BenchmarkUnmarshalJSON(b *testing.B) {
	data := []byte(`[1,-22,333,-4444,55555,-666666,7777777,-88888888,999999999,0]`)
	b.Run("Int[int64]", func(b *testing.B) {
		var v []Int[int64]
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			if err := json.Unmarshal(data, &v); err != nil {
				b.Fatal(err)
			}
		}
	})
	b.Run("native_int64", func(b *testing.B) {
		var v []int64
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			if err := json.Unmarshal(data, &v); err != nil {
				b.Fatal(err)
			}
		}
	})
}

// FuzzUnmarshalJSONFastPath checks that UnmarshalJSON, including its fast
// path, gives the same value and error as the general decodeJSONNumber path.
func FuzzUnmarshalJSONFastPath(f *testing.F) {
	for _, s := range []string{
		"0", "-0", "1", "-1", "00", "01", "-01", "-", "+1", " 1", "1 ", "1.0", "1e2",
		"127", "128", "-128", "-129", "255", "256",
		"9223372036854775807", "9223372036854775808",
		"-9223372036854775808", "-9223372036854775809",
		"18446744073709551615", "18446744073709551616",
		"null", `"1"`, "", "1a", "true", "[]",
	} {
		f.Add([]byte(s))
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		checkJSONFastPath[int8](t, data)
		checkJSONFastPath[uint8](t, data)
		checkJSONFastPath[int64](t, data)
		checkJSONFastPath[uint64](t, data)
		checkJSONFastPath[jsonQuotedUint](t, data)
	})
}

func checkJSONFastPath[T Integer](t *testing.T, data []byte) {
	t.Helper()
	got, want := New(T(7)), New(T(7))
	gotErr := got.UnmarshalJSON(data)
	wantErr := want.decodeJSONNumber(data)
	if got != want || !reflect.DeepEqual(gotErr, wantErr) {
		t.Fatalf("%T input %q: UnmarshalJSON = (%d, %v), decodeJSONNumber = (%d, %v)",
			got, data, got.Val(), gotErr, want.Val(), wantErr)
	}
}

// TestSQL_NullRoundTrip guards the documented way to keep SQL NULL distinct:
// sql.Null[Int[T]] must both write (through Int[T].Value) and scan.
func TestSQL_NullRoundTrip(t *testing.T) {
	v, err := sql.Null[Int[int64]]{V: New[int64](5), Valid: true}.Value()
	if err != nil || v != int64(5) {
		t.Fatalf("Value = (%#v, %v), want (int64(5), nil)", v, err)
	}
	// Above MaxInt64, Int[T].Value returns a decimal string.
	v, err = sql.Null[Int[uint64]]{V: New[uint64](math.MaxUint64), Valid: true}.Value()
	if err != nil || v != "18446744073709551615" {
		t.Fatalf("Value = (%#v, %v), want (\"18446744073709551615\", nil)", v, err)
	}
	var n sql.Null[Int[int64]]
	if err := n.Scan(int64(7)); err != nil || !n.Valid || n.V.Val() != 7 {
		t.Fatalf("Scan(7) = %v, got %+v", err, n)
	}
	if err := n.Scan(nil); err != nil || n.Valid {
		t.Fatalf("Scan(nil) = %v, got %+v", err, n)
	}
}
