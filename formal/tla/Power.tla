------------------------------ MODULE Power ------------------------------
EXTENDS Word
CONSTANTS Bits, Signed, MaxExponent
VARIABLES base, originalExp, exp, factor, result, ok, done
vars == <<base,originalExp,exp,factor,result,ok,done>>
Init == /\ base \in MinVal(Bits,Signed)..MaxVal(Bits,Signed)
        /\ originalExp \in 0..MaxExponent
        /\ exp=originalExp /\ factor=base /\ result=1 /\ ok=TRUE
        /\ done=(exp=0 \/ base=1)
Step == /\ ~done
        /\ UNCHANGED <<base,originalExp>>
        /\ LET product == IF exp%2=1 THEN MulImpl(result,factor,Bits,Signed)
                          ELSE Ret(result,TRUE)
               remaining == exp \div 2
               square == IF product.ok /\ remaining>0
                         THEN MulImpl(factor,factor,Bits,Signed)
                         ELSE Ret(factor,TRUE)
           IN /\ result'=product.value /\ factor'=square.value
              /\ exp'=remaining /\ ok'=(product.ok /\ square.ok)
              /\ done'=(~ok' \/ remaining=0)
Next == Step \/ (done /\ UNCHANGED vars)
Spec == Init /\ [][Next]_vars /\ WF_vars(Step)
TypeOK == /\ exp \in 0..MaxExponent /\ ok \in BOOLEAN /\ done \in BOOLEAN
          /\ result \in MinVal(Bits,Signed)..MaxVal(Bits,Signed)
Exact == done => LET expected == PowerOracle(base,originalExp,Bits)
                IN /\ ok=Fits(expected,Bits,Signed)
                   /\ (ok => result=expected)
Terminates == <>done
=============================================================================
