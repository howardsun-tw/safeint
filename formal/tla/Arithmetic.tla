--------------------------- MODULE Arithmetic ---------------------------
EXTENDS Word, TLC
CONSTANTS Bits, Signed, Full, Operations, LegacyMulDiv
VARIABLES op, a, b, c, n, targetBits, targetSigned
vars == <<op, a, b, c, n, targetBits, targetSigned>>
Domain == MinVal(Bits, Signed)..MaxVal(Bits, Signed)
Samples == {MinVal(Bits, Signed), MinVal(Bits, Signed)+1,
            MaxVal(Bits, Signed), MaxVal(Bits, Signed)-1, -2,-1,0,1,2}
Inputs == IF Full THEN Domain ELSE Samples \cap Domain
Binary == {"Add", "Sub", "Mul", "Div", "DivMod", "Mod"}
Ternary == {"MulDivSmall", "MulModSmall", "MulDivWord", "MulModWord"}
Init == /\ op \in Operations
        /\ a \in Inputs
        /\ b \in (IF op \in Binary \cup Ternary THEN Inputs ELSE {0})
        /\ c \in (IF op \in Ternary THEN Inputs ELSE {0})
        /\ n \in (IF op="Lsh" THEN 0..(Bits+2) ELSE {0})
        /\ targetBits \in (IF op="Convert" THEN {4,8} ELSE {Bits})
        /\ targetSigned \in (IF op="Convert" THEN BOOLEAN ELSE {Signed})
Next == UNCHANGED vars

DivImpl ==
    IF b=0 \/ (Signed /\ a=MinVal(Bits, Signed) /\ b = -1)
    THEN Ret(0, FALSE)
    ELSE LET q == Quot(a,b)
             o == q=0 \/ ((q<0) = ((a<0) /= (b<0)))
         IN [value |-> q, remainder |-> IF op="DivMod" THEN Rem(a,b) ELSE 0,
             ok |-> o]
SmallDiv ==
    IF c=0 THEN Ret(0,FALSE)
    ELSE LET q == Quot(a*b,c)
             cv == ConvertImpl(q,2*Bits,Signed,Bits,Signed)
         IN Ret(IF cv.ok \/ LegacyMulDiv THEN cv.value ELSE 0, cv.ok)

\* Scale the 64-bit math/bits implementation down to a Bits-bit word.
\* hi and lo are unsigned words; the reference below uses exact product.
WordDiv ==
    IF c=0 THEN Ret(0,FALSE)
    ELSE LET ua == AbsVal(a)
             ub == AbsVal(b)
             uc == AbsVal(c)
             hi == (ua*ub) \div (2^Bits)
             lo == (ua*ub) % (2^Bits)
             negative == ((a<0) /= (b<0)) /= (c<0)
         IN IF hi>=uc THEN Ret(0,FALSE)
            ELSE LET q == (hi*2^Bits+lo) \div uc
                 IN IF Signed
                    THEN IF negative
                         THEN IF q>2^(Bits-1) THEN Ret(0,FALSE)
                              ELSE Ret(Wrap(-q,Bits,TRUE),TRUE)
                         ELSE IF q>2^(Bits-1)-1 THEN Ret(0,FALSE)
                              ELSE Ret(q,TRUE)
                    ELSE Ret(q,TRUE)
WordMod ==
    IF c=0 THEN Ret(0,FALSE)
    ELSE LET ua == AbsVal(a)
             ub == AbsVal(b)
             uc == AbsVal(c)
             hi == ((ua*ub) \div (2^Bits)) % uc
             lo == (ua*ub) % (2^Bits)
             r == (hi*2^Bits+lo) % uc
         IN Ret(IF (a<0) /= (b<0) THEN -r ELSE r, TRUE)

Actual == CASE op="Add" -> AddImpl(a,b,Bits,Signed)
            [] op="Sub" -> SubImpl(a,b,Bits,Signed)
            [] op="Mul" -> MulImpl(a,b,Bits,Signed)
            [] op \in {"Div","DivMod"} -> DivImpl
            [] op="Mod" -> IF b=0 THEN Ret(0,FALSE) ELSE Ret(Rem(a,b),TRUE)
            [] op="Neg" -> SubImpl(0,a,Bits,Signed)
            [] op="Abs" -> IF a<0 THEN SubImpl(0,a,Bits,Signed) ELSE Ret(a,TRUE)
            [] op="Lsh" -> LshImpl(a,n,Bits,Signed)
            [] op="Convert" -> ConvertImpl(a,Bits,Signed,targetBits,targetSigned)
            [] op="MulDivSmall" -> SmallDiv
            [] op="MulDivWord" -> WordDiv
            [] op="MulModSmall" -> IF c=0 THEN Ret(0,FALSE) ELSE Ret(Rem(a*b,c),TRUE)
            [] op="MulModWord" -> WordMod

MathResult == CASE op="Add" -> a+b
                [] op="Sub" -> a-b
                [] op="Mul" -> a*b
                [] op \in {"Div","DivMod"} -> IF b=0 THEN 0 ELSE Quot(a,b)
                [] op="Mod" -> IF b=0 THEN 0 ELSE Rem(a,b)
                [] op="Neg" -> -a
                [] op="Abs" -> AbsVal(a)
                [] op="Lsh" -> a*2^n
                [] op="Convert" -> a
                [] op \in {"MulDivSmall","MulDivWord"} -> IF c=0 THEN 0 ELSE Quot(a*b,c)
                [] op \in {"MulModSmall","MulModWord"} -> IF c=0 THEN 0 ELSE Rem(a*b,c)
MathOK == /\ Fits(MathResult,targetBits,targetSigned)
          /\ (op \in {"Div","DivMod","Mod"} => b/=0)
          /\ (op \in Ternary => c/=0)
          /\ (op="Lsh" => (a=0 \/ n<Bits))
TypeOK == /\ op \in Operations /\ a \in Domain
          /\ Actual.value \in Domain \cup (MinVal(targetBits,targetSigned)..MaxVal(targetBits,targetSigned))
          /\ Actual.ok \in BOOLEAN
Exact == /\ Actual.ok = MathOK
         /\ (Actual.ok => Actual.value=MathResult)
         /\ (op="DivMod" /\ Actual.ok => Actual.remainder=Rem(a,b))
FailureZero == (op \in Ternary \cup {"Div","DivMod","Mod"} /\ ~Actual.ok)
               => (Actual.value=0 /\ Actual.remainder=0)
\* Current public contract, including the returned value on overflow.
LshContract == op="Lsh" =>
    IF a=0 THEN Actual=Ret(0,TRUE)
    ELSE IF n>=Bits THEN Actual=Ret(0,FALSE)
    ELSE Actual=Ret(Wrap(a*2^n,Bits,Signed),Fits(a*2^n,Bits,Signed))
\* Historical README at 0f96ba4, retained only as negative controls.
ReadmeOvershift == (op="Lsh" /\ n>=Bits) => Actual=Ret(0,FALSE)
ReadmeLostBits == (op="Lsh" /\ ~MathOK) => Actual=Ret(0,FALSE)
=============================================================================
