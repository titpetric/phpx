<?php
// test double for a class the library type-hints but composer does not install
namespace Elastica;

class Document
{
    public array $data = [];
    public ?string $index = null;
    public function setData($d) { $this->data = $d; return $this; }
    public function setIndex($i) { $this->index = $i; return $this; }
}

class Client
{
    public array $added = [];
    public ?\Throwable $fail = null;
    public function addDocuments(array $docs) { if ($this->fail) { throw $this->fail; } $this->added[] = $docs; }
}
