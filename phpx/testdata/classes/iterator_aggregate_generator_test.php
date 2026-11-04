<?php
class Bag implements IteratorAggregate, Countable {
    public function __construct(private array $a) {}
    public function getIterator(): Generator { foreach ($this->a as $k => $v) { yield $k => $v; } }
    public function count(): int { return count($this->a); }
}
$b = new Bag(['x' => 1, 'y' => 2]);
foreach ($b as $k => $v) { echo "$k=$v "; }
echo count($b), json_encode(iterator_to_array($b)), "\n";
echo pathinfo('/a/b/c.txt', PATHINFO_EXTENSION), pathinfo('/a/b/c.txt', PATHINFO_FILENAME), "\n";
echo json_encode(pathinfo('/a/b/c.txt'), JSON_UNESCAPED_SLASHES), json_encode(pathinfo('noext')), json_encode(pathinfo('/x'), JSON_UNESCAPED_SLASHES), pathinfo('/a/b/c.txt', PATHINFO_DIRNAME), pathinfo('c.tar.gz', PATHINFO_BASENAME), "\n";
echo json_encode(pathinfo('dir/')), json_encode(pathinfo('/a/b/'), JSON_UNESCAPED_SLASHES), "\n";
