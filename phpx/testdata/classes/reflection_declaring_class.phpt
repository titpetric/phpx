name: a reflected property names the class that declared it
description: >
  getProperties() on a child lists the child's own properties first and the
  inherited ones after it, a parent's private property is not listed at all,
  and every ReflectionProperty answers the declaring class in
  getDeclaringClass() rather than the class it was read from. Expected output
  is php 8.4's.
---
<?php
class Base {
    protected int $depth = 0;
    public string $name = 'b';
    private string $secret = 's';
}
class Child extends Base {
    public string $own = 'c';
}
$r = new ReflectionClass('Child');
foreach ($r->getProperties() as $p) {
    echo $p->getName(), ':', $p->getDeclaringClass()->getName(), ' ';
}
echo "\n";
$p = new ReflectionProperty('Child', 'depth');
echo $p->getDeclaringClass()->getName(), ' ', $p->isProtected() ? 'protected' : 'other', "\n";
---
own:Child depth:Base name:Base 
Base protected
