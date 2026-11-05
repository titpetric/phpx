<?php
use PHPUnit\Framework\TestCase;
use Dotenv\Dotenv;
use Dotenv\Parser\Parser;

final class DotenvEntryTest extends TestCase
{
    public function testParse(): void
    {
        $vars = Dotenv::parse("A=1\nB=\"two words\"\nC='single'\n# comment\nD=\${A}x\nE=\n");
        $this->assertSame(['A' => '1', 'B' => 'two words', 'C' => 'single', 'D' => '1x', 'E' => ''], $vars);
    }
    public function testExport(): void
    {
        $vars = Dotenv::parse("export F=f\nG=\"line\\nbreak\"\n");
        $this->assertSame('f', $vars['F']);
        $this->assertSame("line\nbreak", $vars['G']);
    }
}
