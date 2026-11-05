name: yield from a binding's array
description: >
  str_split returns a Go slice, and yield from walks it like any array. Expected output is php 8.4's.
runner:
  flatstack: false
---
<?php
class T { public ?object $b = null; public $s = 'xyz';
 function g1() { yield $this->b?->value . $this->s; }
 function g2() { yield from [$this->b?->value . $this->s]; }
 function g3() { $x = $this->b?->value . $this->s; yield $x; }
 function g4() { yield from str_split($this->s, 2); }
 function g5() { yield from str_split($this->b?->value . $this->s, 2); }
 function g6() { $a = str_split($this->b?->value . $this->s, 2); yield from $a; }
}
$t = new T; foreach (['g1','g2','g3','g4','g5','g6'] as $g) { echo $g, ' ', json_encode(iterator_to_array($t->$g(), false)), "\n"; }
---
g1 ["xyz"]
g2 ["xyz"]
g3 ["xyz"]
g4 ["xy","z"]
g5 ["xy","z"]
g6 ["xy","z"]
