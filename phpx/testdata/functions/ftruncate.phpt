name: ftruncate cuts a stream to a size and leaves the position alone
description: >
  phpunit's test-run history truncates the cache file through ftruncate before
  writing it. Expected output is php 8.4's.
runner:
  flatstack: false
---
<?php
$p = 'ftruncate_tmp.txt';
$f = fopen($p, 'w+');
fwrite($f, 'abcdef');
var_dump(ftruncate($f, 2));
var_dump(ftell($f));
rewind($f);
var_dump(fread($f, 10));
fclose($f);
echo file_get_contents($p), "\n";
unlink($p);
---
bool(true)
int(6)
string(2) "ab"
ab
