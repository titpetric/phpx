# Changelog

## 2026-11-02 - inheritance

- `extends` lands: `runner/phpx_inheritance.go` flattens the parent's fields, statics, constants and methods into the child at hoist, the child's own declaration of a name wins, and `model.Class` grows no parent field
- `parent::method()` and `parent::__construct()` resolve against the neighbour in the chain and keep `$this` bound; a closure declared in a method carries the same parent
- `self::` inside an inherited body resolves to the class that declared it (`model.Class.MethodOwner`), which is php's rule and not the called class
- `instanceof`, `get_parent_class`, `is_subclass_of` and a catch clause naming an ancestor read `model.Class.Ancestors`
- a parent that is a host class (`extends Exception`) contributes its name and no member; `parent::__construct()` on one fails, and a static property keeps one bag per class rather than php's shared storage
- the bytecode engine declines a program that declares `extends`, so the whole program runs on the interpreter
- the linter stops reporting `extends`; upstream's `oop/extends_noop.phpt` and `oop/parent_refused.phpt` are dropped through `patches/REMOVED`, and `phpx/testdata/inheritance` settles the behaviour against php
- `go get ./...` changed no dependency

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
