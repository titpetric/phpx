<?php
class P { private array $a = []; function f() { return [...$this->a]; } }
class C extends P { protected $a; }
var_dump((new C)->f());
