------------------------- MODULE APIFamilyConsistency -------------------------
EXTENDS Word
CONSTANTS Bits, Signed, FlipWrapperFlag, MutateInput
VARIABLES op, originalA, originalB, left, right, checked, wrappedMethod,
          nativeWrapped, panicked, mustValue, done
vars == <<op,originalA,originalB,left,right,checked,wrappedMethod,nativeWrapped,panicked,mustValue,done>>
Domain == MinVal(Bits,Signed)..MaxVal(Bits,Signed)
Checked(a,b) == CASE op="Add" -> AddImpl(a,b,Bits,Signed)
                 [] op="Sub" -> SubImpl(a,b,Bits,Signed)
                 [] op="Mul" -> MulImpl(a,b,Bits,Signed)
                 [] op="Div" -> IF b=0 \/ (Signed /\ a=MinVal(Bits,Signed) /\ b = -1)
                                THEN Ret(0,FALSE) ELSE Ret(Quot(a,b),TRUE)
Expected == CASE op="Add" -> originalA+originalB
              [] op="Sub" -> originalA-originalB
              [] op="Mul" -> originalA*originalB
              [] op="Div" -> IF originalB=0 THEN 0 ELSE Quot(originalA,originalB)
ExpectedOK == Fits(Expected,Bits,Signed) /\ (op="Div" => originalB/=0)
Init == /\ op \in {"Add","Sub","Mul","Div"}
        /\ originalA \in Domain /\ originalB \in Domain
        /\ left=originalA /\ right=originalB
        /\ checked=Ret(0,FALSE) /\ wrappedMethod=Ret(0,FALSE)
        /\ nativeWrapped=0 /\ panicked=FALSE /\ mustValue=0 /\ done=FALSE
Call == /\ ~done /\ done'=TRUE
        /\ LET r == Checked(left,right)
           IN /\ checked'=r
              /\ wrappedMethod'=Ret(r.value,IF FlipWrapperFlag THEN ~r.ok ELSE r.ok)
              /\ panicked'=~r.ok /\ mustValue'=IF r.ok THEN r.value ELSE 0
        /\ nativeWrapped'=Wrap(Expected,Bits,Signed)
        /\ left'=IF MutateInput THEN Wrap(left+1,Bits,Signed) ELSE left
        /\ UNCHANGED <<op,originalA,originalB,right>>
Next == Call \/ (done /\ UNCHANGED vars)
TypeOK == left \in Domain /\ right \in Domain /\ done \in BOOLEAN
CheckedSound == done => (checked.ok=ExpectedOK /\ (checked.ok => checked.value=Expected))
WrapperMatches == done => wrappedMethod=checked
WrappingMatches == (done /\ checked.ok /\ op/="Div") => nativeWrapped=checked.value
MustMatches == done => (panicked=~ExpectedOK /\ (~panicked => mustValue=Expected))
InputUnchanged == left=originalA /\ right=originalB
=============================================================================
