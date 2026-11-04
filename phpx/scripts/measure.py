#!/usr/bin/env python3
"""Measures every test unit on its own: wall time, result and the Go statements
it covers, with a coverage build of phpx. Usage: measure.py <coverage binary> > table.md

A unit is a fixture area (phpx test --matrix <dir>/...), or one phpunit test file
of phpx/testdata/phpunit and phpx/testdata/packages. Each runs with its own
GOCOVERDIR, so the statement count is the unit's alone."""
import os, re, subprocess, sys, tempfile, time, shutil

BIN = os.path.abspath(sys.argv[1])
ROOT = subprocess.check_output(["git", "rev-parse", "--show-toplevel"], text=True).strip()
os.chdir(ROOT)

def covered(covdir):
    """Statements covered overall and in fork-owned files, and the module total."""
    out = tempfile.mktemp(suffix=".cov")
    subprocess.run(["go", "tool", "covdata", "textfmt", f"-i={covdir}", f"-o={out}"], check=True, capture_output=True)
    blocks = {}
    for line in open(out):
        m = re.match(r"(.+?):(\d+\.\d+,\d+\.\d+) (\d+) (\d+)", line)
        if m:
            f, k, n, c = m.groups()
            blocks[(f, k)] = (int(n), max(blocks.get((f, k), (0, 0))[1], int(c)))
    os.remove(out)
    total = cov = fork = 0
    for (f, _), (n, c) in blocks.items():
        total += n
        if c:
            cov += n
            if os.path.basename(f).startswith("phpx") or "/stdlib/phpx/" in f:
                fork += n
    return cov, fork, total

def run(cmd, cwd, label):
    d = tempfile.mkdtemp(prefix="cov-")
    env = dict(os.environ, GOCOVERDIR=d, GOFLAGS="")
    t = time.time()
    p = subprocess.run(cmd, cwd=cwd, env=env, capture_output=True, text=True, timeout=600)
    secs = time.time() - t
    out = p.stdout + p.stderr
    cov = covered(d) if os.listdir(d) else (0, 0, 0)
    shutil.rmtree(d)
    return out, secs, cov

rows = []
def fixture_area(path):
    out, secs, cov = run([BIN, "test", "--skip-php", "--matrix", f"./{path}/..."], ROOT, path)
    m = re.search(r"(\d+) passed, (\d+) failed out of (\d+)", out)
    res = f"{m.group(1)}/{m.group(3)}" if m else "n/a"
    rows.append((path, "fixtures", res, secs, cov))

def phpunit_file(dirpath, f):
    out, secs, cov = run([BIN, "run", "launcher.php", f"tests/{f}"], os.path.join(ROOT, dirpath), f)
    m = re.search(r"OK \((\d+) tests?", out) or re.search(r"Tests: (\d+)", out)
    total = int(m.group(1)) if m else 0
    fail = re.search(r"(?:Failures|Errors): (\d+)", out)
    bad = sum(int(x) for x in re.findall(r"(?:Failures|Errors|Skipped|Incomplete): (\d+)", out))
    res = f"{max(total - bad, 0)}/{total}" if total else "did not run"
    rows.append((f"{dirpath}/tests/{f}", "phpunit", res, secs, cov))

for base in ("phpx/testdata", "tests/fixtures"):
    for d in sorted(os.listdir(base)):
        p = os.path.join(base, d)
        if os.path.isdir(p) and d not in ("patches", "frameworks", "phpunit", "packages"):
            fixture_area(p)
for d in ("phpx/testdata/phpunit", "phpx/testdata/packages"):
    for f in sorted(os.listdir(os.path.join(d, "tests"))):
        if f.endswith("Test.php"):
            phpunit_file(d, f)

print("| Unit | Kind | Passed | Seconds | Go statements covered | of which fork-owned | % of module |")
print("|---|---|---:|---:|---:|---:|---:|")
for name, kind, res, secs, (cov, fork, total) in rows:
    pct = f"{100 * cov / total:.1f}%" if total else "-"
    print(f"| `{name}` | {kind} | {res} | {secs:.1f} | {cov} | {fork} | {pct} |")
