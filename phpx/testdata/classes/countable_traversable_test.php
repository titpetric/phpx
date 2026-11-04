<?php
class C implements Countable { function count(): int { return 3; } }
class I implements IteratorAggregate { function getIterator(): Iterator { return new ArrayIterator([1]); } }
class J implements Iterator { function current(): mixed { return 1; } function key(): mixed { return 0; } function next(): void {} function rewind(): void {} function valid(): bool { return false; } }
var_dump(is_countable(new C), is_countable([]), is_countable(new I), is_countable('s'), is_countable(new ArrayObject([])));
var_dump(new C instanceof Traversable, new I instanceof Traversable, new J instanceof Traversable, new ArrayIterator([]) instanceof Traversable, new ArrayObject([]) instanceof Traversable, new stdClass instanceof Traversable);
