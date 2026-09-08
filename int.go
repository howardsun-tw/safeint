package safeint

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"strconv"
)

// Int is a generic wrapper around Go's native integer types, providing
// method-based arithmetic with overflow detection.
//
// It uses value semantics — all methods return new values, never mutate.
// Inspired by holiman/uint256's dual API: wrapping methods (Add, Sub, Mul)
// that silently wrap on overflow, and checked methods (*Overflow) that
// report overflow via a bool return.
type Int[T Integer] struct {
	val T
}

// ---------------------------------------------------------------------------
// Constructors
// ---------------------------------------------------------------------------

// New creates an Int[T] from a raw value.
func New[T Integer](v T) Int[T] {
	return Int[T]{val: v}
}

// Zero returns the zero value of Int[T].
func Zero[T Integer]() Int[T] {
	return Int[T]{}
}

// ---------------------------------------------------------------------------
// Checked methods — return (result, overflow bool)
// ---------------------------------------------------------------------------

// AddOverflow returns a + b and true if the result does not overflow.
func (a Int[T]) AddOverflow(b Int[T]) (Int[T], bool) {
	r, ok := Add(a.val, b.val)
	return Int[T]{val: r}, ok
}

// SubOverflow returns a - b and true if the result does not overflow.
func (a Int[T]) SubOverflow(b Int[T]) (Int[T], bool) {
	r, ok := Sub(a.val, b.val)
	return Int[T]{val: r}, ok
}

// MulOverflow returns a * b and true if the result does not overflow.
func (a Int[T]) MulOverflow(b Int[T]) (Int[T], bool) {
	r, ok := Mul(a.val, b.val)
	return Int[T]{val: r}, ok
}

// DivOverflow returns a / b and true if the result does not overflow.
func (a Int[T]) DivOverflow(b Int[T]) (Int[T], bool) {
	r, ok := Div(a.val, b.val)
	return Int[T]{val: r}, ok
}

// DivModOverflow returns (a/b, a%b) and true if the quotient does not overflow.
func (a Int[T]) DivModOverflow(b Int[T]) (Int[T], Int[T], bool) {
	q, r, ok := DivMod(a.val, b.val)
	return Int[T]{val: q}, Int[T]{val: r}, ok
}

// ModOverflow returns a % b and true. Returns (0, false) on division by zero.
func (a Int[T]) ModOverflow(b Int[T]) (Int[T], bool) {
	r, ok := Mod(a.val, b.val)
	return Int[T]{val: r}, ok
}

// NegOverflow returns -a and true if the result does not overflow.
func (a Int[T]) NegOverflow() (Int[T], bool) {
	r, ok := Neg(a.val)
	return Int[T]{val: r}, ok
}

// AbsOverflow returns |a| and true if the result does not overflow.
func (a Int[T]) AbsOverflow() (Int[T], bool) {
	r, ok := Abs(a.val)
	return Int[T]{val: r}, ok
}

// PowOverflow returns a^exp and true if no overflow occurs.
func (a Int[T]) PowOverflow(exp uint) (Int[T], bool) {
	r, ok := Pow(a.val, exp)
	return Int[T]{val: r}, ok
}

// LshOverflow returns the result of Lsh on the underlying value as an Int[T].
// The bool is true on success and false on overflow. A zero value always
// succeeds; a nonzero value shifted by at least the bit width of T returns
// zero and false. Other overflow results are not guaranteed to be zero.
func (a Int[T]) LshOverflow(n uint) (Int[T], bool) {
	r, ok := Lsh(a.val, n)
	return Int[T]{val: r}, ok
}

// MulDivOverflow returns (a*b)/c with full intermediate precision.
func (a Int[T]) MulDivOverflow(b, c Int[T]) (Int[T], bool) {
	r, ok := MulDiv(a.val, b.val, c.val)
	return Int[T]{val: r}, ok
}

// MulModOverflow returns (a*b)%c with full intermediate precision.
func (a Int[T]) MulModOverflow(b, c Int[T]) (Int[T], bool) {
	r, ok := MulMod(a.val, b.val, c.val)
	return Int[T]{val: r}, ok
}

// ---------------------------------------------------------------------------
// Wrapping methods — silently wrap on overflow
// ---------------------------------------------------------------------------

// Add returns a + b, wrapping on overflow.
func (a Int[T]) Add(b Int[T]) Int[T] {
	return Int[T]{val: a.val + b.val}
}

// Sub returns a - b, wrapping on overflow.
func (a Int[T]) Sub(b Int[T]) Int[T] {
	return Int[T]{val: a.val - b.val}
}

// Mul returns a * b, wrapping on overflow.
func (a Int[T]) Mul(b Int[T]) Int[T] {
	return Int[T]{val: a.val * b.val}
}

// ---------------------------------------------------------------------------
// Must methods — panic on overflow
// ---------------------------------------------------------------------------

// MustAdd returns a + b, panicking on overflow.
func (a Int[T]) MustAdd(b Int[T]) Int[T] {
	return Int[T]{val: MustAdd(a.val, b.val)}
}

// MustSub returns a - b, panicking on overflow.
func (a Int[T]) MustSub(b Int[T]) Int[T] {
	return Int[T]{val: MustSub(a.val, b.val)}
}

// MustMul returns a * b, panicking on overflow.
func (a Int[T]) MustMul(b Int[T]) Int[T] {
	return Int[T]{val: MustMul(a.val, b.val)}
}

// MustDiv returns a / b, panicking on overflow or division by zero.
func (a Int[T]) MustDiv(b Int[T]) Int[T] {
	return Int[T]{val: MustDiv(a.val, b.val)}
}

// ---------------------------------------------------------------------------
// Comparison
// ---------------------------------------------------------------------------

// Cmp compares a and b, returning -1, 0, or +1.
func (a Int[T]) Cmp(b Int[T]) int {
	switch {
	case a.val < b.val:
		return -1
	case a.val > b.val:
		return 1
	default:
		return 0
	}
}

// Eq returns true if a == b.
func (a Int[T]) Eq(b Int[T]) bool { return a.val == b.val }

// Lt returns true if a < b.
func (a Int[T]) Lt(b Int[T]) bool { return a.val < b.val }

// Gt returns true if a > b.
func (a Int[T]) Gt(b Int[T]) bool { return a.val > b.val }

// Lte returns true if a <= b.
func (a Int[T]) Lte(b Int[T]) bool { return a.val <= b.val }

// Gte returns true if a >= b.
func (a Int[T]) Gte(b Int[T]) bool { return a.val >= b.val }

// IsZero returns true if a == 0.
func (a Int[T]) IsZero() bool { return a.val == 0 }

// ---------------------------------------------------------------------------
// Value access
// ---------------------------------------------------------------------------

// Val returns the underlying raw value.
func (a Int[T]) Val() T { return a.val }

// String implements fmt.Stringer.
func (a Int[T]) String() string { return fmt.Sprint(a.val) }

// ---------------------------------------------------------------------------
// Conversion (standalone — Go methods cannot have extra type parameters)
// ---------------------------------------------------------------------------

// ConvertInt converts an Int[T] to Int[U], returning true if the value is
// preserved exactly.
func ConvertInt[T Integer, U Integer](a Int[T]) (Int[U], bool) {
	r, ok := Convert[T, U](a.val)
	return Int[U]{val: r}, ok
}

// ---------------------------------------------------------------------------
// Serialization — JSON
// ---------------------------------------------------------------------------

// MarshalJSON implements json.Marshaler using the underlying integer value.
// Custom JSON methods on T are not used.
func (a Int[T]) MarshalJSON() ([]byte, error) {
	if isSigned[T]() {
		return strconv.AppendInt(nil, int64(a.val), 10), nil
	}
	return strconv.AppendUint(nil, uint64(a.val), 10), nil
}

// UnmarshalJSON implements json.Unmarshaler using the underlying integer value.
// Custom JSON methods on T are not used. Null and invalid input leave the value unchanged.
func (a *Int[T]) UnmarshalJSON(data []byte) error {
	var value T
	var ok bool
	if isSigned[T]() {
		n := int64(a.val)
		if err := json.Unmarshal(data, &n); err != nil {
			return err
		}
		value, ok = Convert[int64, T](n)
	} else {
		n := uint64(a.val)
		if err := json.Unmarshal(data, &n); err != nil {
			return err
		}
		value, ok = Convert[uint64, T](n)
	}
	if !ok {
		return &json.UnmarshalTypeError{Value: "number " + string(data), Type: reflect.TypeOf(a.val)}
	}
	a.val = value
	return nil
}

// ---------------------------------------------------------------------------
// Serialization — database/sql
// ---------------------------------------------------------------------------

// Value implements driver.Valuer.
// Returns the underlying value as int64, or as a decimal string if it
// overflows int64 (e.g. uint64 > MaxInt64).
func (a Int[T]) Value() (driver.Value, error) {
	v, ok := Convert[T, int64](a.val)
	if !ok {
		return strconv.FormatUint(uint64(a.val), 10), nil
	}
	return v, nil
}

// Scan implements sql.Scanner.
// Accepts int64, float64 (whole numbers only), []byte, string, or nil.
func (a *Int[T]) Scan(src interface{}) error {
	if src == nil {
		a.val = 0
		return nil
	}
	switch v := src.(type) {
	case int64:
		r, ok := Convert[int64, T](v)
		if !ok {
			return fmt.Errorf("safeint: value %d overflows %T", v, a.val)
		}
		a.val = r
		return nil
	case float64:
		if math.IsNaN(v) || math.IsInf(v, 0) || math.Trunc(v) != v {
			return fmt.Errorf("safeint: float64 value %g is not an integer", v)
		}
		var r T
		var ok bool
		// Check exact, exclusive upper bounds before converting. MaxInt64
		// and MaxUint64 round up to these bounds when converted to float64.
		if isSigned[T]() {
			if v < -0x1p63 || v >= 0x1p63 {
				return fmt.Errorf("safeint: value %g overflows %T", v, a.val)
			}
			r, ok = Convert[int64, T](int64(v))
		} else {
			if v < 0 || v >= 0x1p64 {
				return fmt.Errorf("safeint: value %g overflows %T", v, a.val)
			}
			r, ok = Convert[uint64, T](uint64(v))
		}
		if !ok {
			return fmt.Errorf("safeint: value %g overflows %T", v, a.val)
		}
		a.val = r
		return nil
	case []byte:
		return a.scanString(string(v))
	case string:
		return a.scanString(v)
	default:
		return fmt.Errorf("safeint: unsupported Scan source type %T", src)
	}
}

func (a *Int[T]) scanString(s string) error {
	if isSigned[T]() {
		i, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return fmt.Errorf("safeint: cannot scan %q: %w", s, err)
		}
		r, ok := Convert[int64, T](i)
		if !ok {
			return fmt.Errorf("safeint: value %d overflows %T", i, a.val)
		}
		a.val = r
		return nil
	}
	u, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return fmt.Errorf("safeint: cannot scan %q: %w", s, err)
	}
	r, ok := Convert[uint64, T](u)
	if !ok {
		return fmt.Errorf("safeint: value %d overflows %T", u, a.val)
	}
	a.val = r
	return nil
}
