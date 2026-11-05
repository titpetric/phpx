<?php
use Monolog\{Logger, Level, LogRecord, Registry, Utils, DateTimeImmutable, JsonSerializableDateTimeImmutable};
use Monolog\Handler\{TestHandler, BufferHandler, OverflowHandler, FingersCrossedHandler, SamplingHandler, HandlerWrapper, GroupHandler, WhatFailureGroupHandler, NullHandler, NoopHandler, StreamHandler, AbstractProcessingHandler, ErrorLogHandler};
use Monolog\Formatter\{LineFormatter, JsonFormatter, NormalizerFormatter, RedactingFormatter, ScalarFormatter};
use Monolog\Processor\TagProcessor;

final class MonologCov8Str implements \Stringable
{
    public function __toString(): string { return 'stringable-value'; }
}

final class MonologCov8BadStr implements \Stringable
{
    public function __toString(): string { throw new \RuntimeException('no string'); }
}

final class MonologCov8JsonEx extends \Exception implements \JsonSerializable
{
    public function jsonSerialize(): array { return ['custom' => $this->getMessage()]; }
}

final class MonologCov8Closing extends AbstractProcessingHandler
{
    public static int $closed = 0;
    protected function write(LogRecord $record): void {}
    public function close(): void { self::$closed++; }
}

final class MonologCov8ThrowClose extends AbstractProcessingHandler
{
    protected function write(LogRecord $record): void {}
    public function close(): void { throw new \RuntimeException('close fails'); }
}

final class MonologCov8Test extends \Monolog\Test\MonologTestCase
{
    public function testMonologTestCaseHelpers(): void
    {
        $r = $this->getRecord();
        $this->assertSame(Level::Warning, $r->level);
        $this->assertSame('test', $r->message);
        $this->assertSame('test', $r->channel);
        $r = $this->getRecord('error', new MonologCov8Str(), ['a' => 1], 'ch', new \DateTimeImmutable('2020-01-01'), ['e' => 1]);
        $this->assertSame(Level::Error, $r->level);
        $this->assertSame('stringable-value', $r->message);
        $this->assertSame(['e' => 1], $r->extra);
        $multi = $this->getMultipleRecords();
        $this->assertCount(5, $multi);
        $this->assertSame('error', $multi[4]->message);
        $f = $this->getIdentityFormatter();
        $this->assertSame('warning', $f->format($multi[3]));
    }

    public function testDeprecatedTestCaseAlias(): void
    {
        $this->assertInstanceOf(\Monolog\Test\MonologTestCase::class, new \Monolog\Test\TestCase('x'));
    }

    public function testNestedFormatterDelegation(): void
    {
        $line = new LineFormatter("%message%");
        foreach ([
            fn($t) => new BufferHandler($t),
            fn($t) => new OverflowHandler($t),
            fn($t) => new FingersCrossedHandler($t),
            fn($t) => new SamplingHandler($t, 1),
            fn($t) => new HandlerWrapper($t),
        ] as $mk) {
            $inner = new TestHandler();
            $outer = $mk($inner);
            $this->assertSame($outer, $outer->setFormatter($line));
            $this->assertSame($line, $outer->getFormatter());
            $this->assertSame($line, $inner->getFormatter());
            $plain = $mk(new NoopHandler());
            try { $plain->setFormatter($line); $this->fail(); } catch (\UnexpectedValueException | \LogicException $e) { $this->assertNotSame('', $e->getMessage()); }
            try { $plain->getFormatter(); $this->fail(); } catch (\UnexpectedValueException | \LogicException $e) { $this->assertNotSame('', $e->getMessage()); }
        }
    }

    public function testHandlerWrapperNonProcessable(): void
    {
        $w = new HandlerWrapper(new NoopHandler());
        try { $w->pushProcessor(fn($r) => $r); $this->fail(); } catch (\LogicException $e) { $this->assertStringContainsString('ProcessableHandlerInterface', $e->getMessage()); }
        try { $w->popProcessor(); $this->fail(); } catch (\LogicException $e) { $this->assertStringContainsString('ProcessableHandlerInterface', $e->getMessage()); }
    }

    public function testSamplingFactoryClosure(): void
    {
        $s = new SamplingHandler(fn() => 5, 1);
        try { $s->handle($this->getRecord()); $this->fail(); } catch (\RuntimeException $e) { $this->assertStringContainsString('should return a HandlerInterface', $e->getMessage()); }
    }

    public function testOverflowBelowLevel(): void
    {
        $t = new TestHandler();
        $o = new OverflowHandler($t, [], Level::Error);
        $this->assertFalse($o->handle($this->getRecord(Level::Info)));
        $this->assertCount(0, $t->getRecords());
    }

    public function testGroupProcessorsOnBatch(): void
    {
        $a = new TestHandler();
        $g = new GroupHandler([$a]);
        $g->pushProcessor(new TagProcessor(['g']));
        $g->handleBatch([$this->getRecord(Level::Info), $this->getRecord(Level::Error)]);
        $this->assertSame(['tags' => ['g']], $a->getRecords()[1]->extra);
        $w = new WhatFailureGroupHandler([$a]);
        $w->pushProcessor(new TagProcessor(['w']));
        $w->handle($this->getRecord(Level::Info, 'single'));
        $w->handleBatch([$this->getRecord(Level::Info, 'batch')]);
        $this->assertSame(['tags' => ['w']], $a->getRecords()[3]->extra);
        $before = MonologCov8Closing::$closed;
        $w2 = new WhatFailureGroupHandler([new MonologCov8Closing(), new MonologCov8ThrowClose(), new MonologCov8Closing()]);
        $w2->close();
        $this->assertSame($before + 2, MonologCov8Closing::$closed);
    }

    public function testHandlerDestructAndSerialize(): void
    {
        $before = MonologCov8Closing::$closed;
        $h = new MonologCov8Closing();
        unset($h);
        $this->assertSame($before + 1, MonologCov8Closing::$closed);
        $t = new MonologCov8ThrowClose();
        unset($t);
        $n = new NullHandler();
        $arr = $n->__serialize();
        $this->assertIsArray($arr);
    }

    public function testFingersCrossedProcessorsOnHandle(): void
    {
        $t = new TestHandler();
        $f = new FingersCrossedHandler($t, Level::Error);
        $f->pushProcessor(new TagProcessor(['fc']));
        $f->handle($this->getRecord(Level::Error));
        $this->assertSame(['tags' => ['fc']], $t->getRecords()[0]->extra);
    }

    public function testJsonFormatterObjects(): void
    {
        $f = new JsonFormatter();
        $d = json_decode($f->format($this->getRecord(Level::Info, 'm', [
            'str' => new MonologCov8Str(),
            'bad' => new MonologCov8BadStr(),
            'js' => new MonologCov8JsonEx('jm'),
            'plain' => new \stdClass(),
            'ao' => new \ArrayObject([1, 2]),
            'dt' => new \DateTimeImmutable('2020-02-03 04:05:06', new \DateTimeZone('UTC')),
        ])), true);
        $this->assertSame('stringable-value', $d['context']['str']);
        $this->assertSame(MonologCov8BadStr::class, $d['context']['bad']);
        $this->assertSame(['custom' => 'jm'], $d['context']['js']);
        $this->assertSame('2020-02-03T04:05:06+00:00', $d['context']['dt']);
    }

    public function testNormalizerSettings(): void
    {
        $n = new NormalizerFormatter();
        $this->assertSame($n, $n->setMaxTraceLength(5));
        $this->assertSame(5, $n->getMaxTraceLength());
        $this->assertSame($n, $n->setBasePath(__DIR__));
        $this->assertSame($n, $n->setBasePath(''));
        $this->assertSame($n, $n->addJsonEncodeOption(JSON_PRETTY_PRINT));
        $this->assertSame($n, $n->removeJsonEncodeOption(JSON_PRETTY_PRINT));
        $e = new \RuntimeException('base path');
        $n->setBasePath(dirname(__DIR__));
        $out = $n->format($this->getRecord(Level::Info, 'm', ['e' => $e]));
        $this->assertSame('tests/MonologCov8Test.php:' . $e->getLine(), $out['context']['e']['file']);
        $f = new JsonFormatter();
        $this->assertIsString($f->format($this->getRecord()));
    }

    public function testNormalizerJsonSerializableException(): void
    {
        $n = new NormalizerFormatter();
        $out = $n->format($this->getRecord(Level::Info, 'm', ['e' => new MonologCov8JsonEx('ser')]));
        $this->assertSame(['custom' => 'ser'], $out['context']['e']);
    }

    public function testNormalizerIncompleteClass(): void
    {
        $obj = unserialize('O:16:"MonologCovNoSuch":1:{s:1:"a";i:1;}');
        $n = new NormalizerFormatter();
        $out = $n->format($this->getRecord(Level::Info, 'm', ['o' => $obj]));
        $this->assertSame(['__PHP_Incomplete_Class' => 'MonologCovNoSuch'], $out['context']['o']);
        $out = json_decode((new JsonFormatter())->format($this->getRecord(Level::Info, 'm', ['o' => $obj])), true);
        $this->assertArrayHasKey('o', $out['context']);
    }

    public function testNormalizerStringableAndDates(): void
    {
        $n = new NormalizerFormatter();
        $out = $n->format($this->getRecord(Level::Info, 'm', ['s' => new MonologCov8Str(), 'bad' => new MonologCov8BadStr(), 'mono' => new DateTimeImmutable(false, new \DateTimeZone('UTC'))]));
        $this->assertIsString($out['context']['mono']);
        $this->assertIsArray($out['context']['bad']);
    }

    public function testLineFormatterBasePathAndIndent(): void
    {
        $l = new LineFormatter(null, null, true, true);
        $this->assertSame($l, $l->setBasePath(dirname(__DIR__)));
        $this->assertSame($l, $l->indentStacktraces('    '));
        $this->assertSame($l, $l->includeStacktraces(true));
        $e = new \RuntimeException('indent me');
        $out = $l->format($this->getRecord(Level::Error, 'm', ['exception' => $e]));
        $this->assertStringContainsString('tests/MonologCov8Test.php:' . $e->getLine(), $out);
        $this->assertStringNotContainsString(dirname(__DIR__) . '/tests/MonologCov8Test.php:' . $e->getLine(), $out);
        $this->assertStringContainsString('[stacktrace]', $out);
        $this->assertSame($l, $l->setBasePath(''));
    }

    public function testLineFormatterStringifyScalars(): void
    {
        $l = new LineFormatter("%context%", null, true, false);
        $this->assertSame('{"t":true,"f":false,"n":null,"i":5,"d":1.5}', $l->format($this->getRecord(Level::Info, 'm', ['t' => true, 'f' => false, 'n' => null, 'i' => 5, 'd' => 1.5])));
        $this->assertSame('plain', $l->stringify('plain'));
        $this->assertSame('NULL', $l->stringify(null));
        $this->assertSame('true', $l->stringify(true));
    }

    public function testLineFormatterMaxDepthPrevious(): void
    {
        $l = new class(null, null, true, true) extends LineFormatter {
            public function ne(\Throwable $e): string { return $this->normalizeException($e); }
        };
        $l->setMaxNormalizeDepth(1);
        $e = new \RuntimeException('a', 0, new \LogicException('b', 0, new \DomainException('c')));
        $out = $l->ne($e);
        $this->assertStringContainsString('[previous exception]', $out);
        $this->assertStringContainsString('Over 1 levels deep, aborting normalization', $out);
    }

    public function testRedactingEdgeCases(): void
    {
        $f = new RedactingFormatter(new LineFormatter("%message% %context%\n"), ['token'], [], '[X]', true, 3);
        $obj = new \stdClass();
        $obj->self = $obj;
        $obj->list = ['a' => ['b' => ['c' => 'deepsecret']]];
        $out = $f->format($this->getRecord(Level::Info, 'm deepsecret', ['o' => $obj, 'token' => ['n' => ['m' => 'nestedsec']]]));
        $this->assertStringNotContainsString('nestedsec', $out);
        $out = $f->format($this->getRecord(Level::Info, 'no secrets here', ['a' => 1]));
        $this->assertStringContainsString('no secrets here', $out);
        $deep = ['token' => 'zzzzzz'];
        for ($i = 0; $i < 15; $i++) { $deep = ['k' => $deep]; }
        $out = $f->formatBatch([$this->getRecord(Level::Info, 'deep', $deep)]);
        $this->assertStringContainsString('deep', $out);
        $arr = new RedactingFormatter(new NormalizerFormatter(), ['token'], ['/zz/']);
        $res = $arr->format($this->getRecord(Level::Info, 'zz', ['token' => 'abcdef']));
        $this->assertSame('[REDACTED]', $res['message']);
    }

    public function testLoggerLoopDepthGuard(): void
    {
        $h = new TestHandler();
        $l = new Logger('guard', [$h]);
        $l->pushHandler(new class($l) extends AbstractProcessingHandler {
            public function __construct(private Logger $lg) { parent::__construct(); }
            protected function write(LogRecord $record): void { $this->lg->info('re-enter'); }
        });
        $l->info('start');
        $this->assertTrue($h->hasWarningThatContains('infinite logging loop'));
        $this->assertLessThanOrEqual(6, count($h->getRecords()));
    }

    public function testLoggerBubbleStop(): void
    {
        $a = new TestHandler(Level::Debug, false);
        $b = new TestHandler();
        $l = new Logger('bubble', [$a, $b]);
        $l->info('x');
        $this->assertCount(1, $a->getRecords());
        $this->assertCount(0, $b->getRecords());
        $this->assertTrue($l->addRecord(Level::Info, 'direct'));
    }

    public function testLoggerLevelErrors(): void
    {
        try { Logger::toMonologLevel('nope'); $this->fail(); } catch (\Psr\Log\InvalidArgumentException $e) { $this->assertStringContainsString('Level "nope" is not defined', $e->getMessage()); }
        try { Logger::toMonologLevel(12345); $this->fail(); } catch (\Psr\Log\InvalidArgumentException $e) { $this->assertStringContainsString('Level "12345" is not defined', $e->getMessage()); }
        $l = new Logger('lv');
        try { $l->log([], 'x'); $this->fail(); } catch (\InvalidArgumentException $e) { $this->assertStringContainsString('expected to be a string, int or', $e->getMessage()); }
        $h = new TestHandler();
        $l->pushHandler($h);
        $l->log(5, 'rfc int');
        $this->assertSame(Level::Notice, $h->getRecords()[0]->level);
    }

    public function testStreamHandlerCustomMode(): void
    {
        $f = sys_get_temp_dir() . '/monologcov8-' . getmypid() . '.log';
        try {
            $h = new StreamHandler($f, Level::Debug, true, null, false, 'w');
            $h->setFormatter(new LineFormatter("%message%\n"));
            $h->handle($this->getRecord(Level::Info, 'first'));
            $h->close();
            $h2 = new StreamHandler($f, Level::Debug, true, null, false, 'w');
            $h2->setFormatter(new LineFormatter("%message%\n"));
            $h2->handle($this->getRecord(Level::Info, 'second'));
            $h2->close();
            $this->assertSame("second\n", file_get_contents($f));
            $this->assertGreaterThan(0, $h2->getStreamChunkSize());
        } finally {
            if (is_file($f)) { unlink($f); }
        }
    }

    public function testStreamHandlerPhpMemoryUrl(): void
    {
        $h = new StreamHandler('php://memory');
        $h->setFormatter(new LineFormatter("%message%\n"));
        $h->handle($this->getRecord(Level::Info, 'mem'));
        $this->assertIsResource($h->getStream());
        $this->assertSame('php://memory', $h->getUrl());
        $h->reset();
        $this->assertIsResource($h->getStream());
        $h->close();
    }

    public function testErrorLogHandlerFormatter(): void
    {
        $h = new ErrorLogHandler();
        $this->assertInstanceOf(LineFormatter::class, $h->getFormatter());
        $this->assertTrue($h->isHandling($this->getRecord()));
    }
}
