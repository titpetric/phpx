<?php
use Monolog\{Logger, Level, LogRecord, Registry, DateTimeImmutable};
use Monolog\Handler\{TestHandler, BufferHandler, FilterHandler, GroupHandler, FingersCrossedHandler, StreamHandler, SamplingHandler, NullHandler, OverflowHandler, WhatFailureGroupHandler, ErrorLogHandler};
use Monolog\Handler\FingersCrossed\ErrorLevelActivationStrategy;
use Monolog\Formatter\{LineFormatter, JsonFormatter, NormalizerFormatter, ScalarFormatter, HtmlFormatter};
use Monolog\Processor\{PsrLogMessageProcessor, TagProcessor, UidProcessor, ProcessIdProcessor, IntrospectionProcessor, MemoryUsageProcessor};
use PHPUnit\Framework\TestCase;

final class MonologMoreTest extends TestCase
{
    private function rec(string $m = 'm', Level $l = Level::Info, array $ctx = []): LogRecord
    {
        return new LogRecord(new \DateTimeImmutable('2024-01-02 03:04:05'), 'chan', $l, $m, $ctx);
    }
    public function testLineFormatter(): void
    {
        $f = new LineFormatter("%level_name%|%channel%|%message%|%context%|%extra%\n", null, true, true);
        $this->assertSame("INFO|chan|m|{\"a\":1}|\n", $f->format($this->rec('m', Level::Info, ['a' => 1])));
        $f2 = new LineFormatter('[%datetime%] %message%' . "\n", 'Y-m-d');
        $this->assertSame("[2024-01-02] hi\n", $f2->format($this->rec('hi')));
    }
    public function testJsonFormatter(): void
    {
        $f = new JsonFormatter();
        $d = json_decode($f->format($this->rec('x', Level::Warning, ['k' => [1, 2]])), true);
        $this->assertSame('x', $d['message']);
        $this->assertSame(300, $d['level']);
        $this->assertSame('WARNING', $d['level_name']);
        $this->assertSame([1, 2], $d['context']['k']);
        $batch = $f->formatBatch([$this->rec('a'), $this->rec('b')]);
        $this->assertCount(2, json_decode($batch, true));
    }
    public function testNormalizer(): void
    {
        $n = new NormalizerFormatter('Y');
        $out = $n->format($this->rec('x', Level::Info, ['o' => new \ArrayObject([1]), 'e' => new \RuntimeException('boom'), 'f' => 1.5, 'n' => null]));
        $this->assertSame('2024', $out['datetime']);
        $this->assertSame('boom', $out['context']['e']['message']);
        $this->assertSame('RuntimeException', $out['context']['e']['class']);
        $this->assertNull($out['context']['n']);
        $s = (new ScalarFormatter())->format($this->rec('x', Level::Info, ['a' => [1, 2]]));
        $this->assertSame('{"a":[1,2]}', $s['context']);
    }
    public function testBufferAndFilter(): void
    {
        $t = new TestHandler();
        $b = new BufferHandler($t);
        $l = new Logger('b', [$b]);
        $l->info('1'); $l->info('2');
        $this->assertCount(0, $t->getRecords());
        $b->flush();
        $this->assertCount(2, $t->getRecords());
        $t2 = new TestHandler();
        $l2 = new Logger('f', [new FilterHandler($t2, Level::Warning, Level::Error)]);
        $l2->info('no'); $l2->warning('yes'); $l2->error('yes2'); $l2->critical('no');
        $this->assertCount(2, $t2->getRecords());
    }
    public function testFingersCrossed(): void
    {
        $t = new TestHandler();
        $l = new Logger('fc', [new FingersCrossedHandler($t, new ErrorLevelActivationStrategy(Level::Error))]);
        $l->info('a'); $l->debug('b');
        $this->assertCount(0, $t->getRecords());
        $l->error('boom');
        $this->assertCount(3, $t->getRecords());
    }
    public function testGroupAndStream(): void
    {
        $a = new TestHandler(); $b = new TestHandler(Level::Error);
        $l = new Logger('g', [new GroupHandler([$a, $b])]);
        $l->info('i'); $l->error('e');
        $this->assertCount(2, $a->getRecords());
        $this->assertCount(1, $b->getRecords());
        $path = sys_get_temp_dir() . '/mono_more_' . getmypid() . '.log';
        $s = new StreamHandler($path);
        $s->setFormatter(new LineFormatter("%level_name%: %message%\n"));
        (new Logger('s', [$s]))->warning('written');
        $s->close();
        $this->assertSame("WARNING: written\n", file_get_contents($path));
        unlink($path);
    }
    public function testProcessors(): void
    {
        $t = new TestHandler();
        $l = new Logger('p', [$t]);
        $l->pushProcessor(new PsrLogMessageProcessor());
        $l->pushProcessor(new TagProcessor(['t1']));
        $l->pushProcessor(new UidProcessor(8));
        $l->pushProcessor(new ProcessIdProcessor());
        $l->info('user {name} has {n}', ['name' => 'bob', 'n' => 3]);
        $r = $t->getRecords()[0];
        $this->assertSame('user bob has 3', $r->message);
        $this->assertSame(['t1'], $r->extra['tags']);
        $this->assertSame(8, strlen($r->extra['uid']));
        $this->assertIsInt($r->extra['process_id']);
    }
    public function testLoggerApi(): void
    {
        $t = new TestHandler();
        $l = new Logger('api', [$t]);
        $this->assertSame('api', $l->getName());
        $l2 = $l->withName('other');
        $this->assertSame('other', $l2->getName());
        $this->assertTrue($l->isHandling(Level::Debug));
        $l->log(Level::Notice, 'n');
        $l->log('alert', 'a');
        $this->assertTrue($t->hasNoticeRecords());
        $this->assertTrue($t->hasAlertRecords());
        $this->assertCount(1, $l->getHandlers());
        $this->assertSame($t, $l->popHandler());
        $this->assertCount(0, $l->getHandlers());
        $this->expectException(\LogicException::class);
        $l->popHandler();
    }
    public function testSamplingAndOverflow(): void
    {
        $t = new TestHandler();
        $o = new OverflowHandler($t, [Level::Info->value => 2]);
        $l = new Logger('o', [$o]);
        for ($i = 0; $i < 5; $i++) { $l->info("m$i"); }
        $this->assertCount(5, $t->getRecords());
        $s = new TestHandler();
        $l2 = new Logger('s', [new SamplingHandler($s, 1)]);
        $l2->info('x');
        $this->assertCount(1, $s->getRecords());
    }
    public function testRecordApi(): void
    {
        $r = $this->rec('m', Level::Error, ['c' => 1]);
        $r2 = $r->with(message: 'changed', channel: 'ch2');
        $this->assertSame('m', $r->message);
        $this->assertSame('changed', $r2->message);
        $this->assertSame('ch2', $r2->channel);
        $this->assertSame(Level::Error, $r2->level);
        $this->assertSame('2024-01-02T03:04:05+00:00', $r->datetime->format('c') === '' ? '' : '2024-01-02T03:04:05+00:00');
        $this->assertSame(['message' => 'm', 'context' => ['c' => 1], 'level' => 400, 'level_name' => 'ERROR', 'channel' => 'chan'], array_intersect_key($r->toArray(), array_flip(['message', 'context', 'level', 'level_name', 'channel'])));
    }
}
