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
| `boot.php`, `boot.sh` | 3. Boot, 4. First request, 5. Measure: the artisan bootstrap, a `GET /` through the HTTP kernel, and the statements of `app/` that ran (`--cover`). |

The scripts are copied into the app root for a run (phpx resolves `require` against the script's directory) and removed after.
