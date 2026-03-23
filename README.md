# safeint

Overflow-checked integer arithmetic for Go, built on generics.

Go's native integer operations silently wrap on overflow. `safeint` makes overflow detection explicit — you decide whether to check, panic, or wrap.

## Install

```
go get github.com/howardsun-tw/safeint
```

Requires Go 1.18+ (generics). Tested on Go 1.18 through 1.26.

## API Overview

The package provides three API styles for different use cases:

### 1. Checked Functions — `(result, ok)`

Functions return the result and a boolean. `ok` is `false` when overflow (or division by zero) occurs.

```go
sum, ok := safeint.Add(a, b)
if !ok {
    // handle overflow
}

product, ok := safeint.Mul(a, b)
diff, ok := safeint.Sub(a, b)
quotient, ok := safeint.Div(a, b)
quotient, remainder, ok := safeint.DivMod(a, b)
remainder, ok := safeint.Mod(a, b)
negated, ok := safeint.Neg(a)
absolute, ok := safeint.Abs(a)
power, ok := safeint.Pow(base, exp)
shifted, ok := safeint.Lsh(a, n)
converted, ok := safeint.Convert[int64, int32](a)
```

All functions are generic over any Go integer type (`int`, `int8`...`int64`, `uint`, `uint8`...`uint64`, and named types based on them).

### 2. Must Functions — panic on overflow

Convenience wrappers that panic instead of returning a bool. Useful for cases where overflow is a programming error.

```go
sum := safeint.MustAdd(a, b)       // panics: "safeint: Add overflow"
product := safeint.MustMul(a, b)
quotient := safeint.MustDiv(a, b)
negated := safeint.MustNeg(a)
converted := safeint.MustConvert[int64, int32](a)
```

> **Warning:** Do not use `Must*` functions with untrusted input in server contexts — a panic will crash the goroutine.

### 3. `Int[T]` Wrapper Type — method-based API

A value-type wrapper providing three method families:

```go
a := safeint.New[int64](100)
b := safeint.New[int64](200)

// Checked — returns (Int[T], bool)
sum, ok := a.AddOverflow(b)

// Wrapping — silently wraps, like native Go
sum := a.Add(b)

// Must — panics on overflow
sum := a.MustAdd(b)
```

**Comparison methods:**

```go
a.Eq(b)       // ==
a.Lt(b)       // <
a.Gt(b)       // >
a.Lte(b)      // <=
a.Gte(b)      // >=
a.Cmp(b)      // -1, 0, +1
a.IsZero()
```

**Value access:**

```go
a.Val()        // returns the underlying T
a.String()     // implements fmt.Stringer
```

**Cross-type conversion:**

```go
// Go methods can't have extra type params, so this is a standalone function:
narrow, ok := safeint.ConvertInt[int64, int32](wideInt)
```

## Full-Precision MulDiv / MulMod

Compute `(a * b) / c` or `(a * b) % c` without intermediate overflow. The product `a * b` is computed at double width internally (128-bit for 64-bit types via `math/bits`).

```go
// (price * quantity) / divisor — no intermediate overflow
result, ok := safeint.MulDiv[int64](price, quantity, divisor)

// (a * b) % modulus
remainder, ok := safeint.MulMod[uint64](a, b, modulus)
```

Returns `(0, false)` when the divisor/modulus is zero or the final result overflows `T`.

## Supported Operations

| Operation | Checked | Must | Int[T] Checked | Int[T] Wrapping | Int[T] Must |
|-----------|---------|------|----------------|-----------------|-------------|
| Add       | `Add`   | `MustAdd` | `AddOverflow` | `Add` | `MustAdd` |
| Subtract  | `Sub`   | `MustSub` | `SubOverflow` | `Sub` | `MustSub` |
| Multiply  | `Mul`   | `MustMul` | `MulOverflow` | `Mul` | `MustMul` |
| Divide    | `Div`   | `MustDiv` | `DivOverflow` | — | `MustDiv` |
| DivMod    | `DivMod`| — | `DivModOverflow` | — | — |
| Modulo    | `Mod`   | — | `ModOverflow` | — | — |
| Negate    | `Neg`   | `MustNeg` | `NegOverflow` | — | — |
| Abs       | `Abs`   | — | `AbsOverflow` | — | — |
| Power     | `Pow`   | — | `PowOverflow` | — | — |
| Left Shift| `Lsh`   | — | `LshOverflow` | — | — |
| Convert   | `Convert` | `MustConvert` | `ConvertInt` | — | — |
| MulDiv    | `MulDiv`| `MustMulDiv` | `MulDivOverflow` | — | — |
| MulMod    | `MulMod`| — | `MulModOverflow` | — | — |

## Overflow Semantics

Some edge cases worth noting:

- **Division by zero** — returns `(0, false)`, never panics.
- **Signed MinInt / -1** — detected as overflow (Go would silently wrap).
- **MinInt % -1** — returns `(0, true)` (correct in Go, no overflow).
- **Neg on unsigned** — overflow for any non-zero value.
- **Abs on signed MinInt** — overflow (|MinInt| > MaxInt).
- **0^0** — returns `(1, true)` by convention.
- **Lsh** — returns `(0, false)` if shift >= bit width or bits are lost.

## Type Constraints

```go
type Integer interface { Signed | Unsigned }

type Signed interface {
    ~int | ~int8 | ~int16 | ~int32 | ~int64
}

type Unsigned interface {
    ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64
}
```

Named types are supported via the `~` constraint:

```go
type UserID int64
id, ok := safeint.Add(UserID(1), UserID(2)) // id is UserID
```

## Testing

Tests include exhaustive verification of all 8-bit type ranges (int8, uint8) against reference implementations using `math/big`, plus targeted edge-case tests for 16/32/64-bit types.

```
go test ./...
```

### Go Version Compatibility

CI tests against every Go release from 1.18 to 1.26. The minimum version is Go 1.18 (the release that introduced generics).

## Security Note

All algorithms use data-dependent branches and are **not constant-time**. Do not use in cryptographic contexts where timing side-channels matter.

## Acknowledgements

This project was inspired by and references ideas from:

- [g-utils/overflow](https://github.com/g-utils/overflow) — checked arithmetic functions for Go
- [holiman/uint256](https://github.com/holiman/uint256) — high-performance 256-bit integer library with dual API design (wrapping + checked methods)

## License

MIT
