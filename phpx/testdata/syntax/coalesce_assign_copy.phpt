name: a coalescing assignment copies the array it stores
description: >
  `$x ??= $arr` reads $arr and stores it, and php's array value copies at the
  assignment, so a write through $x does not reach $arr. Pest's ArgvInput reads
  `$argv ??= $_SERVER['argv']` and shifts the script name off, which emptied the
  superglobal when one handle served both. Expected output is php 8.4's. Limit:
  the bytecode engine stores the handle and not a copy, so this runs on the
  interpreter alone.
runner:
  flatstack: false
---
<?php
// `??=` assigns a value, and php's array value copies: a write through the
// target does not reach the variable it was read from.
$main = ['a', 'b', 'c'];
$copy = null;
$copy ??= $main;
array_shift($copy);
var_dump($main);

// The same through an array element and a property.
$bag = ['list' => null];
$source = ['x', 'y'];
$bag['list'] ??= $source;
$bag['list'][] = 'z';
var_dump($source);

class Holder
{
    public $items;

    public function fill(array $from): void
    {
        $this->items ??= $from;
        $this->items[] = 'end';
    }
}

$h = new Holder();
$h->fill(['one']);
var_dump($h->items);
---
array(3) {
  [0]=>
  string(1) "a"
  [1]=>
  string(1) "b"
  [2]=>
  string(1) "c"
}
array(2) {
  [0]=>
  string(1) "x"
  [1]=>
  string(1) "y"
}
array(2) {
  [0]=>
  string(3) "one"
  [1]=>
  string(3) "end"
}
