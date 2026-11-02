name: PHP_VERSION reaches a patch level a vendor tree can require
description: >
  composer's platform_check.php compares PHP_VERSION_ID with the lowest php a
  vendor tree requires, and laravel/framework asks for 8.4.1. The fork reported
  upstream's 8.4.0. The expected section is what `php` printed.
---
<?php
var_dump(PHP_VERSION_ID >= 80401);
var_dump(PHP_VERSION === phpversion());
var_dump(PHP_VERSION_ID === PHP_MAJOR_VERSION * 10000 + PHP_MINOR_VERSION * 100 + PHP_RELEASE_VERSION);
var_dump(strpos(PHP_VERSION, PHP_MAJOR_VERSION . '.' . PHP_MINOR_VERSION . '.' . PHP_RELEASE_VERSION) === 0);
---
bool(true)
bool(true)
bool(true)
bool(true)
