# report

The blockers the framework snippets met, one entry per framework, and the
standard library each reached for. A framework that cannot run yet ticks no box
in [ACCEPTANCE.md](../../ACCEPTANCE.md), and this file names the first wall it
met and the walls behind it.

The standing prompt settles each gap: work inside the five reversals is the
fork's, and anything else waits for an issue. Every entry below states that
decision.

Each snippet runs through `php` first, from
[frameworks/](frameworks/), after `composer install` in its folder.

## Shared walls

These sit under more than one framework.

- **(Cleared by the fork.) A namespaced file may not hold a top-level statement.** composer's
  autoload files include `nikic/fast-route/src/functions.php`, which wraps its
  function declarations in `if (!function_exists(...))`. phpscript reads a
  namespaced file for the symbols it declares and refuses a statement there.
  [docs/README.md](../../docs/README.md) records the divergence. Outside the
  OOP reversals.
- **(Cleared by the fork.) A conditional function declaration does not run.** In
  `if (!function_exists('f')) { function f() {} }` in a file with no
  namespace, the parser reads the declaration and the file never defines `f`.
  `php` prints `called` for a call after the `if`; phpx raises
  `call to undefined function f()`. The same guard wraps the helpers in
  `symfony/polyfill-*` and `ralouphie/getallheaders`. Outside the OOP
  reversals.
- **`??` and `??=` are a parse error.** The operator reference lists both as
  not implemented. Outside the OOP reversals.
- **`fn($x) => ...` is a parse error.** phpscript treats `fn` as an alias for
  a block-bodied `function`, recorded as a documented PHP incompatibility in
  [the extensions reference](../../docs/reference/extensions/README.md).
  Outside the OOP reversals.
- **`PHP_VERSION` reads `8.4.0`.** A vendor tree whose `require.php` names a
  later patch, `>=8.4.1` among them, fails composer's generated
  `platform_check.php` before any package file loads. Outside the OOP
  reversals.

## Slim router

Snippet: [frameworks/slim/](frameworks/slim/), `slim/slim` 4.15.3.

`php index.php` prints `Hello, world`. The program builds the app through
`AppFactory`, declares `GET /hello/{name}`, and handles a `ServerRequest`
without the built-in server.

- First wall: the `nikic/fast-route` autoload file from the shared list.
- Next wall: `Slim\Factory\AppFactory::create()` uses `??`.

## Latte template engine

Snippet: [frameworks/latte/](frameworks/latte/), `latte/latte` 3.1.7.

`php index.php` prints `Hello world!`. The program renders a template through
`Latte\Engine` with a `StringLoader`.

- First wall: `Latte\Engine` declares `private \stdClass $providers;`, and the
  class body parser refused a property type that opens with `\`. The fork adds
  the case to `parser/parser.go`, and
  [syntax/typed_property.phpt](syntax/typed_property.phpt) settles the syntax
  against php. Inside the OOP work: a typed property is class syntax.
- Next wall: `Engine::renderToString()` passes `fn() => ...` as a capture,
  from the shared list.

## Laravel

Snippet: [frameworks/laravel/](frameworks/laravel/), `laravel/framework`
12.70.2.

`php index.php` prints `bool(true)`. The program loads the composer autoloader
and checks for `Illuminate\Foundation\Application`.

- First wall: the generated composer platform check requires
  `PHP_VERSION_ID >= 80401`, from the shared list.
- Next wall: `symfony/polyfill-mbstring/bootstrap80.php` uses `??` in
  `mb_convert_encoding()`.

## Filament

Snippet: [frameworks/filament/](frameworks/filament/), `filament/filament`
4.16.0.

Filament has no standalone entry: `Panel::make()->id()` calls
`Container::runningInConsole()`, and `php` fails that outside a booted Laravel
application. The snippet is the `PanelProvider` a Laravel application
registers, and it cannot run alone.

- First wall: the Laravel walls above, met from the same autoloader.
- Livewire check: not reached. Filament renders through Livewire, and a probe
  needs the full application, which no point past the first wall reaches.

## The standard library

No snippet reached for a function phpscript lacks. The polyfills ship their own
`mb_*` fallback, and the probes stopped at syntax and include semantics before
a binding came into play. A snippet that does reach for an extension adds a
line here, and the binding waits for the issue the scope section requires.
