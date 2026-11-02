# Laravel

`laravel/framework` 12.70.2. `index.php` loads the composer autoloader and
checks for `Illuminate\Foundation\Application`. `php index.php` prints
`bool(true)`.

    composer install
    php index.php

phpx meets two walls, and both sit outside the OOP reversals:

- `vendor/composer/platform_check.php` requires `PHP_VERSION_ID >= 80401`.
  phpscript reports `8.4.0`, which is `80400`, and the generated check fails
  before any framework file loads.
- with the check bypassed, `symfony/polyfill-mbstring/bootstrap80.php` uses
  `??` in `mb_convert_encoding()`, and
  [the operator reference](../../../../docs/reference/operators/README.md)
  lists the operator as not implemented.

[The report](../../report.md) carries the entry and the scope decision.
