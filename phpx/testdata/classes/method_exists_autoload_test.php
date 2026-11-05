<?php
spl_autoload_register(function ($c) {
    echo "load $c\n";
    if ($c === 'AutoloadProbe') { require 'autoload_probe_support.php'; }
});
var_dump(method_exists('AutoloadProbe', 'hello'));
var_dump(method_exists('AutoloadProbe', 'nope'));
var_dump(method_exists('Missing', 'x'));
$r = new ReflectionFunction(Closure::fromCallable('strtoupper'));
echo $r->getName(), " ", $r->getNumberOfParameters() >= 1 ? 'params' : 'none', "\n";
var_dump($r->getClosureThis());
