name: a pattern closes at its first unescaped delimiter and carries known modifiers
description: >
  php reads the closing delimiter as the first unescaped one, so "//x//i" is
  an empty body with the modifiers "x//i" and does not compile; an unknown
  modifier, a missing delimiter, an alphanumeric delimiter and an empty pattern
  all answer false. The Mention extension relies on an invalid pattern answering
  false. Expected output is php 8.4's.
---
<?php
var_dump(@preg_match('//x//i', ''));
var_dump(@preg_match('/a/b/', 'a/b'));
var_dump(preg_match('#a/b#', 'a/b'));
var_dump(preg_match('/a\/b/', 'a/b'));
var_dump(@preg_match('', ''));
var_dump(@preg_match('abc', 'abc'));
var_dump(preg_match('//', 'x'));
---
bool(false)
bool(false)
int(1)
int(1)
bool(false)
bool(false)
int(1)
