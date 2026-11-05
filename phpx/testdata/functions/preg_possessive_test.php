<?php
var_dump(preg_match('/a\w*+b/', 'axxb'), preg_match('/^[a-z]++$/', 'abc'), preg_match('/(?:ab)?+c/', 'abc'), preg_match('/\++/', '++'), preg_match('/[+]+x/', '++x'), preg_match('/x{2}+y/', 'xxy'), preg_match('/(?:[.:-]\w+)*+(?!::)/', 'a.b'));
var_dump(preg_match('/^\w++(?:\.\w++)*+$/', 'foo.bar.baz'), preg_match('/a++a/', 'aaa'));
