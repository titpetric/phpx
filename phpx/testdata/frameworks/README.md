# Frameworks

One folder per framework, each holding a `composer.json` and the program that
runs the framework's own entry path. git ignores `vendor/`, and a probe
starts with `composer install` in the folder.

| Folder      | Entry path                          | Tests | php | phpx | State |
|-------------|-------------------------------------|------:|----:|-----:|-------|
| `slim/`     | an app, routes, middleware, groups  | 5     | 5   | 5    | none |
| `latte/`    | templates rendered to a string      | 5     | 5   | 5    | none |
| `laravel/`  | container, collections, `Str`, `Arr` | 4     | 4   | 4    | none |
| `filament/` | panel classes and colour support    | 2     | 2   | 2    | passes |
| `twig/`     | templates, inheritance, macros      | 5     | 5   | 5    | none |
| `commonmark/` | markdown to html, gfm             | 4     | 4   | 4    | none |
| `yaml/`     | symfony/yaml parse and dump         | 5     | 5   | 5    | none |
| `carbon/`   | nesbot/carbon dates                 | 4     | 4   | 0    | `class Carbon extends DateTime`: the host date classes lack `createFromFormat`, `getLastErrors` and the timelib relative grammar |
| `phpparser/` | nikic/php-parser parse and print   | 3     | 3   | 3    | none |
| `mustache/` | mustache templates                  | 4     | 4   | 4    | none |
| `uuid/`     | ramsey/uuid                         | 3     | 3   | 3    | none |
| `csv/`      | league/csv read and write           | 3     | 3   | 3    | none |
| `brickmath/` | brick/math big numbers             | 3     | 3   | 3    | none |
| `symfonystring/` | symfony/string                 | 2     | 2   | 2    | none |
| `emailvalidator/` | egulias/email-validator       | 2     | 2   | 2    | none |
| `dotenv/`   | vlucas/phpdotenv                    | 2     | 2   | 2    | none |
| `collection/` | ramsey/collection                 | 4     | 4   | 4    | none |
| `eventdispatcher/` | symfony/event-dispatcher     | 2     | 2   | 2    | none |
| `routing/`  | symfony/routing                     | 3     | 3   | 3    | none |
| `faker/`    | fakerphp/faker, seeded              | 2     | 2   | 2    | none |
| `pimple/`   | pimple/pimple                       | 2     | 2   | 2    | none |

Each folder is a phpunit shim: `composer install --ignore-platform-reqs` fetches
the framework and phpunit, `launcher.php` runs `tests/` the way the packages
shim does (the recursion-context stand-in is copied to `shims/`, because an
include cannot climb out of the folder), and `summary.sh [phpx]` prints the
table above from a real run, with the first wall phpx meets. The tests were
written against php 8.4.27 first and all pass there.

Scope: the fork enables the OOP baseline, and extension compatibility is out of
scope. A framework's walls are classed in [report.md](../report.md) as in scope
(syntax and OOP) or out of scope (an extension), and only the first kind blocks
a tick in ACCEPTANCE.md.

A blocked folder holds no `.phpt` fixture, because a fixture runs only once the
snippet does. [report.md](../report.md) holds the history of the walls; the table above is the current state.

The shims above are a first look. `../upstream/` runs a library's own phpunit
suite under php and phpx, which is the larger measure.
