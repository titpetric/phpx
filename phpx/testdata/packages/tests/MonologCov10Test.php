<?php
foreach (['Gelf', 'MongoBson', 'Elasticsearch', 'Global', 'AmqpMsg', 'AmqpChan'] as $stub) {
    require_once __DIR__ . '/MonologCovStub' . $stub . '.php';
}

use Monolog\{Logger, Level, LogRecord, Utils, SignalHandler};
use Monolog\Handler\{SocketHandler, StreamHandler, BrowserConsoleHandler, ChromePHPHandler, AmqpHandler, ElasticsearchHandler, TestHandler};
use Monolog\Formatter\{LineFormatter, NormalizerFormatter, GelfMessageFormatter, ElasticsearchFormatter};
use Monolog\Processor\IntrospectionProcessor;
use PHPUnit\Framework\TestCase;

final class MonologCov10Conn extends SocketHandler
{
    public array $sent = [];
    protected function fsockopen() { return fopen('php://memory', 'w+'); }
    protected function pfsockopen() { return fopen('php://memory', 'w+'); }
    protected function streamSetTimeout(): bool { return true; }
    protected function streamSetChunkSize(): int|bool { return 1; }
    protected function fwrite(string $data): int|bool { $this->sent[] = $data; return strlen($data); }
    public function isConnected(): bool { return is_resource($this->getResource()); }
}

final class MonologCov10Pair extends SocketHandler
{
    public $peer;
    public function __construct(?int $chunk = null, string $cs = 'unix://pair')
    {
        parent::__construct($cs, Level::Debug, true, false, 1.5, 5.0, 1.0, $chunk);
    }
    protected function fsockopen()
    {
        $pair = stream_socket_pair(STREAM_PF_UNIX, STREAM_SOCK_STREAM, STREAM_IPPROTO_IP);
        $this->peer = $pair[1];
        return $pair[0];
    }
}

final class MonologCov10Browser extends BrowserConsoleHandler
{
    public static function fmt(string $ct): string { return self::getResponseFormatFromContentType($ct); }
    public static function fmtNow(): string { return self::getResponseFormat(); }
}

final class MonologCov10BrowserHtml extends BrowserConsoleHandler
{
    protected static function getResponseFormat(): string { return 'html'; }
}

final class MonologCov10Es6 extends \Elasticsearch\Client
{
    const VERSION = '6.8.0';
}

final class MonologCov10Chrome extends ChromePHPHandler
{
    public array $headers = [];
    public static function resetAll(): void { static::$initialized = false; static::$overflowed = false; static::$sendHeaders = true; static::$json = ['version' => static::VERSION, 'columns' => ['label', 'log', 'backtrace', 'type'], 'rows' => []]; }
    protected function isWebRequest(): bool { return true; }
    protected function sendHeader(string $header, string $content): void { $this->headers[$header] = $content; }
}

final class MonologCov10Test extends TestCase
{
    private function rec(string $m = 'm', Level $l = Level::Info, array $ctx = [], array $extra = []): LogRecord
    {
        return new LogRecord(new \DateTimeImmutable('2024-01-02 03:04:05', new \DateTimeZone('UTC')), 'chan', $l, $m, $ctx, $extra);
    }

    public function testSocketSettersOnStub(): void
    {
        $h = new MonologCov10Conn('tcp://x.invalid:1');
        $this->assertSame($h, $h->setConnectionTimeout(4.0));
        $this->assertSame($h, $h->setTimeout(0.25));
        $this->assertSame($h, $h->setWritingTimeout(0.0));
        $this->assertSame($h, $h->setChunkSize(2048));
        $this->assertSame(4.0, $h->getConnectionTimeout());
        $this->assertSame(0.25, $h->getTimeout());
        $this->assertSame(0.0, $h->getWritingTimeout());
        $this->assertSame(2048, $h->getChunkSize());
        $h->setFormatter(new LineFormatter("%message%\n"));
        $h->setPersistent(true);
        $h->handle($this->rec('persistent'));
        $this->assertSame(["persistent\n"], $h->sent);
        $h->closeSocket();
        foreach (['setConnectionTimeout', 'setTimeout', 'setWritingTimeout'] as $m) {
            try { $h->$m(-0.5); $this->fail(); } catch (\InvalidArgumentException $e) { $this->assertStringContainsString('Timeout must be 0 or a positive float', $e->getMessage()); }
        }
    }

    public function testSocketWithRealStreamPair(): void
    {
        $h = new MonologCov10Pair(4096);
        $h->setFormatter(new LineFormatter("%message%\n"));
        $h->handle($this->rec('over a socket pair'));
        $this->assertSame("over a socket pair\n", fread($h->peer, 100));
        $this->assertTrue($h->isConnected());
        $h->handle($this->rec('second'));
        $this->assertSame("second\n", fread($h->peer, 100));
        $h->close();
        $this->assertFalse($h->isConnected());
        fclose($h->peer);
    }

    public function testSocketRealStreamPairPersistent(): void
    {
        $h = new MonologCov10Pair();
        $h->setFormatter(new LineFormatter("%message%\n"));
        $h->handle($this->rec('no chunk size'));
        $this->assertSame("no chunk size\n", fread($h->peer, 100));
        $h->closeSocket();
        fclose($h->peer);
    }

    public function testStreamHandlerReadOnlyResourceFails(): void
    {
        $f = sys_get_temp_dir() . '/monologcov10-ro-' . getmypid();
        file_put_contents($f, 'x');
        $res = fopen($f, 'r');
        try {
            $h = new StreamHandler($res);
            $h->setFormatter(new LineFormatter("%message%\n"));
            try { $h->handle($this->rec('cannot write')); $this->fail('expected failure'); } catch (\UnexpectedValueException $e) {
                $this->assertStringContainsString('Writing to the log file failed', $e->getMessage());
                $this->assertStringContainsString('cannot write', $e->getMessage());
            }
        } finally {
            fclose($res);
            unlink($f);
        }
    }

    public function testStreamHandlerReopensAfterExternalClose(): void
    {
        $f = sys_get_temp_dir() . '/monologcov10-reopen-' . getmypid();
        try {
            $h = new StreamHandler($f);
            $h->setFormatter(new LineFormatter("%message%\n"));
            $h->handle($this->rec('one'));
            fclose($h->getStream());
            $h->handle($this->rec('two'));
            $h->close();
            $this->assertSame("one\ntwo\n", file_get_contents($f));
        } finally {
            if (is_file($f)) { unlink($f); }
        }
    }

    public function testStreamHandlerMemoryLimitVariants(): void
    {
        $old = ini_get('memory_limit');
        try {
            foreach (['-1', '512M', '64M'] as $limit) {
                ini_set('memory_limit', $limit);
                $h = new StreamHandler('php://memory');
                $this->assertGreaterThan(0, $h->getStreamChunkSize());
            }
        } finally {
            ini_set('memory_limit', $old);
        }
    }

    public function testBrowserConsoleContentTypes(): void
    {
        $this->assertSame('js', MonologCov10Browser::fmt('Content-Type: application/javascript'));
        $this->assertSame('js', MonologCov10Browser::fmt('content-type: text/javascript'));
        $this->assertSame('html', MonologCov10Browser::fmt('Content-Type: text/html; charset=utf-8'));
        $this->assertSame('unknown', MonologCov10Browser::fmt('Content-Type: application/json'));
    }

    public function testBrowserConsoleDefaultFormatInCli(): void
    {
        $this->assertSame('html', MonologCov10Browser::fmtNow());
    }

    public function testBrowserConsoleRealOutput(): void
    {
        MonologCov10BrowserHtml::resetStatic();
        $h = new MonologCov10BrowserHtml();
        $h->handle($this->rec('buffered', Level::Warning, ['a' => 1]));
        $h->handle($this->rec('custom style', Level::Info));
        ob_start();
        MonologCov10BrowserHtml::send();
        $out = ob_get_clean();
        $this->assertStringStartsWith('<script>(function (c)', $out);
        $this->assertStringContainsString('c.groupCollapsed(', $out);
        $this->assertStringContainsString('c.info(', $out);
        MonologCov10BrowserHtml::resetStatic();
    }

    public function testBrowserConsoleCustomMacro(): void
    {
        MonologCov10BrowserHtml::resetStatic();
        $h = new MonologCov10BrowserHtml();
        $h->setFormatter(new LineFormatter('[[%level_name%]]{macro: custom;color: red} [[%channel%]]{font-weight: bold} %message%'));
        $h->handle($this->rec('styled'));
        ob_start();
        MonologCov10BrowserHtml::send();
        $out = ob_get_clean();
        $this->assertStringContainsString(' custom', $out);
        $this->assertStringContainsString('font-weight: bold', $out);
        MonologCov10BrowserHtml::resetStatic();
    }

    public function testChromePhpCliNoop(): void
    {
        $cli = new ChromePHPHandler();
        $cli->handle($this->rec('cli'));
        $cli->handleBatch([$this->rec('cli')]);
        $this->assertTrue(true);
    }

    public function testChromePhpOverflowAndUnaccepted(): void
    {
        $_SERVER['HTTP_USER_AGENT'] = 'Mozilla/5.0 Chrome/120.0.0.0 Safari/537.36';
        MonologCov10Chrome::resetAll();
        $h = new MonologCov10Chrome();
        $h->handle($this->rec(str_repeat('big', 2000), Level::Error));
        $data = json_decode(base64_decode($h->headers['X-ChromeLogger-Data']), true);
        $this->assertStringContainsString('Incomplete logs', json_encode($data['rows']));
        $before = $h->headers;
        $h->handle($this->rec('after overflow'));
        $this->assertSame($before, $h->headers);
        unset($_SERVER['HTTP_USER_AGENT']);
        MonologCov10Chrome::resetAll();
        $h2 = new MonologCov10Chrome();
        $h2->handle($this->rec('no ua'));
        $this->assertSame([], $h2->headers);
        $h2->handle($this->rec('second call skips'));
        MonologCov10Chrome::resetAll();
    }

    public function testUtilsJsonIgnoreErrorsWithoutPartialOutput(): void
    {
        $this->assertSame('null', Utils::jsonEncode(NAN, 0, true));
        try { Utils::jsonEncode(NAN, 0); $this->fail(); } catch (\RuntimeException $e) { $this->assertStringContainsString('Unknown error', $e->getMessage()); }
    }

    public function testUtilsExpandIniFalsePaths(): void
    {
        $this->assertFalse(Utils::expandIniShorthandBytes('abc'));
        $this->assertFalse(Utils::expandIniShorthandBytes(12));
        $this->assertFalse(Utils::expandIniShorthandBytes(null));
        $this->assertSame(-1, Utils::expandIniShorthandBytes('-1'));
        $this->assertSame(5, Utils::expandIniShorthandBytes(' 5 '));
        $this->assertSame(2048, Utils::expandIniShorthandBytes('2K'));
        $this->assertSame(3 * 1048576, Utils::expandIniShorthandBytes('3m'));
        $this->assertSame(1073741824, Utils::expandIniShorthandBytes('1G'));
    }

    public function testUtilsHandleJsonErrorArray(): void
    {
        $this->assertSame('["abc"]', Utils::handleJsonError(JSON_ERROR_UTF8, ['abc'], JSON_UNESCAPED_SLASHES));
    }

    public function testLoggerWithoutLoopDetection(): void
    {
        $h = new TestHandler();
        $l = new Logger('nl', [$h]);
        $l->useLoggingLoopDetection(false);
        $l->info('one');
        $l->warning('two');
        $this->assertCount(2, $h->getRecords());
    }

    public function testLoggerInvalidLevelString(): void
    {
        $this->expectException(\Psr\Log\InvalidArgumentException::class);
        $this->expectExceptionMessage('Level "999" is not defined');
        (new Logger('x'))->addRecord(999, 'bad');
    }

    public function testNormalizerExceptionDepth(): void
    {
        $n = new NormalizerFormatter();
        $n->setMaxNormalizeDepth(2);
        $out = $n->format($this->rec('m', Level::Info, ['e' => new \RuntimeException('deep', 0, new \LogicException('prev'))]));
        $this->assertSame(['Over 2 levels deep, aborting normalization'], $out['context']['e']['previous']);
    }

    public function testSignalHandlerCallsPreviousCallable(): void
    {
        $h = new TestHandler();
        $sh = new SignalHandler(new Logger('sig2', [$h]));
        $calls = [];
        $prev = new \ReflectionProperty($sh, 'previousSignalHandler');
        $prev->setValue($sh, [15 => function (int $signo, $info) use (&$calls) { $calls[] = [$signo, $info]; }]);
        $levels = new \ReflectionProperty($sh, 'signalLevelMap');
        $levels->setValue($sh, [15 => 'warning']);
        $sh->handleSignal(15, ['pid' => 7]);
        $this->assertSame([[15, ['pid' => 7]]], $calls);
        $this->assertSame(Level::Warning, $h->getRecords()[0]->level);
        $this->assertSame(['pid' => 7], $h->getRecords()[0]->context);
        $sh->handleSignal(99);
        $this->assertCount(2, $h->getRecords());
        $this->assertSame(Level::Critical, $h->getRecords()[1]->level);
    }

    public function testAmqpGelfMessageAndDeprecatedName(): void
    {
        $ex = new \AMQPExchange();
        $h = new AmqpHandler($ex, 'ignored-name', Level::Debug);
        $h->setFormatter(new GelfMessageFormatter('sys'));
        $h->handle($this->rec('gelf over amqp'));
        $this->assertSame('gelf over amqp', json_decode($ex->published[0][0], true)['short']);
        $ch = new \PhpAmqpLib\Channel\AMQPChannel();
        $h2 = new AmqpHandler($ch, 'logs');
        $h2->setFormatter(new GelfMessageFormatter('sys'));
        $h2->handleBatch([$this->rec('gelf batch', Level::Error)]);
        $this->assertSame('gelf batch', json_decode($ch->batch[0][0]->body, true)['short']);
    }

    public function testElasticsearchVersionSix(): void
    {
        $old = new MonologCov10Es6();
        $h = new ElasticsearchHandler($old, ['type' => 'record']);
        $this->assertSame('record', $h->getOptions()['type']);
        $h->handle($this->rec('v6'));
        $this->assertSame('record', $old->bulks[0]['body'][0]['index']['_type']);
        $this->assertSame('monolog', $old->bulks[0]['body'][0]['index']['_index']);
    }

    public function testIntrospectionSkipsConfiguredClasses(): void
    {
        $h = new TestHandler();
        $l = new Logger('intro', [$h], [new IntrospectionProcessor(Level::Debug, ['MonologCov10Test'])]);
        $l->info('x');
        $e = $h->getRecords()[0]->extra;
        $this->assertArrayHasKey('function', $e);
        $this->assertNotSame(self::class, $e['class'] ?? null);
    }

    public function testIntrospectionSkipsCallUserFunc(): void
    {
        $h = new TestHandler();
        $l = new Logger('intro2', [$h], [new IntrospectionProcessor()]);
        call_user_func([$l, 'info'], 'via call_user_func');
        $this->assertSame(self::class, $h->getRecords()[0]->extra['class']);
    }
}
