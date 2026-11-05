<?php
// test double for a class the library type-hints but composer does not install
namespace PhpAmqpLib\Channel;

class AMQPChannel
{
    public array $published = [];
    public array $batch = [];
    public int $flushed = 0;
    public function basic_publish($msg, $exchange, $key) { $this->published[] = [$msg, $exchange, $key]; }
    public function batch_basic_publish($msg, $exchange, $key) { $this->batch[] = [$msg, $exchange, $key]; }
    public function publish_batch() { $this->flushed++; }
}
