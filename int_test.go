package safeint

import (
	"database/sql/driver"
	"encoding/json"
	"math"
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
			if ok && got.Val() != tc.want {
				t.Fatalf("result = %d, want %d", got.Val(), tc.want)
			}
		})
	}
}

func TestAddOverflowUnsigned(t *testing.T) {
	tests := []struct {
		name string
		a, b Int[uint8]
		want uint8
		ok   bool
	}{
		{"normal", New[uint8](100), New[uint8](50), 150, true},
		{"max_boundary", New[uint8](254), New[uint8](1), 255, true},
		{"overflow", New[uint8](255), New[uint8](1), 0, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := tc.a.AddOverflow(tc.b)
			if ok != tc.ok {
				t.Fatalf("ok = %v, want %v", ok, tc.ok)
			}
			if ok && got.Val() != tc.want {
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
			if ok && got.Val() != tc.want {
				t.Fatalf("result = %d, want %d", got.Val(), tc.want)
			}
		})
	}
}

func TestSubOverflowUnsigned(t *testing.T) {
	tests := []struct {
		name string
		a, b Int[uint8]
		want uint8
		ok   bool
	}{
		{"normal", New[uint8](50), New[uint8](30), 20, true},
		{"zero_result", New[uint8](10), New[uint8](10), 0, true},
		{"underflow", New[uint8](0), New[uint8](1), 255, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := tc.a.SubOverflow(tc.b)
			if ok != tc.ok {
				t.Fatalf("ok = %v, want %v", ok, tc.ok)
			}
			if ok && got.Val() != tc.want {
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
			if ok && got.Val() != tc.want {
				t.Fatalf("result = %d, want %d", got.Val(), tc.want)
			}
		})
	}
}

func TestMulOverflowUnsigned(t *testing.T) {
	tests := []struct {
		name string
		a, b Int[uint8]
		want uint8
		ok   bool
	}{
		{"normal", New[uint8](10), New[uint8](25), 250, true},
		{"overflow", New[uint8](16), New[uint8](16), 0, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := tc.a.MulOverflow(tc.b)
			if ok != tc.ok {
				t.Fatalf("ok = %v, want %v", ok, tc.ok)
			}
			if ok && got.Val() != tc.want {
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
		{"min_div_neg1", New[int8](-128), New[int8](-1), -128, false},
		{"div_by_zero", New[int8](1), New[int8](0), 0, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := tc.a.DivOverflow(tc.b)
			if ok != tc.ok {
				t.Fatalf("ok = %v, want %v", ok, tc.ok)
			}
			if ok && got.Val() != tc.want {
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
		{"min_div_neg1", New[int8](-128), New[int8](-1), -128, 0, false},
		{"div_by_zero", New[int8](10), New[int8](0), 0, 0, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			q, r, ok := tc.a.DivModOverflow(tc.b)
			if ok != tc.ok {
				t.Fatalf("ok = %v, want %v", ok, tc.ok)
			}
			if ok {
				if q.Val() != tc.wantQ {
					t.Fatalf("quotient = %d, want %d", q.Val(), tc.wantQ)
				}
				if r.Val() != tc.wantR {
					t.Fatalf("remainder = %d, want %d", r.Val(), tc.wantR)
				}
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
			if ok && got.Val() != tc.want {
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
			if ok && got.Val() != tc.want {
				t.Fatalf("result = %d, want %d", got.Val(), tc.want)
			}
		})
	}
}

func TestNegOverflowUnsigned(t *testing.T) {
	tests := []struct {
		name string
		a    Int[uint8]
		ok   bool
	}{
		{"zero_ok", New[uint8](0), true},
		{"nonzero_overflow", New[uint8](1), false},
		{"max_overflow", New[uint8](255), false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, ok := tc.a.NegOverflow()
			if ok != tc.ok {
				t.Fatalf("ok = %v, want %v", ok, tc.ok)
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
			if ok && got.Val() != tc.want {
				t.Fatalf("result = %d, want %d", got.Val(), tc.want)
			}
		})
	}
}

func TestAbsOverflowUnsigned(t *testing.T) {
	tests := []struct {
		name string
		a    Int[uint8]
		want uint8
	}{
		{"zero", New[uint8](0), 0},
		{"normal", New[uint8](200), 200},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := tc.a.AbsOverflow()
			if !ok {
				t.Fatalf("ok = false, want true")
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
			if ok && got.Val() != tc.want {
				t.Fatalf("result = %d, want %d", got.Val(), tc.want)
			}
		})
	}
}

func TestPowOverflowSigned(t *testing.T) {
	tests := []struct {
		name string
		a    Int[int8]
		exp  uint
		want int8
		ok   bool
	}{
		{"neg2^7", New[int8](-2), 7, -128, true},
		{"neg2^6", New[int8](-2), 6, 64, true},
		{"neg1^99", New[int8](-1), 99, -1, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := tc.a.PowOverflow(tc.exp)
			if ok != tc.ok {
				t.Fatalf("ok = %v, want %v", ok, tc.ok)
			}
			if ok && got.Val() != tc.want {
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
			if ok && got.Val() != tc.want {
				t.Fatalf("result = %d, want %d", got.Val(), tc.want)
			}
		})
	}
}

func TestLshOverflowUnsigned(t *testing.T) {
	tests := []struct {
		name string
		a    Int[uint8]
		n    uint
		want uint8
		ok   bool
	}{
		{"normal", New[uint8](1), 7, 128, true},
		{"overflow", New[uint8](1), 8, 0, false},
		{"bits_lost", New[uint8](3), 7, 128, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := tc.a.LshOverflow(tc.n)
			if ok != tc.ok {
				t.Fatalf("ok = %v, want %v", ok, tc.ok)
			}
			if ok && got.Val() != tc.want {
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
			if ok && got.Val() != tc.want {
				t.Fatalf("result = %d, want %d", got.Val(), tc.want)
			}
		})
	}
}

func TestMulDivOverflowSmall(t *testing.T) {
	tests := []struct {
		name    string
		a, b, c Int[int8]
		want    int8
		ok      bool
	}{
		{
			"intermediate_would_overflow",
			New[int8](100), New[int8](100), New[int8](100),
			100, true,
		},
		{
			"result_overflows_int8",
			New[int8](100), New[int8](100), New[int8](1),
			0, false,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := tc.a.MulDivOverflow(tc.b, tc.c)
			if ok != tc.ok {
				t.Fatalf("ok = %v, want %v", ok, tc.ok)
			}
			if ok && got.Val() != tc.want {
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
		{
			"large_intermediate",
			New[int64](math.MaxInt64), New[int64](math.MaxInt64), New[int64](10),
			9, true, // MaxInt64 mod 10 = 7, 7*7 mod 10 = 9
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := tc.a.MulModOverflow(tc.b, tc.c)
			if ok != tc.ok {
				t.Fatalf("ok = %v, want %v", ok, tc.ok)
			}
			if ok && got.Val() != tc.want {
				t.Fatalf("result = %d, want %d", got.Val(), tc.want)
			}
		})
	}
}

func TestMulModOverflowSmall(t *testing.T) {
	// int8: (7*5)%3 = 35%3 = 2
	got, ok := New[int8](7).MulModOverflow(New[int8](5), New[int8](3))
	if !ok {
		t.Fatalf("ok = false, want true")
	}
	if got.Val() != 2 {
		t.Fatalf("result = %d, want 2", got.Val())
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
		expectPanic(t, "MustAdd overflow", func() {
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
		expectPanic(t, "MustSub overflow", func() {
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
		expectPanic(t, "MustMul overflow", func() {
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
		expectPanic(t, "MustDiv div-by-zero", func() {
			New[int8](1).MustDiv(New[int8](0))
		})
	})
	t.Run("overflow_panics", func(t *testing.T) {
		expectPanic(t, "MustDiv overflow", func() {
			New[int8](-128).MustDiv(New[int8](-1))
		})
	})
}

// ---------------------------------------------------------------------------
// 5. Comparison tests
// ---------------------------------------------------------------------------

func TestCmp(t *testing.T) {
	tests := []struct {
		name string
		a, b Int[int64]
		want int
	}{
		{"less", New[int64](1), New[int64](2), -1},
		{"equal", New[int64](42), New[int64](42), 0},
		{"greater", New[int64](10), New[int64](5), 1},
		{"negative_less", New[int64](-5), New[int64](3), -1},
		{"negative_greater", New[int64](3), New[int64](-5), 1},
		{"both_negative", New[int64](-10), New[int64](-5), -1},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.a.Cmp(tc.b)
			if got != tc.want {
				t.Fatalf("Cmp = %d, want %d", got, tc.want)
			}
		})
	}
}

func TestCmpUnsigned(t *testing.T) {
	tests := []struct {
		name string
		a, b Int[uint8]
		want int
	}{
		{"less", New[uint8](0), New[uint8](255), -1},
		{"equal", New[uint8](128), New[uint8](128), 0},
		{"greater", New[uint8](255), New[uint8](0), 1},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.a.Cmp(tc.b)
			if got != tc.want {
				t.Fatalf("Cmp = %d, want %d", got, tc.want)
			}
		})
	}
}

func TestEq(t *testing.T) {
	tests := []struct {
		name string
		a, b Int[int64]
		want bool
	}{
		{"equal", New[int64](42), New[int64](42), true},
		{"not_equal", New[int64](42), New[int64](43), false},
		{"neg_equal", New[int64](-7), New[int64](-7), true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.a.Eq(tc.b); got != tc.want {
				t.Fatalf("Eq = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestLt(t *testing.T) {
	tests := []struct {
		name string
		a, b Int[int64]
		want bool
	}{
		{"true", New[int64](1), New[int64](2), true},
		{"false_equal", New[int64](2), New[int64](2), false},
		{"false_greater", New[int64](3), New[int64](2), false},
		{"negative_true", New[int64](-5), New[int64](3), true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.a.Lt(tc.b); got != tc.want {
				t.Fatalf("Lt = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestGt(t *testing.T) {
	tests := []struct {
		name string
		a, b Int[int64]
		want bool
	}{
		{"true", New[int64](3), New[int64](2), true},
		{"false_equal", New[int64](2), New[int64](2), false},
		{"false_less", New[int64](1), New[int64](2), false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.a.Gt(tc.b); got != tc.want {
				t.Fatalf("Gt = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestLte(t *testing.T) {
	tests := []struct {
		name string
		a, b Int[int64]
		want bool
	}{
		{"less", New[int64](1), New[int64](2), true},
		{"equal", New[int64](2), New[int64](2), true},
		{"greater", New[int64](3), New[int64](2), false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.a.Lte(tc.b); got != tc.want {
				t.Fatalf("Lte = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestGte(t *testing.T) {
	tests := []struct {
		name string
		a, b Int[int64]
		want bool
	}{
		{"greater", New[int64](3), New[int64](2), true},
		{"equal", New[int64](2), New[int64](2), true},
		{"less", New[int64](1), New[int64](2), false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.a.Gte(tc.b); got != tc.want {
				t.Fatalf("Gte = %v, want %v", got, tc.want)
			}
		})
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

func TestIsZeroUnsigned(t *testing.T) {
	tests := []struct {
		name string
		a    Int[uint8]
		want bool
	}{
		{"zero", New[uint8](0), true},
		{"nonzero", New[uint8](1), false},
		{"max", New[uint8](255), false},
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
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.got != tc.val {
				t.Fatalf("String() = %q, want %q", tc.got, tc.val)
			}
		})
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

func TestUnmarshalJSON(t *testing.T) {
	t.Run("int32_positive", func(t *testing.T) {
		var v Int[int32]
		if err := json.Unmarshal([]byte("42"), &v); err != nil {
			t.Fatalf("UnmarshalJSON error: %v", err)
		}
		if v.Val() != 42 {
			t.Fatalf("Val = %d, want 42", v.Val())
		}
	})
	t.Run("int32_negative", func(t *testing.T) {
		var v Int[int32]
		if err := json.Unmarshal([]byte("-100"), &v); err != nil {
			t.Fatalf("UnmarshalJSON error: %v", err)
		}
		if v.Val() != -100 {
			t.Fatalf("Val = %d, want -100", v.Val())
		}
	})
	t.Run("int32_zero", func(t *testing.T) {
		var v Int[int32]
		if err := json.Unmarshal([]byte("0"), &v); err != nil {
			t.Fatalf("UnmarshalJSON error: %v", err)
		}
		if v.Val() != 0 {
			t.Fatalf("Val = %d, want 0", v.Val())
		}
	})
	// --- boundary: exact min/max accepted ---
	t.Run("int8_exact_max", func(t *testing.T) {
		var v Int[int8]
		if err := json.Unmarshal([]byte("127"), &v); err != nil {
			t.Fatalf("UnmarshalJSON error: %v", err)
		}
		if v.Val() != 127 {
			t.Fatalf("Val = %d, want 127", v.Val())
		}
	})
	t.Run("int8_exact_min", func(t *testing.T) {
		var v Int[int8]
		if err := json.Unmarshal([]byte("-128"), &v); err != nil {
			t.Fatalf("UnmarshalJSON error: %v", err)
		}
		if v.Val() != -128 {
			t.Fatalf("Val = %d, want -128", v.Val())
		}
	})
	t.Run("uint8_exact_max", func(t *testing.T) {
		var v Int[uint8]
		if err := json.Unmarshal([]byte("255"), &v); err != nil {
			t.Fatalf("UnmarshalJSON error: %v", err)
		}
		if v.Val() != 255 {
			t.Fatalf("Val = %d, want 255", v.Val())
		}
	})
	t.Run("uint8_exact_zero", func(t *testing.T) {
		var v Int[uint8]
		if err := json.Unmarshal([]byte("0"), &v); err != nil {
			t.Fatalf("UnmarshalJSON error: %v", err)
		}
		if v.Val() != 0 {
			t.Fatalf("Val = %d, want 0", v.Val())
		}
	})
	// --- boundary: one past min/max rejected ---
	t.Run("uint8_overflow", func(t *testing.T) {
		var v Int[uint8]
		err := json.Unmarshal([]byte("256"), &v)
		if err == nil {
			t.Fatalf("expected error for uint8 overflow, got nil")
		}
	})
	t.Run("int8_overflow", func(t *testing.T) {
		var v Int[int8]
		err := json.Unmarshal([]byte("128"), &v)
		if err == nil {
			t.Fatalf("expected error for int8 overflow, got nil")
		}
	})
	t.Run("int8_underflow", func(t *testing.T) {
		var v Int[int8]
		err := json.Unmarshal([]byte("-129"), &v)
		if err == nil {
			t.Fatalf("expected error for int8 underflow, got nil")
		}
	})
	t.Run("uint8_negative", func(t *testing.T) {
		var v Int[uint8]
		err := json.Unmarshal([]byte("-1"), &v)
		if err == nil {
			t.Fatalf("expected error for negative into uint8, got nil")
		}
	})
	// --- invalid JSON types ---
	t.Run("invalid_json_string", func(t *testing.T) {
		var v Int[int32]
		err := json.Unmarshal([]byte(`"hello"`), &v)
		if err == nil {
			t.Fatalf("expected error for string input, got nil")
		}
	})
	t.Run("invalid_json_bool", func(t *testing.T) {
		var v Int[int32]
		err := json.Unmarshal([]byte("true"), &v)
		if err == nil {
			t.Fatalf("expected error for bool input, got nil")
		}
	})
	t.Run("invalid_json_array", func(t *testing.T) {
		var v Int[int32]
		err := json.Unmarshal([]byte("[1]"), &v)
		if err == nil {
			t.Fatalf("expected error for array input, got nil")
		}
	})
	t.Run("invalid_json_object", func(t *testing.T) {
		var v Int[int32]
		err := json.Unmarshal([]byte(`{"x":1}`), &v)
		if err == nil {
			t.Fatalf("expected error for object input, got nil")
		}
	})
	t.Run("invalid_json_syntax", func(t *testing.T) {
		var v Int[int32]
		err := json.Unmarshal([]byte("{bad"), &v)
		if err == nil {
			t.Fatalf("expected error for invalid JSON, got nil")
		}
	})
	t.Run("invalid_json_empty", func(t *testing.T) {
		var v Int[int32]
		err := json.Unmarshal([]byte(""), &v)
		if err == nil {
			t.Fatalf("expected error for empty input, got nil")
		}
	})
	t.Run("json_float_literal", func(t *testing.T) {
		// JSON "1.0" is a valid JSON number but not a valid integer
		var v Int[int32]
		err := json.Unmarshal([]byte("1.0"), &v)
		if err == nil {
			t.Fatalf("expected error for float literal, got nil")
		}
	})
}

func TestJSON_RoundTrip(t *testing.T) {
	t.Run("int32", func(t *testing.T) {
		orig := New[int32](-12345)
		data, err := json.Marshal(orig)
		if err != nil {
			t.Fatalf("Marshal error: %v", err)
		}
		var got Int[int32]
		if err := json.Unmarshal(data, &got); err != nil {
			t.Fatalf("Unmarshal error: %v", err)
		}
		if got.Val() != orig.Val() {
			t.Fatalf("round-trip: got %d, want %d", got.Val(), orig.Val())
		}
	})
	t.Run("int8_min", func(t *testing.T) {
		orig := New[int8](math.MinInt8)
		data, err := json.Marshal(orig)
		if err != nil {
			t.Fatalf("Marshal error: %v", err)
		}
		var got Int[int8]
		if err := json.Unmarshal(data, &got); err != nil {
			t.Fatalf("Unmarshal error: %v", err)
		}
		if got.Val() != orig.Val() {
			t.Fatalf("round-trip: got %d, want %d", got.Val(), orig.Val())
		}
	})
	t.Run("int64_min", func(t *testing.T) {
		orig := New[int64](math.MinInt64)
		data, err := json.Marshal(orig)
		if err != nil {
			t.Fatalf("Marshal error: %v", err)
		}
		var got Int[int64]
		if err := json.Unmarshal(data, &got); err != nil {
			t.Fatalf("Unmarshal error: %v", err)
		}
		if got.Val() != orig.Val() {
			t.Fatalf("round-trip: got %d, want %d", got.Val(), orig.Val())
		}
	})
	t.Run("int64_max", func(t *testing.T) {
		orig := New[int64](math.MaxInt64)
		data, err := json.Marshal(orig)
		if err != nil {
			t.Fatalf("Marshal error: %v", err)
		}
		var got Int[int64]
		if err := json.Unmarshal(data, &got); err != nil {
			t.Fatalf("Unmarshal error: %v", err)
		}
		if got.Val() != orig.Val() {
			t.Fatalf("round-trip: got %d, want %d", got.Val(), orig.Val())
		}
	})
	t.Run("uint64_max", func(t *testing.T) {
		orig := New[uint64](math.MaxUint64)
		data, err := json.Marshal(orig)
		if err != nil {
			t.Fatalf("Marshal error: %v", err)
		}
		var got Int[uint64]
		if err := json.Unmarshal(data, &got); err != nil {
			t.Fatalf("Unmarshal error: %v", err)
		}
		if got.Val() != orig.Val() {
			t.Fatalf("round-trip: got %d, want %d", got.Val(), orig.Val())
		}
	})
	t.Run("uint64_max_int64_boundary", func(t *testing.T) {
		orig := New[uint64](uint64(math.MaxInt64) + 1)
		data, err := json.Marshal(orig)
		if err != nil {
			t.Fatalf("Marshal error: %v", err)
		}
		var got Int[uint64]
		if err := json.Unmarshal(data, &got); err != nil {
			t.Fatalf("Unmarshal error: %v", err)
		}
		if got.Val() != orig.Val() {
			t.Fatalf("round-trip: got %d, want %d", got.Val(), orig.Val())
		}
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

func TestScan_Uint64_Int64Source(t *testing.T) {
	var v Int[uint64]
	err := v.Scan(int64(-1))
	if err == nil {
		t.Fatalf("expected error for negative int64 into uint64, got nil")
	}
}

func TestScan_PreservesZeroOnNil(t *testing.T) {
	// Scan(nil) should reset to zero even if previously set
	v := New[int32](12345)
	if err := v.Scan(nil); err != nil {
		t.Fatalf("Scan(nil) error: %v", err)
	}
	if v.Val() != 0 {
		t.Fatalf("Val = %d, want 0", v.Val())
	}
}

func TestScan_ByteSlice_Uint64(t *testing.T) {
	// []byte is common from MySQL for BIGINT UNSIGNED
	var v Int[uint64]
	if err := v.Scan([]byte("18446744073709551615")); err != nil {
		t.Fatalf("Scan error: %v", err)
	}
	if v.Val() != math.MaxUint64 {
		t.Fatalf("Val = %d, want MaxUint64", v.Val())
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
