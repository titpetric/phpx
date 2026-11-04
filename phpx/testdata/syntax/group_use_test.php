<?php
namespace Main;
require_once 'group_use_lib.inc';
use Lib\Util\{A, B as Bee, C};
use function Lib\Util\{helper};
echo (new A)->n(), (new Bee)->n(), (new C)->n(), helper(), "\n";
