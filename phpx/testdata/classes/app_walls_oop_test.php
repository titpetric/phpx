<?php
namespace W {
    enum Suit: string { case Hearts = 'H'; case Spades = 'S'; }

    // An object that declares only __call is callable as a method but not as a
    // function: is_callable($proxy) is false and $proxy() raises.
    class Proxy {
        public array $calls = [];
        public function __call($name, $args) {
            $this->calls[] = $name . '(' . implode(',', $args) . ')';
            return $name;
        }
    }

    // The container an app reads a nested write through: offsetGet answers an
    // object, and the write lands in that object, not in the container.
    class Config implements \ArrayAccess {
        public array $bags = [];
        public function __construct() { $this->bags['app'] = new Bag(['name' => 'demo']); }
        public function offsetExists($o): bool { return isset($this->bags[$o]); }
        public function offsetGet($o): mixed { return $this->bags[$o]; }
        public function offsetSet($o, $v): void { $this->bags[$o] = $v; }
        public function offsetUnset($o): void {}
    }
    class Bag implements \ArrayAccess {
        public function __construct(public array $data) {}
        public function offsetExists($o): bool { return isset($this->data[$o]); }
        public function offsetGet($o): mixed { return $this->data[$o] ?? null; }
        public function offsetSet($o, $v): void { $this->data[$o] = $v; }
        public function offsetUnset($o): void { unset($this->data[$o]); }
    }
    class Container implements \ArrayAccess {
        public array $items = [];
        public function offsetExists($o): bool { return isset($this->items[$o]); }
        public function offsetGet($o): mixed { return $this->items[$o]; }
        public function offsetSet($o, $v): void { $this->items[$o] = $v; }
        public function offsetUnset($o): void { unset($this->items[$o]); }
    }

    $p = new Proxy();
    echo 'is_callable=', is_callable($p) ? 'yes' : 'no', "\n";
    echo 'method=', $p->hello('a', 'b'), "\n";
    echo 'calls=', implode('|', $p->calls), "\n";
    try { $p(); } catch (\Error $e) { echo 'call error=', $e->getMessage(), "\n"; }

    $config = new Config();
    $config['app']['name'] = 'changed';
    echo 'nested write=', $config['app']['name'], "\n";

    $c = new Container();
    echo 'isset one=', isset($c['missing']) ? 'yes' : 'no', "\n";
    echo 'isset chain=', isset($c['missing'], $c['also']) ? 'yes' : 'no', "\n";

    echo 'named defaults=', str_getcsv('a;b', escape: '\\')[0], "\n";
    echo 'round=', round(1.955, precision: 2), "\n";
    echo 'slice=', implode(',', array_slice([1, 2, 3, 4], 1, preserve_keys: true)), "\n";

    $extra = compact('p');
    $merged = ['x' => 1] + $extra;
    echo 'union=', implode(',', array_keys($merged)), "\n";

    echo 'enum_exists=', enum_exists('W\\Suit') ? 'yes' : 'no', ' value=', Suit::from('H')->value, "\n";

    // A closure literal captures what the variable holds where the expression
    // reaches it: the second closure sees what the first argument assigned.
    $f = null;
    $pair = call_user_func(function ($x, $y) { return [$x(), $y()]; },
        $f = fn () => 'one', fn () => $f());
    echo 'closure chain=', implode(',', $pair), "\n";
}
