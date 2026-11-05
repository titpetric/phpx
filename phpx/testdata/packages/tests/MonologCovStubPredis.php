<?php
// test double for a class the library type-hints but composer does not install
namespace Predis;

class Client
{
    public array $calls = [];
    public function rpush($k, $v) { $this->calls[] = ['rpush', $k, $v]; }
    public function publish($k, $v) { $this->calls[] = ['publish', $k, $v]; }
    public function ltrim($k, $a, $b) { $this->calls[] = ['ltrim', $k, $a, $b]; }
    public function transaction(callable $cb) { $this->calls[] = ['transaction']; $cb($this); }
}
