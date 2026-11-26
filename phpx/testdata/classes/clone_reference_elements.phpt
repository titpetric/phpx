name: clone gives a copy its own arrays and every node its own data
description: >
  clone copies an object's array property by value, including one whose element
  was made a reference by a setter that has already returned, so a write
  through the copy leaves the original alone. A class that declares __clone
  gets it on the copy, and cloning its object property gives the copy an
  instance of its own; a subtree written through the copy stays out of the
  original. Expected output is php 8.4's. The bytecode engine declines a
  program with a reference outside `foreach`, so its column skips.
runner:
  flatstack: false
---
<?php
class Bag {
    public array $values = ['attributes' => []];
    public function set($key, $value) {
        $current = &$this->values;
        $parts = explode('/', $key);
        $last = array_pop($parts);
        foreach ($parts as $part) {
            if (!isset($current[$part])) { $current[$part] = []; }
            $current = &$current[$part];
        }
        $current[$last] = $value;
    }
    public function get($key) {
        $current = $this->values;
        foreach (explode('/', $key) as $part) { $current = $current[$part]; }
        return $current;
    }
}
$a = new Bag;
$a->set('attributes/class', 'original');
$b = clone $a;
$b->set('attributes/class', 'copy');
echo $a->get('attributes/class'), ' ', $b->get('attributes/class'), "\n";

class Nodey {
    public $data;
    public $child = null;
    public function __construct() { $this->data = new Bag; }
    public function __clone() {
        $children = $this->children();
        $this->child = null;
        foreach ($children as $child) { $this->child = clone $child; }
        $this->data = clone $this->data;
    }
    public function children() { return $this->child === null ? [] : [$this->child]; }
}
$n = new Nodey;
$n->child = new Nodey;
$n->data->set('attributes/class', 'root');
$n->child->data->set('attributes/class', 'leaf');
$copy = clone $n;
$copy->data->set('attributes/class', 'copy-root');
$copy->child->data->set('attributes/class', 'copy-leaf');
echo $n->data->get('attributes/class'), ' ', $n->child->data->get('attributes/class'), "\n";
echo $copy->data->get('attributes/class'), ' ', $copy->child->data->get('attributes/class'), "\n";
---
original copy
root leaf
copy-root copy-leaf
