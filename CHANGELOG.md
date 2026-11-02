# Changelog

## 2026-11-02 - upstream 672379e

- fork of `titpetric/phpscript` at `672379e`, with `origin` on `titpetric/phpx`
- `README.phpx.md` holds the standing prompt, and `README.md` is the copy of it
- `atkins.phpx.yml` holds the fork pipeline, and `phpx/scripts/follow.sh` the one-way follow
- `migrations/` holds the rewrites, one step per file, run again on every follow
- `migrations/001-module-path.sh` renamed the module to `github.com/titpetric/phpx` over 446 files
- `lefthook.yml` refuses a `README.md` staged by hand and a push where the copy drifted
- `atkins.phpx.yml` carries a counterpart for all 24 upstream jobs, with splint measured as a delta
- `ACCEPTANCE.md` holds the checklist, one box per claim and the command that settles it
- Docker Hub serves `titpetric/phpx:latest`, with the git stamp and `phpscript` linked to `phpx` inside it
- `phpx/testdata/patches/` holds the fork's version of an upstream file, copied over the tree on every follow
- the first reversal landed: `global` binds a name to the global frame, settled against php
- the five reversals wait for a worker: `global`, `extends`, traits, `abstract class`, `stdClass`
