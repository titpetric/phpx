name: clone copies a date object
description: >
  `clone $date` is an independent value: changing the copy leaves the original.
  Expected output is php 8.4's.
runner:
  flatstack: false
---
<?php
$d = new DateTime('2020-01-02 03:04:05');
$e = clone $d; $e->modify('+1 day');
echo $d->format('Y-m-d'), ' ', $e->format('Y-m-d'), "\n";
$i = new DateTimeImmutable('2020-01-02'); $j = clone $i; echo $j->format('Y'), "\n";
$a = new ArrayObject([1,2]); $b = clone $a; $b[] = 3; echo count($a), count($b), "\n";
$s = new SplObjectStorage; $t = clone $s; echo count($t), "\n";
$x = new DateInterval('P1D'); $y = clone $x; echo $y->d, "\n";
$tz = new DateTimeZone('UTC'); $tz2 = clone $tz; echo $tz2->getName(), "\n";
---
2020-01-02 2020-01-03
2020
23
0
1
UTC
