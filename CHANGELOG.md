# Changelog

## 2026-11-03 - __CLASS__ and __TRAIT__ in a trait

- upstream is still at `e40e38b`; `go get ./...` was not run
- `__CLASS__` inside a trait method named the trait; it now reads the frame's `__class__`, which hoist set to the using class. `__TRAIT__` is a literal of the trait's name. `get_class()` with no argument answered false and names the running method's class (`stdlib/phpx/classes.go` wraps upstream's binding)
- seams: `parser/parser.go` (a `trait` field) and `parser/expr.go` (two cases), both overlaid; the logic is in fork-owned `parser/phpx_traits.go`
- `phpx/testdata/traits/magic_names.phpt`, expected output from `php`

## 2026-11-03 - __toString

- `echo $obj`, `"x" . $obj`, `strval`, `(string)`, `sprintf('%s')`, `implode` and `str_replace` now call `__toString`, own, inherited or from a trait; upstream gave an object no string form
- `model.Class.Stringer` (overlay) is set at hoist by `runner/phpx_stringer.go`; `internal/phpval` (`phpval.go`, `collection.go`, overlaid) calls it. The runtime reaches phpval through the class, so concurrent runtimes share no global
- `phpx/testdata/classes/tostring.phpt`, expected output from `php`
- upstream `e40e38b` followed first; `go get ./...` changed nothing
- found: a plain object with no `__toString` echoes empty where php raises; `class_implements()` still unregistered (`ACCEPTANCE.md`)

## 2026-11-03 - upstream e40e38b

- follow `e40e38b`, 5 commits ahead of the last one
- `migrations/run.sh` wrote the generated tree again, 1 files dropped

## 2026-11-03 - top-ten package survey and the bindings it named

- upstream is still at `2983cb8`; `go get ./...` was not run
- `phpx/testdata/packages/` lists the ten most downloaded composer libraries with code (psr/log, symfony/console, psr/http-message, guzzlehttp/psr7, psr/container, symfony/finder, guzzlehttp/guzzle, nikic/php-parser, monolog/monolog, doctrine/inflector), and `phpx/scripts/package-survey.sh` lints each one with `phpx lint --include vendor/autoload.php` into `survey.md`
- the lint named 40 undefined functions. Those with at most two arguments and no constant are bound in `stdlib/phpx/`, one file per group; the rest are flagged in `ACCEPTANCE.md` as compatibility issues
- `runner/phpx_classinfo.go` gives `class_parents()` the chain `extends` recorded
- `phpx/testdata/functions/` settles the new names against php (`vendor_basics.phpt`, `clock_arrays.phpt`)
- the widest wall is `??`: 99 of the failing files parse-fail on it, ahead of the namespaced-file statement rule (32) and array spread (19)

## 2026-11-03 - late static binding

- `static::` names the called class and not the class that wrote the body: `new static`, `static::class`, `static::CONST`, `static::method()` and `get_called_class()` answer `Admin` for `Admin::make()` where the inherited body belongs to `Model`. The frame records it as `__static__` beside `__class__`; `self::` and `parent::` forward it, a named class resets it, and a closure made in the method carries it
- `new self` and `new parent` resolve to the declaring class. Upstream refused both and `new static` as won't-build; `tests/fixtures/oop/new_self_refused.phpt` and `new_static_refused.phpt` are dropped through `patches/REMOVED`, and `phpx/testdata/classes/late_static_binding.phpt` settles php's answer
- `static::$prop` keeps resolving to the declaring class's storage, which is php's answer for a property a child inherits; a child that redeclares it shares with its parent here and not in php
- new fork-owned files: `runner/phpx_called.go` and `stdlib/core/phpx_called.go`; the seams sit in `static.go`, `runner_frame.go`, `runtime_eval.go` and `helpers.go`, all overlaid
- found: `fn() => expr` is a parse error, `__TRAIT__` is undefined, `class_parents()` and `strval()` are unregistered, and `echo $obj` with `__toString` prints nothing (a magic method, outside the fork's scope); recorded in `ACCEPTANCE.md`
- `go get ./...` was not run, upstream is still at `2983cb8`

## 2026-11-03 - version_compare()

- upstream is still at `2983cb8`; `go get ./...` was not run, so no dependency changed
- `version_compare()` was unregistered. `stdlib/core/phpx_version_compare.go` registers it with php's canonical form (separators and digit/letter boundaries split parts), the special-form order `dev < alpha < beta < RC < # < pl`, and every operator spelling (`<`, `lt`, `<=`, `le`, `>`, `gt`, `>=`, `ge`, `==`, `eq`, `!=`, `<>`, `ne`)
- `phpx/testdata/constants/version_compare.phpt` settles it against php on all three runtimes
- found while writing the fixture: `foreach ($rows as [$a, $b])` is a parse error, recorded in `ACCEPTANCE.md`

## 2026-11-02 - PHP_VERSION reaches 8.4.27

- upstream is still at `2983cb8`; `go get ./...` changed no dependency
- `PHP_VERSION` read `8.4.0`, so composer's `platform_check.php` refused a vendor tree that requires 8.4.1 or later (`laravel/framework`). `stdlib/phpx/version.go` sets `PHP_VERSION`, `PHP_RELEASE_VERSION`, `PHP_VERSION_ID`, `phpversion()` and `zend_version()` to 8.4.27, the php on this machine
- the binding sits in a new package `stdlib/phpx`, because a file in `stdlib/core` named `phpx_*.go` initialises before `platform.go` and upstream's registration would overwrite it. `stdlib/imports.go` (overlay) imports the package; the untagged build compiles an empty one
- `stdlib/core/platform_test.go` (overlay) pinned `8.4.0`; it now checks the three spellings agree
- `phpx/testdata/constants/php_version.phpt` settles it against php
- found while probing: `version_compare()` is unregistered, recorded in `ACCEPTANCE.md`

## 2026-11-02 - is_a()

- upstream is still at `2983cb8`; `go get ./...` changed no dependency
- `is_a($o, 'A')` raised `call to undefined function is_a()` while `is_subclass_of()` existed. `stdlib/core/phpx_isa.go` registers it over `Runtime.PhpxIsA` (`runner/phpx_isa.go`), which reuses `instanceof` for an object and reads a class name's ancestors and interfaces when `$allow_string` holds
- `phpx/testdata/classes/is_a.phpt` settles object, name, case, leading backslash, interface, scalar and unknown-class cases against php on all three runtimes
- the `is_a` item leaves "Discovered, not implemented" in `ACCEPTANCE.md`

## 2026-11-02 - a class name is declared once

- upstream is still at `2983cb8`; `go get ./...` changed no dependency; the previous run's commit `b938ab9` is pushed now
- `phpxClassChain` refuses a second declaration of one class name in a program, and a class named `stdClass`, with php's `Cannot redeclare class A`; the check reads the `decls` index hoist already builds, so no upstream file changes
- `phpx/testdata/classes/` holds the two fixtures; `runner/phpx_inheritance_test.go` covers the seam
- limit: php appends `(previously declared in FILE:LINE)`, and two declarations in different included files pass

## 2026-11-02 - self:: inside a constant expression

- upstream is still at `2983cb8`, the commit the last follow landed on, so there was no follow to absorb; `go get ./...` changed no dependency
- `const B = self::A + 1` raised `class constant self::A: unknown class`: `helperClassConst` evaluated a constant's expression in the scope that asked for it, which carries no class. It now evaluates in the declaring class's scope (`classDefaultScope`, the one a static default already ran in), and an interface constant in a scope named for the interface. `runner/runtime_eval.go` is an upstream file, so the whole-file copy joins `phpx/testdata/patches/runner/`
- `phpx/testdata/constants/self_in_constant.phpt` settles it against php on all three runtimes: `self::` and `K::` in a class constant, a static default, and an interface constant read through the interface and through an implementing class
- found while probing: this was the wall behind the `Illuminate` constant tables the framework probes will meet next, and it closes the discovery item in `ACCEPTANCE.md`
- limit: an inherited constant evaluates `self::` as the inheriting class, because `model.Class.Consts` keeps no declaring class; recorded under the new acceptance box

## 2026-11-02 - interfaces: the contract meets the finished class

- `implements` is checked against the class the fork hoists, not the class's own AST: a method a parent carried, a trait contributed, or an autoloaded parent brought satisfies the contract, which is how php reads the finished class. `runner/phpx_interfaces.go` moves the check to the end of `hoist`; the untagged build keeps upstream's AST check in the stub beside it
- `model/phpx_interfaces.go` adds the effective method set the linter needs and `PhpxCheckInterfaces`; `lint/lint.go` joins the overlay so `phpx lint` reads the same set and stops reporting `class B extends A implements I` as missing `f()`, and a class whose parent the file does not declare is left to the runtime
- an interface contributes its constants: `Store::LIMIT` through the implementing class and `B::LIMIT` through an interface that extends the declarer, with a class constant of the same name winning; `runner/phpx_interfaces.go` folds them into the class and into each interface declaration at hoist, and reads an interface the runtime registered from an earlier included file, which is the one-interface-per-file layout a framework uses
- a class that includes the interface's file itself (the file hoists before its include statements run), or that leaves the interface to the autoloader, folds no constant: the runtime must know the interface when the class is built
- the bytecode engine's contract check moves after its decline loop, so a program that declares `extends` or a trait is declined instead of being refused by a check that cannot see the parent; the fixture's Flat stack column skips
- `tests/fixtures/oop/interface_constants_not_acquired.phpt` and `interface_constants_extends.phpt` drop through `phpx/testdata/patches/REMOVED`: their own descriptions name php's answer, which is now the fork's
- `phpx/testdata/interfaces/interfaces.phpt` settles the whole behaviour against php on the matrix; `runner/phpx_interfaces_test.go` covers the seam and `runner/phpx_interfaces_stub_test.go` pins the untagged build
- discovered and not implemented: a constant expression that names `self::` does not resolve (`class constant self::A: unknown class` where php answers 2)
- `go get ./...` changed no dependency

## 2026-11-02 - upstream 2983cb8

- follow `2983cb8`, 4 commits ahead of the last one
- `migrations/run.sh` wrote the generated tree again, 1 file dropped
- the overlay copies of `model/ast.go`, `parser/parser.go`, `lint/names.go`, `formatter/print_decl.go` and `flatstack/engine/program_compiler.go` take upstream's hunks: `#194` adds `PropAccess.NameExpr` and has the bytecode engine decline a dynamic property name, `#191` keeps the leading backslash on a name equal to the namespace, and a whole-file copy that misses either drops the change on the floor; the build caught the missing field and the fixture tray caught the engine compiling a name it must decline
- upstream `#194` implements `$o->{'name'}` and `$o->$name`, which drops the discovery that waited on an issue; `#192` reads an out-parameter live and `#193` lists class and interface methods with `--api`
- `go get ./...` changed no dependency

## 2026-11-02 - autoloaded parents, typed property types, frameworks

- upstream is still at `672379e`, the commit the last follow landed on, and `go get ./...` changed no dependency
- `extends` resolves a parent no declaration in the program answers through the autoload queue. composer registers its loader with `--include vendor/autoload.php`, and `runner/phpx_inheritance.go` calls `rt.autoload` after the declaration list and the class table; `runner/phpx_autoload_test.go` covers the resolved parent and the unchanged missing-parent fatal
- a property type may open with `\` (`private \stdClass $providers;`), and `parser/parser.go` accepts it where a property type starts; `phpx/testdata/syntax/typed_property.phpt` settles a nullable, a union and a static typed property against php on all three runtimes, and `phpx/testdata/patches/parser/parser.go` carries the overlay copy
- the four framework snippets land under `phpx/testdata/frameworks/`, each a `composer.json` and its entry program, and `phpx/testdata/report.md` records the first wall and the walls behind it for each
- the walls the probes met, all outside the five reversals: a namespaced file with a top-level statement (composer's `nikic/fast-route` autoload file), a function declared inside a top-level `if` (the `autoload_files` guard), `??`/`??=` and `fn($x) => ...`, and `PHP_VERSION` reading `8.4.0` against a vendor tree that requires `8.4.1`
- `atkins.phpx.yml` skips `phpx/testdata/frameworks/` in the fixture loop, because a blocked framework holds no `.phpt` and the runner treats an empty tree as an error; a framework that runs adds its fixture and drops the skip
- `README.phpx.md` ticks `phpx/testdata/report.md`, and `ACCEPTANCE.md` gains an "OOP syntax" section with the typed-property box and two discovered items

## 2026-11-02 - stdClass

- `json_decode` answers stdClass for a JSON object, which is PHP's default shape and the one framework code reads with `->name`; upstream built an `*model.Array` and refused `$associative = false`
- `stdlib/core/phpx_json.go` re-registers the name from a fork-owned file: the bindings apply in contribution order, `phpx_json.go` sorts after `json.go`, and the untagged build keeps upstream's decoder
- `$associative = true` keeps the array shape; `false`, `null` and an omitted argument answer stdClass; a nested object follows at any depth; the document's order survives a decode and re-encode; a decoded object is an ordinary property bag (instanceof, foreach, unset, dynamic property, `(array)` cast)
- the extension `JSON\Decoder` is left as upstream built it, because it has no PHP counterpart and the fork changes the PHP name alone
- `phpx/testdata/stdclass/decode.phpt` settles it against php on all three runtimes: the php column passes, and the bytecode engine runs the program rather than skipping it
- `stdlib/core/phpx_json_test.go` covers the same shape from the binding under `-tags phpx`
- `phpx/testdata/README.md` moves the `stdclass/` row to done and `README.phpx.md` ticks the fifth reversal
- three gaps the reversal met and did not take, recorded under "Discovered, not implemented" in `ACCEPTANCE.md`: `is_a()` is unregistered where `is_subclass_of()` exists, `$o->{'name'}` is a parse error where PHP allows a braced property name, and a class, interface or trait named `stdClass` is accepted where php refuses to redeclare it
- `go get ./...` changed no dependency

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
