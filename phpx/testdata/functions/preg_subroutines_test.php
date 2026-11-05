<?php
$re = '~(?(DEFINE) (?<label> [a-z_][a-z0-9_]* ) (?<num> \d+ ) ) \$ (?&label) | (?&num) \. (?&num)~xi';
foreach (['$foo_1', '12.5', 'x', '$1'] as $s) { var_dump(preg_match($re, $s, $m)); echo json_encode($m), "\n"; }
$brace = '~\{ ( [^{}]* | (?R) )* \}~x';
var_dump(preg_match('~\{(?:[^{}]|(?R))*\}~', 'a {b {c} d} e', $m), $m[0]);
var_dump(preg_match('~(\((?:[^()]|(?1))*\))~', 'f(a(b)c)(d)', $m), $m[1]);
var_dump(preg_match('~(?<p>\((?:[^()]|(?&p))*\))~', 'x((y))z', $m), $m['p']);
var_dump(preg_match('~(a)(?-1)~', 'aa'), preg_match('~(?+1)(b)~', 'bb'));
$var = '\{ \s* ([a-zA-Z_][a-zA-Z0-9_-]*) \s* (?: : \s* ([^{}]*(?:\{(?-1)\}[^{}]*)*) )? \}';
var_dump(preg_match("~$var~x", '{id:[0-9]+}', $m), $m);
$variable = '\{ \s* ([a-zA-Z_][a-zA-Z0-9_-]*) \s* (?: : \s* ( [^{}]* (?: \{ (?-1) \} [^{}]* )* ) )? \}';
print_r(preg_split("~$variable(*SKIP)(*F) | \[~x", 'a[b{id:[0-9]+}c[d'));
var_dump(preg_match_all('~"[^"]*"(*SKIP)(*F)|\d+~', 'a1 "b2" c3 "4" 55', $mm), $mm);
echo preg_replace('~\([^)]*\)(*SKIP)(*FAIL)|,~', ';', 'a,b(c,d),e'), "\n";
var_dump(preg_match('~x(*SKIP)(*F)|y~', 'xy', $m4, PREG_OFFSET_CAPTURE), $m4);
// branch reset
var_dump(preg_match('~^(?|/a/(\d+)()|/b/(\w+)/(\w+)())$~', '/b/x/y', $m), $m);
var_dump(preg_match('~(?|(a)|(b))(c)~', 'bc', $m), $m);
var_dump(preg_match('~(?|(?<n>a)|(?<n>b))-(?<z>c)~', 'b-c', $m), $m);
