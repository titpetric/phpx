# Frameworks

One folder per framework, each holding a `composer.json` and the program that
runs the framework's own entry path. git ignores `vendor/`, and a probe
starts with `composer install` in the folder.

| Folder      | Entry path                          | State |
|-------------|-------------------------------------|-------|
| `slim/`     | an app, a route and a handled request | blocked |
| `latte/`    | a template rendered to a string     | blocked |
| `laravel/`  | the framework's application class   | blocked |
| `filament/` | a panel provider                    | blocked |

A blocked folder holds no `.phpt` fixture, because a fixture runs only once the
snippet does. [report.md](../report.md) names the wall each one meets.
