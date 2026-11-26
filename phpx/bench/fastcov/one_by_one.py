#!/usr/bin/env python3
"""one_by_one.py runs one test per process, on php and on phpx, and records
what each invocation cost.

PHPUnit normally runs a whole suite in one process, so one slow or failing test
is buried in the suite total and a heap profile covers every test together.
Running exactly one test per invocation gives a 1-1 row: the test, the outcome
php gave it, the outcome phpx gave it, each wall time, and for phpx the Go heap
the invocation allocated (from its own --memprofile).

    one_by_one.py --checkout ../commonmark --tests failing.txt
    one_by_one.py --checkout ../commonmark --from-junit /tmp/phpx.xml --limit 20

A tests file holds one PHPUnit test ID a line, the `Class::method with data set
...` form the JUnit report writes. --from-junit derives the list from a report
instead: by default the tests phpx did not pass.
"""
import argparse
import json
import os
import re
import shutil
import subprocess
import sys
import tempfile
import time
import xml.etree.ElementTree as ET


def junit_tests(path, only_failing):
    out = []
    root = ET.parse(path).getroot()
    for case in root.iter("testcase"):
        cls = case.get("class") or case.get("classname") or ""
        name = case.get("name") or ""
        if only_failing:
            bad = any(child.tag in ("failure", "error") for child in case)
            if not bad:
                continue
        out.append(f"{cls}::{name}" if cls else name)
    return out


def php_filter(test_id):
    """--filter takes a PCRE over `Class::method with data set ...`.

    PHPUnit reads `Class::method#0` as one data set of that method; a plain
    test is anchored. A namespace backslash is doubled because PCRE reads `\\C`
    as a control escape, so the single spelling matches nothing.
    """
    m = re.match(r"^(.*) with data set #(\d+)$", test_id)
    if m:
        head, data = m.group(1), "#" + m.group(2)
    else:
        m = re.match(r'^(.*) with data set "(.*)"$', test_id)
        if m:
            head, data = m.group(1), "#" + m.group(2)
        else:
            head, data = test_id, ""
    head = head.replace("\\", "\\\\")
    return head + data if data else "^" + head + "$"


TEST_TIME = re.compile(r"^Time: ([0-9:.]+)", re.MULTILINE)


def phpunit_seconds(seconds):
    """The test time PHPUnit measured, which leaves the bootstrap out.

    A phpx process pays tens of seconds loading the functional test tree before
    the first test runs; PHPUnit's own clock starts after that, so the two
    numbers answer different questions.
    """
    if not seconds:
        return None
    parts = seconds.split(":")
    total = 0.0
    for part in parts:
        total = total * 60 + float(part)
    return total


def run(cmd, cwd, timeout):
    start = time.monotonic()
    try:
        proc = subprocess.run(
            cmd, cwd=cwd, stdout=subprocess.PIPE, stderr=subprocess.PIPE,
            timeout=timeout, text=True,
        )
        measured = TEST_TIME.search(proc.stdout or "")
        return (proc.returncode, time.monotonic() - start, proc.stderr,
                phpunit_seconds(measured.group(1)) if measured else None)
    except subprocess.TimeoutExpired:
        return 124, time.monotonic() - start, "timeout", None


def pprof_alloc(profile, binary):
    """The heap the invocation allocated, from phpx's own memprofile."""
    if not profile or not os.path.exists(profile):
        return None
    try:
        proc = subprocess.run(
            ["go", "tool", "pprof", "-top", "-sample_index=alloc_space", binary, profile],
            capture_output=True, text=True, timeout=120,
        )
    except (subprocess.SubprocessError, OSError):
        return None
    total = None
    rows = []
    for line in proc.stdout.splitlines():
        # pprof prints: flat flat% sum% cum cum% site
        fields = line.split(None, 5)
        if len(fields) != 6 or not fields[0].endswith("B"):
            continue
        try:
            pct = float(fields[1].rstrip("%"))
        except ValueError:
            continue
        rows.append({"flat": fields[0], "pct": pct, "site": fields[5].strip()})
    if rows:
        total = rows[0]
    return {"top": rows[:5], "total": total}


def main():
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument("--checkout", default="../commonmark", help="the commonmark checkout")
    ap.add_argument("--config", default="phpunit-plain.xml")
    ap.add_argument("--php", default="php")
    ap.add_argument("--phpx", default="phpx")
    ap.add_argument("--shim", default="phpx-shim.sh", help="the PHP_BINARY shim phpx answers")
    ap.add_argument("--tests", help="a file of test IDs, one a line")
    ap.add_argument("--from-junit", help="derive the list from this phpx JUnit report")
    ap.add_argument("--passing", action="store_true", help="with --from-junit, take the passing tests")
    ap.add_argument("--limit", type=int, default=0, help="run at most this many, 0 for all")
    ap.add_argument("--timeout", type=int, default=600)
    ap.add_argument("--out", default="one-by-one.md")
    ap.add_argument("--jsonl", default="one-by-one.jsonl")
    args = ap.parse_args()

    tests = []
    if args.tests:
        with open(args.tests, encoding="utf-8") as f:
            tests = [line.strip() for line in f if line.strip()]
    elif args.from_junit:
        tests = junit_tests(args.from_junit, only_failing=not args.passing)
    else:
        sys.exit("one_by_one: give --tests or --from-junit")
    if args.limit:
        tests = tests[: args.limit]

    phpx = os.path.abspath(args.phpx)
    shim = os.path.abspath(args.shim)
    env = dict(os.environ, PHPX_PHP_BINARY=shim)
    profiles = tempfile.mkdtemp(prefix="fastcov-profiles-")

    rows = []
    with open(args.jsonl, "w", encoding="utf-8") as log:
        for i, test in enumerate(tests, 1):
            flt = php_filter(test)
            php_code, php_s, _, php_test = run(
                [args.php, "vendor/bin/phpunit", "-c", args.config, "--filter", flt,
                 "--do-not-cache-result"],
                args.checkout, args.timeout,
            )
            profile = os.path.join(profiles, f"{i}.pprof")
            # --memprofile is phpx's own flag, so it goes before the script
            # name; everything after the script is the script's.
            px_code, px_s, px_err, px_test = run(
                [phpx, "run", "--memprofile", profile, "vendor/bin/phpunit", "-c", args.config,
                 "--filter", flt, "--do-not-cache-result"],
                args.checkout, args.timeout,
            )
            row = {
                "test": test,
                "php": "pass" if php_code == 0 else "fail",
                "phpx": "pass" if px_code == 0 else "fail",
                "php_s": round(php_s, 3),
                "phpx_s": round(px_s, 3),
                "php_test_s": None if php_test is None else round(php_test, 3),
                "phpx_test_s": None if px_test is None else round(px_test, 3),
                "phpx_alloc": pprof_alloc(profile, phpx),
            }
            if px_code not in (0, 1):
                row["phpx_note"] = (px_err or "").strip().splitlines()[-1:] or []
            rows.append(row)
            log.write(json.dumps(row) + "\n")
            log.flush()
            print(f"[{i}/{len(tests)}] {test}: php={row['php']} phpx={row['phpx']} "
                  f"{row['php_s']:.2f}s/{row['phpx_s']:.2f}s", file=sys.stderr)

    lines = [
        "# One test per invocation",
        "",
        "php s and phpx s are process walls. php test s and phpx test s are the",
        "times PHPUnit itself reports, which leave out each process's bootstrap;",
        "phpx's bootstrap loads the test tree and answers the data providers.",
        "",
        "| test | php | phpx | php s | phpx s | php test s | phpx test s | phpx alloc |",
        "|---|---|---|---:|---:|---:|---:|---:|",
    ]
    for row in rows:
        alloc = row["phpx_alloc"]["total"]["flat"] if row["phpx_alloc"] and row["phpx_alloc"]["total"] else ""
        def secs(v):
            return "" if v is None else f"{v:.3f}"
        lines.append(
            f"| `{row['test']}` | {row['php']} | {row['phpx']} | {row['php_s']:.3f} |"
            f" {row['phpx_s']:.3f} | {secs(row['php_test_s'])} | {secs(row['phpx_test_s'])} | {alloc} |"
        )
    agree = sum(1 for r in rows if r["php"] == r["phpx"])
    lines += ["", f"**{agree}/{len(rows)} agree** on one test per invocation.", ""]
    with open(args.out, "w", encoding="utf-8") as f:
        f.write("\n".join(lines))
    print("\n".join(lines))


if __name__ == "__main__":
    main()
