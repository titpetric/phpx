# Slim router

`slim/slim` 4.15.3 with `slim/psr7` 1.8.0. `index.php` builds the app through
`AppFactory`, declares `GET /hello/{name}`, and handles a request without the
built-in server. `php index.php` prints `Hello, world`.

    composer install
    php index.php

phpx meets two walls, and both sit outside the OOP reversals:

- composer's autoload files include `nikic/fast-route/src/functions.php`, a
  namespaced file whose top level holds an `if (!function_exists(...))`.
  phpscript reads a namespaced file for the symbols it declares and refuses a
  statement there, a divergence [docs/README.md](../../../../docs/README.md)
  records.
- past that file, `Slim\Factory\AppFactory::create()` uses `??`, which
  [the operator reference](../../../../docs/reference/operators/README.md)
  lists as not implemented.

[The report](../../report.md) carries the entry and the scope decision.
