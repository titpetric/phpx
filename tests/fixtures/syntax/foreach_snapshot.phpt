name: foreach over an array the body writes
description: >
  php iterates the array as it was when the loop began. Appending to it,
  unsetting from it or writing an element inside the body leaves the walk
  alone, and the source keeps the writes.
---
<?php

$a = [1, 2, 3];
foreach ($a as $v) {
  echo $v, "\n";
  $a[] = $v;
}
echo count($a), "\n";

$b = [1, 2, 3];
foreach ($b as $k => $v) {
  echo $v, "\n";
  unset($b[$k]);
}
echo count($b), "\n";

$c = [1, 2, 3];
foreach ($c as $v) {
  echo $v, "\n";
  $c[0] = 99;
}
echo implode(",", $c), "\n";

?>
---
1
2
3
6
1
2
3
0
1
2
3
99,2,3
