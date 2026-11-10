# Bounded TLA+ audit

The initial audit checked `0f96ba4fc460b378836c97d16485ff63653eb096` and found
one README contract defect with two concrete `Lsh` counterexamples. That
contract was corrected in `e030f05c3cae2effb67c939c5b4d15bd7bdcf4cf`.
The current verifier adds explicit checks of all three corrected `Lsh` rules;
the old README predicates remain historical negative controls.
The audit did not find a new arithmetic implementation defect in its checked domains.

## Initial results at 0f96ba4

- Nine current-model configurations completed successfully: 851,286 distinct
  states summed across those configurations.
- Two legacy controls and two README checks produced the exact expected
  invariant violations (TLC exit 12). These four checks are counterexamples,
  not successful proofs.
- The Go replay passed 1,895,122 mathematical assertions and reproduced both
  README counterexamples.
- The existing root-package tests passed with race detection and no skips;
  `go vet ./...` and `git diff --check` also passed.
- A verifier mutation that removed one configured check was rejected before
  execution, proving that the 13-check count is enforced.

## Run

Prerequisites: Java 11 or later, Python 3, Go 1.24 or later, and the pinned
[TLA+ tools 1.7.4 release](https://github.com/tlaplus/tlaplus/releases/tag/v1.7.4).
The verifier validates the JAR's SHA-256 before executing it.

From the repository root:

```sh
mkdir -p "$HOME/.cache/safeint-tla"
curl -fL --retry 2 \
  -o "$HOME/.cache/safeint-tla/tla2tools-1.7.4.jar" \
  https://github.com/tlaplus/tlaplus/releases/download/v1.7.4/tla2tools.jar
python3 formal/tla/verify.py
go run ./formal/tla/replay.go
```

Use `--jar /path/to/tla2tools-1.7.4.jar` or `TLA_JAR` to select another location.
Logs and `summary.json` default to `review-artifacts/tla-current-results/`; `--output`
selects a different directory. Git ignores `review-artifacts/`, so run the
verifier to produce them. Run the Go replay from this repository so its
module import resolves to the current local package.

The runner expects exactly 15 model checks: eleven current-model passes, two
legacy implementation violations, and two historical README violations. It fails if
any check is missing, exits unexpectedly, times out, or violates the wrong
invariant. Expected violations are reported as such, never as proofs that the
violating model is correct. The Go replay is separate and must also exit zero.

## Model-to-code mapping

| Model | Go implementation | Finite domain and properties |
| --- | --- | --- |
| `Word.tla`, `Arithmetic.tla`, `arithmetic-*.cfg` | `Add`, `Sub`, `Mul`, `Div`, `DivMod`, `Mod`, `Neg`, `Abs`, `Lsh`, `Convert` | Every signed/unsigned 8-bit input pair; shifts 0..10; conversion targets signed/unsigned 4 and 8 bits. Compare `ok`, successful values, remainder, and documented zero-on-failure returns against exact arithmetic. |
| `Arithmetic.tla`, `doubleword-*.cfg` | `mulDivSmall`, `mulDiv64*`, `mulModSmall`, `mulMod64*` | Every signed/unsigned 4-bit input triple; also sampled 8-bit extrema, their neighbors, and -2..2 where representable. Model high/low word multiplication, quotient guard, remainder high-word reduction, and sign restoration. |
| `Arithmetic.tla`, `lsh-contract-*.cfg` | `Lsh` and its corrected README contract | Every signed/unsigned 8-bit value and shift 0..10. Explicitly check zero success, nonzero overshift failure, and the returned shifted value plus overflow flag for all other shifts. |
| `Power.tla`, `power-*.cfg` | `Pow` | Every signed/unsigned 8-bit base and exponent 0..16. Model each binary-exponentiation iteration, intermediate overflow, final result, and termination under weak fairness. |
| `Scan.tla`, `scan.cfg` | `Int[T].Scan` float branch | An 8-bit intermediate word with 4-bit float significand; every representable integral input in -512..512, four half-integers, and symbolic NaN/infinities; 4/8-bit signed/unsigned targets; previous values min/zero/max. Require exact successful conversion and unchanged stored value after failure. |

All models are handwritten abstractions, not an automatic translation or
formal proof of the Go source. `Word.tla` explicitly models Go's two's-complement
wrap, division toward zero, and arithmetic right shift. Basic arithmetic
checks enumerate input states; only `Power` and `Scan` model internal state
transitions. `Power` collapses each loop iteration into one transition.

The doubleword model scales the 64-bit limb algorithm down to 4/8-bit limbs.
It relies on the standard `math/bits` multiplication/division contracts and
does not model CPU instructions. The small-division model uses the same
roundtrip/sign-check conversion as `Convert`, followed by explicit zeroing on
failure. The independent oracle uses range membership.

## Independent Go replay

`replay.go` invokes the actual package, with `math/big` integers and rationals
as independent oracles. It checks:

- All `int8`/`uint8` input pairs for basic binary arithmetic and remainders.
- All 8-bit values for unary operations, shifts, powers, and conversions.
- Boundary grids and 10,000 seeded random triples for each of the ten built-in
  supported integer types, including actual 64-bit `MulDiv`/`MulMod` paths.
- Floating-point boundary cases and 10,000 arbitrary float64 bit patterns for
  each of `int8`, `uint8`, `int64`, and `uint64` SQL scan targets. Rational
  comparison determines whether the represented input is an in-range integer.
- The two exact TLC counterexamples to the old README, now positive examples
  of the corrected contract.

The seed is 73. On the audited darwin/arm64 environment, this run performs
1,895,122 mathematical assertions plus the two README counterexample checks.
Go replay failure is fatal; it does not silently skip failed cases.

## Findings and negative controls

### Historical P3: README gave the wrong `Lsh` failure contract (fixed in e030f05)

At `0f96ba4`, `README.md:143` promised `(0, false)` for shifts at or above the bit width or
when bits are lost. TLC finds:

| Configuration | Counterexample | Actual Go result | README result |
| --- | --- | --- | --- |
| `readme-lsh-overshift.cfg` | `Lsh[int8](0, 8)` | `(0, true)` | `(0, false)` |
| `readme-lsh-lost-bits.cfg` | `Lsh[int8](-127, 1)` | `(2, false)` | `(0, false)` |

The first behavior agrees with the explicit zero exception in `safeint.go`
and the existing tests. The second preserves the native wrapped result when
only the roundtrip check fails. It does correctly report overflow. Prefer
these existing, tested behaviors when maintaining the contract. Commit
`e030f05` corrected the README and API comments without changing arithmetic.

The accepted contract is equivalent to:

> `Lsh`: zero always returns `(0, true)`. For nonzero values, a shift at or above
> the bit width returns `(0, false)`. Otherwise it returns the shifted value,
> with `ok = false` if the result overflows.

### Previously fixed bugs: controls, not new findings

- `legacy-muldiv.cfg` reenables returning the truncated quotient on overflow.
  TLC violates `FailureZero`: at four bits, `(-8 * -8) / -7` truncates to -9,
  which wraps to 7 instead of returning zero. Current-model checks pass.
- `legacy-scan.cfg` reenables the old signed conversion plus floating
  roundtrip check, modeling ARM's saturating out-of-range conversion. TLC
  violates `ScanExact`: an input of 128 becomes 127 with success for an
  unsigned 8-bit target. This is the reduced-width counterpart of the
  previously reproduced `2^63` float64 bug. Current-model checks pass.

## Limits

- TLC's standard arithmetic implementation uses bounded machine integers;
  using small domains avoids overflowing the checker itself. See the official
  [TLC tools documentation](https://github.com/tlaplus/tlaplus/blob/master/general/docs/current-tools.md).
  The power oracle caps magnitudes outside the representable domain, so it
  preserves fit decisions without constructing huge intermediate powers.
- Reduced-width exhaustive checking is not exhaustive verification of all
  64-bit inputs. Concrete 64-bit replay supplies additional sampled evidence.
- The floating model is not a complete IEEE-754 model. It models nearest-even
  rounding in a small significand and special non-finite categories; concrete
  Go tests exercise actual float64 behavior.
- JSON/custom named-type serialization, decimal string parsing, driver-specific
  SQL behavior, Must-method panic recovery, and concurrent caller behavior are
  not modeled. The package's existing tests cover additional behavior but do
  not remove these formal-model limits.
- No production source was modified by this audit. The models and this report
  are versioned; logs are not.
