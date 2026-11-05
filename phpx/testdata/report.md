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
- **`extension_loaded()` answers false for every name.** phpunit 11's entry script requires dom, filter, json, libxml, mbstring, tokenizer and xmlwriter and stops before it loads a test. Past that check, `??` (39 of 1481 files in a phpunit and sebastian install), `fn()` (49), `match` (13) and `enum` (13) are the parse walls. Outside the OOP reversals.
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

## Extension totals

Scope: the fork enables the OOP baseline, and extensions are out of scope
([README.phpx.md](../../README.phpx.md#scope)). The counts below say how much of
each project sits behind an extension, so a failing snippet is read against
them. `php phpx/scripts/extension-tally.php DIR` produces them with php's own
answer for which extension owns a name: calls to a function, and token
occurrences of a class, outside Core, standard, SPL, date, pcre, hash, random,
Reflection, ctype and json, in non-test files, with `vendor/phpunit`, `sebastian`, `phar-io` and `theseer` skipped (the framework columns run it over each shim's `vendor/`; the phpunit column runs `PHPX_TALLY_SKIP=` empty). A class count is occurrences of
the name (`use` lines included), so read it as weight and not as call count.

| Extension  | phpunit 11 + sebastian + phar-io | Ten packages (survey) | Slim 4 | Latte 3 | Laravel 12 | Filament 4 |
|------------|---------------------------------:|----------------------:|-------:|--------:|-----------:|-----------:|
| mbstring   | 15 calls  | 25 calls | 0  | 5 calls   | 397 calls | 438 calls |
| dom        | 326 uses  | 13 uses  | 0  | 0         | 77 | 144 |
| xmlwriter  | 186 uses  | 0        | 0  | 0         | 0  | 0   |
| libxml     | 16 calls  | 0        | 4 calls | 0    | 17 calls | 23 calls |
| tokenizer  | 6         | 1        | 0  | 2 calls   | 0  | 0   |
| pcntl      | 10 calls  | 21 calls | 0  | 2 calls   | 55 calls | 56 calls |
| posix      | 3 calls   | 3 calls  | 0  | 0         | 0  | 0   |
| curl       | 0         | 47 calls | 0  | 0         | 50 calls | 65 calls |
| sockets    | 0         | 6 calls, 16 uses | 0 | 0   | 12 calls, 58 uses | 12 calls, 58 uses |
| intl       | 0         | 8 uses   | 0  | 32 uses   | 80 calls, 264 uses | 84 calls, 274 uses |
| filter     | 2 calls   | 4 calls  | 2 calls | 1 call | 88 calls | 114 calls |
| iconv      | 0         | 0        | 0  | 4 calls   | 31 calls | 0 (not in top 11) |
| PDO        | 0         | 0        | 0  | 0         | 306 uses | 329 uses |
| session    | 0         | 0        | 0  | 0         | 29 calls, 52 uses | 35 calls, 52 uses |
| gd         | 0         | 0        | 0  | 0         | 31 calls | 59 calls |
| xmlreader  | 0         | 0        | 0  | 0         | 0  | 92 uses |
| SimpleXML  | 0         | 0        | 1 call | 0   | 0  | 0   |
| fileinfo   | 0         | 0        | 0  | 2 calls   | 0  | 0   |
| Phar       | 4 uses    | 0        | 0  | 0         | 0  | 0   |

The only extension on the list the fork may bind is PDO, and only the two
frameworks that ship an ORM touch it. Everything else is a wall the fork leaves
standing. `mbstring` and `filter` are the two that are small enough for a
binding of the standard library kind, and the polyfills already ship an
`mb_*` fallback.

## Biggest blockers by project, aside from OOP

Source: [packages/survey.md](packages/survey.md) for the ten packages, and the
probes above for the frameworks. "In scope" is a syntax wall; "out of scope" is
an extension.

| Project | Biggest in-scope wall | Biggest out-of-scope wall |
|---------|-----------------------|---------------------------|
| phpunit 11 | `extension_loaded()` gate in the entry script (skipped by the minimal launcher), then `<=>`, `?->`, `match`, `enum` (13 files each of the last two) | dom (`DOMDocument` in the XML configuration, coverage and logging, 326 uses) and xmlwriter (186 uses): the minimal launcher never reaches them |
| psr/log, psr/container, psr/http-message | none, 16 of 18 files lint clean | none |
| guzzlehttp/psr7 | `??` and `...$spread` in an array literal, 11 failing files | none |
| guzzlehttp/guzzle | `??`, trailing parameter forms, 19 failing files | curl (8 calls), the whole handler layer |
| symfony/console | `??`, conditional top-level statements, `match`, 66 failing files | mbstring (19), pcntl (13), intl (grapheme) and dom (the XML descriptor) |
| symfony/finder | `??`, 6 failing files | none |
| monolog/monolog | `??`, `match`, 40 failing files | curl (39 calls), sockets (22), pcntl (8): the handlers that use them are opt-in |
| nikic/php-parser | `??`, `match`, `enum`, 52 failing files | tokenizer (`PhpToken`, `token_get_all`): its lexer is built on it and cannot run without it |
| doctrine/inflector | `??`, 11 failing files | none |
| Slim | `??` in `AppFactory::create()` | libxml and SimpleXML in the body parser, on a path a plain route never takes |
| Latte | `fn() => ...` (now parsed), then the compiler's `match` and `enum` | intl formatters (32 uses), iconv, token_get_all in the compiler |
| Laravel | `PHP_VERSION` 8.4.0 below the platform check, `??` in `symfony/polyfill-mbstring` | mbstring (397 calls: the polyfill answers), PDO (306 uses, the one in-reach extension), intl, session, curl |
| Filament | Laravel's walls, then Livewire | the same set as Laravel plus xmlreader (92) and gd (59) |

The widest non-OOP wall across the survey is `??` (99 failing files), now
cleared, followed by `match` and `enum`. Of the extensions, mbstring is the one
every project reaches and the only one cheap enough to consider; dom and
xmlwriter weigh most in phpunit and are the clearest reason it needs the
minimal launcher.

## The standard library

No snippet reached for a function phpscript lacks. The polyfills ship their own
`mb_*` fallback, and the probes stopped at syntax and include semantics before
a binding came into play. A snippet that does reach for an extension adds a
line here, and the binding waits for the issue the scope section requires.
