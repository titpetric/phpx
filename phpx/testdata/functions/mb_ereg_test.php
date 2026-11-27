<?php
// mbstring's regex calls, as urlhub registers `mb_ereg` as a sqlite REGEXP
// function. The ASCII patterns match as php matches them; the runtime answers
// over Go's regexp, so a backreference or a look-around answers false where
// Oniguruma would read one.
var_dump(mb_regex_encoding());
var_dump(mb_regex_encoding('UTF-8'));
var_dump(mb_ereg('^[a-zA-Z]+$', 'abc'));
var_dump(mb_ereg('^[a-zA-Z]+$', 'ab1'));
var_dump(mb_eregi('^[a-z]+$', 'ABC'));
var_dump(mb_ereg('^[0-9]{3}$', '123'));
