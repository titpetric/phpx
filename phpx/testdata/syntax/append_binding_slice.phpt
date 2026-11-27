name: an append into an array a binding answered grows that array
description: >
  An array literal keeps the list a binding answered, and appending through an
  element of it must grow the array and leave the value it was built from alone.
  array_values(array_filter(...)) answers a Go slice, and Laravel's
  Kernel::appendMiddlewareToGroup appends to exactly that at boot. Expected
  output is php 8.4's. Limit: the bytecode engine refuses to append to a slice a
  binding answered, so this runs on the interpreter alone.
runner:
  flatstack: false
---
<?php
// A binding answers a list, php an array: appending through an element the
// literal holds must grow the array and leave the source alone.
class Groups
{
    public $groups;

    public function __construct()
    {
        $this->groups = ['web' => array_values(array_filter(['csrf', 'session']))];
    }

    public function add(string $middleware): void
    {
        $this->groups['web'][] = $middleware;
    }
}

$groups = new Groups();
$groups->add('bindings');
$groups->add('auth');
echo implode(',', $groups->groups['web']), "\n";

$source = array_values(['x']);
$holder = ['list' => $source];
$holder['list'][] = 'y';
echo implode(',', $source), "\n";
echo implode(',', $holder['list']), "\n";
---
csrf,session,bindings,auth
x
x,y
