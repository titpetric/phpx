# Latte template engine

`latte/latte` 3.1.7. `index.php` renders `Hello {$name}!` through
`Latte\Engine` with a `StringLoader`. `php index.php` prints `Hello world!`.

    composer install
    php index.php

phpx meets two walls:

- `Latte\Engine` declares `private \stdClass $providers;`. The class body
  parser had no case for a leading backslash on a property type. This run adds
  the case to `parser/parser.go`, and
  [typed_property.phpt](../../syntax/typed_property.phpt) settles it against
  php.
- past that file, `Engine::renderToString()` passes `fn() => ...` as a
  capture. phpscript's `fn` is an alias for a block-bodied `function`, a
  documented PHP incompatibility in
  [the extensions reference](../../../../docs/reference/extensions/README.md).

[The report](../../report.md) carries the entry and the scope decision.
