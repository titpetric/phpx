name: preg_replace over lists, with a limit and a count
description: >
  A list of patterns, a list of subjects, a limit, the count by reference, preg_filter and preg_replace_callback_array. Expected output is php 8.4's.
runner:
  flatstack: false
---
<?php
var_dump(preg_replace(['/a/', '/b/'], 'X', 'aabb'), preg_replace(['/a/', '/b/'], ['1'], 'aabb'), preg_replace(['/(a)/', '/(b)/'], '<\1>', 'ab'), preg_replace('/(a)(b)/', '\2${1}$0\\\\x', 'ab'));
var_dump(preg_replace('/a/', 'X', ['k1' => 'aa', 'k2' => 'bb', 'ccc'], 1, $count), $count);
var_dump(preg_replace('/\s+/', ' ', "a  b\n\tc"), preg_replace('/x/', 'y', 'abc', -1, $c2), $c2);
var_dump(preg_filter('/\d/', '#', ['a1', 'b', 'c22']), preg_filter('/\d/', '#', 'abc'), preg_filter(['/a/', '/b/'], ['A', 'B'], 'abc'));
var_dump(preg_replace_callback('/\d+/', fn($m) => $m[0] * 2, ['a1', 'b22'], -1, $n), $n);
var_dump(preg_replace_callback_array(['/a/' => fn($m) => 'b', '/b/' => fn($m) => 'c'], 'aabb', -1, $k), $k);
var_dump(preg_replace('/(?<w>\w+)@/', '[$1]', 'me@x'), preg_replace('~/+~', '/', 'a//b///c'), preg_replace('/é/u', 'e', 'café'));
---
string(4) "XXXX"
string(2) "11"
string(6) "<a><b>"
string(6) "baab\x"
array(3) {
  ["k1"]=>
  string(2) "Xa"
  ["k2"]=>
  string(2) "bb"
  [0]=>
  string(3) "ccc"
}
int(1)
string(5) "a b c"
string(3) "abc"
int(0)
array(2) {
  [0]=>
  string(2) "a#"
  [2]=>
  string(3) "c##"
}
NULL
string(3) "ABc"
array(2) {
  [0]=>
  string(2) "a2"
  [1]=>
  string(3) "b44"
}
int(2)
string(4) "cccc"
int(6)
string(5) "[me]x"
string(5) "a/b/c"
string(4) "cafe"
