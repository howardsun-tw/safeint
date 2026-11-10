#!/usr/bin/env python3
"""Run bounded TLA+ checks and require every expected result, including controls."""
import argparse
import hashlib
import json
import os
from pathlib import Path
import re
import subprocess
import tempfile
import time

ROOT = Path(__file__).resolve().parent
REPO = ROOT.parent.parent
CHECKS = [
    ("arithmetic-signed", "Arithmetic", None),
    ("arithmetic-unsigned", "Arithmetic", None),
    ("doubleword-signed", "Arithmetic", None),
    ("doubleword-unsigned", "Arithmetic", None),
    ("doubleword-edges-signed", "Arithmetic", None),
    ("doubleword-edges-unsigned", "Arithmetic", None),
    ("power-signed", "Power", None),
    ("power-unsigned", "Power", None),
    ("scan", "Scan", None),
    ("lsh-contract-signed", "Arithmetic", None),
    ("lsh-contract-unsigned", "Arithmetic", None),
    ("legacy-muldiv", "Arithmetic", "FailureZero"),
    ("legacy-scan", "Scan", "ScanExact"),
    ("readme-lsh-overshift", "Arithmetic", "ReadmeOvershift"),
    ("readme-lsh-lost-bits", "Arithmetic", "ReadmeLostBits"),
]
EXPECTED_CHECKS = 15
JAR_SHA256 = "936a262061c914694dfd669a543be24573c45d5aa0ff20a8b96b23d01e050e88"


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--jar", type=Path, default=Path(os.environ.get(
        "TLA_JAR", str(Path.home()/".cache/safeint-tla/tla2tools-1.7.4.jar"))))
    parser.add_argument("--output", type=Path, default=REPO/"review-artifacts/tla-current-results")
    args = parser.parse_args()
    if len(CHECKS) != EXPECTED_CHECKS:
        raise SystemExit(f"FAIL: expected {EXPECTED_CHECKS} checks, configured {len(CHECKS)}")
    if hashlib.sha256(args.jar.read_bytes()).hexdigest() != JAR_SHA256:
        raise SystemExit("FAIL: use the pinned tla2tools 1.7.4 JAR documented in README.md")
    args.output.mkdir(parents=True, exist_ok=True)
    summary = {
        "head": subprocess.check_output(["git", "rev-parse", "HEAD"], cwd=REPO, text=True).strip(),
        "source_sha256": {p: hashlib.sha256((REPO/p).read_bytes()).hexdigest()
                          for p in ("safeint.go", "int.go", "README.md")},
        "model_sha256": {p.name: hashlib.sha256(p.read_bytes()).hexdigest()
                         for p in sorted(ROOT.glob("*.tla"))},
        "config_sha256": {p.name: hashlib.sha256(p.read_bytes()).hexdigest()
                          for p in sorted(ROOT.glob("*.cfg"))},
        "jar_sha256": JAR_SHA256,
        "checks": [],
    }
    for name, module, expected in CHECKS:
        start = time.monotonic()
        with tempfile.TemporaryDirectory(prefix="safeint-tlc-") as work:
            cmd = ["java", "-XX:+UseParallelGC", "-Xmx1g", "-cp", str(args.jar.resolve()),
                   "tlc2.TLC", "-workers", "1", "-seed", "73", "-fp", "0",
                   "-metadir", work, "-config", str(ROOT/(name+".cfg")),
                   str(ROOT/(module+".tla"))]
            result = subprocess.run(cmd, cwd=work, text=True, stdout=subprocess.PIPE,
                                    stderr=subprocess.STDOUT, timeout=180)
        (args.output/(name+".log")).write_text(result.stdout)
        passed = ((result.returncode == 0 and
                   "Model checking completed. No error has been found." in result.stdout)
                  if expected is None else
                  (result.returncode == 12 and re.search(
                      rf"Invariant {re.escape(expected)} is violated(?:\.| by the initial state:)",
                      result.stdout) is not None))
        counts = re.search(r"([\d,]+) states generated, ([\d,]+) distinct states found", result.stdout)
        entry = {"name": name, "expected_violation": expected, "exit_code": result.returncode,
                 "expectation_met": passed, "seconds": round(time.monotonic()-start, 2),
                 "states": int(counts[2].replace(",", "")) if counts else None}
        summary["checks"].append(entry)
        (args.output/"summary.json").write_text(json.dumps(summary, indent=2)+"\n")
        print(f"{'PASS' if passed else 'FAIL'} {name}: "
              f"{expected or 'invariants hold'}, states={entry['states']}, exit={result.returncode}", flush=True)
        if not passed:
            print(result.stdout)
            raise SystemExit(1)
    if len(summary["checks"]) != EXPECTED_CHECKS:
        raise SystemExit("FAIL: incomplete verification run")
    print(f"{EXPECTED_CHECKS}/{EXPECTED_CHECKS} expectations met: "
          "11 current models, 2 legacy implementation controls, 2 historical README controls.")


if __name__ == "__main__":
    main()
