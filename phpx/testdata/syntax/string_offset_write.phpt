name: a string offset takes a write
description: >
  `$s[0] = 'J'` rebuilds the string and stores it back. The bytecode engine
  writes an index without the place the string came from, so the fixture opts
  it out; the interpreter and php agree.
runner:
  flatstack: false
---
<?php
// The class below keeps the program on the interpreter: the bytecode engine
// writes an index without the place the string came from.
class W { public string $name = 'abc'; function up() { $this->name[0] = strtoupper($this->name[0]); return $this->name; } }
$s = 'hello';
$s[0] = 'J';
echo $s, "\n";
$s[1] = 'EXTRA';
echo $s, "\n";
$s[-1] = '!';
echo $s, "\n";
$s[7] = 'x';
var_dump($s);
echo (new W)->up(), "\n";
$parts = ['word'];
$parts[0][0] = 'W';
echo $parts[0], "\n";
?>
---
Jello
JEllo
JEll!
string(8) "JEll!  x"
Abc
Word
