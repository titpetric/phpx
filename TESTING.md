# TESTING

Testing status of the fork against its objective: OOP and PHP compatibility. Measured 2026-11-04 with `phpx` built `-tags phpx` at `5f6d45a`, phpunit 11 from composer in `/tmp/pu` (`vendor/` is not tracked). Scope is [README.phpx.md](README.phpx.md#scope): extensions are out.

## Answers

**Can phpunit run without dom?** The dom is not the blocker. Two readings:

- *Entry script.* `vendor/phpunit/phpunit/phpunit` stops at `extension_loaded()` for dom, libxml, tokenizer, xmlwriter. phpx answers false for them by decision, so the script is skipped. A launcher (`require vendor/autoload.php; exit(PHPUnit\TextUI\Application::run($argv));`) with no `phpunit.xml` avoids the gate.
- *Core.* dom/xmlwriter/tokenizer are used in 49 files of phpunit, sebastian and phar-io: 36 are XML configuration migrations, plus the XML loader/validator, baseline, `Util/Xml`, the coverage and logging reports. All of it sits behind the config file, coverage, logging and `--migrate-configuration`. A plain run reaches none of it.

**Does it load?** No. `phpx run --include vendor/autoload.php launch.php FooTest.php` dies parsing `TextUI/Application.php:1138`: `$x->a()['k'] ?? false` ("?? needs a variable, index or property on its left"). It never reaches dom.

**Restrictions of use once it loads** (by design): no `phpunit.xml` (use CLI flags and a `bootstrap`), no coverage (`--coverage-*`, needs xdebug/pcov and dom), no XML/JUnit/TeamCity logging, no phpt runner, no `--migrate-configuration`, no process isolation (`ChildProcessReason`, needs proc functions). Also `PHP_VERSION` reads 8.4.0.

## Walls, by files failing `phpx lint --include vendor/autoload.php` (parse errors only)

| Dependency | Pass | Warn | Fail |
|---|--:|--:|--:|
| phpunit/phpunit (src) | 939 | 861 | 84 |
| sebastian/* | 43 | 120 | 24 |
| phar-io/* | 67 | 18 | 3 |
| nikic/php-parser | 230 | 59 | 11 |
| myclabs/deep-copy | 22 | 16 | 0 |
| staabm/side-effects-detector | 2 | 1 | 0 |

Failing files by cause (all are OOP or OOP-code syntax, so in scope; none is an extension):

| Files | Wall | Example |
|--:|---|---|
| 32 | `$obj::class` (parse stops at `::` after a variable) | `DirectDispatcher.php:58` |
| 13 | `enum` / backed enum | `ChildProcessReason.php:17` |
| 12 | `=` unexpected (named args, `readonly`/typed forms, `new` in initializer, to be pinned per file) | `TestSuiteBuilder.php:69` |
| 8 | `=>` unexpected (`match` arms) | `sebastian/.../Parser.php:263` |
| 0 | `??` with a call on the left (landed `phpx/testdata/functions/coalesce_call_test.php`; was 6) | `Application.php:1138` |
| 3 | `const VERSION` after a type (typed class constants) | `Requirement.php:26` |
| 3 | `->` after `?` (`?->`) | `ClosureComparator.php:95` |
| 3 | `<<` (heredoc/shift, to be pinned) | `CodeExporter.php:39` |
| 2 | `xor` | `NumericComparator.php:53` |

"Pass" is a lint pass, not an execution. Warnings are undefined names the lint cannot resolve and are not counted as blockers.

## Coverage per dependency

Not measurable yet: nothing from phpunit executes, so `--cover` has no lines to count. The lint pass rate above is the only figure: phpunit 51% pass / 46% warn / 4% fail, sebastian 23% / 64% / 13%. The first dependency line coverage can be read from is after `$x::class`, `enum`, call-left `??` land.

## Not done, and why

- Agents were not dispatched. The brief limits this run to TESTING.md, and bulk test work needs the parser walls fixed first (code changes); there is no run to push coverage on. Order for a later run: (1) `$expr::class`, (2) `enum`, (3) `??` on a call, (4) `?->`, (5) `match`, (6) typed const, (7) launcher run, then dispatch one agent per dependency for coverage.
- The `=` and `<<` rows are unclassified; pin them with the failing lines before building.
- Check each fix against php first (`docs/testing.md`) and keep fixtures under `phpx/testdata/syntax/`.

## Notes for the next session

- Reproduce: `cd /tmp/pu && phpx lint --include vendor/autoload.php ./vendor/phpunit/phpunit/src/...`; rebuild with `go build -tags phpx -o /tmp/phpx-t .`
- `bin/phpx` in the tree was not rebuilt by this run.
