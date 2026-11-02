# Changelog

## 2026-11-02 - abstract

- `abstract class` lands: `runner/phpx_abstract.go` refuses `new` on a class that declared the modifier, and `model.Class.Abstract` carries the flag the runtime reads
- hoist refuses a concrete class that leaves a method to a child: the class's own abstract declarations first, then what a parent left, then what a trait contributed, which is php's order and php's message (`Class C contains 2 abstract methods and must therefore be declared abstract or implement the remaining methods (B::b2, A::a1)`)
- a child that implements every method instantiates and inherits the concrete methods; an abstract class with no abstract method instantiates through a concrete child; a trait's abstract method is satisfied by the using class or by a concrete method the class inherits, and the inherited body is the one that runs
- a class may not turn a method its parent implements back into a declaration without a body: php's `Cannot make non abstract method P::f() abstract in class C`
- a class's own abstract declaration is reported before traits bind and a parent's members arrive, which is php's compile-time order, and a private method without a body is refused (`Abstract function A::f() cannot be declared private`); a trait may declare a private abstract method, because the using class supplies the body
- `parent::f()` on an abstract method is php's `Cannot call abstract method A::f()`; the message names the declaring class, which `phpxDeclOwner` reads from the class table rather than from the instance
- the bytecode engine declines a program that declares an abstract class or an abstract method, so the whole program runs on the interpreter
- the linter stops reporting `abstract` as a no-op, in `lint/walk.go` and its test
- `phpx/testdata/abstract/` settles it against php: one happy-path fixture and four `error:` fixtures for the four fatals, with the php column skipped on the error fixtures because the host renders an uncaught error where php writes stderr
- the parser already recorded `abstract` on the class and the method and the formatter printed both back, so this reversal needed no overlay syntax half
- `go get ./...` changed no dependency

## 2026-11-02 - traits

- `trait Name { ... }` parses into `model.TraitDecl`, and `use A, B;` inside a class or trait body records `model.ClassDecl.Traits`; the formatter prints both back, so a trait file lints and reformats in either build
- `runner/phpx_traits.go` folds a trait's properties, statics, constants and methods into each class that names it, nested `use` lines included, and it holds the collision bookkeeping and php's three messages
- a class's own method wins over the trait's, which is php's rule; a class's own property or constant is composed with the trait's and must agree on visibility, type and default, or php's fatal names both
- two traits declaring one name compose only when their definitions agree: a shared method is always a collision, a property or constant collides only when the definitions differ, and the three messages are php's word for word (`Trait method B::foo has not been applied as C::foo`, `A and B define the same property ($x) in the composition of C`)
- a trait contributes no ancestor: `instanceof TraitName` is false, `method_exists` finds what the trait installed, and `self::` in a trait method resolves to the using class
- the bytecode engine declines a program that declares or uses a trait, so the whole program runs on the interpreter
- `phpx/testdata/traits/` settles it against php: one happy-path fixture and two `error:` fixtures for the method and property fatals, with the php column skipped on the latter because the host renders an uncaught error where php writes stderr
- limits recorded in `README.phpx.md`: `insteadof`/`as` adaptations are a parse error, a trait declared in one included file is not seen by a class in another, `__CLASS__` in a trait method answers the trait's name, and an abstract trait method the class leaves unimplemented is accepted until the abstract reversal lands
- the linter walks trait bodies and counts a trait declaration as a declared name
- the syntax half lives in the overlay, because a follow lands upstream's file verbatim: `parser/parser.go`, `model/ast.go`, the three `formatter` files and `lint/names.go` join `phpx/testdata/patches/`, and the untagged build parses a trait while taking nothing from it
- `go get ./...` changed no dependency

## 2026-11-02 - inheritance

- `extends` lands: `runner/phpx_inheritance.go` flattens the parent's fields, statics, constants and methods into the child at hoist, the child's own declaration of a name wins, and `model.Class` grows no parent field
- `parent::method()` and `parent::__construct()` resolve against the neighbour in the chain and keep `$this` bound; a closure declared in a method carries the same parent
- `self::` inside an inherited body resolves to the class that declared it (`model.Class.MethodOwner`), which is php's rule and not the called class
- `instanceof`, `get_parent_class`, `is_subclass_of` and a catch clause naming an ancestor read `model.Class.Ancestors`
- a parent that is a host class (`extends Exception`) contributes its name and no member; `parent::__construct()` on one fails, and a static property keeps one bag per class rather than php's shared storage
- the bytecode engine declines a program that declares `extends`, so the whole program runs on the interpreter
- the linter stops reporting `extends`; upstream's `oop/extends_noop.phpt` and `oop/parent_refused.phpt` are dropped through `patches/REMOVED`, and `phpx/testdata/inheritance` settles the behaviour against php
- `go get ./...` changed no dependency
- `migrations/004-atkins-state.sh` adds `.atkins/atkins.db*` to `.gitignore`: atkins' run database and its SQLite side files are run state, and a pipeline run left a tree `follow.sh` refuses as dirty

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
