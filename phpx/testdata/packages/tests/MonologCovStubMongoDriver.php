<?php
// test double for a class the library type-hints but composer does not install
namespace MongoDB\Driver;

class BulkWrite { public array $docs = []; public function insert($d) { $this->docs[] = $d; } }

class Manager { public array $writes = []; public function executeBulkWrite($ns, $bulk) { $this->writes[] = [$ns, $bulk->docs]; } }
