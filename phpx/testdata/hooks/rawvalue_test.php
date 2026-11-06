<?php
class P {
    public string $name = 'a' {
        get => strtoupper($this->name);
        set => trim($value);
    }
}
$p = new P();
$r = new ReflectionProperty(P::class, 'name');
$r->setRawValue($p, '  raw  ');
var_dump($r->getRawValue($p));
var_dump($p->name);
