<?php
$s = 'helloWorld';
var_dump(preg_replace('/([\p{Lu}]+)([\p{Lu}][\p{Ll}])/u', '\1_\2', $s));
var_dump(preg_replace('/([\p{Ll}0-9])([\p{Lu}])/u', '\1_\2', $s));
var_dump(preg_replace('/(?<=\p{Ll})\p{Lu}/u', '_$0', $s));
var_dump(mb_strtolower(preg_replace(['/(\p{Lu}+)(\p{Lu}\p{Ll})/u', '/([\p{Ll}0-9])(\p{Lu})/u'], '\1_\2', 'helloWorld'), 'UTF-8'));
var_dump(preg_replace('/\p{Lu}/u', '[$0]', 'aBcD'));
