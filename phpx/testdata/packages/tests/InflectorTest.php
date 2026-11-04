<?php
use Doctrine\Inflector\{InflectorFactory, Language};
use PHPUnit\Framework\Attributes\DataProvider;
use PHPUnit\Framework\TestCase;

final class InflectorTest extends TestCase
{
    public static function plurals(): array
    {
        return [['user', 'users'], ['box', 'boxes'], ['city', 'cities'], ['person', 'people'], ['child', 'children'], ['mouse', 'mice'], ['bus', 'buses']];
    }
    #[DataProvider('plurals')]
    public function testPluralizeAndBack(string $s, string $p): void
    {
        $i = InflectorFactory::create()->build();
        $this->assertSame($p, $i->pluralize($s));
        $this->assertSame($s, $i->singularize($p));
    }
    public function testCase(): void
    {
        $i = InflectorFactory::create()->build();
        $this->assertSame('FooBarBaz', $i->classify('foo_bar_baz'));
        $this->assertSame('fooBarBaz', $i->camelize('foo_bar_baz'));
        $this->assertSame('foo_bar', $i->tableize('FooBar'));
        $this->assertSame('foo-bar-baz', $i->urlize('Foo Bar Baz'));
        $this->assertSame('Foo Bar', $i->capitalize('foo bar'));
    }
    public function testLanguage(): void
    {
        $this->assertSame('spanish', Language::SPANISH);
        $this->assertSame('boxes', InflectorFactory::createForLanguage(Language::ENGLISH)->build()->pluralize('box'));
    }
}
