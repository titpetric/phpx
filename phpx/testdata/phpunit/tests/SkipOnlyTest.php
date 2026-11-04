<?php
use PHPUnit\Framework\TestCase;
final class SkipOnlyTest extends TestCase
{
    public function testSkipped(): void { $this->markTestSkipped('not today'); }
    public function testIncomplete(): void { $this->markTestIncomplete('later'); }
}
