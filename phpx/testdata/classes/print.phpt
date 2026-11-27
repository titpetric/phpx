name: print renders like echo
description: >
  print is echo that answers 1, so it renders a value through the same
  conversion: a stringable object prints __toString, and an object without one
  raises php's Error instead of printing nothing.
---
<?php
class Shown { function __toString() { return "shown"; } }
print new Shown();
class Plain {}
try {
	print new Plain();
} catch (Error $e) {
	echo "\n", $e->getMessage(), "\n";
}
---
shown
Object of class Plain could not be converted to string
