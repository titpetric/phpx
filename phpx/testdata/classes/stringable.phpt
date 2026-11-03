name: a class with __toString is Stringable, one without has no string form
description: >
  php adds the Stringable interface to every class that declares __toString,
  and a subclass or a trait user inherits it. An object without the method is
  not Stringable. The expected section is what `php` printed.
---
<?php
class A { function __toString() { return "a"; } }
class B extends A {}
class C {}
trait T { function __toString() { return "t"; } }
class D { use T; }
function show(Stringable|string $s): string { return (string) $s; }
var_dump(new A instanceof Stringable, new B instanceof Stringable);
var_dump(new C instanceof Stringable, new D instanceof Stringable);
echo show(new B), show("x"), "\n";
try { echo new C; } catch (Error $e) { echo get_class($e), ": ", $e->getMessage(), "\n"; }
try { $x = "v" . new C; } catch (Error $e) { echo $e->getMessage(), "\n"; }
echo "done\n";
---
bool(true)
bool(true)
bool(false)
bool(true)
ax
Error: Object of class C could not be converted to string
Object of class C could not be converted to string
done
