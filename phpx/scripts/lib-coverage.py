#!/usr/bin/env python3
"""Statement coverage of the nine composer libraries, from a phpx profile.

usage: lib-coverage.py <profile.cov> [--files LIBRARY]

The profile is what `phpx --cover --coverfile=P run coverage_launcher.php tests/`
writes from phpx/testdata/packages: one `file:l.c,l.c statements count` line per
statement. A library's denominator is every statement of every .php file in its
source directories, which coverage_launcher.php includes after the suites end,
so a file no test reached counts as uncovered and not as absent.
"""
import re
import sys
from collections import defaultdict

LIBRARIES = [
    ("psr/log", "vendor/psr/log/src/"),
    ("psr/http-message", "vendor/psr/http-message/src/"),
    ("psr/container", "vendor/psr/container/src/"),
    ("guzzlehttp/psr7", "vendor/guzzlehttp/psr7/src/"),
    ("guzzlehttp/guzzle", "vendor/guzzlehttp/guzzle/src/"),
    ("doctrine/inflector", "vendor/doctrine/inflector/src/"),
    ("symfony/finder", "vendor/symfony/finder/"),
    ("symfony/console", "vendor/symfony/console/"),
    ("monolog/monolog", "vendor/monolog/monolog/src/"),
]
THRESHOLD = 80.0

def main():
    profile = sys.argv[1]
    want = sys.argv[sys.argv.index("--files") + 1] if "--files" in sys.argv else None
    files = defaultdict(lambda: [0, 0])  # file -> [statements, covered]
    line = re.compile(r"^(.*):\d+\.\d+,\d+\.\d+ (\d+) (\d+)$")
    # A statement can be listed more than once (one entry per time its file was
    # registered); keep the largest count for each span.
    seen = {}
    for raw in open(profile):
        m = line.match(raw.strip())
        if not m:
            continue
        key = raw.split(" ")[0]
        stmts, count = int(m.group(2)), int(m.group(3))
        seen[key] = (m.group(1), stmts, max(count, seen.get(key, (0, 0, 0))[2]))
    for path, stmts, count in seen.values():
        files[path][0] += stmts
        if count > 0:
            files[path][1] += stmts
    rows = []
    for name, prefix in LIBRARIES:
        total = covered = 0
        for path, (s, c) in files.items():
            if path.startswith(prefix) and "/Resources/" not in path:
                total += s
                covered += c
        rows.append((name, prefix, total, covered))
    if want:
        prefix = dict(LIBRARIES)[want]
        listing = sorted(((c / s if s else 1, path, s, c) for path, (s, c) in files.items() if path.startswith(prefix)), key=lambda r: (r[0], -r[2]))
        for pct, path, s, c in listing[:25]:
            print(f"{pct*100:5.1f}%  {c:4}/{s:<4} {path[len(prefix):]}")
        return
    print("| Library | Statements | Covered | Coverage | Gate (80%) |")
    print("|---|--:|--:|--:|:-:|")
    tt = tc = 0
    for name, prefix, total, covered in rows:
        pct = 100.0 * covered / total if total else 0.0
        tt += total
        tc += covered
        gate = "pass" if (pct >= THRESHOLD or total == 0) else "FAIL"
        note = " (interfaces only)" if total == 0 else ""
        print(f"| `{name}` | {total} | {covered} | {pct:.1f}%{note} | {gate} |")
    print(f"| **all nine** | {tt} | {tc} | {100.0*tc/tt:.1f}% | |")

main()
