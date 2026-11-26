#!/usr/bin/env python3
"""compare.py puts two PHPUnit JUnit reports side by side, one test at a time.

The point is the 1-1 view: every test that php and phpx both ran appears once,
with the outcome each stack gave it, so the tests phpx still gets wrong are the
only ones left to read. Both files come from the same checkout and the same
PHPUnit; the only difference is the interpreter.

    compare.py --php php.xml --phpx phpx.xml [--out report.md] [--all]

The exit status is 1 when a test passes on php and not on phpx, so a pipeline
can gate on parity. A test a stack did not run at all counts as "missing".
"""
import argparse
import os
import sys
import xml.etree.ElementTree as ET


def outcome(case):
    for child in case:
        tag = child.tag
        if tag in ("failure", "error"):
            return tag
        if tag == "skipped":
            return "skipped"
    return "passed"


def load(path):
    tests = {}
    if not os.path.exists(path):
        return tests
    root = ET.parse(path).getroot()
    for case in root.iter("testcase"):
        cls = case.get("class") or case.get("classname") or ""
        name = case.get("name") or ""
        key = f"{cls}::{name}" if cls else name
        # A JUnit file may hold one suite per stack; keep the worst outcome a
        # test took, and the sum of its time.
        got = outcome(case)
        prev = tests.get(key)
        if prev is None:
            tests[key] = {"outcome": got, "time": float(case.get("time") or 0)}
            continue
        prev["time"] += float(case.get("time") or 0)
        rank = {"passed": 0, "skipped": 1, "failure": 2, "error": 3}
        if rank.get(got, 0) > rank.get(prev["outcome"], 0):
            prev["outcome"] = got
    return tests


def counts(tests):
    out = {"passed": 0, "failure": 0, "error": 0, "skipped": 0}
    for t in tests.values():
        out[t["outcome"]] += 1
    return out


def main():
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument("--php", required=True, help="JUnit report from the php run")
    ap.add_argument("--phpx", required=True, help="JUnit report from the phpx run")
    ap.add_argument("--out", help="write the markdown report here")
    ap.add_argument("--all", action="store_true", help="table every test, not only the mismatches")
    args = ap.parse_args()

    php = load(args.php)
    phpx = load(args.phpx)
    keys = sorted(set(php) | set(phpx))
    mismatches = []
    for key in keys:
        a = php.get(key, {}).get("outcome", "missing")
        b = phpx.get(key, {}).get("outcome", "missing")
        if a != b:
            mismatches.append((key, a, b))

    lines = []
    lines.append("# fastcov 1-1 test invocation")
    lines.append("")
    lines.append("| stack | tests | passed | failures | errors | skipped |")
    lines.append("|---|---:|---:|---:|---:|---:|")
    for name, table in (("php", php), ("phpx", phpx)):
        c = counts(table)
        lines.append(
            f"| {name} | {len(table)} | {c['passed']} | {c['failure']} | {c['error']} | {c['skipped']} |"
        )
    lines.append("")
    agree = len(keys) - len(mismatches)
    lines.append(f"**{agree}/{len(keys)} tests agree** ({100.0 * agree / max(len(keys), 1):.1f}%).")
    lines.append("")
    if mismatches:
        lines.append("## Where the stacks differ")
        lines.append("")
        lines.append("| test | php | phpx |")
        lines.append("|---|---|---|")
        for key, a, b in mismatches:
            lines.append(f"| `{key}` | {a} | {b} |")
        lines.append("")
    if args.all:
        lines.append("## Every test")
        lines.append("")
        lines.append("| test | php | phpx | php s | phpx s |")
        lines.append("|---|---|---|---:|---:|")
        for key in keys:
            a = php.get(key, {})
            b = phpx.get(key, {})
            lines.append(
                f"| `{key}` | {a.get('outcome', 'missing')} | {b.get('outcome', 'missing')} |"
                f" {a.get('time', 0):.3f} | {b.get('time', 0):.3f} |"
            )

    report = "\n".join(lines) + "\n"
    if args.out:
        with open(args.out, "w", encoding="utf-8") as f:
            f.write(report)
    else:
        sys.stdout.write(report)

    failing_on_phpx = [m for m in mismatches if m[1] == "passed" and m[2] != "passed"]
    sys.exit(1 if failing_on_phpx else 0)


if __name__ == "__main__":
    main()
