<?php
// phpx blocker: the braced namespace form ("namespace {") fails to parse ("expected name, got {").
namespace {
    use PHPUnit\Framework\TestCase;

    final class MonologCovNsTest extends TestCase
    {
        public function testBracedGlobalNamespace(): void
        {
            $this->assertTrue(true);
        }
    }
}
