name: str_getcsv, fputcsv and fgetcsv follow php
description: >
  A blank line is a record of one null, a field may span lines, the escape character is kept, and fputcsv encloses a field holding a space, a tab or a line break. Expected output is php 8.4's.
runner:
  flatstack: false
---
<?php
var_dump(str_getcsv('a,b,c'), str_getcsv(''), str_getcsv('a, b ,"c d",,"e""f"'), str_getcsv("x;y;'z;w'", ';', "'"), str_getcsv('"a\\"b",c'), str_getcsv('"a\\"b",c', ',', '"', ''));
$h = fopen('php://memory', 'w+');
echo fputcsv($h, ['plain', 'has space', 'has,comma', 'has"quote', "new\nline", '', 12, 1.5, true, null]);
echo fputcsv($h, ['a', 'b'], ';', "'", '\\', "!\n");
rewind($h); echo stream_get_contents($h);
$m = fopen('php://memory', 'w+'); fwrite($m, "a,b\n\nc,\"d\ne\"\nlast"); rewind($m);
while (($r = fgetcsv($m)) !== false) { echo json_encode($r), "\n"; }
---
array(3) {
  [0]=>
  string(1) "a"
  [1]=>
  string(1) "b"
  [2]=>
  string(1) "c"
}
array(1) {
  [0]=>
  NULL
}
array(5) {
  [0]=>
  string(1) "a"
  [1]=>
  string(3) " b "
  [2]=>
  string(3) "c d"
  [3]=>
  string(0) ""
  [4]=>
  string(3) "e"f"
}
array(3) {
  [0]=>
  string(1) "x"
  [1]=>
  string(1) "y"
  [2]=>
  string(3) "z;w"
}
array(2) {
  [0]=>
  string(4) "a\"b"
  [1]=>
  string(1) "c"
}
array(2) {
  [0]=>
  string(4) "a\b""
  [1]=>
  string(1) "c"
}
655plain,"has space","has,comma","has""quote","new
line",,12,1.5,1,
a;b!
["a","b"]
[null]
["c","d\ne"]
["last"]
