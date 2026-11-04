<?php
interface A {}
interface B extends A {}
interface S { public function __toString(): string; public function with(string $u, ?string $p = null): static; }
class C implements B {}
class D extends C {}
class T { public function __toString(): string { return 't'; } }
var_dump(is_subclass_of('B', 'A'));
var_dump(is_subclass_of('C', 'A'));
var_dump(is_subclass_of('D', 'C'));
var_dump(is_subclass_of(new D, 'A'));
var_dump(is_subclass_of('C', 'C'));
var_dump(is_subclass_of('A', 'B'));
var_dump(is_subclass_of('S', 'Stringable'));
var_dump(is_subclass_of('T', 'Stringable'));
var_dump(is_subclass_of('C', 'D'));
var_dump(interface_exists('Throwable'), interface_exists('Stringable'), interface_exists('Countable'));
$p = (new ReflectionMethod('S', 'with'))->getParameters();
var_dump($p[1]->isOptional(), $p[1]->getDefaultValue());
var_dump(gettype(fopen('php://memory', 'r')));
