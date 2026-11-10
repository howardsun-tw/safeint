#!/usr/bin/env python3
"""Execute all four extended model families, retaining failures and controls."""
import argparse
import hashlib
import json
from pathlib import Path
import re
import subprocess
import tempfile
import time

ROOT=Path(__file__).resolve().parent
REPO=ROOT.parent.parent
CHECKS=[
 ('extended-decode-signed','DecodeState','current',None),
 ('extended-decode-unsigned','DecodeState','current',None),
 ('extended-roundtrip-signed','SerializationRoundTrip','current',None),
 ('extended-roundtrip-unsigned','SerializationRoundTrip','current',None),
 ('extended-named-numeric','NamedTypeCodec','current',None),
 ('extended-named-legacy-plain','NamedTypeCodec','legacy',None),
 ('extended-named-legacy-quoted','NamedTypeCodec','legacy','WrappedRoundTrip'),
 ('extended-named-legacy-index','NamedTypeCodec','legacy','WrappedRoundTrip'),
 ('extended-named-delegate','NamedTypeCodec','counterfactual',None),
 ('extended-api-signed','APIFamilyConsistency','current',None),
 ('extended-api-unsigned','APIFamilyConsistency','current',None),
 ('extended-decode-null-control','DecodeState','negative_control','SuccessExact'),
 ('extended-roundtrip-control','SerializationRoundTrip','negative_control','RoundTrip'),
 ('extended-api-flag-control','APIFamilyConsistency','negative_control','WrapperMatches'),
 ('extended-api-input-control','APIFamilyConsistency','negative_control','InputUnchanged'),
]
EXPECTED=15
parser=argparse.ArgumentParser(description=__doc__)
parser.add_argument('--jar',type=Path,default=Path.home()/'.cache/safeint-tla/tla2tools-1.7.4.jar')
parser.add_argument('--output',type=Path,default=REPO/'review-artifacts/extended-models/tlc')
args=parser.parse_args()
assert len(CHECKS)==EXPECTED
assert hashlib.sha256(args.jar.read_bytes()).hexdigest()=='936a262061c914694dfd669a543be24573c45d5aa0ff20a8b96b23d01e050e88'
args.output.mkdir(parents=True,exist_ok=True)
summary={'head':subprocess.check_output(['git','rev-parse','HEAD'],cwd=REPO,text=True).strip(),
 'source_sha256':{f:hashlib.sha256((REPO/f).read_bytes()).hexdigest() for f in ('int.go','safeint.go')},
 'artifact_sha256':{f.name:hashlib.sha256(f.read_bytes()).hexdigest() for f in ROOT.iterdir() if f.suffix in ('.tla','.cfg')},
 'checks':[]}
for name,module,kind,expected in CHECKS:
 start=time.monotonic()
 with tempfile.TemporaryDirectory(prefix='safeint-extended-tlc-') as work:
  cmd=['java','-XX:+UseParallelGC','-Xmx1g','-cp',str(args.jar.resolve()),'tlc2.TLC',
       '-workers','1','-seed','73','-fp','0','-metadir',work,'-config',str(ROOT/(name+'.cfg')),str(ROOT/(module+'.tla'))]
  try:
   result=subprocess.run(cmd,cwd=work,text=True,stdout=subprocess.PIPE,stderr=subprocess.STDOUT,timeout=180)
   output=result.stdout;code=result.returncode
  except subprocess.TimeoutExpired as e:
   output=(e.stdout or b'').decode() if isinstance(e.stdout,bytes) else (e.stdout or '')
   output+='\nTIMEOUT: 180 seconds\n';code=124
 (args.output/(name+'.log')).write_text(output)
 match=re.search(r'Invariant ([A-Za-z0-9_]+) is violated',output)
 observed=match[1] if match else None
 correct=(code==0 and 'Model checking completed. No error has been found.' in output) if expected is None else code==12 and observed==expected
 counts=re.search(r'([\d,]+) states generated, ([\d,]+) distinct states found',output)
 entry={'name':name,'kind':kind,'expected_violation':expected,'observed_violation':observed,
        'exit_code':code,'expectation_met':correct,'seconds':round(time.monotonic()-start,2),
        'states':int(counts[2].replace(',','')) if counts else None}
 summary['checks'].append(entry)
 (args.output/'summary.json').write_text(json.dumps(summary,indent=2)+'\n')
 print(f'{name}: exit={code}, invariant={observed}, expectation_met={correct}, kind={kind}',flush=True)
assert len(summary['checks'])==EXPECTED
bad=[r for r in summary['checks'] if not r['expectation_met']]
print(f'Completed {EXPECTED}/{EXPECTED} checks; {len(bad)} unexpected outcomes. Expected violations are controls, not proofs.',flush=True)
raise SystemExit(1 if bad else 0)
