<?php
// test double for a class the library type-hints but composer does not install
namespace Psr\Clock;

interface ClockInterface
{
    public function now(): \DateTimeImmutable;
}
