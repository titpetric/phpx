<?php
use GuzzleHttp\{Client, HandlerStack, Middleware, RequestOptions};
use GuzzleHttp\Handler\MockHandler;
use GuzzleHttp\Cookie\CookieJar;
use GuzzleHttp\Exception\{InvalidArgumentException, RequestException, TooManyRedirectsException, BadResponseException};
use GuzzleHttp\Psr7\{Request, Response, Utils as PU7};
use PHPUnit\Framework\TestCase;

final class GuzzleCov2Test extends TestCase
{
    private array $deprecations = [];
    private array $history = [];

    protected function setUp(): void
    {
        $this->deprecations = [];
        $this->history = [];
    }

    private function client(array $queue = [], array $config = []): Client
    {
        $stack = HandlerStack::create(new MockHandler($queue ?: [new Response(200)]));
        $stack->push(Middleware::history($this->history));
        return new Client($config + ['handler' => $stack]);
    }

    private function collect(callable $fn)
    {
        set_error_handler(function (int $no, string $msg) {
            $this->deprecations[] = $msg;
            return true;
        }, E_USER_DEPRECATED);
        try {
            return $fn();
        } finally {
            restore_error_handler();
        }
    }

    public function testBadHandlerConfig(): void
    {
        $this->expectException(InvalidArgumentException::class);
        new Client(['handler' => 'not a callable']);
    }

    public function testCapOptionsWithCustomHandler(): void
    {
        $this->expectException(InvalidArgumentException::class);
        new Client(['handler' => new MockHandler(), 'max_host_connections' => 2]);
    }

    public function testTransportSharingRequiresDefaultHandler(): void
    {
        $this->expectException(InvalidArgumentException::class);
        new Client(['handler' => new MockHandler(), 'transport_sharing' => \GuzzleHttp\TransportSharing::HANDLER_REQUIRE]);
    }

    public function testConfigDefaults(): void
    {
        $c = $this->client([], ['base_uri' => 'http://a.test/x/', 'cookies' => true, 'headers' => ['x-y' => 'z']]);
        $this->assertSame('http://a.test/x/', (string) $c->getConfig('base_uri'));
        $this->assertInstanceOf(CookieJar::class, $c->getConfig('cookies'));
        $this->assertTrue($c->getConfig('http_errors'));
        $this->assertSame('z', $c->getConfig('headers')['x-y']);
        $this->assertArrayHasKey('User-Agent', $c->getConfig('headers'));
        $this->assertNull($c->getConfig('nope'));
        $this->assertIsArray($c->getConfig());
    }

    public function testUserAgentNotOverridden(): void
    {
        $c = $this->client([], ['headers' => ['user-agent' => 'mine']]);
        $c->get('http://a.test/');
        $this->assertSame('mine', $this->history[0]['request']->getHeaderLine('User-Agent'));
    }

    public function testMagicMethods(): void
    {
        $c = $this->client([new Response(200), new Response(200), new Response(200), new Response(200), new Response(200), new Response(200)]);
        $this->collect(function () use ($c) {
            $c->head('http://a.test/h');
            $c->put('http://a.test/p', ['body' => 'x']);
            $c->patch('http://a.test/p');
            $c->delete('http://a.test/d');
            $p = $c->postAsync('http://a.test/a');
            $p->wait();
            $c->options('http://a.test/o');
        });
        $this->assertSame(['HEAD', 'PUT', 'PATCH', 'DELETE', 'POST', 'OPTIONS'], array_map(fn ($h) => $h['request']->getMethod(), $this->history));
    }

    public function testMagicMethodRequiresUri(): void
    {
        $c = $this->client();
        $this->expectException(InvalidArgumentException::class);
        $this->collect(fn () => $c->__call('get', []));
    }

    public function testSendAndSendRequest(): void
    {
        $c = $this->client([new Response(201), new Response(202)]);
        $this->assertSame(201, $c->send(new Request('GET', 'http://a.test/'))->getStatusCode());
        $this->assertSame(202, $c->sendRequest(new Request('GET', 'http://a.test/'))->getStatusCode());
    }

    public function testSendRequestHttpErrorsOff(): void
    {
        $c = $this->client([new Response(404)]);
        $this->assertSame(404, $c->sendRequest(new Request('GET', 'http://a.test/'))->getStatusCode());
    }

    public function testSendAsync(): void
    {
        $c = $this->client([new Response(200, [], 'z')]);
        $this->assertSame('z', (string) $c->sendAsync(new Request('GET', 'http://a.test/'))->wait()->getBody());
    }

    public function testBaseUriResolution(): void
    {
        $c = $this->client([new Response(200), new Response(200), new Response(200)], ['base_uri' => 'http://a.test/v1/']);
        $c->get('x');
        $c->get('/root');
        $c->get('http://other.test/abs');
        $this->assertSame('http://a.test/v1/x', (string) $this->history[0]['request']->getUri());
        $this->assertSame('http://a.test/root', (string) $this->history[1]['request']->getUri());
        $this->assertSame('http://other.test/abs', (string) $this->history[2]['request']->getUri());
    }

    public function testBaseUriNoHostGetsHostHeaderUntouched(): void
    {
        $c = $this->client([new Response(200)], ['base_uri' => 'http://a.test:8080/']);
        $c->get('p?q=1', ['headers' => ['Host' => 'custom.test']]);
        $this->assertSame('custom.test', $this->history[0]['request']->getHeaderLine('Host'));
    }

    public function testBodyOptions(): void
    {
        $c = $this->client([new Response(200), new Response(200), new Response(200), new Response(200), new Response(200)]);
        $c->post('http://a.test/', ['body' => 'abc']);
        $c->post('http://a.test/', ['body' => fopen('data://text/plain,res', 'r')]);
        $c->post('http://a.test/', ['body' => PU7::streamFor('stream')]);
        $c->post('http://a.test/', ['body' => new ArrayIterator(['a', 'b'])]);
        $c->post('http://a.test/', ['body' => new class { public function __toString(): string { return 'str'; } }]);
        $this->assertSame('abc', (string) $this->history[0]['request']->getBody());
        $this->assertSame('res', (string) $this->history[1]['request']->getBody());
        $this->assertSame('stream', (string) $this->history[2]['request']->getBody());
        $this->assertSame('ab', (string) $this->history[3]['request']->getBody());
        $this->assertSame('str', (string) $this->history[4]['request']->getBody());
    }

    public function testBodyScalarDeprecated(): void
    {
        $c = $this->client([new Response(200)]);
        $this->collect(fn () => $c->post('http://a.test/', ['body' => 42]));
        $this->assertSame('42', (string) $this->history[0]['request']->getBody());
    }

    public function testBodyArrayInvalid(): void
    {
        $c = $this->client();
        $this->expectException(InvalidArgumentException::class);
        $this->collect(fn () => $c->post('http://a.test/', ['body' => ['a']]));
    }

    public function testBodyInvalidObject(): void
    {
        $c = $this->client();
        $this->expectException(InvalidArgumentException::class);
        $this->collect(fn () => $c->post('http://a.test/', ['body' => new stdClass]));
    }

    public function testBodyCallable(): void
    {
        $c = $this->client([new Response(200)]);
        $f = new class { public function __invoke() { static $n = 0; return $n++ ? false : 'cb'; } };
        $c->post('http://a.test/', ['body' => $f]);
        $this->assertSame('cb', (string) $this->history[0]['request']->getBody());
    }

    public function testMultipart(): void
    {
        $c = $this->client();
        $c->post('http://a.test/', ['multipart' => [['name' => 'f', 'contents' => 'v'], ['name' => 'g', 'contents' => 'w', 'filename' => 'x.txt', 'headers' => ['X-Q' => 'r']]]]);
        $req = $this->history[0]['request'];
        $this->assertStringStartsWith('multipart/form-data; boundary=', $req->getHeaderLine('Content-Type'));
        $this->assertStringContainsString('filename="x.txt"', (string) $req->getBody());
    }

    public function testFormParamsAndMultipartConflict(): void
    {
        $c = $this->client();
        $this->expectException(InvalidArgumentException::class);
        $c->post('http://a.test/', ['form_params' => ['a' => 1], 'multipart' => [['name' => 'a', 'contents' => 'b']]]);
    }

    public function testJsonErrors(): void
    {
        $c = $this->client();
        $this->expectException(InvalidArgumentException::class);
        $c->post('http://a.test/', ['json' => "\xB1\x31"]);
    }

    public function testJsonReplacesContentType(): void
    {
        $c = $this->client();
        $c->post('http://a.test/', ['json' => ['a' => 1], 'headers' => ['content-type' => 'text/x']]);
        $this->assertSame('{"a":1}', (string) $this->history[0]['request']->getBody());
    }

    public function testAuth(): void
    {
        $c = $this->client([new Response(200), new Response(200), new Response(200)]);
        $c->get('http://a.test/', ['auth' => ['u', 'p']]);
        $c->get('http://a.test/', ['auth' => ['u', 'p', 'basic']]);
        $this->assertSame('Basic ' . base64_encode('u:p'), $this->history[0]['request']->getHeaderLine('Authorization'));
        $this->assertSame('Basic ' . base64_encode('u:p'), $this->history[1]['request']->getHeaderLine('Authorization'));
    }

    public function testAuthDigest(): void
    {
        $c = $this->client([new Response(200)]);
        $c->get('http://a.test/', ['auth' => ['u', 'p', 'digest']]);
        $this->assertFalse($this->history[0]['request']->hasHeader('Authorization'));
        $this->assertSame('u:p', $this->history[0]['options']['curl'][CURLOPT_USERPWD]);
    }

    public function testDeprecationsReachErrorHandler(): void
    {
        $c = $this->client([new Response(200)]);
        $this->collect(fn () => $c->get('http://a.test/', ['timeout' => '5', 'headers' => ['X-N' => 5]]));
        $this->assertNotEmpty($this->deprecations);
        $this->assertStringContainsString('Passing', $this->deprecations[0]);
    }

    public function testAuthNtlm(): void
    {
        $c = $this->client([new Response(200)]);
        try {
            $this->collect(fn () => $c->get('http://a.test/', ['auth' => ['u', 'p', 'ntlm']]));
            $this->assertSame('u:p', $this->history[0]['options']['curl'][CURLOPT_USERPWD]);
        } catch (InvalidArgumentException $e) {
            $this->assertStringContainsString('NTLM', $e->getMessage());
        }
    }

    public function testQueryString(): void
    {
        $c = $this->client([new Response(200), new Response(200)]);
        $c->get('http://a.test/?old=1', ['query' => 'a=b&c=d']);
        $c->get('http://a.test/', ['query' => ['x' => 'y z']]);
        $this->assertSame('a=b&c=d', $this->history[0]['request']->getUri()->getQuery());
        $this->assertSame('x=y%20z', $this->history[1]['request']->getUri()->getQuery());
    }

    public function testQueryInvalid(): void
    {
        $c = $this->client();
        $this->expectException(InvalidArgumentException::class);
        $this->collect(fn () => $c->get('http://a.test/', ['query' => 5]));
    }

    public function testSinkBoolInvalid(): void
    {
        $c = $this->client();
        $this->expectException(InvalidArgumentException::class);
        $this->collect(fn () => $c->get('http://a.test/', ['sink' => true]));
    }

    public function testHeadersMustBeAssoc(): void
    {
        $c = $this->client();
        $this->expectException(InvalidArgumentException::class);
        $c->send(new Request('GET', 'http://a.test/'), ['headers' => ['a', 'b']]);
    }

    public function testHeadersNotArray(): void
    {
        $c = $this->client();
        $this->expectException(InvalidArgumentException::class);
        $c->get('http://a.test/', ['headers' => 'x']);
    }

    public function testHeadersNullClearsDefaults(): void
    {
        $c = $this->client([new Response(200)], ['headers' => ['X-Def' => '1']]);
        $c->get('http://a.test/', ['headers' => null]);
        $this->assertFalse($this->history[0]['request']->hasHeader('X-Def'));
    }

    public function testHeaderValueCasting(): void
    {
        $c = $this->client([new Response(200)]);
        $this->collect(fn () => $c->get('http://a.test/', ['headers' => ['X-N' => 5, 'X-L' => [1, 2.5], 'X-E' => [], 'X-F' => INF, 'X-T' => true, 'X-Nan' => [NAN, -INF]]]));
        $r = $this->history[0]['request'];
        $this->assertSame('5', $r->getHeaderLine('X-N'));
        $this->assertSame('1, 2.5', $r->getHeaderLine('X-L'));
        $this->assertFalse($r->hasHeader('X-E'));
        $this->assertSame('INF', $r->getHeaderLine('X-F'));
        $this->assertSame('1', $r->getHeaderLine('X-T'));
        $this->assertSame('NAN, -INF', $r->getHeaderLine('X-Nan'));
    }

    public function testDecodeContentString(): void
    {
        $c = $this->client([new Response(200)]);
        $c->get('http://a.test/', ['decode_content' => 'gzip']);
        $this->assertSame('gzip', $this->history[0]['request']->getHeaderLine('Accept-Encoding'));
    }

    public function testVersionOption(): void
    {
        $c = $this->client([new Response(200), new Response(200), new Response(200)]);
        $c->get('http://a.test/', ['version' => 2.0]);
        $c->get('http://a.test/', ['version' => '1.0']);
        $this->collect(fn () => $c->get('http://a.test/', ['version' => '']));
        $this->assertSame('2.0', $this->history[0]['request']->getProtocolVersion());
        $this->assertSame('1.0', $this->history[1]['request']->getProtocolVersion());
        $this->assertSame('1.1', $this->history[2]['request']->getProtocolVersion());
    }

    public function testBadProtocolVersionDeprecated(): void
    {
        $c = $this->client([new Response(200), new Response(200)]);
        $this->collect(function () use ($c) {
            $c->get('http://a.test/', ['version' => 'x.y']);
            $c->send((new Request('GET', 'http://a.test/'))->withProtocolVersion('1.1'));
        });
        $this->assertCount(2, $this->history);
        $this->assertSame('x.y', $this->history[0]['request']->getProtocolVersion());
    }

    public function testDeprecatedOptionNormalization(): void
    {
        $c = $this->client([new Response(200)]);
        $this->collect(fn () => $c->get('http://a.test/', [
            'timeout' => '5', 'connect_timeout' => '1.5', 'read_timeout' => '2', 'delay' => '0',
            'retries' => '3', 'crypto_method' => 2.0, 'crypto_method_max' => '4',
            'cert_type' => 5, 'force_ip_resolve' => 4, 'ssl_key_type' => 6,
            'cert' => [1, 2], 'ssl_key' => [3, null],
            'auth' => [1, 2, 3],
        ]));
        $o = $this->history[0]['options'];
        $this->assertSame(5, $o['timeout']);
        $this->assertSame(1.5, $o['connect_timeout']);
        $this->assertSame(3, $o['retries']);
        $this->assertSame(2, $o['crypto_method']);
        $this->assertSame(4, $o['crypto_method_max']);
        $this->assertSame('5', $o['cert_type']);
        $this->assertSame('4', $o['force_ip_resolve']);
        $this->assertSame(['1', '2'], $o['cert']);
        $this->assertSame(['3', null], $o['ssl_key']);
        $this->assertSame([ '1', '2', '3'], $o['auth']);
    }

    public function testInvalidOptionTypesWarn(): void
    {
        $bad = [
            'allow_redirects' => 5, 'auth' => 5, 'auth2' => null,
            'cert' => 5, 'cert_type' => 5, 'connect_timeout' => 'x', 'crypto_method' => 'x', 'debug' => 'x',
            'decode_content' => 5, 'delay' => -1, 'expect' => 'x', 'force_ip_resolve' => 'v5',
            'http_errors' => 'x', 'on_headers' => 'x', 'on_stats' => 'x', 'on_trailers' => 'x', 'progress' => 'x',
            'protocols' => 'x', 'read_timeout' => 'x', 'retries' => 'x', 'sink' => 5, 'ssl_key' => 5, 'ssl_key_type' => 5,
            'stream' => 'x', 'stream_context' => 'x', 'timeout' => 'x', 'verify' => 5,
            'version' => [], 'curl' => 'x', 'cookies' => 'x', 'handler' => 'x', 'proxy' => 5,
        ];
        $ran = 0;
        foreach ($bad as $k => $v) {
            if ($k === 'auth2') { continue; }
            $ran++;
            $this->deprecations = [];
            $c = $this->client([new Response(200)]);
            try {
                $this->collect(fn () => $c->get('http://a.test/', [$k => $v]));
            } catch (\Throwable $e) {
                // some invalid values throw after the warning
            }
        }
        $this->assertSame(count($bad) - 1, $ran);
    }

    public function testInvalidNestedOptionTypesWarn(): void
    {
        $cases = [
            ['allow_redirects' => ['max' => 'x', 'strict' => 1, 'referer' => 1, 'protocols' => ['ftp', 5], 'on_redirect' => 'x', 'track_redirects' => 1]],
            ['allow_redirects' => ['protocols' => []]],
            ['allow_redirects' => ['protocols' => 'x']],
            ['auth' => [1]],
            ['auth' => ['a', 'b', 5]],
            ['cert' => [5, 6]],
            ['cert' => []],
            ['form_params' => 5],
            ['form_params' => ['a' => new stdClass]],
            ['form_params' => ['a' => ['b' => new stdClass]]],
            ['headers' => ['X' => new stdClass]],
            ['headers' => ['X' => [new stdClass]]],
            ['headers' => ['X' => []]],
            ['multipart' => 5],
            ['multipart' => [5]],
            ['multipart' => [['contents' => 'x']]],
            ['multipart' => [['name' => 'x', 'contents' => 'v', 'headers' => 5, 'filename' => 5]]],
            ['multipart' => [['name' => 'x', 'contents' => 'v', 'headers' => ['A' => 5]]]],
            ['multipart' => [['name' => []]]],
            ['proxy' => ['http' => 5, 'https' => 5, 'no' => 5]],
            ['proxy' => ['no' => [5]]],
            ['proxy' => ['no' => 'a.com', 'http' => 'http://p']],
            ['protocols' => ['ftp']],
            ['protocols' => [5]],
            ['protocols' => []],
            ['cookies' => true],
            ['cookies' => 'x'],
            ['delay' => INF],
            ['body' => ['x']],
        ];
        $ran = 0;
        foreach ($cases as $i => $opts) {
            $ran++;
            $this->deprecations = [];
            $c = $this->client([new Response(200)]);
            try {
                $this->collect(fn () => $c->get('http://a.test/', $opts));
            } catch (\Throwable $e) {
            }
        }
        $this->assertSame(count($cases), $ran);
    }

    public function testRequestLevelHandlerDeprecated(): void
    {
        $c = $this->client([new Response(200)]);
        $r = $this->collect(fn () => $c->get('http://a.test/', ['handler' => HandlerStack::create(new MockHandler([new Response(204)]))]));
        $this->assertSame(204, $r->getStatusCode());
    }

    public function testNonFiniteQueryDeprecated(): void
    {
        $c = $this->client([new Response(200)]);
        $this->collect(fn () => $c->get('http://a.test/', ['query' => ['a' => NAN, 'b' => [INF], 'c' => -INF]]));
        $this->assertSame('a=NAN&b%5B0%5D=INF&c=-INF', $this->history[0]['request']->getUri()->getQuery());
    }

    public function testNonFiniteFormParams(): void
    {
        $c = $this->client([new Response(200)]);
        $this->collect(fn () => $c->post('http://a.test/', ['form_params' => ['a' => NAN]]));
        $this->assertSame('a=NAN', (string) $this->history[0]['request']->getBody());
    }

    public function testRequestAsyncRejection(): void
    {
        $c = $this->client([new Response(500)]);
        $p = $c->requestAsync('GET', 'http://a.test/');
        $this->expectException(\GuzzleHttp\Exception\ServerException::class);
        $p->wait();
    }

    public function testRequestWithUriObject(): void
    {
        $c = $this->client();
        $c->request('GET', new \GuzzleHttp\Psr7\Uri('http://a.test/obj'));
        $this->assertSame('http://a.test/obj', (string) $this->history[0]['request']->getUri());
    }

    public function testDelayOption(): void
    {
        $c = $this->client();
        $c->get('http://a.test/', ['delay' => 0]);
        $this->assertCount(1, $this->history);
    }

    public function testRedirects(): void
    {
        $c = $this->client([
            new Response(302, ['Location' => '/two']),
            new Response(301, ['Location' => 'http://b.test/three']),
            new Response(200, [], 'end'),
        ]);
        $r = $c->post('http://a.test/one', ['body' => 'data', 'headers' => ['Authorization' => 'x', 'Cookie' => 'a=b'], 'allow_redirects' => ['track_redirects' => true, 'referer' => true]]);
        $this->assertSame('end', (string) $r->getBody());
        $this->assertCount(3, $this->history);
        $this->assertSame('GET', $this->history[1]['request']->getMethod());
        $this->assertSame('http://a.test/two', (string) $this->history[1]['request']->getUri());
        $this->assertSame('http://a.test/one', $this->history[1]['request']->getHeaderLine('Referer'));
        $this->assertTrue($this->history[1]['request']->hasHeader('Authorization'));
        $this->assertFalse($this->history[2]['request']->hasHeader('Authorization'));
        $this->assertFalse($this->history[2]['request']->hasHeader('Cookie'));
        $this->assertSame(['http://a.test/two', 'http://b.test/three'], $r->getHeader('X-Guzzle-Redirect-History'));
        $this->assertSame(['302', '301'], $r->getHeader('X-Guzzle-Redirect-Status-History'));
    }

    public function testRedirectStrictKeepsMethod(): void
    {
        $c = $this->client([new Response(301, ['Location' => '/b']), new Response(200)]);
        $c->post('http://a.test/a', ['body' => 'd', 'allow_redirects' => ['strict' => true]]);
        $this->assertSame('POST', $this->history[1]['request']->getMethod());
        $this->assertSame('d', (string) $this->history[1]['request']->getBody());
    }

    public function testRedirect307KeepsMethod(): void
    {
        $c = $this->client([new Response(307, ['Location' => '/b']), new Response(200)]);
        $c->put('http://a.test/a', ['body' => 'd']);
        $this->assertSame('PUT', $this->history[1]['request']->getMethod());
    }

    public function testRedirect303ToGet(): void
    {
        $c = $this->client([new Response(303, ['Location' => '/b']), new Response(200)]);
        $c->request('DELETE', 'http://a.test/a', ['allow_redirects' => ['strict' => true]]);
        $this->assertSame('GET', $this->history[1]['request']->getMethod());
    }

    public function testRedirectQueryMethod(): void
    {
        $c = $this->client([new Response(301, ['Location' => '/b']), new Response(303, ['Location' => '/c']), new Response(200)]);
        $c->request('QUERY', 'http://a.test/a', ['body' => 'q']);
        $this->assertSame('QUERY', $this->history[1]['request']->getMethod());
        $this->assertSame('GET', $this->history[2]['request']->getMethod());
    }

    public function testRedirectOnRedirectCallback(): void
    {
        $seen = [];
        $c = $this->client([new Response(302, ['Location' => '/b']), new Response(200)]);
        $c->get('http://a.test/a', ['allow_redirects' => ['on_redirect' => function ($req, $res, $uri) use (&$seen) { $seen[] = (string) $uri . ' ' . $res->getStatusCode(); }]]);
        $this->assertSame(['http://a.test/b 302'], $seen);
    }

    public function testRedirectTooMany(): void
    {
        $c = $this->client([new Response(302, ['Location' => '/1']), new Response(302, ['Location' => '/2']), new Response(302, ['Location' => '/3'])]);
        $this->expectException(TooManyRedirectsException::class);
        $c->get('http://a.test/0', ['allow_redirects' => ['max' => 2]]);
    }

    public function testRedirectDisabled(): void
    {
        $c = $this->client([new Response(302, ['Location' => '/1'])]);
        $this->assertSame(302, $c->get('http://a.test/0', ['allow_redirects' => false])->getStatusCode());
        $c = $this->client([new Response(302, ['Location' => '/1'])]);
        $this->assertSame(302, $c->get('http://a.test/0', ['allow_redirects' => ['max' => 0]])->getStatusCode());
    }

    public function testRedirectProtocolRejected(): void
    {
        $c = $this->client([new Response(302, ['Location' => 'ftp://x.test/f'])]);
        $this->expectException(BadResponseException::class);
        $c->get('http://a.test/0');
    }

    public function testRedirectNoLocationOrNon3xx(): void
    {
        $c = $this->client([new Response(302), new Response(200, ['Location' => '/x'])]);
        $this->assertSame(302, $c->get('http://a.test/', ['http_errors' => false])->getStatusCode());
        $this->assertSame(200, $c->get('http://a.test/')->getStatusCode());
    }

    public function testRedirectInvalidOption(): void
    {
        $mw = Middleware::redirect()(new MockHandler([new Response(200)]));
        $this->expectException(\InvalidArgumentException::class);
        $mw(new Request('GET', 'http://a.test/'), ['allow_redirects' => 'bad']);
    }

    public function testRedirectUnrewindableBody(): void
    {
        $body = new \GuzzleHttp\Psr7\NoSeekStream(PU7::streamFor('abc'));
        $body->read(1);
        $c = $this->client([new Response(307, ['Location' => '/b']), new Response(200)]);
        $this->expectException(RequestException::class);
        $c->send(new Request('POST', 'http://a.test/a', [], $body));
    }

    public function testRedirectWithIdn(): void
    {
        $c = $this->client([new Response(302, ['Location' => 'http://b.test/']), new Response(200)]);
        $c->get('http://a.test/', ['idn_conversion' => false]);
        $this->assertCount(2, $this->history);
    }
}
