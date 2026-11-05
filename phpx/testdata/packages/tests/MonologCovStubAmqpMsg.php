<?php
// test double for a class the library type-hints but composer does not install
namespace PhpAmqpLib\Message;

class AMQPMessage
{
    public function __construct(public string $body = '', public array $props = []) {}
}
