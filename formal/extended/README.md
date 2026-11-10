# Extended formal audit

Follow-up: the finding is fixed in `24cd137`. `Int[T]` now uses the numeric JSON
policy in both directions, and `replay.go` asserts preserved values for both
custom codecs.

The models follow the fix. In `NamedTypeCodec.tla`, `Policy="numeric"` is the
current implementation and `Policy="legacy"` is the pre-fix behavior, kept as a
control. The TLC results below were rerun on 2026-11-10 with these names. The
finding, the Z3 results and the concrete verification below remain the pre-fix
audit pinned to `e030f05`; they are not verification of the current
implementation. Post-fix checks were recorded under
`review-artifacts/json-codec-fix/`, which Git ignores.

Audited source: `e030f05c3cae2effb67c939c5b4d15bd7bdcf4cf` (2026-09-08).
No package implementation was changed. Models and replay programs are handwritten
audit artifacts, not a compiler-verified translation of Go.

## Confirmed finding: asymmetric named-type JSON codecs (P2)

`int.go:218` formats the integer directly, ignoring `T.MarshalJSON`.
`int.go:226` decodes into `T`, invoking `T.UnmarshalJSON` when present.
Consequently, a named integer with an internally consistent custom codec can fail
to roundtrip through `Int[T]`, including silent value changes with a nil error.

The executable reproduction in `replay.go` first verifies each bare named type's
roundtrip, then exercises the actual package through `encoding/json`:

| Named codec | Bare behavior for value 1 | Wrapped behavior |
| --- | --- | --- |
| `QuotedInt` | Encodes `"1"`, decodes to 1 | Encodes `1`, decoding returns an error |
| `OneBasedIndex` | Encodes `2`, decodes to 1 | Encodes `1`, decodes to **0 with nil error** |

Both cases reproduce for original values 1, 2, and 7. They are two manifestations
of one root cause, not two independent findings. Ordinary built-in integers and
a plain named integer without custom codecs passed the concrete checks.

A fix needs a consistent policy: either encode and decode as built-in numbers,
or delegate both directions to the named type. Both policies pass the bounded
counterfactual model. Neither policy has been implemented in this audit. The
failure-atomicity checks do not promise atomic behavior for arbitrary custom
unmarshalers that mutate themselves before returning an error.

## TLC results

All 15 configurations completed with their expected outcomes, visiting 25,198
distinct states in total across runs. This includes intentional invariant
violations; it does not mean every configuration satisfied its invariants.

| Model | Scope | Outcome |
| --- | --- | --- |
| `DecodeState` | Signed/unsigned 4-bit integers, three decode calls, JSON/SQL input categories | Current configurations pass; null-clearing mutation is caught |
| `SerializationRoundTrip` | Signed/unsigned 8-bit values, JSON and SQL | Current configurations pass; high unsigned truncation mutation is caught |
| `NamedTypeCodec` | Values 1 through 7, plain/string/index codecs | The numeric policy (current) passes for all three codecs. The legacy policy passes for plain and violates wrapped roundtrip for both custom codecs. The delegate policy (counterfactual) passes |
| `APIFamilyConsistency` | Signed/unsigned 4-bit Add/Sub/Mul/Div | Current configurations pass; incorrect flag and input mutation controls are caught |

Classification: 7 passing current configurations, 3 legacy configurations
(1 passing, 2 violating), 1 passing counterfactual policy, and 4 deliberately
failing negative controls. Go replay at `e030f05` confirmed the two legacy
violation traces as the defect described above.

Parsing and wire payloads are abstracted, including SQL's signed intermediate:
the bounded model uses the modeled word width, while production uses int64.
This is a structural model, not a literal int8 SQL execution. Full-width SQL
boundary behavior is checked against the actual Go package. Wrapper delegation
in the model is explicit and partially true by construction; concrete wrapper
checks provide a separate connection to the implementation.

## Z3 results

Z3 4.14.1 completed 73 queries: 70 `unsat`, 2 deliberate negative controls `sat`,
and 1 `unknown` due to timeout. The runner returns exit code 2 for this incomplete
proof set. No positive query returned a counterexample.

| Query group | Outcome and exact scope |
| --- | --- |
| Convert | All 64 signed/unsigned 8/16/32/64-bit source-target combinations: acceptance and successful value preservation proved in the bit-vector model |
| Lsh | Signed and unsigned 64-bit acceptance agrees with mathematical representability for all modeled operands and shift counts |
| Unsigned 64-bit MulDiv | High-word overflow guard agrees with representability of the full 128-bit quotient; this query checks acceptance, not an independent implementation of quotient computation |
| Signed 64-bit MulDiv | Combined acceptance/value query remains **unknown** after 20 seconds; a separate CLI retry also timed out after 90 seconds |
| Three helper lemmas | Absolute magnitude, high-word bound, and signed quotient restoration: `unsat`; these do not establish the timed-out combined proof |
| Two negative controls | Missing Convert sign check and logical signed Lsh restoration each produce `sat` |

These are fixed-width bit-vector proofs of the stated expressions. They do not
prove Go compiler behavior, `math/bits.Div64` internals, parsing, or arbitrary
runtime execution. Full-width arithmetic also has the independent sampled Go
oracle described below; sampled tests cannot replace the unknown proof.

## Concrete verification

- Extended replay: **952,430 assertions** completed. Covers JSON and Value/Scan
  roundtrips, failed-decode state preservation, null behavior and recovery,
  checked method/function agreement, wrapping agreement on success, Must panic
  and result agreement, and input immutability. Exercises all ten supported
  built-in integer types plus `PlainInt`, using 47 boundary/seeded values per
  type. Covers all checked arithmetic wrappers; the four Must methods are
  compared with their standalone counterparts. At `e030f05` the replay expected
  the confirmed custom-codec defect; it now asserts the fixed behavior.
- Existing arithmetic replay: **1,895,122 assertions** passed against independent
  big-integer/rational expectations, plus two current Lsh contract checks.
- `go test -race -count=1 -json .`: **943 test/subtest pass events**, zero failures,
  zero skips. This is the root package test run.
- `go vet ./...`: exit 0 after correcting an audit-harness compile error.
  The initial harness mistakenly included unsupported `uintptr`; it was removed
  because `safeint.Integer` excludes that type. No supported type was skipped.

## Reproduction and evidence

From the repository root:

```sh
python3 formal/tla/verify_extended.py
PYTHONPATH=/opt/homebrew/Cellar/z3/4.14.1/lib/python3.13/site-packages python3 formal/z3/verify.py
go run ./formal/extended/replay.go
go run ./formal/tla/replay.go
go test -race -count=1 -json .
go vet ./...
```

The Python path above is the local Homebrew Z3 binding location; other machines
need their own working `z3` Python module. TLC uses the pinned jar verified by
the runner, Java 17, one worker, seed 73, and a 180-second per-run timeout.

Git ignores `review-artifacts/`, so the raw evidence is not in the repository.
The two verifier commands above write fresh `tlc/` and `z3/` output there for
the current source. No command regenerates the other files listed below; they
exist only where the audit ran.

Raw evidence is under `review-artifacts/extended-models/`: `tlc/summary.json`
and per-configuration traces, `z3/summary.json` and all 73 SMT2 queries,
`z3/muldiv64-signed-full-retry.log`, `go-replay.log`,
`go-arithmetic-replay.log`, `go-test.json`, and `go-vet.log`.
The solver summaries pin HEAD and SHA-256 hashes of `int.go` and `safeint.go`;
the TLC summary additionally records model/configuration hashes.
