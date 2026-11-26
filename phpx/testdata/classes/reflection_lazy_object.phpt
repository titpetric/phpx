name: ReflectionClass::isUninitializedLazyObject answers false for an ordinary object
description: >
  Livewire asks it for every model it hydrates, and phpx's ReflectionClass
  had no such method, so the dashboard stopped with "call to undefined method
  ReflectionClass::isUninitializedLazyObject()". php 8.4 answers true only
  for an object its lazy-object syntax left uninitialised, which this
  runtime never builds. Expected output is php 8.4's.
runner:
  flatstack: false
---
<?php
class Plain {}
$r = new ReflectionClass('Plain');
var_dump($r->isUninitializedLazyObject(new Plain));
$m = new ReflectionMethod('ReflectionClass', 'isUninitializedLazyObject');
var_dump($m->getName());
---
bool(false)
string(25) "isUninitializedLazyObject"
