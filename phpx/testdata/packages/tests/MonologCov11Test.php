<?php
require_once __DIR__ . '/MonologCovStubSoap.php';

use Monolog\{Logger, Level, LogRecord};
use Monolog\Handler\{TestHandler, DeduplicationHandler};
use Monolog\Formatter\{LineFormatter, NormalizerFormatter, RedactingFormatter};
use PHPUnit\Framework\TestCase;

final class MonologCov11Test extends TestCase
{
    private string $dir;
    private string $cwd;

    protected function setUp(): void
    {
        $this->cwd = getcwd();
        $this->dir = sys_get_temp_dir() . '/monologcov11-' . getmypid();
        mkdir($this->dir);
    }

    protected function tearDown(): void
    {
        chdir($this->cwd);
        foreach (glob($this->dir . '/*') ?: [] as $f) { unlink($f); }
        rmdir($this->dir);
    }

    private function rec(string $m = 'm', Level $l = Level::Info, array $ctx = [], ?\DateTimeImmutable $dt = null): LogRecord
    {
        return new LogRecord($dt ?? new \DateTimeImmutable(), 'chan', $l, $m, $ctx);
    }

    public function testDeduplicationWithRelativeStore(): void
    {
        chdir($this->dir);
        $t = new TestHandler();
        $d = new DeduplicationHandler($t, 'dedup.log', Level::Error, 60);
        $d->handle($this->rec("same\nsecond line", Level::Error));
        $d->handle($this->rec('info', Level::Info));
        $d->flush();
        $this->assertCount(2, $t->getRecords());
        $this->assertFileExists($this->dir . '/dedup.log');
        $d->handle($this->rec('same', Level::Error));
        $d->close();
        $this->assertCount(2, $t->getRecords());
        $d->handle($this->rec('different', Level::Error));
        $d->handle($this->rec('different', Level::Error));
        $d->flush();
        $this->assertCount(4, $t->getRecords());
    }

    public function testDeduplicationGarbageCollection(): void
    {
        chdir($this->dir);
        file_put_contents('dedup.log', "1000:ERROR:old\nbadline\n" . time() . ":ERROR:fresh\n");
        $t = new TestHandler();
        $d = new DeduplicationHandler($t, 'dedup.log', Level::Error, 60);
        $d->handle($this->rec('new', Level::Error));
        $d->flush();
        $this->assertCount(1, $t->getRecords());
        $contents = file_get_contents('dedup.log');
        $this->assertStringNotContainsString('1000:ERROR:old', $contents);
        $this->assertStringContainsString('fresh', $contents);
        $this->assertStringContainsString(':ERROR:new', $contents);
    }

    public function testDeduplicationSkipsKnownDuplicate(): void
    {
        chdir($this->dir);
        $now = new \DateTimeImmutable();
        file_put_contents('dedup.log', $now->getTimestamp() . ":ERROR:dup message\n");
        $t = new TestHandler();
        $d = new DeduplicationHandler($t, 'dedup.log', Level::Error, 60);
        $d->handle($this->rec('dup message', Level::Error, [], $now));
        $d->flush();
        $this->assertCount(0, $t->getRecords());
        $d->handle($this->rec('dup message', Level::Warning, [], $now));
        $d->flush();
        $this->assertCount(1, $t->getRecords());
    }

    public function testLineFormatterSoapFault(): void
    {
        $l = new LineFormatter(null, null, true, true);
        $out = $l->format($this->rec('m', Level::Error, ['exception' => new SoapFault('Server.Code', 'boom', 'actor-uri', 'detail text')]));
        $this->assertStringContainsString('SoapFault(code: 0 faultcode: Server.Code faultactor: actor-uri detail: detail text): boom', $out);
        $out = $l->format($this->rec('m', Level::Error, ['exception' => new SoapFault('C', 'boom', null, ['k' => 'v'])]));
        $this->assertStringContainsString('faultcode: C detail: {\\"k\\":\\"v\\"}', $out);
        $this->assertStringNotContainsString('faultactor', $out);
    }

    public function testNormalizerSoapFault(): void
    {
        $n = new NormalizerFormatter();
        $out = $n->format($this->rec('m', Level::Error, ['e' => new SoapFault('Client', 'bad', 'actor', 'detail string')]));
        $this->assertSame('Client', $out['context']['e']['faultcode']);
        $this->assertSame('actor', $out['context']['e']['faultactor']);
        $this->assertSame('detail string', $out['context']['e']['detail']);
        $out = $n->format($this->rec('m', Level::Error, ['e' => new SoapFault('Client', 'bad', null, ['x' => 1])]));
        $this->assertSame('{"x":1}', $out['context']['e']['detail']);
        $this->assertArrayNotHasKey('faultactor', $out['context']['e']);
    }

    public function testRedactingPatternFailureDropsRecord(): void
    {
        $f = new RedactingFormatter(new LineFormatter("%message%\n"), [], ['/(a+)+$/']);
        $old = ini_get('pcre.backtrack_limit');
        ini_set('pcre.backtrack_limit', '100');
        $oldJit = ini_get('pcre.jit');
        ini_set('pcre.jit', '0');
        try {
            $out = $f->format($this->rec(str_repeat('a', 40) . 'b'));
        } finally {
            ini_set('pcre.backtrack_limit', $old);
            ini_set('pcre.jit', $oldJit);
        }
        $this->assertStringStartsWith('[RedactingFormatter dropped this record', $out);
    }

    public function testLoggerInsideFiber(): void
    {
        $h = new TestHandler();
        $l = new Logger('fiber', [$h]);
        $fiber = new \Fiber(function () use ($l) {
            $l->info('before suspend');
            \Fiber::suspend(1);
            $l->warning('after resume');
        });
        $this->assertSame(1, $fiber->start());
        $fiber->resume();
        $this->assertCount(2, $h->getRecords());
        $this->assertSame('after resume', $h->getRecords()[1]->message);
    }
}
