#!/usr/bin/env python3
"""Find counterexamples to full-width arithmetic models; report unknown honestly."""
import argparse
import hashlib
import json
from pathlib import Path
import subprocess
import time
import z3 as z

ROOT = Path(__file__).resolve().parents[2]
parser = argparse.ArgumentParser(description=__doc__)
parser.add_argument('--timeout-ms', type=int, default=20000)
parser.add_argument('--output', type=Path, default=ROOT/'review-artifacts/extended-models/z3')
args = parser.parse_args()
args.output.mkdir(parents=True, exist_ok=True)
summary = {'solver': z.get_version_string(), 'head': subprocess.check_output(
    ['git','rev-parse','HEAD'],cwd=ROOT,text=True).strip(),
    'source_sha256': {p: hashlib.sha256((ROOT/p).read_bytes()).hexdigest() for p in ('safeint.go','int.go')},
    'timeout_ms': args.timeout_ms, 'checks': []}

def check(name, defect, variables, assumptions=(), control=False):
    solver = z.Solver()
    solver.set(timeout=args.timeout_ms)
    solver.add(*assumptions, defect)
    (args.output/(name+'.smt2')).write_text(solver.to_smt2())
    start=time.monotonic()
    answer=solver.check()
    entry={'name':name,'result':str(answer),'seconds':round(time.monotonic()-start,3),'negative_control':control}
    if answer==z.sat:
        m=solver.model()
        entry['counterexample']={str(v):str(m.eval(v,model_completion=True)) for v in variables}
    if answer==z.unknown:
        entry['reason']=solver.reason_unknown()
    summary['checks'].append(entry)
    (args.output/'summary.json').write_text(json.dumps(summary,indent=2)+'\n')
    print(f'{name}: {answer} ({entry["seconds"]}s)'+ (' [negative control]' if control else ''),flush=True)
    return answer

def cast(a, source_signed, width):
    if width<a.size(): return z.Extract(width-1,0,a)
    return (z.SignExt if source_signed else z.ZeroExt)(width-a.size(),a)

# Every input bit pattern for all 64 combinations of fixed-width integer types.
for sw in (8,16,32,64):
 for ss in (False,True):
  for dw in (8,16,32,64):
   for ds in (False,True):
    a=z.BitVec('a',sw)
    b=cast(a,ss,dw)
    back=cast(b,ds,sw)
    neg_a=(a<0) if ss else z.BoolVal(False)
    neg_b=(b<0) if ds else z.BoolVal(False)
    ok=z.And(back==a,neg_a==neg_b)
    exact=cast(a,ss,65)
    lower=-(1<<(dw-1)) if ds else 0
    upper=(1<<(dw-1))-1 if ds else (1<<dw)-1
    fits=z.And(exact>=lower,exact<=upper)
    defect=z.Or(ok!=fits,z.And(ok,cast(b,ds,65)!=exact))
    check(f'convert-{sw}-{"s" if ss else "u"}-to-{dw}-{"s" if ds else "u"}',defect,[a])

for signed in (False,True):
 a,n=z.BitVecs('a n',64)
 shifted=a<<n
 restored=(shifted>>n) if signed else z.LShR(shifted,n)
 ok=z.If(a==0,True,z.If(z.UGE(n,64),False,restored==a))
 wide=cast(a,signed,128)<<z.ZeroExt(64,n)
 lower=-(1<<63) if signed else 0
 upper=(1<<63)-1 if signed else (1<<64)-1
 fits=z.Or(a==0,z.And(z.ULT(n,64),wide>=lower,wide<=upper))
 check(f'lsh64-{"signed" if signed else "unsigned"}',ok!=fits,[a,n])

# Full expressions for the actual 128-bit intermediate and reference quotient.
a,b,c=z.BitVecs('a b c',64)
product=z.ZeroExt(64,a)*z.ZeroExt(64,b)
q=z.UDiv(product,z.ZeroExt(64,c))
actual_ok=z.And(c!=0,z.ULT(z.Extract(127,64,product),c))
expected_ok=z.And(c!=0,z.ULE(q,(1<<64)-1))
check('muldiv64-unsigned-full',actual_ok!=expected_ok,[a,b,c])

abs_a=z.If(a<0,-a,a)
abs_b=z.If(b<0,-b,b)
abs_c=z.If(c<0,-c,c)
mag_product=z.ZeroExt(64,abs_a)*z.ZeroExt(64,abs_b)
mag_q=z.UDiv(mag_product,z.ZeroExt(64,abs_c))
q64=z.Extract(63,0,mag_q)
negative=z.Xor(z.Xor(a<0,b<0),c<0)
actual_ok=z.And(c!=0,z.ULT(z.Extract(127,64,mag_product),abs_c),
               z.If(negative,z.ULE(q64,1<<63),z.ULE(q64,(1<<63)-1)))
actual_value=z.If(actual_ok,z.If(negative,-q64,q64),z.BitVecVal(0,64))
reference=(z.SignExt(64,a)*z.SignExt(64,b))/z.SignExt(64,c)
expected_ok=z.And(c!=0,reference>=-(1<<63),reference<=(1<<63)-1)
expected_value=z.If(expected_ok,z.Extract(63,0,reference),z.BitVecVal(0,64))
check('muldiv64-signed-full',z.Or(actual_ok!=expected_ok,actual_value!=expected_value),[a,b,c])

# Smaller independent lemmas do not turn an unknown full query into a proof.
x=z.BitVec('x',64)
mag=z.If(x<0,-x,x)
wide_x=z.SignExt(64,x)
check('signed64-absolute-magnitude',z.ZeroExt(64,mag)!=z.If(wide_x<0,-wide_x,wide_x),[x])
p=z.BitVec('p',128)
d=z.BitVec('d',64)
check('unsigned128-high-word-bound',z.ULT(z.Extract(127,64,p),d)!=
      z.ULT(p,z.ZeroExt(64,d)<<64),[p,d])
q=z.BitVec('magnitude',64)
negative=z.Bool('negative')
ok=z.If(negative,z.ULE(q,1<<63),z.ULE(q,(1<<63)-1))
exact=z.If(negative,-z.ZeroExt(1,q),z.ZeroExt(1,q))
fits=z.And(exact>=-(1<<63),exact<=(1<<63)-1)
value=z.If(negative,-q,q)
check('signed64-quotient-sign-restore',z.Or(ok!=fits,z.And(ok,z.SignExt(1,value)!=exact)),[q,negative])

# A missing sign-change check in Convert must yield a concrete counterexample.
a=z.BitVec('a',64)
mutant_ok=cast(cast(a,True,64),False,64)==a
exact=z.SignExt(1,a)
fits=z.And(exact>=0,exact<=(1<<64)-1)
check('control-convert-without-sign-check',mutant_ok!=fits,[a],control=True)
# Treating signed right shift as logical must break the negative Lsh case.
a,n=z.BitVecs('a n',64)
wrong=z.If(a==0,True,z.If(z.UGE(n,64),False,z.LShR(a<<n,n)==a))
wide=z.SignExt(64,a)<<z.ZeroExt(64,n)
fits=z.Or(a==0,z.And(z.ULT(n,64),wide>=-(1<<63),wide<=(1<<63)-1))
check('control-lsh-logical-right-shift',wrong!=fits,[a,n],control=True)
assert len(summary['checks'])==73, len(summary['checks'])
bad=[c for c in summary['checks'] if not c['negative_control'] and c['result']=='sat']
unknown=[c for c in summary['checks'] if c['result']=='unknown']
controls=[c for c in summary['checks'] if c['negative_control']]
if any(c['result']!='sat' for c in controls): raise SystemExit('FAIL: negative control did not find a counterexample')
print(f'Completed {len(summary["checks"])} queries: {len(bad)} candidate defects, {len(unknown)} unknown.',flush=True)
raise SystemExit(1 if bad else 2 if unknown else 0)
