package safeint

import (
	"fmt"
	"math"
	"math/big"
	"testing"
)

// ---------------------------------------------------------------------------
// A. Exhaustive 8-bit tests
// ---------------------------------------------------------------------------

// refAdd computes a+b in wide arithmetic and checks if it fits in [lo, hi].
func refAdd64(a, b, lo, hi int64) (int64, bool) {
	r := a + b
	return r, r >= lo && r <= hi
}

func refSub64(a, b, lo, hi int64) (int64, bool) {
	r := a - b
	return r, r >= lo && r <= hi
}

func refMul64(a, b, lo, hi int64) (int64, bool) {
	r := big.NewInt(a)
	r.Mul(r, big.NewInt(b))
	if r.Cmp(big.NewInt(lo)) < 0 || r.Cmp(big.NewInt(hi)) > 0 {
		low := r.Int64()
		return low, false
	}
	return r.Int64(), true
}

func refDiv64(a, b, lo, hi int64) (int64, bool) {
	if b == 0 {
		return 0, false
	}
	r := big.NewInt(a)
	r.Quo(r, big.NewInt(b))
	if r.Cmp(big.NewInt(lo)) < 0 || r.Cmp(big.NewInt(hi)) > 0 {
		return 0, false
	}
	return r.Int64(), true
}

func refMod64(a, b int64) (int64, bool) {
	if b == 0 {
		return 0, false
	}
	r := big.NewInt(a)
	r.Rem(r, big.NewInt(b))
	return r.Int64(), true
}

func TestExhaustiveInt8Add(t *testing.T) {
	errors := 0
	for ai := -128; ai <= 127; ai++ {
		for bi := -128; bi <= 127; bi++ {
			a, b := int8(ai), int8(bi)
			got, gotOk := Add(a, b)
			ref, refOk := refAdd64(int64(a), int64(b), math.MinInt8, math.MaxInt8)
			wantOk := refOk
			var want int8
			if wantOk {
				want = int8(ref)
			}
			if gotOk != wantOk {
				t.Errorf("Add(%d, %d): ok=%v, want ok=%v", a, b, gotOk, wantOk)
				errors++
			} else if gotOk && got != want {
				t.Errorf("Add(%d, %d): got=%d, want=%d", a, b, got, want)
				errors++
			}
			if errors >= 10 {
				t.Fatalf("too many errors, stopping")
			}
		}
	}
}

func TestExhaustiveInt8Sub(t *testing.T) {
	errors := 0
	for ai := -128; ai <= 127; ai++ {
		for bi := -128; bi <= 127; bi++ {
			a, b := int8(ai), int8(bi)
			got, gotOk := Sub(a, b)
			ref, refOk := refSub64(int64(a), int64(b), math.MinInt8, math.MaxInt8)
			wantOk := refOk
			var want int8
			if wantOk {
				want = int8(ref)
			}
			if gotOk != wantOk {
				t.Errorf("Sub(%d, %d): ok=%v, want ok=%v", a, b, gotOk, wantOk)
				errors++
			} else if gotOk && got != want {
				t.Errorf("Sub(%d, %d): got=%d, want=%d", a, b, got, want)
				errors++
			}
			if errors >= 10 {
				t.Fatalf("too many errors, stopping")
			}
		}
	}
}

func TestExhaustiveInt8Mul(t *testing.T) {
	errors := 0
	for ai := -128; ai <= 127; ai++ {
		for bi := -128; bi <= 127; bi++ {
			a, b := int8(ai), int8(bi)
			got, gotOk := Mul(a, b)
			ref, refOk := refMul64(int64(a), int64(b), math.MinInt8, math.MaxInt8)
			wantOk := refOk
			var want int8
			if wantOk {
				want = int8(ref)
			}
			if gotOk != wantOk {
				t.Errorf("Mul(%d, %d): ok=%v, want ok=%v", a, b, gotOk, wantOk)
				errors++
			} else if gotOk && got != want {
				t.Errorf("Mul(%d, %d): got=%d, want=%d", a, b, got, want)
				errors++
			}
			if errors >= 10 {
				t.Fatalf("too many errors, stopping")
			}
		}
	}
}

func TestExhaustiveInt8Div(t *testing.T) {
	errors := 0
	for ai := -128; ai <= 127; ai++ {
		for bi := -128; bi <= 127; bi++ {
			a, b := int8(ai), int8(bi)
			got, gotOk := Div(a, b)
			ref, refOk := refDiv64(int64(a), int64(b), math.MinInt8, math.MaxInt8)
			wantOk := refOk
			var want int8
			if wantOk {
				want = int8(ref)
			}
			if gotOk != wantOk {
				t.Errorf("Div(%d, %d): ok=%v, want ok=%v", a, b, gotOk, wantOk)
				errors++
			} else if gotOk && got != want {
				t.Errorf("Div(%d, %d): got=%d, want=%d", a, b, got, want)
				errors++
			}
			if errors >= 10 {
				t.Fatalf("too many errors, stopping")
			}
		}
	}
}

func TestExhaustiveInt8DivMod(t *testing.T) {
	errors := 0
	for ai := -128; ai <= 127; ai++ {
		for bi := -128; bi <= 127; bi++ {
			a, b := int8(ai), int8(bi)
			gotQ, gotR, gotOk := DivMod(a, b)
			refQ, refQOk := refDiv64(int64(a), int64(b), math.MinInt8, math.MaxInt8)
			refR, _ := refMod64(int64(a), int64(b))

			if gotOk != refQOk {
				t.Errorf("DivMod(%d, %d): ok=%v, want ok=%v", a, b, gotOk, refQOk)
				errors++
			} else if gotOk {
				wantQ := int8(refQ)
				wantR := int8(refR)
				if gotQ != wantQ || gotR != wantR {
					t.Errorf("DivMod(%d, %d): got=(%d, %d), want=(%d, %d)", a, b, gotQ, gotR, wantQ, wantR)
					errors++
				}
			}
			if errors >= 10 {
				t.Fatalf("too many errors, stopping")
			}
		}
	}
}

func TestExhaustiveInt8Mod(t *testing.T) {
	errors := 0
	for ai := -128; ai <= 127; ai++ {
		for bi := -128; bi <= 127; bi++ {
			a, b := int8(ai), int8(bi)
			got, gotOk := Mod(a, b)
			ref, refOk := refMod64(int64(a), int64(b))
			wantOk := refOk
			var want int8
			if wantOk {
				want = int8(ref)
			}
			if gotOk != wantOk {
				t.Errorf("Mod(%d, %d): ok=%v, want ok=%v", a, b, gotOk, wantOk)
				errors++
			} else if gotOk && got != want {
				t.Errorf("Mod(%d, %d): got=%d, want=%d", a, b, got, want)
				errors++
			}
			if errors >= 10 {
				t.Fatalf("too many errors, stopping")
			}
		}
	}
}

func TestExhaustiveUint8Add(t *testing.T) {
	errors := 0
	for ai := 0; ai <= 255; ai++ {
		for bi := 0; bi <= 255; bi++ {
			a, b := uint8(ai), uint8(bi)
			got, gotOk := Add(a, b)
			sum := uint64(a) + uint64(b)
			wantOk := sum <= math.MaxUint8
			var want uint8
			if wantOk {
				want = uint8(sum)
			}
			if gotOk != wantOk {
				t.Errorf("Add(%d, %d): ok=%v, want ok=%v", a, b, gotOk, wantOk)
				errors++
			} else if gotOk && got != want {
				t.Errorf("Add(%d, %d): got=%d, want=%d", a, b, got, want)
				errors++
			}
			if errors >= 10 {
				t.Fatalf("too many errors, stopping")
			}
		}
	}
}

func TestExhaustiveUint8Sub(t *testing.T) {
	errors := 0
	for ai := 0; ai <= 255; ai++ {
		for bi := 0; bi <= 255; bi++ {
			a, b := uint8(ai), uint8(bi)
			got, gotOk := Sub(a, b)
			wantOk := a >= b
			var want uint8
			if wantOk {
				want = a - b
			}
			if gotOk != wantOk {
				t.Errorf("Sub(%d, %d): ok=%v, want ok=%v", a, b, gotOk, wantOk)
				errors++
			} else if gotOk && got != want {
				t.Errorf("Sub(%d, %d): got=%d, want=%d", a, b, got, want)
				errors++
			}
			if errors >= 10 {
				t.Fatalf("too many errors, stopping")
			}
		}
	}
}

func TestExhaustiveUint8Mul(t *testing.T) {
	errors := 0
	for ai := 0; ai <= 255; ai++ {
		for bi := 0; bi <= 255; bi++ {
			a, b := uint8(ai), uint8(bi)
			got, gotOk := Mul(a, b)
			prod := uint64(a) * uint64(b)
			wantOk := prod <= math.MaxUint8
			var want uint8
			if wantOk {
				want = uint8(prod)
			}
			if gotOk != wantOk {
				t.Errorf("Mul(%d, %d): ok=%v, want ok=%v", a, b, gotOk, wantOk)
				errors++
			} else if gotOk && got != want {
				t.Errorf("Mul(%d, %d): got=%d, want=%d", a, b, got, want)
				errors++
			}
			if errors >= 10 {
				t.Fatalf("too many errors, stopping")
			}
		}
	}
}

func TestExhaustiveUint8Div(t *testing.T) {
	errors := 0
	for ai := 0; ai <= 255; ai++ {
		for bi := 0; bi <= 255; bi++ {
			a, b := uint8(ai), uint8(bi)
			got, gotOk := Div(a, b)
			if b == 0 {
				if gotOk {
					t.Errorf("Div(%d, 0): ok=true, want ok=false", a)
					errors++
				}
			} else {
				want := a / b
				if !gotOk {
					t.Errorf("Div(%d, %d): ok=false, want ok=true", a, b)
					errors++
				} else if got != want {
					t.Errorf("Div(%d, %d): got=%d, want=%d", a, b, got, want)
					errors++
				}
			}
			if errors >= 10 {
				t.Fatalf("too many errors, stopping")
			}
		}
	}
}

func TestExhaustiveUint8DivMod(t *testing.T) {
	errors := 0
	for ai := 0; ai <= 255; ai++ {
		for bi := 0; bi <= 255; bi++ {
			a, b := uint8(ai), uint8(bi)
			gotQ, gotR, gotOk := DivMod(a, b)
			if b == 0 {
				if gotOk {
					t.Errorf("DivMod(%d, 0): ok=true, want ok=false", a)
					errors++
				}
			} else {
				wantQ := a / b
				wantR := a % b
				if !gotOk {
					t.Errorf("DivMod(%d, %d): ok=false, want ok=true", a, b)
					errors++
				} else if gotQ != wantQ || gotR != wantR {
					t.Errorf("DivMod(%d, %d): got=(%d,%d), want=(%d,%d)", a, b, gotQ, gotR, wantQ, wantR)
					errors++
				}
			}
			if errors >= 10 {
				t.Fatalf("too many errors, stopping")
			}
		}
	}
}

func TestExhaustiveUint8Mod(t *testing.T) {
	errors := 0
	for ai := 0; ai <= 255; ai++ {
		for bi := 0; bi <= 255; bi++ {
			a, b := uint8(ai), uint8(bi)
			got, gotOk := Mod(a, b)
			if b == 0 {
				if gotOk {
					t.Errorf("Mod(%d, 0): ok=true, want ok=false", a)
					errors++
				}
			} else {
				want := a % b
				if !gotOk {
					t.Errorf("Mod(%d, %d): ok=false, want ok=true", a, b)
					errors++
				} else if got != want {
					t.Errorf("Mod(%d, %d): got=%d, want=%d", a, b, got, want)
					errors++
				}
			}
			if errors >= 10 {
				t.Fatalf("too many errors, stopping")
			}
		}
	}
}

// ---------------------------------------------------------------------------
// B. Boundary value pair tests
// ---------------------------------------------------------------------------

// Helper: check a binary op using big.Int as reference for signed types.
func checkSignedOp[T Signed](t *testing.T, opName string, op func(T, T) (T, bool), a, b T, lo, hi *big.Int) {
	t.Helper()
	got, gotOk := op(a, b)
	var ref *big.Int
	switch opName {
	case "Add":
		ref = new(big.Int).Add(big.NewInt(int64(a)), big.NewInt(int64(b)))
	case "Sub":
		ref = new(big.Int).Sub(big.NewInt(int64(a)), big.NewInt(int64(b)))
	case "Mul":
		ref = new(big.Int).Mul(big.NewInt(int64(a)), big.NewInt(int64(b)))
	case "Div":
		if b == 0 {
			if gotOk {
				t.Errorf("%s(%d, %d): ok=true, want false (div by zero)", opName, a, b)
			}
			return
		}
		ref = new(big.Int).Quo(big.NewInt(int64(a)), big.NewInt(int64(b)))
	default:
		t.Fatalf("unknown op: %s", opName)
		return
	}
	wantOk := ref.Cmp(lo) >= 0 && ref.Cmp(hi) <= 0
	if gotOk != wantOk {
		t.Errorf("%s(%d, %d): ok=%v, want ok=%v (ref=%s)", opName, a, b, gotOk, wantOk, ref)
	} else if gotOk {
		want := T(ref.Int64())
		if got != want {
			t.Errorf("%s(%d, %d): got=%d, want=%d", opName, a, b, got, want)
		}
	}
}

// checkUnsignedOp uses big.Int as reference for unsigned types.
func checkUnsignedOp[T Unsigned](t *testing.T, opName string, op func(T, T) (T, bool), a, b T, max *big.Int) {
	t.Helper()
	ba := new(big.Int).SetUint64(uint64(a))
	bb := new(big.Int).SetUint64(uint64(b))
	var ref *big.Int
	switch opName {
	case "Add":
		ref = new(big.Int).Add(ba, bb)
	case "Sub":
		ref = new(big.Int).Sub(ba, bb)
	case "Mul":
		ref = new(big.Int).Mul(ba, bb)
	case "Div":
		if b == 0 {
			got, gotOk := op(a, b)
			_ = got
			if gotOk {
				t.Errorf("%s(%d, %d): ok=true, want false (div by zero)", opName, a, b)
			}
			return
		}
		ref = new(big.Int).Quo(ba, bb)
	default:
		t.Fatalf("unknown op: %s", opName)
		return
	}
	got, gotOk := op(a, b)
	wantOk := ref.Sign() >= 0 && ref.Cmp(max) <= 0
	if gotOk != wantOk {
		t.Errorf("%s(%d, %d): ok=%v, want ok=%v (ref=%s)", opName, a, b, gotOk, wantOk, ref)
	} else if gotOk {
		want := T(ref.Uint64())
		if got != want {
			t.Errorf("%s(%d, %d): got=%d, want=%d", opName, a, b, got, want)
		}
	}
}

func TestBoundaryInt16(t *testing.T) {
	const minV, maxV = math.MinInt16, math.MaxInt16
	lo := big.NewInt(minV)
	hi := big.NewInt(maxV)
	pairs := [][2]int16{
		{minV, minV}, {minV, minV + 1}, {minV, -1}, {minV, 0}, {minV, 1}, {minV, maxV - 1}, {minV, maxV},
		{-1, -1}, {-1, 0}, {-1, 1}, {-1, maxV},
		{0, 0}, {0, 1}, {0, maxV},
		{1, 1}, {1, maxV - 1}, {1, maxV},
		{maxV - 1, maxV}, {maxV, maxV},
	}
	ops := []struct {
		name string
		fn   func(int16, int16) (int16, bool)
	}{
		{"Add", Add[int16]}, {"Sub", Sub[int16]}, {"Mul", Mul[int16]}, {"Div", Div[int16]},
	}
	for _, op := range ops {
		for _, p := range pairs {
			t.Run(fmt.Sprintf("%s/%d,%d", op.name, p[0], p[1]), func(t *testing.T) {
				checkSignedOp(t, op.name, op.fn, p[0], p[1], lo, hi)
			})
		}
	}
}

func TestBoundaryInt32(t *testing.T) {
	const minV, maxV = math.MinInt32, math.MaxInt32
	lo := big.NewInt(minV)
	hi := big.NewInt(maxV)
	pairs := [][2]int32{
		{minV, minV}, {minV, minV + 1}, {minV, -1}, {minV, 0}, {minV, 1}, {minV, maxV - 1}, {minV, maxV},
		{-1, -1}, {-1, 0}, {-1, 1}, {-1, maxV},
		{0, 0}, {0, 1}, {0, maxV},
		{1, 1}, {1, maxV - 1}, {1, maxV},
		{maxV - 1, maxV}, {maxV, maxV},
	}
	ops := []struct {
		name string
		fn   func(int32, int32) (int32, bool)
	}{
		{"Add", Add[int32]}, {"Sub", Sub[int32]}, {"Mul", Mul[int32]}, {"Div", Div[int32]},
	}
	for _, op := range ops {
		for _, p := range pairs {
			t.Run(fmt.Sprintf("%s/%d,%d", op.name, p[0], p[1]), func(t *testing.T) {
				checkSignedOp(t, op.name, op.fn, p[0], p[1], lo, hi)
			})
		}
	}
}

func TestBoundaryInt64(t *testing.T) {
	const minV, maxV = math.MinInt64, math.MaxInt64
	lo := new(big.Int).SetInt64(minV)
	hi := new(big.Int).SetInt64(maxV)
	pairs := [][2]int64{
		{minV, minV}, {minV, minV + 1}, {minV, -1}, {minV, 0}, {minV, 1}, {minV, maxV - 1}, {minV, maxV},
		{-1, -1}, {-1, 0}, {-1, 1}, {-1, maxV},
		{0, 0}, {0, 1}, {0, maxV},
		{1, 1}, {1, maxV - 1}, {1, maxV},
		{maxV - 1, maxV}, {maxV, maxV},
	}

	// For int64, use big.Int-based reference directly since values may overflow int64 intermediates.
	checkInt64 := func(t *testing.T, opName string, op func(int64, int64) (int64, bool), a, b int64) {
		t.Helper()
		got, gotOk := op(a, b)
		ba := new(big.Int).SetInt64(a)
		bb := new(big.Int).SetInt64(b)
		var ref *big.Int
		switch opName {
		case "Add":
			ref = new(big.Int).Add(ba, bb)
		case "Sub":
			ref = new(big.Int).Sub(ba, bb)
		case "Mul":
			ref = new(big.Int).Mul(ba, bb)
		case "Div":
			if b == 0 {
				if gotOk {
					t.Errorf("%s(%d, %d): ok=true, want false", opName, a, b)
				}
				return
			}
			ref = new(big.Int).Quo(ba, bb)
		}
		wantOk := ref.Cmp(lo) >= 0 && ref.Cmp(hi) <= 0
		if gotOk != wantOk {
			t.Errorf("%s(%d, %d): ok=%v, want ok=%v (ref=%s)", opName, a, b, gotOk, wantOk, ref)
		} else if gotOk {
			want := ref.Int64()
			if got != want {
				t.Errorf("%s(%d, %d): got=%d, want=%d", opName, a, b, got, want)
			}
		}
	}

	ops := []struct {
		name string
		fn   func(int64, int64) (int64, bool)
	}{
		{"Add", Add[int64]}, {"Sub", Sub[int64]}, {"Mul", Mul[int64]}, {"Div", Div[int64]},
	}
	for _, op := range ops {
		for _, p := range pairs {
			t.Run(fmt.Sprintf("%s/%d,%d", op.name, p[0], p[1]), func(t *testing.T) {
				checkInt64(t, op.name, op.fn, p[0], p[1])
			})
		}
	}
}

func TestBoundaryUint16(t *testing.T) {
	const maxV = math.MaxUint16
	bmax := new(big.Int).SetUint64(maxV)
	pairs := [][2]uint16{
		{0, 0}, {0, 1}, {0, maxV}, {1, 1}, {1, maxV - 1}, {1, maxV}, {maxV - 1, maxV}, {maxV, maxV},
	}
	ops := []struct {
		name string
		fn   func(uint16, uint16) (uint16, bool)
	}{
		{"Add", Add[uint16]}, {"Sub", Sub[uint16]}, {"Mul", Mul[uint16]}, {"Div", Div[uint16]},
	}
	for _, op := range ops {
		for _, p := range pairs {
			t.Run(fmt.Sprintf("%s/%d,%d", op.name, p[0], p[1]), func(t *testing.T) {
				checkUnsignedOp(t, op.name, op.fn, p[0], p[1], bmax)
			})
		}
	}
}

func TestBoundaryUint32(t *testing.T) {
	const maxV = math.MaxUint32
	bmax := new(big.Int).SetUint64(maxV)
	pairs := [][2]uint32{
		{0, 0}, {0, 1}, {0, maxV}, {1, 1}, {1, maxV - 1}, {1, maxV}, {maxV - 1, maxV}, {maxV, maxV},
	}
	ops := []struct {
		name string
		fn   func(uint32, uint32) (uint32, bool)
	}{
		{"Add", Add[uint32]}, {"Sub", Sub[uint32]}, {"Mul", Mul[uint32]}, {"Div", Div[uint32]},
	}
	for _, op := range ops {
		for _, p := range pairs {
			t.Run(fmt.Sprintf("%s/%d,%d", op.name, p[0], p[1]), func(t *testing.T) {
				checkUnsignedOp(t, op.name, op.fn, p[0], p[1], bmax)
			})
		}
	}
}

func TestBoundaryUint64(t *testing.T) {
	const maxV = math.MaxUint64
	bmax := new(big.Int).SetUint64(maxV)
	pairs := [][2]uint64{
		{0, 0}, {0, 1}, {0, maxV}, {1, 1}, {1, maxV - 1}, {1, maxV}, {maxV - 1, maxV}, {maxV, maxV},
	}

	checkUint64 := func(t *testing.T, opName string, op func(uint64, uint64) (uint64, bool), a, b uint64) {
		t.Helper()
		got, gotOk := op(a, b)
		ba := new(big.Int).SetUint64(a)
		bb := new(big.Int).SetUint64(b)
		var ref *big.Int
		switch opName {
		case "Add":
			ref = new(big.Int).Add(ba, bb)
		case "Sub":
			ref = new(big.Int).Sub(ba, bb)
		case "Mul":
			ref = new(big.Int).Mul(ba, bb)
		case "Div":
			if b == 0 {
				if gotOk {
					t.Errorf("%s(%d, %d): ok=true, want false", opName, a, b)
				}
				return
			}
			ref = new(big.Int).Quo(ba, bb)
		}
		wantOk := ref.Sign() >= 0 && ref.Cmp(bmax) <= 0
		if gotOk != wantOk {
			t.Errorf("%s(%d, %d): ok=%v, want ok=%v (ref=%s)", opName, a, b, gotOk, wantOk, ref)
		} else if gotOk {
			want := ref.Uint64()
			if got != want {
				t.Errorf("%s(%d, %d): got=%d, want=%d", opName, a, b, got, want)
			}
		}
	}

	ops := []struct {
		name string
		fn   func(uint64, uint64) (uint64, bool)
	}{
		{"Add", Add[uint64]}, {"Sub", Sub[uint64]}, {"Mul", Mul[uint64]}, {"Div", Div[uint64]},
	}
	for _, op := range ops {
		for _, p := range pairs {
			t.Run(fmt.Sprintf("%s/%d,%d", op.name, p[0], p[1]), func(t *testing.T) {
				checkUint64(t, op.name, op.fn, p[0], p[1])
			})
		}
	}
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
		{"int8/MinInt8", func() (int64, bool) { r, ok := Neg[int8](math.MinInt8); return int64(r), ok }, 0, false},
		{"int8/-1", func() (int64, bool) { r, ok := Neg[int8](-1); return int64(r), ok }, 1, true},
		{"int8/0", func() (int64, bool) { r, ok := Neg[int8](0); return int64(r), ok }, 0, true},
		{"int8/1", func() (int64, bool) { r, ok := Neg[int8](1); return int64(r), ok }, -1, true},
		{"int8/MaxInt8", func() (int64, bool) { r, ok := Neg[int8](math.MaxInt8); return int64(r), ok }, -math.MaxInt8, true},

		{"int16/MinInt16", func() (int64, bool) { r, ok := Neg[int16](math.MinInt16); return int64(r), ok }, 0, false},
		{"int16/-1", func() (int64, bool) { r, ok := Neg[int16](-1); return int64(r), ok }, 1, true},
		{"int16/0", func() (int64, bool) { r, ok := Neg[int16](0); return int64(r), ok }, 0, true},
		{"int16/1", func() (int64, bool) { r, ok := Neg[int16](1); return int64(r), ok }, -1, true},
		{"int16/MaxInt16", func() (int64, bool) { r, ok := Neg[int16](math.MaxInt16); return int64(r), ok }, -math.MaxInt16, true},

		{"int32/MinInt32", func() (int64, bool) { r, ok := Neg[int32](math.MinInt32); return int64(r), ok }, 0, false},
		{"int32/-1", func() (int64, bool) { r, ok := Neg[int32](-1); return int64(r), ok }, 1, true},
		{"int32/0", func() (int64, bool) { r, ok := Neg[int32](0); return int64(r), ok }, 0, true},
		{"int32/1", func() (int64, bool) { r, ok := Neg[int32](1); return int64(r), ok }, -1, true},
		{"int32/MaxInt32", func() (int64, bool) { r, ok := Neg[int32](math.MaxInt32); return int64(r), ok }, -math.MaxInt32, true},

		{"int64/MinInt64", func() (int64, bool) { r, ok := Neg[int64](math.MinInt64); return r, ok }, 0, false},
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
			if gotOk && got != tc.want {
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
		{"int8/MinInt8", func() (int64, bool) { r, ok := Abs[int8](math.MinInt8); return int64(r), ok }, 0, false},
		{"int8/-1", func() (int64, bool) { r, ok := Abs[int8](-1); return int64(r), ok }, 1, true},
		{"int8/0", func() (int64, bool) { r, ok := Abs[int8](0); return int64(r), ok }, 0, true},
		{"int8/1", func() (int64, bool) { r, ok := Abs[int8](1); return int64(r), ok }, 1, true},
		{"int8/MaxInt8", func() (int64, bool) { r, ok := Abs[int8](math.MaxInt8); return int64(r), ok }, math.MaxInt8, true},

		{"int16/MinInt16", func() (int64, bool) { r, ok := Abs[int16](math.MinInt16); return int64(r), ok }, 0, false},
		{"int16/-1", func() (int64, bool) { r, ok := Abs[int16](-1); return int64(r), ok }, 1, true},
		{"int16/0", func() (int64, bool) { r, ok := Abs[int16](0); return int64(r), ok }, 0, true},
		{"int16/MaxInt16", func() (int64, bool) { r, ok := Abs[int16](math.MaxInt16); return int64(r), ok }, math.MaxInt16, true},

		{"int32/MinInt32", func() (int64, bool) { r, ok := Abs[int32](math.MinInt32); return int64(r), ok }, 0, false},
		{"int32/-1", func() (int64, bool) { r, ok := Abs[int32](-1); return int64(r), ok }, 1, true},
		{"int32/0", func() (int64, bool) { r, ok := Abs[int32](0); return int64(r), ok }, 0, true},
		{"int32/MaxInt32", func() (int64, bool) { r, ok := Abs[int32](math.MaxInt32); return int64(r), ok }, math.MaxInt32, true},

		{"int64/MinInt64", func() (int64, bool) { r, ok := Abs[int64](math.MinInt64); return r, ok }, 0, false},
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
			if gotOk && got != tc.want {
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
		_, ok := MulDiv[int64](1, 2, 0)
		if ok {
			t.Error("MulDiv(1,2,0): ok=true, want false")
		}
	})

	t.Run("c_zero_uint", func(t *testing.T) {
		_, ok := MulDiv[uint64](1, 2, 0)
		if ok {
			t.Error("MulDiv(1,2,0): ok=true, want false")
		}
	})

	t.Run("MinInt64_neg1_1", func(t *testing.T) {
		// MinInt64 * -1 = |MinInt64| which overflows int64.
		_, ok := MulDiv[int64](math.MinInt64, -1, 1)
		if ok {
			t.Error("MulDiv(MinInt64,-1,1): ok=true, want false")
		}
	})

	t.Run("int8_100_50_25", func(t *testing.T) {
		// 100*50 = 5000, 5000/25 = 200, but 200 > MaxInt8 (127) so overflow.
		_, ok := MulDiv[int8](100, 50, 25)
		if ok {
			t.Error("MulDiv[int8](100,50,25): ok=true, want false (200 overflows int8)")
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

func TestMulMod(t *testing.T) {
	t.Run("MaxUint64_MaxUint64_MaxUint64", func(t *testing.T) {
		r, ok := MulMod[uint64](math.MaxUint64, math.MaxUint64, math.MaxUint64)
		if !ok || r != 0 {
			t.Errorf("MulMod(MaxUint64,MaxUint64,MaxUint64): got=(%d,%v), want=(0,true)", r, ok)
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
		_, ok := MulMod[int64](1, 2, 0)
		if ok {
			t.Error("MulMod(1,2,0): ok=true, want false")
		}
	})

	t.Run("c_zero_uint", func(t *testing.T) {
		_, ok := MulMod[uint64](1, 2, 0)
		if ok {
			t.Error("MulMod(1,2,0): ok=true, want false")
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
		_, ok := Pow[int8](2, 7)
		if ok {
			t.Error("Pow[int8](2,7): ok=true, want false (128 overflows int8)")
		}
	})
	t.Run("2^7_uint8_ok", func(t *testing.T) {
		r, ok := Pow[uint8](2, 7)
		if !ok || r != 128 {
			t.Errorf("Pow[uint8](2,7): got=(%d,%v), want=(128,true)", r, ok)
		}
	})
	t.Run("2^8_uint8_overflow", func(t *testing.T) {
		_, ok := Pow[uint8](2, 8)
		if ok {
			t.Error("Pow[uint8](2,8): ok=true, want false (256 overflows uint8)")
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
		_, ok := Pow[int64](2, 63)
		if ok {
			t.Error("Pow[int64](2,63): ok=true, want false")
		}
	})
	t.Run("2^63_uint64_ok", func(t *testing.T) {
		r, ok := Pow[uint64](2, 63)
		if !ok || r != 1<<63 {
			t.Errorf("Pow[uint64](2,63): got=(%d,%v), want=(%d,true)", r, ok, uint64(1<<63))
		}
	})
	t.Run("2^64_uint64_overflow", func(t *testing.T) {
		_, ok := Pow[uint64](2, 64)
		if ok {
			t.Error("Pow[uint64](2,64): ok=true, want false")
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
		_, ok := Pow[uint8](3, 6)
		if ok {
			t.Error("Pow[uint8](3,6): ok=true, want false (729 overflows uint8)")
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
		_, ok := Lsh[uint8](1, 8)
		if ok {
			t.Error("Lsh[uint8](1,8): ok=true, want false")
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
		_, ok := Lsh[uint64](1, 64)
		if ok {
			t.Error("Lsh[uint64](1,64): ok=true, want false")
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
	// Division by zero for various types.
	t.Run("div_by_zero_int8", func(t *testing.T) {
		_, ok := Div[int8](42, 0)
		if ok {
			t.Error("Div(42, 0): ok=true, want false")
		}
	})
	t.Run("div_by_zero_uint8", func(t *testing.T) {
		_, ok := Div[uint8](42, 0)
		if ok {
			t.Error("Div(42, 0): ok=true, want false")
		}
	})
	t.Run("div_by_zero_int64", func(t *testing.T) {
		_, ok := Div[int64](42, 0)
		if ok {
			t.Error("Div(42, 0): ok=true, want false")
		}
	})
	t.Run("div_by_zero_uint64", func(t *testing.T) {
		_, ok := Div[uint64](42, 0)
		if ok {
			t.Error("Div(42, 0): ok=true, want false")
		}
	})
	t.Run("divmod_by_zero_int64", func(t *testing.T) {
		_, _, ok := DivMod[int64](42, 0)
		if ok {
			t.Error("DivMod(42, 0): ok=true, want false")
		}
	})
	t.Run("mod_by_zero_int64", func(t *testing.T) {
		_, ok := Mod[int64](42, 0)
		if ok {
			t.Error("Mod(42, 0): ok=true, want false")
		}
	})
	t.Run("div_by_zero_0_int64", func(t *testing.T) {
		_, ok := Div[int64](0, 0)
		if ok {
			t.Error("Div(0, 0): ok=true, want false")
		}
	})

	// MinInt / -1 → Div overflow, Mod ok with result 0.
	t.Run("MinInt8_div_neg1", func(t *testing.T) {
		_, ok := Div[int8](math.MinInt8, -1)
		if ok {
			t.Error("Div(MinInt8, -1): ok=true, want false")
		}
	})
	t.Run("MinInt8_mod_neg1", func(t *testing.T) {
		r, ok := Mod[int8](math.MinInt8, -1)
		if !ok || r != 0 {
			t.Errorf("Mod(MinInt8, -1): got=(%d,%v), want=(0,true)", r, ok)
		}
	})
	t.Run("MinInt16_div_neg1", func(t *testing.T) {
		_, ok := Div[int16](math.MinInt16, -1)
		if ok {
			t.Error("Div(MinInt16, -1): ok=true, want false")
		}
	})
	t.Run("MinInt16_mod_neg1", func(t *testing.T) {
		r, ok := Mod[int16](math.MinInt16, -1)
		if !ok || r != 0 {
			t.Errorf("Mod(MinInt16, -1): got=(%d,%v), want=(0,true)", r, ok)
		}
	})
	t.Run("MinInt32_div_neg1", func(t *testing.T) {
		_, ok := Div[int32](math.MinInt32, -1)
		if ok {
			t.Error("Div(MinInt32, -1): ok=true, want false")
		}
	})
	t.Run("MinInt32_mod_neg1", func(t *testing.T) {
		r, ok := Mod[int32](math.MinInt32, -1)
		if !ok || r != 0 {
			t.Errorf("Mod(MinInt32, -1): got=(%d,%v), want=(0,true)", r, ok)
		}
	})
	t.Run("MinInt64_div_neg1", func(t *testing.T) {
		_, ok := Div[int64](math.MinInt64, -1)
		if ok {
			t.Error("Div(MinInt64, -1): ok=true, want false")
		}
	})
	t.Run("MinInt64_mod_neg1", func(t *testing.T) {
		r, ok := Mod[int64](math.MinInt64, -1)
		if !ok || r != 0 {
			t.Errorf("Mod(MinInt64, -1): got=(%d,%v), want=(0,true)", r, ok)
		}
	})

	// DivMod for MinInt/-1: q overflows, r=0.
	t.Run("MinInt8_divmod_neg1", func(t *testing.T) {
		_, _, ok := DivMod[int8](math.MinInt8, -1)
		if ok {
			t.Error("DivMod(MinInt8, -1): ok=true, want false")
		}
	})
	t.Run("MinInt64_divmod_neg1", func(t *testing.T) {
		_, _, ok := DivMod[int64](math.MinInt64, -1)
		if ok {
			t.Error("DivMod(MinInt64, -1): ok=true, want false")
		}
	})

	// Truncation to zero: (1, -3), (-1, 3), (1, MaxInt) → ok=true.
	// This tests the bug fix: q==0 should be ok.
	t.Run("truncation_1_neg3", func(t *testing.T) {
		q, ok := Div[int64](1, -3)
		if !ok || q != 0 {
			t.Errorf("Div(1, -3): got=(%d,%v), want=(0,true)", q, ok)
		}
	})
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

func expectPanic(t *testing.T, name string, fn func()) {
	t.Helper()
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("%s: expected panic but did not panic", name)
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
		expectPanic(t, "MustAdd(MaxInt64,1)", func() {
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
		expectPanic(t, "MustSub(MinInt64,1)", func() {
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
		expectPanic(t, "MustMul(MaxInt64,2)", func() {
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
		expectPanic(t, "MustDiv(1,0)", func() {
			MustDiv[int64](1, 0)
		})
	})
	t.Run("MinInt64_neg1", func(t *testing.T) {
		expectPanic(t, "MustDiv(MinInt64,-1)", func() {
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
		expectPanic(t, "MustNeg(MinInt64)", func() {
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
		expectPanic(t, "MustMulDiv(1,2,0)", func() {
			MustMulDiv[int64](1, 2, 0)
		})
	})
	t.Run("overflow_result", func(t *testing.T) {
		expectPanic(t, "MustMulDiv(MinInt64,-1,1)", func() {
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
		expectPanic(t, "MustConvert[int64,int8](128)", func() {
			MustConvert[int64, int8](128)
		})
	})
	t.Run("sign_overflow", func(t *testing.T) {
		expectPanic(t, "MustConvert[int64,uint64](-1)", func() {
			MustConvert[int64, uint64](-1)
		})
	})
}

// ---------------------------------------------------------------------------
// J. Property-based tests
// ---------------------------------------------------------------------------

func TestPropertyCommutativity(t *testing.T) {
	signedVals := []int64{
		math.MinInt64, math.MinInt64 + 1, -1000, -1, 0, 1, 1000, math.MaxInt64 - 1, math.MaxInt64,
	}
	unsignedVals := []uint64{
		0, 1, 1000, math.MaxUint64/2 - 1, math.MaxUint64 / 2, math.MaxUint64 - 1, math.MaxUint64,
	}

	t.Run("Add_int64", func(t *testing.T) {
		for _, a := range signedVals {
			for _, b := range signedVals {
				r1, ok1 := Add(a, b)
				r2, ok2 := Add(b, a)
				if ok1 != ok2 || (ok1 && r1 != r2) {
					t.Errorf("Add(%d,%d) != Add(%d,%d): (%d,%v) vs (%d,%v)", a, b, b, a, r1, ok1, r2, ok2)
				}
			}
		}
	})

	t.Run("Mul_int64", func(t *testing.T) {
		for _, a := range signedVals {
			for _, b := range signedVals {
				r1, ok1 := Mul(a, b)
				r2, ok2 := Mul(b, a)
				if ok1 != ok2 || (ok1 && r1 != r2) {
					t.Errorf("Mul(%d,%d) != Mul(%d,%d): (%d,%v) vs (%d,%v)", a, b, b, a, r1, ok1, r2, ok2)
				}
			}
		}
	})

	t.Run("Add_uint64", func(t *testing.T) {
		for _, a := range unsignedVals {
			for _, b := range unsignedVals {
				r1, ok1 := Add(a, b)
				r2, ok2 := Add(b, a)
				if ok1 != ok2 || (ok1 && r1 != r2) {
					t.Errorf("Add(%d,%d) != Add(%d,%d): (%d,%v) vs (%d,%v)", a, b, b, a, r1, ok1, r2, ok2)
				}
			}
		}
	})

	t.Run("Mul_uint64", func(t *testing.T) {
		for _, a := range unsignedVals {
			for _, b := range unsignedVals {
				r1, ok1 := Mul(a, b)
				r2, ok2 := Mul(b, a)
				if ok1 != ok2 || (ok1 && r1 != r2) {
					t.Errorf("Mul(%d,%d) != Mul(%d,%d): (%d,%v) vs (%d,%v)", a, b, b, a, r1, ok1, r2, ok2)
				}
			}
		}
	})
}

func TestPropertyIdentity(t *testing.T) {
	signedVals := []int64{
		math.MinInt64, math.MinInt64 + 1, -1000, -1, 0, 1, 1000, math.MaxInt64 - 1, math.MaxInt64,
	}
	unsignedVals := []uint64{
		0, 1, 1000, math.MaxUint64 / 2, math.MaxUint64 - 1, math.MaxUint64,
	}

	// Add(a, 0) == (a, true)
	t.Run("Add_zero_int64", func(t *testing.T) {
		for _, a := range signedVals {
			r, ok := Add(a, int64(0))
			if !ok || r != a {
				t.Errorf("Add(%d, 0): got=(%d,%v), want=(%d,true)", a, r, ok, a)
			}
		}
	})

	t.Run("Add_zero_uint64", func(t *testing.T) {
		for _, a := range unsignedVals {
			r, ok := Add(a, uint64(0))
			if !ok || r != a {
				t.Errorf("Add(%d, 0): got=(%d,%v), want=(%d,true)", a, r, ok, a)
			}
		}
	})

	// Mul(a, 1) == (a, true)
	t.Run("Mul_one_int64", func(t *testing.T) {
		for _, a := range signedVals {
			r, ok := Mul(a, int64(1))
			if !ok || r != a {
				t.Errorf("Mul(%d, 1): got=(%d,%v), want=(%d,true)", a, r, ok, a)
			}
		}
	})

	t.Run("Mul_one_uint64", func(t *testing.T) {
		for _, a := range unsignedVals {
			r, ok := Mul(a, uint64(1))
			if !ok || r != a {
				t.Errorf("Mul(%d, 1): got=(%d,%v), want=(%d,true)", a, r, ok, a)
			}
		}
	})
}

func TestPropertyZeroMul(t *testing.T) {
	signedVals := []int64{
		math.MinInt64, math.MinInt64 + 1, -1000, -1, 0, 1, 1000, math.MaxInt64 - 1, math.MaxInt64,
	}
	unsignedVals := []uint64{
		0, 1, 1000, math.MaxUint64 / 2, math.MaxUint64 - 1, math.MaxUint64,
	}

	// Mul(a, 0) == (0, true)
	t.Run("int64", func(t *testing.T) {
		for _, a := range signedVals {
			r, ok := Mul(a, int64(0))
			if !ok || r != 0 {
				t.Errorf("Mul(%d, 0): got=(%d,%v), want=(0,true)", a, r, ok)
			}
		}
	})

	t.Run("uint64", func(t *testing.T) {
		for _, a := range unsignedVals {
			r, ok := Mul(a, uint64(0))
			if !ok || r != 0 {
				t.Errorf("Mul(%d, 0): got=(%d,%v), want=(0,true)", a, r, ok)
			}
		}
	})
}

func TestPropertyInverse(t *testing.T) {
	// If Neg(a) succeeds, then Add(a, Neg(a)) should succeed and return 0.
	signedVals := []int64{
		math.MinInt64, math.MinInt64 + 1, -1000, -1, 0, 1, 1000, math.MaxInt64 - 1, math.MaxInt64,
	}
	for _, a := range signedVals {
		neg, negOk := Neg(a)
		if !negOk {
			continue
		}
		sum, sumOk := Add(a, neg)
		if !sumOk || sum != 0 {
			t.Errorf("Add(%d, Neg(%d)=%d): got=(%d,%v), want=(0,true)", a, a, neg, sum, sumOk)
		}
	}
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
