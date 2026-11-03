---
name: clock, array and edit-distance bindings
description: >
  gmdate, array_product, array_replace, levenshtein, uniqid and hrtime, the
  forms vendor code calls. The expected section is what `php` printed.
---
<?php
var_dump(gmdate("Y-m-d H:i:s", 86400 * 365), array_product([2, "3", 4]), array_product([]), array_product([1.5, 2]));
print_r(array_replace([1, 2, 3], [1 => "b"], [3 => "d"]));
var_dump(levenshtein("kitten", "sitting"), levenshtein("", "abc"), levenshtein("same", "same"));
var_dump(strlen(uniqid()), strlen(uniqid("x_")), is_int(hrtime(true)), count(hrtime()));
---
string(19) "1971-01-01 00:00:00"
int(24)
int(1)
float(3)
Array
(
    [0] => 1
    [1] => b
    [2] => 3
    [3] => d
)
int(3)
int(3)
int(0)
int(13)
int(15)
bool(true)
int(2)
