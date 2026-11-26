<?php
class WithDtor {
    public $id;
    public $data;
    function __construct($id) { $this->id = $id; $this->data = [$id, $id + 1]; }
    function __destruct() {}
}
$n = (int)($argv[1] ?? 100000);
for ($i = 0; $i < $n; $i++) {
    $o = new WithDtor($i);
    $o = null;
}
echo "done $n\n";
