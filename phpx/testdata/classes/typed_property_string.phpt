name: a Stringable written to a string property is stored as its string
description: >
  php converts an object with __toString when it is written to a property
  declared string, and keeps the object for a nullable string and an untyped
  property alike. Expected output is php 8.4's.
---
<?php
class S { public function __toString(): string { return 's'; } }
class P { public string $name = ''; public ?string $maybe = null; public $raw; }
$p = new P;
$p->name = new S;
$p->maybe = new S;
$p->raw = new S;
echo gettype($p->name), ' ', $p->name, "\n";
echo gettype($p->maybe), ' ', $p->maybe, "\n";
echo gettype($p->raw), "\n";
?>
---
string s
string s
object
