<?php
use PHPUnit\Framework\TestCase;

enum Status: string { case Open = 'open'; case Done = 'done';
    public function label(): string { return match ($this) { self::Open => 'Still open', self::Done => 'Finished' }; }
    public static function fromLabel(string $l): self { foreach (self::cases() as $c) { if ($c->label() === $l) { return $c; } } throw new ValueError('no'); } }
enum Suit { case Hearts; case Spades; }

final class EnumMatchTest extends TestCase
{
    public function testBacked(): void
    {
        $this->assertSame('open', Status::Open->value);
        $this->assertSame(Status::Done, Status::from('done'));
        $this->assertNull(Status::tryFrom('x'));
        $this->assertSame('Finished', Status::Done->label());
        $this->assertSame(Status::Open, Status::fromLabel('Still open'));
        $this->assertCount(2, Status::cases());
    }
    public function testPure(): void
    {
        $this->assertSame('Hearts', Suit::Hearts->name);
        $this->assertNotSame(Suit::Hearts, Suit::Spades);
        $this->assertInstanceOf(UnitEnum::class, Suit::Hearts);
    }
    public function testInvalid(): void
    {
        $this->expectException(ValueError::class);
        Status::from('nope');
    }
    public function testMatch(): void
    {
        $r = fn($n) => match (true) { $n < 0 => 'neg', $n === 0 => 'zero', default => 'pos' };
        $this->assertSame(['neg', 'zero', 'pos'], [$r(-1), $r(0), $r(5)]);
        $this->expectException(UnhandledMatchError::class);
        echo match (99) { 1 => 'a' };
    }
    public function testNullsafeAndSpread(): void
    {
        $o = null;
        $this->assertNull($o?->x?->y);
        $f = fn(int ...$n) => array_sum($n);
        $this->assertSame(6, $f(...[1, 2, 3]));
        $this->assertSame(3, $f(...['a' => 1, 'b' => 2]));
        ['p' => $p, 'q' => [$q]] = ['p' => 1, 'q' => [2]];
        $this->assertSame(3, $p + $q);
    }
}
