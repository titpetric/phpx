name: a class declared inside a block is declared when the statement runs
description: >
  if (!class_exists(...)) { class X {} } and a class in a function body: hoist reads the top of the file, and php declares the rest when it reaches them. The loader registered as 'Foo\\Thing::load' needs the class to exist. Expected output is php 8.4's.
runner:
  flatstack: false
---
<?php
namespace Foo;
echo "start ", var_export(\class_exists('Foo\Thing', false), true), "\n";
if (\class_exists('Foo\Thing', false) === false) {
    final class Thing { public static function load($c) { echo "load $c\n"; } }
    echo "declared ", var_export(\class_exists('Foo\Thing', false), true), "\n";
    \spl_autoload_register(__NAMESPACE__ . '\Thing::load');
}
echo "after ", var_export(\class_exists('Foo\Thing', false), true), "\n";
var_dump(class_exists('Foo\Missing'));
if (!function_exists('Foo\helper')) { function helper() { return 1; } }
echo helper(), "\n";
function make() { if (true) { class InFunction { public $v = 5; } } return new InFunction; }
echo make()->v, "\n";
if (false) { class Nothing {} } else { class Chosen extends \ArrayObject {} }
echo class_exists('Foo\Nothing', false) ? 'nothing' : 'no', ' ', class_exists('Foo\Chosen', false) ? 'chosen' : 'no', "\n";
---
start false
declared true
after true
load Foo\Missing
bool(false)
1
5
no chosen
