<?php

// `??=` assigns a value, and php's array value copies: a write through the
// target does not reach the variable it was read from.
$main = ['a', 'b', 'c'];
$copy = null;
$copy ??= $main;
array_shift($copy);
var_dump($main);

// The same through an array element and a property.
$bag = ['list' => null];
$source = ['x', 'y'];
$bag['list'] ??= $source;
$bag['list'][] = 'z';
var_dump($source);

class Holder
{
    public $items;

    public function fill(array $from): void
    {
        $this->items ??= $from;
        $this->items[] = 'end';
    }
}

$h = new Holder();
$h->fill(['one']);
var_dump($h->items);
