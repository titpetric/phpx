<?php
$src = <<<'SRC'
<?php
$a = .5 + 0x1F;
$b = "v $a {$a['k']}";
echo $b; // trailing
SRC;
echo "string:\n", highlight_string($src, true), "\n";
echo "file:\n", highlight_file(__FILE__, true), "\n";
echo "wrote: ";
$ok = highlight_file(__FILE__);
var_dump($ok);
echo "missing: ";
var_dump(@highlight_file(__DIR__ . '/nope.php', true));
