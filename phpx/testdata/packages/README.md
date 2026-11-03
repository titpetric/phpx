# Packages

The ten most downloaded composer libraries that carry code, each linted by
`phpx lint --include vendor/autoload.php` after `composer install` in this
folder. git ignores `vendor/`. `phpx/scripts/package-survey.sh` regenerates
[survey.md](survey.md), the pass, warning and failure count per package and the
walls behind the failures.

The list comes from `https://packagist.org/explore/popular.json` (2026-11-03,
downloads in billions). It leaves out the packages that hold no code worth a
lint of their own: the `symfony/polyfill-*` family, `symfony/*-contracts` and
`symfony/deprecation-contracts`. What remains, by downloads:

| Rank | Package                | Downloads |
|-----:|------------------------|----------:|
| 1    | `psr/log`              | 1.33      |
| 2    | `symfony/console`      | 1.25      |
| 3    | `psr/http-message`     | 1.19      |
| 4    | `guzzlehttp/psr7`      | 1.18      |
| 5    | `psr/container`        | 1.16      |
| 6    | `symfony/finder`       | 1.15      |
| 7    | `guzzlehttp/guzzle`    | 1.15      |
| 8    | `nikic/php-parser`     | 1.05      |
| 9    | `monolog/monolog`      | 1.11      |
| 10   | `doctrine/inflector`   | 0.98      |

## The rule for a missing binding

A binding with at most two arguments, and no constant, goes into
`stdlib/phpx/`, one file per group (`types.go`, `ctype.go`, `clock.go`,
`arrays.go`, `strings.go`, `environment.go`, `classes.go`). A function that
takes more arguments, or a flag from a constant family (`trigger_error`'s
level, `parse_url`'s component, `setlocale`'s category, `curl_setopt`'s option)
is not bound: the survey reports it as a compatibility issue, and the code
that calls it stays unrunnable until a decision says otherwise.
`phpx/testdata/functions/` settles each bound name against php.

## Compatibility issues the survey found

Syntax walls, which no binding fixes. Counted in failing files across the ten
packages:

- `??` and `??=` are a parse error: 99 files. The widest wall.
- a namespaced file with a top-level statement (a `use function`, an `if`
  guard, a `trait` file with a statement): 32 files
- `...$spread` in an array literal: 19 files
- `=>` after `yield`, `fn`, `match` arms: 11 files
- trailing parameters with promotion or default expressions: the `expected ")"`
  family

Functions with no binding, kept out under the rule above: `trigger_error`,
`parse_url`, `assert`, `current`/`next`/`key` (an internal array pointer
no `model.Array` carries), `stream_*`, `flock`, `proc_open`, the `curl_*`
family, `syslog`, `socket_*`, `array_walk_recursive`, `mb_strcut`,
`mt_rand`.

`trigger_deprecation` is no binding: `symfony/deprecation-contracts` declares
it inside `if (!function_exists(...))`, a conditional declaration the shared
walls in [report.md](../report.md) already name.
