<?php
class Node {
    public $id; public $children = [];
    function __construct($id) { $this->id = $id; }
    function __destruct() {}
}
class Holder { public $items = []; }
$n = (int)($argv[1] ?? 2000);
$h = new Holder();
for ($i = 0; $i < $n; $i++) {
    $root = new Node($i);
    for ($j = 0; $j < 20; $j++) { $root->children[] = new Node($i * 20 + $j); }
    $h->items[] = $root;
    $root = null;
}
echo "done ", count($h->items), "\n";
