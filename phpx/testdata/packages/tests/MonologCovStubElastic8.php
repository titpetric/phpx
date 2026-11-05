<?php
// test double for a class the library type-hints but composer does not install
namespace Elastic\Elasticsearch;

class Client
{
    public array $bulks = [];
    public function bulk(array $p) { $this->bulks[] = $p; return ["errors" => false]; }
}
