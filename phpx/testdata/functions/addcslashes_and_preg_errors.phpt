name: addcslashes ranges, an invalid pattern's error, the encodings list
description: >
  addcslashes reads a..b ranges and writes unprintable bytes in octal, an invalid pattern leaves PREG_INTERNAL_ERROR, mb_list_encodings names the encodings the runtime handles. Expected output is php 8.4's.
runner:
  flatstack: false
---
<?php
var_dump(@preg_match('/([[:lower:]{1{/', ''), preg_last_error(), preg_last_error_msg());
var_dump(@preg_match('/ok/', 'ok'), preg_last_error());
var_dump(@preg_match('/([[:lower:]{1{/', ''), preg_last_error());
echo addcslashes("a\x01b\x7f", "\0..\37\177"), '|', addcslashes("foo[bar]", 'A..Z'), "|", addcslashes("tab\there\n", "\t\n"), "\n";
var_dump(in_array('UTF-16LE', mb_list_encodings()), in_array('UTF-8', mb_list_encodings()), in_array('ASCII', mb_list_encodings()));
---
bool(false)
int(1)
string(14) "Internal error"
int(1)
int(0)
bool(false)
int(1)
a\001b\177|foo[bar]|tab\there\n
bool(true)
bool(true)
bool(true)
