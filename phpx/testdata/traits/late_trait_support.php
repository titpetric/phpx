<?php
trait LateGreeter { public function greet() { return 'hello from ' . static::class; } }
