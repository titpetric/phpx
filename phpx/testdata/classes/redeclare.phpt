name: a class name is declared once
description: >
  php refuses a second declaration of one class name at compile time, so the
  statement before the declarations never runs. The expected error names the
  message php printed up to the file and line, which the fork does not carry.
  The bytecode engine collects the classes itself, so it follows the
  interpreter only where the program is declined.
error: 'Cannot redeclare class A'
runner:
  php: false
---
<?php

class A {}
class A {}

echo "unreachable\n";
---
Internal Server Error
