------------------------------ MODULE Word ------------------------------
EXTENDS Integers

AbsVal(x) == IF x < 0 THEN -x ELSE x
MinVal(w, s) == IF s THEN -(2^(w-1)) ELSE 0
MaxVal(w, s) == IF s THEN 2^(w-1)-1 ELSE 2^w-1
Fits(x, w, s) == x >= MinVal(w, s) /\ x <= MaxVal(w, s)
Wrap(x, w, s) ==
    LET u == x % (2^w)
    IN IF s /\ u >= 2^(w-1) THEN u-2^w ELSE u

\* Go division truncates toward zero; signed right shift rounds down.
Quot(a, b) ==
    IF (a < 0) /= (b < 0)
    THEN -(AbsVal(a) \div AbsVal(b)) ELSE AbsVal(a) \div AbsVal(b)
Rem(a, b) == a - Quot(a, b)*b
Rshift(a, n) ==
    IF a < 0 THEN -((AbsVal(a)+2^n-1) \div 2^n) ELSE a \div 2^n
Ret(v, o) == [value |-> v, remainder |-> 0, ok |-> o]

AddImpl(a, b, w, s) ==
    LET v == Wrap(a+b, w, s) IN Ret(v, (v>a) = (b>0))
SubImpl(a, b, w, s) ==
    LET v == Wrap(a-b, w, s) IN Ret(v, (v<a) = (b>0))
MulImpl(a, b, w, s) ==
    IF a=0 \/ b=0 THEN Ret(0, TRUE)
    ELSE LET v == Wrap(a*b, w, s)
         IN Ret(v, ((v<0) = ((a<0) /= (b<0)))
                   /\ Wrap(Quot(v, b), w, s)=a)
ConvertImpl(a, sw, ss, dw, ds) ==
    LET v == Wrap(a, dw, ds)
    IN Ret(v, Wrap(v, sw, ss)=a /\ ((a<0) = (v<0)))
LshImpl(a, n, w, s) ==
    IF a=0 THEN Ret(0, TRUE)
    ELSE IF n>=w THEN Ret(0, FALSE)
    ELSE LET v == Wrap(a*2^n, w, s) IN Ret(v, Rshift(v, n)=a)

\* Saturating magnitude is an independent oracle: only fit and exact
\* in-range values matter. It avoids TLC's 32-bit arithmetic limit.
RECURSIVE CappedMagnitude(_, _, _)
CappedMagnitude(a, n, cap) ==
    IF n=0 THEN 1
    ELSE LET p == a*CappedMagnitude(a, n-1, cap)
         IN IF p>cap THEN cap ELSE p
PowerOracle(a, n, w) ==
    LET m == CappedMagnitude(AbsVal(a), n, 2^w)
    IN IF a<0 /\ n%2=1 THEN -m ELSE m
=============================================================================
