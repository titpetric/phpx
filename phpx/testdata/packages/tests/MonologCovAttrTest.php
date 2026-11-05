<?php
use PHPUnit\Framework\TestCase;

// phpx blocker: an attribute written inline before a parameter fails to parse
// ("expected parameter $var"); the same attribute on its own line parses.
final class MonologCovAttrTest extends TestCase
{
    public static function go(#[\SensitiveParameter] string $v): string { return $v; }

    public function testInlineParameterAttribute(): void
    {
        $this->assertSame('ok', self::go('ok'));
    }
}
