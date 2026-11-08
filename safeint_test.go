package safeint

import (
	"fmt"
	"math"
	"math/big"
	"testing"
	"unsafe"
)

// ---------------------------------------------------------------------------
// A. Exhaustive 8-bit tests
// ---------------------------------------------------------------------------

// Exponents and shift counts for the 8-bit Pow and Lsh checks. They include
// values where the result first overflows (8, 16) and values of at least 64,
// which run the whole binary-exponentiation loop.
var (
	exhaustivePowExps   = []uint{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 31, 32, 63, 64, 65, 1000, math.MaxUint}
	exhaustiveLshShifts = []uint{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 15, 16, 63, 64, 65, math.MaxUint}
)

func TestExhaustive8Bit(t *testing.T) {
	t.Run("int8", exhaustive8[int8])
	t.Run("uint8", exhaustive8[uint8])
}

// exhaustive8 checks every operation for every operand of the 8-bit type T
// against exact int64 arithmetic, and math/big for Pow's wrapped value. On
// failure it expects each function's documented result: Add, Sub, Mul, Neg,
// Abs and Pow return the wrapped value; Div, DivMod, Mod, MulDiv and MulMod
// return zero; Lsh returns 0 when the shift count is at least 8. The value
// that Lsh returns for other overflows, and that Convert returns on failure,
// is not specified, so only ok is checked there.
func exhaustive8[T int8 | uint8](t *testing.T) {
	lo, hi := int64(0), int64(math.MaxUint8)
	if T(0)-1 < 0 {
		lo, hi = math.MinInt8, math.MaxInt8
	}
	// want returns (exact, true) when exact fits in T. Otherwise it returns
	// exact wrapped to T, or zero, with false.
	want := func(exact int64, wrap bool) (T, bool) {
		if exact >= lo && exact <= hi {
			return T(exact), true
		}
		if wrap {
			return T(exact), false
		}
		return 0, false
	}
	expect := func(op string, args [3]int64, got T, ok bool, w T, wok bool) {
		if got != w || ok != wok {
			t.Helper() // only on failure: Helper is too slow for the hot loop
			t.Fatalf("%s%v = (%d, %v), want (%d, %v)", op, args, got, ok, w, wok)
		}
	}

	checkConvert8[T, int8](t)
	checkConvert8[T, uint8](t)
	checkConvert8[T, int16](t)
	checkConvert8[T, uint16](t)
	checkConvert8[T, int32](t)
	checkConvert8[T, uint32](t)
	checkConvert8[T, int64](t)
	checkConvert8[T, uint64](t)
	checkConvert8[T, int](t)
	checkConvert8[T, uint](t)

	for i := range 256 {
		a := T(i)
		x := int64(a)

		r, ok := Neg(a)
		w, wok := want(-x, true)
		expect("Neg", [3]int64{x}, r, ok, w, wok)

		r, ok = Abs(a)
		w, wok = want(max(x, -x), true)
		expect("Abs", [3]int64{x}, r, ok, w, wok)

		for _, e := range exhaustivePowExps {
			r, ok = Pow(a, e)
			wrapped := new(big.Int).Exp(big.NewInt(x&0xFF), new(big.Int).SetUint64(uint64(e)), big.NewInt(256))
			expect("Pow", [3]int64{x, int64(e)}, r, ok, T(wrapped.Int64()), powFits(x, e, lo, hi))
		}

		for _, n := range exhaustiveLshShifts {
			r, ok = Lsh(a, n)
			switch {
			case a == 0:
				expect("Lsh", [3]int64{x, int64(n)}, r, ok, 0, true)
			case n >= 8:
				expect("Lsh", [3]int64{x, int64(n)}, r, ok, 0, false)
			default:
				if w, wok = want(x<<n, false); wok {
					expect("Lsh", [3]int64{x, int64(n)}, r, ok, w, true)
				} else if ok {
					t.Fatalf("Lsh(%d, %d) = (%d, true), want ok == false", x, n, r)
				}
			}
		}

		for j := range 256 {
			b := T(j)
			y := int64(b)

			r, ok = Add(a, b)
			w, wok = want(x+y, true)
			expect("Add", [3]int64{x, y}, r, ok, w, wok)

			r, ok = Sub(a, b)
			w, wok = want(x-y, true)
			expect("Sub", [3]int64{x, y}, r, ok, w, wok)

			r, ok = Mul(a, b)
			w, wok = want(x*y, true)
			expect("Mul", [3]int64{x, y}, r, ok, w, wok)

			var q, rem T
			var qok bool
			if y != 0 {
				q, qok = want(x/y, false)
				rem = T(x % y)
			}
			r, ok = Div(a, b)
			expect("Div", [3]int64{x, y}, r, ok, q, qok)

			r, ok = Mod(a, b)
			expect("Mod", [3]int64{x, y}, r, ok, rem, y != 0)

			if !qok {
				rem = 0
			}
			gq, gr, ok := DivMod(a, b)
			if gq != q || gr != rem || ok != qok {
				t.Fatalf("DivMod(%d, %d) = (%d, %d, %v), want (%d, %d, %v)", x, y, gq, gr, ok, q, rem, qok)
			}

			for k := range 256 {
				c := T(k)
				z := int64(c)
				var dw, mw T
				var dok bool
				if z != 0 {
					dw, dok = want(x*y/z, false)
					mw = T(x * y % z)
				}
				r, ok = MulDiv(a, b, c)
				expect("MulDiv", [3]int64{x, y, z}, r, ok, dw, dok)
				r, ok = MulMod(a, b, c)
				expect("MulMod", [3]int64{x, y, z}, r, ok, mw, z != 0)
			}
		}
	}
}

// powFits reports whether x^e lies in [lo, hi] for an 8-bit x. Powers of
// 0, 1 and -1 never grow, and for |x| >= 2 every e >= 8 exceeds 8 bits.
func powFits(x int64, e uint, lo, hi int64) bool {
	switch {
	case x >= -1 && x <= 1:
		return true
	case e >= 8:
		return false
	}
	p := int64(1)
	for range e {
		p *= x
	}
	return p >= lo && p <= hi
}

// checkConvert8 checks Convert from every value of the 8-bit type T to U.
func checkConvert8[T int8 | uint8, U Integer](t *testing.T) {
	t.Helper()
	lo, hi := boundsOf[U]()
	for i := range 256 {
		a := T(i)
		exact := big.NewInt(int64(a))
		got, ok := Convert[T, U](a)
		wok := exact.Cmp(lo) >= 0 && exact.Cmp(hi) <= 0
		if ok != wok || (ok && toBig(got).Cmp(exact) != 0) {
			t.Fatalf("Convert[%T, %T](%d) = (%d, %v), want ok == %v", a, got, a, got, ok, wok)
		}
	}
}

// ---------------------------------------------------------------------------
// B. Boundary value pair tests
// ---------------------------------------------------------------------------

// boundsOf returns the minimum and maximum values of T.
func boundsOf[T Integer]() (lo, hi *big.Int) {
	bits := uint(unsafe.Sizeof(T(0))) * 8
	one := big.NewInt(1)
	if T(0)-1 < 0 {
		hi = new(big.Int).Sub(new(big.Int).Lsh(one, bits-1), one)
		lo = new(big.Int).Neg(new(big.Int).Lsh(one, bits-1))
		return lo, hi
	}
	return new(big.Int), new(big.Int).Sub(new(big.Int).Lsh(one, bits), one)
}

func toBig[T Integer](v T) *big.Int {
	if v < 0 {
		return big.NewInt(int64(v))
	}
	return new(big.Int).SetUint64(uint64(v))
}

// wrapBig returns v modulo 2^64, converted to T. This is the two's
// complement value that native arithmetic produces.
func wrapBig[T Integer](v *big.Int) T {
	return T(new(big.Int).And(v, new(big.Int).SetUint64(math.MaxUint64)).Uint64())
}

// checkBinaryOp compares op(a, b) with exact math/big arithmetic. In range
// it wants (exact, true). Out of range, Add, Sub and Mul want the wrapped
// result and Div wants 0, all with false.
func checkBinaryOp[T Integer](t *testing.T, opName string, op func(T, T) (T, bool), a, b T) {
	t.Helper()
	got, ok := op(a, b)
	var want T
	var wantOk bool
	ba, bb := toBig(a), toBig(b)
	if opName == "Div" {
		if b != 0 {
			exact := new(big.Int).Quo(ba, bb)
			lo, hi := boundsOf[T]()
			if exact.Cmp(lo) >= 0 && exact.Cmp(hi) <= 0 {
				want, wantOk = wrapBig[T](exact), true
			}
		}
	} else {
		exact := new(big.Int)
		switch opName {
		case "Add":
			exact.Add(ba, bb)
		case "Sub":
			exact.Sub(ba, bb)
		case "Mul":
			exact.Mul(ba, bb)
		default:
			t.Fatalf("unknown op: %s", opName)
		}
		lo, hi := boundsOf[T]()
		want, wantOk = wrapBig[T](exact), exact.Cmp(lo) >= 0 && exact.Cmp(hi) <= 0
	}
	if got != want || ok != wantOk {
		t.Errorf("%s(%d, %d) = (%d, %v), want (%d, %v)", opName, a, b, got, ok, want, wantOk)
	}
}

// checkBoundaryPairs runs Add, Sub, Mul and Div on each pair in both
// operand orders.
func checkBoundaryPairs[T Integer](t *testing.T, pairs [][2]T) {
	ops := []struct {
		name string
		fn   func(T, T) (T, bool)
	}{
		{"Add", Add[T]}, {"Sub", Sub[T]}, {"Mul", Mul[T]}, {"Div", Div[T]},
	}
	for _, op := range ops {
		for _, p := range pairs {
			checkBinaryOp(t, op.name, op.fn, p[0], p[1])
			checkBinaryOp(t, op.name, op.fn, p[1], p[0])
		}
	}
}

func signedBoundaryPairs[T Signed]() [][2]T {
	bits := uint(unsafe.Sizeof(T(0))) * 8
	maxV := T(uint64(1)<<(bits-1) - 1)
	minV := -maxV - 1
	return [][2]T{
		{minV, minV}, {minV, minV + 1}, {minV, -1}, {minV, 0}, {minV, 1}, {minV, maxV - 1}, {minV, maxV},
		{-1, -1}, {-1, 0}, {-1, 1}, {-1, maxV},
		{0, 0}, {0, 1}, {0, maxV},
		{1, 1}, {1, maxV - 1}, {1, maxV},
		{maxV - 1, maxV}, {maxV, maxV},
		{-100, 100}, {maxV / 2, 2}, {maxV/2 + 1, 2}, {minV / 2, 2}, {minV/2 - 1, 2},
	}
}

func unsignedBoundaryPairs[T Unsigned]() [][2]T {
	maxV := ^T(0)
	return [][2]T{
		{0, 0}, {0, 1}, {0, maxV}, {1, 1}, {1, maxV - 1}, {1, maxV}, {maxV - 1, maxV}, {maxV, maxV},
		{100, 200}, {maxV / 2, 2}, {maxV/2 + 1, 2},
	}
}

// checkBoundaryTriples runs MulDiv and MulMod on every triple of values that
// appear in pairs, against exact math/big arithmetic. MulDiv wants the
// quotient when it fits in T and (0, false) otherwise. MulMod wants the
// remainder, which always fits. Both want (0, false) when c == 0.
func checkBoundaryTriples[T Integer](t *testing.T, pairs [][2]T) {
	t.Helper()
	seen := map[T]bool{}
	var vals []T
	for _, p := range pairs {
		for _, v := range p {
			if !seen[v] {
				seen[v] = true
				vals = append(vals, v)
			}
		}
	}
	lo, hi := boundsOf[T]()
	for _, a := range vals {
		for _, b := range vals {
			for _, c := range vals {
				var wq, wr T
				var qok bool
				if c != 0 {
					exact := new(big.Int).Mul(toBig(a), toBig(b))
					q, r := new(big.Int).QuoRem(exact, toBig(c), new(big.Int))
					if q.Cmp(lo) >= 0 && q.Cmp(hi) <= 0 {
						wq, qok = wrapBig[T](q), true
					}
					wr = wrapBig[T](r)
				}
				if got, ok := MulDiv(a, b, c); got != wq || ok != qok {
					t.Errorf("MulDiv(%d, %d, %d) = (%d, %v), want (%d, %v)", a, b, c, got, ok, wq, qok)
				}
				if got, ok := MulMod(a, b, c); got != wr || ok != (c != 0) {
					t.Errorf("MulMod(%d, %d, %d) = (%d, %v), want (%d, %v)", a, b, c, got, ok, wr, c != 0)
				}
			}
		}
	}
}

func TestBoundary(t *testing.T) {
	t.Run("int16", func(t *testing.T) { testBoundary(t, signedBoundaryPairs[int16]()) })
	t.Run("int32", func(t *testing.T) { testBoundary(t, signedBoundaryPairs[int32]()) })
	t.Run("int64", func(t *testing.T) { testBoundary(t, signedBoundaryPairs[int64]()) })
	t.Run("uint16", func(t *testing.T) { testBoundary(t, unsignedBoundaryPairs[uint16]()) })
	t.Run("uint32", func(t *testing.T) { testBoundary(t, unsignedBoundaryPairs[uint32]()) })
	t.Run("uint64", func(t *testing.T) { testBoundary(t, unsignedBoundaryPairs[uint64]()) })
}

func testBoundary[T Integer](t *testing.T, pairs [][2]T) {
	checkBoundaryPairs(t, pairs)
	checkBoundaryTriples(t, pairs)
}

// ---------------------------------------------------------------------------
// C. Neg / Abs tests
// ---------------------------------------------------------------------------

func TestNegSigned(t *testing.T) {
	tests := []struct {
		name   string
		fn     func() (int64, bool)
		want   int64
		wantOk bool
	}{
		{"int8/MinInt8", func() (int64, bool) { r, ok := Neg[int8](math.MinInt8); return int64(r), ok }, math.MinInt8, false},
		{"int8/-1", func() (int64, bool) { r, ok := Neg[int8](-1); return int64(r), ok }, 1, true},
		{"int8/0", func() (int64, bool) { r, ok := Neg[int8](0); return int64(r), ok }, 0, true},
		{"int8/1", func() (int64, bool) { r, ok := Neg[int8](1); return int64(r), ok }, -1, true},
		{"int8/MaxInt8", func() (int64, bool) { r, ok := Neg[int8](math.MaxInt8); return int64(r), ok }, -math.MaxInt8, true},

		{"int16/MinInt16", func() (int64, bool) { r, ok := Neg[int16](math.MinInt16); return int64(r), ok }, math.MinInt16, false},
		{"int16/-1", func() (int64, bool) { r, ok := Neg[int16](-1); return int64(r), ok }, 1, true},
		{"int16/0", func() (int64, bool) { r, ok := Neg[int16](0); return int64(r), ok }, 0, true},
		{"int16/1", func() (int64, bool) { r, ok := Neg[int16](1); return int64(r), ok }, -1, true},
		{"int16/MaxInt16", func() (int64, bool) { r, ok := Neg[int16](math.MaxInt16); return int64(r), ok }, -math.MaxInt16, true},

		{"int32/MinInt32", func() (int64, bool) { r, ok := Neg[int32](math.MinInt32); return int64(r), ok }, math.MinInt32, false},
		{"int32/-1", func() (int64, bool) { r, ok := Neg[int32](-1); return int64(r), ok }, 1, true},
		{"int32/0", func() (int64, bool) { r, ok := Neg[int32](0); return int64(r), ok }, 0, true},
		{"int32/1", func() (int64, bool) { r, ok := Neg[int32](1); return int64(r), ok }, -1, true},
		{"int32/MaxInt32", func() (int64, bool) { r, ok := Neg[int32](math.MaxInt32); return int64(r), ok }, -math.MaxInt32, true},

		{"int64/MinInt64", func() (int64, bool) { r, ok := Neg[int64](math.MinInt64); return r, ok }, math.MinInt64, false},
		{"int64/-1", func() (int64, bool) { r, ok := Neg[int64](-1); return r, ok }, 1, true},
		{"int64/0", func() (int64, bool) { r, ok := Neg[int64](0); return r, ok }, 0, true},
		{"int64/1", func() (int64, bool) { r, ok := Neg[int64](1); return r, ok }, -1, true},
		{"int64/MaxInt64", func() (int64, bool) { r, ok := Neg[int64](math.MaxInt64); return r, ok }, -math.MaxInt64, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, gotOk := tc.fn()
			if gotOk != tc.wantOk {
				t.Errorf("ok=%v, want %v", gotOk, tc.wantOk)
			}
			if got != tc.want {
				t.Errorf("got=%d, want=%d", got, tc.want)
			}
		})
	}
}

func TestNegUnsigned(t *testing.T) {
	// Neg for unsigned: 0 is ok, everything else overflows.
	t.Run("uint8/0", func(t *testing.T) {
		r, ok := Neg[uint8](0)
		if !ok || r != 0 {
			t.Errorf("Neg(0): got=(%d, %v), want=(0, true)", r, ok)
		}
	})
	t.Run("uint8/1", func(t *testing.T) {
		_, ok := Neg[uint8](1)
		if ok {
			t.Error("Neg(1): ok=true, want false")
		}
	})
	t.Run("uint8/128", func(t *testing.T) {
		_, ok := Neg[uint8](128)
		if ok {
			t.Error("Neg(128): ok=true, want false")
		}
	})
	t.Run("uint8/255", func(t *testing.T) {
		_, ok := Neg[uint8](255)
		if ok {
			t.Error("Neg(255): ok=true, want false")
		}
	})
	t.Run("uint16/0", func(t *testing.T) {
		r, ok := Neg[uint16](0)
		if !ok || r != 0 {
			t.Errorf("Neg(0): got=(%d, %v), want=(0, true)", r, ok)
		}
	})
	t.Run("uint16/MaxUint16", func(t *testing.T) {
		_, ok := Neg[uint16](math.MaxUint16)
		if ok {
			t.Error("Neg(MaxUint16): ok=true, want false")
		}
	})
	t.Run("uint32/0", func(t *testing.T) {
		r, ok := Neg[uint32](0)
		if !ok || r != 0 {
			t.Errorf("Neg(0): got=(%d, %v), want=(0, true)", r, ok)
		}
	})
	t.Run("uint32/MaxUint32", func(t *testing.T) {
		_, ok := Neg[uint32](math.MaxUint32)
		if ok {
			t.Error("Neg(MaxUint32): ok=true, want false")
		}
	})
	t.Run("uint64/0", func(t *testing.T) {
		r, ok := Neg[uint64](0)
		if !ok || r != 0 {
			t.Errorf("Neg(0): got=(%d, %v), want=(0, true)", r, ok)
		}
	})
	t.Run("uint64/1", func(t *testing.T) {
		_, ok := Neg[uint64](1)
		if ok {
			t.Error("Neg(1): ok=true, want false")
		}
	})
	t.Run("uint64/MaxUint64", func(t *testing.T) {
		_, ok := Neg[uint64](math.MaxUint64)
		if ok {
			t.Error("Neg(MaxUint64): ok=true, want false")
		}
	})
}

func TestAbsSigned(t *testing.T) {
	tests := []struct {
		name   string
		fn     func() (int64, bool)
		want   int64
		wantOk bool
	}{
		{"int8/MinInt8", func() (int64, bool) { r, ok := Abs[int8](math.MinInt8); return int64(r), ok }, math.MinInt8, false},
		{"int8/-1", func() (int64, bool) { r, ok := Abs[int8](-1); return int64(r), ok }, 1, true},
		{"int8/0", func() (int64, bool) { r, ok := Abs[int8](0); return int64(r), ok }, 0, true},
		{"int8/1", func() (int64, bool) { r, ok := Abs[int8](1); return int64(r), ok }, 1, true},
		{"int8/MaxInt8", func() (int64, bool) { r, ok := Abs[int8](math.MaxInt8); return int64(r), ok }, math.MaxInt8, true},

		{"int16/MinInt16", func() (int64, bool) { r, ok := Abs[int16](math.MinInt16); return int64(r), ok }, math.MinInt16, false},
		{"int16/-1", func() (int64, bool) { r, ok := Abs[int16](-1); return int64(r), ok }, 1, true},
		{"int16/0", func() (int64, bool) { r, ok := Abs[int16](0); return int64(r), ok }, 0, true},
		{"int16/MaxInt16", func() (int64, bool) { r, ok := Abs[int16](math.MaxInt16); return int64(r), ok }, math.MaxInt16, true},

		{"int32/MinInt32", func() (int64, bool) { r, ok := Abs[int32](math.MinInt32); return int64(r), ok }, math.MinInt32, false},
		{"int32/-1", func() (int64, bool) { r, ok := Abs[int32](-1); return int64(r), ok }, 1, true},
		{"int32/0", func() (int64, bool) { r, ok := Abs[int32](0); return int64(r), ok }, 0, true},
		{"int32/MaxInt32", func() (int64, bool) { r, ok := Abs[int32](math.MaxInt32); return int64(r), ok }, math.MaxInt32, true},

		{"int64/MinInt64", func() (int64, bool) { r, ok := Abs[int64](math.MinInt64); return r, ok }, math.MinInt64, false},
		{"int64/-1", func() (int64, bool) { r, ok := Abs[int64](-1); return r, ok }, 1, true},
		{"int64/0", func() (int64, bool) { r, ok := Abs[int64](0); return r, ok }, 0, true},
		{"int64/1", func() (int64, bool) { r, ok := Abs[int64](1); return r, ok }, 1, true},
		{"int64/MaxInt64", func() (int64, bool) { r, ok := Abs[int64](math.MaxInt64); return r, ok }, math.MaxInt64, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, gotOk := tc.fn()
			if gotOk != tc.wantOk {
				t.Errorf("ok=%v, want %v", gotOk, tc.wantOk)
			}
			if got != tc.want {
				t.Errorf("got=%d, want=%d", got, tc.want)
			}
		})
	}
}

func TestAbsUnsigned(t *testing.T) {
	// Unsigned Abs always returns (a, true).
	vals8 := []uint8{0, 1, 128, 255}
	for _, v := range vals8 {
		t.Run(fmt.Sprintf("uint8/%d", v), func(t *testing.T) {
			r, ok := Abs(v)
			if !ok || r != v {
				t.Errorf("Abs(%d): got=(%d, %v), want=(%d, true)", v, r, ok, v)
			}
		})
	}
	vals64 := []uint64{0, 1, math.MaxUint64 / 2, math.MaxUint64}
	for _, v := range vals64 {
		t.Run(fmt.Sprintf("uint64/%d", v), func(t *testing.T) {
			r, ok := Abs(v)
			if !ok || r != v {
				t.Errorf("Abs(%d): got=(%d, %v), want=(%d, true)", v, r, ok, v)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// D. MulDiv / MulMod tests
// ---------------------------------------------------------------------------

func TestMulDiv(t *testing.T) {
	t.Run("basic_6_7_2", func(t *testing.T) {
		r, ok := MulDiv[int64](6, 7, 2)
		if !ok || r != 21 {
			t.Errorf("MulDiv(6,7,2): got=(%d,%v), want=(21,true)", r, ok)
		}
	})

	t.Run("MaxInt64_2_2", func(t *testing.T) {
		// MaxInt64*2 overflows int64, but MaxInt64*2/2 = MaxInt64 fits.
		r, ok := MulDiv[int64](math.MaxInt64, 2, 2)
		if !ok || r != math.MaxInt64 {
			t.Errorf("MulDiv(MaxInt64,2,2): got=(%d,%v), want=(%d,true)", r, ok, int64(math.MaxInt64))
		}
	})

	t.Run("MaxInt64_2_3", func(t *testing.T) {
		// MaxInt64*2 / 3 = 6148914691236517204 (fits int64).
		r, ok := MulDiv[int64](math.MaxInt64, 2, 3)
		// Reference: big.Int
		ref := new(big.Int).Mul(big.NewInt(math.MaxInt64), big.NewInt(2))
		ref.Quo(ref, big.NewInt(3))
		want := ref.Int64()
		if !ok || r != want {
			t.Errorf("MulDiv(MaxInt64,2,3): got=(%d,%v), want=(%d,true)", r, ok, want)
		}
	})

	t.Run("c_zero", func(t *testing.T) {
		r, ok := MulDiv[int64](1, 2, 0)
		if ok || r != 0 {
			t.Errorf("MulDiv(1,2,0): got=(%d,%v), want=(0,false)", r, ok)
		}
	})

	t.Run("c_zero_uint", func(t *testing.T) {
		r, ok := MulDiv[uint64](1, 2, 0)
		if ok || r != 0 {
			t.Errorf("MulDiv(1,2,0): got=(%d,%v), want=(0,false)", r, ok)
		}
	})

	t.Run("MinInt64_neg1_1", func(t *testing.T) {
		// MinInt64 * -1 = |MinInt64| which overflows int64.
		_, ok := MulDiv[int64](math.MinInt64, -1, 1)
		if ok {
			t.Error("MulDiv(MinInt64,-1,1): ok=true, want false")
		}
	})

	t.Run("uint8_100_50_25", func(t *testing.T) {
		// 100*50 = 5000, 5000/25 = 200, and 200 fits uint8 (max 255).
		r, ok := MulDiv[uint8](100, 50, 25)
		if !ok || r != 200 {
			t.Errorf("MulDiv[uint8](100,50,25): got=(%d,%v), want=(200,true)", r, ok)
		}
	})

	t.Run("uint64_max_max_max", func(t *testing.T) {
		// MaxUint64 * MaxUint64 / MaxUint64 = MaxUint64
		r, ok := MulDiv[uint64](math.MaxUint64, math.MaxUint64, math.MaxUint64)
		if !ok || r != math.MaxUint64 {
			t.Errorf("MulDiv(MaxUint64,MaxUint64,MaxUint64): got=(%d,%v), want=(%d,true)", r, ok, uint64(math.MaxUint64))
		}
	})

	t.Run("int8_small_exact", func(t *testing.T) {
		r, ok := MulDiv[int8](10, 12, 4)
		// 10*12=120, 120/4=30, fits int8.
		if !ok || r != 30 {
			t.Errorf("MulDiv[int8](10,12,4): got=(%d,%v), want=(30,true)", r, ok)
		}
	})

	t.Run("int64_negative_result", func(t *testing.T) {
		r, ok := MulDiv[int64](-6, 7, 2)
		if !ok || r != -21 {
			t.Errorf("MulDiv(-6,7,2): got=(%d,%v), want=(-21,true)", r, ok)
		}
	})

	t.Run("int64_two_negatives", func(t *testing.T) {
		r, ok := MulDiv[int64](-6, -7, 2)
		if !ok || r != 21 {
			t.Errorf("MulDiv(-6,-7,2): got=(%d,%v), want=(21,true)", r, ok)
		}
	})
}

func checkMulDivOverflowZero[T Integer](t *testing.T, a, b, c T) {
	t.Helper()
	if r, ok := MulDiv(a, b, c); ok || r != 0 {
		t.Errorf("MulDiv[%T](%d,%d,%d) = (%d,%v), want (0,false)", a, a, b, c, r, ok)
	}
	if r, ok := New(a).MulDivOverflow(New(b), New(c)); ok || r.Val() != 0 {
		t.Errorf("MulDivOverflow[%T] = (%d,%v), want (0,false)", a, r.Val(), ok)
	}
}

func TestMulDivOverflowZero(t *testing.T) {
	checkMulDivOverflowZero[int8](t, 100, 50, 25)
	checkMulDivOverflowZero[int8](t, -100, 50, 25)
	checkMulDivOverflowZero[int16](t, math.MaxInt16, 2, 1)
	checkMulDivOverflowZero[int32](t, math.MaxInt32, 2, 1)
	checkMulDivOverflowZero[int64](t, math.MaxInt64, 2, 1)
	checkMulDivOverflowZero[int](t, int(^uint(0)>>1), 2, 1)
	checkMulDivOverflowZero[uint8](t, math.MaxUint8, 2, 1)
	checkMulDivOverflowZero[uint16](t, math.MaxUint16, 2, 1)
	checkMulDivOverflowZero[uint32](t, math.MaxUint32, 2, 1)
	checkMulDivOverflowZero[uint64](t, math.MaxUint64, 2, 1)
	checkMulDivOverflowZero[uint](t, ^uint(0), 2, 1)
}

func TestMulMod(t *testing.T) {
	t.Run("MaxUint64_MaxUint64_MaxUint64", func(t *testing.T) {
		r, ok := MulMod[uint64](math.MaxUint64, math.MaxUint64, math.MaxUint64)
		if !ok || r != 0 {
			t.Errorf("MulMod(MaxUint64,MaxUint64,MaxUint64): got=(%d,%v), want=(0,true)", r, ok)
		}
	})

	t.Run("int64_MaxInt64_MaxInt64_10", func(t *testing.T) {
		// MaxInt64 % 10 == 7 and 7*7 % 10 == 9.
		r, ok := MulMod[int64](math.MaxInt64, math.MaxInt64, 10)
		if !ok || r != 9 {
			t.Errorf("MulMod(MaxInt64,MaxInt64,10): got=(%d,%v), want=(9,true)", r, ok)
		}
	})

	t.Run("int64_7_5_3", func(t *testing.T) {
		r, ok := MulMod[int64](7, 5, 3)
		// 7*5=35, 35%3=2.
		if !ok || r != 2 {
			t.Errorf("MulMod(7,5,3): got=(%d,%v), want=(2,true)", r, ok)
		}
	})

	t.Run("c_zero", func(t *testing.T) {
		r, ok := MulMod[int64](1, 2, 0)
		if ok || r != 0 {
			t.Errorf("MulMod(1,2,0): got=(%d,%v), want=(0,false)", r, ok)
		}
	})

	t.Run("c_zero_uint", func(t *testing.T) {
		r, ok := MulMod[uint64](1, 2, 0)
		if ok || r != 0 {
			t.Errorf("MulMod(1,2,0): got=(%d,%v), want=(0,false)", r, ok)
		}
	})

	t.Run("int8_basic", func(t *testing.T) {
		r, ok := MulMod[int8](7, 5, 3)
		if !ok || r != 2 {
			t.Errorf("MulMod[int8](7,5,3): got=(%d,%v), want=(2,true)", r, ok)
		}
	})

	t.Run("uint8_basic", func(t *testing.T) {
		r, ok := MulMod[uint8](100, 50, 25)
		// 100*50=5000, 5000%25=0.
		if !ok || r != 0 {
			t.Errorf("MulMod[uint8](100,50,25): got=(%d,%v), want=(0,true)", r, ok)
		}
	})

	t.Run("int64_intermediate_overflow", func(t *testing.T) {
		r, ok := MulMod[int64](math.MaxInt64, 2, 3)
		// ref: MaxInt64*2 mod 3
		ref := new(big.Int).Mul(big.NewInt(math.MaxInt64), big.NewInt(2))
		ref.Rem(ref, big.NewInt(3))
		want := ref.Int64()
		if !ok || r != want {
			t.Errorf("MulMod(MaxInt64,2,3): got=(%d,%v), want=(%d,true)", r, ok, want)
		}
	})

	t.Run("uint64_intermediate_overflow", func(t *testing.T) {
		r, ok := MulMod[uint64](math.MaxUint64, 2, 3)
		// ref: MaxUint64*2 mod 3
		ref := new(big.Int).Mul(new(big.Int).SetUint64(math.MaxUint64), big.NewInt(2))
		ref.Rem(ref, big.NewInt(3))
		want := ref.Uint64()
		if !ok || r != want {
			t.Errorf("MulMod(MaxUint64,2,3): got=(%d,%v), want=(%d,true)", r, ok, want)
		}
	})

	// c == MinInt64 exercises the upper bound of toUint64Abs(c) == 1<<63,
	// where the remainder can reach MaxInt64 in magnitude.
	t.Run("int64_modulus_min_int64", func(t *testing.T) {
		cases := []struct{ a, b int64 }{
			{math.MaxInt64, -1},
			{math.MaxInt64, 2},
			{math.MinInt64, 3},
			{math.MinInt64, math.MinInt64},
			{-math.MaxInt64, math.MaxInt64},
		}
		for _, tc := range cases {
			r, ok := MulMod[int64](tc.a, tc.b, math.MinInt64)
			ref := new(big.Int).Mul(big.NewInt(tc.a), big.NewInt(tc.b))
			ref.Rem(ref, big.NewInt(math.MinInt64))
			if !ok || !ref.IsInt64() || r != ref.Int64() {
				t.Errorf("MulMod(%d,%d,MinInt64): got=(%d,%v), want=(%s,true)", tc.a, tc.b, r, ok, ref)
			}
		}
	})

	t.Run("int64_negative_dividend", func(t *testing.T) {
		r, ok := MulMod[int64](-7, 5, 3)
		// -7*5 = -35, -35 % 3 = -2 (Go truncates toward zero).
		if !ok || r != -2 {
			t.Errorf("MulMod(-7,5,3): got=(%d,%v), want=(-2,true)", r, ok)
		}
	})
}

// ---------------------------------------------------------------------------
// E. Pow tests
// ---------------------------------------------------------------------------

func TestPow(t *testing.T) {
	t.Run("0^0=1", func(t *testing.T) {
		r, ok := Pow[int64](0, 0)
		if !ok || r != 1 {
			t.Errorf("Pow(0,0): got=(%d,%v), want=(1,true)", r, ok)
		}
	})
	t.Run("0^5=0", func(t *testing.T) {
		r, ok := Pow[int64](0, 5)
		if !ok || r != 0 {
			t.Errorf("Pow(0,5): got=(%d,%v), want=(0,true)", r, ok)
		}
	})
	t.Run("1^100=1", func(t *testing.T) {
		r, ok := Pow[int64](1, 100)
		if !ok || r != 1 {
			t.Errorf("Pow(1,100): got=(%d,%v), want=(1,true)", r, ok)
		}
	})
	t.Run("(-1)^even=1", func(t *testing.T) {
		r, ok := Pow[int64](-1, 42)
		if !ok || r != 1 {
			t.Errorf("Pow(-1,42): got=(%d,%v), want=(1,true)", r, ok)
		}
	})
	t.Run("(-1)^odd=-1", func(t *testing.T) {
		r, ok := Pow[int64](-1, 43)
		if !ok || r != -1 {
			t.Errorf("Pow(-1,43): got=(%d,%v), want=(-1,true)", r, ok)
		}
	})

	// 2^7 = 128
	t.Run("2^7_int8_overflow", func(t *testing.T) {
		r, ok := Pow[int8](2, 7)
		if ok || r != -128 {
			t.Errorf("Pow[int8](2,7): got=(%d,%v), want=(%d,false) (128 overflows int8; wrapped value is -128)", r, ok, -128)
		}
	})
	t.Run("2^7_uint8_ok", func(t *testing.T) {
		r, ok := Pow[uint8](2, 7)
		if !ok || r != 128 {
			t.Errorf("Pow[uint8](2,7): got=(%d,%v), want=(128,true)", r, ok)
		}
	})
	t.Run("2^8_uint8_overflow", func(t *testing.T) {
		r, ok := Pow[uint8](2, 8)
		if ok || r != 0 {
			t.Errorf("Pow[uint8](2,8): got=(%d,%v), want=(%d,false) (256 overflows uint8; wrapped value is 0)", r, ok, 0)
		}
	})

	// 2^62 for int64
	t.Run("2^62_int64_ok", func(t *testing.T) {
		r, ok := Pow[int64](2, 62)
		if !ok || r != 1<<62 {
			t.Errorf("Pow[int64](2,62): got=(%d,%v), want=(%d,true)", r, ok, int64(1<<62))
		}
	})
	t.Run("2^63_int64_overflow", func(t *testing.T) {
		r, ok := Pow[int64](2, 63)
		if ok || r != math.MinInt64 {
			t.Errorf("Pow[int64](2,63): got=(%d,%v), want=(%d,false) (wrapped value is MinInt64)", r, ok, int64(math.MinInt64))
		}
	})
	t.Run("2^63_uint64_ok", func(t *testing.T) {
		r, ok := Pow[uint64](2, 63)
		if !ok || r != 1<<63 {
			t.Errorf("Pow[uint64](2,63): got=(%d,%v), want=(%d,true)", r, ok, uint64(1<<63))
		}
	})
	t.Run("2^64_uint64_overflow", func(t *testing.T) {
		r, ok := Pow[uint64](2, 64)
		if ok || r != 0 {
			t.Errorf("Pow[uint64](2,64): got=(%d,%v), want=(%d,false) (wrapped value is 0)", r, ok, 0)
		}
	})

	// 3^5 = 243 for uint8 (ok), 3^6 = 729 for uint8 (overflow)
	t.Run("3^5_uint8_ok", func(t *testing.T) {
		r, ok := Pow[uint8](3, 5)
		if !ok || r != 243 {
			t.Errorf("Pow[uint8](3,5): got=(%d,%v), want=(243,true)", r, ok)
		}
	})
	t.Run("3^6_uint8_overflow", func(t *testing.T) {
		r, ok := Pow[uint8](3, 6)
		if ok || r != 217 {
			t.Errorf("Pow[uint8](3,6): got=(%d,%v), want=(%d,false) (729 overflows uint8; wrapped value is 729-512=217)", r, ok, 217)
		}
	})

	// The overflow happens while squaring the base (2^8 for int8) or while
	// multiplying into the result (3^5 for int8). Both return the wrapped value.
	t.Run("2^8_int8_overflow_wraps", func(t *testing.T) {
		r, ok := Pow[int8](2, 8)
		if ok || r != 0 {
			t.Errorf("Pow[int8](2,8): got=(%d,%v), want=(0,false)", r, ok)
		}
	})
	t.Run("3^5_int8_overflow_wraps", func(t *testing.T) {
		r, ok := Pow[int8](3, 5)
		if ok || r != -13 {
			t.Errorf("Pow[int8](3,5): got=(%d,%v), want=(-13,false)", r, ok)
		}
	})

	// uint64: 0^0
	t.Run("0^0_uint64", func(t *testing.T) {
		r, ok := Pow[uint64](0, 0)
		if !ok || r != 1 {
			t.Errorf("Pow[uint64](0,0): got=(%d,%v), want=(1,true)", r, ok)
		}
	})

	// int8: 2^6 = 64 (ok, fits in int8)
	t.Run("2^6_int8_ok", func(t *testing.T) {
		r, ok := Pow[int8](2, 6)
		if !ok || r != 64 {
			t.Errorf("Pow[int8](2,6): got=(%d,%v), want=(64,true)", r, ok)
		}
	})
}

// ---------------------------------------------------------------------------
// F. Lsh tests
// ---------------------------------------------------------------------------

func TestLsh(t *testing.T) {
	// a=0: all shifts ok
	t.Run("0_n0", func(t *testing.T) {
		r, ok := Lsh[int64](0, 0)
		if !ok || r != 0 {
			t.Errorf("Lsh(0,0): got=(%d,%v), want=(0,true)", r, ok)
		}
	})
	t.Run("0_n1", func(t *testing.T) {
		r, ok := Lsh[int64](0, 1)
		if !ok || r != 0 {
			t.Errorf("Lsh(0,1): got=(%d,%v), want=(0,true)", r, ok)
		}
	})
	t.Run("0_n64", func(t *testing.T) {
		r, ok := Lsh[int64](0, 64)
		if !ok || r != 0 {
			t.Errorf("Lsh(0,64): got=(%d,%v), want=(0,true)", r, ok)
		}
	})
	t.Run("0_n100", func(t *testing.T) {
		r, ok := Lsh[int64](0, 100)
		if !ok || r != 0 {
			t.Errorf("Lsh(0,100): got=(%d,%v), want=(0,true)", r, ok)
		}
	})

	// a=1: various shifts
	t.Run("1_n0", func(t *testing.T) {
		r, ok := Lsh[int8](1, 0)
		if !ok || r != 1 {
			t.Errorf("Lsh[int8](1,0): got=(%d,%v), want=(1,true)", r, ok)
		}
	})
	t.Run("1_n7_int8_overflow", func(t *testing.T) {
		// 1 << 7 = 128, which is MinInt8 for int8 (sign bit set) — overflow.
		_, ok := Lsh[int8](1, 7)
		if ok {
			t.Error("Lsh[int8](1,7): ok=true, want false")
		}
	})
	t.Run("1_n7_uint8_ok", func(t *testing.T) {
		r, ok := Lsh[uint8](1, 7)
		if !ok || r != 128 {
			t.Errorf("Lsh[uint8](1,7): got=(%d,%v), want=(128,true)", r, ok)
		}
	})
	t.Run("1_n8_uint8_overflow", func(t *testing.T) {
		r, ok := Lsh[uint8](1, 8)
		if ok || r != 0 {
			t.Errorf("Lsh[uint8](1,8): got=(%d,%v), want=(0,false)", r, ok)
		}
	})
	t.Run("1_n63_int64_overflow", func(t *testing.T) {
		_, ok := Lsh[int64](1, 63)
		if ok {
			t.Error("Lsh[int64](1,63): ok=true, want false")
		}
	})
	t.Run("1_n63_uint64_ok", func(t *testing.T) {
		r, ok := Lsh[uint64](1, 63)
		if !ok || r != uint64(1)<<63 {
			t.Errorf("Lsh[uint64](1,63): got=(%d,%v), want=(%d,true)", r, ok, uint64(1)<<63)
		}
	})
	t.Run("1_n64_uint64_overflow", func(t *testing.T) {
		r, ok := Lsh[uint64](1, 64)
		if ok || r != 0 {
			t.Errorf("Lsh[uint64](1,64): got=(%d,%v), want=(0,false)", r, ok)
		}
	})

	// a=-1 (int8) with n=0..7
	t.Run("neg1_n0_int8", func(t *testing.T) {
		r, ok := Lsh[int8](-1, 0)
		if !ok || r != -1 {
			t.Errorf("Lsh[int8](-1,0): got=(%d,%v), want=(-1,true)", r, ok)
		}
	})
	t.Run("neg1_n1_int8", func(t *testing.T) {
		// -1 << 1 = -2, and (-2 >> 1) = -1, so roundtrip passes.
		r, ok := Lsh[int8](-1, 1)
		if !ok || r != -2 {
			t.Errorf("Lsh[int8](-1,1): got=(%d,%v), want=(-2,true)", r, ok)
		}
	})
	t.Run("neg1_n6_int8", func(t *testing.T) {
		// -1 << 6 = -64, (-64 >> 6) = -1, ok.
		r, ok := Lsh[int8](-1, 6)
		if !ok || r != -64 {
			t.Errorf("Lsh[int8](-1,6): got=(%d,%v), want=(-64,true)", r, ok)
		}
	})
	t.Run("neg1_n7_int8", func(t *testing.T) {
		// -1 << 7 = -128 (MinInt8), (-128 >> 7) = -1, ok.
		r, ok := Lsh[int8](-1, 7)
		if !ok || r != math.MinInt8 {
			t.Errorf("Lsh[int8](-1,7): got=(%d,%v), want=(%d,true)", r, ok, int8(math.MinInt8))
		}
	})

	// MinInt with n=0 (ok), n=1 (overflow)
	t.Run("MinInt8_n0", func(t *testing.T) {
		r, ok := Lsh[int8](math.MinInt8, 0)
		if !ok || r != math.MinInt8 {
			t.Errorf("Lsh(MinInt8,0): got=(%d,%v), want=(%d,true)", r, ok, int8(math.MinInt8))
		}
	})
	t.Run("MinInt8_n1_overflow", func(t *testing.T) {
		_, ok := Lsh[int8](math.MinInt8, 1)
		if ok {
			t.Error("Lsh(MinInt8,1): ok=true, want false")
		}
	})
	t.Run("MinInt64_n0", func(t *testing.T) {
		r, ok := Lsh[int64](math.MinInt64, 0)
		if !ok || r != math.MinInt64 {
			t.Errorf("Lsh(MinInt64,0): got=(%d,%v), want=(%d,true)", r, ok, int64(math.MinInt64))
		}
	})
	t.Run("MinInt64_n1_overflow", func(t *testing.T) {
		_, ok := Lsh[int64](math.MinInt64, 1)
		if ok {
			t.Error("Lsh(MinInt64,1): ok=true, want false")
		}
	})

	// MaxInt with n=1 (overflow)
	t.Run("MaxInt8_n1_overflow", func(t *testing.T) {
		_, ok := Lsh[int8](math.MaxInt8, 1)
		if ok {
			t.Error("Lsh(MaxInt8,1): ok=true, want false")
		}
	})
	t.Run("MaxInt64_n1_overflow", func(t *testing.T) {
		_, ok := Lsh[int64](math.MaxInt64, 1)
		if ok {
			t.Error("Lsh(MaxInt64,1): ok=true, want false")
		}
	})
	t.Run("MaxUint64_n1_overflow", func(t *testing.T) {
		_, ok := Lsh[uint64](math.MaxUint64, 1)
		if ok {
			t.Error("Lsh(MaxUint64,1): ok=true, want false")
		}
	})

	// a=1 n=62 for int64 (ok: 1<<62 = 4611686018427387904, fits)
	t.Run("1_n62_int64_ok", func(t *testing.T) {
		r, ok := Lsh[int64](1, 62)
		if !ok || r != 1<<62 {
			t.Errorf("Lsh[int64](1,62): got=(%d,%v), want=(%d,true)", r, ok, int64(1<<62))
		}
	})
}

// ---------------------------------------------------------------------------
// G. Convert tests
// ---------------------------------------------------------------------------

func TestConvert(t *testing.T) {
	// int8 -> uint8
	t.Run("int8_to_uint8_neg1", func(t *testing.T) {
		_, ok := Convert[int8, uint8](-1)
		if ok {
			t.Error("Convert[int8,uint8](-1): ok=true, want false")
		}
	})
	t.Run("int8_to_uint8_0", func(t *testing.T) {
		r, ok := Convert[int8, uint8](0)
		if !ok || r != 0 {
			t.Errorf("Convert[int8,uint8](0): got=(%d,%v), want=(0,true)", r, ok)
		}
	})
	t.Run("int8_to_uint8_127", func(t *testing.T) {
		r, ok := Convert[int8, uint8](127)
		if !ok || r != 127 {
			t.Errorf("Convert[int8,uint8](127): got=(%d,%v), want=(127,true)", r, ok)
		}
	})

	// uint8 -> int8
	t.Run("uint8_to_int8_127", func(t *testing.T) {
		r, ok := Convert[uint8, int8](127)
		if !ok || r != 127 {
			t.Errorf("Convert[uint8,int8](127): got=(%d,%v), want=(127,true)", r, ok)
		}
	})
	t.Run("uint8_to_int8_128", func(t *testing.T) {
		_, ok := Convert[uint8, int8](128)
		if ok {
			t.Error("Convert[uint8,int8](128): ok=true, want false")
		}
	})
	t.Run("uint8_to_int8_255", func(t *testing.T) {
		_, ok := Convert[uint8, int8](255)
		if ok {
			t.Error("Convert[uint8,int8](255): ok=true, want false")
		}
	})

	// int64 -> int8
	t.Run("int64_to_int8_neg129", func(t *testing.T) {
		_, ok := Convert[int64, int8](-129)
		if ok {
			t.Error("Convert[int64,int8](-129): ok=true, want false")
		}
	})
	t.Run("int64_to_int8_neg128", func(t *testing.T) {
		r, ok := Convert[int64, int8](-128)
		if !ok || r != -128 {
			t.Errorf("Convert[int64,int8](-128): got=(%d,%v), want=(-128,true)", r, ok)
		}
	})
	t.Run("int64_to_int8_127", func(t *testing.T) {
		r, ok := Convert[int64, int8](127)
		if !ok || r != 127 {
			t.Errorf("Convert[int64,int8](127): got=(%d,%v), want=(127,true)", r, ok)
		}
	})
	t.Run("int64_to_int8_128", func(t *testing.T) {
		_, ok := Convert[int64, int8](128)
		if ok {
			t.Error("Convert[int64,int8](128): ok=true, want false")
		}
	})

	// uint64 -> int64
	t.Run("uint64_to_int64_MaxInt64", func(t *testing.T) {
		r, ok := Convert[uint64, int64](math.MaxInt64)
		if !ok || r != math.MaxInt64 {
			t.Errorf("Convert[uint64,int64](MaxInt64): got=(%d,%v), want=(%d,true)", r, ok, int64(math.MaxInt64))
		}
	})
	t.Run("uint64_to_int64_MaxInt64_plus1", func(t *testing.T) {
		_, ok := Convert[uint64, int64](uint64(math.MaxInt64) + 1)
		if ok {
			t.Error("Convert[uint64,int64](MaxInt64+1): ok=true, want false")
		}
	})

	// int64 -> uint64
	t.Run("int64_to_uint64_neg1", func(t *testing.T) {
		_, ok := Convert[int64, uint64](-1)
		if ok {
			t.Error("Convert[int64,uint64](-1): ok=true, want false")
		}
	})
	t.Run("int64_to_uint64_0", func(t *testing.T) {
		r, ok := Convert[int64, uint64](0)
		if !ok || r != 0 {
			t.Errorf("Convert[int64,uint64](0): got=(%d,%v), want=(0,true)", r, ok)
		}
	})
	t.Run("int64_to_uint64_MaxInt64", func(t *testing.T) {
		r, ok := Convert[int64, uint64](math.MaxInt64)
		if !ok || r != uint64(math.MaxInt64) {
			t.Errorf("Convert[int64,uint64](MaxInt64): got=(%d,%v), want=(%d,true)", r, ok, uint64(math.MaxInt64))
		}
	})

	// Same type: always ok
	t.Run("int64_to_int64_MinInt64", func(t *testing.T) {
		r, ok := Convert[int64, int64](math.MinInt64)
		if !ok || r != math.MinInt64 {
			t.Errorf("Convert[int64,int64](MinInt64): got=(%d,%v), want=(%d,true)", r, ok, int64(math.MinInt64))
		}
	})
	t.Run("uint64_to_uint64_MaxUint64", func(t *testing.T) {
		r, ok := Convert[uint64, uint64](math.MaxUint64)
		if !ok || r != uint64(math.MaxUint64) {
			t.Errorf("Convert[uint64,uint64](MaxUint64): got=(%d,%v), want=(%d,true)", r, ok, uint64(math.MaxUint64))
		}
	})
	t.Run("int8_to_int8_MinInt8", func(t *testing.T) {
		r, ok := Convert[int8, int8](math.MinInt8)
		if !ok || r != math.MinInt8 {
			t.Errorf("Convert[int8,int8](MinInt8): got=(%d,%v), want=(%d,true)", r, ok, int8(math.MinInt8))
		}
	})

	// Wider signed to narrower unsigned
	t.Run("int32_to_uint8_neg1", func(t *testing.T) {
		_, ok := Convert[int32, uint8](-1)
		if ok {
			t.Error("Convert[int32,uint8](-1): ok=true, want false")
		}
	})
	t.Run("int32_to_uint8_255", func(t *testing.T) {
		r, ok := Convert[int32, uint8](255)
		if !ok || r != 255 {
			t.Errorf("Convert[int32,uint8](255): got=(%d,%v), want=(255,true)", r, ok)
		}
	})
	t.Run("int32_to_uint8_256", func(t *testing.T) {
		_, ok := Convert[int32, uint8](256)
		if ok {
			t.Error("Convert[int32,uint8](256): ok=true, want false")
		}
	})

	// Wider unsigned to narrower signed
	t.Run("uint32_to_int16_32767", func(t *testing.T) {
		r, ok := Convert[uint32, int16](32767)
		if !ok || r != 32767 {
			t.Errorf("Convert[uint32,int16](32767): got=(%d,%v), want=(32767,true)", r, ok)
		}
	})
	t.Run("uint32_to_int16_32768", func(t *testing.T) {
		_, ok := Convert[uint32, int16](32768)
		if ok {
			t.Error("Convert[uint32,int16](32768): ok=true, want false")
		}
	})
}

// ---------------------------------------------------------------------------
// H. DivMod special cases
// ---------------------------------------------------------------------------

func TestDivModSpecialCases(t *testing.T) {
	// Division by zero and MinInt / -1 for 64-bit types (8-bit types are
	// covered by TestExhaustive8Bit, Div by TestBoundary).
	t.Run("divmod_by_zero_int64", func(t *testing.T) {
		q, r, ok := DivMod[int64](42, 0)
		if ok || q != 0 || r != 0 {
			t.Errorf("DivMod(42, 0): got=(%d,%d,%v), want=(0,0,false)", q, r, ok)
		}
	})
	t.Run("mod_by_zero_int64", func(t *testing.T) {
		r, ok := Mod[int64](42, 0)
		if ok || r != 0 {
			t.Errorf("Mod(42, 0): got=(%d,%v), want=(0,false)", r, ok)
		}
	})

	// MinInt / -1 → Div overflow, Mod ok with result 0.
	t.Run("MinInt16_mod_neg1", func(t *testing.T) {
		r, ok := Mod[int16](math.MinInt16, -1)
		if !ok || r != 0 {
			t.Errorf("Mod(MinInt16, -1): got=(%d,%v), want=(0,true)", r, ok)
		}
	})
	t.Run("MinInt32_mod_neg1", func(t *testing.T) {
		r, ok := Mod[int32](math.MinInt32, -1)
		if !ok || r != 0 {
			t.Errorf("Mod(MinInt32, -1): got=(%d,%v), want=(0,true)", r, ok)
		}
	})
	t.Run("MinInt64_mod_neg1", func(t *testing.T) {
		r, ok := Mod[int64](math.MinInt64, -1)
		if !ok || r != 0 {
			t.Errorf("Mod(MinInt64, -1): got=(%d,%v), want=(0,true)", r, ok)
		}
	})

	// DivMod for MinInt/-1: q overflows, r=0.
	t.Run("MinInt64_divmod_neg1", func(t *testing.T) {
		q, r, ok := DivMod[int64](math.MinInt64, -1)
		if ok || q != 0 || r != 0 {
			t.Errorf("DivMod(MinInt64, -1): got=(%d,%d,%v), want=(0,0,false)", q, r, ok)
		}
	})

	// Truncation to zero: (-1, 3), (1, MaxInt) → ok=true.
	// This tests the bug fix: q==0 should be ok.
	t.Run("truncation_neg1_3", func(t *testing.T) {
		q, ok := Div[int64](-1, 3)
		if !ok || q != 0 {
			t.Errorf("Div(-1, 3): got=(%d,%v), want=(0,true)", q, ok)
		}
	})
	t.Run("truncation_1_MaxInt64", func(t *testing.T) {
		q, ok := Div[int64](1, math.MaxInt64)
		if !ok || q != 0 {
			t.Errorf("Div(1, MaxInt64): got=(%d,%v), want=(0,true)", q, ok)
		}
	})

	// DivMod consistency: q*b + r == a when ok.
	t.Run("divmod_consistency_7_3", func(t *testing.T) {
		q, r, ok := DivMod[int64](7, 3)
		if !ok || q != 2 || r != 1 {
			t.Errorf("DivMod(7,3): got=(%d,%d,%v), want=(2,1,true)", q, r, ok)
		}
	})
	t.Run("divmod_consistency_neg7_3", func(t *testing.T) {
		q, r, ok := DivMod[int64](-7, 3)
		if !ok || q != -2 || r != -1 {
			t.Errorf("DivMod(-7,3): got=(%d,%d,%v), want=(-2,-1,true)", q, r, ok)
		}
	})
	t.Run("divmod_consistency_7_neg3", func(t *testing.T) {
		q, r, ok := DivMod[int64](7, -3)
		if !ok || q != -2 || r != 1 {
			t.Errorf("DivMod(7,-3): got=(%d,%d,%v), want=(-2,1,true)", q, r, ok)
		}
	})
}

// ---------------------------------------------------------------------------
// I. Must* panic tests
// ---------------------------------------------------------------------------

// expectPanic checks that fn panics with exactly the message want. Comparing
// the message makes sure the panic comes from the library, not from the Go
// runtime (for example a native division by zero).
func expectPanic(t *testing.T, name, want string, fn func()) {
	t.Helper()
	defer func() {
		r := recover()
		if r == nil {
			t.Errorf("%s: expected panic %q but did not panic", name, want)
		} else if r != want {
			t.Errorf("%s: panic = %#v, want %q", name, r, want)
		}
	}()
	fn()
}

func expectNoPanic(t *testing.T, name string, fn func()) {
	t.Helper()
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("%s: unexpected panic: %v", name, r)
		}
	}()
	fn()
}

func TestMustAddStandalone(t *testing.T) {
	t.Run("no_overflow", func(t *testing.T) {
		expectNoPanic(t, "MustAdd(1,2)", func() {
			r := MustAdd[int64](1, 2)
			if r != 3 {
				t.Errorf("MustAdd(1,2): got=%d, want=3", r)
			}
		})
	})
	t.Run("overflow", func(t *testing.T) {
		expectPanic(t, "MustAdd(MaxInt64,1)", "safeint: Add overflow", func() {
			MustAdd[int64](math.MaxInt64, 1)
		})
	})
}

func TestMustSubStandalone(t *testing.T) {
	t.Run("no_overflow", func(t *testing.T) {
		expectNoPanic(t, "MustSub(5,3)", func() {
			r := MustSub[int64](5, 3)
			if r != 2 {
				t.Errorf("MustSub(5,3): got=%d, want=2", r)
			}
		})
	})
	t.Run("overflow", func(t *testing.T) {
		expectPanic(t, "MustSub(MinInt64,1)", "safeint: Sub overflow", func() {
			MustSub[int64](math.MinInt64, 1)
		})
	})
}

func TestMustMulStandalone(t *testing.T) {
	t.Run("no_overflow", func(t *testing.T) {
		expectNoPanic(t, "MustMul(6,7)", func() {
			r := MustMul[int64](6, 7)
			if r != 42 {
				t.Errorf("MustMul(6,7): got=%d, want=42", r)
			}
		})
	})
	t.Run("overflow", func(t *testing.T) {
		expectPanic(t, "MustMul(MaxInt64,2)", "safeint: Mul overflow", func() {
			MustMul[int64](math.MaxInt64, 2)
		})
	})
}

func TestMustDivStandalone(t *testing.T) {
	t.Run("no_overflow", func(t *testing.T) {
		expectNoPanic(t, "MustDiv(42,7)", func() {
			r := MustDiv[int64](42, 7)
			if r != 6 {
				t.Errorf("MustDiv(42,7): got=%d, want=6", r)
			}
		})
	})
	t.Run("div_by_zero", func(t *testing.T) {
		expectPanic(t, "MustDiv(1,0)", "safeint: Div overflow", func() {
			MustDiv[int64](1, 0)
		})
	})
	t.Run("MinInt64_neg1", func(t *testing.T) {
		expectPanic(t, "MustDiv(MinInt64,-1)", "safeint: Div overflow", func() {
			MustDiv[int64](math.MinInt64, -1)
		})
	})
}

func TestMustNegStandalone(t *testing.T) {
	t.Run("no_overflow", func(t *testing.T) {
		expectNoPanic(t, "MustNeg(5)", func() {
			r := MustNeg[int64](5)
			if r != -5 {
				t.Errorf("MustNeg(5): got=%d, want=-5", r)
			}
		})
	})
	t.Run("overflow", func(t *testing.T) {
		expectPanic(t, "MustNeg(MinInt64)", "safeint: Neg overflow", func() {
			MustNeg[int64](math.MinInt64)
		})
	})
}

func TestMustMulDivStandalone(t *testing.T) {
	t.Run("no_overflow", func(t *testing.T) {
		expectNoPanic(t, "MustMulDiv(6,7,2)", func() {
			r := MustMulDiv[int64](6, 7, 2)
			if r != 21 {
				t.Errorf("MustMulDiv(6,7,2): got=%d, want=21", r)
			}
		})
	})
	t.Run("overflow_c_zero", func(t *testing.T) {
		expectPanic(t, "MustMulDiv(1,2,0)", "safeint: MulDiv overflow", func() {
			MustMulDiv[int64](1, 2, 0)
		})
	})
	t.Run("overflow_result", func(t *testing.T) {
		expectPanic(t, "MustMulDiv(MinInt64,-1,1)", "safeint: MulDiv overflow", func() {
			MustMulDiv[int64](math.MinInt64, -1, 1)
		})
	})
}

func TestMustConvertStandalone(t *testing.T) {
	t.Run("no_overflow", func(t *testing.T) {
		expectNoPanic(t, "MustConvert[int64,int8](42)", func() {
			r := MustConvert[int64, int8](42)
			if r != 42 {
				t.Errorf("MustConvert(42): got=%d, want=42", r)
			}
		})
	})
	t.Run("overflow", func(t *testing.T) {
		expectPanic(t, "MustConvert[int64,int8](128)", "safeint: Convert overflow", func() {
			MustConvert[int64, int8](128)
		})
	})
	t.Run("sign_overflow", func(t *testing.T) {
		expectPanic(t, "MustConvert[int64,uint64](-1)", "safeint: Convert overflow", func() {
			MustConvert[int64, uint64](-1)
		})
	})
}

// ---------------------------------------------------------------------------
// K. Fuzz tests
// ---------------------------------------------------------------------------

func FuzzAddInt64(f *testing.F) {
	f.Add(int64(0), int64(0))
	f.Add(int64(math.MinInt64), int64(-1))
	f.Add(int64(math.MaxInt64), int64(1))
	f.Add(int64(math.MinInt64), int64(math.MaxInt64))
	f.Add(int64(1), int64(-1))
	f.Fuzz(func(t *testing.T, a, b int64) {
		got, ok := Add(a, b)
		ref := new(big.Int).Add(big.NewInt(a), big.NewInt(b))
		lo := new(big.Int).SetInt64(math.MinInt64)
		hi := new(big.Int).SetInt64(math.MaxInt64)
		fits := ref.Cmp(lo) >= 0 && ref.Cmp(hi) <= 0
		if ok != fits {
			t.Errorf("Add(%d, %d): ok=%v, want %v", a, b, ok, fits)
		}
		if ok && got != ref.Int64() {
			t.Errorf("Add(%d, %d): got=%d, want=%d", a, b, got, ref.Int64())
		}
	})
}

func FuzzSubInt64(f *testing.F) {
	f.Add(int64(0), int64(0))
	f.Add(int64(math.MinInt64), int64(1))
	f.Add(int64(math.MaxInt64), int64(-1))
	f.Add(int64(0), int64(math.MinInt64))
	f.Fuzz(func(t *testing.T, a, b int64) {
		got, ok := Sub(a, b)
		ref := new(big.Int).Sub(big.NewInt(a), big.NewInt(b))
		lo := new(big.Int).SetInt64(math.MinInt64)
		hi := new(big.Int).SetInt64(math.MaxInt64)
		fits := ref.Cmp(lo) >= 0 && ref.Cmp(hi) <= 0
		if ok != fits {
			t.Errorf("Sub(%d, %d): ok=%v, want %v", a, b, ok, fits)
		}
		if ok && got != ref.Int64() {
			t.Errorf("Sub(%d, %d): got=%d, want=%d", a, b, got, ref.Int64())
		}
	})
}

func FuzzMulInt64(f *testing.F) {
	f.Add(int64(0), int64(0))
	f.Add(int64(math.MinInt64), int64(-1))
	f.Add(int64(math.MaxInt64), int64(2))
	f.Add(int64(-1), int64(-1))
	f.Add(int64(1), int64(math.MaxInt64))
	f.Fuzz(func(t *testing.T, a, b int64) {
		got, ok := Mul(a, b)
		ref := new(big.Int).Mul(big.NewInt(a), big.NewInt(b))
		lo := new(big.Int).SetInt64(math.MinInt64)
		hi := new(big.Int).SetInt64(math.MaxInt64)
		fits := ref.Cmp(lo) >= 0 && ref.Cmp(hi) <= 0
		if ok != fits {
			t.Errorf("Mul(%d, %d): ok=%v, want %v", a, b, ok, fits)
		}
		if ok && got != ref.Int64() {
			t.Errorf("Mul(%d, %d): got=%d, want=%d", a, b, got, ref.Int64())
		}
	})
}

func FuzzDivInt64(f *testing.F) {
	f.Add(int64(0), int64(1))
	f.Add(int64(math.MinInt64), int64(-1))
	f.Add(int64(42), int64(0))
	f.Add(int64(1), int64(-3))
	f.Add(int64(math.MaxInt64), int64(1))
	f.Fuzz(func(t *testing.T, a, b int64) {
		got, ok := Div(a, b)
		if b == 0 {
			if ok {
				t.Errorf("Div(%d, 0): ok=true, want false", a)
			}
			return
		}
		ref := new(big.Int).Quo(big.NewInt(a), big.NewInt(b))
		lo := new(big.Int).SetInt64(math.MinInt64)
		hi := new(big.Int).SetInt64(math.MaxInt64)
		fits := ref.Cmp(lo) >= 0 && ref.Cmp(hi) <= 0
		if ok != fits {
			t.Errorf("Div(%d, %d): ok=%v, want %v (ref=%s)", a, b, ok, fits, ref)
		}
		if ok && got != ref.Int64() {
			t.Errorf("Div(%d, %d): got=%d, want=%d", a, b, got, ref.Int64())
		}
	})
}

func FuzzAddUint64(f *testing.F) {
	f.Add(uint64(0), uint64(0))
	f.Add(uint64(math.MaxUint64), uint64(1))
	f.Add(uint64(math.MaxUint64), uint64(math.MaxUint64))
	f.Add(uint64(1), uint64(1))
	f.Fuzz(func(t *testing.T, a, b uint64) {
		got, ok := Add(a, b)
		ref := new(big.Int).Add(new(big.Int).SetUint64(a), new(big.Int).SetUint64(b))
		bmax := new(big.Int).SetUint64(math.MaxUint64)
		fits := ref.Sign() >= 0 && ref.Cmp(bmax) <= 0
		if ok != fits {
			t.Errorf("Add(%d, %d): ok=%v, want %v", a, b, ok, fits)
		}
		if ok && got != ref.Uint64() {
			t.Errorf("Add(%d, %d): got=%d, want=%d", a, b, got, ref.Uint64())
		}
	})
}

func FuzzSubUint64(f *testing.F) {
	f.Add(uint64(0), uint64(0))
	f.Add(uint64(0), uint64(1))
	f.Add(uint64(math.MaxUint64), uint64(math.MaxUint64))
	f.Add(uint64(1), uint64(0))
	f.Fuzz(func(t *testing.T, a, b uint64) {
		got, ok := Sub(a, b)
		ref := new(big.Int).Sub(new(big.Int).SetUint64(a), new(big.Int).SetUint64(b))
		bmax := new(big.Int).SetUint64(math.MaxUint64)
		fits := ref.Sign() >= 0 && ref.Cmp(bmax) <= 0
		if ok != fits {
			t.Errorf("Sub(%d, %d): ok=%v, want %v", a, b, ok, fits)
		}
		if ok && got != ref.Uint64() {
			t.Errorf("Sub(%d, %d): got=%d, want=%d", a, b, got, ref.Uint64())
		}
	})
}

func FuzzMulUint64(f *testing.F) {
	f.Add(uint64(0), uint64(0))
	f.Add(uint64(math.MaxUint64), uint64(2))
	f.Add(uint64(math.MaxUint64), uint64(math.MaxUint64))
	f.Add(uint64(1), uint64(1))
	f.Fuzz(func(t *testing.T, a, b uint64) {
		got, ok := Mul(a, b)
		ref := new(big.Int).Mul(new(big.Int).SetUint64(a), new(big.Int).SetUint64(b))
		bmax := new(big.Int).SetUint64(math.MaxUint64)
		fits := ref.Sign() >= 0 && ref.Cmp(bmax) <= 0
		if ok != fits {
			t.Errorf("Mul(%d, %d): ok=%v, want %v", a, b, ok, fits)
		}
		if ok && got != ref.Uint64() {
			t.Errorf("Mul(%d, %d): got=%d, want=%d", a, b, got, ref.Uint64())
		}
	})
}

func FuzzDivUint64(f *testing.F) {
	f.Add(uint64(0), uint64(1))
	f.Add(uint64(42), uint64(0))
	f.Add(uint64(math.MaxUint64), uint64(1))
	f.Add(uint64(math.MaxUint64), uint64(math.MaxUint64))
	f.Fuzz(func(t *testing.T, a, b uint64) {
		got, ok := Div(a, b)
		if b == 0 {
			if ok {
				t.Errorf("Div(%d, 0): ok=true, want false", a)
			}
			return
		}
		want := a / b
		if !ok {
			t.Errorf("Div(%d, %d): ok=false, want true", a, b)
		}
		if ok && got != want {
			t.Errorf("Div(%d, %d): got=%d, want=%d", a, b, got, want)
		}
	})
}
