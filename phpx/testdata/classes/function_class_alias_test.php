<?php
namespace App;
use Foo\Support\Config;
function config($k) { return "c:$k"; }
echo config('a'), "\n";
echo Config::class, "\n";
