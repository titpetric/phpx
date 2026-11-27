# oop

One application written in the OOP baseline, run on php, on upstream phpscript
and on phpx, so the fork's feature and its cost are one number each.

- `app_oop.php` uses what [README.phpx.md](../../../README.phpx.md) calls the
  OOP baseline and nothing else: an interface, a trait, an abstract class,
  `extends`, `parent::`, visibility, static storage, readonly promotion,
  `__toString` and late static binding. No extension is reached for, so a
  difference is the baseline and not a binding.
- `app_flat.php` is the same loop and the same output without a class. Every
  stack runs it, which makes it the control: it measures the interpreter the
  fork inherits, and leaves the OOP app to measure what the fork adds.

Both print the same five lines, and php's output is the reference.

## Running

```sh
# upstream's binary, built outside this tree with the same toolchain
git clone --depth 1 https://github.com/titpetric/phpscript /tmp/phpscript
(cd /tmp/phpscript && GOFLAGS= CGO_ENABLED=0 go build -o /tmp/phpscript .)

phpx/bench/oop/run.sh --phpscript /tmp/phpscript --phpx bin/phpx
```

`run.sh` takes the measurement lock and `measure.py` pins every run to one CPU
(`--cpu`, default 3; a negative value runs unpinned). `--runs` sets the sample
(default 15) and the table reports the median and the minimum. Peak RSS is the
child's `ru_maxrss`, read in a fresh interpreter so it describes that run and
not the largest one before it. `bin/phpscript` in this tree is a link to phpx,
so `run.sh` refuses it as the phpscript binary.

## What it answers, 2026-11-27

php 8.4, upstream phpscript at `30e1dce7`, phpx at `63f08cd1`, all built
`GOFLAGS= CGO_ENABLED=0`; 15 runs, one core.

### app_oop.php

| stack | median | min | peak RSS | output |
|---|---|---|---|---|
| php | 38.1 ms | 37.5 ms | 39,256 KiB | reference |
| phpscript | - | - | - | exit 1: parse error, line 20 |
| phpx | 193.2 ms | 191.1 ms | 28,144 KiB | same as php |

### app_flat.php

| stack | median | min | peak RSS | output |
|---|---|---|---|---|
| php | 34.6 ms | 34.2 ms | 39,340 KiB | reference |
| phpscript | 39.5 ms | 38.9 ms | 24,184 KiB | same as php |
| phpx | 47.8 ms | 47.0 ms | 26,868 KiB | same as php |

Three readings:

1. **The OOP app is a feature measure.** phpscript stops at
   `protected string $name = 'shape';` inside the trait: upstream reads no
   typed property in a trait body, and the fork does. Only phpx runs the app,
   and it prints php's output.
2. **The control is a cost measure.** On code every stack runs, phpx is 21%
   slower than upstream and 38% slower than php. That gap is the fork's work
   on every property read and write (hooks, magic accessors) and on every
   `new` (destructor tracking), which code without those features still pays.
3. **The OOP app costs 5.1× php.** `--cpuprofile` over 300,000 iterations puts
   64% of the time under `invokeMethod` and almost all of the rest in
   `mapaccess2_faststr` and `mapassign_faststr`: a scope and a property bag are
   maps, and the interpreter builds a frame per call. Closing that gap is a
   change to the model, not to a line.

## What this run changed

Before this run, every property access asked for a magic method by name and
`lookupPHPMethod` answered a miss by scanning all of the class's methods with
`strings.EqualFold`. That is one scan per `$this->x = ...`, per `$this->x`, and
per `new`: `lookupPHPMethod` was 11.5% of the profile.

`runner/phpx_methods.go` memoises the answer per class and spelling on the
Runtime, and the fork's call sites (`__get`, `__set`, `__isset`, `__unset`,
`__call`, `__callStatic`, `__destruct`, the property hooks) read it. Measured
in warm blocks of seven runs, alternating the two binaries six times:

| | before | after | paired |
|---|---|---|---|
| app_oop.php | 210.9 ms | 190.0 ms | 0.901 ± 0.003 |
| app_flat.php | 48.2 ms | 48.0 ms | unchanged |

- **Better.** A class's methods are fixed when it hoists, so the answer cannot
  change and the memory is paid once per class and name. The table is on the
  Runtime, so it is collected with the program; a host reusing one Runtime
  keeps the entries for the classes it has already seen.
- **Worse.** One map entry per class and spelling, and a host that reuses a
  Runtime across programs keeps every class it parsed. That is the same
  lifetime `phpxHiddenProps` already has, and the entries are small.
- **Not done.** The map access that dominates now needs indexed property slots
  and a frame that is not a map, which is a change to `model.Object` and
  `Scope` and a rewrite of the paths upstream owns. It is not a fork edit.

The paired ratio is the number to trust for a delta: absolute times move with
the machine and with how long a measurement loop runs (unpinned, on eight
cores, the same binary reads 167 ms). Both binaries are measured under one
condition, one after the other, which is what the protocol in
[docs/agents/README.md](../../../docs/agents/README.md#the-measurement-protocol)
asks for.

## Limits

- One app shape. A program that constructs no objects, or that uses hooks and
  `__get` heavily, weighs the pieces differently.
- Wall time and peak RSS only; no allocation count per operation.
- The flat control mirrors the arithmetic by hand. It is the same output and
  not the same machine code, so its ratio is an order of magnitude and not a
  conversion.
