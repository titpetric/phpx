<?php
class Evens extends FilterIterator {
    public function accept(): bool { return $this->current() % 2 === 0; }
}
echo implode(',', iterator_to_array(new Evens(new ArrayIterator([1, 2, 3, 4, 5, 6])), false)), "\n";
echo implode(',', iterator_to_array(new CallbackFilterIterator(new ArrayIterator(['a', 'bb', 'ccc']), fn($v) => strlen($v) > 1), false)), "\n";
echo implode(',', iterator_to_array(new LimitIterator(new ArrayIterator([10, 20, 30, 40]), 1, 2), false)), "\n";
$nested = new RecursiveArrayIterator([1, [2, 3, [4]], 5]);
echo implode(',', iterator_to_array(new RecursiveIteratorIterator($nested), false)), "\n";
echo implode(',', iterator_to_array(new RecursiveIteratorIterator($nested, RecursiveIteratorIterator::SELF_FIRST), false) ? array_map(fn($v) => is_array($v) ? 'A' : $v, iterator_to_array(new RecursiveIteratorIterator($nested, RecursiveIteratorIterator::SELF_FIRST), false)) : []), "\n";
foreach (new EmptyIterator as $never) { echo "no\n"; }
class Name extends RecursiveFilterIterator {
    public function accept(): bool { return $this->hasChildren() || $this->current() !== 3; }
}
echo implode(',', iterator_to_array(new RecursiveIteratorIterator(new Name($nested)), false)), "\n";
$it = new IteratorIterator(new ArrayObject(['x' => 1, 'y' => 2]));
foreach ($it as $k => $v) { echo $k, $v; }
echo "\n";
