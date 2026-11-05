<?php
use GuzzleHttp\Psr7\{Utils, Stream, LimitStream, AppendStream, NoSeekStream, PumpStream, FnStream};
use PHPUnit\Framework\TestCase;

final class Psr7Cov4Test extends TestCase
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

    public function testAppendStreamReadAcrossStreams(): void
    {
        $a = new AppendStream([Utils::streamFor('abc'), Utils::streamFor('def')]);
        $a->addStream(Utils::streamFor('ghi'));
        $this->assertTrue($a->isReadable());
        $this->assertFalse($a->isWritable());
        $this->assertTrue($a->isSeekable());
        $this->assertSame(9, $a->getSize());
        $this->assertSame('abcd', $a->read(4));
        $this->assertSame(4, $a->tell());
        $this->assertFalse($a->eof());
        $this->assertSame('efghi', $a->read(100));
        $this->assertTrue($a->eof());
        $this->assertSame('', $a->read(5));
        $this->assertSame(9, $a->tell());
    }

    public function testAppendStreamSeek(): void
    {
        $a = new AppendStream([Utils::streamFor('abc'), Utils::streamFor('def')]);
        $a->seek(4);
        $this->assertSame(4, $a->tell());
        $this->assertSame('ef', $a->getContents());
        $a->rewind();
        $this->assertSame('abcdef', (string) $a);
        $a->seek(100);
        $this->assertTrue($a->eof());
        $this->assertSame('', $a->read(1));
    }

    public function testAppendStreamEmpty(): void
    {
        $a = new AppendStream();
        $this->assertSame('', $a->read(5));
        $this->assertSame(0, $a->getSize());
        $this->assertSame('', (string) $a);
        $this->assertTrue($a->eof());
    }

    public function testAppendStreamErrors(): void
    {
        $a = new AppendStream();
        $unreadable = FnStream::decorate(Utils::streamFor('x'), ['isReadable' => fn () => false]);
        try {
            $a->addStream($unreadable);
            $this->fail();
        } catch (\InvalidArgumentException $e) {
            $this->assertSame('Each stream must be readable', $e->getMessage());
        }
        try {
            $a->write('x');
            $this->fail();
        } catch (\RuntimeException $e) {
            $this->assertSame('Cannot write to an AppendStream', $e->getMessage());
        }
        $b = new AppendStream([Utils::streamFor('abc')]);
        try {
            $b->seek(0, SEEK_END);
            $this->fail();
        } catch (\RuntimeException $e) {
            $this->assertSame('The AppendStream can only seek with SEEK_SET', $e->getMessage());
        }
    }

    public function testAppendStreamNotSeekable(): void
    {
        $a = new AppendStream([Utils::streamFor('abc'), new NoSeekStream(Utils::streamFor('def'))]);
        $this->assertFalse($a->isSeekable());
        $this->assertSame('abcdef', $a->getContents());
        try {
            $a->seek(0);
            $this->fail();
        } catch (\RuntimeException $e) {
            $this->assertSame('This AppendStream is not seekable', $e->getMessage());
        }
        try {
            $a->rewind();
            $this->fail();
        } catch (\RuntimeException $e) {
            $this->assertSame('This AppendStream is not seekable', $e->getMessage());
        }
    }

    public function testAppendStreamSeekFailureWrapped(): void
    {
        $bad = FnStream::decorate(Utils::streamFor('abc'), ['rewind' => function () {
            throw new \RuntimeException('nope');
        }]);
        $a = new AppendStream([$bad]);
        try {
            $a->seek(0);
            $this->fail();
        } catch (\RuntimeException $e) {
            $this->assertSame('Unable to seek stream 0 of the AppendStream', $e->getMessage());
            $this->assertSame('nope', $e->getPrevious()->getMessage());
        }
    }

    public function testAppendStreamUnknownSizeAndMetadata(): void
    {
        $p = new PumpStream(fn () => false);
        $a = new AppendStream([Utils::streamFor('ab'), $p]);
        $this->assertNull($a->getSize());
        $this->assertSame([], $a->getMetadata());
        $this->assertNull($a->getMetadata('x'));
    }

    public function testAppendStreamCloseDetach(): void
    {
        $inner = Utils::streamFor('abc');
        $a = new AppendStream([$inner]);
        $a->read(2);
        $this->assertNull($a->detach());
        $this->assertSame(0, $a->tell());
        $this->assertSame(0, $a->getSize());
        $this->assertSame('', $a->read(1));
        $b = new AppendStream([$inner = Utils::streamFor('xyz')]);
        $b->close();
        $this->assertNull($inner->getSize());
        $this->assertSame(0, $b->tell());
    }

    public function testAppendStreamToStringError(): void
    {
        $bad = FnStream::decorate(Utils::streamFor('abc'), ['read' => function () {
            throw new \LogicException('read failed');
        }, 'eof' => fn () => false]);
        $a = new AppendStream([$bad]);
        $this->expectException(\LogicException::class);
        (string) $a;
    }

    public function testAppendStreamSkipsEmptyMiddleStream(): void
    {
        $a = new AppendStream([Utils::streamFor('ab'), Utils::streamFor(''), Utils::streamFor('cd')]);
        $this->assertSame('abcd', $a->read(10));
    }

    public function testLimitStream(): void
    {
        $l = new LimitStream(Utils::streamFor('0123456789'), 4, 2);
        $this->assertSame(4, $l->getSize());
        $this->assertSame(0, $l->tell());
        $this->assertFalse($l->eof());
        $this->assertSame('23', $l->read(2));
        $this->assertSame(2, $l->tell());
        $this->assertSame('45', $l->read(100));
        $this->assertTrue($l->eof());
        $this->assertSame('', $l->read(5));
        $l->seek(1);
        $this->assertSame('34', $l->read(2));
        $l->seek(100);
        $this->assertSame('', $l->read(1));
        $this->assertSame(4, $l->tell());
        $l->rewind();
        $this->assertSame('2345', (string) $l);
    }

    public function testLimitStreamUnlimited(): void
    {
        $l = new LimitStream(Utils::streamFor('0123456789'), -1, 3);
        $this->assertSame(7, $l->getSize());
        $this->assertFalse($l->eof());
        $this->assertSame('3456789', $l->read(100));
        $this->assertTrue($l->eof());
        $l->setLimit(2);
        $l->setOffset(1);
        $this->assertSame(2, $l->getSize());
        $this->assertSame('12', $l->getContents());
    }

    public function testLimitStreamSizeClamped(): void
    {
        $this->assertSame(0, (new LimitStream(Utils::streamFor('abc'), 5, 3))->getSize());
        $this->assertSame(1, (new LimitStream(Utils::streamFor('abc'), 5, 2))->getSize());
        $unknown = new LimitStream(new PumpStream(fn () => false), 5, 0);
        $this->assertNull($unknown->getSize());
    }

    public function testLimitStreamBadSeek(): void
    {
        $l = new LimitStream(Utils::streamFor('abcdef'), 3, 1);
        foreach ([[-1, SEEK_SET], [0, SEEK_CUR], [0, SEEK_END]] as [$o, $w]) {
            try {
                $l->seek($o, $w);
                $this->fail();
            } catch (\RuntimeException $e) {
                $this->assertSame("Cannot seek to offset $o with whence $w", $e->getMessage());
            }
        }
    }

    public function testLimitStreamNonSeekableOffset(): void
    {
        $l = new LimitStream(new NoSeekStream(Utils::streamFor('0123456789')), 3, 4);
        $this->assertSame('456', (string) $l);
        $src = new NoSeekStream(Utils::streamFor('0123456789'));
        $src->read(5);
        $this->expectException(\RuntimeException::class);
        $this->expectExceptionMessage('Could not seek to stream offset 2');
        new LimitStream($src, 3, 2);
    }

    public function testStreamErrors(): void
    {
        try {
            new Stream('notaresource');
            $this->fail();
        } catch (\InvalidArgumentException $e) {
            $this->assertSame('Stream must be a resource', $e->getMessage());
        }
        $s = new Stream(fopen('php://temp', 'r+'));
        $s->write('abc');
        try {
            $s->read(-1);
            $this->fail();
        } catch (\RuntimeException $e) {
            $this->assertSame('Length parameter cannot be negative', $e->getMessage());
        }
        $this->assertSame('', $s->read(0));
        $s->close();
        $this->assertNull($s->getSize());
        $this->assertSame([], $s->getMetadata());
        $this->assertNull($s->getMetadata('uri'));
        $this->assertNull($s->detach());
        $this->assertFalse($s->isReadable());
        $this->assertFalse($s->isWritable());
        $this->assertFalse($s->isSeekable());
        foreach (['tell' => [], 'eof' => [], 'rewind' => [], 'seek' => [0], 'read' => [1], 'write' => ['x'], 'getContents' => []] as $m => $args) {
            try {
                $s->$m(...$args);
                $this->fail("$m did not throw");
            } catch (\RuntimeException $e) {
                $this->assertSame('Stream is detached', $e->getMessage(), $m);
            }
        }
        try {
            (string) $s;
            $this->fail();
        } catch (\RuntimeException $e) {
            $this->assertSame('Stream is detached', $e->getMessage());
        }
    }

    public function testStreamReadOnlyWriteOnly(): void
    {
        $f = sys_get_temp_dir() . '/psr7cov4_' . getmypid();
        $this->tmp[] = $f;
        file_put_contents($f, 'content');
        $ro = new Stream(fopen($f, 'r'));
        $this->assertTrue($ro->isReadable());
        $this->assertFalse($ro->isWritable());
        $this->assertSame(7, $ro->getSize());
        try {
            $ro->write('x');
            $this->fail();
        } catch (\RuntimeException $e) {
            $this->assertSame('Cannot write to a non-writable stream', $e->getMessage());
        }
        $this->assertSame($f, $ro->getMetadata('uri'));
        $this->assertSame('r', $ro->getMetadata('mode'));
        $this->assertIsArray($ro->getMetadata());
        $ro->close();

        $wo = new Stream(fopen($f, 'w'));
        $this->assertFalse($wo->isReadable());
        try {
            $wo->read(1);
            $this->fail();
        } catch (\RuntimeException $e) {
            $this->assertSame('Cannot read from non-readable stream', $e->getMessage());
        }
        try {
            (string) $wo;
            $this->fail();
        } catch (\RuntimeException $e) {
            $this->assertSame('Cannot read from non-readable stream', $e->getMessage());
        }
        try {
            $wo->getContents();
            $this->fail();
        } catch (\RuntimeException $e) {
            $this->assertSame('Cannot read from non-readable stream', $e->getMessage());
        }
        $this->assertSame(3, $wo->write('abc'));
        $this->assertSame(3, $wo->getSize());
        $wo->close();
    }

    public function testStreamOptionsAndSeek(): void
    {
        $r = fopen('php://temp', 'r+');
        fwrite($r, 'hello');
        $s = new Stream($r, ['size' => 42, 'metadata' => ['custom' => 'yes']]);
        $this->assertSame(42, $s->getSize());
        $this->assertSame('yes', $s->getMetadata('custom'));
        $this->assertSame('yes', $s->getMetadata()['custom']);
        $this->assertSame(5, $s->tell());
        $s->write('!');
        $this->assertSame(6, $s->getSize());
        try {
            $s->seek(-100);
            $this->fail();
        } catch (\RuntimeException $e) {
            $this->assertStringContainsString('Unable to seek to stream position -100 with whence 0', $e->getMessage());
        }
        $s->seek(-1, SEEK_END);
        $this->assertSame('!', $s->getContents());
        $this->assertTrue($s->eof());
        $this->assertSame($r, $s->detach());
        fclose($r);
    }

    public function testStreamNotSeekable(): void
    {
        $p = fopen('php://output', 'w');
        $s = new Stream($p);
        $this->assertFalse($s->isSeekable());
        try {
            $s->seek(0);
            $this->fail();
        } catch (\RuntimeException $e) {
            $this->assertSame('Stream is not seekable', $e->getMessage());
        }
        $s->detach();
        fclose($p);
    }
}
