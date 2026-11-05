name: the u modifier reads Unicode and refuses invalid UTF-8
description: >
  With /u, \w \d and \s are Unicode classes, \b follows them, and a subject that is not valid UTF-8 makes preg_match, preg_split and preg_match_all answer false with preg_last_error() 4. Expected output is php 8.4's.
runner:
  flatstack: false
---
<?php
var_dump(preg_match('/\p{Cc}+/u', "\xe2"), preg_last_error(), preg_last_error_msg());
var_dump(preg_match('/./u', "\xe2\x82"), preg_match('/./', "\xe2"));
var_dump(preg_replace('/a/u', 'b', "a\xff"), preg_split('/,/u', "a,\xffb"), preg_match_all('/./u', "a\xffb", $m), preg_last_error());
var_dump(preg_match('/^[\x{4e00}-\x{9fa5}]+$/u', '漢字'), preg_match('/\p{Han}/u', '漢'), preg_match('/^\p{Lu}\p{Ll}+$/u', 'Éa'));
var_dump(preg_match('/\w+/u', 'héllo', $mm), $mm, preg_match('/\w+/', 'héllo', $m2), $m2);
---
bool(false)
int(4)
string(56) "Malformed UTF-8 characters, possibly incorrectly encoded"
bool(false)
int(1)
NULL
bool(false)
bool(false)
int(4)
int(1)
int(1)
int(1)
int(1)
array(1) {
  [0]=>
  string(6) "héllo"
}
int(1)
array(1) {
  [0]=>
  string(1) "h"
}
