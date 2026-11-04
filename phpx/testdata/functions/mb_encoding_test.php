<?php
var_dump(mb_detect_encoding('plain', null, true));
var_dump(mb_detect_encoding('café', null, true));
var_dump(mb_detect_encoding("\xff\xfe", ['ASCII', 'UTF-8'], true));
var_dump(mb_detect_encoding('café', 'UTF-8, ISO-8859-1', true));
var_dump(bin2hex(mb_convert_encoding('é', 'ISO-8859-1', 'UTF-8')));
var_dump(mb_convert_encoding("\xe9", 'UTF-8', 'ISO-8859-1'));
var_dump(mb_convert_encoding('same', 'utf8', 'UTF-8'));
print_r(mb_convert_encoding(['a', 'k' => 'é'], 'UTF-8', 'UTF-8'));
