<?php
use GuzzleHttp\Psr7\{Message, Request, Response, Query, MultipartStream, Utils, Uri, ServerRequest, Stream};
use PHPUnit\Framework\TestCase;

final class Psr7Cov5Test extends TestCase
{
    public function testToStringRequestWithoutHostHeader(): void
    {
        $r = new Request('GET', '/path');
        $this->assertSame("GET /path HTTP/1.1\r\nHost: \r\n\r\n", Message::toString($r));
        $r = (new Request('POST', 'http://h.test/p?x=1', [], 'body'))->withoutHeader('Host');
        $this->assertSame("POST /p?x=1 HTTP/1.1\r\nHost: h.test\r\n\r\nbody", Message::toString($r));
    }

    public function testToStringResponseAndSetCookie(): void
    {
        $r = new Response(404, ['Set-Cookie' => ['a=1', 'b=2'], 'X-List' => ['x', 'y']], 'nope', '1.0');
        $this->assertSame(
            "HTTP/1.0 404 Not Found\r\nSet-Cookie: a=1\r\nSet-Cookie: b=2\r\nX-List: x, y\r\n\r\nnope",
            Message::toString($r)
        );
    }

    public function testToStringUnknownMessage(): void
    {
        $m = new class implements \Psr\Http\Message\MessageInterface {
            public function getProtocolVersion(): string { return '1.1'; }
            public function withProtocolVersion(string $version): \Psr\Http\Message\MessageInterface { return $this; }
            public function getHeaders(): array { return []; }
            public function hasHeader(string $name): bool { return false; }
            public function getHeader(string $name): array { return []; }
            public function getHeaderLine(string $name): string { return ''; }
            public function withHeader(string $name, $value): \Psr\Http\Message\MessageInterface { return $this; }
            public function withAddedHeader(string $name, $value): \Psr\Http\Message\MessageInterface { return $this; }
            public function withoutHeader(string $name): \Psr\Http\Message\MessageInterface { return $this; }
            public function getBody(): \Psr\Http\Message\StreamInterface { return Utils::streamFor(''); }
            public function withBody(\Psr\Http\Message\StreamInterface $body): \Psr\Http\Message\MessageInterface { return $this; }
        };
        $this->expectException(\InvalidArgumentException::class);
        $this->expectExceptionMessage('Unknown message type');
        Message::toString($m);
    }

    public function testBodySummary(): void
    {
        $this->assertSame('hello', Message::bodySummary(new Response(200, [], 'hello')));
        $this->assertNull(Message::bodySummary(new Response(200)));
        $this->assertNull(Message::bodySummary(new Response(200, [], "bin\x00ary")));
        $this->assertSame('abcde (truncated...)', Message::bodySummary(new Response(200, [], 'abcdefghij'), 5));
        $nonSeekable = new Response(200, [], new \GuzzleHttp\Psr7\NoSeekStream(Utils::streamFor('abc')));
        $this->assertNull(Message::bodySummary($nonSeekable));
        $this->assertNull(Message::bodySummary(new Response(200, [], new \GuzzleHttp\Psr7\PumpStream(fn () => false))));
    }

    public function testBodySummaryTruncatesMultibyteSafely(): void
    {
        $body = str_repeat('a', 4) . 'é' . 'xyz';
        $this->assertSame('aaaa (truncated...)', Message::bodySummary(new Response(200, [], $body), 5));
        $this->assertSame('aaaaé (truncated...)', Message::bodySummary(new Response(200, [], $body), 6));
        $body3 = 'ab' . '€' . 'z';
        $this->assertSame('ab (truncated...)', Message::bodySummary(new Response(200, [], $body3), 3));
        $this->assertSame('ab (truncated...)', Message::bodySummary(new Response(200, [], $body3), 4));
        $body4 = 'a' . "\u{1F600}" . 'z';
        $this->assertSame('a (truncated...)', Message::bodySummary(new Response(200, [], $body4), 4));
    }

    public function testRewindBody(): void
    {
        $r = new Response(200, [], 'abcdef');
        $r->getBody()->read(3);
        Message::rewindBody($r);
        $this->assertSame(0, $r->getBody()->tell());
        Message::rewindBody($r);
        $this->assertSame('abcdef', (string) $r->getBody());
    }

    public function testParseMessageErrors(): void
    {
        $cases = [
            '' => 'Invalid message',
            "GET / HTTP/1.1" => 'Invalid message: Missing header delimiter',
            "GET / HTTP/1.1\r\nHost: a" => 'Invalid message: Missing header delimiter',
            "GET / HTTP/1.1\r\nHost: a\r\nbad header\r\n\r\n" => 'Invalid header syntax',
            "GET / HTTP/1.1\r\nX: a\r\n b\r\n\r\n" => 'Invalid header syntax: Obsolete line folding',
        ];
        foreach ($cases as $msg => $expected) {
            try {
                Message::parseMessage((string) $msg);
                $this->fail("no exception for " . json_encode($msg));
            } catch (\InvalidArgumentException $e) {
                $this->assertSame($expected, $e->getMessage());
            }
        }
    }

    public function testParseMessageOk(): void
    {
        $p = Message::parseMessage("\r\n\nGET /x HTTP/1.1\nA: 1\nA: 2\nB:  v \n\nbody\n\nmore");
        $this->assertSame('GET /x HTTP/1.1', $p['start-line']);
        $this->assertSame(['A' => ['1', '2'], 'B' => ['v']], $p['headers']);
        $this->assertSame("body\n\nmore", $p['body']);
    }

    public function testParseMessageHttp10UnfoldsHeaders(): void
    {
        $p = Message::parseMessage("HTTP/1.0 200 OK\r\nX-Fold: one\r\n two\r\n\r\nb");
        $this->assertSame(['X-Fold' => ['one two']], $p['headers']);
        $p = Message::parseMessage("GET / HTTP/1.0\r\nX-Fold: one\r\n\ttwo\r\n\r\n");
        $this->assertSame(['X-Fold' => ['one two']], $p['headers']);
    }

    public function testParseRequestUri(): void
    {
        $this->assertSame('/a/b', Message::parseRequestUri('/a/b', []));
        $this->assertSame('/a//b', Message::parseRequestUri('/a//b', []));
        $this->assertSame('/evil.test/x', Message::parseRequestUri('//evil.test/x', []));
        $this->assertSame('http://h.test/a', Message::parseRequestUri('/a', ['Host' => ['h.test']]));
        $this->assertSame('http://h.test/a', Message::parseRequestUri('a', ['HOST' => ['h.test']]));
        $this->assertSame('https://h.test:443/a', Message::parseRequestUri('/a', ['host' => ['h.test:443']]));
        $this->assertSame('http://h.test:8080/a', Message::parseRequestUri('/a', ['Host' => ['h.test:8080']]));
    }

    public function testParseRequestUriInvalidHost(): void
    {
        $this->expectException(\InvalidArgumentException::class);
        $this->expectExceptionMessage('Invalid request string');
        Message::parseRequestUri('/a', ['Host' => ['bad host']]);
    }

    public function testParseRequestVariants(): void
    {
        $r = Message::parseRequest("GET http://abs.test/p?q=1 HTTP/1.0\r\nX-A: b\r\n\r\n");
        $this->assertSame('http://abs.test/p?q=1', $r->getRequestTarget());
        $this->assertSame('1.0', $r->getProtocolVersion());
        $this->assertSame('abs.test', $r->getUri()->getHost());
        $r = Message::parseRequest("POST /submit HTTP/1.1\r\nHost: h.test\r\nContent-Length: 4\r\n\r\ndata");
        $this->assertSame('POST', $r->getMethod());
        $this->assertSame('http://h.test/submit', (string) $r->getUri());
        $this->assertSame('data', (string) $r->getBody());
        $r = Message::parseRequest("GET /nohost HTTP/1.1\r\nA: b\r\n\r\n");
        $this->assertSame('/nohost', (string) $r->getUri());
    }

    public function testParseRequestInvalid(): void
    {
        foreach (["GET\r\nX: y\r\n\r\n", "GARBAGE\r\nX: y\r\n\r\n", "GET nopath HTTP/1.1\r\nX: y\r\n\r\n"] as $msg) {
            try {
                Message::parseRequest($msg);
                $this->fail('accepted ' . json_encode($msg));
            } catch (\InvalidArgumentException $e) {
                $this->assertSame('Invalid request string', $e->getMessage());
            }
        }
    }

    public function testParseResponseVariants(): void
    {
        $r = Message::parseResponse("HTTP/1.1 404\r\nA: b\r\n\r\n");
        $this->assertSame(404, $r->getStatusCode());
        $this->assertSame('Not Found', $r->getReasonPhrase());
        $r = Message::parseResponse("HTTP/2 200 All Good Here\r\nA: b\r\n\r\nx");
        $this->assertSame('2', $r->getProtocolVersion());
        $this->assertSame('All Good Here', $r->getReasonPhrase());
    }

    public function testParseResponseInvalid(): void
    {
        foreach (["NOTHTTP 200 OK\r\nX: y\r\n\r\n", "HTTP/1.1 20 OK\r\nX: y\r\n\r\n", "HTTP/1.1 2000 OK\r\nX: y\r\n\r\n"] as $msg) {
            try {
                Message::parseResponse($msg);
                $this->fail('accepted ' . json_encode($msg));
            } catch (\InvalidArgumentException $e) {
                $this->assertStringContainsString('Invalid response string', $e->getMessage());
            }
        }
    }

    public function testMessageTraitHeaders(): void
    {
        $r = new Request('GET', '/');
        $a = $r->withHeader('X-One', 'a')->withAddedHeader('x-one', 'b')->withAddedHeader('X-Two', ['c', 'd']);
        $this->assertSame(['a', 'b'], $a->getHeader('X-ONE'));
        $this->assertSame('a, b', $a->getHeaderLine('x-one'));
        $this->assertSame('c, d', $a->getHeaderLine('X-Two'));
        $this->assertSame([], $a->getHeader('Absent'));
        $this->assertSame('', $a->getHeaderLine('Absent'));
        $this->assertTrue($a->hasHeader('X-TWO'));
        $b = $a->withHeader('x-ONE', ['z']);
        $this->assertSame(['z'], $b->getHeader('X-One'));
        $this->assertArrayHasKey('x-ONE', $b->getHeaders());
        $this->assertArrayNotHasKey('X-One', $b->getHeaders());
        $this->assertSame($a, $a->withoutHeader('Absent'));
        $c = $a->withoutHeader('x-one');
        $this->assertFalse($c->hasHeader('X-One'));
        $this->assertTrue($a->hasHeader('X-One'));
        $d = $r->withAddedHeader('Fresh', 'v');
        $this->assertSame(['v'], $d->getHeader('fresh'));
        $this->assertSame(['t'], $r->withHeader('T', "  t\t")->getHeader('T'));
    }

    public function testMessageTraitHeaderValidation(): void
    {
        $r = new Request('GET', '/');
        $bad = [
            fn () => $r->withHeader('Bad Name', 'v'),
            fn () => $r->withHeader('', 'v'),
            fn () => $r->withAddedHeader('Bad:Name', 'v'),
            fn () => $r->withHeader('X', "line1\nline2"),
            fn () => $r->withHeader('X', "a\x00b"),
            fn () => $r->withHeader('X', [new \stdClass()]),
            fn () => $r->withHeader('X', [['nested']]),
            fn () => $r->withAddedHeader('X', "v\r\nInjected: 1"),
            fn () => new Request('GET', '/', ["Bad Header" => 'v']),
            fn () => new Request('GET', '/', ['X' => "v\n"]),
        ];
        foreach ($bad as $i => $fn) {
            try {
                $fn();
                $this->fail("case $i accepted");
            } catch (\InvalidArgumentException $e) {
                $this->assertNotSame('', $e->getMessage());
            }
        }
        try {
            $r->withHeader('Bad Name', 'v');
        } catch (\InvalidArgumentException $e) {
            $this->assertSame('"Bad Name" is not valid header name.', $e->getMessage());
        }
        try {
            $r->withHeader('X', new \stdClass());
        } catch (\InvalidArgumentException $e) {
            $this->assertSame('Header value must be scalar or null but stdClass provided.', $e->getMessage());
        }
    }

    public function testMessageTraitProtocolAndBody(): void
    {
        $r = new Request('GET', '/', [], 'b');
        $this->assertSame($r, $r->withProtocolVersion('1.1'));
        $n = $r->withProtocolVersion('2');
        $this->assertSame('2', $n->getProtocolVersion());
        $this->assertSame('1.1', $r->getProtocolVersion());
        try {
            $r->withProtocolVersion("1.1\r\nX: y");
            $this->fail();
        } catch (\InvalidArgumentException $e) {
            $this->assertSame('Protocol version must not contain CR or LF characters.', $e->getMessage());
        }
        $this->assertSame($r, $r->withBody($r->getBody()));
        $nb = $r->withBody(Utils::streamFor('other'));
        $this->assertSame('other', (string) $nb->getBody());
        $this->assertSame('b', (string) $r->getBody());
        $empty = new Request('GET', '/');
        $this->assertSame('', (string) $empty->getBody());
        $this->assertSame($empty->getBody(), $empty->getBody());
    }

    public function testRequestTargetAndMethod(): void
    {
        $r = new Request('GET', 'http://h.test');
        $this->assertSame('/', $r->getRequestTarget());
        $r2 = new Request('GET', 'http://h.test/a b?q=1');
        $this->assertSame('/a%20b?q=1', $r2->getRequestTarget());
        $t = $r->withRequestTarget('*');
        $this->assertSame('*', $t->getRequestTarget());
        $this->assertSame('/', $r->getRequestTarget());
        try {
            $r->withRequestTarget('has space');
            $this->fail();
        } catch (\InvalidArgumentException $e) {
            $this->assertStringContainsString('Invalid request target', $e->getMessage());
        }
        foreach (['', "GE\nT"] as $m) {
            try {
                new Request($m, '/');
                $this->fail();
            } catch (\InvalidArgumentException $e) {
                $this->assertNotSame('', $e->getMessage());
            }
        }
        $this->assertSame('POST', $r->withMethod('POST')->getMethod());
        $this->expectException(\InvalidArgumentException::class);
        $this->expectExceptionMessage('Method must be a non-empty string.');
        $r->withMethod('');
    }

    public function testRequestWithUriHostHandling(): void
    {
        $r = new Request('GET', 'http://a.test:81/x');
        $this->assertSame(['a.test:81'], $r->getHeader('Host'));
        $this->assertSame($r, $r->withUri($r->getUri()));
        $n = $r->withUri(new Uri('http://b.test/y'));
        $this->assertSame(['b.test'], $n->getHeader('Host'));
        $p = $r->withUri(new Uri('http://b.test/y'), true);
        $this->assertSame(['a.test:81'], $p->getHeader('Host'));
        $np = (new Request('GET', '/x'))->withUri(new Uri('http://c.test/y'), true);
        $this->assertSame(['c.test'], $np->getHeader('Host'));
        $custom = new Request('GET', 'http://a.test/', ['host' => 'explicit.test']);
        $this->assertSame(['explicit.test'], $custom->getHeader('Host'));
        $this->assertSame(['d.test'], $custom->withUri(new Uri('http://d.test/'))->getHeader('Host'));
        $this->assertSame(['http://d.test/'], [(string) $custom->withUri(new Uri('http://d.test/'))->getUri()]);
        $this->assertSame([], (new Request('GET', '/nohost'))->getHeader('Host'));
    }

    public function testQueryParse(): void
    {
        $this->assertSame([], Query::parse(''));
        $this->assertSame(['a' => '1', 'b' => ['2', '3'], 'c' => null], Query::parse('a=1&b=2&b=3&c'));
        $this->assertSame(['a b' => 'c d'], Query::parse('a+b=c+d'));
        $this->assertSame(['a+b' => 'c+d'], Query::parse('a+b=c+d', PHP_QUERY_RFC3986));
        $this->assertSame(['a b' => 'c d%2'], Query::parse('a+b=c+d%2', PHP_QUERY_RFC1738));
        $this->assertSame(['a%20b' => 'c%20d'], Query::parse('a%20b=c%20d', false));
        $this->assertSame(['k' => 'v=w'], Query::parse('k=v=w'));
        $this->assertSame(['x' => ['1', '2', '3']], Query::parse('x=1&x=2&x=3'));
    }

    public function testQueryBuild(): void
    {
        $this->assertSame('', Query::build([]));
        $this->assertSame('a=b%20c&d=1&e', Query::build(['a' => 'b c', 'd' => 1, 'e' => null]));
        $this->assertSame('a=b+c', Query::build(['a' => 'b c'], PHP_QUERY_RFC1738));
        $this->assertSame('a=b c', Query::build(['a' => 'b c'], false));
        $this->assertSame('x=1&x=2&y', Query::build(['x' => [1, 2], 'y' => [null]]));
        $this->assertSame('t=1&f=0', Query::build(['t' => true, 'f' => false]));
        $this->assertSame('t=true&f=false', Query::build(['t' => true, 'f' => false], PHP_QUERY_RFC3986, false));
        $this->assertSame('l=true&l=false', Query::build(['l' => [true, false]], PHP_QUERY_RFC3986, false));
        $this->assertSame('l=1&l=0', Query::build(['l' => [true, false]]));
        $this->assertSame('f=1.5', Query::build(['f' => 1.5]));
        $this->assertSame('a%26b=c%3Dd', Query::build(['a&b' => 'c=d']));
        $this->expectException(\InvalidArgumentException::class);
        $this->expectExceptionMessage('Invalid type');
        Query::build(['a' => 'b'], 999);
    }

    public function testMultipartStreamBasics(): void
    {
        $m = new MultipartStream([
            ['name' => 'field', 'contents' => 'value'],
            ['name' => 'num', 'contents' => 42],
            ['name' => 'file', 'contents' => 'filedata', 'filename' => 'a.txt'],
            ['name' => 'hdr', 'contents' => 'x', 'headers' => ['X-Custom' => 'yes', 'Content-Type' => 'text/x-custom']],
        ], 'BOUNDARY');
        $this->assertSame('BOUNDARY', $m->getBoundary());
        $this->assertFalse($m->isWritable());
        $s = (string) $m;
        $this->assertStringContainsString("--BOUNDARY\r\nContent-Disposition: form-data; name=\"field\"\r\nContent-Length: 5\r\n\r\nvalue\r\n", $s);
        $this->assertStringContainsString("name=\"num\"\r\nContent-Length: 2\r\n\r\n42\r\n", $s);
        $this->assertStringContainsString('name="file"; filename="a.txt"', $s);
        $this->assertStringContainsString("Content-Type: text/plain\r\n\r\nfiledata", $s);
        $this->assertStringContainsString("X-Custom: yes\r\n", $s);
        $this->assertStringContainsString("Content-Type: text/x-custom\r\n", $s);
        $this->assertStringEndsWith("--BOUNDARY--\r\n", $s);
    }

    public function testMultipartStreamGeneratedBoundaryAndFilenameFromStream(): void
    {
        $f = sys_get_temp_dir() . '/psr7cov5_' . getmypid() . '.bin';
        file_put_contents($f, "bin");
        try {
            $m = new MultipartStream([['name' => 'up', 'contents' => fopen($f, 'r')]]);
            $this->assertMatchesRegularExpression('/^[0-9a-f]{40}$/', $m->getBoundary());
            $s = (string) $m;
            $this->assertStringContainsString('filename="psr7cov5_' . getmypid() . '.bin"', $s);
            $this->assertStringContainsString('Content-Type: application/octet-stream', $s);
            $m->close();
        } finally {
            unlink($f);
        }
    }

    public function testMultipartStreamNested(): void
    {
        $m = new MultipartStream([
            ['name' => 'user', 'contents' => ['name' => 'bob', 'tags' => ['a', 'b']]],
            ['name' => 0, 'contents' => 'zero'],
        ], 'B');
        $s = (string) $m;
        $this->assertStringContainsString('name="user[name]"', $s);
        $this->assertStringContainsString('name="user[tags][0]"', $s);
        $this->assertStringContainsString('name="user[tags][1]"', $s);
        $this->assertStringContainsString('name="0"', $s);
    }

    public function testMultipartStreamNestedWithEmptyRoot(): void
    {
        $m = new MultipartStream([['name' => '', 'contents' => ['k' => 'v']]], 'B');
        $this->assertStringContainsString('name="[k]"', (string) $m);
    }

    public function testMultipartStreamErrors(): void
    {
        $cases = [
            [[['contents' => 'x']], "A 'name' key is required", \InvalidArgumentException::class],
            [[['name' => 'x']], "A 'contents' key is required", \InvalidArgumentException::class],
            [[['name' => ['x'], 'contents' => 'x']], "The 'name' key must be a string or integer", \InvalidArgumentException::class],
            [[['name' => 'x', 'contents' => ['a'], 'filename' => 'f']], "The 'filename' and 'headers' options cannot be used when 'contents' is an array", \InvalidArgumentException::class],
            [[['name' => 'x', 'contents' => ['a'], 'headers' => []]], "The 'filename' and 'headers' options cannot be used when 'contents' is an array", \InvalidArgumentException::class],
            [['notarray'], 'An array is expected', \UnexpectedValueException::class],
        ];
        foreach ($cases as [$elements, $message, $class]) {
            try {
                new MultipartStream($elements, 'B');
                $this->fail("accepted: $message");
            } catch (\Throwable $e) {
                $this->assertInstanceOf($class, $e);
                $this->assertSame($message, $e->getMessage());
            }
        }
    }

    public function testMultipartStreamInvalidHeaderValueType(): void
    {
        $this->expectException(\InvalidArgumentException::class);
        $this->expectExceptionMessage('Multipart part header value must be a string or stringable value but array provided.');
        new MultipartStream([['name' => 'x', 'contents' => 'y', 'headers' => ['X-Bad' => ['arr']]]], 'B');
    }

    public function testMultipartStreamDispositionOverride(): void
    {
        $m = new MultipartStream([
            ['name' => 'x', 'contents' => 'y', 'headers' => ['content-disposition' => 'attachment; name="x"', 'Content-Length' => '1']],
        ], 'B');
        $s = (string) $m;
        $this->assertStringContainsString("content-disposition: attachment; name=\"x\"\r\n", $s);
        $this->assertStringNotContainsString('form-data', $s);
        $this->assertSame(1, substr_count($s, 'Content-Length'));
    }

    public function testMultipartStreamFilenameZero(): void
    {
        $m = new MultipartStream([['name' => 'x', 'contents' => 'y', 'filename' => '0']], 'B');
        $this->assertStringContainsString('name="x"; filename="0"', (string) $m);
    }
}
