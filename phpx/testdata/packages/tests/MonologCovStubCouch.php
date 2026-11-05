<?php
// test double for a class the library type-hints but composer does not install
namespace Doctrine\CouchDB;

class CouchDBClient { public array $docs = []; public function postDocument($d) { $this->docs[] = $d; } }
