<?php
use PHPUnit\Framework\TestCase;

// phpx blocker: a class declared inside an if block is never declared.
if (!class_exists('MonologCovCondStub', false)) {
    class MonologCovCondStub { public function hi(): string { return 'hi'; } }
}

final class MonologCovCondTest extends TestCase
{
    public function testConditionalClassDeclaration(): void
    {
        $this->assertSame('hi', (new MonologCovCondStub())->hi());
    }
}
