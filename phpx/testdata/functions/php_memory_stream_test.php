<?php
$h = fopen('php://memory', 'r+b');
fwrite($h, "line one\nline two\n");
rewind($h);
echo fgets($h);
echo stream_get_contents($h);
fclose($h);
$t = fopen('php://temp', 'w+');
fwrite($t, 'abc');
fseek($t, 1);
echo fread($t, 2), ftell($t), "\n";
