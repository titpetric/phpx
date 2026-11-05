<?php
namespace A {
    use const NAN as PHP_NAN;
    const NAN = PHP_NAN;
    echo is_nan(NAN) ? "nan\n" : "no\n";
}

namespace B {
    class O { function m($a) { return $a * 2; } static function s($a) { return $a + 1; } }
    class Q extends \ArrayObject {}
    class R implements \ArrayAccess {
        function __construct(public $d) {}
        function offsetExists($o): bool { return isset($this->d[$o]); }
        function offsetGet($o): mixed { return $this->d[$o]; }
        function offsetSet($o, $v): void { $this->d[$o] = $v; }
        function offsetUnset($o): void {}
    }

    $cb = [new O, 'm'];
    $c = $cb(...);
    echo $c(4), "\n";
    $st = ['B\O', 's'];
    $s = $st(...);
    echo $s(4), "\n";
    $r = new \ReflectionFunction(function (array $x) {});
    echo $r->isAnonymous() ? "anon\n" : "named\n";
    $r = new \ReflectionFunction('strlen');
    echo $r->isAnonymous() ? "anon\n" : "named\n";

    echo strstr("user@host", "@", true), "|", strstr("user@host", "@"), "|", var_export(strstr("a", "b"), true), "\n";
    echo in_array('zlib.*', stream_get_filters()) ? "filters\n" : "none\n";

    [$x, $y] = new R([1, 2]);
    echo $x, $y, "\n";
    [$x, $y] = new Q([3, 4]);
    echo $x, $y, "\n";
    [$a, $b] = new R([new \stdClass, 5]);
    echo get_class($a), $b, "\n";
}
