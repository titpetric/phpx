<?php
class Suite {
    public function check($t) { return $t instanceof self ? 'self' : ($t instanceof static ? 'static' : 'no'); }
    public function check2(Suite|Other $t) { return $t instanceof self; }
}
final class Sub extends Suite {}
class Other {}
$s = new Suite;
echo $s->check(new Sub), $s->check(new Suite), $s->check(new Other), (new Sub)->check(new Sub), "\n";
var_dump($s->check2(new Sub));
