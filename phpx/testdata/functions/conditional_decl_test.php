<?php
if (!function_exists('plain')) {
    function plain() { return "plain"; }
}
echo plain(), "\n";
echo function_exists('later') ? "yes" : "no", "\n";
if (true) {
    function later() { return "later"; }
}
echo function_exists('later') ? "yes" : "no", "\n";
echo later(), "\n";
function outer() {
    function inner() { return "inner"; }
}
echo function_exists('inner') ? "yes" : "no", "\n";
outer();
echo inner(), "\n";
require 'conditional_decl_ns.inc';
echo \Lib\twice(21), "\n";
