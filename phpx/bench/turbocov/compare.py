#!/usr/bin/env python3
"""Compare a PHPUnit Clover report with a phpx statement profile.

usage: compare.py <clover.xml> <phpx.cov>

Both sides are reduced to the commonmark `src/` tree. Clover counts executable
lines; a phpx profile block counts statements, so a file's percentage is
compared against its own denominator and never line by line.
"""
import re
import sys
import xml.etree.ElementTree as ET

LINE = re.compile(r"^(.*):(\d+)\.\d+,(\d+)\.\d+ (\d+) (\d+)$")


def norm(path):
    # Clover names /app/src/Foo.php; a profile trims the leading separator, so
    # the same file arrives as app/src/Foo.php. Both then reduce to src/Foo.php.
    if path.startswith("/app/"):
        path = path[len("/app/"):]
    elif path.startswith("app/"):
        path = path[len("app/"):]
    at = path.find("commonmark/")
    if at >= 0:
        path = path[at + len("commonmark/"):]
    return path


def clover(path):
    rows = {}
    for file in ET.parse(path).getroot().iter("file"):
        name = norm(file.get("name", ""))
        if not name.startswith("src/"):
            continue
        lines = [int(l.get("count", "0")) for l in file.iter("line") if l.get("type") == "stmt"]
        if lines:
            rows[name] = (sum(1 for c in lines if c > 0), len(lines))
    return rows


def profile(path):
    rows = {}
    for line in open(path):
        if line.startswith("mode:"):
            continue
        m = LINE.match(line.strip())
        if not m:
            continue
        name, start, end, nstmt, count = m.groups()
        name = norm(name)
        if not name.startswith("src/"):
            continue
        covered, total = rows.get(name, (0, 0))
        rows[name] = (covered + (int(nstmt) if int(count) > 0 else 0), total + int(nstmt))
    return rows


def main():
    cl = clover(sys.argv[1])
    pr = profile(sys.argv[2])
    common = sorted(set(cl) & set(pr))
    agree = 0
    rows = []
    for name in common:
        covered, total = cl[name]
        pcovered, ptotal = pr[name]
        cp = 100 * covered / total if total else 100
        pp = 100 * pcovered / ptotal if ptotal else 100
        rows.append((abs(pp - cp), name, cp, pp, covered, total, pcovered, ptotal))
        if abs(pp - cp) <= 5:
            agree += 1
    rows.sort(reverse=True)

    print(f"files: clover={len(cl)} phpx={len(pr)} common={len(common)}")
    print(f"within 5 points: {agree}/{len(common)}")
    print(f"{'file':58} {'clover%':>8} {'phpx%':>7} {'delta':>7} {'clover':>10} {'phpx':>10}")
    for delta, name, cp, pp, covered, total, pcovered, ptotal in rows[:20]:
        print(f"{name:58} {cp:8.1f} {pp:7.1f} {pp - cp:+7.1f} {covered:5}/{total:<4} {pcovered:5}/{ptotal:<4}")
    covered = sum(cl[n][0] for n in common)
    total = sum(cl[n][1] for n in common)
    pcovered = sum(pr[n][0] for n in common)
    ptotal = sum(pr[n][1] for n in common)
    print(f"weighted: clover {100 * covered / total:.1f}% of {total} lines; phpx {100 * pcovered / ptotal:.1f}% of {ptotal} statements")


if __name__ == "__main__":
    main()
