<?php
use PHPUnit\Framework\TestCase;
use Symfony\Component\Yaml\Yaml;

final class YamlEntryTest extends TestCase
{
    public function testParseScalarsAndMaps(): void
    {
        $this->assertSame(['a' => 1, 'b' => 'two', 'c' => true, 'd' => null, 'e' => 1.5], Yaml::parse("a: 1\nb: two\nc: true\nd: ~\ne: 1.5\n"));
    }
    public function testParseNested(): void
    {
        $y = "list:\n  - x\n  - y\nmap:\n  k: v\n  n: [1, 2, 3]\n";
        $this->assertSame(['list' => ['x', 'y'], 'map' => ['k' => 'v', 'n' => [1, 2, 3]]], Yaml::parse($y));
    }
    public function testMultilineStrings(): void
    {
        $this->assertSame(['t' => "a\nb\n"], Yaml::parse("t: |\n  a\n  b\n"));
        $this->assertSame(['t' => 'a b'], Yaml::parse("t: >-\n  a\n  b\n"));
    }
    public function testDump(): void
    {
        $this->assertSame("a: 1\nb:\n    - x\n    - z\n", Yaml::dump(['a' => 1, 'b' => ['x', 'z']]));
        $this->assertSame("{ a: 1, b: [x, z] }", Yaml::dump(['a' => 1, 'b' => ['x', 'z']], 0));
    }
    public function testRoundTrip(): void
    {
        $data = ['name' => 'phpx', 'tags' => ['a', 'b'], 'deep' => ['x' => ['y' => 'z']], 'n' => 3];
        $this->assertSame($data, Yaml::parse(Yaml::dump($data, 4, 2)));
    }
}
