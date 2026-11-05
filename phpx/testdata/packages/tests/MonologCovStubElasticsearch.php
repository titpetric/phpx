<?php
// test double for a class the library type-hints but composer does not install
namespace Elasticsearch;

class Client
{
    const VERSION = "7.17.0";
    public array $bulks = [];
    public array $response = ["errors" => false];
    public function bulk(array $p) { $this->bulks[] = $p; return $this->response; }
}
