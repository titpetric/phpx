<?php
use GuzzleHttp\Psr7\{Uri, UriResolver, UriNormalizer, Utils, Header, Message, ServerRequest, MultipartStream, LimitStream, AppendStream, BufferStream, Request, Response, UploadedFile, Rfc7230};
use PHPUnit\Framework\TestCase;

final class Psr7MoreTest extends TestCase
{
    public function testResolve(): void
    {
        $base = new Uri('http://a/b/c/d;p?q');
        $this->assertSame('http://a/b/c/g', (string) UriResolver::resolve($base, new Uri('g')));
        $this->assertSame('http://a/g', (string) UriResolver::resolve($base, new Uri('/g')));
        $this->assertSame('http://a/b/g', (string) UriResolver::resolve($base, new Uri('../g')));
        $this->assertSame('http://a/b/c/d;p?y', (string) UriResolver::resolve($base, new Uri('?y')));
    }
    public function testNormalize(): void
    {
        $u = new Uri('HTTP://Example.COM:80/a/./b/../c?b=2&a=1');
        $n = UriNormalizer::normalize($u, UriNormalizer::PRESERVING_NORMALIZATIONS | UriNormalizer::REMOVE_DOT_SEGMENTS | UriNormalizer::SORT_QUERY_PARAMETERS);
        $this->assertSame('http://example.com/a/c?a=1&b=2', (string) $n);
        $this->assertTrue(UriNormalizer::isEquivalent(new Uri('http://a/%7euser'), new Uri('http://a/~user')));
    }
    public function testHeaderParse(): void
    {
        $p = Header::parse('text/html; charset=utf-8, application/json;q=0.5');
        $this->assertSame('text/html', $p[0][0]);
        $this->assertSame('utf-8', $p[0]['charset']);
        $this->assertSame(['a', 'b', '"c d"'], Header::splitList('a, b, "c d"'));
    }
    public function testMessageToString(): void
    {
        $r = new Request('GET', 'http://x.test/p', ['Host' => 'x.test', 'A' => 'b']);
        $this->assertSame("GET /p HTTP/1.1\r\nHost: x.test\r\nA: b\r\n\r\n", Message::toString($r));
        $parsed = Message::parseRequest("GET /q?z=1 HTTP/1.1\r\nHost: h.test\r\n\r\n");
        $this->assertSame('h.test', $parsed->getUri()->getHost());
        $this->assertSame('z=1', $parsed->getUri()->getQuery());
        $resp = Message::parseResponse("HTTP/1.1 201 Created\r\nX: 1\r\n\r\nbody");
        $this->assertSame(201, $resp->getStatusCode());
        $this->assertSame('body', (string) $resp->getBody());
    }
    public function testStreams(): void
    {
        $a = new AppendStream([Utils::streamFor('ab'), Utils::streamFor('cd')]);
        $this->assertSame('abcd', (string) $a);
        $l = new LimitStream(Utils::streamFor('0123456789'), 4, 2);
        $this->assertSame('2345', (string) $l);
        $b = new BufferStream();
        $b->write('xyz');
        $this->assertSame('xy', $b->read(2));
        $this->assertSame(1, $b->getSize());
    }
    public function testMultipart(): void
    {
        $m = new MultipartStream([['name' => 'f', 'contents' => 'v'], ['name' => 'g', 'contents' => 'w', 'filename' => 'a.txt']], 'BOUND');
        $s = (string) $m;
        $this->assertStringContainsString("--BOUND\r\nContent-Disposition: form-data; name=\"f\"", $s);
        $this->assertStringContainsString('filename="a.txt"', $s);
        $this->assertSame('BOUND', $m->getBoundary());
    }
    public function testServerRequest(): void
    {
        $r = new ServerRequest('POST', '/x', [], null, '1.1', ['REMOTE_ADDR' => '1.2.3.4']);
        $r = $r->withQueryParams(['a' => '1'])->withParsedBody(['b' => '2'])->withAttribute('id', 7)->withCookieParams(['c' => 'd']);
        $this->assertSame(['a' => '1'], $r->getQueryParams());
        $this->assertSame(['b' => '2'], $r->getParsedBody());
        $this->assertSame(7, $r->getAttribute('id'));
        $this->assertSame('dflt', $r->getAttribute('none', 'dflt'));
        $this->assertSame('1.2.3.4', $r->getServerParams()['REMOTE_ADDR']);
        $this->assertSame(['c' => 'd'], $r->getCookieParams());
    }
    public function testUriEncoding(): void
    {
        $u = (new Uri('http://h'))->withPath('/a b/ü')->withQuery('k=v w');
        $this->assertSame('/a%20b/%C3%BC', $u->getPath());
        $this->assertSame('k=v%20w', $u->getQuery());
        $this->assertSame('h', Uri::fromParts(['host' => 'h'])->getAuthority());
        $this->assertTrue(Uri::isDefaultPort(new Uri('https://h:443/')));
        $this->assertSame('http://h/p?x=1', (string) Uri::withQueryValue(new Uri('http://h/p'), 'x', '1'));
        $this->assertSame('http://h/p', (string) Uri::withoutQueryValue(new Uri('http://h/p?x=1'), 'x'));
    }
    public function testUtils(): void
    {
        $this->assertSame("a=1\n", Utils::readLine(Utils::streamFor("a=1\nb=2")));
        $this->assertSame('x', Utils::copyToString(Utils::streamFor('x')));
        $res = Utils::tryFopen('php://memory', 'r+');
        $this->assertIsResource($res);
        $this->assertSame('', Utils::modifyRequest(new Request('GET', 'http://h/'), ['set_headers' => ['X' => '1']])->getBody()->getContents());
    }
    public function testUploadedFile(): void
    {
        $f = new UploadedFile(Utils::streamFor('data'), 4, UPLOAD_ERR_OK, 'a.txt', 'text/plain');
        $this->assertSame('a.txt', $f->getClientFilename());
        $this->assertSame(4, $f->getSize());
        $this->assertSame('data', (string) $f->getStream());
    }
}
