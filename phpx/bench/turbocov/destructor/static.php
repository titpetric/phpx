<?php
class Reg {
    public static $items = [];
}
class Item {
    public $id;
    function __construct($id) { $this->id = $id; }
    function __destruct() {}
}
$n = (int)($argv[1] ?? 100000);
for ($i = 0; $i < $n; $i++) {
    Reg::$items[] = new Item($i);
    $tmp = new Item($i);   // dropped every iteration, reachable from static
    $tmp = null;
}
echo "done ", count(Reg::$items), "\n";
