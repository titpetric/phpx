<?php
$it = new RecursiveDirectoryIterator('/tmp/fx', FilesystemIterator::SKIP_DOTS);
foreach (new RecursiveIteratorIterator($it, RecursiveIteratorIterator::SELF_FIRST) as $k => $f) { echo $k, ' ', get_class($f), ' ', $f->getFilename(), "\n"; }
$f = (new Symfony\Component\Finder\Finder)->files()->in('/tmp/fx');
echo count(iterator_to_array($f)), "\n";
