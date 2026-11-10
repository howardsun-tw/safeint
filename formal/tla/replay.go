// Run with: go run ./formal/tla/replay.go
// Independent concrete checks for the bounded TLA+ audit. No package mutation.
package main

import (
	"fmt"
	"math"
	"math/big"
	"math/rand"
	"unsafe"

	s "github.com/howardsun-tw/safeint"
)

var assertions int

func integer[T s.Integer](x T) *big.Int {
	if ^T(0) < 0 {
		return big.NewInt(int64(x))
	}
	return new(big.Int).SetUint64(uint64(x))
}

func bounds[T s.Integer]() (*big.Int, *big.Int) {
	width := uint(unsafe.Sizeof(T(0)) * 8)
	if ^T(0) < 0 {
		upper := new(big.Int).Lsh(big.NewInt(1), width-1)
		return new(big.Int).Neg(upper), new(big.Int).Sub(upper, big.NewInt(1))
	}
	return new(big.Int), new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), width), big.NewInt(1))
}

func verify[T s.Integer](label string, got T, ok bool, want *big.Int, valid, zeroOnError bool) {
	assertions++
	lo, hi := bounds[T]()
	valid = valid && want.Cmp(lo) >= 0 && want.Cmp(hi) <= 0
	if ok != valid || ok && integer(got).Cmp(want) != 0 || !ok && zeroOnError && got != 0 {
		panic(fmt.Sprintf("%s [%T]: got (%v,%v), expected mathematical (%v,%v), zeroOnError=%v", label, got, got, ok, want, valid, zeroOnError))
	}
}

func arithmetic[T s.Integer](a, b T) {
	x, y := integer(a), integer(b)
	v, ok := s.Add(a, b)
	verify("Add", v, ok, new(big.Int).Add(x, y), true, false)
	v, ok = s.Sub(a, b)
	verify("Sub", v, ok, new(big.Int).Sub(x, y), true, false)
	v, ok = s.Mul(a, b)
	verify("Mul", v, ok, new(big.Int).Mul(x, y), true, false)
	q, r := new(big.Int), new(big.Int)
	if b != 0 {
		q.QuoRem(x, y, r)
	}
	v, ok = s.Div(a, b)
	verify("Div", v, ok, q, b != 0, true)
	v, ok = s.Mod(a, b)
	verify("Mod", v, ok, r, b != 0, true)
	v, rem, ok := s.DivMod(a, b)
	verify("DivMod quotient", v, ok, q, b != 0, true)
	assertions++
	if ok && integer(rem).Cmp(r) != 0 || !ok && rem != 0 {
		panic(fmt.Sprintf("DivMod remainder [%T]: (%v,%v) -> %v, ok=%v", a, a, b, rem, ok))
	}
}

func mulDivMod[T s.Integer](a, b, c T) {
	product := new(big.Int).Mul(integer(a), integer(b))
	q, r := new(big.Int), new(big.Int)
	if c != 0 {
		q.QuoRem(product, integer(c), r)
	}
	v, ok := s.MulDiv(a, b, c)
	verify("MulDiv", v, ok, q, c != 0, true)
	v, ok = s.MulMod(a, b, c)
	verify("MulMod", v, ok, r, c != 0, true)
}

func convert[T s.Integer, U s.Integer](a T) {
	v, ok := s.Convert[T, U](a)
	verify("Convert", v, ok, integer(a), true, false)
}

func unary[T s.Integer](a T) {
	v, ok := s.Neg(a)
	verify("Neg", v, ok, new(big.Int).Neg(integer(a)), true, false)
	v, ok = s.Abs(a)
	verify("Abs", v, ok, new(big.Int).Abs(integer(a)), true, false)
	width := uint(unsafe.Sizeof(a) * 8)
	for n := uint(0); n <= width+2; n++ {
		v, ok = s.Lsh(a, n)
		verify("Lsh", v, ok, new(big.Int).Lsh(integer(a), n), a == 0 || n < width, false)
	}
	for n := uint(0); n <= 16; n++ {
		v, ok = s.Pow(a, n)
		verify("Pow", v, ok, new(big.Int).Exp(integer(a), big.NewInt(int64(n)), nil), true, false)
	}
	convert[T, int8](a)
	convert[T, uint8](a)
	convert[T, int16](a)
	convert[T, uint16](a)
	convert[T, int32](a)
	convert[T, uint32](a)
	convert[T, int64](a)
	convert[T, uint64](a)
	convert[T, int](a)
	convert[T, uint](a)
}

func exhaustive8[T s.Integer]() {
	start := assertions
	for i := 0; i < 256; i++ {
		a := T(i)
		unary(a)
		for j := 0; j < 256; j++ {
			arithmetic(a, T(j))
		}
	}
	fmt.Printf("PASS exhaustive %T: %d assertions\n", T(0), assertions-start)
}

func sampleType[T s.Integer]() {
	start := assertions
	lo, hi := bounds[T]()
	values := []T{0, 1, 2, T(hi.Uint64()), T(new(big.Int).Sub(hi, big.NewInt(1)).Uint64())}
	if ^T(0) < 0 {
		values = append(values, T(lo.Int64()), T(lo.Int64()+1), ^T(0), ^T(1))
	}
	for _, a := range values {
		unary(a)
		for _, b := range values {
			arithmetic(a, b)
			for _, c := range values {
				mulDivMod(a, b, c)
			}
		}
	}
	rng := rand.New(rand.NewSource(73))
	for i := 0; i < 10000; i++ {
		a, b, c := T(rng.Uint64()), T(rng.Uint64()), T(rng.Uint64())
		arithmetic(a, b)
		mulDivMod(a, b, c)
	}
	fmt.Printf("PASS boundaries and 10000 seeded triples %T: %d assertions\n", T(0), assertions-start)
}

func scan[T s.Integer]() {
	start := assertions
	lo, hi := bounds[T]()
	inputs := []float64{0, math.Copysign(0, -1), -1, 1, 1.5, -1.5,
		math.NaN(), math.Inf(1), math.Inf(-1)}
	for _, x := range []float64{-0x1p63, 0x1p63, 0x1p64} {
		inputs = append(inputs, math.Nextafter(x, math.Inf(-1)), x, math.Nextafter(x, math.Inf(1)))
	}
	rng := rand.New(rand.NewSource(73))
	for i := 0; i < 10000; i++ {
		inputs = append(inputs, math.Float64frombits(rng.Uint64()))
	}
	for _, input := range inputs {
		v := s.New[T](7)
		err := v.Scan(input)
		want := new(big.Rat).SetFloat64(input)
		valid := want != nil && want.IsInt() && want.Num().Cmp(lo) >= 0 && want.Num().Cmp(hi) <= 0
		assertions++
		if (err == nil) != valid || valid && integer(v.Val()).Cmp(want.Num()) != 0 || !valid && v.Val() != 7 {
			panic(fmt.Sprintf("Scan[%T](%g) = (%v,%v), valid=%v", T(0), input, v.Val(), err, valid))
		}
	}
	fmt.Printf("PASS Scan %T: %d exact-rational assertions\n", T(0), assertions-start)
}

func main() {
	exhaustive8[int8]()
	exhaustive8[uint8]()
	sampleType[int8]()
	sampleType[uint8]()
	sampleType[int16]()
	sampleType[uint16]()
	sampleType[int32]()
	sampleType[uint32]()
	sampleType[int64]()
	sampleType[uint64]()
	sampleType[int]()
	sampleType[uint]()
	scan[int8]()
	scan[uint8]()
	scan[int64]()
	scan[uint64]()
	fmt.Printf("PASS %d concrete assertions\n", assertions)
	zero, ok := s.Lsh[int8](0, 8)
	if zero != 0 || !ok {
		panic("Lsh zero contract violated; re-run TLC and update the report")
	}
	fmt.Printf("PASS current Lsh contract: Lsh[int8](0,8) = (%v,%v)\n", zero, ok)
	v, ok := s.Lsh[int8](-127, 1)
	if v != 2 || ok {
		panic("Lsh overflow result contract violated; re-run TLC and update the report")
	}
	fmt.Printf("PASS current Lsh contract: Lsh[int8](-127,1) = (%v,%v)\n", v, ok)
}
