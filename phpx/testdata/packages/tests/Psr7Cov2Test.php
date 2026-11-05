<?php
use GuzzleHttp\Psr7\{Rfc7230, ServerRequest, Response, UploadedFile, BufferStream, Utils, Stream, Request};
use PHPUnit\Framework\Attributes\DataProvider;
use PHPUnit\Framework\TestCase;

final class Psr7Cov2Test extends TestCase
{
    private array $serverBackup = [];
    private array $tmp = [];

    protected function setUp(): void
    {
        $this->serverBackup = $_SERVER;
    }

    protected function tearDown(): void
    {
        $_SERVER = $this->serverBackup;
        foreach ($this->tmp as $f) {
            if (is_file($f)) {
                unlink($f);
            }
        }
        $this->tmp = [];
    }

    private function tmpFile(string $content = ''): string
    {
        $f = sys_get_temp_dir() . '/psr7cov2_' . getmypid() . '_' . count($this->tmp);
        file_put_contents($f, $content);
        $this->tmp[] = $f;
        return $f;
    }

    public static function hostProvider(): array
    {
        return [
            'empty' => ['', null],
            'plain' => ['example.com', ['example.com', null]],
            'port' => ['example.com:8080', ['example.com', 8080]],
            'zero port' => ['example.com:000', ['example.com', 0]],
            'leading zero port' => ['example.com:0080', ['example.com', 80]],
            'max port' => ['example.com:65535', ['example.com', 65535]],
            'port too large' => ['example.com:65536', null],
            'port too long' => ['example.com:123456', null],
            'port empty' => ['example.com:', null],
            'port alpha' => ['example.com:ab', null],
            'empty host' => [':80', null],
            'ipv6' => ['[::1]', ['[::1]', null]],
            'ipv6 port' => ['[::1]:8080', ['[::1]', 8080]],
            'ipv6 bad port' => ['[::1]:x', null],
            'ipv6 garbage after' => ['[::1]x', null],
            'ipv6 unclosed' => ['[::1', null],
            'ipv6 invalid' => ['[nothex]', null],
            'ipvfuture' => ['[v1.abc]', ['[v1.abc]', null]],
            'bracket mid' => ['a[b]', null],
            'two colons' => ['a:1:2', null],
            'slash' => ['a/b', null],
            'at sign' => ['u@h', null],
            'space' => ['a b', null],
        ];
    }

    #[DataProvider('hostProvider')]
    public function testParseHostHeader(string $in, ?array $expected): void
    {
        $this->assertSame($expected, Rfc7230::parseHostHeader($in));
    }

    public function testNormalizeFiles(): void
    {
        $file = $this->tmpFile('x');
        $files = ServerRequest::normalizeFiles([
            'a' => ['tmp_name' => $file, 'size' => 1, 'error' => 0, 'name' => 'a.txt', 'type' => 'text/plain'],
            'b' => [
                'tmp_name' => ['x' => $file, 'y' => $file],
                'size' => ['x' => 1, 'y' => 2],
                'error' => ['x' => 0, 'y' => 0],
                'name' => ['x' => 'x.txt', 'y' => 'y.txt'],
                'type' => ['x' => 't1', 'y' => 't2'],
            ],
            'c' => ['deep' => ['tmp_name' => $file, 'size' => 3, 'error' => 0, 'name' => 'd', 'type' => 't']],
            'd' => new UploadedFile($file, 1, UPLOAD_ERR_OK),
        ]);
        $this->assertInstanceOf(UploadedFile::class, $files['a']);
        $this->assertSame('a.txt', $files['a']->getClientFilename());
        $this->assertSame('t2', $files['b']['y']->getClientMediaType());
        $this->assertSame(2, $files['b']['y']->getSize());
        $this->assertSame('d', $files['c']['deep']->getClientFilename());
        $this->assertInstanceOf(UploadedFile::class, $files['d']);
    }

    public function testNormalizeFilesInvalid(): void
    {
        $this->expectException(\InvalidArgumentException::class);
        $this->expectExceptionMessage('Invalid value in files specification');
        ServerRequest::normalizeFiles(['a' => 'str']);
    }

    public function testServerRequestWithers(): void
    {
        $r = new ServerRequest('GET', '/p', [], null, '1.1', ['k' => 'v']);
        $this->assertSame(['k' => 'v'], $r->getServerParams());
        $this->assertSame([], $r->getAttributes());
        $this->assertSame('dflt', $r->getAttribute('x', 'dflt'));
        $r2 = $r->withAttribute('x', 1)->withAttribute('n', null);
        $this->assertSame(1, $r2->getAttribute('x'));
        $this->assertNull($r2->getAttribute('n', 'dflt'));
        $this->assertSame(['x' => 1, 'n' => null], $r2->getAttributes());
        $this->assertSame($r2, $r2->withoutAttribute('missing'));
        $r3 = $r2->withoutAttribute('x');
        $this->assertSame(['n' => null], $r3->getAttributes());
        $this->assertSame(['x' => 1, 'n' => null], $r2->getAttributes());

        $c = $r->withCookieParams(['a' => 'b'])->withQueryParams(['q' => '1'])->withParsedBody(['p' => 2]);
        $this->assertSame(['a' => 'b'], $c->getCookieParams());
        $this->assertSame(['q' => '1'], $c->getQueryParams());
        $this->assertSame(['p' => 2], $c->getParsedBody());
        $this->assertNull($r->getParsedBody());
        $o = new \stdClass();
        $this->assertSame($o, $r->withParsedBody($o)->getParsedBody());
        $this->assertNull($r->withParsedBody(null)->getParsedBody());

        $uf = new UploadedFile('/nonexistent', 1, UPLOAD_ERR_OK);
        $w = $r->withUploadedFiles(['a' => $uf, 'b' => ['c' => $uf]]);
        $this->assertSame(['a' => $uf, 'b' => ['c' => $uf]], $w->getUploadedFiles());
        $this->assertSame([], $r->getUploadedFiles());
    }

    public function testUriFromGlobalsHttpHost(): void
    {
        $_SERVER = [
            'HTTPS' => 'on',
            'HTTP_HOST' => 'host.test:8443',
            'REQUEST_URI' => '/path/to?x=1&y=2',
            'QUERY_STRING' => 'ignored=1',
        ];
        $this->assertSame('https://host.test:8443/path/to?x=1&y=2', (string) ServerRequest::getUriFromGlobals());
    }

    public function testUriFromGlobalsServerName(): void
    {
        $_SERVER = [
            'HTTPS' => 'off',
            'SERVER_NAME' => 'name.test',
            'SERVER_PORT' => '8080',
            'REQUEST_URI' => '/a',
            'QUERY_STRING' => 'z=9',
        ];
        $this->assertSame('http://name.test:8080/a?z=9', (string) ServerRequest::getUriFromGlobals());
    }

    public function testUriFromGlobalsServerAddr(): void
    {
        $_SERVER = ['SERVER_ADDR' => '10.0.0.1', 'SERVER_PORT' => 'abc', 'REQUEST_URI' => '/'];
        $u = ServerRequest::getUriFromGlobals();
        $this->assertSame('10.0.0.1', $u->getHost());
        $this->assertNull($u->getPort());
        $this->assertSame('/', $u->getPath());
    }

    public function testUriFromGlobalsInvalidHostAndNone(): void
    {
        $_SERVER = ['HTTP_HOST' => 'bad host', 'SERVER_PORT' => '81'];
        $u = ServerRequest::getUriFromGlobals();
        $this->assertSame('localhost', $u->getHost());
        $this->assertSame(81, $u->getPort());
        $_SERVER = ['HTTP_HOST' => 'h.test', 'SERVER_PORT' => '81'];
        $this->assertSame(81, ServerRequest::getUriFromGlobals()->getPort());
        $_SERVER = [];
        $this->assertSame('http://localhost', (string) ServerRequest::getUriFromGlobals());
    }

    public function testResponseWithStatus(): void
    {
        $r = new Response(200);
        $n = $r->withStatus(404);
        $this->assertSame(404, $n->getStatusCode());
        $this->assertSame('Not Found', $n->getReasonPhrase());
        $this->assertSame(200, $r->getStatusCode());
        $c = $r->withStatus(299, 'Custom');
        $this->assertSame('Custom', $c->getReasonPhrase());
        $this->assertSame('', $r->withStatus(299)->getReasonPhrase());
        $this->assertSame(503, $r->withStatus(503)->getStatusCode());
        $this->assertSame('I\'m a teapot', $r->withStatus(418)->getReasonPhrase());
    }

    public function testResponseWithStatusInvalid(): void
    {
        $r = new Response();
        foreach ([99, 600, 1000] as $code) {
            try {
                $r->withStatus($code);
                $this->fail("accepted $code");
            } catch (\InvalidArgumentException $e) {
                $this->assertSame('Status code must be an integer value between 1xx and 5xx.', $e->getMessage());
            }
        }
        try {
            $r->withStatus('abc');
            $this->fail();
        } catch (\InvalidArgumentException $e) {
            $this->assertSame('Status code must be an integer value.', $e->getMessage());
        }
        try {
            $r->withStatus(200, "a\nb");
            $this->fail();
        } catch (\InvalidArgumentException $e) {
            $this->assertStringContainsString('Reason phrase', $e->getMessage());
        }
    }

    public function testResponseConstruct(): void
    {
        $r = new Response(201, ['X-A' => 'b'], 'body', '1.0', 'Made');
        $this->assertSame('1.0', $r->getProtocolVersion());
        $this->assertSame('Made', $r->getReasonPhrase());
        $this->assertSame('b', $r->getHeaderLine('x-a'));
        $this->assertSame('body', (string) $r->getBody());
        $this->assertSame('Continue', (new Response(100))->getReasonPhrase());
    }

    public function testUploadedFileStreamAndMove(): void
    {
        $src = $this->tmpFile('upload-body');
        $dst = $src . '.moved';
        $this->tmp[] = $dst;
        $u = new UploadedFile($src, 11, UPLOAD_ERR_OK, 'c.txt', 'text/plain');
        $this->assertSame(UPLOAD_ERR_OK, $u->getError());
        $this->assertSame('text/plain', $u->getClientMediaType());
        $this->assertSame('upload-body', (string) $u->getStream());
        $u->moveTo($dst);
        $this->assertSame('upload-body', file_get_contents($dst));
        $this->assertFileDoesNotExist($src);
        try {
            $u->getStream();
            $this->fail();
        } catch (\RuntimeException $e) {
            $this->assertStringContainsString('already been moved', $e->getMessage());
        }
        try {
            $u->moveTo($dst . 'x');
            $this->fail();
        } catch (\RuntimeException $e) {
            $this->assertStringContainsString('already been moved', $e->getMessage());
        }
    }

    public function testUploadedFileFromStreamAndResource(): void
    {
        $dst = $this->tmpFile('');
        $u = new UploadedFile(Utils::streamFor('from-stream'), null, UPLOAD_ERR_OK);
        $this->assertNull($u->getSize());
        $this->assertNull($u->getClientFilename());
        $u->moveTo($dst);
        $this->assertSame('from-stream', file_get_contents($dst));

        $res = fopen('php://temp', 'r+');
        fwrite($res, 'res-data');
        rewind($res);
        $u2 = new UploadedFile($res, 8, UPLOAD_ERR_OK);
        $u2->moveTo($dst);
        $this->assertSame('res-data', file_get_contents($dst));
    }

    public function testUploadedFileErrors(): void
    {
        try {
            new UploadedFile('x', 1, 99);
            $this->fail();
        } catch (\InvalidArgumentException $e) {
            $this->assertSame('Invalid error status for UploadedFile', $e->getMessage());
        }
        try {
            new UploadedFile(123, 1, UPLOAD_ERR_OK);
            $this->fail();
        } catch (\InvalidArgumentException $e) {
            $this->assertSame('Invalid stream or file provided for UploadedFile', $e->getMessage());
        }
        $bad = new UploadedFile('x', 1, UPLOAD_ERR_NO_FILE);
        $this->assertSame(UPLOAD_ERR_NO_FILE, $bad->getError());
        try {
            $bad->getStream();
            $this->fail();
        } catch (\RuntimeException $e) {
            $this->assertStringContainsString('Cannot retrieve stream due to upload error', $e->getMessage());
        }
        try {
            $bad->moveTo('/tmp/whatever');
            $this->fail();
        } catch (\RuntimeException $e) {
            $this->assertStringContainsString('upload error', $e->getMessage());
        }
        $ok = new UploadedFile(Utils::streamFor('a'), 1, UPLOAD_ERR_OK);
        foreach (['', 123] as $path) {
            try {
                $ok->moveTo($path);
                $this->fail();
            } catch (\InvalidArgumentException $e) {
                $this->assertStringContainsString('path', $e->getMessage());
            }
        }
    }

    public function testBufferStream(): void
    {
        $b = new BufferStream(5);
        $this->assertTrue($b->isReadable());
        $this->assertTrue($b->isWritable());
        $this->assertFalse($b->isSeekable());
        $this->assertSame(0, $b->getSize());
        $this->assertTrue($b->eof());
        $this->assertSame(3, $b->write('abc'));
        $this->assertSame(3, $b->getSize());
        $this->assertFalse($b->eof());
        $this->assertSame(5, $b->getMetadata('hwm'));
        $this->assertSame([], $b->getMetadata());
        $this->assertNull($b->getMetadata('other'));
        $this->assertSame(0, $b->write('defgh'));
        $this->assertSame(0, $b->write('more'));
        $this->assertSame('abcdefghmore', $b->read(100));
        $this->assertSame('', $b->read(1));
        $b->write('xyz');
        $this->assertSame('xyz', (string) $b);
        $this->assertSame('', (string) $b);
        $b->write('q');
        $b->close();
        $this->assertSame(0, $b->getSize());
        $b->write('r');
        $this->assertNull($b->detach());
        $this->assertSame(0, $b->getSize());
        $b->write('z');
        $this->assertSame('z', $b->getContents());
    }

    public function testBufferStreamNotSeekable(): void
    {
        $b = new BufferStream();
        foreach (['seek' => [0], 'rewind' => []] as $m => $args) {
            try {
                $b->$m(...$args);
                $this->fail();
            } catch (\RuntimeException $e) {
                $this->assertSame('Cannot seek a BufferStream', $e->getMessage());
            }
        }
        try {
            $b->tell();
            $this->fail();
        } catch (\RuntimeException $e) {
            $this->assertSame('Cannot determine the position of a BufferStream', $e->getMessage());
        }
    }
}
