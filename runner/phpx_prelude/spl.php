<?php
interface SplObserver { public function update(SplSubject $subject): void; }
interface SplSubject {
    public function attach(SplObserver $observer): void;
    public function detach(SplObserver $observer): void;
    public function notify(): void;
}

class ArrayIterator implements Iterator, Countable, ArrayAccess {
    protected $storage = [];
    private $keys = [];
    private $pos = 0;
    public function __construct($array = []) { $this->storage = $array; $this->keys = array_keys($array); }
    public function current(): mixed { return $this->storage[$this->keys[$this->pos]]; }
    public function key(): mixed { return $this->keys[$this->pos] ?? null; }
    public function next(): void { $this->pos++; }
    public function rewind(): void { $this->keys = array_keys($this->storage); $this->pos = 0; }
    public function valid(): bool { return $this->pos < count($this->keys); }
    public function count(): int { return count($this->storage); }
    public function offsetExists(mixed $o): bool { return isset($this->storage[$o]); }
    public function offsetGet(mixed $o): mixed { return $this->storage[$o]; }
    public function offsetSet(mixed $o, mixed $v): void { if ($o === null) { $this->storage[] = $v; } else { $this->storage[$o] = $v; } }
    public function offsetUnset(mixed $o): void { unset($this->storage[$o]); }
    public function getArrayCopy() { return $this->storage; }
}

class ArrayObject implements IteratorAggregate, Countable, ArrayAccess {
    private $storage = [];
    public function __construct($array = []) { $this->storage = $array; }
    public function getIterator(): Iterator { return new ArrayIterator($this->storage); }
    public function count(): int { return count($this->storage); }
    public function offsetExists(mixed $o): bool { return isset($this->storage[$o]); }
    public function offsetGet(mixed $o): mixed { return $this->storage[$o]; }
    public function offsetSet(mixed $o, mixed $v): void { if ($o === null) { $this->storage[] = $v; } else { $this->storage[$o] = $v; } }
    public function offsetUnset(mixed $o): void { unset($this->storage[$o]); }
    public function append($v) { $this->storage[] = $v; }
    public function getArrayCopy() { return $this->storage; }
    public function exchangeArray($array) { $old = $this->storage; $this->storage = $array; return $old; }
    // json_encode writes the storage as an object, as php does.
    public function jsonSerialize() { return (object) $this->storage; }
}

class SplObjectStorage implements Countable, Iterator, ArrayAccess {
    private $objects = [];
    private $data = [];
    private $pos = 0;
    private function id($o) { return spl_object_id($o); }
    public function attach($object, $info = null) { $i = $this->id($object); $this->objects[$i] = $object; $this->data[$i] = $info; }
    public function detach($object) { $i = $this->id($object); unset($this->objects[$i], $this->data[$i]); }
    public function contains($object) { return isset($this->objects[$this->id($object)]); }
    public function addAll($storage) { foreach ($storage as $i => $o) { $this->attach($o, $storage[$o]); } return count($this->objects); }
    public function count(): int { return count($this->objects); }
    public function offsetExists($object): bool { return $this->contains($object); }
    public function offsetGet($object): mixed {
        $i = $this->id($object);
        if (!isset($this->objects[$i])) { throw new UnexpectedValueException('Object not found'); }
        return $this->data[$i];
    }
    public function offsetSet($object, $info = null): void { $this->attach($object, $info); }
    public function offsetUnset($object): void { $this->detach($object); }
    public function getInfo() { $k = array_keys($this->objects); return $this->data[$k[$this->pos]] ?? null; }
    public function setInfo($info) { $k = array_keys($this->objects); if (isset($k[$this->pos])) { $this->data[$k[$this->pos]] = $info; } }
    public function current(): mixed { $k = array_keys($this->objects); return $this->objects[$k[$this->pos]]; }
    public function key(): mixed { return $this->pos; }
    public function next(): void { $this->pos++; }
    public function rewind(): void { $this->pos = 0; }
    public function valid(): bool { return $this->pos < count($this->objects); }
}

class SplDoublyLinkedList implements Iterator, Countable, ArrayAccess {
    protected $items = [];
    protected $pos = 0;
    protected $lifo = false;
    public function push($v) { $this->items[] = $v; }
    public function pop() {
        if (!$this->items) { throw new RuntimeException("Can't pop from an empty datastructure"); }
        return array_pop($this->items);
    }
    public function shift() {
        if (!$this->items) { throw new RuntimeException("Can't shift from an empty datastructure"); }
        return array_shift($this->items);
    }
    public function unshift($v) { array_unshift($this->items, $v); }
    public function top() {
        if (!$this->items) { throw new RuntimeException("Can't peek at an empty datastructure"); }
        return $this->items[count($this->items) - 1];
    }
    public function bottom() {
        if (!$this->items) { throw new RuntimeException("Can't peek at an empty datastructure"); }
        return $this->items[0];
    }
    public function isEmpty() { return count($this->items) === 0; }
    public function count(): int { return count($this->items); }
    public function toArray() { return $this->items; }
    public function offsetExists($o): bool { return isset($this->items[$o]); }
    public function offsetGet($o): mixed {
        if (!isset($this->items[$o]) && !array_key_exists($o, $this->items)) { throw new OutOfRangeException('SplDoublyLinkedList::offsetGet(): Argument #1 ($index) is out of range'); }
        return $this->items[$o];
    }
    public function offsetSet($o, $v): void { if ($o === null) { $this->items[] = $v; } else { $this->items[$o] = $v; } }
    public function offsetUnset($o): void { unset($this->items[$o]); $this->items = array_values($this->items); }
    public function rewind(): void { $this->pos = $this->lifo ? count($this->items) - 1 : 0; }
    public function valid(): bool { return $this->pos >= 0 && $this->pos < count($this->items); }
    public function current(): mixed { return $this->items[$this->pos] ?? null; }
    public function key(): mixed { return $this->pos; }
    public function next(): void { if ($this->lifo) { $this->pos--; } else { $this->pos++; } }
}
class SplQueue extends SplDoublyLinkedList {
    public function enqueue($v) { $this->push($v); }
    public function dequeue() { return $this->shift(); }
}
class SplStack extends SplDoublyLinkedList {
    protected $lifo = true;
}

class SplFixedArray implements Iterator, Countable, ArrayAccess {
    private $items = [];
    private $size = 0;
    private $pos = 0;
    public function __construct(int $size = 0) { $this->size = $size; for ($i = 0; $i < $size; $i++) { $this->items[$i] = null; } }
    public function getSize() { return $this->size; }
    public function setSize(int $size) {
        for ($i = $this->size; $i < $size; $i++) { $this->items[$i] = null; }
        for ($i = $size; $i < $this->size; $i++) { unset($this->items[$i]); }
        $this->size = $size;
    }
    public function toArray() { return $this->items; }
    public function count(): int { return $this->size; }
    public function offsetExists($o): bool { return $o >= 0 && $o < $this->size && $this->items[$o] !== null; }
    public function offsetGet($o): mixed {
        if (!is_int($o) || $o < 0 || $o >= $this->size) { throw new RuntimeException('Index invalid or out of range'); }
        return $this->items[$o];
    }
    public function offsetSet($o, $v): void {
        if (!is_int($o) || $o < 0 || $o >= $this->size) { throw new RuntimeException('Index invalid or out of range'); }
        $this->items[$o] = $v;
    }
    public function offsetUnset($o): void { $this->items[$o] = null; }
    public function rewind(): void { $this->pos = 0; }
    public function valid(): bool { return $this->pos < $this->size; }
    public function current(): mixed { return $this->items[$this->pos]; }
    public function key(): mixed { return $this->pos; }
    public function next(): void { $this->pos++; }
}

interface OuterIterator extends Iterator { public function getInnerIterator(): ?Iterator; }
interface RecursiveIterator extends Iterator {
    public function hasChildren(): bool;
    public function getChildren(): ?RecursiveIterator;
}

class IteratorIterator implements OuterIterator {
    protected $inner;
    public function __construct(Traversable $iterator) {
        while ($iterator instanceof IteratorAggregate) { $iterator = $iterator->getIterator(); }
        $this->inner = $iterator;
    }
    public function getInnerIterator(): ?Iterator { return $this->inner; }
    public function rewind(): void { $this->inner->rewind(); }
    public function valid(): bool { return $this->inner->valid(); }
    public function current(): mixed { return $this->inner->current(); }
    public function key(): mixed { return $this->inner->key(); }
    public function next(): void { $this->inner->next(); }
}

abstract class FilterIterator extends IteratorIterator {
    abstract public function accept(): bool;
    public function __construct(Iterator $iterator) { parent::__construct($iterator); }
    private function fetch() { while ($this->inner->valid() && !$this->accept()) { $this->inner->next(); } }
    public function rewind(): void { $this->inner->rewind(); $this->fetch(); }
    public function next(): void { $this->inner->next(); $this->fetch(); }
}

class CallbackFilterIterator extends FilterIterator {
    private $callback;
    public function __construct(Iterator $iterator, callable $callback) { parent::__construct($iterator); $this->callback = $callback; }
    public function accept(): bool { return (bool) call_user_func($this->callback, $this->current(), $this->key(), $this->inner); }
}

abstract class RecursiveFilterIterator extends FilterIterator implements RecursiveIterator {
    public function __construct(RecursiveIterator $iterator) { parent::__construct($iterator); }
    public function hasChildren(): bool { return $this->inner->hasChildren(); }
    public function getChildren(): ?RecursiveIterator { return new static($this->inner->getChildren()); }
}

class LimitIterator extends IteratorIterator {
    private $offset;
    private $limit;
    private $pos = 0;
    public function __construct(Iterator $iterator, int $offset = 0, int $limit = -1) { parent::__construct($iterator); $this->offset = $offset; $this->limit = $limit; }
    public function rewind(): void { $this->inner->rewind(); $this->pos = 0; while ($this->pos < $this->offset && $this->inner->valid()) { $this->inner->next(); $this->pos++; } }
    public function valid(): bool { return ($this->limit === -1 || $this->pos < $this->offset + $this->limit) && $this->inner->valid(); }
    public function next(): void { $this->pos++; $this->inner->next(); }
}

class EmptyIterator implements Iterator {
    public function current(): never { throw new BadMethodCallException('Accessing the value of an EmptyIterator'); }
    public function key(): never { throw new BadMethodCallException('Accessing the key of an EmptyIterator'); }
    public function next(): void {}
    public function rewind(): void {}
    public function valid(): bool { return false; }
}

class RecursiveArrayIterator extends ArrayIterator implements RecursiveIterator {
    public function hasChildren(): bool { $c = $this->current(); return is_array($c) || is_object($c); }
    public function getChildren(): ?RecursiveIterator { return new static($this->current()); }
}

class RecursiveIteratorIterator implements OuterIterator {
    const LEAVES_ONLY = 0;
    const SELF_FIRST = 1;
    const CHILD_FIRST = 2;
    private $root;
    private $mode;
    private $items = [];
    private $pos = 0;
    public function __construct(Traversable $iterator, int $mode = self::LEAVES_ONLY, int $flags = 0) {
        while ($iterator instanceof IteratorAggregate) { $iterator = $iterator->getIterator(); }
        $this->root = $iterator;
        $this->mode = $mode;
    }
    private function walk($it, $depth) {
        for ($it->rewind(); $it->valid(); $it->next()) {
            $key = $it->key();
            $value = $it->current();
            if ($it->hasChildren()) {
                if ($this->mode === self::SELF_FIRST) { $this->items[] = [$key, $value, $depth]; }
                $this->walk($it->getChildren(), $depth + 1);
                if ($this->mode === self::CHILD_FIRST) { $this->items[] = [$key, $value, $depth]; }
            } else {
                $this->items[] = [$key, $value, $depth];
            }
        }
    }
    public function getInnerIterator(): ?Iterator { return $this->root; }
    public function rewind(): void { $this->items = []; $this->pos = 0; $this->walk($this->root, 0); }
    public function valid(): bool { return $this->pos < count($this->items); }
    public function current(): mixed { return $this->items[$this->pos][1]; }
    public function key(): mixed { return $this->items[$this->pos][0]; }
    public function next(): void { $this->pos++; }
    public function getDepth() { return $this->items[$this->pos][2] ?? 0; }
}

interface UnitEnum { public static function cases(): array; }
interface BackedEnum extends UnitEnum {}
class UnhandledMatchError extends Error {}
