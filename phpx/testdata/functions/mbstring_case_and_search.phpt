name: mb_convert_case and the mb search family
description: >
  Full Unicode case mapping, mb_ord and mb_chr, the mb_str* searches that answer a part of the haystack, the 8.4 trims. Expected output is php 8.4's.
runner:
  flatstack: false
---
<?php
var_dump(mb_convert_case('hello wORLD', MB_CASE_TITLE, 'UTF-8'), mb_convert_case('Straße', MB_CASE_UPPER), mb_convert_case('ÀB', MB_CASE_LOWER), mb_convert_case("o'neil mc-donald jr.", MB_CASE_TITLE), mb_strtoupper('éàß'), mb_strtolower('ÉÀ'));
var_dump(mb_ord('€'), mb_chr(8364), mb_chr(0x1F600), mb_chr(0xD800));
var_dump(mb_internal_encoding(), mb_internal_encoding('UTF-8'), mb_substitute_character(), mb_language());
var_dump(mb_strstr('user@example.com', '@'), mb_strstr('user@example.com', '@', true), mb_stristr('HeLLo', 'LL'), mb_strrchr('a/b/c', '/'), mb_strrichr('aXbxc', 'X', true), mb_strripos('ababAB', 'ab'), mb_strripos('héllo hÉllo', 'É'));
var_dump(mb_trim("  héllo \u{00a0}"), mb_ltrim("xxhixx", 'x'), mb_rtrim("xxhixx", 'x'), mb_strcut('héllo', 2, 3), mb_scrub("a\xffb"));
---
string(11) "Hello World"
string(7) "STRASSE"
string(3) "àb"
string(20) "O'neil Mc-Donald Jr."
string(6) "ÉÀSS"
string(4) "éà"
int(8364)
string(3) "€"
string(4) "😀"
bool(false)
string(5) "UTF-8"
bool(true)
int(63)
string(7) "neutral"
string(12) "@example.com"
string(4) "user"
string(3) "LLo"
string(2) "/c"
string(3) "aXb"
int(4)
int(7)
string(6) "héllo"
string(4) "hixx"
string(4) "xxhi"
string(3) "él"
string(3) "a?b"
