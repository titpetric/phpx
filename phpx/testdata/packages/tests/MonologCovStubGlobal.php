<?php
// test double for a class the library type-hints but composer does not install
class Redis
{
    const MULTI = 1;
    public array $calls = [];
    public function rpush($k, $v) { $this->calls[] = ['rpush', $k, $v]; return $this; }
    public function multi($mode) { $this->calls[] = ['multi', $mode]; return $this; }
    public function ltrim($k, $a, $b) { $this->calls[] = ['ltrim', $k, $a, $b]; return $this; }
    public function exec() { $this->calls[] = ['exec']; return []; }
    public function publish($k, $v) { $this->calls[] = ['publish', $k, $v]; return 1; }
}
class AMQPExchange
{
    public array $published = [];
    public function publish($data, $key, $flags, $attrs) { $this->published[] = [$data, $key, $flags, $attrs]; }
}
