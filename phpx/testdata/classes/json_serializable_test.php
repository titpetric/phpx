<?php
class Pt implements JsonSerializable {
    public function __construct(public int $x = 1, private int $secret = 9) {}
    public function jsonSerialize(): mixed { return ['x' => $this->x, 'sum' => $this->x + $this->secret]; }
}
class Plain { public $a = 1; protected $b = 2; private $c = 3; public $list = [1, 2]; public ?Plain $child = null; }
class Wrapper { public function __construct(public Pt $pt, public array $items = []) {} }
echo json_encode(new Pt(4)), "\n";
echo json_encode([new Pt(1), new Pt(2)]), "\n";
echo json_encode(new Plain), "\n";
$p = new Plain; $p->child = new Plain;
echo json_encode($p), "\n";
echo json_encode(new Wrapper(new Pt(5), ['k' => new Pt(6)])), "\n";
echo json_encode(['n' => null, 'e' => new stdClass, 'o' => (object) ['z' => 1]]), "\n";
echo json_encode(new ArrayObject([1, 2])), "\n";
echo json_encode(['html' => '<b>&</b>', 'q' => "it's"], JSON_UNESCAPED_SLASHES), "\n";
$m = fopen('php://memory', 'w+b'); fwrite($m, "abcdef"); echo stream_get_contents($m, -1, 0), '|', stream_get_contents($m, 2, 1), "\n";
