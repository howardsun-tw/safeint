----------------------------- MODULE DecodeState -----------------------------
EXTENDS Word
CONSTANTS Bits, Signed, NullClears
VARIABLES value, previous, source, kind, number, ok, step
vars == <<value,previous,source,kind,number,ok,step>>
Domain == MinVal(Bits,Signed)..MaxVal(Bits,Signed)
Numbers == (MinVal(Bits,Signed)-1)..(MaxVal(Bits,Signed)+1)
NumericInputs == {[source |-> s,kind |-> k,number |-> n]:
                  s \in {"json","sql"}, k \in {"number","string","bytes","float"}, n \in Numbers}
OtherInputs == {[source |-> s,kind |-> k,number |-> 0]:
                s \in {"json","sql"}, k \in {"null","fraction","invalid"}}
Inputs == NumericInputs \cup OtherInputs
Accepted(s,k,n) ==
    IF k="null" THEN TRUE
    ELSE IF s="json" THEN k="number" /\ Fits(n,Bits,Signed)
    ELSE k \in {"number","string","bytes","float"} /\ Fits(n,Bits,Signed)
ExpectedValue(s,k,n,old) ==
    IF k="null" THEN IF s="json" THEN old ELSE 0
    ELSE n
Init == /\ value \in Domain /\ previous=value
        /\ source="none" /\ kind="none" /\ number=0 /\ ok=TRUE /\ step=0
Decode(input) ==
    /\ step<3 /\ step'=step+1
    /\ previous'=value
    /\ source'=input.source /\ kind'=input.kind /\ number'=input.number
    /\ ok'=Accepted(input.source,input.kind,input.number)
    /\ value'=IF ~ok' THEN value
              ELSE IF input.kind="null"
                   THEN IF input.source="sql" \/ NullClears THEN 0 ELSE value
                   ELSE input.number
Next == (\E input \in Inputs: Decode(input)) \/ (step=3 /\ UNCHANGED vars)
TypeOK == value \in Domain /\ previous \in Domain /\ ok \in BOOLEAN /\ step \in 0..3
ErrorAtomic == (step>0 /\ ~ok) => value=previous
SuccessExact == (step>0 /\ ok) => value=ExpectedValue(source,kind,number,previous)
StatusExact == step>0 => ok=Accepted(source,kind,number)
=============================================================================
