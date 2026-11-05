name: identifiers may hold bytes above 0x7F
description: >
  php identifiers are [a-zA-Z_\x80-\xff][a-zA-Z0-9_\x80-\xff]*, so latte's
  generated `$ʟ_args` and a function named with non-ASCII letters both parse.
---
<?php
function ʟ_greet(string $ʟ_name): string { return "hi " . $ʟ_name; }
$ʟ_args = ['a' => 1];
echo ʟ_greet('x'), " ", $ʟ_args['a'], "\n";
---
hi x 1
