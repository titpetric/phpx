<?php
use Monolog\{Logger, Level, LogRecord};
use Monolog\Handler\{TestHandler, BufferHandler, FilterHandler, GroupHandler, FingersCrossedHandler, StreamHandler, SamplingHandler, NullHandler, NoopHandler, OverflowHandler, WhatFailureGroupHandler, FallbackGroupHandler, ErrorLogHandler, ProcessHandler, PsrHandler, HandlerWrapper, DeduplicationHandler, RotatingFileHandler, AbstractProcessingHandler};
use Monolog\Handler\FingersCrossed\{ErrorLevelActivationStrategy, ChannelLevelActivationStrategy};
use Monolog\Formatter\{LineFormatter, JsonFormatter};
use Monolog\Processor\TagProcessor;
use PHPUnit\Framework\TestCase;

final class MonologCovBoom extends AbstractProcessingHandler
{
    public int $n = 0;
    protected function write(LogRecord $record): void { $this->n++; throw new \RuntimeException('boom'); }
}

final class MonologCov2Test extends TestCase
{
    private string $dir;

    protected function setUp(): void
    {
        $this->dir = sys_get_temp_dir() . '/monologcov2-' . getmypid();
        if (!is_dir($this->dir)) { mkdir($this->dir); }
    }

    protected function tearDown(): void
    {
        $this->rm($this->dir);
    }

    private function rm(string $d): void
    {
        if (!is_dir($d)) { return; }
        foreach (scandir($d) as $e) {
            if ($e === '.' || $e === '..') { continue; }
            $p = $d . '/' . $e;
            is_dir($p) ? $this->rm($p) : unlink($p);
        }
        rmdir($d);
    }

    private function rec(string $m = 'm', Level $l = Level::Info, array $ctx = [], string $ch = 'chan', ?\DateTimeImmutable $dt = null): LogRecord
    {
        return new LogRecord($dt ?? new \DateTimeImmutable('2024-01-02 03:04:05'), $ch, $l, $m, $ctx);
    }

    public function testTestHandlerApi(): void
    {
        $h = new TestHandler();
        $h->handle($this->rec('hello world', Level::Warning, ['a' => 1]));
        $h->handle($this->rec('second', Level::Error));
        $this->assertTrue($h->hasRecords(Level::Warning));
        $this->assertTrue($h->hasRecords('error'));
        $this->assertFalse($h->hasRecords(Level::Debug));
        $this->assertTrue($h->hasRecord('hello world', Level::Warning));
        $this->assertTrue($h->hasRecord(['message' => 'hello world', 'context' => ['a' => 1]], Level::Warning));
        $this->assertFalse($h->hasRecord(['message' => 'hello world', 'context' => ['a' => 2]], Level::Warning));
        $this->assertFalse($h->hasRecord('nope', Level::Warning));
        $this->assertTrue($h->hasRecordThatContains('world', Level::Warning));
        $this->assertTrue($h->hasRecordThatMatches('/^sec/', Level::Error));
        $this->assertTrue($h->hasRecordThatPasses(fn($r, $i) => $i === 0, Level::Error));
        $this->assertFalse($h->hasRecordThatPasses(fn($r) => false, Level::Warning));
        $this->assertFalse($h->hasRecordThatPasses(fn($r) => true, Level::Debug));
        $this->assertTrue($h->hasWarningRecords());
        $this->assertTrue($h->hasWarning('hello world'));
        $this->assertTrue($h->hasErrorThatContains('sec'));
        $this->assertFalse($h->hasCriticalRecords());
        $this->assertTrue($h->hasErrorThatMatches('/second/'));
        $this->assertTrue($h->hasWarningThatPasses(fn($r) => true));
        $this->assertCount(2, $h->getRecords());
        $h->setSkipReset(true);
        $h->reset();
        $this->assertCount(2, $h->getRecords());
        $h->setSkipReset(false);
        $h->reset();
        $this->assertCount(0, $h->getRecords());
        $h->handle($this->rec());
        $h->clear();
        $this->assertFalse($h->hasInfoRecords());
        try { $h->nonsense(); $this->fail(); } catch (\BadMethodCallException $e) {
            $this->assertStringContainsString('Call to undefined method', $e->getMessage());
        }
    }

    public function testAbstractHandlerBasics(): void
    {
        $h = new TestHandler(Level::Warning, false);
        $this->assertSame(Level::Warning, $h->getLevel());
        $this->assertFalse($h->getBubble());
        $this->assertTrue($h->isHandling($this->rec('x', Level::Error)));
        $this->assertFalse($h->isHandling($this->rec('x', Level::Info)));
        $this->assertFalse($h->handle($this->rec('x', Level::Info)));
        $this->assertTrue($h->handle($this->rec('x', Level::Error)));
        $this->assertSame($h, $h->setLevel('debug'));
        $this->assertSame(Level::Debug, $h->getLevel());
        $this->assertSame($h, $h->setBubble(true));
        $this->assertFalse($h->handle($this->rec('y')));
        $h->handleBatch([$this->rec('1'), $this->rec('2')]);
        $this->assertCount(4, $h->getRecords());
        $h->close();
        $this->assertNotNull($h->getFormatter());
        $h->setFormatter(new JsonFormatter());
        $this->assertInstanceOf(JsonFormatter::class, $h->getFormatter());
        $h->pushProcessor(new TagProcessor(['t']));
        $h->handle($this->rec('z'));
        $last = $h->getRecords()[count($h->getRecords()) - 1];
        $this->assertSame(['tags' => ['t']], $last->extra);
        $this->assertNotNull($h->popProcessor());
        $this->expectException(\LogicException::class);
        $h->popProcessor();
    }

    public function testStreamHandlerFileNoLock(): void
    {
        $file = $this->dir . '/sub/dir/nolock.log';
        $h = new StreamHandler($file, Level::Debug, true, 0644, false, 'a');
        $h->setFormatter(new LineFormatter("%level_name%:%message%\n"));
        $h->handle($this->rec('one'));
        $h->handle($this->rec('two', Level::Error));
        $this->assertSame($file, $h->getUrl());
        $this->assertIsResource($h->getStream());
        $h->close();
        $this->assertNull($h->getStream());
        $this->assertSame("INFO:one\nERROR:two\n", file_get_contents($file));
        $h->handle($this->rec('three'));
        $h->reset();
        $this->assertSame("INFO:one\nERROR:two\nINFO:three\n", file_get_contents($file));
    }

    public function testStreamHandlerFile(): void
    {
        $file = $this->dir . '/sub/dir/app.log';
        $h = new StreamHandler($file, Level::Debug, true, 0644, true);
        $h->setFormatter(new LineFormatter("%level_name%:%message%\n"));
        $h->handle($this->rec('one'));
        $h->handle($this->rec('two', Level::Error));
        $this->assertSame($file, $h->getUrl());
        $this->assertIsResource($h->getStream());
        $this->assertGreaterThan(0, $h->getStreamChunkSize());
        $h->close();
        $this->assertNull($h->getStream());
        $this->assertSame("INFO:one\nERROR:two\n", file_get_contents($file));
        $h->handle($this->rec('three'));
        $h->reset();
        $this->assertSame("INFO:one\nERROR:two\nINFO:three\n", file_get_contents($file));
        $h->close();
    }

    public function testStreamHandlerResourceAndErrors(): void
    {
        $r = fopen('php://memory', 'w+');
        $h = new StreamHandler($r);
        $h->setFormatter(new LineFormatter("%message%\n"));
        $h->handle($this->rec('mem'));
        rewind($r);
        $this->assertSame("mem\n", stream_get_contents($r));
        $this->assertNull($h->getUrl());
        $h->reset();
        fclose($r);

        $caught = null;
        try { new StreamHandler(123); } catch (\InvalidArgumentException $e) { $caught = $e; }
        $this->assertNotNull($caught);


        // a regular file standing where a directory is needed
        file_put_contents($this->dir . '/plain', 'x');
        $h = new StreamHandler($this->dir . '/plain/inner/x.log');
        try { $h->handle($this->rec()); $this->fail(); } catch (\UnexpectedValueException $e) {
            $this->assertStringContainsString('could not be created', $e->getMessage());
        }
        $h = new StreamHandler($this->dir . '/missing/../nonexist/' . str_repeat('a', 300) . '.log');
        try { $h->handle($this->rec()); $this->fail(); } catch (\UnexpectedValueException $e) {
            $this->assertStringContainsString('could not be', $e->getMessage());
        }
        $this->assertTrue(true);
    }

    public function testStreamHandlerFileUrl(): void
    {
        $h = new StreamHandler('file://' . $this->dir . '/viaurl.log');
        $h->handle($this->rec('u'));
        $this->assertFileExists($this->dir . '/viaurl.log');
        $h->close();
    }

    public function testStreamHandlerInodeChange(): void
    {
        $file = $this->dir . '/ino.log';
        $h = new StreamHandler($file);
        $h->setFormatter(new LineFormatter("%message%\n"));
        $h->handle($this->rec('a'));
        unlink($file);
        $h->handle($this->rec('b'));
        $this->assertSame("b\n", file_get_contents($file));
        $h->close();
    }

    public function testRotatingFileHandler(): void
    {
        $base = $this->dir . '/rot.log';
        $h = new RotatingFileHandler($base, 2);
        $h->setFormatter(new LineFormatter("%message%\n"));
        $today = date('Y-m-d');
        $expected = $this->dir . '/rot-' . $today . '.log';
        $h->handle($this->rec('x', Level::Info, [], 'c', new \DateTimeImmutable()));
        $h->close();
        $this->assertFileExists($expected);
        unlink($expected);
        foreach (['2001-01-01', '2001-01-02', '2001-01-03', '2001-01-04'] as $d) {
            file_put_contents($this->dir . "/rot-$d.log", 'old');
        }
        file_put_contents($this->dir . '/other.log', 'keep');
        $h = new RotatingFileHandler($base, 2);
        $h->setFormatter(new LineFormatter("%message%\n"));
        $h->handle($this->rec('y', Level::Info, [], 'c', new \DateTimeImmutable()));
        $h->close();
        $files = array_values(array_diff(scandir($this->dir), ['.', '..']));
        sort($files);
        $this->assertSame(['other.log', 'rot-2001-01-04.log', "rot-$today.log"], $files);
        $h->reset();
    }

    public function testRotatingFormatsAndNextRotation(): void
    {
        $base = $this->dir . '/r.log';
        $h = new RotatingFileHandler($base, 0, Level::Debug, true, null, false, RotatingFileHandler::FILE_PER_MONTH, '{date}/{filename}');
        $h->setFormatter(new LineFormatter("%message%\n"));
        $h->handle($this->rec('m', Level::Info, [], 'c', new \DateTimeImmutable()));
        $h->close();
        $this->assertFileExists($this->dir . '/' . date('Y-m') . '/r.log');
        $h2 = new RotatingFileHandler($base, 0, Level::Debug, true, null, false, RotatingFileHandler::FILE_PER_YEAR, '{filename}-{date}', new \DateTimeZone('UTC'));
        $h2->setFormatter(new LineFormatter("%message%\n"));
        // a record dated far in the future forces a rotation
        $h2->handle($this->rec('future', Level::Info, [], 'c', new \DateTimeImmutable('+3 years')));
        $h2->close();
        $this->assertFileExists($this->dir . '/r-' . gmdate('Y') . '.log');
        $this->assertSame($h2, $h2->setFilenameFormat('{date}-{filename}', RotatingFileHandler::FILE_PER_DAY));
        $this->expectException(\InvalidArgumentException::class);
        $h2->setFilenameFormat('nodate', RotatingFileHandler::FILE_PER_DAY);
    }

    public function testRotatingBadDateFormat(): void
    {
        $h = new RotatingFileHandler($this->dir . '/z.log');
        $this->expectException(\InvalidArgumentException::class);
        $h->setFilenameFormat('{filename}-{date}', 'invalid');
    }

    public function testRotatingNested(): void
    {
        $base = $this->dir . '/n.log';
        for ($i = 1; $i <= 3; $i++) {
            mkdir($this->dir . "/200$i/01", 0777, true);
            file_put_contents($this->dir . "/200$i/01/n.log", 'x');
        }
        $h = new RotatingFileHandler($base, 1, Level::Debug, true, null, false, 'Y/m', '{date}/{filename}');
        $h->setFormatter(new LineFormatter("%message%\n"));
        $h->handle($this->rec('now', Level::Info, [], 'c', new \DateTimeImmutable()));
        $h->close();
        $this->assertDirectoryDoesNotExist($this->dir . '/2001');
        $this->assertFileExists($this->dir . '/' . date('Y/m') . '/n.log');
    }

    public function testBufferHandler(): void
    {
        $t = new TestHandler();
        $b = new BufferHandler($t, 2, Level::Info, true, true);
        $b->handle($this->rec('dbg', Level::Debug));
        $b->handle($this->rec('a'));
        $b->handle($this->rec('b'));
        $this->assertCount(0, $t->getRecords());
        $b->handle($this->rec('c'));
        $this->assertCount(2, $t->getRecords());
        $b->flush();
        $this->assertCount(3, $t->getRecords());
        $b->flush();
        $b->handle($this->rec('d'));
        $b->clear();
        $b->close();
        $this->assertCount(3, $t->getRecords());
        $b->handle($this->rec('e'));
        $b->flush();
        $this->assertCount(4, $t->getRecords());
        $b->reset();
        $this->assertCount(0, $t->getRecords());

        $t2 = new TestHandler();
        $b2 = new BufferHandler($t2, 2, Level::Debug, true, false);
        foreach (['1', '2', '3', '4'] as $m) { $b2->handle($this->rec($m)); }
        $b2->close();
        $this->assertCount(2, $t2->getRecords());
        $this->assertSame('3', $t2->getRecords()[0]->message);
        // processors on the buffer and a closed buffer
        $t3 = new TestHandler();
        $b3 = new BufferHandler($t3);
        $b3->pushProcessor(new TagProcessor(['p']));
        $b3->handle($this->rec('p'));
        $b3->flush();
        $this->assertSame(['tags' => ['p']], $t3->getRecords()[0]->extra);
        $this->assertFalse((new BufferHandler($t3, 0, Level::Error, false))->handle($this->rec('lo', Level::Debug)));
    }

    public function testGroupHandlers(): void
    {
        $a = new TestHandler(); $b = new TestHandler(Level::Error);
        $g = new GroupHandler([$a, $b], false);
        $this->assertTrue($g->isHandling($this->rec()));
        $this->assertFalse((new GroupHandler([$b]))->isHandling($this->rec('x', Level::Debug)));
        $this->assertTrue($g->handle($this->rec('1')));
        $g->handle($this->rec('2', Level::Error));
        $g->handleBatch([$this->rec('3'), $this->rec('4', Level::Error)]);
        $this->assertCount(4, $a->getRecords());
        $this->assertCount(2, $b->getRecords());
        $g->pushProcessor(new TagProcessor(['g']));
        $g->handle($this->rec('5'));
        $this->assertSame(['tags' => ['g']], $a->getRecords()[4]->extra);
        $g->popProcessor();
        $f = new LineFormatter("%message%\n");
        $this->assertSame($g, $g->setFormatter($f));
        $this->assertSame($f, $a->getFormatter());
        $g->reset();
        $this->assertCount(0, $a->getRecords());
        $g->close();
        $this->assertFalse((new GroupHandler([$a]))->handle($this->rec()));
        $this->expectException(\InvalidArgumentException::class);
        new GroupHandler([new \stdClass()]);
    }

    public function testWhatFailureGroup(): void
    {
        $boom = new MonologCovBoom(); $t = new TestHandler();
        $g = new WhatFailureGroupHandler([$boom, $t]);
        $g->handle($this->rec('x'));
        $g->handleBatch([$this->rec('y'), $this->rec('z')]);
        $this->assertSame(2, $boom->n);
        $this->assertCount(3, $t->getRecords());
        $g2 = new WhatFailureGroupHandler([$boom, $t], false);
        $this->assertTrue($g2->handle($this->rec('w')));
        $this->expectException(\InvalidArgumentException::class);
        new WhatFailureGroupHandler([1]);
    }

    public function testFallbackGroup(): void
    {
        $boom = new MonologCovBoom(); $t = new TestHandler(); $t2 = new TestHandler();
        $g = new FallbackGroupHandler([$boom, $t, $t2]);
        $g->handle($this->rec('x'));
        $this->assertCount(1, $t->getRecords());
        $this->assertCount(0, $t2->getRecords());
        $g->handleBatch([$this->rec('a'), $this->rec('b')]);
        $this->assertCount(3, $t->getRecords());
        $only = new FallbackGroupHandler([new MonologCovBoom()]);
        $this->assertFalse($only->handle($this->rec('x')));
        $pp = new FallbackGroupHandler([$t2], false);
        $pp->pushProcessor(new TagProcessor(['fb']));
        $this->assertTrue($pp->handle($this->rec('x')));
        $pp->handleBatch([$this->rec('y')]);
        $this->assertSame(['tags' => ['fb']], $t2->getRecords()[1]->extra);
    }

    public function testFallbackGroupBatchAllFail(): void
    {
        $g = new FallbackGroupHandler([new MonologCovBoom()]);
        $g->handleBatch([$this->rec('x')]);
        $this->assertTrue(true);
    }

    public function testFilterHandler(): void
    {
        $t = new TestHandler();
        $f = new FilterHandler($t, Level::Warning, Level::Error);
        $this->assertSame([Level::Warning, Level::Error, Level::Critical, Level::Alert, Level::Emergency][0], $f->getAcceptedLevels()[0]);
        foreach ([Level::Debug, Level::Warning, Level::Error, Level::Critical] as $l) { $f->handle($this->rec('m', $l)); }
        $this->assertCount(2, $t->getRecords());
        $this->assertTrue($f->isHandling($this->rec('x', Level::Warning)));
        $this->assertFalse($f->isHandling($this->rec('x', Level::Info)));
        $f->setAcceptedLevels([Level::Info, 'notice']);
        $this->assertCount(2, $f->getAcceptedLevels());
        $f->handle($this->rec('i', Level::Info));
        $f->handleBatch([$this->rec('n', Level::Notice), $this->rec('w', Level::Warning)]);
        $this->assertCount(4, $t->getRecords());
        $f->pushProcessor(new TagProcessor(['f']));
        $f->handle($this->rec('t', Level::Info));
        $this->assertSame(['tags' => ['f']], $t->getRecords()[4]->extra);
        $f->popProcessor();
        $f->setFormatter(new JsonFormatter());
        $this->assertInstanceOf(JsonFormatter::class, $f->getFormatter());
        $f->reset();
        $f->close();
        $cl = new FilterHandler(fn($rec, $h) => new TestHandler(), Level::Debug, Level::Emergency, false);
        $this->assertTrue($cl->handle($this->rec()));
        $bad = new FilterHandler(fn() => 5);
        try { $bad->handle($this->rec()); $this->fail(); } catch (\RuntimeException $e) {
            $this->assertStringContainsString('should return a HandlerInterface', $e->getMessage());
        }
        $this->assertCount(0, (new TestHandler())->getRecords());
        try { $bad->getHandler(); $this->fail(); } catch (\RuntimeException $e) { $this->assertTrue(true); }
        $this->assertInstanceOf(TestHandler::class, (new FilterHandler($t))->getHandler());
    }

    public function testFingersCrossed(): void
    {
        $t = new TestHandler();
        $f = new FingersCrossedHandler($t, Level::Error, 2);
        $f->handle($this->rec('1', Level::Debug));
        $f->handle($this->rec('2', Level::Debug));
        $f->handle($this->rec('3', Level::Debug));
        $this->assertCount(0, $t->getRecords());
        $this->assertTrue($f->isHandling($this->rec()));
        $f->handle($this->rec('boom', Level::Error));
        $this->assertSame(['3', 'boom'], array_map(fn($r) => $r->message, $t->getRecords()));
        $f->handle($this->rec('after', Level::Debug));
        $this->assertCount(3, $t->getRecords());
        $f->close();
        $f->handle($this->rec('again', Level::Debug));
        $this->assertCount(3, $t->getRecords());
        $f->reset();
        $this->assertCount(0, $t->getRecords());

        $t = new TestHandler();
        $f = new FingersCrossedHandler($t, new ErrorLevelActivationStrategy(Level::Warning), 0, true, false, Level::Notice);
        $f->handle($this->rec('w', Level::Warning));
        $f->handle($this->rec('d', Level::Debug));
        $this->assertCount(1, $t->getRecords());
        $f->clear();
        $f->close();
        $t = new TestHandler();
        $f = new FingersCrossedHandler($t, Level::Error, 0, true, true, Level::Notice);
        $f->handle($this->rec('n', Level::Notice));
        $f->handle($this->rec('i', Level::Info));
        $f->close();
        $this->assertSame(['n'], array_map(fn($r) => $r->message, $t->getRecords()));
        $f->pushProcessor(new TagProcessor(['x']));
        $this->assertNotNull($f->getHandler());
        $f->setFormatter(new JsonFormatter());
        $f->clear();
        $f->activate();
        $cl = new FingersCrossedHandler(fn($r, $h) => new TestHandler(), Level::Error);
        $cl->handle($this->rec('x', Level::Error));
        $this->expectException(\RuntimeException::class);
        (new FingersCrossedHandler(fn() => 1))->getHandler();
    }

    public function testActivationStrategies(): void
    {
        $e = new ErrorLevelActivationStrategy('warning');
        $this->assertTrue($e->isHandlerActivated($this->rec('x', Level::Warning)));
        $this->assertFalse($e->isHandlerActivated($this->rec('x', Level::Notice)));
        $c = new ChannelLevelActivationStrategy(Level::Critical, ['chan' => Level::Info, 'other' => 'error']);
        $this->assertTrue($c->isHandlerActivated($this->rec('x', Level::Info)));
        $this->assertFalse($c->isHandlerActivated($this->rec('x', Level::Error, [], 'zzz')));
        $this->assertTrue($c->isHandlerActivated($this->rec('x', Level::Critical, [], 'zzz')));
        $this->assertFalse($c->isHandlerActivated($this->rec('x', Level::Warning, [], 'other')));
        $this->expectException(\InvalidArgumentException::class);
        new FingersCrossedHandler(new TestHandler(), 'not-a-level-xyz');
    }

    public function testSamplingOverflowNullNoop(): void
    {
        $t = new TestHandler();
        $s = new SamplingHandler($t, 1);
        for ($i = 0; $i < 5; $i++) { $s->handle($this->rec('m')); }
        $this->assertCount(5, $t->getRecords());
        $this->assertTrue($s->isHandling($this->rec()));
        $s->pushProcessor(new TagProcessor(['s']));
        $s->handle($this->rec('p'));
        $this->assertSame(['tags' => ['s']], $t->getRecords()[5]->extra);
        $s->popProcessor();
        $s->setFormatter(new JsonFormatter());
        $this->assertInstanceOf(JsonFormatter::class, $s->getFormatter());
        $s->close();
        $s2 = new SamplingHandler(fn($r, $h) => new TestHandler(), 1);
        $s2->handle($this->rec());
        $this->assertSame($t, $s->getHandler());
        $s2->reset();

        $t = new TestHandler();
        $o = new OverflowHandler($t, [Level::Debug->value => 1, Level::Info->value => 2], Level::Debug);
        for ($i = 0; $i < 2; $i++) { $o->handle($this->rec('i')); }
        $this->assertCount(0, $t->getRecords());
        $o->handle($this->rec('i'));
        $this->assertCount(3, $t->getRecords());
        $o->handle($this->rec('w', Level::Warning));
        $o->handle($this->rec('d', Level::Debug));
        $this->assertGreaterThanOrEqual(2, count($t->getRecords()));

        $n = new NullHandler(Level::Error);
        $this->assertTrue($n->isHandling($this->rec('x', Level::Error)));
        $this->assertFalse($n->isHandling($this->rec('x', Level::Info)));
        $this->assertTrue($n->handle($this->rec('x', Level::Error)));
        $this->assertFalse($n->handle($this->rec('x', Level::Info)));
        $nb = new NoopHandler();
        $this->assertTrue($nb->isHandling($this->rec()));
        $this->assertFalse($nb->handle($this->rec()));
        $nb->handleBatch([$this->rec()]);
        $nb->close();
    }

    public function testHandlerWrapper(): void
    {
        $t = new TestHandler();
        $w = new HandlerWrapper($t);
        $this->assertTrue($w->isHandling($this->rec()));
        $w->handle($this->rec('a'));
        $w->handleBatch([$this->rec('b')]);
        $this->assertCount(2, $t->getRecords());
        $w->pushProcessor(new TagProcessor(['w']));
        $w->handle($this->rec('c'));
        $this->assertSame(['tags' => ['w']], $t->getRecords()[2]->extra);
        $this->assertNotNull($w->popProcessor());
        $f = new LineFormatter();
        $w->setFormatter($f);
        $this->assertSame($f, $w->getFormatter());
        $w->reset();
        $this->assertCount(0, $t->getRecords());
        $w->close();
    }

    public function testErrorLogHandler(): void
    {
        $file = $this->dir . '/err.log';
        $h = new ErrorLogHandler(ErrorLogHandler::OPERATING_SYSTEM, Level::Debug);
        $this->assertSame(ErrorLogHandler::OPERATING_SYSTEM, 0);
        $h->setFormatter(new LineFormatter("%message%"));
        $old = ini_set('error_log', $file);
        try {
            $h2 = new ErrorLogHandler(ErrorLogHandler::OPERATING_SYSTEM, Level::Debug, true, true);
            $h2->setFormatter(new LineFormatter("%message%"));
            $h2->handle($this->rec("a\nb"));
        } finally {
            ini_set('error_log', $old === false ? '' : $old);
        }
        $this->assertFileExists($file);
        $this->assertStringContainsString('a', file_get_contents($file));
        $this->assertStringContainsString('b', file_get_contents($file));
    }

    public function testErrorLogHandlerTypes(): void
    {
        $this->assertSame([0, 4], ErrorLogHandler::getAvailableTypes());
        $h = new ErrorLogHandler(ErrorLogHandler::SAPI, Level::Warning, false, true);
        $this->assertSame(Level::Warning, $h->getLevel());
        $this->assertFalse($h->isHandling($this->rec('x', Level::Info)));
        $this->expectException(\InvalidArgumentException::class);
        new ErrorLogHandler(99);
    }

    public function testPsrHandler(): void
    {
        $inner = new Logger('inner', [$th = new TestHandler()]);
        $p = new PsrHandler($inner, Level::Info);
        $p->handle($this->rec('x', Level::Warning, ['a' => 1]));
        $p->handle($this->rec('y', Level::Debug));
        $this->assertCount(1, $th->getRecords());
        $this->assertSame(['a' => 1], $th->getRecords()[0]->context);
        $p2 = new PsrHandler($inner);
        $p2->setFormatter(new LineFormatter("%message%!"));
        $p2->handle($this->rec('fmt'));
        $this->assertSame('fmt!', $th->getRecords()[1]->message);
        $this->assertNotNull($p2->getFormatter());
    }

    public function testProcessHandlerArguments(): void
    {
        try { new ProcessHandler(''); $this->fail(); } catch (\InvalidArgumentException $e) { $this->assertStringContainsString('non-empty', $e->getMessage()); }
        try { new ProcessHandler('cat', Level::Debug, true, ''); $this->fail(); } catch (\InvalidArgumentException $e) { $this->assertStringContainsString('CWD', $e->getMessage()); }
        $h = new ProcessHandler('cat', Level::Debug, true, '/tmp', 0.5);
        $h->close();
        $this->assertTrue($h->isHandling($this->rec()));
    }

    public function testProcessHandler(): void
    {
        $out = $this->dir . '/proc.out';
        $h = new ProcessHandler('cat > ' . escapeshellarg($out), Level::Debug, true, null, 2.0);
        $h->setFormatter(new LineFormatter("%message%\n"));
        $h->handle($this->rec('piped'));
        $h->close();
        $h->close();
        $this->assertSame("piped\n", file_get_contents($out));
        $bad = new ProcessHandler('echo oops 1>&2', Level::Debug, true, $this->dir, 1.0);
        $bad->setFormatter(new LineFormatter("%message%\n"));
        try { $bad->handle($this->rec()); $this->fail(); } catch (\UnexpectedValueException $e) {
            $this->assertStringContainsString('oops', $e->getMessage());
        }
        $bad->close();
    }

    public function testDeduplication(): void
    {
        $store = $this->dir . '/dedup.log';
        $t = new TestHandler();
        $d = new DeduplicationHandler($t, $store, Level::Error, 60);
        $now = new \DateTimeImmutable();
        $d->handle($this->rec("same\nsecond line", Level::Error, [], 'c', $now));
        $d->handle($this->rec('info', Level::Info, [], 'c', $now));
        $d->flush();
        $this->assertCount(2, $t->getRecords());
        $this->assertFileExists($store);
        $d->handle($this->rec('same', Level::Error, [], 'c', $now));
        $d->close();
        $this->assertCount(2, $t->getRecords());
        $d->handle($this->rec('different', Level::Error, [], 'c', $now));
        $d->handle($this->rec('different', Level::Error, [], 'c', $now));
        $d->flush();
        $this->assertCount(4, $t->getRecords());
        $d->flush();
        // old entries trigger gc
        file_put_contents($store, "1000:ERROR:old\nbadline\n" . time() . ":ERROR:fresh\n");
        $d->handle($this->rec('new', Level::Error, [], 'c', $now));
        $d->flush();
        $this->assertStringNotContainsString('old', file_get_contents($store));
        $this->assertStringContainsString('fresh', file_get_contents($store));
    }

    public function testDeduplicationBelowLevel(): void
    {
        $t = new TestHandler();
        $d2 = new DeduplicationHandler($t, $this->dir . '/none.log', 'critical');
        $d2->handle($this->rec('low', Level::Debug));
        $d2->flush();
        $this->assertTrue($t->hasDebug('low'));
        $d2->flush();
        $d2->close();
    }
}
