name: a runtime error reports the file and line it surfaced at
description: >
  getLine() on an error the runtime raised (DivisionByZeroError) answers the
  line of the failing statement. The bytecode engine does not record the site
  of an error, so the fixture opts it out.
runner:
  flatstack: false
---
<?php
function f() {
    return intdiv(1, 0);
}
try {
    f();
} catch (DivisionByZeroError $e) {
    echo $e->getFile() === __FILE__ ? "here" : "elsewhere", ":", $e->getLine(), "\n";
}
try {
    $x = 1 % 0;
} catch (DivisionByZeroError $e) {
    echo $e->getLine(), "\n";
}
---
here:3
11
