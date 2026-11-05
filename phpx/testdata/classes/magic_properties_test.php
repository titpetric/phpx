<?php
class Lazy {
    private $built;
    private array $data = ['a' => 1];
    public function __construct() { unset($this->built); }
    public function __get($name) { echo "get $name\n"; return $this->built = "built-$name"; }
    public function __set($name, $value) { echo "set $name\n"; $this->data[$name] = $value; }
    public function __isset($name) { echo "isset $name\n"; return isset($this->data[$name]); }
    public function __unset($name) { echo "unset $name\n"; unset($this->data[$name]); }
    public static function __callStatic($name, $args) { return "static $name(" . implode(',', $args) . ')'; }
    public function __call($name, $args) { return "call $name"; }
}
$l = new Lazy;
echo $l->built, "\n";
echo $l->built, "\n";
$l->extra = 5;
var_dump(isset($l->a), isset($l->nope), empty($l->a));
unset($l->extra);
echo Lazy::make(1, 2), "\n", $l->go(), "\n";
class Plain { public $x = 1; private $hidden = 2; public function __get($n) { return "magic $n"; } }
$p = new Plain; echo $p->x, $p->hidden, $p->missing, "\n";
class NoMagic { private $v = 1; }
$n = new NoMagic; var_dump(isset($n->v), isset($n->w));
