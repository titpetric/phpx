<?php
class Addr { public $city = 'Ljubljana'; public ?Addr $next = null; function city() { return $this->city; } }
class User { public ?Addr $addr = null; public $n = 0; function addr() { return $this->addr; } function bump() { $this->n++; return $this; } }

function show($v) { var_dump($v); }

$u = new User;
$n = null;
show($n?->addr);
show($n?->addr());
show($u?->addr);
show($u->addr?->city);
$u->addr = new Addr;
show($u->addr?->city);
show($u?->addr()?->city());
show($u->addr?->next?->city);

// short circuit: the rest of the chain is skipped, arguments are not evaluated
$calls = 0;
function side() { global $calls; $calls++; return 1; }
show($n?->foo(side())->bar->baz());
show($calls);

// the base evaluates once
$x = new User;
show($x->bump()?->n);
show($x->n);

// combined with ??
show($n?->addr ?? 'none');
show($u->addr?->city ?? 'none');

// in a condition and a string
if ($n?->addr) { echo "yes\n"; } else { echo "no\n"; }
echo "city: " . $u->addr?->city . "\n";

// a loop
foreach ([$u, null, $x] as $i => $o) {
    echo $i, ':', $o?->n, "|\n";
}
