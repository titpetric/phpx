# Changelog

## 2026-11-02 - upstream 672379e

- fork of `titpetric/phpscript` at `672379e`, with `origin` on `phpscript/phpx`
- `README.phpx.md` holds the standing prompt, and `README.md` is the copy of it
- `atkins.phpx.yml` holds the fork pipeline, and `phpx/scripts/pull.sh` the rebase
- `migrations/` holds the rewrites, one step per file, run again on every pull
- `migrations/001-module-path.sh` has not run against this tree: go.mod still reads `github.com/titpetric/phpscript`
- `lefthook.yml` refuses a `README.md` staged by hand and a push where the copy drifted
- the five reversals wait for a worker: `global`, `extends`, traits, `abstract class`, `stdClass`
