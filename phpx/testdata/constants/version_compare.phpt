---
name: version_compare orders versions as php does
description: >
  Framework bootstrap code gates on version_compare(). The expected section is
  what `php` printed.
---
<?php
$p = [["8.4.27","8.4.1"],["1.0","1.0.0"],["1.0rc1","1.0"],["1.0-dev","1.0alpha"],["1.0beta","1.0RC1"],["1.0pl1","1.0"],["5.2","5.10"],["1.0.0","1.0.0"],["1.0a","1.0b"],["1.0-p1","1.0-pl1"],["1","1.0.1"],["1.0foo","1.0"],["", "1"],["1.0.0-beta2","1.0.0-beta10"]];
foreach ($p as $pair) { $a = $pair[0]; $b = $pair[1]; echo "$a $b ", version_compare($a,$b), "\n"; }
foreach (['<','lt','<=','le','>','gt','>=','ge','==','eq','!=','<>','ne'] as $o) { var_dump(version_compare("1.2","1.10",$o)); }
var_dump(version_compare(PHP_VERSION, "8.4.1", ">="));
---
8.4.27 8.4.1 1
1.0 1.0.0 -1
1.0rc1 1.0 -1
1.0-dev 1.0alpha -1
1.0beta 1.0RC1 -1
1.0pl1 1.0 1
5.2 5.10 -1
1.0.0 1.0.0 0
1.0a 1.0b -1
1.0-p1 1.0-pl1 0
1 1.0.1 -1
1.0foo 1.0 -1
 1 -1
1.0.0-beta2 1.0.0-beta10 -1
bool(true)
bool(true)
bool(true)
bool(true)
bool(false)
bool(false)
bool(false)
bool(false)
bool(false)
bool(false)
bool(true)
bool(true)
bool(true)
bool(true)
