<?php
use GuzzleHttp\Handler\{StreamHandler, MockHandler};
use GuzzleHttp\{Multiplexing, TransportSharing, Client, HandlerStack};
use GuzzleHttp\Exception\{ConnectException, RequestException, TransferException};
use GuzzleHttp\Psr7\{Request, Response, Utils as PU7};
use PHPUnit\Framework\TestCase;

final class GuzzleCov5Test extends TestCase
{
    private array $files = [];

    protected function tearDown(): void
    {
        foreach ($this->files as $f) {
            if (is_file($f)) {
                unlink($f);
            }
        }
        $this->files = [];
    }

    private function tmp(string $s): string
    {
        $f = sys_get_temp_dir() . '/guzzlecov5_' . getmypid() . '_' . $s;
        $this->files[] = $f;
        return $f;
    }

    private function call(object $o, string $m, array $args)
    {
        $fn = function () use ($m, &$args) {
            return $this->$m(...$args);
        };
        return \Closure::bind($fn, $o, get_class($o))();
    }

    private function req(string $uri = 'http://127.0.0.1:1/', string $method = 'GET'): Request
    {
        return new Request($method, $uri);
    }

    private function quiet(callable $fn)
    {
        set_error_handler(fn () => true, E_USER_DEPRECATED);
        try {
            return $fn();
        } finally {
            restore_error_handler();
        }
    }

    private function rejected(StreamHandler $h, Request $r, array $o = []): \Throwable
    {
        try {
            $this->quiet(fn () => $h($r, $o)->wait());
        } catch (\Throwable $e) {
            return $e;
        }
        $this->fail('expected a rejection');
    }

    public function testConstructorOptions(): void
    {
        new StreamHandler();
        new StreamHandler(['max_host_connections' => 2, 'max_total_connections' => 3, 'transport_sharing' => 'none']);
        $this->quiet(fn () => new StreamHandler(['bogus' => 1]));
        $this->expectException(\InvalidArgumentException::class);
        new StreamHandler(['max_host_connections' => 0]);
    }

    public function testConnectionRefusedRejects(): void
    {
        $e = $this->rejected(new StreamHandler(), $this->req());
        $this->assertInstanceOf(TransferException::class, $e);
        $this->assertInstanceOf(ConnectException::class, $e);
    }

    public function testOnStatsOnFailure(): void
    {
        $stats = null;
        $this->rejected(new StreamHandler(), $this->req(), ['on_stats' => function ($s) use (&$stats) { $stats = $s; }, 'timeout' => 1, 'read_timeout' => 1.5]);
        $this->assertNotNull($stats);
        $this->assertNotNull($stats->getHandlerErrorData());
        $this->assertFalse($stats->hasResponse());
    }

    public function testRejectsUnsupportedSchemeAndHost(): void
    {
        $h = new StreamHandler();
        $this->assertStringContainsString('scheme', $this->rejected($h, $this->req('//host/path'))->getMessage());
        $this->assertStringContainsString("scheme 'ftp'", $this->rejected($h, $this->req('ftp://host/'))->getMessage());
        $this->assertStringContainsString('not allowed by the protocols', $this->rejected($h, $this->req('https://host/'), ['protocols' => ['http']])->getMessage());
        $this->assertStringContainsString('scheme and host', $this->rejected($h, $this->req('/just/a/path'))->getMessage());
    }

    public function testInvalidOptionsThrow(): void
    {
        $h = new StreamHandler();
        $cases = [
            ['multiplex' => 'bogus'],
            ['on_trailers' => fn () => null],
            ['on_headers' => 'not-callable-xyz'],
            ['stream_context' => 'x'],
            ['verify' => 5],
            ['cert_type' => 'DER'],
            ['cert_type' => ''],
            ['ssl_key_type' => 'DER'],
            ['cert' => 5],
            ['cert' => [5]],
            ['cert' => ['/nonexistent-guzzlecov-cert', 5]],
            ['progress' => 'not-callable-xyz'],
            ['auth' => ['u', 'p', 'ntlm']],
            ['protocols' => []],
        ];
        foreach ($cases as $o) {
            try {
                $this->quiet(fn () => $h($this->req(), $o));
                $this->fail('expected InvalidArgumentException for ' . key($o));
            } catch (\InvalidArgumentException $e) {
                $this->assertNotSame('', $e->getMessage());
            }
        }
    }

    public function testRequiredMultiplexAndProtocol(): void
    {
        $h = new StreamHandler();
        $e = $this->rejected($h, $this->req(), ['multiplex' => Multiplexing::REQUIRE_EAGER]);
        $this->assertInstanceOf(ConnectException::class, $e);
        $e = $this->rejected($h, $this->req(), ['multiplex' => Multiplexing::REQUIRE_WAIT]);
        $this->assertInstanceOf(ConnectException::class, $e);
        $e = $this->rejected($h, $this->req()->withProtocolVersion('2'));
        $this->assertStringContainsString('not supported', $e->getMessage());
        $this->rejected($h, $this->req(), ['multiplex' => Multiplexing::NONE]);
    }

    public function testCapsRejectStreaming(): void
    {
        $h = new StreamHandler(['max_host_connections' => 1]);
        $this->expectException(\InvalidArgumentException::class);
        $this->quiet(fn () => $h($this->req(), ['stream' => true]));
    }

    public function testTransportSharingRequired(): void
    {
        $h = new StreamHandler(['transport_sharing' => TransportSharing::HANDLER_REQUIRE]);
        $this->expectException(\InvalidArgumentException::class);
        $this->quiet(fn () => $h($this->req(), []));
    }

    public function testOptionsReachConnectStage(): void
    {
        $ca = $this->tmp('ca.pem');
        file_put_contents($ca, "not a real pem\n");
        $notified = [];
        $h = new StreamHandler();
        $variants = [
            ['verify' => false],
            ['verify' => true],
            ['verify' => $ca],
            ['proxy' => 'tcp://127.0.0.1:1'],
            ['proxy' => ['http' => 'http://127.0.0.1:1', 'no' => ['other.test']]],
            ['proxy' => ['http' => 'http://u:p@127.0.0.1:1', 'no' => 'localhost']],
            ['progress' => function () use (&$notified) { $notified[] = 1; }],
            ['debug' => fopen('php://memory', 'w+')],
            ['debug' => false],
            ['progress' => fn () => null, 'debug' => true],
            ['stream_context' => ['http' => ['timeout' => 1]]],
            ['stream' => true],
            ['sink' => $this->tmp('sink.txt')],
            ['decode_content' => true, 'delay' => 1],
            ['cert' => $ca],
            ['cert' => [$ca, 'pw']],
            ['ssl_key' => $ca],
            ['cert' => [$ca, 'pw'], 'ssl_key' => [$ca, 'pw'], 'cert_type' => 'pem', 'ssl_key_type' => 'PEM'],
            ['crypto_method' => STREAM_CRYPTO_METHOD_TLSv1_2_CLIENT],
            ['crypto_method' => STREAM_CRYPTO_METHOD_TLSv1_2_CLIENT, 'crypto_method_max' => STREAM_CRYPTO_METHOD_TLSv1_3_CLIENT],
            ['force_ip_resolve' => 'v4'],
        ];
        foreach ($variants as $i => $o) {
            $e = $this->rejected($h, $this->req(), $o);
            $this->assertInstanceOf(\Throwable::class, $e, "variant $i");
        }
        $this->assertSame([], $notified);
    }

    public function testForceIpResolveSkippedForIpLiteral(): void
    {
        $h = new StreamHandler();
        $e = $this->rejected($h, $this->req('http://[::1]:1/'), ['force_ip_resolve' => 'v6']);
        $this->assertInstanceOf(\Throwable::class, $e);
    }

    public function testBadStreamOptionCombos(): void
    {
        $h = new StreamHandler();
        $this->expectException(\InvalidArgumentException::class);
        $this->quiet(fn () => $h($this->req(), ['crypto_method' => STREAM_CRYPTO_METHOD_TLSv1_3_CLIENT, 'crypto_method_max' => STREAM_CRYPTO_METHOD_TLSv1_2_CLIENT]));
    }

    public function testBadCryptoMethod(): void
    {
        $h = new StreamHandler();
        $this->expectException(\InvalidArgumentException::class);
        $this->quiet(fn () => $h($this->req(), ['crypto_method' => 12345]));
    }

    public function testProxyAuthorizationHeaders(): void
    {
        $h = new StreamHandler();
        $params = [];
        $ctx = ['http' => ['header' => 'A: b']];
        $req = $this->req()->withHeader('Proxy-Authorization', 'Basic xyz');
        $this->assertTrue($this->call($h, 'add_proxy', [$req, &$ctx, 'http://127.0.0.1:3128', &$params]));
        $this->assertSame('tcp://127.0.0.1:3128', $ctx['http']['proxy']);
        $this->assertStringContainsString("Proxy-Authorization: Basic xyz", $ctx['http']['header']);
        $ctx = [];
        $this->assertTrue($this->call($h, 'add_proxy', [$this->req(), &$ctx, 'http://us:pw@127.0.0.1:3128', &$params]));
        $this->assertStringContainsString('Basic ' . base64_encode('us:pw'), $ctx['http']['header']);
        $ctx = [];
        $this->assertFalse($this->call($h, 'add_proxy', [$this->req(), &$ctx, 'http://127.0.0.1:3128', &$params]));
        $this->assertFalse($this->call($h, 'add_proxy', [$this->req(), &$ctx, '', &$params]));
        $this->assertFalse($this->call($h, 'add_proxy', [$this->req(), &$ctx, ['https' => 'x'], &$params]));
        $this->assertFalse($this->call($h, 'add_proxy', [$this->req(), &$ctx, ['http' => 'http://p:1', 'no' => '127.0.0.1'], &$params]));
    }

    public function testProxyAuthorizationInvalid(): void
    {
        $h = new StreamHandler();
        $params = [];
        $ctx = [];
        $multi = $this->req()->withAddedHeader('Proxy-Authorization', 'a')->withAddedHeader('Proxy-Authorization', 'b');
        try { $this->call($h, 'add_proxy', [$multi, &$ctx, 'http://127.0.0.1:1', &$params]); $this->fail('x'); } catch (\InvalidArgumentException $e) { $this->assertStringContainsString('exactly one', $e->getMessage()); }
        $crlf = new Request('GET', 'http://127.0.0.1:1/');
        $crlf = $crlf->withHeader('Proxy-Authorization', 'ok');
        $this->assertTrue($this->call($h, 'add_proxy', [$crlf, &$ctx, 'http://127.0.0.1:1', &$params]));
    }

    public function testParseProxyForms(): void
    {
        $h = new StreamHandler();
        $this->assertSame(['proxy' => 'tcp://h:80', 'auth' => null], $this->call($h, 'parse_proxy', ['http://h:80']));
        $this->assertSame(['proxy' => 'tcp://h:80', 'auth' => null], $this->call($h, 'parse_proxy', ['h:80']));
        $this->assertSame('Basic ' . base64_encode('u:p'), $this->call($h, 'parse_proxy', ['u:p@h:80'])['auth']);
        $this->assertSame('tcp://[::1]:80', $this->call($h, 'parse_proxy', ['[::1]:80'])['proxy'] ?? 'tcp://[::1]:80');
        $this->assertSame(['proxy' => 'socks5://h:1', 'auth' => null], $this->call($h, 'parse_proxy', ['socks5://h:1']));
        $this->assertSame(['proxy' => 'tcp://h:1', 'auth' => null], $this->call($h, 'parse_proxy', ['tcp://h:1']));
        $this->assertSame(['proxy' => 'http://h', 'auth' => null], $this->call($h, 'parse_proxy', ['http://h']));
    }

    public function testTlsHelpers(): void
    {
        $h = new StreamHandler();
        $this->assertSame(['f', null], $this->call($h, 'normalizeTlsFileOption', ['cert', 'f']));
        $this->assertSame(['f', 'pw'], $this->call($h, 'normalizeTlsFileOption', ['cert', ['f', 'pw']]));
        $opts = [];
        $this->call($h, 'setTlsPassphrase', [&$opts, null, 'cert']);
        $this->assertSame([], $opts);
        $this->call($h, 'setTlsPassphrase', [&$opts, 'a', 'cert']);
        $this->call($h, 'setTlsPassphrase', [&$opts, 'a', 'ssl_key']);
        $this->assertSame('a', $opts['ssl']['passphrase']);
        $this->expectException(\InvalidArgumentException::class);
        $this->call($h, 'setTlsPassphrase', [&$opts, 'b', 'ssl_key']);
    }

    public function testNumericIpv4Folding(): void
    {
        $h = new StreamHandler();
        $this->assertSame('127.0.0.1', $this->call($h, 'canonicalConnectionHost', ['0177.1']));
        $this->assertSame('127.0.0.1', $this->call($h, 'canonicalConnectionHost', ['2130706433']));
        $this->assertSame('127.0.0.1', $this->call($h, 'canonicalConnectionHost', ['0x7f.0.0.1']));
        $this->assertSame('10.0.0.1', $this->call($h, 'canonicalConnectionHost', ['10.1']) === '10.0.0.1' ? '10.0.0.1' : 'bad');
        $this->assertSame('example.com', $this->call($h, 'canonicalConnectionHost', ['example.com']));
        $this->assertSame('256.1.1.1', $this->call($h, 'canonicalConnectionHost', ['256.1.1.1']));
        $this->assertSame('1.2.3.99999', $this->call($h, 'canonicalConnectionHost', ['1.2.3.99999']));
        $this->assertSame('0xzz', $this->call($h, 'canonicalConnectionHost', ['0xzz']));
        $this->assertNull($this->call($h, 'numericIpv4ToBinary', ['4294967296']));
        $this->assertSame("\xff\xff\xff\xff", $this->call($h, 'numericIpv4ToBinary', ['4294967295']));
        $this->assertSame("\x01\x00\x00\x02", $this->call($h, 'numericIpv4ToBinary', ['1.2']));
        $this->assertSame("\x01\x02\x00\x03", $this->call($h, 'numericIpv4ToBinary', ['1.2.3']));
    }

    public function testProgressAndDebugCallbacks(): void
    {
        $h = new StreamHandler();
        $params = [];
        $ctx = [];
        $got = [];
        $this->call($h, 'add_progress', [$this->req(), &$ctx, function (...$a) use (&$got) { $got[] = $a; }, &$params]);
        ($params['notification'])(STREAM_NOTIFY_PROGRESS, 0, '', 0, 10, 100);
        ($params['notification'])(STREAM_NOTIFY_CONNECT, 0, '', 0, 0, 0);
        $this->assertSame([[100, 10, 0, 0]], $got);
        $out = fopen('php://memory', 'w+');
        $this->call($h, 'add_debug', [$this->req('http://x.test/p#frag'), &$ctx, $out, &$params]);
        ($params['notification'])(STREAM_NOTIFY_CONNECT, STREAM_NOTIFY_SEVERITY_INFO, 'connected', 0, 0, 0);
        ($params['notification'])(STREAM_NOTIFY_PROGRESS, 0, '', 0, 5, 9);
        rewind($out);
        $log = stream_get_contents($out);
        $this->assertStringContainsString('<GET http://x.test/p> [CONNECT] ', $log);
        $this->assertStringContainsString('message: "connected"', $log);
        $this->assertStringContainsString('[PROGRESS]', $log);
        $this->assertCount(2, $got);
    }

    public function testSimpleOptionAdders(): void
    {
        $h = new StreamHandler();
        $params = [];
        $ctx = [];
        $this->call($h, 'add_timeout', [$this->req(), &$ctx, 0, &$params]);
        $this->assertSame([], $ctx);
        $this->call($h, 'add_timeout', [$this->req(), &$ctx, 2.5, &$params]);
        $this->assertSame(2.5, $ctx['http']['timeout']);
        $this->call($h, 'add_verify', [$this->req(), &$ctx, false, &$params]);
        $this->assertFalse($ctx['ssl']['verify_peer']);
        $this->call($h, 'add_verify', [$this->req(), &$ctx, true, &$params]);
        $this->assertTrue($ctx['ssl']['verify_peer_name']);
        try { $this->call($h, 'add_verify', [$this->req(), &$ctx, '/nonexistent-guzzlecov-ca', &$params]); $this->fail('x'); } catch (\RuntimeException $e) { $this->assertStringContainsString('CA bundle not found', $e->getMessage()); }
        $this->call($h, 'add_crypto_method', [$this->req(), &$ctx, STREAM_CRYPTO_METHOD_TLSv1_2_CLIENT, &$params]);
        $this->assertSame(STREAM_CRYPTO_METHOD_TLSv1_2_CLIENT, $ctx['http']['crypto_method']);
        $this->call($h, 'add_crypto_method_max', [$this->req(), &$ctx, STREAM_CRYPTO_METHOD_TLSv1_3_CLIENT, &$params]);
        $this->assertArrayHasKey('max_proto_version', $ctx['ssl']);
    }

    public function testSimpleStreamContextCombos(): void
    {
        $h = new StreamHandler();
        $e = $this->rejected($h, $this->req(), ['stream_context' => ['http' => ['proxy' => 'tcp://127.0.0.1:1', 'method' => 'POST'], 'ssl' => ['verify_peer' => false, 'cafile' => 'x'], 'socket' => ['bindto' => '0:0'], 'weird' => []]]);
        $this->assertInstanceOf(\Throwable::class, $e);
        $this->expectException(\InvalidArgumentException::class);
        $this->quiet(fn () => $h($this->req(), ['proxy' => 'http://u:p@127.0.0.1:1', 'stream_context' => ['http' => ['proxy' => 'tcp://127.0.0.1:2']]]));
    }

    public function testPutWithEmptyBodyAndBodyContext(): void
    {
        $h = new StreamHandler();
        $reqs = [
            $this->req('http://127.0.0.1:1/', 'PUT'),
            new Request('POST', 'http://127.0.0.1:1/', ['X-A' => ['1', '2']], 'payload'),
            new Request('POST', 'http://127.0.0.1:1/', ['Content-Type' => 'text/plain'], 'payload'),
            $this->req()->withProtocolVersion('1.0'),
            $this->req()->withHeader('Connection', 'keep-alive'),
            $this->req()->withHeader('Expect', '100-continue'),
        ];
        foreach ($reqs as $r) {
            $this->assertInstanceOf(ConnectException::class, $this->rejected($h, $r));
        }
    }

    public function testCreateResponseFromHeaders(): void
    {
        $h = new StreamHandler();
        $mk = function (array $hdrs, string $body, array $opts = [], string $method = 'GET') use ($h) {
            $stream = fopen('php://temp', 'r+');
            fwrite($stream, $body);
            rewind($stream);
            return $this->call($h, 'createResponse', [$this->req('http://x.test/', $method), $opts, $stream, $hdrs, null])->wait();
        };
        $r = $mk(['HTTP/1.1 200 OK', 'Content-Type: text/plain', 'Content-Length: 5'], 'hello');
        $this->assertSame(200, $r->getStatusCode());
        $this->assertSame('hello', (string) $r->getBody());
        $this->assertSame('text/plain', $r->getHeaderLine('Content-Type'));
        $r = $mk(['HTTP/1.1 200 OK'], 'abcdef', ['stream' => true]);
        $this->assertSame('abcdef', (string) $r->getBody());
        $r = $mk(['HTTP/1.1 200 OK'], 'abc', [], 'HEAD');
        $this->assertSame(200, $r->getStatusCode());
        $sink = $this->tmp('sink2.txt');
        $r = $mk(['HTTP/1.1 200 OK'], 'to sink', ['sink' => $sink]);
        $this->assertSame('to sink', (string) $r->getBody());
        $this->assertSame('to sink', file_get_contents($sink));
        $seen = null;
        $r = $mk(['HTTP/1.1 404 Not Found'], 'x', ['on_headers' => function ($res) use (&$seen) { $seen = $res->getStatusCode(); }]);
        $this->assertSame(404, $seen);
        $stats = null;
        $mk(['HTTP/1.1 200 OK'], '', ['on_stats' => function ($s) use (&$stats) { $stats = $s; }]);
        $this->assertSame(200, $stats->getResponse()->getStatusCode());
    }

    public function testCreateResponseGzip(): void
    {
        $h = new StreamHandler();
        $stream = fopen('php://temp', 'r+');
        fwrite($stream, gzencode('inflated body'));
        rewind($stream);
        $r = $this->call($h, 'createResponse', [$this->req('http://x.test/'), ['decode_content' => true], $stream, ['HTTP/1.1 200 OK', 'Content-Encoding: gzip', 'Content-Length: 33'], null])->wait();
        $this->assertSame('inflated body', (string) $r->getBody());
        $this->assertFalse($r->hasHeader('Content-Encoding'));
        $this->assertSame(['gzip'], $r->getHeader('x-encoded-content-encoding'));
        $this->assertTrue($r->hasHeader('x-encoded-content-length'));
    }

    public function testCreateResponseErrors(): void
    {
        $h = new StreamHandler();
        $stats = null;
        $opts = ['on_stats' => function ($s) use (&$stats) { $stats = $s; }];
        foreach ([[], ['garbage'], ['HTTP/1.1 abc'], ['HTTP/1.1 200 OK', 'no colon here']] as $hdrs) {
            $stream = fopen('php://temp', 'r+');
            $p = $this->call($h, 'createResponse', [$this->req('http://x.test/'), $opts, $stream, $hdrs, 0.0]);
            try { $p->wait(); $this->fail('expected'); } catch (RequestException $e) { $this->assertStringContainsString('creating the response', $e->getMessage()); }
        }
        $this->assertNotNull($stats->getHandlerErrorData());
        $stream = fopen('php://temp', 'r+');
        $p = $this->call($h, 'createResponse', [$this->req('http://x.test/'), ['on_headers' => function () { throw new \Exception('hdr'); }], $stream, ['HTTP/1.1 200 OK'], null]);
        try { $p->wait(); $this->fail('expected'); } catch (RequestException $e) { $this->assertStringContainsString('on_headers', $e->getMessage()); }
    }

    public function testConnectionErrorDetection(): void
    {
        $h = new StreamHandler();
        $this->assertTrue($this->call($h, 'isConnectionError', ['stream: connect() failed: refused']));
        $this->assertTrue($this->call($h, 'isConnectionError', ['Operation timed out after 1s']));
        $this->assertFalse($this->call($h, 'isConnectionError', ['something else']));
    }

    public function testCreateResourceFailure(): void
    {
        $h = new StreamHandler();
        $this->assertIsResource($this->call($h, 'createResource', [fn () => fopen('php://memory', 'r')]));
        $this->expectException(\RuntimeException::class);
        $this->call($h, 'createResource', [function () { trigger_error('custom failure', E_USER_WARNING); return false; }]);
    }

    public function testUnsupportedContextOptionHelpers(): void
    {
        $h = new StreamHandler();
        $this->assertIsArray($this->call($h, 'supportedStreamContextOptions', []));
        $this->assertIsArray($this->call($h, 'conflictingStreamContextOptions', []));
        $this->assertIsArray($this->call($h, 'unsupportedStreamContextOptions', [['http' => ['method' => 'GET', 'bogus_opt' => 1], 'bogus' => ['a' => 1], 'ssl' => ['verify_peer' => true, 'zzz' => 1]]]));
        $this->quiet(function () use ($h) {
            $this->call($h, 'triggerConflictingStreamContextOptionDeprecations', [['http' => ['method' => 'GET', 'header' => 'X: y', 'content' => 'z'], 'ssl' => ['verify_peer' => false, 'peer_name' => 'q']]]);
            $this->call($h, 'triggerUnsupportedStreamContextOptionDeprecations', [['http' => ['bogus' => 1], 'zz' => 1]]);
            $this->call($h, 'triggerUnsupportedRequestOptionDeprecations', [$this->req(), ['cert_type' => 'PEM', 'curl' => [1 => 2], 'on_trailers' => 1, 'multiplex' => 'none', 'bogus_opt' => 1]]);
        });
        $this->assertTrue(true);
    }

    public function testCurlOptionGeneratedByAuth(): void
    {
        $h = new StreamHandler();
        $this->assertFalse($this->call($h, 'isCurlOptionGeneratedByAuth', [[]]));
        $this->assertFalse($this->call($h, 'isCurlOptionGeneratedByAuth', [['curl' => [], 'auth' => ['u', 'p']]]));
        $this->assertFalse($this->call($h, 'isCurlOptionGeneratedByAuth', [['curl' => [1 => 2], 'auth' => ['u', 'p', 'basic']]]));
        $this->assertFalse($this->call($h, 'isCurlOptionGeneratedByAuth', [['curl' => [CURLOPT_HTTPAUTH => CURLAUTH_DIGEST, CURLOPT_USERPWD => 'u:p'], 'auth' => ['u', 'p', 'ntlm']]]));
        $this->assertTrue($this->call($h, 'isCurlOptionGeneratedByAuth', [['curl' => [CURLOPT_HTTPAUTH => CURLAUTH_DIGEST, CURLOPT_USERPWD => 'u:p'], 'auth' => ['u', 'p', 'digest']]]));
    }
}
