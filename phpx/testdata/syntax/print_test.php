<?php
print "a\n";
$r = print "b\n";
echo $r, "\n";
if (print("c\n")) { echo "d\n"; }
echo 1 . print "x";
echo "\n";
print 'concat' . ' ' . 'parts' . PHP_EOL;
class O { public function print(string $s): string { return "<$s>"; } }
echo (new O)->print('m'), "\n";
