<?php
require_once __DIR__ . '/MonologCovStubClock.php';

use Monolog\{Logger, Level, JsonSerializableDateTimeImmutable};
use Monolog\Handler\TestHandler;
use PHPUnit\Framework\TestCase;

final class MonologCov14Clock implements \Psr\Clock\ClockInterface
{
    public function __construct(private \DateTimeImmutable $at) {}
    public function now(): \DateTimeImmutable { return $this->at; }
}

final class MonologCov14Test extends TestCase
{
    public function testLoggerUsesClock(): void
    {
        $h = new TestHandler();
        $clock = new MonologCov14Clock(new \DateTimeImmutable('2001-02-03 04:05:06.250000', new \DateTimeZone('Europe/Paris')));
        $l = new Logger('clk', [$h], [], new \DateTimeZone('UTC'), $clock);
        $this->assertSame($clock, $l->getClock());
        $l->info('timed');
        $dt = $h->getRecords()[0]->datetime;
        $this->assertInstanceOf(JsonSerializableDateTimeImmutable::class, $dt);
        $this->assertSame('2001-02-03 03:05:06', $dt->format('Y-m-d H:i:s'));
        $this->assertSame('250000', $dt->format('u'));
        $this->assertSame('UTC', $dt->getTimezone()->getName());
    }

    public function testClockReturningMonologDateTime(): void
    {
        $h = new TestHandler();
        $fixed = new JsonSerializableDateTimeImmutable(false, new \DateTimeZone('Asia/Tokyo'));
        $l = new Logger('clk2', [$h]);
        $this->assertSame($l, $l->setClock(new MonologCov14Clock($fixed)));
        $l->info('same instance');
        $this->assertSame($fixed->getTimestamp(), $h->getRecords()[0]->datetime->getTimestamp());
        $this->assertSame('Asia/Tokyo', $h->getRecords()[0]->datetime->getTimezone()->getName());
        $l->setClock(null);
        $this->assertNull($l->getClock());
        $l->info('real clock');
        $this->assertCount(2, $h->getRecords());
    }

    public function testClockWithoutMicroseconds(): void
    {
        $h = new TestHandler();
        $l = new Logger('clk3', [$h], [], null, new MonologCov14Clock(new \DateTimeImmutable('2020-05-06 07:08:09', new \DateTimeZone('UTC'))));
        $l->useMicrosecondTimestamps(false);
        $l->warning('whole second');
        $this->assertSame('2020-05-06T07:08:09+00:00', $h->getRecords()[0]->datetime->format('c'));
    }
}
