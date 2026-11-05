<?php
// test double for a class the library type-hints but composer does not install
namespace Aws\Sqs;

class SqsClient
{
    public array $sent = [];
    public function sendMessage(array $args) { $this->sent[] = $args; }
}
