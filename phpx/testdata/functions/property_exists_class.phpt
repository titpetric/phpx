name: property_exists reads a class name and a static property
description: >
  php accepts a class name where an object goes and reads the declarations,
  static ones among them. Livewire's Synthesizer::getKey() asks
  property_exists(static::class, 'key') and phpx answered false for every
  class name, so a component that held an array stopped with "You need to
  define static $key property". Expected output is php 8.4's.
runner:
  flatstack: false
---
<?php
class Synth {
    public static $key = 'arr';
    public $path;
    protected $hidden = 1;
    private $secret = 2;
}
$s = new Synth;
var_dump(property_exists('Synth', 'key'));
var_dump(property_exists('Synth', 'path'));
var_dump(property_exists('Synth', 'hidden'));
var_dump(property_exists('Synth', 'secret'));
var_dump(property_exists('Synth', 'nope'));
var_dump(property_exists($s, 'key'));
var_dump(property_exists($s, 'secret'));
var_dump(property_exists($s, 'nope'));
var_dump(property_exists('Synth', 'KEY'));
---
bool(true)
bool(true)
bool(true)
bool(true)
bool(false)
bool(true)
bool(true)
bool(false)
bool(false)
