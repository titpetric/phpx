#!/usr/bin/env python3
"""Merge phpx statement profiles written by separate runs of one suite.

usage: mergecov.py <out.cov> <in.cov> [<in.cov> ...]

A profile is one line per statement block, `file:start.1,end.1 count`. The
commonmark suite split into its two PHPUnit testsuites keeps the memory of one
combined process below the soft heap limit, and the two profiles then describe
disjoint executions of the same statements. Under atomic mode a block is 0 or
1, so the union takes the larger count; a block absent from one profile keeps
its count from the other. Every input must name the same mode.
"""
import sys

LINE = __import__("re").compile(r"^(.*):(\d+)\.\d+,(\d+)\.\d+ (\d+) (\d+)$")


def read(path):
    mode = ""
    rows = {}
    for line in open(path):
        line = line.strip()
        if line.startswith("mode:"):
            if mode and mode != line:
                raise SystemExit(f"{path}: mode {line!r} differs from {mode!r}")
            mode = line
            continue
        m = LINE.match(line)
        if not m:
            continue
        name, start, end, nstmt, count = m.groups()
        key = (name, int(start), int(end))
        rows[key] = (int(nstmt), max(int(count), rows.get(key, (0, 0))[1]))
    return mode, rows


def main():
    if len(sys.argv) < 3:
        raise SystemExit(__doc__.strip())
    mode, rows = read(sys.argv[2])
    for path in sys.argv[3:]:
        other_mode, other = read(path)
        if other_mode != mode:
            raise SystemExit(f"{path}: mode {other_mode!r} differs from {mode!r}")
        for key, (nstmt, count) in other.items():
            have = rows.get(key)
            if have is not None and have[0] != nstmt:
                raise SystemExit(f"{key}: {have[0]} statements against {nstmt}")
            rows[key] = (nstmt, max(count, have[1] if have else 0))
    with open(sys.argv[1], "w") as out:
        out.write(mode + "\n")
        for (name, start, end), (nstmt, count) in sorted(rows.items()):
            out.write(f"{name}:{start}.1,{end}.1 {nstmt} {count}\n")


if __name__ == "__main__":
    main()
