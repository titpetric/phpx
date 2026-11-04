<?php
namespace Lib;

if (!function_exists('Lib\twice')) {
    function twice($v) { return $v * 2; }
}
if (!function_exists('Lib\twice')) {
    function twice($v) { return 'never'; }
}
