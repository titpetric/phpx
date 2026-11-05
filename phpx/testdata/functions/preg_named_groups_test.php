<?php
preg_match('/(?<a>x)(?P<b>y)(z)?/', 'xy', $m); print_r($m);
preg_match('/(?<year>\d{4})-(\d\d)-(?<day>\d\d)/', 'on 2024-05-17', $m); print_r($m);
preg_match('/(?<a>x)(?<b>y)?/', 'x', $m, PREG_UNMATCHED_AS_NULL); var_dump($m);
preg_match('/(?<w>\w+)/', 'hello world', $m, PREG_OFFSET_CAPTURE); print_r($m);
preg_match_all('/(?<k>\w)=(?<v>\d)/', 'a=1,b=2', $m); print_r($m);
preg_match_all('/(?<k>\w)=(\d)/', 'a=1,b=2', $m, PREG_SET_ORDER); print_r($m);
echo preg_replace_callback('/(?<n>\d+)/', fn($m) => $m['n'] * 2, 'a1 b22'), "\n";
var_dump(preg_match('/(?<q>["\']).*?\k<q>/', 'say "hi" now', $m), $m['q'], $m[0]);
var_dump(preg_match('/(?P<d>\w)(?P=d)/', 'aab', $m), $m[0]);
var_dump(preg_match("/(?<Text>.+?)??(?<Ind>(?<=\\n|^)[ \\t]+)?((?<Open>\\{)|\$)/xsiAuD", 'Hello {$n}', $m, PREG_UNMATCHED_AS_NULL), array_keys($m));
var_dump(preg_match('/(?J)(?<n>a)|(?<n>b)/', 'b', $m, PREG_UNMATCHED_AS_NULL), $m);
var_dump(preg_match('/(?J)(?<n>a)|(?<n>b)/', 'a', $m), $m);
