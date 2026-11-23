name: __destruct runs when the last reference goes, and at the end
description: >
  php calls an instance's __destruct when the last reference to it goes away:
  a variable overwritten or unset, an element removed, a frame that returns,
  and a statement whose value nobody keeps. An inherited destructor runs on
  the child, and what is still alive when the script ends is destroyed then.
  Upstream compiles no destructor at all. Laravel's Route::resource() defers
  its registration to one. The expected section is what `php` printed.
---
<?php
class P {
    public $n;
    function __construct($n) { $this->n = $n; echo " new ", $n; }
    function __destruct() { echo " drop ", $this->n; }
}
class C extends P {}
class Chain extends P {
    function __destruct() { echo " chain"; parent::__destruct(); }
}
function makes() { $local = new P("local"); return new P("returned"); }

echo "1:"; $a = new P("a"); unset($a); echo "\n";
echo "2:"; $b = new P("b"); $b = new P("c"); unset($b); echo "\n";
echo "3:"; $kept = makes(); echo " kept"; unset($kept); echo "\n";
echo "4:"; (new P("temp")); echo "\n";
echo "5:"; $rows = [new P("r0"), new P("r1")]; unset($rows[0]); $rows = null; echo "\n";
echo "6:"; $d = new Chain("d"); unset($d); echo "\n";
echo "7:"; $e = new C("e"); unset($e); echo "\n";
echo "8:"; $o = new stdClass; $o->held = new P("held"); unset($o); echo "\n";
echo "9:"; $orig = new P("orig"); $copy = clone $orig; unset($orig); unset($copy); echo "\n";
echo "10:"; $still = new P("still"); echo " end\n";
---
1: new a drop a
2: new b new c drop b drop c
3: new local new returned drop local kept drop returned
4: new temp drop temp
5: new r0 new r1 drop r0 drop r1
6: new d chain drop d
7: new e drop e
8: new held drop held
9: new orig drop orig drop orig
10: new still end
 drop still