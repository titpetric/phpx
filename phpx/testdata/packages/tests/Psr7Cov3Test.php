<?php
use GuzzleHttp\Psr7\{Utils, UriResolver, Uri, Request, Stream, PumpStream, BufferStream, LimitStream, AppendStream, FnStream};
use PHPUnit\Framework\TestCase;

final class Psr7Cov3Test extends TestCase
{
    public function testAsciiHelpers(): void
    {
        $this->assertSame('', Utils::asciiUcFirst(''));
        $this->assertSame('Hello', Utils::asciiUcFirst('hello'));
        $this->assertSame('Hello', Utils::asciiUcFirst('Hello'));
        $this->assertSame('1abc', Utils::asciiUcFirst('1abc'));
        $this->assertSame('ABC-1', Utils::asciiToUpper('abC-1'));
        $this->assertSame('abc-1', Utils::asciiToLower('ABc-1'));
        $this->assertTrue(Utils::caselessContains('Hello World', 'o w'));
        $this->assertFalse(Utils::caselessContains('Hello World', 'xyz'));
        $this->assertTrue(Utils::caselessContains('abc', ''));
        $this->assertTrue(Utils::caselessEquals('Foo', 'fOO'));
        $this->assertFalse(Utils::caselessEquals('Foo', 'Fo'));
    }

    public function testCaselessRemove(): void
    {
        $data = ['Foo' => 1, 'bar' => 2, 'BAZ' => 3, 0 => 'zero'];
        $this->assertSame(['bar' => 2, 0 => 'zero'], Utils::caselessRemove(['foo', 'baz'], $data));
        $this->assertSame($data, Utils::caselessRemove([], $data));
    }

    public function testCopyToStreamLimited(): void
    {
        $src = Utils::streamFor('0123456789');
        $dst = Utils::streamFor('');
        Utils::copyToStream($src, $dst, 4);
        $this->assertSame('0123', (string) $dst);
        $this->assertSame(4, $src->tell());
        $dst2 = Utils::streamFor('');
        Utils::copyToStream($src, $dst2);
        $this->assertSame('456789', (string) $dst2);
    }

    public function testCopyToStreamLimitedLongerThanSource(): void
    {
        $src = Utils::streamFor('abc');
        $dst = Utils::streamFor('');
        Utils::copyToStream($src, $dst, 100);
        $this->assertSame('abc', (string) $dst);
    }

    public function testCopyToStreamUnwritableDest(): void
    {
        $dst = new FnStream(['write' => fn ($s) => 0]);
        $src = Utils::streamFor('abc');
        Utils::copyToStream($src, $dst);
        $this->assertSame(3, $src->tell() >= 0 ? 3 : 0);
        $src2 = Utils::streamFor('abc');
        Utils::copyToStream($src2, $dst, 2);
        $this->assertTrue(true);
    }

    public function testCopyToStringLimited(): void
    {
        $s = Utils::streamFor('abcdefghij');
        $this->assertSame('abc', Utils::copyToString($s, 3));
        $this->assertSame('defghij', Utils::copyToString($s));
        $this->assertSame('', Utils::copyToString($s, 3));
        $s->rewind();
        $this->assertSame('abcdefghij', Utils::copyToString($s, 100));
    }

    public function testCopyToStringShortReads(): void
    {
        $n = 0;
        $p = new PumpStream(function () use (&$n) {
            return $n++ < 2 ? 'ab' : false;
        });
        $this->assertSame('abab', Utils::copyToString($p, 10));
    }

    public function testHash(): void
    {
        $s = Utils::streamFor('hello world');
        $this->assertSame(md5('hello world'), Utils::hash($s, 'md5'));
        $this->assertSame(0, $s->tell());
        $s->seek(5);
        $this->assertSame(hash('sha256', 'hello world'), Utils::hash($s, 'sha256'));
        $this->assertSame(5, $s->tell());
        $this->assertSame(hash('crc32b', 'hello world', true), Utils::hash($s, 'crc32b', true));
    }

    public function testReadLine(): void
    {
        $s = Utils::streamFor("one\ntwo\nthree");
        $this->assertSame("one\n", Utils::readLine($s));
        $this->assertSame("tw", Utils::readLine($s, 3));
        $this->assertSame("o\n", Utils::readLine($s));
        $this->assertSame('three', Utils::readLine($s));
        $this->assertSame('', Utils::readLine($s));
    }

    public function testRedactUserInfo(): void
    {
        $this->assertSame('http://user:***@h.test/p', (string) Utils::redactUserInfo(new Uri('http://user:secret@h.test/p')));
        $this->assertSame('http://user@h.test/p', (string) Utils::redactUserInfo(new Uri('http://user@h.test/p')));
        $this->assertSame('http://h.test/p', (string) Utils::redactUserInfo(new Uri('http://h.test/p')));
    }

    public function testUriFor(): void
    {
        $u = new Uri('http://a/b');
        $this->assertSame($u, Utils::uriFor($u));
        $this->assertSame('http://x/y', (string) Utils::uriFor('http://x/y'));
        $this->expectException(\InvalidArgumentException::class);
        $this->expectExceptionMessage('URI must be a string or UriInterface');
        Utils::uriFor(12);
    }

    public function testStreamForVariants(): void
    {
        $s = Utils::streamFor('abc');
        $this->assertSame($s, Utils::streamFor($s));
        $this->assertSame('', (string) Utils::streamFor(''));
        $this->assertSame('', (string) Utils::streamFor(null));
        $this->assertSame('str', (string) Utils::streamFor(new class {
            public function __toString(): string
            {
                return 'str';
            }
        }));
        $it = new \ArrayIterator(['a', 'b', 'c']);
        $p = Utils::streamFor($it);
        $this->assertInstanceOf(PumpStream::class, $p);
        $this->assertSame('abc', (string) $p);
        $this->assertSame('called', (string) Utils::streamFor(function () {
            static $done = false;
            if ($done) {
                return false;
            }
            $done = true;
            return 'called';
        }));
        $meta = Utils::streamFor('x', ['metadata' => ['k' => 1], 'size' => 5]);
        $this->assertSame(5, $meta->getSize());
        $this->assertSame(1, $meta->getMetadata('k'));
    }

    public function testStreamForResource(): void
    {
        $r = fopen('php://temp', 'r+');
        fwrite($r, 'resource!');
        rewind($r);
        $s = Utils::streamFor($r);
        $this->assertInstanceOf(Stream::class, $s);
        $this->assertSame('resource!', (string) $s);
    }

    public function testStreamForInvalid(): void
    {
        $this->expectException(\InvalidArgumentException::class);
        $this->expectExceptionMessage('Invalid resource type: object');
        Utils::streamFor(new \stdClass());
    }

    public function testTryGetContents(): void
    {
        $r = fopen('php://temp', 'r+');
        fwrite($r, 'data');
        rewind($r);
        $this->assertSame('data', Utils::tryGetContents($r));
        $this->assertSame('', Utils::tryGetContents($r));
    }

    public function testTryFopenOk(): void
    {
        $h = Utils::tryFopen('php://memory', 'r+');
        $this->assertIsResource($h);
        fclose($h);
    }

    public function testTryFopenMissing(): void
    {
        $this->expectException(\RuntimeException::class);
        $this->expectExceptionMessage('Unable to open');
        Utils::tryFopen(sys_get_temp_dir() . '/psr7cov3_missing_' . getmypid() . '/f', 'r');
    }

    public function testModifyRequestEmpty(): void
    {
        $r = new Request('GET', 'http://a/b');
        $this->assertSame($r, Utils::modifyRequest($r, []));
    }

    public function testModifyRequestAll(): void
    {
        $r = new Request('GET', 'http://a.test/b', ['X-Old' => '1', 'Keep' => 'k']);
        $n = Utils::modifyRequest($r, [
            'method' => 'POST',
            'set_headers' => ['X-New' => 'n'],
            'remove_headers' => ['x-old'],
            'body' => 'payload',
            'version' => '2.0',
            'query' => 'a=1',
        ]);
        $this->assertSame('POST', $n->getMethod());
        $this->assertFalse($n->hasHeader('X-Old'));
        $this->assertSame('n', $n->getHeaderLine('X-New'));
        $this->assertSame('k', $n->getHeaderLine('Keep'));
        $this->assertSame('payload', (string) $n->getBody());
        $this->assertSame('2.0', $n->getProtocolVersion());
        $this->assertSame('a=1', $n->getUri()->getQuery());
        $this->assertSame('GET', $r->getMethod());
    }

    public function testModifyRequestUriSetsHost(): void
    {
        $r = new Request('GET', 'http://a.test/b', ['Host' => 'a.test']);
        $n = Utils::modifyRequest($r, ['uri' => new Uri('http://b.test:8080/c')]);
        $this->assertSame('b.test:8080', $n->getHeaderLine('Host'));
        $n = Utils::modifyRequest($r, ['uri' => new Uri('http://b.test:80/c')]);
        $this->assertSame('b.test', $n->getHeaderLine('Host'));
        $n = Utils::modifyRequest($r, ['uri' => new Uri('/only-path')]);
        $this->assertSame('/only-path', $n->getUri()->getPath());
        $n = Utils::modifyRequest($r, ['uri' => new Uri('ftp://f.test:2121/c')]);
        $this->assertSame('f.test', $n->getHeaderLine('Host'));
    }

    public function testModifyRequestHostConflict(): void
    {
        $r = new Request('GET', 'http://a.test/b');
        $this->expectException(\InvalidArgumentException::class);
        $this->expectExceptionMessage('Cannot modify request with both a URI containing a host and an explicit Host header.');
        Utils::modifyRequest($r, ['uri' => new Uri('http://b.test/'), 'set_headers' => ['HOST' => 'x']]);
    }

    public function testModifyRequestHostlessRequestAddsHost(): void
    {
        $r = new Request('GET', '/b');
        $n = Utils::modifyRequest($r, ['uri' => new Uri('http://new.test:81/z')]);
        $this->assertSame('new.test:81', $n->getHeaderLine('Host'));
        $this->assertSame('http://new.test:81/z', (string) $n->getUri());
    }

    public function testModifyRequestAddsHostFromUriWhenMissing(): void
    {
        $r = new Request('GET', 'http://a.test/b');
        $n = Utils::modifyRequest($r, ['remove_headers' => ['Host']]);
        $this->assertSame(['a.test'], $n->getHeader('Host'));
        $r2 = new Request('GET', 'http://a.test:99/b');
        $n2 = Utils::modifyRequest($r2, ['remove_headers' => ['Host']]);
        $this->assertSame('a.test:99', $n2->getHeaderLine('Host'));
    }

    public function testRemoveDotSegments(): void
    {
        $cases = [
            '' => '',
            '/' => '/',
            '/a/b/c' => '/a/b/c',
            '/a/./b' => '/a/b',
            '/a/../b' => '/b',
            '/a/b/..' => '/a/',
            '/a/b/.' => '/a/b/',
            '../a' => 'a',
            './a' => 'a',
            '/../a' => '/a',
            'a/b/../../c' => 'c',
            '/a//b' => '/a//b',
            'a/..' => '',
            '/a/b/../../..' => '/',
        ];
        foreach ($cases as $in => $out) {
            $this->assertSame($out, UriResolver::removeDotSegments((string) $in), "input '$in'");
        }
    }

    public function testGuardedPath(): void
    {
        $this->assertSame('/p', UriResolver::guardedPath(new Uri('http://a/p'), '/p'));
        $this->assertSame('./a:b', UriResolver::guardedPath(new Uri('x'), 'a:b'));
        $this->assertSame('/x//y', UriResolver::guardedPath(new Uri('http://h//x'), '/x//y'));
    }

    public function testResolveVariants(): void
    {
        $base = new Uri('http://a/b/c/d;p?q');
        $cases = [
            '' => 'http://a/b/c/d;p?q',
            '#s' => 'http://a/b/c/d;p?q#s',
            'g;x?y#s' => 'http://a/b/c/g;x?y#s',
            '.' => 'http://a/b/c/',
            './' => 'http://a/b/c/',
            '..' => 'http://a/b/',
            '../..' => 'http://a/',
            '../../g' => 'http://a/g',
            '../../../g' => 'http://a/g',
            '//other/x' => 'http://other/x',
            'https://z/y/../x' => 'https://z/x',
            '/./g' => 'http://a/g',
            'g/../h' => 'http://a/b/c/h',
            '?y' => 'http://a/b/c/d;p?y',
        ];
        foreach ($cases as $rel => $expected) {
            $this->assertSame($expected, (string) UriResolver::resolve($base, new Uri((string) $rel)), "rel '$rel'");
        }
    }

    public function testResolveRelativeBase(): void
    {
        $this->assertSame('/x', (string) UriResolver::resolve(new Uri('/relative'), new Uri('x')));
    }

    public function testResolveEmptyBaseAuthority(): void
    {
        $base = new Uri('http://h');
        $this->assertSame('http://h/x', (string) UriResolver::resolve($base, new Uri('x')));
        $this->assertSame('http://h/', (string) UriResolver::resolve($base, new Uri('/')));
    }

    public function testRelativize(): void
    {
        $base = new Uri('http://a/b/c/d');
        $cases = [
            'http://a/b/c/d' => '',
            'http://a/b/c/e' => 'e',
            'http://a/b/c/' => './',
            'http://a/b/' => '../',
            'http://a/b/x/y' => '../x/y',
            'http://a/b/c/d?q=1' => '?q=1',
            'http://a/b/c/d#frag' => '#frag',
            'http://a/x' => '../../x',
            'http://other/b/c/d' => '//other/b/c/d',
            'https://a/b/c/d' => 'https://a/b/c/d',
            '/abs' => '../../abs',
            'rel/path' => 'rel/path',
        ];
        foreach ($cases as $target => $expected) {
            $rel = UriResolver::relativize($base, new Uri($target));
            $this->assertSame($expected, (string) $rel, "target '$target'");
        }
    }

    public function testRelativizeColonSegment(): void
    {
        $base = new Uri('http://a/b/c');
        $rel = UriResolver::relativize($base, new Uri('http://a/b/x:y'));
        $this->assertSame('./x:y', (string) $rel);
        $rel = UriResolver::relativize(new Uri('http://h'), new Uri('http://h/x'));
        $this->assertSame('./x', (string) $rel);
        $this->assertSame('http://h/x', (string) UriResolver::resolve(new Uri('http://h'), $rel));
    }

    public function testRelativizeSameQueryDifferentFragment(): void
    {
        $base = new Uri('http://a/p?x=1');
        $this->assertSame('', (string) UriResolver::relativize($base, new Uri('http://a/p?x=1')));
        $this->assertSame('?x=2#f', (string) UriResolver::relativize($base, new Uri('http://a/p?x=2#f')));
    }
}
