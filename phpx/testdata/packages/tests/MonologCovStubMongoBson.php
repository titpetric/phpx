<?php
// test double for a class the library type-hints but composer does not install
namespace MongoDB\BSON;

class UTCDateTime { public function __construct(public int $ms = 0) {} }
