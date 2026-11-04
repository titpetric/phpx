<?php
namespace App;

final class Calc
{
    private array $log = [];
    public function __construct(private int $base = 0) {}
    public function add(int $a, int $b): int { $this->log[] = "add"; return $a + $b + $this->base; }
    public function div(int $a, int $b): float
    {
        if ($b === 0) { throw new \InvalidArgumentException('division by zero'); }
        return $a / $b;
    }
    public function log(): array { return $this->log; }
}
