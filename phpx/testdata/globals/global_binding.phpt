name: global binds a name to the global frame
description: >
  Every behaviour php gives the statement, in one fixture: a read sees the
  global, a write reaches it, several names bind at once, a name the global
  frame does not hold arrives as null and a write to it creates the global,
  unset() drops the binding and leaves the global, GLOBAL is the same keyword,
  and a local value assigned before the statement falls out of reach. The
  expected section is what `php` printed. The bytecode engine declines a
  program that uses the statement, so its column skips and the whole program
  runs on the interpreter.
---
<?php

$counter = 1;
$name = "top";

function reads()
{
	global $counter;
	echo "reads: ", $counter, "\n";
}

function writes()
{
	global $counter;
	$counter = $counter + 10;
}

function several()
{
	global $counter, $name;
	echo "several: ", $counter, " ", $name, "\n";
}

function creates()
{
	global $fresh;
	echo "creates: ", var_export($fresh, true), "\n";
	$fresh = "made in a function";
}

function drops()
{
	global $counter;
	unset($counter);
	echo "drops: ", isset($counter) ? "still bound" : "unbound here", "\n";
}

function upper()
{
	GLOBAL $name;
	echo "upper: ", $name, "\n";
}

reads();
writes();
reads();
several();
creates();
echo "after creates: ", $fresh, "\n";
drops();
echo "after drops: ", $counter, "\n";
upper();

$shadow = "global value";
function shadows()
{
	$shadow = "local value";
	global $shadow;
	echo "shadows: ", $shadow, "\n";
}
shadows();
echo "after shadows: ", $shadow, "\n";
---
reads: 1
reads: 11
several: 11 top
creates: NULL
after creates: made in a function
drops: unbound here
after drops: 11
upper: top
shadows: global value
after shadows: global value
