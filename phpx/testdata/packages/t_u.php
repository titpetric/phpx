<?php
$url = 'https://example.com/a';
var_dump(strcspn($url, '/?#'));
var_dump(preg_replace_callback('%[^:/@?&=#]+%usD', static function ($m) { return urlencode($m[0]); }, $url));
var_dump(array_map('urldecode', parse_url('https://example.com/a')));
