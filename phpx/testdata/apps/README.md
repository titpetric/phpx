# Real apps

Five open-source Laravel and Filament apps, run through five consumer tests under php and under phpx. The results
and what they mean are in [../../TESTING.md](../../TESTING.md).

```sh
mkdir -p /data/apps && cd /data/apps
for r in relaticle/relaticle andrewdwallo/erpsaas devaslanphp/project-management realodix/urlhub akaunting/akaunting; do
  git clone --depth 1 https://github.com/$r
done
for a in */; do (cd $a && composer install --no-interaction --no-scripts --ignore-platform-reqs); done
# the production tree: the same commit with no dev dependencies
mkdir /data/apps-prod && for a in */; do mkdir /data/apps-prod/$a && (cd $a && git archive HEAD | tar -x -C /data/apps-prod/$a) \
  && (cd /data/apps-prod/$a && composer install --no-dev --no-interaction --no-scripts --ignore-platform-reqs); done
phpx/testdata/apps/run.sh bin/phpx /data/apps /data/apps-prod
```

| File | Test |
|------|------|
| `survey.sh` | 1. Parse: `phpx lint` over `app/`, `routes/`, `database/`, `config/`, `tests/`, `resources/` and the framework trees. |
| `load.php`, `load.sh` | 2. Class load: one process per class under php and phpx, `class_exists` on every class of `app/`. |
| `boot.php`, `boot.sh` | 3. Boot, 4. First request, 5. Measure: the artisan bootstrap, a `GET /` through the HTTP kernel, the statements of `app/`, the Laravel framework and all of `vendor/` that ran (`--cover`), and the classes the run declared on each runtime. |

The scripts are copied into the app root for a run (phpx resolves `require` against the script's directory) and removed after.

## An app's own suite

A suite runs through the phpunit in the app's own `vendor/`, with the minimal [launcher](../phpunit/launcher.php) that
skips the entry script's extension gate. The `--no-configuration` flag stands in for `phpunit.xml` while the DOM
binding has no `loadXML` or `DOMXPath`, so the bootstrap and the env the file names go on the command line:

```sh
cd /data/apps/urlhub
cp phpx/testdata/phpunit/launcher.php .
APP_ENV=testing APP_KEY=base64:... DB_CONNECTION=sqlite DB_DATABASE=:memory: \
  SESSION_DRIVER=array CACHE_STORE=array QUEUE_CONNECTION=sync \
  phpx run launcher.php --no-configuration tests/Feature/Auth/RegisterTest.php
```

On 2026-11-26 that prints `OK (13 tests, 75 assertions)` for urlhub's `RegisterTest`, where the same command under
php prints the same totals.
