name: IteratorIterator asks its inner iterator once per step
description: >
  The wrapper caches the inner current and key when it moves and answers valid() from the cache, which the call trace shows. A stream-backed inner iterator loses its last record otherwise. Expected output is php 8.4's.
runner:
  flatstack: false
---
<?php
class Src implements SeekableIterator {
  private $i = 0; private $d = ['a', 'b', 'c', 'd'];
  function rewind(): void { echo "[rw]"; $this->i = 0; } function valid(): bool { echo "[v", $this->i, "]"; return $this->i < count($this->d); }
  function current(): mixed { echo "[c", $this->i, "]"; return $this->d[$this->i]; } function key(): mixed { return $this->i; }
  function next(): void { echo "[n]"; $this->i++; } function seek(int $o): void { $this->i = $o; }
}
$it = new CallbackFilterIterator(new IteratorIterator(new Src), fn($v) => true);
foreach ($it as $k => $v) { echo "$k=$v "; } echo "\n";
---
[rw][v0][c0]0=a [n][v1][c1]1=b [n][v2][c2]2=c [n][v3][c3]3=d [n][v4]
