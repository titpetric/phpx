<?php
use GuzzleHttp\Psr7\{Uri, Request, Response, Utils, Query, MessageTrait, Stream};
use PHPUnit\Framework\TestCase;

final class Psr7Test extends TestCase
{
    public function testUri(): void
    {
        $u = new Uri('https://user@example.com:8080/a/b?x=1&y=2#frag');
        $this->assertSame('https', $u->getScheme());
        $this->assertSame('example.com', $u->getHost());
        $this->assertSame(8080, $u->getPort());
        $this->assertSame('/a/b', $u->getPath());
        $this->assertSame('x=1&y=2', $u->getQuery());
        $this->assertSame('frag', $u->getFragment());
        $v = $u->withPath('/z')->withScheme('http')->withPort(null);
        $this->assertSame('http://user@example.com/z?x=1&y=2#frag', (string) $v);
        $this->assertSame('/a/b', $u->getPath());
    }
    public function testRequest(): void
    {
        $r = new Request('POST', 'http://x.test/p?q=1', ['Content-Type' => 'application/json'], '{"a":1}');
        $this->assertSame('POST', $r->getMethod());
        $this->assertSame('application/json', $r->getHeaderLine('content-type'));
        $this->assertTrue($r->hasHeader('CONTENT-TYPE'));
        $this->assertSame('/p?q=1', $r->getRequestTarget());
        $this->assertSame('{"a":1}', (string) $r->getBody());
        $r2 = $r->withHeader('X-A', ['1', '2'])->withMethod('PUT');
        $this->assertSame('1, 2', $r2->getHeaderLine('x-a'));
        $this->assertSame('POST', $r->getMethod());
    }
    public function testResponse(): void
    {
        $r = new Response(404, ['X' => 'y'], 'nope');
        $this->assertSame(404, $r->getStatusCode());
        $this->assertSame('Not Found', $r->getReasonPhrase());
        $this->assertSame('nope', $r->getBody()->getContents());
    }
    public function testQuery(): void
    {
        $this->assertSame(['a' => '1', 'b[]' => ['x', 'y']], Query::parse('a=1&b[]=x&b[]=y'));
        $this->assertSame('a=1&b=2', Query::build(['a' => 1, 'b' => 2]));
    }
    public function testStream(): void
    {
        $s = Utils::streamFor('hello world');
        $this->assertSame('hello', $s->read(5));
        $this->assertSame(5, $s->tell());
        $s->seek(0);
        $this->assertSame('hello world', (string) $s);
        $this->assertSame(11, $s->getSize());
        $this->assertTrue($s->isReadable());
    }
}
