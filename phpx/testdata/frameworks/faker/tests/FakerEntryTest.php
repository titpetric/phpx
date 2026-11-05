<?php
use PHPUnit\Framework\TestCase;

final class FakerEntryTest extends TestCase
{
    public function testSeededDeterminism(): void
    {
        $f = Faker\Factory::create();
        $f->seed(42); $first = [$f->name(), $f->numberBetween(1, 1000), $f->city()];
        $f->seed(42); $second = [$f->name(), $f->numberBetween(1, 1000), $f->city()];
        $this->assertSame($first, $second);
    }
    public function testShapes(): void
    {
        $f = Faker\Factory::create(); $f->seed(1);
        $this->assertMatchesRegularExpression('/^[^@\s]+@[^@\s]+\.[a-z]+$/', $f->safeEmail());
        $this->assertSame(36, strlen($f->uuid()));
        $this->assertMatchesRegularExpression('/^\d{4}-\d{2}-\d{2}$/', $f->date());
        $this->assertCount(3, $f->words(3));
    }
}
