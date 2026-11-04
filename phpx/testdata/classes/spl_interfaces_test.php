<?php
class Col implements Iterator {
    private $i = 0;
    public function __construct(private array $a) {}
    public function current(): mixed { return $this->a[$this->i]; }
    public function key(): mixed { return $this->i; }
    public function next(): void { $this->i++; }
    public function rewind(): void { $this->i = 0; }
    public function valid(): bool { return isset($this->a[$this->i]); }
}
class Agg implements IteratorAggregate, Countable, ArrayAccess {
    public function __construct(private array $a) {}
    public function getIterator(): Iterator { return new Col($this->a); }
    public function count(): int { return count($this->a); }
    public function offsetExists(mixed $o): bool { return isset($this->a[$o]); }
    public function offsetGet(mixed $o): mixed { return $this->a[$o]; }
    public function offsetSet(mixed $o, mixed $v): void { if ($o === null) { $this->a[] = $v; } else { $this->a[$o] = $v; } }
    public function offsetUnset(mixed $o): void { unset($this->a[$o]); }
}
foreach (new Col(['a', 'b']) as $k => $v) { echo "$k=$v "; }
echo "\n";
$g = new Agg(['x', 'y', 'z']);
foreach ($g as $k => $v) { echo "$k=$v "; }
echo count($g), isset($g[1]) ? 'Y' : 'N', $g[2], "\n";
$g[] = 'w'; unset($g[0]);
echo count($g), "\n";
echo implode(',', iterator_to_array($g)), "\n";
