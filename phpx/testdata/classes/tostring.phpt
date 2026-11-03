name: __toString supplies the string form of an object
description: >
  php converts an object to a string through its __toString method, in echo,
  concatenation, strval, sprintf, implode and str_replace. Upstream treats an
  object as having no string form. The method may be inherited or come from a
  trait. The expected section is what `php` printed.
---
<?php
class P { public function __toString(): string { return "p" . $this->n(); } function n() { return static::class; } }
class C extends P {}
trait T { function __toString() { return "trait"; } }
class U { use T; }
class N {}
echo new C, "\n", "x" . new C . "\n", strval(new U), "\n";
$s = sprintf("%s|%s", new C, new U); echo $s, "\n";
echo implode(",", [new C, new U]), "\n";
echo str_replace("p", "Q", new C), "\n";
echo (string) new C, "\n";
var_dump((string) new U === "trait", new C == "pC");
---
pC
xpC
trait
pC|trait
pC,trait
QC
pC
bool(true)
bool(true)
