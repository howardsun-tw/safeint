------------------------- MODULE SerializationRoundTrip -------------------------
EXTENDS Word
CONSTANTS Bits, Signed, LoseHigh
VARIABLES original, transport, wire, decoded, ok, phase
vars == <<original,transport,wire,decoded,ok,phase>>
Domain == MinVal(Bits,Signed)..MaxVal(Bits,Signed)
\* Numeric payloads stand for exact JSON/decimal serialization, not float64.
Encode(v, t) ==
    IF t="json" THEN [kind |-> "number", number |-> v]
    ELSE IF Fits(v,Bits,TRUE) THEN [kind |-> "int",number |-> v]
    ELSE [kind |-> "decimal",number |-> IF LoseHigh THEN Wrap(v,Bits,TRUE) ELSE v]
Init == /\ original \in Domain /\ transport \in {"json","sql"}
        /\ wire=[kind |-> "none",number |-> 0] /\ decoded=0 /\ ok=FALSE /\ phase="encode"
Serialize == /\ phase="encode" /\ phase'="decode"
             /\ wire'=Encode(original,transport)
             /\ UNCHANGED <<original,transport,decoded,ok>>
Deserialize == /\ phase="decode" /\ phase'="done"
               /\ ok'=Fits(wire.number,Bits,Signed)
               /\ decoded'=IF ok' THEN wire.number ELSE decoded
               /\ UNCHANGED <<original,transport,wire>>
Next == Serialize \/ Deserialize \/ (phase="done" /\ UNCHANGED vars)
TypeOK == original \in Domain /\ decoded \in Domain /\ ok \in BOOLEAN
RoundTrip == phase="done" => (ok /\ decoded=original)
WireKind == phase/="encode" =>
    IF transport="json" THEN wire.kind="number"
    ELSE wire.kind=(IF Fits(original,Bits,TRUE) THEN "int" ELSE "decimal")
=============================================================================
