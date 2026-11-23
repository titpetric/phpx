name: attributes on a property of a class with no methods
description: >
  ReflectionProperty reads the attributes, their named arguments, the default
  value and the type from the declaration. A class that declares only
  properties has no method to take its file from, and an attribute's
  getArguments keeps the names php keeps. A union type answers a
  ReflectionUnionType. Expected output is php 8.4's.
---
<?php
#[Attribute(Attribute::TARGET_PROPERTY)]
class Tag {
    public string $label;
    public int $n;
    public function __construct($label = '', $n = 0) { $this->label = $label; $this->n = $n; }
}
class Holder {
    #[Tag(label: 'x', n: 2)]
    public string $p = 'v';
    #[Tag('y')]
    public int $q = 1;
}
foreach (['p' => 'x', 'q' => 'y'] as $name => $label) {
    $rp = new ReflectionProperty('Holder', $name);
    $attrs = $rp->getAttributes(Tag::class);
    echo $name, ': ', count($attrs), "\n";
    echo 'args: ', json_encode($attrs[0]->getArguments()), "\n";
    $inst = $attrs[0]->newInstance();
    echo 'label: ', $inst->label, ' ', $inst->n, "\n";
    echo 'default: ', var_export($rp->getDefaultValue(), true), "\n";
    echo 'type: ', get_class($rp->getType()), "\n";
}
class U { #[Tag] public bool|string $both = false; }
$t = (new ReflectionProperty('U', 'both'))->getType();
echo get_class($t), ' ', count($t->getTypes()), "\n";
?>
---
p: 1
args: {"label":"x","n":2}
label: x 2
default: 'v'
type: ReflectionNamedType
q: 1
args: ["y"]
label: y 0
default: 1
type: ReflectionNamedType
ReflectionUnionType 2
