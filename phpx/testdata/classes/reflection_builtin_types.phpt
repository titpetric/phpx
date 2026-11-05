name: reflection and class_exists know the engine's own types
description: >
  ReflectionClass names Closure, Generator and the core interfaces, reports them internal, and class_exists is true for Closure and Generator. Expected output is php 8.4's.
runner:
  flatstack: false
---
<?php
foreach (['Closure', 'stdClass', 'Generator', 'ArrayObject', 'Exception', 'DateTime', 'Countable', 'Traversable', 'Iterator', 'WeakMap', 'Stringable', 'Throwable', 'JsonSerializable', 'ArrayAccess', 'SplStack', 'Error', 'TypeError', 'DateTimeImmutable', 'DateTimeInterface', 'UnitEnum', 'SplObjectStorage'] as $c) {
  try { $r = new ReflectionClass($c); echo $c, ':', $r->getName(), $r->isInterface() ? ' iface' : '', $r->isInternal() ? ' internal' : '', "\n"; } catch (Throwable $e) { echo $c, ' ERR ', $e->getMessage(), "\n"; }
}
var_dump(class_exists('Closure'), class_exists('Generator'), class_exists('stdClass'), interface_exists('Traversable'), class_exists('ArrayObject'), class_exists('WeakMap'));
$f = function () {}; var_dump($f instanceof Closure, get_class($f), (new ReflectionObject($f))->getName());
---
Closure:Closure internal
stdClass:stdClass internal
Generator:Generator internal
ArrayObject:ArrayObject internal
Exception:Exception internal
DateTime:DateTime internal
Countable:Countable iface internal
Traversable:Traversable iface internal
Iterator:Iterator iface internal
WeakMap:WeakMap internal
Stringable:Stringable iface internal
Throwable:Throwable iface internal
JsonSerializable:JsonSerializable iface internal
ArrayAccess:ArrayAccess iface internal
SplStack:SplStack internal
Error:Error internal
TypeError:TypeError internal
DateTimeImmutable:DateTimeImmutable internal
DateTimeInterface:DateTimeInterface iface internal
UnitEnum:UnitEnum iface internal
SplObjectStorage:SplObjectStorage internal
bool(true)
bool(true)
bool(true)
bool(true)
bool(true)
bool(true)
bool(true)
string(7) "Closure"
string(7) "Closure"
