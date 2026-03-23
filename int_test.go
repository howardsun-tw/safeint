package safeint

import (
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
