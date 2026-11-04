name: AppendIterator walks its iterators in turn, skipping empty ones
description: >
  The bytecode engine reads an object's properties in a foreach where the
  interpreter calls the Iterator methods, so the fixture opts it out.
runner:
  flatstack: false
---
<?php
$a = new AppendIterator;
$a->append(new ArrayIterator([1, 2]));
$a->append(new ArrayIterator([]));
$a->append(new ArrayIterator(['x' => 3]));
foreach ($a as $k => $v) { echo $k, '=', $v, ' idx ', $a->getIteratorIndex(), "\n"; }
echo implode(',', iterator_to_array($a, false)), "\n";
$e = new AppendIterator; var_dump(iterator_count($e), $e->valid());
---
0=1 idx 0
1=2 idx 0
x=3 idx 2
1,2,3
int(0)
bool(false)
