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
