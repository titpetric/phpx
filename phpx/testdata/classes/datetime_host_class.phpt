name: the date classes as classes
description: >
  ReflectionObject over a date, is_subclass_of and class_implements for DateTimeInterface, and a subclass of DateTimeImmutable that declares no constructor. Expected output is php 8.4's.
runner:
  flatstack: false
---
<?php
var_dump(class_exists('DateTimeImmutable'), class_exists('DateTime'), class_exists('DateTimeZone'), class_exists('DateInterval'), interface_exists('DateTimeInterface'));
$r = new ReflectionClass('DateTimeImmutable'); var_dump($r->getName(), $r->isInternal());
$d = new DateTimeImmutable('2020-01-01'); var_dump($d instanceof DateTimeInterface, $d instanceof DateTimeImmutable, get_class($d), (new ReflectionObject($d))->getName());
var_dump(is_a($d, 'DateTimeInterface'), is_subclass_of($d, 'DateTimeInterface'), in_array('DateTimeInterface', class_implements($d)));
class MyDT extends DateTimeImmutable {} $m = new MyDT('2020-01-01'); var_dump($m instanceof DateTimeImmutable, $m->format('Y'));
---
bool(true)
bool(true)
bool(true)
bool(true)
bool(true)
string(17) "DateTimeImmutable"
bool(true)
bool(true)
bool(true)
string(17) "DateTimeImmutable"
string(17) "DateTimeImmutable"
bool(true)
bool(true)
bool(true)
bool(true)
string(4) "2020"
