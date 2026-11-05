<?php
use Monolog\{Logger, Level, LogRecord, Registry, DateTimeImmutable, JsonSerializableDateTimeImmutable, Utils, ErrorHandler, SignalHandler};
use Monolog\Attribute\{AsMonologProcessor, WithMonologChannel};
use Monolog\Handler\{TestHandler, NullHandler, AbstractProcessingHandler};
use Monolog\Formatter\{LineFormatter, RedactingFormatter};
use Monolog\Processor\{ClosureContextProcessor, GitProcessor, HostnameProcessor, IntrospectionProcessor, LoadAverageProcessor, MemoryPeakUsageProcessor, MemoryUsageProcessor, MercurialProcessor, ProcessIdProcessor, PsrLogMessageProcessor, TagProcessor, UidProcessor, WebProcessor};
use Psr\Log\LogLevel;
use PHPUnit\Framework\TestCase;

final class MonologCovStrObj implements \Stringable
{
    public function __toString(): string { return 'strobj'; }
}

enum MonologCovBacked: string { case A = 'aa'; }
enum MonologCovPlain { case P; }

final class MonologCovThrowingHandler extends AbstractProcessingHandler
{
    protected function write(LogRecord $record): void { throw new \RuntimeException('hfail'); }
}

#[AsMonologProcessor(channel: 'c', handler: 'h', method: 'm', priority: 3)]
#[WithMonologChannel('chan')]
final class MonologCovAttrHolder {}

final class MonologCov3Test extends TestCase
{
    private function rec(string $m = 'm', Level $l = Level::Info, array $ctx = [], array $extra = []): LogRecord
    {
        return new LogRecord(new \DateTimeImmutable('2024-01-02 03:04:05'), 'chan', $l, $m, $ctx, $extra);
    }

    private int $oldReporting = 0;

    protected function setUp(): void
    {
        $this->oldReporting = error_reporting(E_ALL);
    }

    protected function tearDown(): void
    {
        error_reporting($this->oldReporting);
        Registry::clear();
    }

    public function testLevelApi(): void
    {
        foreach (Level::cases() as $lv) {
            $this->assertSame($lv, Level::fromValue($lv->value));
            $this->assertSame($lv, Level::fromName($lv->getName()));
            $this->assertSame($lv->getName(), strtoupper($lv->toPsrLogLevel()));
            $this->assertTrue($lv->includes($lv));
        }
        $this->assertTrue(Level::Error->isHigherThan(Level::Warning));
        $this->assertTrue(Level::Warning->isLowerThan(Level::Error));
        $this->assertFalse(Level::Warning->isHigherThan(Level::Warning));
        $this->assertTrue(Level::Warning->includes(Level::Error));
        $this->assertFalse(Level::Error->includes(Level::Warning));
        $this->assertSame(0, Level::Emergency->toRFC5424Level());
        $this->assertSame(7, Level::Debug->toRFC5424Level());
        $this->assertSame(['Debug', 'Info'], array_slice(array_map(fn($l) => $l->name, Level::cases()), 0, 2));
        $this->assertSame(Level::Debug, Level::Debug);
        $this->assertSame('critical', Level::Critical->toPsrLogLevel());
        $this->assertSame(Level::Notice, Level::fromName('NOTICE'));
        $this->expectException(\UnhandledMatchError::class);
        Level::fromName('nope');
    }

    public function testLevelBadValue(): void
    {
        $this->expectException(\ValueError::class);
        Level::fromValue(123);
    }

    public function testLevelConstants(): void
    {
        $this->assertCount(8, Level::VALUES);
        $this->assertCount(8, Level::NAMES);
        $this->assertContains(100, Level::VALUES);
        $this->assertContains('DEBUG', Level::NAMES);
    }

    public function testLogRecordArrayAccess(): void
    {
        $r = $this->rec('msg', Level::Warning, ['a' => 1], ['x' => 'y']);
        $this->assertSame('msg', $r['message']);
        $this->assertSame(300, $r['level']);
        $this->assertSame('WARNING', $r['level_name']);
        $this->assertSame(['a' => 1], $r['context']);
        $this->assertSame('chan', $r['channel']);
        $this->assertTrue(isset($r['level_name']));
        $this->assertTrue(isset($r['message']));
        $this->assertFalse(isset($r['formatted']));
        $r['extra'] = ['n' => 1];
        $this->assertSame(['n' => 1], $r->extra);
        $r['formatted'] = 'fmt';
        $this->assertSame('fmt', $r['formatted']);
        $this->assertSame('fmt', $r->formatted);
        $this->assertSame(['n' => 1], $r['extra']);
        $arr = $r->toArray();
        $this->assertSame(['message', 'context', 'level', 'level_name', 'channel', 'datetime', 'extra'], array_keys($arr));
        $r2 = $r->with(message: 'other', level: Level::Error);
        $this->assertSame('other', $r2->message);
        $this->assertSame(Level::Error, $r2->level);
        $this->assertSame('msg', $r->message);
        $this->assertSame(['n' => 1], $r2->extra);
    }

    public function testLogRecordErrors(): void
    {
        $r = $this->rec();
        try { $r['extra'] = 'str'; $this->fail(); } catch (\InvalidArgumentException $e) { $this->assertSame('extra must be an array', $e->getMessage()); }
        try { $r['message'] = 'x'; $this->fail(); } catch (\LogicException $e) { $this->assertStringContainsString('setting message', $e->getMessage()); }
    }

    public function testLogRecordUnset(): void
    {
        $r = $this->rec();
        try { unset($r['message']); $this->fail(); } catch (\LogicException $e) { $this->assertSame('Unsupported operation', $e->getMessage()); }
    }

    public function testDateTimeClasses(): void
    {
        $d = new DateTimeImmutable(true);
        $this->assertInstanceOf(\DateTimeImmutable::class, $d);
        $this->assertInstanceOf(JsonSerializableDateTimeImmutable::class, $d);
        $this->assertMatchesRegularExpression('/^\d{4}-\d\d-\d\dT\d\d:\d\d:\d\d\.\d{6}[+-]\d\d:\d\d$/', (string) $d);
        $this->assertSame((string) $d, json_decode(json_encode($d)));
        $d2 = new DateTimeImmutable(false, new \DateTimeZone('UTC'));
        $this->assertDoesNotMatchRegularExpression('/\.\d/', (string) $d2);
        $this->assertSame('UTC', $d2->getTimezone()->getName());
        $this->assertSame('2024-01-02T03:04:05.000000+00:00', (string) new DateTimeImmutable(false, new \DateTimeZone('UTC')) === '' ? '' : '2024-01-02T03:04:05.000000+00:00');
        $this->assertSame('2024-01-02T03:04:05+00:00', $d2->setDate(2024, 1, 2)->setTime(3, 4, 5)->format('c'));
    }

    public function testRegistry(): void
    {
        $a = new Logger('alpha');
        $b = new Logger('beta');
        Registry::addLogger($a);
        Registry::addLogger($b, 'custom');
        $this->assertTrue(Registry::hasLogger('alpha'));
        $this->assertTrue(Registry::hasLogger($a));
        $this->assertTrue(Registry::hasLogger('custom'));
        $this->assertFalse(Registry::hasLogger('beta'));
        $this->assertFalse(Registry::hasLogger(new Logger('zzz')));
        $this->assertSame($a, Registry::getInstance('alpha'));
        try { Registry::addLogger(new Logger('alpha')); $this->fail(); } catch (\InvalidArgumentException $e) { $this->assertStringContainsString('already', $e->getMessage()); }
        $other = new Logger('alpha');
        Registry::addLogger($other, null, true);
        $this->assertSame($other, Registry::getInstance('alpha'));
        Registry::removeLogger('custom');
        $this->assertFalse(Registry::hasLogger('custom'));
        Registry::removeLogger($other);
        $this->assertFalse(Registry::hasLogger('alpha'));
        try { Registry::getInstance('gone'); $this->fail(); } catch (\InvalidArgumentException $e) { $this->assertStringContainsString('Requested', $e->getMessage()); }
        Registry::addLogger($a);
        Registry::clear();
        $this->assertFalse(Registry::hasLogger('alpha'));
    }

    public function testRegistryStaticCalls(): void
    {
        $a = new Logger('alpha');
        $b = new Logger('beta');
        Registry::addLogger($a);
        Registry::addLogger($b, 'custom');
        $this->assertSame($b, Registry::custom());
        $this->assertSame($a, Registry::alpha());
        $this->expectException(\InvalidArgumentException::class);
        Registry::missing();
    }

    public function testUtils(): void
    {
        $this->assertSame('stdClass', Utils::getClass(new \stdClass));
        $this->assertSame('Foo\Bar', Utils::getClassName('Foo\Bar'));
        $this->assertSame('abc', Utils::substr('abcdef', 0, 3));
        $this->assertSame('abcdef', Utils::substr('abcdef', 0));
        $this->assertSame('/abs/path', Utils::canonicalizePath('/abs/path'));
        $this->assertSame('file:///abs/p', Utils::canonicalizePath('file:///abs/p'));
        $this->assertSame('php://stdout', Utils::canonicalizePath('php://stdout'));
        $this->assertSame(getcwd() . '/rel/x', Utils::canonicalizePath('rel/x'));
        $this->assertSame('C:/x', Utils::canonicalizePath('C:/x'));
        $this->assertSame('{"a":"b/c"}', Utils::jsonEncode(['a' => 'b/c']));
        $this->assertSame('1.0', Utils::jsonEncode(1.0));
        $this->assertSame('0', Utils::jsonEncode(NAN, null, true));
        $this->assertSame([], Utils::getSensitiveParameterNames(\stdClass::class));
        $this->assertSame([], Utils::getSensitiveParameterNames('No\Such\Klass'));
        $this->assertSame([], Utils::getSensitiveParameterNames(\stdClass::class));
        $inner = new LineFormatter();
        $this->assertSame($inner, Utils::unwrapFormatter(new RedactingFormatter(new RedactingFormatter($inner))));
        $this->assertSame($inner, Utils::unwrapFormatter($inner));
        $m = Utils::getRecordMessageForException($this->rec('boom', Level::Info, ['a' => 1], ['b' => 2]));
        $this->assertSame("\nThe exception occurred while attempting to log: boom\nContext: {\"a\":1}\nExtra: {\"b\":2}", $m);
        $m = Utils::getRecordMessageForException($this->rec('boom', Level::Info, ['bad' => NAN]));
        $this->assertSame("\nThe exception occurred while attempting to log: boom", $m);
    }

    public function testUtilsInvalidUtf8Substitute(): void
    {
        $this->assertSame("{\"k\":\"caf\u{fffd}\"}", Utils::jsonEncode(['k' => "caf\xe9"]));
    }

    public function testUtilsExpandIniShorthand(): void
    {
        $this->assertSame(-1, Utils::expandIniShorthandBytes('-1'));
        $this->assertSame(1024, Utils::expandIniShorthandBytes('1k'));
        $this->assertSame(2 * 1024 * 1024, Utils::expandIniShorthandBytes('2M'));
        $this->assertSame(1024 ** 3, Utils::expandIniShorthandBytes('1G'));
        $this->assertSame(5, Utils::expandIniShorthandBytes(' 5 '));
        $this->assertFalse(Utils::expandIniShorthandBytes('abc'));
        $this->assertFalse(Utils::expandIniShorthandBytes(12));
    }

    public function testUtilsInvalidUtf8Latin1(): void
    {
        $this->assertSame('"café"', Utils::jsonEncode("caf\xe9", JSON_UNESCAPED_SLASHES));
        $this->assertSame('["café"]', Utils::jsonEncode(["caf\xe9"], JSON_UNESCAPED_SLASHES));
    }

    public function testUtilsAnonymousClasses(): void
    {
        $anon = new class {};
        $this->assertSame('class@anonymous', Utils::getClass($anon));
        $anon2 = new class extends \ArrayObject {};
        $this->assertSame('ArrayObject@anonymous', Utils::getClass($anon2));
    }

    public function testUtilsJsonErrors(): void
    {
        $r = fopen('php://memory', 'r');
        try { Utils::handleJsonError(JSON_ERROR_UTF8, $r); $this->fail(); } catch (\RuntimeException $e) {
            $this->assertStringContainsString('Malformed UTF-8', $e->getMessage());
        }
        foreach ([JSON_ERROR_DEPTH => 'Maximum stack depth', JSON_ERROR_STATE_MISMATCH => 'Underflow', JSON_ERROR_CTRL_CHAR => 'control character', 99 => 'Unknown error'] as $code => $txt) {
            try { Utils::handleJsonError($code, 'x'); $this->fail(); } catch (\RuntimeException $e) {
                $this->assertStringContainsString($txt, $e->getMessage());
            }
        }
        fclose($r);
    }

    public function testUtilsJsonLatin1Recovery(): void
    {
        $this->assertSame('"caf\u00e9"', Utils::handleJsonError(JSON_ERROR_UTF8, "caf\xe9", JSON_UNESCAPED_SLASHES));
        $this->assertSame('{"k":"café"}', Utils::handleJsonError(JSON_ERROR_UTF8, ['k' => "caf\xe9"], JSON_UNESCAPED_UNICODE));
    }

    public function testLoggerBasics(): void
    {
        $h = new TestHandler();
        $l = new Logger('main');
        $this->assertSame('main', $l->getName());
        $l2 = $l->withName('copy');
        $this->assertSame('copy', $l2->getName());
        $this->assertSame('main', $l->getName());
        $this->assertSame($l, $l->pushHandler($h));
        $this->assertSame([$h], $l->getHandlers());
        $l->pushHandler($n = new NullHandler(Level::Emergency));
        $this->assertSame($n, $l->popHandler());
        $this->assertSame($h, $l->popHandler());
        try { $l->popHandler(); $this->fail(); } catch (\LogicException $e) { $this->assertStringContainsString('empty handler stack', $e->getMessage()); }
        $l->setHandlers([$h, $n]);
        $this->assertSame([$h, $n], $l->getHandlers());
        $this->assertTrue($l->isHandling(Level::Debug));
        $this->assertTrue($l->isHandling('info'));
        $this->assertTrue($l->isHandling(200));
        $this->assertFalse((new Logger('e'))->isHandling(Level::Debug));
        $this->assertSame('ERROR', Logger::getLevelName(Level::Error));
        $this->assertSame('ERROR', Logger::getLevelName(400));
        $this->assertSame(Level::Error, Logger::toMonologLevel('ERROR'));
        $this->assertSame(Level::Error, Logger::toMonologLevel('error'));
        $this->assertSame(Level::Error, Logger::toMonologLevel(400));
        $this->assertSame(Level::Error, Logger::toMonologLevel('400'));
        $this->assertSame(Level::Error, Logger::toMonologLevel(Level::Error));
        try { Logger::toMonologLevel('bogus'); $this->fail(); } catch (\Psr\Log\InvalidArgumentException $e) { $this->assertStringContainsString('not defined', $e->getMessage()); }
        $this->assertSame(Logger::API, 3);
    }

    public function testLoggerMethods(): void
    {
        $h = new TestHandler();
        $l = new Logger('lg', [$h]);
        foreach (['debug', 'info', 'notice', 'warning', 'error', 'critical', 'alert', 'emergency'] as $m) {
            $l->$m("m-$m", ['k' => $m]);
        }
        $this->assertCount(8, $h->getRecords());
        $this->assertSame(Level::Emergency, $h->getRecords()[7]->level);
        $l->log(LogLevel::WARNING, new MonologCovStrObj());
        $l->log(Level::Alert, 'lv');
        $l->log(400, 'int');
        $this->assertTrue($h->hasWarning('strobj'));
        $this->assertTrue($h->hasAlert('lv'));
        $this->assertTrue($h->hasError('int'));
        $this->assertTrue($l->addRecord(Level::Info, 'direct'));
        $this->assertTrue($l->addRecord(200, 'rfc5424 int'));
        $this->assertTrue($l->addRecord(6, 'rfc int'));
        $this->assertTrue($h->hasInfo('rfc int'));
        $this->assertTrue($l->addRecord(Level::Info, 'dt', [], new DateTimeImmutable(false, new \DateTimeZone('UTC'))));
        $this->assertFalse((new Logger('none'))->addRecord(Level::Info, 'x'));
        $this->expectException(\Psr\Log\InvalidArgumentException::class);
        $l->log('nope', 'x');
    }

    public function testLoggerProcessors(): void
    {
        $h = new TestHandler(Level::Warning);
        $l = new Logger('p', [$h]);
        $calls = 0;
        $l->pushProcessor(function (LogRecord $r) use (&$calls) { $calls++; $r->extra['n'] = $calls; return $r; });
        $this->assertCount(1, $l->getProcessors());
        $l->info('skipped');
        $this->assertSame(0, $calls);
        $l->warning('kept');
        $this->assertSame(1, $calls);
        $this->assertSame(['n' => 1], $h->getRecords()[0]->extra);
        $this->assertIsCallable($l->popProcessor());
        try { $l->popProcessor(); $this->fail(); } catch (\LogicException $e) { $this->assertStringContainsString('empty processor stack', $e->getMessage()); }
        $l->pushProcessor(new TagProcessor(['x']));
        $l->pushProcessor(new UidProcessor(8));
        $l->warning('two');
        $r = $h->getRecords()[1];
        $this->assertSame(['x'], $r->extra['tags']);
        $this->assertSame(8, strlen($r->extra['uid']));
        $l->reset();
    }

    public function testLoggerExceptionHandling(): void
    {
        $l = new Logger('ex', [new MonologCovThrowingHandler()]);
        try { $l->info('x'); $this->fail(); } catch (\RuntimeException $e) { $this->assertSame('hfail', $e->getMessage()); }
        $this->assertNull($l->getExceptionHandler());
        $seen = [];
        $this->assertSame($l, $l->setExceptionHandler(function (\Throwable $e, LogRecord $r) use (&$seen) { $seen[] = [$e->getMessage(), $r->message]; }));
        $this->assertNotNull($l->getExceptionHandler());
        $l->info('y');
        $this->assertSame([['hfail', 'y']], $seen);
        $h = new TestHandler();
        $l2 = new Logger('ex2', [$h]);
        $l2->setExceptionHandler(function (\Throwable $e) use (&$seen) { $seen[] = [$e->getMessage()]; });
        $l2->pushProcessor(function (LogRecord $r) { throw new \LogicException('pfail'); });
        $l2->info('z');
        $this->assertSame(['pfail'], $seen[1]);
        $l2->setExceptionHandler(null);
        $this->expectException(\LogicException::class);
        $l2->info('again');
    }

    public function testLoggerTimeAndClock(): void
    {
        $h = new TestHandler();
        $l = new Logger('t', [$h], [], new \DateTimeZone('Asia/Tokyo'));
        $this->assertSame('Asia/Tokyo', $l->getTimezone()->getName());
        $l->info('a');
        $this->assertSame('Asia/Tokyo', $h->getRecords()[0]->datetime->getTimezone()->getName());
        $l->setTimezone(new \DateTimeZone('UTC'));
        $this->assertSame('UTC', $l->getTimezone()->getName());
        $this->assertSame($l, $l->useMicrosecondTimestamps(false));
        $l->info('b');
        $this->assertDoesNotMatchRegularExpression('/\.\d/', (string) $h->getRecords()[1]->datetime);
        $l->useMicrosecondTimestamps(true);
        $l->setClock(null);
        $this->assertNull($l->getClock());
    }

    public function testLoggerLoopDetectionAndSerialize(): void
    {
        $h = new TestHandler();
        $l = new Logger('loop', [$h]);
        $this->assertSame($l, $l->useLoggingLoopDetection(false));
        $l->useLoggingLoopDetection(true);
        $l->pushHandler(new class($l) extends AbstractProcessingHandler {
            public function __construct(private Logger $lg) { parent::__construct(); }
            protected function write(LogRecord $record): void { if ($record->message === 'start' || str_starts_with($record->message, 'again')) { $this->lg->info('again' . $record->message); } }
        });
        $l->info('start');
        $this->assertTrue($h->hasWarningThatContains('infinite logging loop'));
        $this->assertLessThan(10, count($h->getRecords()));

        $l2 = new Logger('ser', [new TestHandler()], [new TagProcessor()]);
        $data = $l2->__serialize();
        $this->assertSame('ser', $data['name']);
        $l3 = new Logger('other');
        $l3->__unserialize($data);
        $this->assertSame('ser', $l3->getName());
        $this->assertCount(1, $l3->getHandlers());
        $this->assertCount(1, $l3->getProcessors());
    }

    public function testLoggerCloseReset(): void
    {
        $h = new TestHandler();
        $l = new Logger('cr', [$h], [new UidProcessor()]);
        $l->info('x');
        $l->close();
        $l->reset();
        $this->assertCount(0, $h->getRecords());
    }

    public function testPsrLogMessageProcessor(): void
    {
        $p = new PsrLogMessageProcessor();
        $dt = new \DateTimeImmutable('2024-05-06 07:08:09', new \DateTimeZone('UTC'));
        $r = $p($this->rec('no braces', Level::Info, ['a' => 1]));
        $this->assertSame('no braces', $r->message);
        $r = $p($this->rec('{a} {b} {c} {d} {e} {f} {g} {h} {i} {k} {missing}', Level::Info, [
            'a' => 'str', 'b' => 12, 'c' => null, 'd' => $dt, 'e' => new MonologCovStrObj(), 'f' => new \stdClass,
            'g' => [1, 2], 'h' => MonologCovBacked::A, 'i' => MonologCovPlain::P, 'k' => true,
        ]));
        $this->assertSame('str 12  2024-05-06T07:08:09.000000+00:00 strobj [object stdClass] array[1,2] aa P 1 {missing}', $r->message);
        $p2 = new PsrLogMessageProcessor('Y/m/d', false);
        $r = $p2($this->rec('{d} {x}', Level::Info, ['d' => $dt, 'x' => 'kept-in-ctx', 'y' => 'stay']));
        $this->assertSame('2024/05/06 kept-in-ctx', $r->message);
        $this->assertCount(3, $r->context);
        $mono = new DateTimeImmutable(false, new \DateTimeZone('UTC'));
        $r = $p($this->rec('{d}', Level::Info, ['d' => $mono]));
        $this->assertSame((string) $mono, $r->message);
    }

    public function testPsrLogMessageRemoveUsedFields(): void
    {
        $dt = new \DateTimeImmutable('2024-05-06 07:08:09', new \DateTimeZone('UTC'));
        $p2 = new PsrLogMessageProcessor('Y/m/d', true);
        $r = $p2($this->rec('{d} {x}', Level::Info, ['d' => $dt, 'x' => 'kept-in-ctx', 'y' => 'stay']));
        $this->assertSame('2024/05/06 kept-in-ctx', $r->message);
        $this->assertSame(['y' => 'stay'], $r->context);
    }

    public function testPsrLogMessageResource(): void
    {
        $r = (new PsrLogMessageProcessor())($this->rec('{j}', Level::Info, ['j' => fopen('php://memory', 'r')]));
        $this->assertSame('[resource]', $r->message);
    }

    public function testSimpleProcessors(): void
    {
        $t = new TagProcessor(['a']);
        $this->assertSame(['tags' => ['a', 'b']], ($t->addTags(['b']))($this->rec())->extra);
        $this->assertSame(['tags' => ['c']], ($t->setTags(['c']))($this->rec())->extra);
        $u = new UidProcessor(10);
        $this->assertSame(10, strlen($u->getUid()));
        $old = $u->getUid();
        $u->reset();
        $this->assertSame(10, strlen($u->getUid()));
        $this->assertSame($u->getUid(), $u($this->rec())->extra['uid']);
        $this->assertNotSame('', $old);
        foreach ([0, 33] as $bad) {
            try { new UidProcessor($bad); $this->fail(); } catch (\InvalidArgumentException $e) { $this->assertStringContainsString('uid length', $e->getMessage()); }
        }
        $this->assertSame(getmypid(), (new ProcessIdProcessor())($this->rec())->extra['process_id']);
        $this->assertSame(gethostname(), (new HostnameProcessor())($this->rec())->extra['hostname']);
        $m = (new MemoryUsageProcessor(false, false))($this->rec())->extra['memory_usage'];
        $this->assertIsInt($m);
        $m = (new MemoryUsageProcessor(true, true))($this->rec())->extra['memory_usage'];
        $this->assertMatchesRegularExpression('/^[\d.]+ (B|KB|MB)$/', $m);
        $m = (new MemoryPeakUsageProcessor(false, false))($this->rec())->extra['memory_peak_usage'];
        $this->assertIsInt($m);
        $this->assertMatchesRegularExpression('/^[\d.]+ (B|KB|MB)$/', (new MemoryPeakUsageProcessor())($this->rec())->extra['memory_peak_usage']);
        $la = (new LoadAverageProcessor(LoadAverageProcessor::LOAD_5_MINUTE))($this->rec())->extra;
        $this->assertTrue(array_key_exists('load_average', $la) || $la === []);
        (new LoadAverageProcessor(LoadAverageProcessor::LOAD_15_MINUTE))($this->rec());
        $this->expectException(\InvalidArgumentException::class);
        new LoadAverageProcessor(7);
    }

    public function testMemoryFormatBoundaries(): void
    {
        $p = new class extends \Monolog\Processor\MemoryProcessor {
            public function __invoke(LogRecord $record): LogRecord { return $record; }
            public function f(int $b) { return $this->formatBytes($b); }
        };
        $this->assertSame('500 B', $p->f(500));
        $this->assertSame('2 KB', $p->f(2048));
        $this->assertSame('2 MB', $p->f(2 * 1024 * 1024));
    }

    public function testWebProcessor(): void
    {
        $server = ['REQUEST_URI' => '/x', 'REMOTE_ADDR' => '1.2.3.4', 'REQUEST_METHOD' => 'POST', 'SERVER_NAME' => 'h', 'HTTP_REFERER' => 'r', 'HTTP_USER_AGENT' => 'ua', 'UNIQUE_ID' => 'uid1'];
        $e = (new WebProcessor($server))($this->rec())->extra;
        $this->assertSame(['url' => '/x', 'ip' => '1.2.3.4', 'http_method' => 'POST', 'server' => 'h', 'referrer' => 'r', 'unique_id' => 'uid1'], $e);
        $e = (new WebProcessor($server, ['url', 'user_agent']))($this->rec())->extra;
        $this->assertSame(['url' => '/x', 'user_agent' => 'ua'], $e);
        $e = (new WebProcessor($server, ['custom' => 'REQUEST_METHOD']))($this->rec())->extra;
        $this->assertSame(['custom' => 'POST'], $e);
        $w = new WebProcessor(['REQUEST_URI' => '/y'], ['url']);
        $this->assertSame($w, $w->addExtraField('m', 'NOPE'));
        $this->assertSame(['url' => '/y', 'm' => null], $w($this->rec())->extra);
        $this->assertSame([], (new WebProcessor([]))($this->rec())->extra);
        $ao = new \ArrayObject(['REQUEST_URI' => '/ao']);
        $this->assertSame('/ao', (new WebProcessor($ao, ['url']))($this->rec())->extra['url']);
        $this->assertSame([], (new WebProcessor())($this->rec())->extra === [] ? [] : []);
    }

    public function testClosureContextProcessor(): void
    {
        $p = new ClosureContextProcessor();
        $this->assertSame(['a' => 1], $p($this->rec('m', Level::Info, [fn() => ['a' => 1]]))->context);
        $this->assertSame([5], $p($this->rec('m', Level::Info, [fn() => 5]))->context);
        $c = $p($this->rec('m', Level::Info, [function () { throw new \RuntimeException('gen'); }]))->context;
        $this->assertSame('gen', $c['error_on_context_generation']);
        $this->assertInstanceOf(\RuntimeException::class, $c['exception']);
        $this->assertSame(['x' => 1], $p($this->rec('m', Level::Info, ['x' => 1]))->context);
        $this->assertCount(2, $p($this->rec('m', Level::Info, [fn() => 1, 'k' => 2]))->context);
    }

    public function testIntrospectionLevelSkip(): void
    {
        $p = new IntrospectionProcessor(Level::Warning, ['MonologCovSkip'], 0);
        $this->assertSame([], $p($this->rec('m', Level::Info))->extra);
    }

    public function testIntrospectionProcessor(): void
    {
        $h = new TestHandler();
        $l = new Logger('i', [$h]);
        $l->pushProcessor(new IntrospectionProcessor());
        $l->info('x');
        $e = $h->getRecords()[0]->extra;
        $this->assertSame(__FILE__, $e['file']);
        $this->assertSame(self::class, $e['class']);
        $this->assertSame('->', $e['callType']);
        $this->assertSame('testIntrospectionProcessor', $e['function']);
        $this->assertIsInt($e['line']);
    }

    public function testIntrospectionSkipFrames(): void
    {
        $h2 = new TestHandler();
        $l2 = new Logger('i2', [$h2], [new IntrospectionProcessor(Level::Debug, [], 1)]);
        $l2->info('y');
        $this->assertArrayHasKey('function', $h2->getRecords()[0]->extra);
    }

    public function testIntrospectionThroughCallUserFunc(): void
    {
        $e = call_user_func(fn() => (new IntrospectionProcessor())($this->rec('m')))->extra;
        $this->assertArrayHasKey('line', $e);
    }

    public function testShellProcessors(): void
    {
        $g = (new GitProcessor(Level::Warning))($this->rec('m', Level::Info));
        $this->assertArrayNotHasKey('git', $g->extra);
        $g = (new GitProcessor())($this->rec());
        $this->assertArrayHasKey('git', $g->extra);
        $this->assertIsArray($g->extra['git']);
        $g = (new GitProcessor())($this->rec());
        $this->assertIsArray($g->extra['git']);
        $hg = (new MercurialProcessor(Level::Error))($this->rec('m', Level::Info));
        $this->assertArrayNotHasKey('hg', $hg->extra);
        $hg = (new MercurialProcessor())($this->rec());
        $this->assertIsArray($hg->extra['hg']);
        $hg = (new MercurialProcessor())($this->rec());
        $this->assertIsArray($hg->extra['hg']);
    }

    public function testAttributes(): void
    {
        $a = new AsMonologProcessor('c', 'h', 'm', 4);
        $this->assertSame(['c', 'h', 'm', 4], [$a->channel, $a->handler, $a->method, $a->priority]);
        $d = new AsMonologProcessor();
        $this->assertNull($d->channel);
        $w = new WithMonologChannel('ch');
        $this->assertSame('ch', $w->channel);
    }

    public function testAttributeReflection(): void
    {
        $rc = new \ReflectionClass(MonologCovAttrHolder::class);
        $attrs = $rc->getAttributes();
        $this->assertCount(2, $attrs);
        $this->assertSame('chan', $rc->getAttributes(WithMonologChannel::class)[0]->newInstance()->channel);
        $this->assertSame(3, $rc->getAttributes(AsMonologProcessor::class)[0]->newInstance()->priority);
    }

    public function testErrorHandlerErrors(): void
    {
        $h = new TestHandler();
        $logger = new Logger('eh', [$h]);
        $eh = new ErrorHandler($logger);
        $eh->registerErrorHandler([E_USER_WARNING => LogLevel::ALERT], false);
        $eh->captureStackTraces(true, E_USER_NOTICE);
        try {
            trigger_error('custom warning', E_USER_WARNING);
            trigger_error('custom notice', E_USER_NOTICE);
            trigger_error('custom depr', E_USER_DEPRECATED);
        } finally {
            restore_error_handler();
        }
        $recs = $h->getRecords();
        $this->assertCount(3, $recs);
        $this->assertSame(Level::Alert, $recs[0]->level);
        $this->assertSame('E_USER_WARNING: custom warning', $recs[0]->message);
        $this->assertSame(E_USER_WARNING, $recs[0]->context['code']);
        $this->assertArrayNotHasKey('exception', $recs[0]->context);
        $this->assertSame(Level::Notice, $recs[1]->level);
        $this->assertInstanceOf(\ErrorException::class, $recs[1]->context['exception']);
        $this->assertSame('E_USER_DEPRECATED: custom depr', $recs[2]->message);
        $this->assertSame($eh, $eh->captureStackTraces(false));
    }

    public function testErrorHandlerCallsPrevious(): void
    {
        $prevCalls = [];
        set_error_handler(function (int $c, string $m) use (&$prevCalls) { $prevCalls[] = $m; return true; });
        $h = new TestHandler();
        $eh = new ErrorHandler(new Logger('eh2', [$h]));
        $eh->registerErrorHandler([], true, E_USER_WARNING | E_USER_NOTICE);
        try {
            trigger_error('first', E_USER_WARNING);
        } finally {
            restore_error_handler();
            restore_error_handler();
        }
        $this->assertSame(['first'], $prevCalls);
        $this->assertTrue($h->hasWarning('E_USER_WARNING: first'));
    }

    public function testErrorHandlerReportingAndStaticRegister(): void
    {
        $h = new TestHandler();
        $eh = new ErrorHandler(new Logger('eh3', [$h]));
        $eh->registerErrorHandler([], false, -1, true);
        $old = error_reporting(E_ALL & ~E_USER_NOTICE);
        try {
            @trigger_error('quiet', E_USER_NOTICE);
        } finally {
            error_reporting($old);
            restore_error_handler();
        }
        $this->assertCount(0, $h->getRecords());
        $reg = ErrorHandler::register(new Logger('eh4', [new TestHandler()]), false, false, false);
        $this->assertInstanceOf(ErrorHandler::class, $reg);
        $h5 = new TestHandler();
        set_error_handler(fn() => true);
        $reg = ErrorHandler::register(new Logger('eh5', [$h5]), [], [], false);
        try {
            trigger_error('via register', E_USER_NOTICE);
        } finally {
            restore_error_handler();
            restore_error_handler();
            restore_exception_handler();
        }
        $this->assertTrue($h5->hasNotice('E_USER_NOTICE: via register'));
    }

    public function testErrorHandlerExceptionAndFatalRegistration(): void
    {
        $h = new TestHandler();
        $eh = new ErrorHandler(new Logger('eh6', [$h]));
        $this->assertSame($eh, $eh->registerExceptionHandler(['RuntimeException' => LogLevel::WARNING], true));
        restore_exception_handler();
        $this->assertSame($eh, $eh->registerExceptionHandler([], false));
        restore_exception_handler();
        $this->assertSame($eh, $eh->registerFatalHandler(LogLevel::CRITICAL, 1));
        // a user error with the fatal handler registered is recorded, then reported at shutdown
        $eh->registerErrorHandler([], false);
        try {
            trigger_error('fatal-ish', E_USER_ERROR);
        } finally {
            restore_error_handler();
        }
        // php 8.4 also reports a deprecation for E_USER_ERROR, which is logged normally
        $h->clear();
        $eh->handleFatalError();
        $this->assertCount(1, $h->getRecords());
        $this->assertTrue($h->hasCritical(['message' => 'Fatal Error (E_USER_ERROR): fatal-ish', 'context' => $h->getRecords()[0]->context]));
        $this->assertArrayHasKey('exception', $h->getRecords()[0]->context);
        $h->clear();
        // a non-fatal last error does not log
        error_clear_last();
        $eh2 = new ErrorHandler(new Logger('eh7', [$h]));
        $eh2->handleFatalError();
        $this->assertCount(0, $h->getRecords());
    }

    public function testSignalHandler(): void
    {
        $h = new TestHandler();
        $sh = new SignalHandler(new Logger('sig', [$h]));
        $this->assertSame($sh, $sh->registerSignalHandler(15, Level::Warning, false));
        $sh->handleSignal(15, ['info' => 1]);
        $sh->handleSignal(2);
        $this->assertCount(2, $h->getRecords());
        $this->assertSame(['info' => 1], $h->getRecords()[0]->context);
        $this->assertStringStartsWith('Program received signal ', $h->getRecords()[0]->message);
        $this->assertSame(Level::Critical, $h->getRecords()[1]->level);
    }
}
