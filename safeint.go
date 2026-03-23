// Package safeint provides overflow-checked arithmetic for Go's native integer types.
//
// It offers three API styles:
//   - Checked functions: Add, Sub, Mul, etc. return (result, ok) where ok is false on overflow.
//   - Must functions: MustAdd, MustSub, etc. panic on overflow. Do not use with untrusted input.
//   - Int[T] wrapper type: provides method-based API with checked, wrapping, and must variants.
//
// All algorithms use data-dependent branches and are NOT constant-time.
// Do not use in cryptographic contexts where timing side-channels matter.
package safeint

import (
	"math/bits"
	"unsafe"
)

// Integer is a constraint that permits any Go built-in integer type (including named types).
type Integer interface {
	Signed | Unsigned
}

// Signed is a constraint that permits any signed integer type.
type Signed interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64
}

// Unsigned is a constraint that permits any unsigned integer type.
type Unsigned interface {
	~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64
}

// ---------------------------------------------------------------------------
// Checked arithmetic
// ---------------------------------------------------------------------------

// Add returns a + b and true if the result does not overflow.
func Add[T Integer](a, b T) (T, bool) {
	c := a + b
	if (c > a) == (b > 0) {
		return c, true
	}
	return c, false
}

// Sub returns a - b and true if the result does not overflow.
func Sub[T Integer](a, b T) (T, bool) {
	c := a - b
	if (c < a) == (b > 0) {
		return c, true
	}
	return c, false
}

// Mul returns a * b and true if the result does not overflow.
// Both a sign check and a division-roundtrip check are required:
// the sign check catches MinInt * -1; the roundtrip catches magnitude overflow.
func Mul[T Integer](a, b T) (T, bool) {
	if a == 0 || b == 0 {
		return 0, true
	}
	c := a * b
	// For unsigned T: c<0 is always false and (a<0)!=(b<0) is always false,
	// so this evaluates to false==false → true, falling through to the
	// roundtrip check c/b==a which is the sole unsigned overflow detector.
	if (c < 0) == ((a < 0) != (b < 0)) {
		if c/b == a {
			return c, true
		}
	}
	return c, false
}

// Div returns a / b and true if the result does not overflow.
// Returns (0, false) on division by zero or signed MinInt / -1 overflow.
func Div[T Integer](a, b T) (T, bool) {
	if b == 0 {
		return 0, false
	}
	if isSigned[T]() && a == minValue[T]() && b == ^T(0) {
		return 0, false
	}
	q := a / b
	ok := q == 0 || (q < 0) == ((a < 0) != (b < 0))
	return q, ok
}

// DivMod returns (a/b, a%b) and true if the quotient does not overflow.
// Returns (0, 0, false) on division by zero or signed MinInt / -1 overflow.
func DivMod[T Integer](a, b T) (T, T, bool) {
	if b == 0 {
		return 0, 0, false
	}
	// Defense-in-depth: explicit MinInt / -1 check.
	// Go guarantees wrapping here (no panic), but we detect it as overflow.
	// ^T(0) is -1 for signed types (all bits set).
	if isSigned[T]() && a == minValue[T]() && b == ^T(0) {
		return 0, 0, false
	}
	q := a / b
	// q == 0 handles truncation-to-zero with mixed signs (fixes g-utils/overflow bug).
	ok := q == 0 || (q < 0) == ((a < 0) != (b < 0))
	return q, a % b, ok
}

// Mod returns a % b and true. Returns (0, false) only on division by zero.
// The remainder never overflows; even MinInt % -1 == 0 is correct in Go.
func Mod[T Integer](a, b T) (T, bool) {
	if b == 0 {
		return 0, false
	}
	return a % b, true
}

// Neg returns -a and true if the result does not overflow.
// Overflows for signed MinInt and all non-zero unsigned values.
func Neg[T Integer](a T) (T, bool) {
	return Sub(T(0), a)
}

// Abs returns |a| and true if the result does not overflow.
// Overflows only for signed MinInt (whose absolute value exceeds MaxInt).
// For unsigned types, always returns (a, true).
func Abs[T Integer](a T) (T, bool) {
	if a < 0 {
		return Neg(a)
	}
	return a, true
}

// Pow returns base^exp and true if no intermediate or final overflow occurs.
// Uses binary exponentiation. 0^0 returns (1, true) by convention.
func Pow[T Integer](base T, exp uint) (T, bool) {
	if exp == 0 {
		return 1, true
	}
	if base == 1 {
		return 1, true
	}
	result := T(1)
	b := base
	for exp > 0 {
		if exp&1 == 1 {
			var ok bool
			result, ok = Mul(result, b)
			if !ok {
				return result, false
			}
		}
		exp >>= 1
		if exp > 0 {
			var ok bool
			b, ok = Mul(b, b)
			if !ok {
				return result, false
			}
		}
	}
	return result, true
}

// Lsh returns a << n and true if no bits are lost.
// Always returns (0, true) when a == 0.
func Lsh[T Integer](a T, n uint) (T, bool) {
	if a == 0 {
		return 0, true
	}
	bitsz := uint(unsafe.Sizeof(a)) * 8
	if n >= bitsz {
		return 0, false
	}
	c := a << n
	if (c >> n) != a {
		return c, false
	}
	return c, true
}

// Convert converts value a of type T to type U, returning true if the value
// is preserved exactly (no truncation or sign change).
func Convert[T Integer, U Integer](a T) (U, bool) {
	b := U(a)
	if T(b) == a && (a < 0) == (b < 0) {
		return b, true
	}
	return b, false
}

// ---------------------------------------------------------------------------
// MulDiv / MulMod — full-precision intermediate
// ---------------------------------------------------------------------------

// MulDiv returns (a*b)/c computed with full intermediate precision (no
// intermediate overflow). Returns (0, false) when c == 0 or the quotient
// overflows T.
func MulDiv[T Integer](a, b, c T) (T, bool) {
	if c == 0 {
		return 0, false
	}
	bitsz := uint(unsafe.Sizeof(a)) * 8
	if bitsz <= 32 {
		return mulDivSmall(a, b, c)
	}
	return mulDiv64(a, b, c)
}

// MulMod returns (a*b)%c computed with full intermediate precision.
// Returns (0, false) when c == 0.
func MulMod[T Integer](a, b, c T) (T, bool) {
	if c == 0 {
		return 0, false
	}
	bitsz := uint(unsafe.Sizeof(a)) * 8
	if bitsz <= 32 {
		return mulModSmall(a, b, c)
	}
	return mulMod64(a, b, c)
}

// mulDivSmall handles MulDiv for types that fit in 64-bit intermediates.
func mulDivSmall[T Integer](a, b, c T) (T, bool) {
	if isSigned[T]() {
		wide := int64(a) * int64(b)
		q := wide / int64(c)
		return Convert[int64, T](q)
	}
	wide := uint64(a) * uint64(b)
	q := wide / uint64(c)
	return Convert[uint64, T](q)
}

// mulModSmall handles MulMod for types that fit in 64-bit intermediates.
func mulModSmall[T Integer](a, b, c T) (T, bool) {
	if isSigned[T]() {
		wide := int64(a) * int64(b)
		r := wide % int64(c)
		return Convert[int64, T](r)
	}
	wide := uint64(a) * uint64(b)
	r := wide % uint64(c)
	return Convert[uint64, T](r)
}

// mulDiv64 handles MulDiv for 64-bit types using math/bits for 128-bit math.
func mulDiv64[T Integer](a, b, c T) (T, bool) {
	if isSigned[T]() {
		return mulDiv64Signed(a, b, c)
	}
	return mulDiv64Unsigned(a, b, c)
}

func mulDiv64Unsigned[T Integer](a, b, c T) (T, bool) {
	ua, ub, uc := uint64(a), uint64(b), uint64(c)
	hi, lo := bits.Mul64(ua, ub)
	if hi >= uc {
		return 0, false // quotient overflows uint64
	}
	quo, _ := bits.Div64(hi, lo, uc)
	return T(quo), true
}

func mulDiv64Signed[T Integer](a, b, c T) (T, bool) {
	negative := (a < 0) != (b < 0) != (c < 0)
	ua := toUint64Abs(a)
	ub := toUint64Abs(b)
	uc := toUint64Abs(c)

	hi, lo := bits.Mul64(ua, ub)
	if hi >= uc {
		return 0, false
	}
	quo, _ := bits.Div64(hi, lo, uc)

	if negative {
		// Result must fit as negative T: quo <= uint64(maxValue[int64]()) + 1
		if quo > uint64(maxSigned64)+1 {
			return 0, false
		}
		return T(-int64(quo)), true
	}
	if quo > uint64(maxSigned64) {
		return 0, false
	}
	return T(int64(quo)), true
}

// mulMod64 handles MulMod for 64-bit types.
func mulMod64[T Integer](a, b, c T) (T, bool) {
	if isSigned[T]() {
		return mulMod64Signed(a, b, c)
	}
	return mulMod64Unsigned(a, b, c)
}

func mulMod64Unsigned[T Integer](a, b, c T) (T, bool) {
	ua, ub, uc := uint64(a), uint64(b), uint64(c)
	hi, lo := bits.Mul64(ua, ub)
	// CRITICAL: reduce hi to prevent bits.Div64 panic.
	// (hi*2^64+lo) mod uc == ((hi mod uc)*2^64+lo) mod uc
	hi = hi % uc
	_, rem := bits.Div64(hi, lo, uc)
	return T(rem), true
}

func mulMod64Signed[T Integer](a, b, c T) (T, bool) {
	// The sign of a%b in Go follows the dividend (a).
	// For (a*b)%c, the sign follows a*b, i.e. (a<0) XOR (b<0).
	negative := (a < 0) != (b < 0)
	ua := toUint64Abs(a)
	ub := toUint64Abs(b)
	uc := toUint64Abs(c)

	hi, lo := bits.Mul64(ua, ub)
	hi = hi % uc
	_, rem := bits.Div64(hi, lo, uc)

	if negative && rem != 0 {
		if rem > uint64(maxSigned64)+1 {
			return 0, false
		}
		return T(-int64(rem)), true
	}
	if rem > uint64(maxSigned64) {
		return 0, false
	}
	return T(int64(rem)), true
}

// ---------------------------------------------------------------------------
// Must variants — panic on overflow
// ---------------------------------------------------------------------------

// MustAdd returns a + b, panicking on overflow.
// WARNING: Do not use with untrusted input in server contexts.
func MustAdd[T Integer](a, b T) T {
	r, ok := Add(a, b)
	if !ok {
		panic("safeint: Add overflow")
	}
	return r
}

// MustSub returns a - b, panicking on overflow.
func MustSub[T Integer](a, b T) T {
	r, ok := Sub(a, b)
	if !ok {
		panic("safeint: Sub overflow")
	}
	return r
}

// MustMul returns a * b, panicking on overflow.
func MustMul[T Integer](a, b T) T {
	r, ok := Mul(a, b)
	if !ok {
		panic("safeint: Mul overflow")
	}
	return r
}

// MustDiv returns a / b, panicking on overflow or division by zero.
func MustDiv[T Integer](a, b T) T {
	r, ok := Div(a, b)
	if !ok {
		panic("safeint: Div overflow")
	}
	return r
}

// MustNeg returns -a, panicking on overflow.
func MustNeg[T Integer](a T) T {
	r, ok := Neg(a)
	if !ok {
		panic("safeint: Neg overflow")
	}
	return r
}

// MustMulDiv returns (a*b)/c, panicking on overflow or division by zero.
func MustMulDiv[T Integer](a, b, c T) T {
	r, ok := MulDiv(a, b, c)
	if !ok {
		panic("safeint: MulDiv overflow")
	}
	return r
}

// MustConvert converts a from T to U, panicking if the value cannot be represented.
func MustConvert[T Integer, U Integer](a T) U {
	r, ok := Convert[T, U](a)
	if !ok {
		panic("safeint: Convert overflow")
	}
	return r
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func isSigned[T Integer]() bool {
	// ^T(0) is all-ones: -1 for signed (negative), MaxT for unsigned (positive).
	return ^T(0) < 0
}

func minValue[T Integer]() T {
	if !isSigned[T]() {
		return 0
	}
	// T(1) shifted to sign bit position wraps to MinInt for signed types.
	return T(1) << (unsafe.Sizeof(T(0))*8 - 1)
}

const maxSigned64 = int64(1<<63 - 1)

// toUint64Abs returns the absolute value of a as uint64.
// Go's two's complement conversion handles MinInt correctly:
// uint64(-MinInt64) == 1 << 63.
func toUint64Abs[T Integer](a T) uint64 {
	if a < 0 {
		return uint64(-a)
	}
	return uint64(a)
}
