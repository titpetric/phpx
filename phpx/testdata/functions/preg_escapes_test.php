<?php
var_dump(preg_split('/(.*\R)/', "a\nb\n", -1, PREG_SPLIT_DELIM_CAPTURE | PREG_SPLIT_NO_EMPTY));
var_dump(preg_split('/,/', 'a,b,,c', -1, PREG_SPLIT_NO_EMPTY));
var_dump(preg_split('/(-)/', 'a-b', -1, PREG_SPLIT_DELIM_CAPTURE));
var_dump(preg_split('/\s+/', 'a  b c', 2));
var_dump(preg_match('/^a\Rb$/', "a\r\nb"), preg_match('/a\hb/', "a\tb"), preg_match('/[\h]+/', "x \t y", $m), $m);
echo json_encode(preg_split('/\R/', "l1\nl2\r\nl3")), "\n";
