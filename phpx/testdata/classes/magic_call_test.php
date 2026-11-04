<?php
class Proxy {
    private array $log = [];
    public function __call(string $name, array $args) { $this->log[] = $name . '(' . implode(',', $args) . ')'; return strtoupper($name); }
    public function real() { return 'real'; }
    public function log(): array { return $this->log; }
}
$p = new Proxy;
echo $p->real(), $p->hasFoo(1, 2), $p->other(), "\n";
echo json_encode($p->log()), "\n";
