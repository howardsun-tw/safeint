------------------------------- MODULE Scan -------------------------------
EXTENDS Word
CONSTANTS Bits, Precision, LegacyScan
VARIABLES src, targetBits, targetSigned, before, stored, ok, done
vars == <<src,targetBits,targetSigned,before,stored,ok,done>>
\* Integer-only IEEE-style nearest-even rounding, with reduced precision.
RoundFloat(x) ==
    IF x=0 THEN 0
    ELSE LET mag == AbsVal(x)
             e == CHOOSE k \in 0..(Bits+2): 2^k<=mag /\ mag<2^(k+1)
             spacing == IF e<Precision THEN 1 ELSE 2^(e-Precision+1)
             q == mag \div spacing
             r == mag % spacing
             up == 2*r>spacing \/ (2*r=spacing /\ q%2=1)
             rounded == (q + IF up THEN 1 ELSE 0)*spacing
         IN IF x<0 THEN -rounded ELSE rounded
Finite == {[kind |-> "finite", num |-> x, den |-> 1]:
           x \in {i \in -(2^(Bits+1))..2^(Bits+1): RoundFloat(i)=i}}
Fractions == {[kind |-> "finite",num |-> x,den |-> 2]: x \in {-3,-1,1,3}}
Special == {[kind |-> k,num |-> 0,den |-> 1]: k \in {"NaN","+Inf","-Inf"}}
Init == /\ src \in Finite \cup Fractions \cup Special
        /\ targetBits \in {4,Bits} /\ targetSigned \in BOOLEAN
        /\ before \in {MinVal(targetBits,targetSigned),0,MaxVal(targetBits,targetSigned)}
        /\ stored=before /\ ok=FALSE /\ done=FALSE
Legacy ==
    IF src.kind/="finite" THEN Ret(0,FALSE)
    ELSE LET v == Quot(src.num,src.den)
             i == IF v<MinVal(Bits,TRUE) THEN MinVal(Bits,TRUE)
                  ELSE IF v>MaxVal(Bits,TRUE) THEN MaxVal(Bits,TRUE) ELSE v
             cv == ConvertImpl(i,Bits,TRUE,targetBits,targetSigned)
         IN Ret(cv.value,RoundFloat(i)*src.den=src.num /\ cv.ok)
Current ==
    IF src.kind/="finite" \/ src.den/=1 THEN Ret(0,FALSE)
    ELSE IF ~Fits(src.num,Bits,targetSigned) THEN Ret(0,FALSE)
    ELSE ConvertImpl(src.num,Bits,targetSigned,targetBits,targetSigned)
Result == IF LegacyScan THEN Legacy ELSE Current
Step == /\ ~done /\ UNCHANGED <<src,targetBits,targetSigned,before>>
        /\ stored'=(IF Result.ok THEN Result.value ELSE before)
        /\ ok'=Result.ok /\ done'=TRUE
Next == Step \/ (done /\ UNCHANGED vars)
ExpectedOK == src.kind="finite" /\ src.den=1 /\ Fits(src.num,targetBits,targetSigned)
TypeOK == /\ stored \in MinVal(targetBits,targetSigned)..MaxVal(targetBits,targetSigned)
          /\ ok \in BOOLEAN /\ done \in BOOLEAN
ScanExact == done => (ok=ExpectedOK /\ (ok => stored=src.num))
FailureAtomic == (done /\ ~ok) => stored=before
=============================================================================
