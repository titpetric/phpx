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
