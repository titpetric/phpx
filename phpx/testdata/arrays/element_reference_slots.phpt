name: an array element can be the slot a reference names
description: >
  A reference stored in an array keeps the two names one: assigning a value to
  the element writes through it, a by-reference loop binds the loop variable to
  the element and leaves it bound after the loop, array_pop through a stored
  reference removes the value everywhere, rebinding the element replaces the
  old reference, and a nested element passed to a by-reference parameter is the
  same slot. Expected output is php 8.4's. The bytecode engine declines a
  program with a reference outside `foreach`, so its column skips.
runner:
  flatstack: false
---
<?php
class Refs {
    public array $slots = [];
    public function add(array &$value) { $this->slots[] = &$value; }
    public function drop() {
        foreach ($this->slots as &$slot) {
            if (count($slot) > 0) { array_pop($slot); }
        }
    }
}
$a = [1, 2, 3];
$c = new Refs;
$c->add($a);
$c->drop();
echo json_encode($a), "\n";

$x = 1;
$arr = [];
$arr[] = &$x;
$arr[0] = 5;
echo $x, "\n";
$x = 6;
echo $arr[0], "\n";

$out = [1, 2];
foreach ($out as &$v) { $v *= 10; }
$v = 99;
echo json_encode($out), "\n";
unset($v);

$p = 'a'; $q = 'b';
$holder = [];
$holder[0] = &$p;
$holder[0] = &$q;
$holder[0] = 'z';
echo $p, $q, "\n";

function bump(array &$inner) { $inner[] = 'bump'; }
$nested = ['k' => ['a']];
bump($nested['k']);
echo json_encode($nested), "\n";
---
[1,2]
5
6
[10,99]
az
{"k":["a","bump"]}
