<?php
// test double for a class the library type-hints but composer does not install
namespace MongoDB;

class Collection { public array $docs = []; public function insertOne($d) { $this->docs[] = $d; } }

class Client { public ?Collection $last = null; public function selectCollection($db, $c) { return $this->last = new Collection(); } }
