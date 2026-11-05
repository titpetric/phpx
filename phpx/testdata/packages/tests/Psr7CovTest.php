<?php
use GuzzleHttp\Psr7\{Uri, UriComparator, Utils, CachingStream, PumpStream, FnStream, StreamWrapper, HttpFactory, DroppingStream, NoSeekStream, LazyOpenStream, Stream, BufferStream, InflateStream, LimitStream, AppendStream};
use PHPUnit\Framework\TestCase;

final class Psr7CovTest extends TestCase
{
    private array $tmp = [];

    protected function tearDown(): void
    {
        foreach ($this->tmp as $f) {
            if (is_file($f)) {
                unlink($f);
            }
        }
        $this->tmp = [];
    }

    private function tmpFile(string $content = ''): string
    {
        $f = sys_get_temp_dir() . '/psr7cov_' . getmypid() . '_' . count($this->tmp);
        file_put_contents($f, $content);
        $this->tmp[] = $f;
        return $f;
    }

    public function testCachingStreamReadAndSeek(): void
    {
        $c = new CachingStream(Utils::streamFor('abcdefghij'));
        $this->assertSame('abc', $c->read(3));
        $this->assertSame(10, $c->getSize());
        $c->seek(6);
        $this->assertSame(6, $c->tell());
        $c->seek(-2, SEEK_CUR);
        $this->assertSame('efg', $c->read(3));
        $c->seek(0, SEEK_END);
        $this->assertSame(10, $c->tell());
        $c->rewind();
        $this->assertSame('abcdefghij', (string) $c);
    }

    public function testCachingStreamBadWhence(): void
    {
        $c = new CachingStream(Utils::streamFor('abc'));
        $this->expectException(\InvalidArgumentException::class);
        $this->expectExceptionMessage('Invalid whence');
        $c->seek(0, 99);
    }

    public function testCachingStreamSeekEndUnknownSize(): void
    {
        $src = new PumpStream(function ($n) {
            static $done = false;
            if ($done) {
                return false;
            }
            $done = true;
            return 'hello world';
        });
        $c = new CachingStream($src);
        $this->assertNull($c->getSize());
        $c->seek(-5, SEEK_END);
        $this->assertSame('world', $c->read(10));
    }

    public function testCachingStreamWriteAndDetachClose(): void
    {
        $c = new CachingStream(Utils::streamFor('0123456789'));
        $c->read(2);
        $this->assertSame(2, $c->write('XY'));
        $c->seek(0);
        $this->assertSame('01XY456789', $c->getContents());
        $res = $c->detach();
        $this->assertIsResource($res);
        $this->assertNull($c->detach());
        $this->assertNull($c->getSize());
        $c2 = new CachingStream(Utils::streamFor('zz'));
        $c2->close();
        $this->assertNull($c2->getSize());
    }

    public function testPumpStream(): void
    {
        $i = 0;
        $p = new PumpStream(function ($n) use (&$i) {
            return $i++ < 3 ? 'ab' : false;
        }, ['size' => 6, 'metadata' => ['k' => 'v']]);
        $this->assertSame(6, $p->getSize());
        $this->assertFalse($p->isSeekable());
        $this->assertFalse($p->isWritable());
        $this->assertTrue($p->isReadable());
        $this->assertSame(['k' => 'v'], $p->getMetadata());
        $this->assertSame('v', $p->getMetadata('k'));
        $this->assertNull($p->getMetadata('zz'));
        $this->assertSame('aba', $p->read(3));
        $this->assertSame(3, $p->tell());
        $this->assertSame('bab', $p->getContents());
        $this->assertTrue($p->eof());
        $this->assertSame('', $p->read(2));
        $p->close();
        $this->assertSame(0, $p->tell());
    }

    public function testPumpStreamToStringAndErrors(): void
    {
        $n = 0;
        $p = new PumpStream(function () use (&$n) {
            return $n++ === 0 ? 'xyz' : null;
        });
        $this->assertSame('xyz', (string) $p);
        $p2 = new PumpStream(fn () => 'a');
        try {
            $p2->seek(0);
            $this->fail();
        } catch (\RuntimeException $e) {
            $this->assertSame('Cannot seek a PumpStream', $e->getMessage());
        }
        try {
            $p2->rewind();
            $this->fail();
        } catch (\RuntimeException $e) {
            $this->assertSame('Cannot seek a PumpStream', $e->getMessage());
        }
        try {
            $p2->write('x');
            $this->fail();
        } catch (\RuntimeException $e) {
            $this->assertSame('Cannot write to a PumpStream', $e->getMessage());
        }
        $this->assertNull($p2->detach());
        $thrower = new PumpStream(function () {
            throw new \LogicException('boom');
        });
        $this->expectException(\LogicException::class);
        (string) $thrower;
    }

    public function testFnStream(): void
    {
        $closed = false;
        $s = new FnStream([
            '__toString' => fn () => 'str',
            'close' => function () use (&$closed) { $closed = true; },
            'detach' => fn () => null,
            'getSize' => fn () => 12,
            'tell' => fn () => 3,
            'eof' => fn () => false,
            'isSeekable' => fn () => true,
            'rewind' => function () {},
            'seek' => function ($o, $w) {},
            'isWritable' => fn () => true,
            'write' => fn ($x) => strlen($x),
            'isReadable' => fn () => true,
            'read' => fn ($n) => str_repeat('r', $n),
            'getContents' => fn () => 'contents',
            'getMetadata' => fn ($k) => $k === null ? ['a'] : "meta-$k",
        ]);
        $this->assertSame('str', (string) $s);
        $this->assertSame(12, $s->getSize());
        $this->assertSame(3, $s->tell());
        $this->assertFalse($s->eof());
        $this->assertTrue($s->isSeekable());
        $s->rewind();
        $s->seek(1);
        $this->assertTrue($s->isWritable());
        $this->assertSame(2, $s->write('ab'));
        $this->assertTrue($s->isReadable());
        $this->assertSame('rrr', $s->read(3));
        $this->assertSame('contents', $s->getContents());
        $this->assertSame(['a'], $s->getMetadata());
        $this->assertSame('meta-x', $s->getMetadata('x'));
        $this->assertNull($s->detach());
        $s->close();
        $this->assertTrue($closed);
    }

    public function testFnStreamMissingSlot(): void
    {
        $s = new FnStream([]);
        $this->expectException(\BadMethodCallException::class);
        $this->expectExceptionMessage('getSize() is not implemented in the FnStream');
        $s->getSize();
    }

    public function testFnStreamDecorate(): void
    {
        $inner = Utils::streamFor('hello');
        $d = FnStream::decorate($inner, ['getSize' => fn () => 99]);
        $this->assertSame(99, $d->getSize());
        $this->assertSame('he', $d->read(2));
        $this->assertSame(2, $d->tell());
        $this->assertSame('hello', (string) $d);
    }

    public function testFnStreamUnserializeRefused(): void
    {
        $s = new FnStream([]);
        $this->expectException(\LogicException::class);
        $s->__wakeup();
    }

    public function testStreamWrapper(): void
    {
        $s = Utils::streamFor('line1' . "\n" . 'line2');
        $r = StreamWrapper::getResource($s);
        $this->assertIsResource($r);
        $this->assertSame("line1\n", fgets($r));
        $this->assertSame('line2', fread($r, 100));
        $this->assertTrue(feof($r));
        $st = fstat($r);
        $this->assertSame(11, $st['size']);
        fclose($r);
        $this->assertContains('guzzle', stream_get_wrappers());
    }

    public function testStreamWrapperWriteSeek(): void
    {
        $s = new Stream(fopen('php://temp', 'r+'));
        $r = StreamWrapper::getResource($s);
        $this->assertSame(3, fwrite($r, 'abc'));
        $this->assertSame(0, fseek($r, 1));
        $this->assertSame('bc', fread($r, 10));
        $this->assertSame(3, ftell($r));
        fclose($r);
    }

    public function testStreamWrapperUnusable(): void
    {
        $s = FnStream::decorate(Utils::streamFor('x'), [
            'isReadable' => fn () => false,
            'isWritable' => fn () => false,
        ]);
        $this->expectException(\InvalidArgumentException::class);
        StreamWrapper::getResource($s);
    }

    public function testStreamWrapperUrlStat(): void
    {
        StreamWrapper::register();
        StreamWrapper::register();
        $ctx = StreamWrapper::createStreamContext(Utils::streamFor('abc'));
        $this->assertIsResource($ctx);
        $st = @stat('guzzle://stream');
        $this->assertSame(0, $st['size']);
    }

    public function testStreamWrapperWriteOnly(): void
    {
        $s = FnStream::decorate(new Stream(fopen('php://temp', 'r+')), [
            'isReadable' => fn () => false,
        ]);
        $r = StreamWrapper::getResource($s);
        $this->assertSame(2, fwrite($r, 'hi'));
        fclose($r);
    }

    public function testHttpFactory(): void
    {
        $f = new HttpFactory();
        $this->assertSame('abc', (string) $f->createStream('abc'));
        $file = $this->tmpFile('filedata');
        $this->assertSame('filedata', (string) $f->createStreamFromFile($file));
        $this->assertSame('rs', (string) $f->createStreamFromResource(fopen('data://text/plain,rs', 'r')));
        $u = $f->createUri('http://h/p?q=1');
        $this->assertSame('h', $u->getHost());
        $req = $f->createRequest('POST', 'http://h/x');
        $this->assertSame('POST', $req->getMethod());
        $resp = $f->createResponse(404, 'Nope');
        $this->assertSame(404, $resp->getStatusCode());
        $this->assertSame('Nope', $resp->getReasonPhrase());
        $this->assertSame('Not Found', $f->createResponse(404)->getReasonPhrase());
        $sr = $f->createServerRequest('GET', '/a', ['X' => 1]);
        $this->assertSame(['X' => 1], $sr->getServerParams());
        $sr2 = $f->createServerRequest('', '/a', ['REQUEST_METHOD' => 'PUT']);
        $this->assertSame('PUT', $sr2->getMethod());
        $uf = $f->createUploadedFile($f->createStream('12345'));
        $this->assertSame(5, $uf->getSize());
        $uf = $f->createUploadedFile($f->createStream('12345'), 3, UPLOAD_ERR_OK, 'n.txt', 'text/plain');
        $this->assertSame(3, $uf->getSize());
        $this->assertSame('n.txt', $uf->getClientFilename());
        $this->assertSame('text/plain', $uf->getClientMediaType());
    }

    public function testHttpFactoryErrors(): void
    {
        $f = new HttpFactory();
        try {
            $f->createServerRequest('', '/a');
            $this->fail();
        } catch (\InvalidArgumentException $e) {
            $this->assertSame('Cannot determine HTTP method', $e->getMessage());
        }
        try {
            $f->createStreamFromFile($this->tmpFile('x'), 'z');
            $this->fail();
        } catch (\InvalidArgumentException $e) {
            $this->assertSame('Invalid file opening mode "z"', $e->getMessage());
        }
        try {
            $f->createStreamFromFile($this->tmpFile('x'), '');
            $this->fail();
        } catch (\InvalidArgumentException $e) {
            $this->assertStringContainsString('Invalid file opening mode', $e->getMessage());
        }
        try {
            $f->createStreamFromFile(sys_get_temp_dir() . '/psr7cov_missing_' . getmypid(), 'r');
            $this->fail();
        } catch (\RuntimeException $e) {
            $this->assertStringContainsString('Unable to open', $e->getMessage());
        }
    }

    public function testUriComparator(): void
    {
        $a = new Uri('http://example.com/a');
        $this->assertFalse(UriComparator::isCrossOrigin($a, new Uri('http://EXAMPLE.com:80/b')));
        $this->assertTrue(UriComparator::isCrossOrigin($a, new Uri('http://other.com/a')));
        $this->assertTrue(UriComparator::isCrossOrigin($a, new Uri('https://example.com/a')));
        $this->assertTrue(UriComparator::isCrossOrigin($a, new Uri('http://example.com:8080/a')));
        $this->assertFalse(UriComparator::isCrossOrigin(new Uri('https://e.com'), new Uri('https://e.com:443')));
        $this->assertFalse(UriComparator::isCrossOrigin(new Uri('ftp://e.com'), new Uri('ftp://e.com')));
        $this->assertFalse(UriComparator::isCrossOrigin(new Uri('ftp://e.com'), new Uri('ftp://e.com:21')));
    }

    public function testDroppingStream(): void
    {
        $inner = new BufferStream(100);
        $d = new DroppingStream($inner, 5);
        $this->assertSame(3, $d->write('abc'));
        $this->assertSame(2, $d->write('defgh'));
        $this->assertSame(0, $d->write('zzz'));
        $this->assertSame('abcde', (string) $d);
    }

    public function testNoSeekStream(): void
    {
        $n = new NoSeekStream(Utils::streamFor('abc'));
        $this->assertFalse($n->isSeekable());
        $this->assertSame('abc', (string) $n);
        $this->expectException(\RuntimeException::class);
        $this->expectExceptionMessage('Cannot seek a NoSeekStream');
        $n->seek(0);
    }

    public function testLazyOpenStream(): void
    {
        $f = $this->tmpFile('lazy-data');
        $l = new LazyOpenStream($f, 'r');
        $this->assertSame(9, $l->getSize());
        $this->assertSame('lazy', $l->read(4));
        $this->assertSame('-data', $l->getContents());
        $l->close();
        $w = new LazyOpenStream($f, 'w');
        $w->write('new');
        $w->close();
        $this->assertSame('new', file_get_contents($f));
    }

    public function testStreamDecoratorMethods(): void
    {
        $d = new NoSeekStream(Utils::streamFor('hello world'));
        $this->assertTrue($d->isReadable());
        $this->assertTrue($d->isWritable());
        $this->assertSame(11, $d->getSize());
        $this->assertSame(0, $d->tell());
        $this->assertFalse($d->eof());
        $this->assertSame('hello', $d->read(5));
        $this->assertSame(' world', $d->getContents());
        $this->assertTrue($d->eof());
        $this->assertSame('php://temp', $d->getMetadata('uri'));
        $this->assertIsArray($d->getMetadata());
        $this->assertSame(3, $d->write('abc'));
        $this->assertIsResource($d->detach());
        $d->close();
    }

    public function testDecoratorCallProxy(): void
    {
        $inner = Utils::streamFor('x');
        $f = FnStream::decorate($inner, []);
        $this->assertSame($inner->getSize(), $f->getSize());
        $l = new LazyOpenStream($this->tmpFile('q'), 'r');
        $this->assertSame('q', (string) $l);
        $this->assertTrue($l->isSeekable());
        $l->rewind();
        $l->seek(0, SEEK_END);
        $this->assertTrue($l->eof() || $l->tell() === 1);
    }

    public function testInflateStream(): void
    {
        $gz = gzcompress('inflate me please');
        $s = new InflateStream(Utils::streamFor($gz));
        $this->assertSame('inflate me please', (string) $s);
    }
}
