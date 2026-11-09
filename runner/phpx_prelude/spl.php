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
    private $has = false;
    private $cv = null;
    private $ck = null;
    public function __construct(Traversable $iterator) {
        while ($iterator instanceof IteratorAggregate) { $iterator = $iterator->getIterator(); }
        $this->inner = $iterator;
    }
    public function getInnerIterator(): ?Iterator { return $this->inner; }
    // php forwards a method the wrapper does not declare to the inner iterator,
    // which is how symfony/finder's filters call getFilename() and isDir().
    public function __call($name, $args) { return $this->inner->$name(...$args); }
    // php caches the inner iterator's current element and key when it moves,
    // and answers valid(), current() and key() from the cache, so an inner
    // iterator that does work when asked (a stream) is asked once per step.
    protected function fetchInner(): bool {
        if ($this->inner->valid()) {
            $this->cv = $this->inner->current();
            $this->ck = $this->inner->key();
            return $this->has = true;
        }
        $this->has = false;
        $this->cv = null;
        $this->ck = null;
        return false;
    }
    public function rewind(): void { $this->inner->rewind(); $this->fetchInner(); }
    public function valid(): bool { return $this->has; }
    public function current(): mixed { return $this->cv; }
    public function key(): mixed { return $this->ck; }
    public function next(): void { $this->inner->next(); $this->fetchInner(); }
}

abstract class FilterIterator extends IteratorIterator {
    abstract public function accept(): bool;
    public function __construct(Iterator $iterator) { parent::__construct($iterator); }
    private function fetchAccepted(): void {
        while ($this->fetchInner()) {
            if ($this->accept()) { return; }
            $this->inner->next();
        }
    }
    public function rewind(): void { $this->inner->rewind(); $this->fetchAccepted(); }
    public function next(): void { $this->inner->next(); $this->fetchAccepted(); }
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
    public function current(): mixed { return $this->inner->current(); }
    public function key(): mixed { return $this->inner->key(); }
}

class AppendIterator extends IteratorIterator {
    private $iterators = [];
    private $index = 0;
    public function __construct() {}
    public function append(Iterator $iterator): void {
        $this->iterators[] = $iterator;
        if (count($this->iterators) === 1) { $this->index = 0; $this->inner = $iterator; $iterator->rewind(); }
        elseif ($this->inner === null || !$this->inner->valid()) { $this->skip(); }
    }
    public function getInnerIterator(): ?Iterator { return $this->inner; }
    public function getIteratorIndex(): ?int { return $this->inner === null ? null : $this->index; }
    public function getArrayIterator(): ArrayIterator { return new ArrayIterator($this->iterators); }
    private function skip(): void {
        while ($this->inner !== null && !$this->inner->valid() && $this->index + 1 < count($this->iterators)) {
            $this->index++;
            $this->inner = $this->iterators[$this->index];
            $this->inner->rewind();
        }
    }
    public function rewind(): void {
        $this->index = 0;
        $this->inner = $this->iterators[0] ?? null;
        if ($this->inner !== null) { $this->inner->rewind(); $this->skip(); }
    }
    public function valid(): bool { return $this->inner !== null && $this->inner->valid(); }
    public function current(): mixed { return $this->valid() ? $this->inner->current() : null; }
    public function key(): mixed { return $this->valid() ? $this->inner->key() : null; }
    public function next(): void { if ($this->inner === null) { return; } $this->inner->next(); $this->skip(); }
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
    const CATCH_GET_CHILD = 16;
    private $root;
    private $mode;
    private $items = [];
    private $pos = 0;
    private $maxDepth = -1;
    public function __construct(Traversable $iterator, int $mode = self::LEAVES_ONLY, int $flags = 0) {
        while ($iterator instanceof IteratorAggregate) { $iterator = $iterator->getIterator(); }
        $this->root = $iterator;
        $this->mode = $mode;
    }
    public function setMaxDepth(int $maxDepth = -1): void { $this->maxDepth = $maxDepth; }
    public function getMaxDepth() { return $this->maxDepth < 0 ? false : $this->maxDepth; }
    private function walk($it, $depth) {
        for ($it->rewind(); $it->valid(); $it->next()) {
            $key = $it->key();
            $value = $it->current();
            $descend = $it->hasChildren() && ($this->maxDepth < 0 || $depth < $this->maxDepth);
            if ($descend) {
                if ($this->mode === self::SELF_FIRST) { $this->items[] = [$key, $value, $depth, clone $it]; }
                $this->walk($it->getChildren(), $depth + 1);
                if ($this->mode === self::CHILD_FIRST) { $this->items[] = [$key, $value, $depth, clone $it]; }
            } else {
                $this->items[] = [$key, $value, $depth, clone $it];
            }
        }
    }
    public function getInnerIterator(): ?Iterator { return $this->items[$this->pos][3] ?? $this->root; }
    public function rewind(): void { $this->items = []; $this->pos = 0; $this->walk($this->root, 0); }
    public function valid(): bool { return $this->pos < count($this->items); }
    public function current(): mixed { return $this->items[$this->pos][1]; }
    public function key(): mixed { return $this->items[$this->pos][0]; }
    public function next(): void { $this->pos++; }
    public function getDepth() { return $this->items[$this->pos][2] ?? 0; }
    public function getSubPathname() { return $this->getInnerIterator()->getSubPathname(); }
    public function getSubPath() { return $this->getInnerIterator()->getSubPath(); }
}

interface UnitEnum { public static function cases(): array; }
interface BackedEnum extends UnitEnum {}
class UnhandledMatchError extends Error {}

final class WeakMap implements Countable, ArrayAccess, IteratorAggregate {
    private $objs = [];
    private $vals = [];
    public function count(): int { return count($this->objs); }
    public function offsetExists($o): bool { return isset($this->vals[spl_object_id($o)]); }
    public function offsetGet($o): mixed {
        $i = spl_object_id($o);
        if (!array_key_exists($i, $this->objs)) { throw new Error('Object WeakMap#' . $i . ' not contained in WeakMap'); }
        return $this->vals[$i];
    }
    public function offsetSet($o, $v): void { $i = spl_object_id($o); $this->objs[$i] = $o; $this->vals[$i] = $v; }
    public function offsetUnset($o): void { $i = spl_object_id($o); unset($this->objs[$i], $this->vals[$i]); }
    public function getIterator(): Iterator { return (function () { foreach ($this->objs as $i => $o) { yield $o => $this->vals[$i]; } })(); }
}

class SplFileInfo implements Stringable {
    protected $path;
    public function __construct(string $filename) { $this->path = $filename; }
    protected function fi(): string { return $this->path; }
    public function getPathname(): string { return $this->fi(); }
    public function getFilename(): string { return basename($this->fi()); }
    public function getPath(): string { $d = dirname($this->fi()); return $d === '.' && strpos($this->fi(), '/') === false ? '' : $d; }
    public function getExtension(): string { return pathinfo($this->fi(), PATHINFO_EXTENSION); }
    public function getBasename(string $suffix = ''): string { return basename($this->fi(), $suffix); }
    public function getRealPath() { return realpath($this->fi()); }
    public function isFile(): bool { return is_file($this->fi()); }
    public function isDir(): bool { return is_dir($this->fi()); }
    public function isLink(): bool { return is_link($this->fi()); }
    public function isReadable(): bool { return is_readable($this->fi()); }
    public function isWritable(): bool { return is_writable($this->fi()); }
    public function getSize() { return filesize($this->fi()); }
    public function getMTime() { return filemtime($this->fi()); }
    public function getType() { return is_dir($this->fi()) ? 'dir' : 'file'; }
    public function getPerms() { return fileperms($this->fi()); }
    public function __toString(): string { return $this->fi(); }
}

class FilesystemIterator extends SplFileInfo implements Iterator {
    const CURRENT_MODE_MASK = 240;
    const CURRENT_AS_PATHNAME = 32;
    const CURRENT_AS_FILEINFO = 0;
    const CURRENT_AS_SELF = 16;
    const KEY_MODE_MASK = 3840;
    const KEY_AS_PATHNAME = 0;
    const FOLLOW_SYMLINKS = 512;
    const KEY_AS_FILENAME = 256;
    const NEW_CURRENT_AND_KEY = 256;
    const OTHER_MODE_MASK = 12288;
    const SKIP_DOTS = 4096;
    const UNIX_PATHS = 8192;
    protected $flags;
    protected $entries = [];
    protected $pos = 0;
    public function __construct(string $directory, int $flags = 4096) {
        parent::__construct($directory);
        $this->flags = $flags;
        $names = scandir($directory);
        if ($names === false) { throw new UnexpectedValueException("FilesystemIterator::__construct($directory): Failed to open directory: No such file or directory"); }
        foreach ($names as $n) {
            if (($n === '.' || $n === '..') && ($this->flags & self::SKIP_DOTS || true)) { continue; }
            $this->entries[] = $n;
        }
    }
    protected function fi(): string { return $this->entries ? rtrim($this->path, '/') . '/' . ($this->entries[$this->pos] ?? '') : $this->path; }
    public function getFlags(): int { return $this->flags; }
    public function setFlags(int $flags): void { $this->flags = $flags; }
    public function rewind(): void { $this->pos = 0; }
    public function valid(): bool { return $this->pos < count($this->entries); }
    public function key(): mixed { return rtrim($this->path, '/') . '/' . $this->entries[$this->pos]; }
    public function current(): mixed {
        $p = rtrim($this->path, '/') . '/' . $this->entries[$this->pos];
        if ($this->flags & self::CURRENT_AS_PATHNAME) { return $p; }
        return new SplFileInfo($p);
    }
    public function next(): void { $this->pos++; }
}

class RecursiveDirectoryIterator extends FilesystemIterator implements RecursiveIterator {
    private $sub = '';
    public function __construct(string $directory, int $flags = 0) { parent::__construct($directory, $flags); }
    public function hasChildren(bool $allowLinks = false): bool {
        $p = rtrim($this->path, '/') . '/' . $this->entries[$this->pos];
        return is_dir($p) && ($allowLinks || !is_link($p) || ($this->flags & self::FOLLOW_SYMLINKS));
    }
    public function getChildren(): RecursiveDirectoryIterator {
        $c = new static(rtrim($this->path, '/') . '/' . $this->entries[$this->pos], $this->flags);
        $c->sub = ltrim($this->sub . '/' . $this->entries[$this->pos], '/');
        return $c;
    }
    public function getSubPath(): string { return $this->sub; }
    public function getSubPathname(): string { return ltrim($this->sub . '/' . $this->entries[$this->pos], '/'); }
}

class DirectoryIterator extends SplFileInfo implements Iterator {
    private $dir;
    private $names = [];
    private $pos = 0;
    public function __construct(string $directory) {
        $names = scandir($directory);
        if ($names === false) { throw new UnexpectedValueException("DirectoryIterator::__construct($directory): Failed to open directory: No such file or directory"); }
        $this->dir = rtrim($directory, '/');
        $this->names = $names;
        parent::__construct($this->dir . '/' . ($names[0] ?? ''));
    }
    protected function fi(): string { return $this->dir . '/' . ($this->names[$this->pos] ?? ''); }
    public function isDot(): bool { $n = $this->names[$this->pos] ?? ''; return $n === '.' || $n === '..'; }
    public function rewind(): void { $this->pos = 0; $this->path = $this->dir . '/' . ($this->names[0] ?? ''); }
    public function valid(): bool { return $this->pos < count($this->names); }
    public function key(): mixed { return $this->pos; }
    public function current(): mixed { return $this; }
    public function next(): void { $this->pos++; $this->path = $this->dir . '/' . ($this->names[$this->pos] ?? ''); }
}

interface DateTimeInterface {
    const ATOM = "Y-m-d\\TH:i:sP";
    const COOKIE = "l, d-M-Y H:i:s T";
    const ISO8601 = "Y-m-d\\TH:i:sO";
    const ISO8601_EXPANDED = "X-m-d\\TH:i:sP";
    const RFC822 = "D, d M y H:i:s O";
    const RFC850 = "l, d-M-y H:i:s T";
    const RFC1036 = "D, d M y H:i:s O";
    const RFC1123 = "D, d M Y H:i:s O";
    const RFC7231 = "D, d M Y H:i:s \\G\\M\\T";
    const RFC2822 = "D, d M Y H:i:s O";
    const RFC3339 = "Y-m-d\\TH:i:sP";
    const RFC3339_EXTENDED = "Y-m-d\\TH:i:s.vP";
    const RSS = "D, d M Y H:i:s O";
    const W3C = "Y-m-d\\TH:i:sP";
}

// Fibers are not implemented; code that asks whether it runs in one is told no.
final class Fiber {
    public static function getCurrent(): ?Fiber { return null; }
}

interface SeekableIterator extends Iterator { public function seek(int $offset): void; }

class SplFileObject extends SplFileInfo implements RecursiveIterator, SeekableIterator {
    const DROP_NEW_LINE = 1;
    const READ_AHEAD = 2;
    const SKIP_EMPTY = 4;
    const READ_CSV = 8;
    protected $fp;
    private $flags = 0;
    private $line = null;
    private $row = null;
    private $lineNum = 0;
    private $csv = [',', '"', '\\'];
    private $maxLen = 0;
    private $eofSeen = false;
    public function __construct(string $filename, string $mode = 'r', bool $useIncludePath = false, $context = null) {
        parent::__construct($filename);
        $fp = @fopen($filename, $mode, $useIncludePath);
        if ($fp === false) {
            throw new RuntimeException('SplFileObject::__construct(' . $filename . '): Failed to open stream: No such file or directory');
        }
        $this->fp = $fp;
    }
    private function freeLine(): void { $this->line = null; $this->row = null; }
    private function readLine(bool $silent, bool $lineAdd) {
        if ($this->eofSeen) {
            if (!$silent) { throw new RuntimeException('Cannot read from file ' . $this->fi()); }
            return false;
        }
        $buf = $this->maxLen > 0 ? fgets($this->fp, $this->maxLen + 1) : fgets($this->fp);
        if ($buf === false) {
            $this->eofSeen = true;
            $this->line = '';
        } else {
            if (substr($buf, -1) !== "\n") { $this->eofSeen = true; }
            if ($this->flags & self::DROP_NEW_LINE) {
                if (substr($buf, -1) === "\n") { $buf = substr($buf, 0, -1); }
                if (substr($buf, -1) === "\r") { $buf = substr($buf, 0, -1); }
            }
            $this->line = $buf;
        }
        if ($lineAdd) { $this->lineNum++; }
        return true;
    }
    private function readCsvRow(?array $ctl, bool $silent) {
        $first = true;
        do {
            if (!$first) { $this->lineNum++; }
            $first = false;
            $ok = $this->readLine($silent, false);
        } while ($ok && $this->line === '' && ($this->flags & self::SKIP_EMPTY));
        if (!$ok) { $this->freeLine(); return false; }
        [$sep, $enc, $esc] = $ctl ?? $this->csv;
        $raw = $this->line;
        while (substr_count($raw, $enc) % 2 === 1 && !$this->eofSeen) {
            $more = fgets($this->fp);
            if ($more === false) { $this->eofSeen = true; break; }
            if (substr($more, -1) !== "\n") { $this->eofSeen = true; }
            $raw .= $more;
        }
        $row = str_getcsv($raw, $sep, $enc, $esc);
        $this->row = $row;
        return true;
    }
    private function emptyLine(): bool {
        if ($this->row !== null) { return count($this->row) === 1 && $this->row[0] === null; }
        return $this->line === null || $this->line === '';
    }
    private function readCurrent(bool $silent): bool {
        $this->freeLine();
        $ok = ($this->flags & self::READ_CSV) ? $this->readCsvRow(null, $silent) : $this->readLine($silent, false);
        while (($this->flags & self::SKIP_EMPTY) && $ok && $this->emptyLine()) {
            $this->freeLine();
            $ok = ($this->flags & self::READ_CSV) ? $this->readCsvRow(null, $silent) : $this->readLine($silent, false);
        }
        if (!$ok) { $this->freeLine(); }
        return $ok;
    }
    public function rewind(): void {
        fseek($this->fp, 0);
        $this->eofSeen = false;
        $this->freeLine();
        $this->lineNum = 0;
        if ($this->flags & self::READ_AHEAD) { $this->readCurrent(true); }
    }
    public function eof(): bool { return $this->eofSeen; }
    public function valid(): bool {
        if ($this->flags & self::READ_AHEAD) { return $this->line !== null || $this->row !== null; }
        return !$this->eofSeen;
    }
    public function fgets(): string {
        $this->readLine(false, true);
        return $this->line;
    }
    public function current(): string|array|false {
        if ($this->line === null && $this->row === null) { $this->readCurrent(true); }
        if ($this->line !== null && (!($this->flags & self::READ_CSV) || $this->row === null)) { return $this->line; }
        if ($this->row !== null) { return $this->row; }
        return false;
    }
    public function key(): int { return $this->lineNum; }
    public function next(): void {
        $this->freeLine();
        if ($this->flags & self::READ_AHEAD) { $this->readCurrent(true); }
        $this->lineNum++;
    }
    public function seek(int $line): void {
        if ($line < 0) { throw new ValueError('SplFileObject::seek(): Argument #1 ($line) must be greater than or equal to 0'); }
        $this->rewind();
        for ($i = 0; $i < $line; $i++) {
            if ($this->readCurrent(true) === false) { return; }
            $this->lineNum++;
            $this->freeLine();
        }
        if ($line > 0 && ($this->flags & self::READ_AHEAD)) { $this->readCurrent(true); }
    }
    public function setFlags(int $flags): void { $this->flags = $flags; }
    public function getFlags(): int { return $this->flags; }
    public function setMaxLineLen(int $maxLength): void { $this->maxLen = $maxLength; }
    public function getMaxLineLen(): int { return $this->maxLen; }
    public function setCsvControl(string $separator = ',', string $enclosure = '"', string $escape = '\\'): void { $this->csv = [$separator, $enclosure, $escape]; }
    public function getCsvControl(): array { return $this->csv; }
    public function fgetcsv(string $separator = ',', string $enclosure = '"', string $escape = '\\'): array|false {
        $ok = $this->readCsvRow([$separator, $enclosure, $escape], true);
        if (!$ok) { return false; }
        $row = $this->row;
        $this->freeLine();
        $this->lineNum++;
        return $row;
    }
    public function fputcsv(array $fields, string $separator = ',', string $enclosure = '"', string $escape = '\\', string $eol = "\n"): int|false {
        return fputcsv($this->fp, $fields, $separator, $enclosure, $escape, $eol);
    }
    public function fwrite(string $data, int $length = 0): int|false { return $length > 0 ? fwrite($this->fp, $data, $length) : fwrite($this->fp, $data); }
    public function fread(int $length): string|false { return fread($this->fp, $length); }
    public function fgetc(): string|false { $c = fgetc($this->fp); $this->freeLine(); if ($c === "\n") { $this->lineNum++; } return $c; }
    public function fpassthru(): int { return fpassthru($this->fp); }
    public function fflush(): bool { return fflush($this->fp); }
    public function ftell(): int|false { return ftell($this->fp); }
    public function fseek(int $offset, int $whence = SEEK_SET): int { $this->freeLine(); $this->eofSeen = false; return fseek($this->fp, $offset, $whence); }
    public function ftruncate(int $size): bool { return ftruncate($this->fp, $size); }
    public function flock(int $operation, &$wouldBlock = null): bool { return flock($this->fp, $operation); }
    public function fstat(): array { return fstat($this->fp); }
    public function getCurrentLine(): string { return $this->fgets(); }
    public function hasChildren(): bool { return false; }
    public function getChildren(): ?RecursiveIterator { return null; }
    public function __toString(): string { return (string) $this->fgets(); }
}

class SplTempFileObject extends SplFileObject {
    public function __construct(int $maxMemory = 2 * 1024 * 1024) {
        $name = $maxMemory < 0 ? 'php://memory' : ($maxMemory === 2 * 1024 * 1024 ? 'php://temp' : 'php://temp/maxmemory:' . $maxMemory);
        parent::__construct($name, 'wb');
    }
}

class DatePeriod implements IteratorAggregate {
    const EXCLUDE_START_DATE = 1;
    const INCLUDE_END_DATE = 2;
    public $start;
    public $current = null;
    public $end = null;
    public $interval;
    public $recurrences;
    public $include_start_date;
    public $include_end_date;
    private $count = null;
    public function __construct($start, $interval, $end = null, $options = 0) {
        $this->start = $start;
        $this->interval = $interval;
        $this->include_start_date = !($options & 1);
        $this->include_end_date = (bool) ($options & 2);
        if (is_int($end)) {
            $this->count = $end;
            $this->recurrences = $end + ($this->include_start_date ? 1 : 0);
        } else {
            $this->end = $end;
            $this->recurrences = 1;
        }
    }
    public function getStartDate(): DateTimeInterface { return clone $this->start; }
    public function getEndDate(): ?DateTimeInterface { return $this->end === null ? null : clone $this->end; }
    public function getDateInterval(): DateInterval { return clone $this->interval; }
    public function getRecurrences(): ?int { return $this->count; }
    public function getIterator(): Iterator {
        $cur = clone $this->start;
        $i = 0;
        $n = 0;
        if (!$this->include_start_date) {
            $cur = $cur->add($this->interval);
        }
        while (true) {
            if ($this->end !== null) {
                if ($this->include_end_date ? $cur > $this->end : $cur >= $this->end) {
                    return;
                }
            } elseif ($n > $this->count - ($this->include_start_date ? 0 : 1)) {
                return;
            }
            yield $i++ => clone $cur;
            $n++;
            $cur = $cur->add($this->interval);
        }
    }
}
