<?php
var_dump(strspn('0x7f', '0123456789abcdefABCDEF', 2), strspn('abc', 'ab'), strspn('abcd', 'cd', -2), strspn('abcd', 'abc', 0, 2), strspn('abcd', 'abc', 1, -1), strcspn('abcd', 'c', 1), strcspn('abcd', 'd'), strcspn('abcd', 'x', 5), strcspn('abcd', 'cd', 0, 3));
