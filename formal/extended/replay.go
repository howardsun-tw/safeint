// Run with go run ./formal/extended/replay.go. Checks the numeric JSON codec policy.
package main

import (
	"encoding/json"
	"fmt"
	s "github.com/howardsun-tw/safeint"
	"math"
	"math/rand"
	"strconv"
	"unsafe"
)

var assertions int

func require(ok bool, label string) {
	assertions++
	if !ok {
		panic(label)
	}
}

type QuotedInt int64

func (v QuotedInt) MarshalJSON() ([]byte, error) {
	return json.Marshal(strconv.FormatInt(int64(v), 10))
}
func (v *QuotedInt) UnmarshalJSON(b []byte) error {
	var str string
	if err := json.Unmarshal(b, &str); err != nil {
		return err
	}
	n, err := strconv.ParseInt(str, 10, 64)
	if err == nil {
		*v = QuotedInt(n)
	}
	return err
}

type OneBasedIndex int64

func (v OneBasedIndex) MarshalJSON() ([]byte, error) { return json.Marshal(int64(v) + 1) }
func (v *OneBasedIndex) UnmarshalJSON(b []byte) error {
	var n int64
	if err := json.Unmarshal(b, &n); err != nil {
		return err
	}
	*v = OneBasedIndex(n - 1)
	return nil
}

type PlainInt int64

func codecRoundTrips() {
	for _, v := range []int64{1, 2, 7} {
		q := QuotedInt(v)
		b, e := json.Marshal(q)
		require(e == nil, "bare quoted encode")
		var qr QuotedInt
		require(json.Unmarshal(b, &qr) == nil && qr == q, "bare quoted roundtrip")
		b, e = json.Marshal(s.New(q))
		require(e == nil, "wrapped quoted encode")
		var qw s.Int[QuotedInt]
		e = json.Unmarshal(b, &qw)
		require(e == nil && qw.Val() == q, "quoted codec numeric roundtrip")
		fmt.Printf("PASS quoted: input=%d wire=%s output=%d\n", v, b, qw.Val())
		i := OneBasedIndex(v)
		b, e = json.Marshal(i)
		require(e == nil, "bare index encode")
		var ir OneBasedIndex
		require(json.Unmarshal(b, &ir) == nil && ir == i, "bare index roundtrip")
		b, e = json.Marshal(s.New(i))
		require(e == nil, "wrapped index encode")
		var iw s.Int[OneBasedIndex]
		e = json.Unmarshal(b, &iw)
		require(e == nil && iw.Val() == i, "index codec numeric roundtrip")
		fmt.Printf("PASS index: input=%d wire=%s output=%d\n", v, b, iw.Val())
	}
}
func must[T s.Integer](f func() T) (v T, panicked bool) {
	panicked = true
	defer func() { recover() }()
	v = f()
	panicked = false
	return
}
func audit[T s.Integer]() {
	width := uint(unsafe.Sizeof(T(0)) * 8)
	hi := ^T(0)
	lo := T(0)
	if hi < 0 {
		hi = T((uint64(1) << (width - 1)) - 1)
		lo = -hi - 1
	}
	vals := []T{lo, lo + 1, 0, 1, 2, hi - 1, hi}
	r := rand.New(rand.NewSource(73))
	for i := 0; i < 40; i++ {
		vals = append(vals, T(r.Uint64()))
	}
	for _, a := range vals {
		x := s.New(a)
		b, e := json.Marshal(x)
		require(e == nil, "JSON encode")
		y := s.New(T(3))
		require(json.Unmarshal(b, &y) == nil && y.Val() == a, "JSON roundtrip")
		dv, e := x.Value()
		require(e == nil, "SQL value")
		if lo == 0 && uint64(a) > math.MaxInt64 {
			_, ok := dv.(string)
			require(ok, "SQL high unsigned string")
		} else {
			_, ok := dv.(int64)
			require(ok, "SQL int64")
		}
		require(y.Scan(dv) == nil && y.Val() == a, "SQL roundtrip")
		for _, token := range []string{"1.5", "\"bad\"", "{}", "18446744073709551616", "-18446744073709551616", "["} {
			old := y.Val()
			require(json.Unmarshal([]byte(token), &y) != nil && y.Val() == old, "JSON failure atomicity")
		}
		old := y.Val()
		require(json.Unmarshal([]byte("null"), &y) == nil && y.Val() == old, "JSON null preserves")
		for _, src := range []interface{}{1.5, math.NaN(), math.Inf(1), "bad", []byte("1.5"), true, "18446744073709551616", "-18446744073709551616"} {
			old := y.Val()
			require(y.Scan(src) != nil && y.Val() == old, "SQL failure atomicity")
		}
		require(y.Scan(nil) == nil && y.Val() == 0, "SQL null clears")
		require(y.Scan([]byte("2")) == nil && y.Val() == 2, "SQL recovery")
		require(json.Unmarshal([]byte("3"), &y) == nil && y.Val() == 3, "JSON recovery")
		check := func(v T, ok bool, w s.Int[T], wk bool) {
			require(v == w.Val() && ok == wk, "checked wrapper mismatch")
			require(x.Val() == a, "input mutated")
		}
		v, ok := s.Neg(a)
		w, wk := x.NegOverflow()
		check(v, ok, w, wk)
		v, ok = s.Abs(a)
		w, wk = x.AbsOverflow()
		check(v, ok, w, wk)
		for _, n := range []uint{0, 1, 2, 7, 31, 32, 63, 64, 65} {
			v, ok = s.Pow(a, n)
			w, wk = x.PowOverflow(n)
			check(v, ok, w, wk)
			v, ok = s.Lsh(a, n)
			w, wk = x.LshOverflow(n)
			check(v, ok, w, wk)
		}
		for _, b := range vals {
			z := s.New(b)
			cases := []struct {
				f           func(T, T) (T, bool)
				wrap        func(s.Int[T]) (s.Int[T], bool)
				panicMethod func(s.Int[T]) s.Int[T]
				panicFunc   func(T, T) T
				raw         func(s.Int[T]) s.Int[T]
			}{
				{s.Add[T], x.AddOverflow, x.MustAdd, s.MustAdd[T], x.Add}, {s.Sub[T], x.SubOverflow, x.MustSub, s.MustSub[T], x.Sub}, {s.Mul[T], x.MulOverflow, x.MustMul, s.MustMul[T], x.Mul}, {s.Div[T], x.DivOverflow, x.MustDiv, s.MustDiv[T], nil},
			}
			for _, c := range cases {
				v, ok := c.f(a, b)
				w, wk := c.wrap(z)
				check(v, ok, w, wk)
				mv, mp := must(func() T { return c.panicMethod(z).Val() })
				fv, fp := must(func() T { return c.panicFunc(a, b) })
				require(mp == !ok && fp == !ok, "Must panic agreement")
				if ok {
					require(mv == v && fv == v, "Must value")
					if c.raw != nil {
						require(c.raw(z).Val() == v, "wrapping agreement")
					}
				}
			}
			v, ok = s.Mod(a, b)
			w, wk = x.ModOverflow(z)
			check(v, ok, w, wk)
			q, rem, qok := s.DivMod(a, b)
			qw, rw, rok := x.DivModOverflow(z)
			check(q, qok, qw, rok)
			require(rem == rw.Val(), "DivMod remainder")
			for _, c := range []T{0, 1, lo, hi} {
				v, ok = s.MulDiv(a, b, c)
				w, wk = x.MulDivOverflow(z, s.New(c))
				check(v, ok, w, wk)
				v, ok = s.MulMod(a, b, c)
				w, wk = x.MulModOverflow(z, s.New(c))
				check(v, ok, w, wk)
			}
			require(z.Val() == b, "second input mutated")
		}
	}
	fmt.Printf("ordinary codec/state/API checks: %T, %d values\n", T(0), len(vals))
}
func main() {
	codecRoundTrips()
	audit[int8]()
	audit[uint8]()
	audit[int16]()
	audit[uint16]()
	audit[int32]()
	audit[uint32]()
	audit[int64]()
	audit[uint64]()
	audit[int]()
	audit[uint]()
	audit[PlainInt]()
	fmt.Printf("Completed %d assertions; both custom codecs preserve numeric values.\n", assertions)
}
