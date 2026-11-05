name: SplFileObject iterates and reads csv as php does
description: >
  Plain iteration yields the empty last line, DROP_NEW_LINE with SKIP_EMPTY and READ_AHEAD does not, READ_CSV counts the blank lines it skips, and SplTempFileObject writes to memory. Expected output is php 8.4's.
runner:
  flatstack: false
---
<?php
$p = 'spl_file_object_tmp.txt'; file_put_contents($p, "a,b\n1,2\n\n3,4\n");
$f = new SplFileObject($p);
foreach ($f as $n => $l) { echo $n, '=', json_encode($l), ' '; } echo "\n";
$f->setFlags(SplFileObject::DROP_NEW_LINE | SplFileObject::SKIP_EMPTY | SplFileObject::READ_AHEAD);
foreach ($f as $n => $l) { echo $n, '=', json_encode($l), ' '; } echo "\n";
$f->setFlags(SplFileObject::READ_CSV | SplFileObject::SKIP_EMPTY | SplFileObject::READ_AHEAD | SplFileObject::DROP_NEW_LINE);
foreach ($f as $n => $l) { echo $n, '=', json_encode($l), ' '; } echo "\n";
$f->setFlags(0); $f->rewind(); echo json_encode($f->fgets()), json_encode($f->fgetcsv()), $f->key(), "\n";
$f->seek(3); echo json_encode($f->current()), "\n";
$t = new SplTempFileObject(); $t->fputcsv(['x', 'y z', 'q"r']); $t->fwrite("tail\n"); $t->rewind(); echo json_encode($t->fgets()), json_encode($t->fgets()), "\n";
echo $f->getFilename() === basename($p) ? 'name' : 'bad', ' ', SplFileObject::READ_CSV, "\n";
unlink($p);
---
0="a,b\n" 1="1,2\n" 2="\n" 3="3,4\n" 4="" 
0="a,b" 1="1,2" 2="3,4" 
0=["a","b"] 1=["1","2"] 3=["3","4"] 
"a,b\n"["1","2"]2
"3,4\n"
"x,\"y z\",\"q\"\"r\"\n""tail\n"
name 8
