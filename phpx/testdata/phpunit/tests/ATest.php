<?php
use PHPUnit\Framework\TestCase;
final class ATest extends TestCase {
    public function testAdd(): void { $this->assertSame(4, 2 + 2); }
}
