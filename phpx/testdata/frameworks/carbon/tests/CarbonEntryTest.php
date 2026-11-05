<?php
use PHPUnit\Framework\TestCase;
use Carbon\Carbon;
use Carbon\CarbonImmutable;

final class CarbonEntryTest extends TestCase
{
    public function testCreateAndFormat(): void
    {
        $d = Carbon::create(2020, 3, 15, 10, 30, 0, 'UTC');
        $this->assertSame('2020-03-15 10:30:00', $d->format('Y-m-d H:i:s'));
        $this->assertSame('Sunday', $d->dayName);
        $this->assertSame(3, $d->month);
    }
    public function testArithmetic(): void
    {
        $d = CarbonImmutable::create(2020, 1, 31, 0, 0, 0, 'UTC');
        $this->assertSame('2020-02-29', $d->addMonthNoOverflow()->format('Y-m-d'));
        $this->assertSame('2020-02-07', $d->addWeek()->format('Y-m-d'));
        $this->assertSame('2020-01-31', $d->format('Y-m-d'));
        $this->assertSame('2019-12-31', $d->subMonth()->format('Y-m-d'));
    }
    public function testDiff(): void
    {
        $a = Carbon::create(2020, 1, 1, 0, 0, 0, 'UTC');
        $b = Carbon::create(2020, 3, 1, 12, 0, 0, 'UTC');
        $this->assertSame(60, (int) $a->diffInDays($b));
        $this->assertSame(true, $a->lessThan($b));
        $this->assertSame('2 months before', $b->diffForHumans($a, ['syntax' => Carbon::DIFF_RELATIVE_TO_OTHER]) === '2 months after' ? '2 months before' : 'x');
    }
    public function testParse(): void
    {
        $d = Carbon::parse('2021-06-05 08:09:10', 'UTC');
        $this->assertSame(1622880550, $d->getTimestamp());
        $this->assertSame('2021-06-01', $d->startOfMonth()->toDateString());
        $this->assertSame('2021-06-30', $d->endOfMonth()->toDateString());
    }
}
