<?php
// test double for a class the library type-hints but composer does not install
namespace Aws\DynamoDb;

class Marshaler
{
    public function marshalItem(array $item) { return array_map(fn($v) => ['S' => (string) json_encode($v)], $item); }
}

class DynamoDbClient
{
    public array $items = [];
    public function putItem(array $args) { $this->items[] = $args; }
}
