<?php
function boom() {
    throw new RuntimeException("x");
}
class MyEx extends Exception {
    public function __construct() {}
}
try { boom(); } catch (Exception $e) {
    echo basename($e->getFile()), ":", $e->getLine(), "\n";
}
$m = new MyEx();
echo basename($m->getFile()), ":", $m->getLine(), "\n";
