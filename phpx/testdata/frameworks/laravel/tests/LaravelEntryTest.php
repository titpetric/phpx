<?php
use PHPUnit\Framework\TestCase;
use Illuminate\Container\Container;
use Illuminate\Support\Collection;
use Illuminate\Support\Arr;
use Illuminate\Support\Str;

interface LaravelEntryGreeter { public function greet(string $n): string; }
final class LaravelEntryEnglish implements LaravelEntryGreeter { public function greet(string $n): string { return "Hello $n"; } }
final class LaravelEntryConsumer { public function __construct(public LaravelEntryGreeter $greeter) {} }

final class LaravelEntryTest extends TestCase
{
    public function testApplicationClassExists(): void
    {
        $this->assertTrue(class_exists('Illuminate\Foundation\Application'));
    }
    public function testContainerBindsAndResolves(): void
    {
        $c = new Container();
        $c->bind(LaravelEntryGreeter::class, LaravelEntryEnglish::class);
        $this->assertSame('Hello Ada', $c->make(LaravelEntryConsumer::class)->greeter->greet('Ada'));
        $c->singleton('shared', fn () => new stdClass);
        $this->assertSame($c->make('shared'), $c->make('shared'));
    }
    public function testCollection(): void
    {
        $c = Collection::make([3, 1, 2])->sort()->values()->map(fn ($n) => $n * 2);
        $this->assertSame([2, 4, 6], $c->all());
        $this->assertSame(12, $c->sum());
        $this->assertSame(['a' => 1], Collection::make(['a' => 1, 'b' => null])->filter()->all());
    }
    public function testStrAndArr(): void
    {
        $this->assertSame('hello-world', Str::slug('Hello World'));
        $this->assertSame('HelloWorld', Str::studly('hello_world'));
        $this->assertSame('hello_world', Str::snake('HelloWorld'));
        $this->assertSame(1, Arr::get(['a' => ['b' => 1]], 'a.b'));
        $this->assertSame(['a' => ['b' => 2]], Arr::undot(['a.b' => 2]));
    }
}
