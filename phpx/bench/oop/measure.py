#!/usr/bin/env python3
"""Time one PHP file under php, phpscript and phpx and check they agree.

Each binary runs the file once for the reference output and N times for the
timing. A binary whose output differs is reported and its time is still shown,
because the difference is the point of the OOP app.
"""

import argparse
import os
import statistics
import subprocess
import sys
import time

HERE = os.path.dirname(os.path.abspath(__file__))

# Runs one command and prints the peak RSS of that child, in KiB. A fresh
# interpreter keeps the number this child's and not the max of every earlier
# one, which is what RUSAGE_CHILDREN answers inside a long-lived process.
RSS = (
    "import resource,subprocess,sys;"
    "subprocess.run(sys.argv[1:],stdout=subprocess.DEVNULL,stderr=subprocess.DEVNULL);"
    "print(resource.getrusage(resource.RUSAGE_CHILDREN).ru_maxrss)"
)


def capture(cmd):
    done = subprocess.run(cmd, capture_output=True, text=True)
    return done.returncode, done.stdout, done.stderr


def measure(cmd, runs):
    times = []
    for _ in range(runs):
        start = time.monotonic()
        subprocess.run(cmd, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
        times.append(time.monotonic() - start)
    peak = subprocess.run(
        [sys.executable, "-I", "-c", RSS, *cmd], capture_output=True, text=True
    ).stdout.strip()
    return statistics.median(times), min(times), peak


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--php", default="php")
    ap.add_argument("--phpscript", default="")
    ap.add_argument("--phpx", required=True)
    ap.add_argument("--runs", type=int, default=15)
    ap.add_argument(
        "--cpu",
        type=int,
        default=3,
        help="CPU to pin every run to; a negative value runs unpinned",
    )
    ap.add_argument("--apps", nargs="+", default=["app_oop.php", "app_flat.php"])
    args = ap.parse_args()

    # php takes the script as its first argument; a phpscript binary takes
    # `run` first, as phpx does.
    stacks = [("php", [args.php])]
    if args.phpscript and os.path.realpath(args.phpscript) != os.path.realpath(args.phpx):
        stacks.append(("phpscript", [args.phpscript, "run"]))
    stacks.append(("phpx", [args.phpx, "run"]))

    for app in args.apps:
        script = os.path.join(HERE, app)
        _, reference, _ = capture([args.php, script])
        print(f"\n### {app}\n")
        print("| stack | median | min | peak RSS | output |")
        print("|---|---|---|---|---|")
        for name, prefix in stacks:
            cmd = [*prefix, script]
            if args.cpu >= 0:
                cmd = ["taskset", "-c", str(args.cpu), *cmd]
            code, out, err = capture(cmd)
            agree = "same as php" if out == reference else "DIFFERS"
            if code != 0:
                first = (err or out).strip().splitlines()
                agree = f"exit {code}: {first[-1][:70] if first else ''}"
                print(f"| {name} | - | - | - | {agree} |")
                continue
            median, low, peak = measure(cmd, args.runs)
            print(
                f"| {name} | {median * 1000:.1f} ms | {low * 1000:.1f} ms | "
                f"{peak} KiB | {agree} |"
            )


if __name__ == "__main__":
    main()
