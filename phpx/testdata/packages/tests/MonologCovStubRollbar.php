<?php
// test double for a class the library type-hints but composer does not install
namespace Rollbar;

class RollbarLogger
{
    public array $logs = [];
    public int $flushes = 0;
    public function log($level, $toLog, $context = []) { $this->logs[] = [$level, $toLog, $context]; }
    public function flush() { $this->flushes++; }
}
