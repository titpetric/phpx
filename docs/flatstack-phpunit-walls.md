# Flatstack and PHPUnit: the walls to a native run

The turbocov comparison leaves the bytecode engine idle. The commonmark suite
passes whole on the interpreter in minutes, where php with turbocov finishes in
seconds. This document records what the engine needs before it can run that
suite natively. It follows a probe, run on a checkout at `c59079b7`, that
turns `flatstack` on for the local PHPUnit suite and compiles every file the
suite loads.

## The probe

`phpx/bench/turbocov/phpx-flatstack.yml` is a copy of the benchmark's
`phpx.yml` with `flatstack.enabled` set. The suite starts at `launcher.php`.
The probe uses that entrypoint and one test file:

```sh
cd phpx/testdata/phpunit
phpx -f ../../bench/turbocov/phpx-flatstack.yml run launcher.php tests/ATest.php
```

The run prints the interpreter's report, `OK (1 test, 1 assertion)`. Flatstack
changes no output and takes no share of the work, for the reason in the first
wall below.

The test file alone declares the test class, so `lint --flatstack` reads the
compiler verdict for it. The command calls `flatstack.Supports`, which calls
`engine.Compile` and returns its error:

```text
$ phpx lint --flatstack tests/ATest.php
| FAIL   | ATest.php | 1 | [flatstack unsupported] flatstack: class.ATest: unsupported extends |
```

A run falls back to the interpreter on that error by contract. A run prints
no verdict. `lint --flatstack` is how to read it.

## The first wall: the engine sees one file

`Runtime.Run` asks `rt.flat` about the entrypoint alone
(`runner/runner.go:72`). `includeFile` parses an included program, hoists it
and runs it with `rt.exec` (`runner/runner_load.go:139`), and never calls
`runFlat`. A precompile pass fills `ExprCache` with flat bytecode
(`runner/precompiler.go:126`), and `includeFile` does not read it either. Only
`runFlat` reads the flat cache.

The ATest run measures the split:

| what the run does                 | count |
|-----------------------------------|------:|
| entrypoints compiled by flatstack |     1 |
| include events on the interpreter |   472 |
| distinct files loaded             |   396 |
| loaded files the compiler rejects |   289 |

PHPUnit stays on the interpreter even if every file below compiles. Closing
the language walls is necessary and not enough: `includeFile` has to run an
included program through the same engine selection `Run` makes.

The work is larger than a call swap. An included program runs in the caller's
scope, returns a value from the include and answers `*_once` on the resolved
path. The flat VM runs a top-level frame. The frame has to seed from the
caller scope and write back when the include ends. Factor the engine
selection out of `Run` into one helper, pass the caller scope to it, and call it
from `includeFile`.

## The language walls

The probe parses every file the ATest run loads and calls `engine.Compile` on
each. Of the 396 files, 289 fail. The compiler stops at the first rejected
node. It builds the class table before it walks a method body. A declaration
check hides every body wall in the same file. The first table
reports the census of declaration features across the loaded tree; the second
reports body walls seen after the probe neutralizes the declaration checks.
Both counts are floors.

### Declaration walls

| Blocker                    | Prevalence                                         | One file                                                                                      | Work to close it                                                                                                                                                                                                                                                                                            |
|----------------------------|---------------------------------------------------:|-----------------------------------------------------------------------------------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Property visibility        |              929 fields; 172 files stop here first | `vendor/phpunit/php-timer/src/Duration.php:20`                                                | The interpreter already refuses the read, the write and the call (`runner/phpx_visibility.go`, `runner/phpx_properties.go`). Pass a caller class to the VM on every property op and apply the same rules; the class table already records each field's visibility.                                          |
| `readonly` class           |              218 classes; 84 files stop here first | `vendor/phpunit/phpunit/src/Event/Value/Telemetry/Php83GarbageCollectorStatusProvider.php:19` | `runner/phpx_promotion.go` folds `readonly class` into every field's `Readonly` flag. Compile from that flattened class and refuse a second write, as `runner/phpx_properties.go` already does.                                                                                                             |
| `extends`                  |               80 classes; 15 files stop here first | `tests/ATest.php:3`                                                                           | `runner.phpxClass` already flattens a parent's members and keeps the chain in `Ancestors` and `MethodOwner`. Move the flatten into `modelutil` so `flatstack/engine` can reach it, then compile method bodies from the flattened class. `parent::` resolves through `MethodOwner` and needs no second pass. |
| `abstract` class or method | 14 classes and 7 methods; 14 files stop here first | `vendor/phpunit/phpunit/src/Event/Value/Test/Test.php:17`                                     | The interpreter refuses `new` on an abstract class and a call to a bodyless method. Record `Abstract` and `AbstractMethods` on the engine's class and apply both checks.                                                                                                                                    |
| Constructor promotion      |                                            1 class | `vendor/phpunit/phpunit/src/Runner/ResultCache/ResultCacheId.php:48`                          | The extends work closes this too. The hoist turns each promoted parameter into a field, and the engine compiles the promoted field.                                                                                                                                                                         |

### Body and call walls

| Blocker                                              | Prevalence                                | One file                                                              | Work to close it                                                                                                                                                                                                                       |
|------------------------------------------------------|------------------------------------------:|-----------------------------------------------------------------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Spread `...$args` in a call                          | 25 files stop here after the declarations | `vendor/phpunit/phpunit/src/Event/Value/Test/TestCollection.php:35`   | `opCall` pops a fixed argument count. Add an unpack flag to the call and expand an `*model.Array` or a Go slice into the argument slice. `opArray` needs the same for `[...$items]`, as `HookMethodCollection.php:68` writes it.       |
| `parent::`                                           |                                   4 files | `vendor/phpunit/phpunit/src/Framework/Constraint/IsIdentical.php:121` | The extends flatten already names each method's owner; this falls out of that change.                                                                                                                                                  |
| Named argument                                       |                                    1 file | `vendor/phpunit/phpunit/src/TextUI/Configuration/Xml/Loader.php:144`  | The parser wraps one in `Unary{Op: "named:<name>"}`. Compile a name table beside the arguments and reorder them against the callee's parameters in the VM. A host call needs the same map.                                             |
| `try` with `finally` and no `catch`                  |                                    1 file | `vendor/phpunit/php-invoker/src/Invoker.php:54`                       | The VM already runs a handler for a catch. Add the no-catch form: run the finally block and rethrow.                                                                                                                                   |
| Function declared below the top level                |                                   2 files | `vendor/phpunit/phpunit/src/Framework/Assert/Functions.php:72`        | Emit a declaration op where the interpreter hoists the nested function, so that the function exists after the block runs and not before.                                                                                               |
| `unset` of an index of a property or static property |                                   2 files | `vendor/composer/ClassLoader.php:398`                                 | Extend the unset lvalue compiler in `flatstack/engine/program_assign.go` to a property base and a static property base. The interpreter already answers both, as `Factory.php:86` and the loader show.                                 |
| `continue 2`                                         |                                    1 file | `vendor/phpunit/phpunit/src/Runner/ErrorHandler.php:352`              | `c.loops` is a stack. Map `Levels` to the nth entry and remove the refusal for a value above one.                                                                                                                                      |
| `func_get_args()`                                    |                                    1 file | `vendor/phpunit/phpunit/src/Framework/Assert/Functions.php:89`        | The compiler refuses the name at every call site. A compiled frame keeps slots and no argument list, so that the VM needs a per-frame argument vector first. The file reaches this wall only after the nested declaration compiles.    |
| PHP constructor                                      |             every class that declares one | `vendor/phpunit/phpunit/src/Event/Value/Test/TestCollection.php:38`   | `Compile` skips `__construct`, and `helperNew` runs it on the interpreter. Compile it as a method and pass the VM a constructed object with `$this` bound. Until then a compiled class still calls the interpreter for every instance. |

## What is not a wall

Traits do not appear in this run: the loaded tree declares none. The compiler
refuses one. Destructors do not appear either. The parser rewrites an
enum into a class before `Compile` sees it (`parser/phpx_enum.go`). The
compiler checks a class contract without inheriting a member from it, which
PHPUnit's interfaces allow. Casts, `switch`, nested loops, closures and
first-class callables already compile.

The probe found no wrong answer from the engine. Every rejection above names a
form the compiler declines as unsupported, and `docs/flatstack.md` lists most
of them. No fixture pins a defect, because the engine answers every program it
accepts correctly.

## The order to close them

1. Compile included files (the first wall). Without it the language work
   changes no wall-clock result.
2. Flatten the class table: `extends`, traits, promotion and `abstract` in one
   change, since all four come from the same hoist.
3. Enforce property visibility and `readonly` writes.
4. Unpack arguments and bind named ones.
5. The small forms: `try`/`finally`, nested declarations, the two `unset`
   shapes, `continue` with a level, `func_get_args`, and constructors.

The benchmark to rerun after each step is
`phpx/bench/turbocov/run.sh` with `PHPX=1`, and the per-file verdict is
`phpx lint --flatstack <file>`.
