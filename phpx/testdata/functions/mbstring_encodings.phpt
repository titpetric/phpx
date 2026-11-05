name: mb_detect_encoding and mb_convert_encoding know the legacy encodings
description: >
  'auto' is the default detection order, strict detection refuses malformed UTF-8, Windows-1252, ISO-8859-1 and UTF-16 convert both ways, and utf8_encode, utf8_decode and iconv agree. Expected output is php 8.4's.
runner:
  flatstack: false
---
<?php
var_dump(mb_detect_encoding("\xe2", 'auto', true), mb_detect_encoding("abc", 'auto', true), mb_detect_encoding("é", 'auto', true), mb_detect_encoding("\xe9", ['UTF-8', 'ISO-8859-1'], true), mb_detect_encoding("\xe9", 'ASCII, UTF-8', true));
var_dump(bin2hex(mb_convert_encoding("\xe2", 'UTF-8', 'Windows-1252')), bin2hex(mb_convert_encoding("é", 'ISO-8859-1', 'UTF-8')), bin2hex(mb_convert_encoding("\xe9", 'UTF-8', 'ISO-8859-1')), bin2hex(mb_convert_encoding("€", 'Windows-1252', 'UTF-8')), bin2hex(mb_convert_encoding("\x80", 'UTF-8', 'CP1252')), bin2hex(mb_convert_encoding("aé", 'UTF-16BE', 'UTF-8')), bin2hex(mb_convert_encoding("\x00a", 'UTF-8', 'UTF-16BE')), mb_convert_encoding("a<b", 'HTML-ENTITIES', 'UTF-8'));
var_dump(utf8_encode("\xe9") === "é", utf8_decode("é") === "\xe9", iconv('UTF-8', 'ISO-8859-1//TRANSLIT', "é") === "\xe9", iconv('ISO-8859-1', 'UTF-8', "\xe9"), mb_check_encoding("\xe9", 'UTF-8'), mb_check_encoding("é", 'UTF-8'));
---
bool(false)
string(5) "ASCII"
string(5) "UTF-8"
string(10) "ISO-8859-1"
bool(false)
string(4) "c3a2"
string(2) "e9"
string(4) "c3a9"
string(2) "80"
string(6) "e282ac"
string(8) "006100e9"
string(2) "61"
string(3) "a<b"
bool(true)
bool(true)
bool(true)
string(2) "é"
bool(false)
bool(true)
