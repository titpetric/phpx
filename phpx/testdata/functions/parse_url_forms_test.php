<?php
foreach (['http://user:pw@example.com:8080/a/b.php?x=1&y=2#frag', 'h:80', 'h:80/p', '//host/path', '/just/path?q', 'mailto:a@b.c', 'file:///etc/passwd', 'http://[::1]:81/x', 'http://[::1]/', 'x://a b/c', 'http:///x', 'http://h:99999/', 'a.b/c', 'localhost:8080', '?q=1', '#f', 'http://user@h', 'data:text/plain,hi', 'http://h?x', 'ws://h:1/p?q#f', 'php://memory', 'C:\\x'] as $u) { echo $u, ' => ', json_encode(parse_url($u)), "\n"; }
var_dump(parse_url('http://h:80/p', PHP_URL_PORT), parse_url('http://h/p', PHP_URL_QUERY), parse_url('http://h/p', PHP_URL_PATH));
