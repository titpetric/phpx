name: preg_grep keeps keys and takes PREG_GREP_INVERT
description: >
  The matching elements keep their keys, and the third argument inverts the selection. Expected output is php 8.4's.
runner:
  flatstack: false
---
<?php
$a = ['apple', 'bob', 'cat', 'apricot', 5 => 'x1', 'k' => 'y2'];
print_r(preg_grep('/^a/', $a)); print_r(preg_grep('/^a/', $a, PREG_GREP_INVERT)); print_r(preg_grep('/\d/', $a)); print_r(preg_grep('/z/', []));
---
Array
(
    [0] => apple
    [3] => apricot
)
Array
(
    [1] => bob
    [2] => cat
    [5] => x1
    [k] => y2
)
Array
(
    [5] => x1
    [k] => y2
)
Array
(
)
