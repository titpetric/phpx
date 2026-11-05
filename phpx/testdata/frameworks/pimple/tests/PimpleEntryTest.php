<?php
use PHPUnit\Framework\TestCase;
use Pimple\Container;

final class PimpleEntryTest extends TestCase
{
    public function testServicesAndParams(): void
    {
        $c = new Container();
        $c['name'] = 'world';
        $c['greeter'] = function ($c) { return new class($c['name']) { public function __construct(public string $n) {} public function hi() { return "hi {$this->n}"; } }; };
        $this->assertSame('hi world', $c['greeter']->hi());
        $this->assertSame($c['greeter'], $c['greeter']);
    }
    public function testFactoryAndExtend(): void
    {
        $c = new Container();
        $c['n'] = 0;
        $c['obj'] = $c->factory(function () { return new \stdClass; });
        $this->assertNotSame($c['obj'], $c['obj']);
        $c['v'] = 2;
        $c['sq'] = fn($c) => $c['v'] ** 2;
        $c->extend('sq', fn($sq, $c) => $sq + 1);
        $this->assertSame(5, $c['sq']);
    }
}
