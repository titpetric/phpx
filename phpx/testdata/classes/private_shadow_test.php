<?php
class ShadowP {
    private array $a = ['p'];
    private string $s = 'ps';
    function read() { return $this->a; }
    function spread() { return [...$this->a]; }
    function write($v) { $this->a = $v; }
    function push($v) { $this->a[] = $v; }
    function vars() { return get_object_vars($this); }
    function has() { return isset($this->a); }
    function drop() { unset($this->a); }
    function loop() { $out = []; foreach ($this as $k => $v) { $out[$k] = $v; } return $out; }
}
class ShadowC extends ShadowP {
    private array $a = ['c'];
    function readC() { return $this->a; }
    function writeC($v) { $this->a = $v; }
    function pushC($v) { $this->a[] = $v; }
    function varsC() { return get_object_vars($this); }
    function hasC() { return isset($this->a); }
    function dropC() { unset($this->a); }
    function loopC() { $out = []; foreach ($this as $k => $v) { $out[$k] = $v; } return $out; }
}
class ShadowG extends ShadowC {
    private array $a = ['g'];
    private string $g = 'g';
    function readG() { return $this->a; }
    function varsG() { return get_object_vars($this); }
}
$c = new ShadowC;
var_dump($c->read(), $c->readC(), $c->spread());
$c->write(['p2']);
$c->writeC(['c2']);
$c->push('x');
var_dump($c->read(), $c->spread(), $c->readC());
var_dump($c->vars(), $c->varsC());
var_dump($c->has(), $c->hasC());
var_dump($c->loop(), $c->loopC());
echo json_encode($c), "\n";
$clone = clone $c;
$c->write(['after']);
var_dump($clone->read(), $clone->readC());
$g = new ShadowG;
var_dump($g->read(), $g->readC(), $g->readG(), $g->varsG());
$c->drop();
var_dump($c->has(), $c->hasC(), $c->vars(), $c->loop());
$c->dropC();
var_dump($c->hasC(), $c->varsC());
