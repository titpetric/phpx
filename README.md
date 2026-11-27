# phpx

The [phpx](https://github.com/titpetric/phpx) project is an automated fork of [phpscript](https://github.com/titpetric/phpscript). To hold fork fidelity, phpx keeps no hand-maintained diff of upstream's files: a follow takes phpscript's tree verbatim and `migrations/` runs again over it, renaming the module path and writing the files the fork owns.

The intent of the fork is to reverse the decisions that make phpscript incompatible with code written for PHP. The decisions this fork reverses:

- phpx implements global values, `global` keyword to use
- phpx implements inheritance, `extends`
- phpx implements traits
- phpx implements `abstract class`
- phpx implements `stdClass`
- phpx implements `goto`, with its labels, inside a function body

This is the complete goal of the fork. It aims to increase compatibility with code written for PHP. For this purpose, the agent should research more compatibility gaps against PHP, but keep to OOP issues.

Agent notes:

- `goto` is the sixth reversal, decided on 2026-11-05 after the framework shims met it: laravel's `Illuminate/Support/helpers.php` (`retry()`), latte's `TagLexer` and `TagParser`. It works (`parser/phpx_goto.go`, `runner/phpx_levels.go`); the "Won't build" row in `docs/design.md` stays upstream's. A fixture settled against php comes first, and the framework shims in `phpx/testdata/frameworks/` are the test that it works.
- [ACCEPTANCE.md](ACCEPTANCE.md) holds the command that settles each claim below. This file holds the status; that file holds the proof.
- `origin` is `git@github.com:titpetric/phpx.git` and `upstream` is `titpetric/phpscript`. `phpx/scripts/follow.sh` adds the second remote when a checkout has none.
- `model.ClassDecl` carries `Parent` for `extends` and an `Abstract` flag, and the parser fills both (`model/ast.go:181`). Of the five reversals, only traits needed parser work.
- `model.Class`, the runtime form, drops `Parent`; `Abstract` survives as a flag the runtime enforces. `runner.TestNoInheritanceAtRuntime` (`runner/runner_control_test.go:128`) fails the build when `model.Class` grows a field named `Parent`, `Extends`, `Parents`, `Super` or `Base`. Flattening the parent's fields, methods, statics and constants into the child's `model.Class` leaves that test passing.
- `rt.hoist` (`runner/runner.go:150`) turns a `ClassDecl` into a `Class`. One seam there carries `extends`, traits and `abstract`.
- `extends` works. `runner/phpx_inheritance.go` flattens the parent's members at hoist, keeps the chain in `model.Class.Ancestors` for `instanceof`, `parent::` and a catch clause, and records the declaring class in `model.Class.MethodOwner` so `self::` inside an inherited body resolves to the class that wrote it. `parent::__construct()` chains a constructor. A parent no declaration in the program answers resolves through the autoload queue: `--include vendor/autoload.php` registers composer's loader before the program runs, and `runner/phpx_autoload_test.go` holds the seam. The engine declines a program that declares `extends`, so the whole of it runs on the interpreter. A class may extend a built-in throwable (`Exception`, `Error` and the SPL family, `runner/phpx_throwable.go`): `parent::__construct($message, $code, $previous)`, `getMessage()`, `getCode()`, `getPrevious()` and the catch and `instanceof` hierarchy answer as php does, while `getFile()`, `getLine()` and `getTrace()` are empty. A static property an inherited class does not redeclare shares the declaring class's storage (`runner/phpx_statics.go`). Two limits stay: any other host class as a parent contributes its name and no member, and a trait a parent's file uses is not seen, because a trait lives in the declaration list of the file that hoisted it.
- `global` works. `runner/phpx.go` binds each name to the global frame through `Scope.bindStatic`, the mechanism `static $x` already used, and `phpx/testdata/patches/` carries the dispatcher case, the bytecode engine's refusal, the linter rule it drops and the two upstream fixtures that pinned the no-op.
- `stdClass` is `model.StdClass()` and `model.NewStdClass()` (`model/stdclass.go`), and `json_decode` answers it. Upstream's decoder builds an `*model.Array` for a JSON object and refuses `$associative = false`; `stdlib/core/phpx_json.go` re-registers the name after `json.go`'s init, because the bindings apply in contribution order and the fork-owned file sorts last. `$associative = true` keeps the array shape, `false`, `null` and an omitted argument all answer stdClass, a nested object follows at any depth, and the document's order survives a decode and re-encode. The extension `JSON\Decoder` is left as upstream built it: it has no PHP counterpart, and the fork changes the PHP name alone. Two limits stay: invalid input raises where php answers null, and `$depth` and `$flags` are ignored. A script may no longer declare its own `stdClass`. `is_a()` is registered in `stdlib/core/phpx_isa.go`.
- traits work. The parser reads `trait Name { ... }` into `model.TraitDecl` and the `use A, B;` line inside a class or trait body into `model.ClassDecl.Traits`, the formatter prints both back, and `runner/phpx_traits.go` folds a trait's properties, statics, constants and methods into every class that names it. A class's own method wins over the trait's; a class's own property or constant is an ordinary definition that must agree with the trait's, as php checks it; two traits declaring one name compose only when their definitions agree, and php's fatal with php's message otherwise. A trait composes another trait through its own `use` line. A trait contributes no ancestor, so `instanceof TraitName` is false. The engine declines a program that declares or uses a trait. The adaptation block works: `A::f insteadof B;`, `B::f as g;` and `f as protected;` parse into `model.ClassDecl.TraitRules` (`parser/phpx_traits.go`), print back, and `runner/phpx_traitrules.go` applies them to the trait bodies before they compose, in a class and in a trait that composes traits. Two limits stay: a trait declared in one included file is not seen by a class in another. `__CLASS__` inside a trait method reads the frame's class (`parser/phpx_traits.go`), `__TRAIT__` names the trait, and `get_class()` with no argument names the running method's class.
- a static property named by an expression works: `Class::${expr}` and `Class::$$name` name the property at run time (`parser/phpx_staticname.go`, `model.StaticProp.NameExpr`), read, written, appended, compound assigned, incremented and `isset()`, statically and through a value base, and the bytecode engine declines the form. The `$` after `::` is the property sigil, so `$$name` names what `$name` holds. Limits: `unset()` of a static property is accepted where php raises, an undeclared one reads null where php raises, and `phpx fmt` prints the value-base `$cls::${...}` form as its intrinsic call.
- late static binding works: `static::` resolves to the called class (`runner/phpx_called.go`), so `new static`, `static::class` and `get_called_class()` name the child that inherited the method, and `new self` names the declaring class.
- method visibility works: `runner/phpx_visibility.go` refuses a private method outside its declaring class and a protected one outside its family at `$obj->f()`, `Class::f()` and `new`, with php's message (`Call to private method A::f() from scope B`). Trait `as private` and `as protected` therefore take effect. A callable (`call_user_func`, a first-class callable) skips the check.
- property visibility works: `runner/phpx_properties.go` records the visibility and declaring class of each `protected` and `private` property at hoist, and a read, write, compound assignment, `unset`, `foreach` and `get_object_vars()` outside the scope php allows raise `Cannot access protected property B::$b` or leave the property out, while `isset()` and `empty()` answer false and true. A private property of an ancestor reads as undefined from a child. A private property an ancestor declares keeps its own slot when a class redeclares the name: the ancestor's slot is stored under php's mangled key, seeded with the ancestor's default at `new`, and the ancestor's own body reads, writes, unsets and lists that slot while the redeclaring class answers its own (`phpx/testdata/classes/private_shadow_test.php`). A child's redeclaration php refuses raises php's message at hoist: a narrowed visibility, a static or readonly change, or another type (`Access level to C::$a must be protected (as in class P) or weaker`, `Cannot redeclare static P::$a as non static C::$a`, `Type of C::$a must be array (as in class P)`). Limits: a write to a private property of an ancestor that no class redeclared lands on the shared slot where php makes a dynamic property; `print_r`, `var_dump` and the `(array)` cast name a visible private or protected property plainly and a hidden slot under its mangled key; an indexed write that auto-creates an array through an unset hidden slot lands on the visible property; and the bytecode engine declines a class that declares a non-public property.
- `__destruct` works. php runs the destructor when the last reference to an instance goes and again for what is still alive when the program ends; the runtime keeps no reference count, so `runner/phpx_destruct.go` remembers every instance whose class declares the method, notes a *candidate* wherever a reference can go (`unset()`, a variable written over, a returned frame, a statement whose value nobody keeps) and destroys what a walk of every live root cannot reach, at the end of the statement that dropped it. `runner/frames.go`, `helperNew` in `runner/helpers.go` and `execUnset` in `runner/phpx_unset.go` carry the seams; the bytecode engine declines a program that declares one. Laravel's `Route::resource()` registers its routes in a `PendingResourceRegistration` destructor, and akaunting went from 282 to 616 routes, php's own count. `phpx/testdata/classes/destruct.phpt` settles the cases against php. Limits: destruction lands at the statement boundary and not inside one, an instance held only by a Go closure's capture is not seen by the walk, and the order at shutdown is reverse creation order where php's follows its object store.
- `__toString` works: `echo`, concatenation, `strval`, `sprintf` and `implode` call it through `model.Class.Stringer`, which `runner/phpx_stringer.go` sets at hoist; a class that declares the method is `Stringable`, and an object with no such method raises php's `Object of class N could not be converted to string` in `echo`, concatenation and the `(string)` cast (`runner/phpx_stringer.go`).
- constructor property promotion works: `runner/phpx_promotion.go` turns each `__construct(private int $x)` parameter that carries a modifier into a property with that visibility and a `$this->x = $x;` ahead of the body, before `extends` and traits compose. A promoted constructor in a trait declares no property, and named arguments do not parse.
- `readonly` works. The parser reads `public readonly int $x;` into `model.Field.Readonly` (`parser/parser.go`, `model/ast.go`, `formatter/print_decl.go`, all overlaid), a promoted parameter's `readonly` and a `readonly class` mark the property too, and `runner/phpx_properties.go` keeps a readonly property out of the new object until its first write. A second write, a compound assignment, an `unset` and an indexed write (`$o->list[] = 1`) raise php's `Cannot modify readonly property A::$x`, `Cannot unset readonly property A::$x` and `Cannot indirectly modify readonly property A::$x`; the first write from outside the declaring class raises php 8.4's `Cannot modify protected(set) readonly property A::$x from global scope`. The bytecode engine declines a class that declares one. A read before initialization raises `Typed property A::$x must not be accessed before initialization`, and `isset()` answers false. Limits: an uninitialized property sits after the initialized ones in `print_r` and `var_dump`, `clone` copies the value but a `__clone` cannot re-initialize it, and the 8.3 wording of the scope message differs from 8.4.
- `Closure::bind` rebinds `$this` and the class scope (`runner/phpx_closure.go`, registered in `stdlib/phpx/classes.go`), which property visibility reads; composer's `ClassLoader` depends on it. `bindTo()` and `call()` answer through `phpxClosureMethod`, one seam in the overlaid `runner/dispatch_method.go`; `Closure::fromCallable` scope is not wired.
- `rt.Vendor(fsys)` (`runner/phpx_vendor.go`) mounts a source tree, typically a `go:embed` of a composer `vendor/`, behind `Options.RootFS` and defaults `Options.Include` to `vendor/autoload.php`. It is opt in. The standing objective is phpunit run from its source tree this way, not from `phpunit.phar`, with a `phpunit.yml` and `phpx` passing `*Test.php` files to it. A namespaced file may hold a top-level statement, and a function declared inside an `if` or a function body is declared when control reaches it (`runner/phpx_funcdecl.go`, one seam in `exec`; the untagged build keeps upstream's no-op), which cleared `deep_copy.php`. Argument unpacking, `f(...$args)`, parses and runs (`runner/expr/phpx_spread.go`; arrays only). Arrow functions, `fn($x) => expr`, parse and run (`parser/phpx_arrow.go`; the body is a `return`, the closure captures every other variable the body names by value, and `fn()` with no `=>` after it stays upstream's function-name reading). `??`, `??=`, `<=>`, `and`, `or` and `xor` are upstream's since `6904f62` (`model.Coalesce`, `model.Binary`, the bytecode opcode `opJumpNotNil`); the fork's lowerings are gone. `$obj::class` parses and runs (`parser/phpx_classof.go`, lowered to `get_class($obj)`). `?->` parses and runs (`parser/phpx_nullsafe.go`; the whole rest of the chain sits in the else arm of `($t = a) === null ? null : $t->b->c`, so a null short-circuits it as php does). phpunit now runs from its source tree: `cd phpx/testdata/phpunit && ./compare.sh phpx` runs five test files under php and under phpx, `#[DataProvider]` and `#[TestWith]` included, and prints what php prints. The syntax that took (destructuring, heredoc, `and`/`or`/`xor`, named arguments, `match`, `enum`, `clone`, `throw` as an expression, `use A\{B, C}`), the library (Reflection with attributes, the SPL structures and iterators, `JsonSerializable`, `php://memory`) and the semantics (arrays copy where php copies, `parent::` through a chain, `instanceof Traversable`) are in [phpx/TESTING.md](phpx/TESTING.md) with the before and after numbers and the coverage. What stays: a throwable records no file or line, and a namespaced `const X = 1;` still defines the unqualified name. Two pull requests turn overlays into registrations; [phpx/testdata/UPSTREAM.md](phpx/testdata/UPSTREAM.md) lists the rest.
- arrays copy on write. `model.Array` holds a `shared` flag (`model/array.go`): `Clone` answers a second handle over one storage and every mutator calls `Split` first, which copies the storage and sets a separate handle on each nested array. The first write to either side is the only copy the pair pays. `modelutil.CopyArray` and `CopyValue` are that clone, and the write paths that write into an array a container holds split it on the way: `runner/assign_container.go` for a nested index or property write, `runner/phpx_value.go` (`phpxSplitBase`: `unset`, a by-reference `foreach`, an array builtin's in-place argument), `runner/phpx_refs.go` for `$r = &$a['k']`, `runner/runner_assign.go` for a `list()` element, and the bytecode engine's `opVivifyIndex` and the interpreter's `__phpx_splitbase` match them. `modelutil.ArrayWriteArg` names each builtin that writes an array argument in place, and its call site resolves the argument through the container, which makes `sort($q['k'])` write `$q` and leave the array `$q` holds a copy of unchanged; a builtin that builds a new array from another array's values clones each array element with `modelutil.CopyArrayValue`, which makes `array_merge($a, [])` answer a result that writes independently. `phpx/testdata/arrays/copy_on_write.phpt` settles it against php on all three columns; [phpx/TESTING.md](phpx/TESTING.md) holds the rule and the limits (the memory walk counts a clone's entries twice).
- property hooks work (php 8.4): `parser/phpx_hooks.go` lowers `get` and `set` hooks to methods named for the property and `runner/phpx_hooks.go` calls them from a read and a write; the hook's own access reaches the backing value. `private(set)` parses and is not enforced; a compound write and `unset` skip the hook. `phpx/testdata/hooks/` settles it against php.
- `abstract` works. `runner/phpx_abstract.go` refuses `new` on a class that declared the modifier and a call to a method with no body, and hoist refuses a concrete class that leaves a method to a child, with php's message and php's order: the class's own declarations, then what the parent left, then what a trait contributed. A child that implements every method instantiates; an abstract class with no abstract method instantiates through a concrete child; a trait's abstract method is satisfied by the class that uses the trait or by a concrete method it inherits. A class may not turn a method its parent implements back into a declaration without a body, and a class's own abstract declaration is reported before traits bind, which is php's compile-time order; a private method without a body is refused, while a trait may declare one because the using class supplies the body. The parser already recorded both spellings and the formatter printed them back, so no overlay syntax half was needed. The engine declines a program that declares an abstract class or an abstract method, so the whole of it runs on the interpreter.
- An interface is a contract and its constants. `extends` and traits satisfy a contract: `class B extends A implements I` needs no second `f()`, and `class C implements I { use T; }` takes it from the trait, because the runtime and the linter read the class after the fork has folded both in. `Store::LIMIT` through the implementing class and `B::X` through `interface B extends A` answer as php answers them, and a class constant of the same name wins. `phpx/testdata/interfaces/` settles it against php, and the two upstream fixtures that pinned the old answers are dropped through `phpx/testdata/patches/REMOVED`. An interface the runtime has not loaded when the class hoists contributes no constant, so the interface's file loads first, and a constant expression may name `self::`, which `runner/runtime_eval.go` (an overlay) resolves in the declaring class's scope; an inherited constant reads `self::` as the inheriting class, a limit recorded in [ACCEPTANCE.md](ACCEPTANCE.md).
- Of the five reversals, traits alone needed parser work, so its syntax half travels in the overlay: `parser/parser.go`, `model/ast.go`, `formatter/print_decl.go`, `formatter/print_stmt.go`, `formatter/printer.go` and `lint/names.go` sit under `phpx/testdata/patches/`. The untagged build parses a trait and takes nothing from it, the same one-direction rule `extends` follows.
- Coverage records a line two ways, named after go's `-covermode`: `count` increments a counter per hit, `atomic` marks the line covered once and stops writing, so `mode: count` reaches counts above one and `mode: atomic` answers 0 or 1. `phpx test --covermode` picks one and `runner/coverage/phpx_default.go` raises the untagged default `count` to `atomic` under `-tags phpx`, where a profile is reviewed and not counted. The pointer-keyed statement table and the mode branch are upstream's since `30e1dce7`.
- Names fold once, at registration. A PHP function, class or interface name folds by lowercase, so a lookup is exact then folded and never a scan; a Go binding's method or field folds by lowercase with underscores removed, which makes `get_all`, `getAll` and `GetAll` one key, as the binding name they are. A PHP variable stays exact: `$A` is not `$a`. `phpx/bench/testcov` measures the suite under both modes.

## Scope

The scope of the fork is the OOP baseline: the five reversals and the class syntax and semantics code written for PHP leans on. PHP extensions are out of scope, and a testing criterion that needs one is not a criterion of this fork.

Agent notes:

- **Strict scope check.** Before a snippet, a package or a framework counts as a blocker, name the wall. A wall is in scope when it is OOP (a class, an interface, a trait, visibility, `static::`, `readonly`, `__toString`, promotion), or a syntax form the OOP code is written in (`??`, `fn`, `?->`, `match`, `enum`, `<=>`). A wall that is an extension is out of scope: report it in [phpx/testdata/report.md](phpx/testdata/report.md) and do not build it. A criterion that passes only with an extension enabled settles nothing about the fork.
- **Extensions that stay off the table:** xmlreader, SimpleXML, intl, curl, sockets, pcntl, posix, gd, session, iconv, fileinfo, Phar. `extension_loaded()` answers false for each and stays that way. `libxml`, `xmlwriter`, `tokenizer` and `xml` answer true because each is bound in Go for phpunit's entry gate (`libxml_get_errors()`, `XMLWriter`, `token_get_all()` and `token_name()`, `utf8_encode()`; xml's expat parser stays unbound), and the host's extension stays off.
- **What may land.** PDO, `openssl` and a DOM subset, when a measured wall needs one and the surface is small, and more standard library surface (`standard`, `SPL`, `ctype`, `json`, `pcre`, `date`, `random`, and the small `mbstring` and `filter` helpers a package guards with `function_exists`). A binding follows the rule in [phpx/testdata/packages/README.md](phpx/testdata/packages/README.md). The three are bound in Go, not enabled from the host: `extension_loaded()` answers true for each because the marker is present, and a criterion that passes through them is fork work. `openssl` is the AES-CBC/GCM encrypter Laravel's `Encrypter` calls ([stdlib/px/phpx/openssl.go](stdlib/px/phpx/openssl.go)); the DOM subset is the parse, traversal and serialisation phpunit's XML configuration and Laravel's console renderer read ([stdlib/px/phpx/dom.go](stdlib/px/phpx/dom.go)). The reassessment on 2026-11-26: three of the five apps stopped on `openssl_cipher_iv_length()` and phpunit's configuration loader stopped on `DOMDocument`, and both are walls no OOP fix moves.
- **phpunit runs, with its XML configuration.** The entry script's extension gate answers for the names it needs, so `phpx run vendor/bin/phpunit` starts. `DOMDocument::loadXML`, `DOMXPath` and the writable `preserveWhiteSpace` are bound (2026-11-27), so `phpunit.xml` loads: `cd phpx/testdata/phpunit && ./compare.sh phpx --config` prints what php prints. urlhub's `tests/Feature` tree runs to a report; the tests that still fail stop on the walls [phpx/TESTING.md](phpx/TESTING.md) ranks, not on the request, and the other apps' runners stop before their tests (Pest's `eval`, akaunting's recursion, phpunit 10.5.66's `getopt()`).
- PDO is bound. `stdlib/px/phpx/pdo.go` and its three neighbours carry `PDO`, `PDOStatement` and `PDOException` over `database/sql`: sqlite through a private modernc driver, mysql and pgsql through the keyword DSN each speaks. `phpx/testdata/pdo/` settles the connection, both bind styles, the fetch modes, transactions, the error modes and `sqliteCreateFunction` against php, and the binding moved urlhub and project-management from the bootstrap to a response (2026-11-18). It answers what the apps call, and `fetchAll()` answers php's whole mode list: `FETCH_GROUP`, `FETCH_UNIQUE`, `FETCH_FUNC`, `FETCH_NAMED` and `FETCH_KEY_PAIR` beside the flat modes, `FETCH_CLASS` runs the class's constructor after it fills the row's properties, and `FETCH_INTO` fills the object `setFetchMode` was given ([phpx/testdata/pdo/fetchmodes_test.php](phpx/testdata/pdo/fetchmodes_test.php)). [phpx/TESTING.md](phpx/TESTING.md) names the limits: eager rows, no `bindColumn`, `bindParam` keeps the value it was given, and `getColumnMeta` answers name and type. An in-memory sqlite database is named in shared-cache mode and held open by a keeper connection, so `sqliteCreateFunction` reaches it after the first statement, as php's one connection does (`phpx/testdata/pdo/udf_memory_test.php`).
- **phpunit runs from its source tree.** `phpx/testdata/phpunit` composes phpunit and runs `phpx run vendor/bin/phpunit --testdox tests/ProviderTest.php`. A small launcher that calls `PHPUnit\TextUI\Application::run($argv)` stays for a tree that lacks one of the seven extension markers. The configuration loader reads `phpunit.xml` through the bound DOM subset. libxml's XSD validation, XML logging and the `phpt` runner need an extension and are not supported.
- **Extension totals.** [phpx/testdata/report.md](phpx/testdata/report.md) carries the count of calls and class uses per extension for every tested project, from `php phpx/scripts/extension-tally.php DIR`, and the biggest non-OOP blockers of each.

## The components

Every part the fork adds, with a box that ticks when the part works. A worker ticks a box in the commit that earns it, and adds a line for a component this list does not name yet. A box stays open while the component exists and fails.

The harness a managed fork needs before a reversal can land:

- [X] `README.phpx.md`, the standing prompt and this list
- [X] `ACCEPTANCE.md`, one box per claim and the command that settles it
- [X] `CHANGELOG.md`, the scratchpad, one entry per follow
- [X] `atkins.phpx.yml`, a counterpart for all 24 jobs in `atkins.yml`
- [X] `migrations/run.sh`, every step in order
- [X] `migrations/001-module-path.sh`, the module rename and the import pass
- [X] `migrations/002-readme.sh`, `README.md` from `README.phpx.md`
- [X] `migrations/003-overlay.sh`, `phpx/testdata/patches/` copied over the tree
- [X] `phpx/scripts/overlay.sh`, the read-back and the report behind that copy
- [X] `phpx/scripts/splint-fix.sh`, `splint fix` with upstream's markdown kept
- [X] `migrations/README.md`, the contract a step holds to
- [X] `phpx/scripts/follow.sh`, upstream's tree verbatim, the steps over it, one commit
- [X] `phpx/scripts/lint-delta.sh`, splint measured against upstream
- [X] `lefthook.yml`, the `README.md` guard on commit and on push
- [X] `phpx/docker/Dockerfile`, the image. Docker Hub serves `titpetric/phpx:latest`, which carries the git stamp and answers `version` as `phpx` and as `phpscript`.
- [ ] `atkins -f atkins.phpx.yml` green from end to end, eleven legs including the 326-fixture matrix on all three runtimes Open, and open before this run: `lint-delta.sh` fails the first leg. At `e1e250e` it counted 62 `grouping/filename` findings against upstream's 1 (an exported `Phpx*` symbol sits in a file named for the area and not for the symbol), 3 `func-args/order`, 4 `coverage/uncovered`, 1 `filecheck/long` and 2 `imports/format`, and the 2026-11-04 run fixed the imports and the pairing. Against `upstream/main` at `30e1dce7`, measured on 2026-11-27, the leg counts 74 `grouping/filename`, 164 `godoc/internals` against upstream's 2, 187 `prose/semicolon` against 2, 26 `prose/placement`, 25 `prose/shape`, 18 `pairing/unpaired`, 10 `godoc/outweighs`, 7 `func-args/order`, 5 `imports/format` and 2 `func-returns/order`, and the run added none of them. Every other leg passes: 328 upstream and 237 fork fixtures on all three runtimes, and `go test ./...` and `go test -tags phpx ./...` under both builds.

The overlay a reversal lands in:

- [X] `runner/phpx.go`, `//go:build phpx`, the overlay's answer in the package that needs it
- [X] `runner/phpx_stub.go`, `//go:build !phpx`, upstream's answer, which keeps an untagged build reviewable
- [X] `phpx/testdata/patches/`, the fork's version of an upstream file
- [X] `phpx/testdata/`, one folder per reversal, documented in its own README.md
- [X] `phpx/testdata/phpunit/`, phpunit 11 from its source tree, five test files and `compare.sh` (`atkins -f atkins.phpx.yml test:phpunit`)
- [X] `phpx/TESTING.md`, the before and after numbers, the coverage and the open walls
- [X] `phpx/testdata/UPSTREAM.md`, the seams that turn an overlay into a registration, and the pull requests that open them
- [X] `stdlib/phpx/`, for a reversal that needs a binding
- [X] `stdlib/px/phpx/pdo.go`, `PDO`, `PDOStatement` and `PDOException` over sqlite, mysql and pgsql, settled by [phpx/testdata/pdo/](phpx/testdata/pdo/)
- [X] `stdlib/px/phpx/openssl.go`, the AES-CBC/GCM encrypter laravel's `Encrypter` calls, settled by `TestOpenSSL` against php 8.4's ciphertexts
- [X] `stdlib/px/phpx/dom.go`, the DOM subset phpunit's XML configuration and the console renderer read, settled by `TestDOM`
- [X] `phpx/testdata/report.md`, the blockers and the standard library each snippet reached for
- [X] `phpx/testdata/streams/`, `stream_filter_register()` and `stream_wrapper_register()` as php's registries, which composer's autoload files call

The five reversals, each a folder under `phpx/testdata/` and a fixture that php settles first:

- [X] `global` values and the `global` keyword that reads them
- [X] inheritance, `extends`
- [X] traits, `trait` and `use`
- [X] `abstract class`
- [X] `stdClass`

The frameworks, one snippet tree each under `phpx/testdata/frameworks/`:

- [X] Slim router
- [X] Latte template engine
- [X] Filament
- [X] Laravel

What the follow already holds, with no box left to tick:

- [X] the module path reads `github.com/titpetric/phpx`, over 446 files
- [X] `README.md` is the generated copy, and a hand edit cannot land
- [X] `origin` is the fork and `upstream` is phpscript
- [X] a follow takes upstream's tree, runs the steps and commits without a hand on it

## The fork

The solution is a fork of phpscript. As such:

1. the fork follows new changes in one direction, phpscript to phpx, and never the other way,
2. prefer an addition over a fix, and treat an edit to an existing file as the last option

Agent notes:

- `atkins -f atkins.phpx.yml` is the whole gate. Every job in `atkins.yml` has a counterpart there, the binary is `bin/phpx` with `bin/phpscript` linked beside it for the shebang fixture, and `default` ends on `fmt` as upstream's does: `test:introspection` regenerates two tracked documents, the generator wraps its prose where mdox does not, and a run without that leg ends on a diff. `fmt` runs `splint fix` through `phpx/scripts/splint-fix.sh`, because splint rewrites markdown as well as Go and the fork keeps upstream's prose. `gen` stays out, and a follow refuses a dirty tree.
- `phpx/scripts/lint-delta.sh` stands in for the bare `splint ./...`. Upstream main reports 236 `godoc/outweighs` findings against the splint on this machine, and the fork's rule is to add none of its own. No checkout of this tree reaches the zero upstream claims.
- `atkins -f atkins.phpx.yml phpx:follow` is that follow, and `phpx/scripts/follow.sh` holds the logic: an `upstream` remote when the checkout has none, the removal of every path upstream stopped carrying, upstream's tree checked out verbatim, `migrations/run.sh` over the result, and one commit.
- The follow neither pulls nor merges. A pull asks git to reconcile two edits of one file, and the fork's edit of an upstream file is a migration over the whole tree: git would report a conflict on every file the rename touched, each one resolved by running the step again. The step runs instead, and no file has two authors.
- Three things come out of that. The history is append-only, and every push is a fast-forward. A conflict cannot happen. An edit the fork wants inside an upstream file belongs in a migration step, because upstream's version of that file lands verbatim and a committed edit to it disappears.
- `follow.sh` carries the list of files the fork owns, and the policy above reads back as code there: `README.phpx.md`, `ACCEPTANCE.md`, `CHANGELOG.md`, `atkins.phpx.yml`, `lefthook.yml`, everything under `phpx/`, `migrations/` and `stdlib/phpx/`, and any file named `phpx.go` or `phpx_*.go` in any package. A new fork-owned path goes on that list in the commit that creates it. [migrations/README.md](migrations/README.md) carries the contract a step holds to.
- `README.md` in the fork is the copy of this file that `migrations/002-readme.sh` writes. A follow takes upstream's version and the step writes over it, and the one file both projects edit produces no conflict. `lefthook.yml` refuses a commit that stages `README.md` by hand and a push where the copy drifted, and the pipeline passes `PHPX_README=1` to write it.
- `CHANGELOG.md` is the fork's scratchpad. Every follow records the upstream commit it landed on, and an agent adds what the overlay had to adjust under that entry.
- Migrations are for a change that reaches the whole tree. The overlay itself is hand-written code in a fork-owned file, and a seam of one line inside an upstream file is a step, because the follow overwrites that file.
- atkins opens its run database at `.atkins/atkins.db`, beside the tracked skills it reads, and SQLite leaves `-shm` and `-wal` files there while a run holds one. `migrations/004-atkins-state.sh` ignores the three in `.gitignore`, because a pipeline run otherwise leaves a tree `follow.sh` refuses as dirty.

## Local changes over upstream files

`phpx/testdata/patches/` mirrors the source tree: `phpx/testdata/patches/runner/runner.go` is the fork's `runner/runner.go`. `migrations/003-overlay.sh` copies each one over the tree after a follow has landed upstream's version, and that is how a local change to an upstream file survives.

Agent notes:

- The overlay carries a whole file. A diff fails on context drift and leaves the tree half-migrated, and a copy always applies. The cost is that an upstream edit to an overlaid file goes unseen, and `atkins -f atkins.phpx.yml phpx:overlay:report` is what reads it back: it diffs each overlay against `upstream/main` and names the lines the fork drops. Read it after a follow.
- The overlay sits under `testdata` because the go tool ignores every path with a `testdata` element. Without that, `go build ./...` compiles the copies as packages of their own and splint reads them as source: the first run of this mechanism reported a 1218-line file and three missing test files, every one of them a copy. `test:phpx:fixtures` skips the folder for the same reason, since the fixture runner would otherwise run an overlaid fixture twice.
- `phpx/testdata/patches/REMOVED` lists the upstream paths the fork drops. The removal runs before the copy, and a path there cannot also be overlaid.
- An overlay with no file under it in the tree is an addition, and that is the fork's route to a new file inside an upstream directory. `follow.sh` drops a tracked path upstream does not carry, and the step writes it again.
- `phpx:overlay:capture` reads every overlaid file back out of the tree, and `default` ends on it. A job in the run edits an overlaid file, gofmt regrouping an import or mdox rewrapping a paragraph, and without the read-back the overlay goes stale and the next follow installs the old version over the fixed one.
- Prefer a fork-owned file to an overlay. A new file named `phpx.go` or `phpx_*.go` in the package that needs it (or any file under a directory named `phpx_*`, which holds what the Go files embed) survives a follow by name and carries no review burden. The overlay is for the line that has to sit inside an upstream file: a `case` in a dispatcher, an expected output in a fixture, a rule the linter stops reporting.

## The module path

`migrations/001-module-path.sh` renames the module to `github.com/titpetric/phpx` and rewrites every import of it. `PHPX_MODULE` names a different target.

Agent notes:

- `go mod edit -module` is the whole of what the go toolchain carries for this. It writes the go.mod module line and leaves every import of the old path alone, and the go documentation names a text rewrite over the sources as the other half. `gomvpkg` and `gorename` under `golang.org/x/tools` move a package and rename an identifier, predate modules, and sit on no path here.
- The step rewrites one string, `github.com/titpetric/phpscript`. The word phpscript stays in the help text, in the fixtures and in every document. Rewriting those would edit the behaviour the fork inherits.
- `splint fix ./...` runs after the rewrite. The new path sorts into a different import group and the lint gate reads the grouping.
- After the rename, `go install .` writes a binary called `phpx`. A binary takes the last element of the module path.

## Expectations for the additions

- `phpx/testdata/{globals,inheritance,traits,...}` and assisting subfolders with code
- extend with the `/phpx` package, or prefer file additions in existing scopes for go code
- smallest change for the integration, enforce with `-tags phpx` in build steps
- can git clone phpscript into a gitignored subfolder or a go workspace

The agent should choose the methods of implementation based on what causes the least friction from phpscript imports. The phpscript project may adjust structure or add hooks that can make this integration smaller. Any change outside of `/phpx` and `/stdlib/phpx` has to carry phpx attribution, the preferred way is to avoid changes in other places, but if required then keep the changes to `phpx.go` in the relevant packages. It may be hard to integrate phpx. The result embeds phpscript and produces an overlay that implements these runtime compatibilities with the official PHP runtime.

Agent notes:

- A go build tag applies to a file in the module the build compiles. `-tags phpx` and an import of phpscript cannot both hold for one file. The choice is per package: a tagged file in this tree, or an exported hook upstream.
- An imported `runner` cannot see the AST of an included file. `runner/runner_load.go:47` and `runner/precompiler.go:124` call `parser.ParseFile` themselves, and a program the overlay never sees runs phpscript's rules. A transform hook upstream is the smallest change that enables it, and the first `phpx: ...` PR to write.
- The shape that holds each seam to one line: `runner/phpx.go` declares a hook variable the stock path calls, and `runner/phpx_overlay.go` carries `//go:build phpx` and assigns the implementation. An untagged build answers as phpscript does.

## Outputs

The intended outputs of the workflow are:

- A published docker image on `titpetric/phpx`
- Commits and pushes to the main branch of `phpx`

Agent notes:

- `phpx:image` builds `phpx/docker/Dockerfile`, which links `phpscript` to the `phpx` binary inside the image. A compose file or a script written against phpscript runs there unchanged.
- `docker/Dockerfile` upstream reads `./bin/phpscript-linux-amd64` while `go:build` writes `bin/phpx-linux-amd64`, which fails upstream's `atkins build` in this checkout. That file takes a `NAME` build argument and ignores it. Reading that argument is a one-file fix and a second `phpx: ...` PR.
- atkins reads the image name off the checkout directory (`$HOME/.atkins/skills/docker.yml`), and `titpetric/phpx` resolves without configuration.
- `phpx:push` and `phpx:image` sit outside the default pipeline. A run that publishes is a run somebody asked to publish.

## The testdata tree

As the fork gets updated with repeated runs of this workflow, phpx should create a `phpx/testdata` tree, documented in its own README.md. On creation, it should perform research over what php frameworks are common, with this starting list:

- Slim router
- Latte template engine
- Filament
- Laravel

The agent should write test snippets against all the frameworks, and note any blockers or incompatibilities that remain. Anything the reverted decisions leave unstated, but that forms a foundational need of the runtime, stays the same. Adding extensions to the standard library is not the intention here, and `phpx/testdata/report.md` records what those extensions would be.

Agent notes:

- A fixture is a `.phpt` document, or a `<name>_test.php` body with its `<name>_test.txt` output beside it. The expected output comes from running the source through `php` first. A fixture's own folder is its include root.
- `phpscript test ./...` runs a tree. `phpscript test .` matches the fixtures sitting directly in one directory and passes over every folder below it.
- Filament is a Laravel package. Its blockers are Laravel's blockers plus a Livewire front end, and the report should separate the two.

## Contributing upstream

The phpscript contribution policy allows you to open PRs against phpscript, with the label of `phpx: ...`. The ideal solution would have phpx integrate against the phpscript `runner` package and wrap the `phpscript` cmd toolchain. The output binary of `phpscript` and `phpx` have to work the same way in use, hence the need for a maintained fork. This prompt lives in `README.phpx.md` to avoid git conflicts in the fork.

Agent notes:

- The first draft of this file named a `runtime/` package. The package is `runner/`, and the toolchain is `cmd/phpscript/`.
- This file answers to the same prose gate as every other document here. `scripts/vale-gate.sh` reads 39 rules at error level against a baseline of zero, and an edit to this file runs `atkins -f atkins.phpx.yml lint`.
