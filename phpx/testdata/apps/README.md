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

A suite runs through the phpunit in the app's own `vendor/`. The runs below pass `--no-configuration` and put the
bootstrap and the env the file names on the command line; the DOM binding loads a `phpunit.xml` since 2026-11-27,
and the app suites are due a re-measure against it. `suite.sh <phpx> <app> <test path>` passes that env and falls
back to the minimal
[launcher](../phpunit/launcher.php) for a tree that has no `vendor/bin/phpunit`:

```sh
cd /data/apps/urlhub
APP_ENV=testing APP_KEY=base64:... DB_CONNECTION=sqlite DB_DATABASE=:memory: \
  SESSION_DRIVER=array CACHE_STORE=array QUEUE_CONNECTION=sync \
  phpx run vendor/bin/phpunit --no-configuration tests/Feature/Auth/RegisterTest.php
# or: phpx/testdata/apps/suite.sh bin/phpx /data/apps/urlhub tests/Feature/Auth/RegisterTest.php
```

On 2026-11-26 that prints `OK (13 tests, 75 assertions)` for urlhub's `RegisterTest`, where the same command under
php prints the same totals. `phpx --cover --coverfile=<file> run ...` records the statements the suite executed, in
the app's `app/` tree as well as vendor: the same file reaches 43 of 94 `app/` statements in the files it loaded.

The runner decides whether a suite starts at all. PHPUnit 11 suites run (urlhub), and a Pest 3.8 suite runs too
since 2026-11-27: erpsaas's 48 tests execute and each stops at the app's `NumberFormatter`, which needs intl.
phpunit 10.5.66's launcher stops on the unbound `getopt()`; and akaunting's Feature tests overflow the Go stack once
the runner starts. The per-project verdicts are in
[../../TESTING.md](../../TESTING.md#the-apps-own-suites).
