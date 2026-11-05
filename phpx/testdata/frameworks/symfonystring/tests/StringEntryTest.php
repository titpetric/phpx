<?php
use PHPUnit\Framework\TestCase;
use function Symfony\Component\String\u;
use Symfony\Component\String\Inflector\EnglishInflector;

final class StringEntryTest extends TestCase
{
    public function testUnicodeStrings(): void
    {
        $this->assertSame('HELLO', (string) u('hello')->upper());
        $this->assertSame('helloWorld', (string) u('hello_world')->camel());
        $this->assertSame('hello_world', (string) u('helloWorld')->snake());
        $this->assertSame(5, u('héllo')->length());
        $this->assertSame('héllo wor…', (string) u('héllo world')->truncate(10, '…'));
    }
    public function testInflector(): void
    {
        $i = new EnglishInflector();
        $this->assertSame(['box'], $i->singularize('boxes'));
        $this->assertSame(['mice'], $i->pluralize('mouse'));
    }
}
