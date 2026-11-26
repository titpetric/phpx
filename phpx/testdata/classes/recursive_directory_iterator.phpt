name: RecursiveIteratorIterator starts on first valid and walks a directory like php
description: >
  A fresh RecursiveIteratorIterator answers valid() true and current() with the
  first leaf, which is the shape urlhub's helper loader reads in a
  `while ($it->valid())` loop, and isDot() forwards to the RecursiveDirectoryIterator.
  Expected output is php 8.4's.
runner:
  flatstack: false
---
<?php
$d = 'spl_rdi_tree';
@mkdir($d); @mkdir("$d/sub");
file_put_contents("$d/GoodOne.php", '<?php');
file_put_contents("$d/bad.php", '<?php');
file_put_contents("$d/sub/GoodTwo.php", '<?php');
$fresh = new RecursiveIteratorIterator(new RecursiveDirectoryIterator($d, FilesystemIterator::SKIP_DOTS));
var_dump($fresh->valid(), $fresh->current()->getFilename(), $fresh->isDot());
$rit = new RecursiveIteratorIterator(new RecursiveDirectoryIterator($d, FilesystemIterator::SKIP_DOTS));
$rit->rewind();
$names = [];
while ($rit->valid()) {
    if (!$rit->isDot() && $rit->isFile() && $rit->current()->getExtension() === 'php' && strpos($rit->current()->getFilename(), 'Good') !== false) {
        $names[] = $rit->current()->getFilename();
    }
    $rit->next();
}
sort($names);
echo implode(',', $names), "\n";
unlink("$d/sub/GoodTwo.php"); unlink("$d/GoodOne.php"); unlink("$d/bad.php");
rmdir("$d/sub"); rmdir($d);
---
bool(true)
string(11) "GoodOne.php"
bool(false)
GoodOne.php,GoodTwo.php
