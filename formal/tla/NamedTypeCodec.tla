---------------------------- MODULE NamedTypeCodec ----------------------------
EXTENDS Integers
CONSTANTS Codecs, Policy
VARIABLES original, codec, wire, decoded, ok, phase
vars == <<original,codec,wire,decoded,ok,phase>>
Number(v) == [kind |-> "number",number |-> v]
BareEncode(v,c) == CASE c="plain" -> Number(v)
                    [] c="quoted" -> [kind |-> "string",number |-> v]
                    [] c="index" -> Number(v+1)
BareDecode(w,c) ==
    CASE c="plain" -> [ok |-> w.kind="number",value |-> w.number]
      [] c="quoted" -> [ok |-> w.kind="string",value |-> w.number]
      [] c="index" -> [ok |-> w.kind="number",value |-> w.number-1]
WrappedEncode(v,c) == IF Policy="delegate" THEN BareEncode(v,c) ELSE Number(v)
WrappedDecode(w,c) == IF Policy="numeric"
                     THEN [ok |-> w.kind="number",value |-> w.number]
                     ELSE BareDecode(w,c)
Init == /\ original \in 1..7 /\ codec \in Codecs
        /\ wire=[kind |-> "none",number |-> 0] /\ decoded=0 /\ ok=FALSE /\ phase="encode"
Serialize == /\ phase="encode" /\ phase'="decode"
             /\ wire'=WrappedEncode(original,codec)
             /\ UNCHANGED <<original,codec,decoded,ok>>
Deserialize == /\ phase="decode" /\ phase'="done"
               /\ LET r == WrappedDecode(wire,codec)
                  IN /\ ok'=r.ok /\ decoded'=IF r.ok THEN r.value ELSE decoded
               /\ UNCHANGED <<original,codec,wire>>
Next == Serialize \/ Deserialize \/ (phase="done" /\ UNCHANGED vars)
TypeOK == original \in 1..7 /\ decoded \in 0..8 /\ ok \in BOOLEAN
\* Both user-defined codecs are internally consistent without the wrapper.
BareCodecSound == LET r == BareDecode(BareEncode(original,codec),codec)
                  IN r.ok /\ r.value=original
WrappedRoundTrip == phase="done" => (ok /\ decoded=original)
=============================================================================
