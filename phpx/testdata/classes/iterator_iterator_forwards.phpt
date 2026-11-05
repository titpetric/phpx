name: IteratorIterator forwards unknown methods to the inner iterator
description: >
  A FilterIterator subclass calls getFilename() and isDir(), which only the SplFileInfo-like inner iterator declares. symfony/finder's ExcludeDirectoryFilterIterator does this. Expected output is php 8.4's.
runner:
  flatstack: false
---
<?php
class Inner extends ArrayIterator {
    public function label() { return 'L' . $this->key(); }
    public function twice($n) { return $n * 2; }
}
class F extends FilterIterator {
    public function accept(): bool { return $this->label() !== 'L1'; }
}
$f = new F(new Inner(['a', 'b', 'c']));
foreach ($f as $k => $v) { echo $k, $v, $f->twice($k), "\n"; }
---
0a0
2c4
