name: global binds the name to the global frame
phpx: true
description: >
  phpx implements the global statement: the name in a function reads and
  writes the variable the file body holds, which is what php prints. Upstream
  phpscript answers "global is a no-op" here and opts the php runner out of
  the fixture. The bytecode engine declines a program that uses the statement
  and the whole program runs on the interpreter, so both engine columns print
  the interpreter's answer.
---
<?php

$x = 1;
function f() {
	global $x;
	echo isset($x) ? "sees x" : "global is a no-op", "\n";
	$x = 2;
}
f();
echo "after f: ", $x, "\n";
---
sees x
after f: 2
