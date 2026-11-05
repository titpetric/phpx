<?php
use GuzzleHttp\{Utils, TransferStats, Client, HandlerStack, RequestOptions};
use GuzzleHttp\Handler\{HostValidator, HeaderProcessor, TlsVersion, Proxy, ProxyEnvironment, CurlVersion, EasyHandle, MockHandler};
use GuzzleHttp\Exception\{InvalidArgumentException, RequestException};
use GuzzleHttp\Promise\{FulfilledPromise};
use GuzzleHttp\Psr7\{Request, Response, Uri};
use PHPUnit\Framework\TestCase;

final class GuzzleCovTest extends TestCase
{
    public function testDescribeType(): void
    {
        $this->assertSame('object(stdClass)', Utils::describeType(new stdClass));
        $this->assertSame('array(2)', Utils::describeType([1, 2]));
        $this->assertSame('string(2) "ab"', Utils::describeType('ab'));
        $this->assertSame('int(3)', Utils::describeType(3));
        $this->assertSame('bool(true)', Utils::describeType(true));
        $this->assertSame('NULL', Utils::describeType(null));
        $this->assertStringStartsWith('string(', Utils::describeType(str_repeat('x', 100)));
    }

    public function testHeadersFromLines(): void
    {
        $h = Utils::headersFromLines(["A: 1\r\n", 'a: 2', 'B:  x ']);
        $this->assertSame(['A' => ['1'], 'a' => ['2'], 'B' => ['x']], $h);
    }

    public function testDebugResource(): void
    {
        $this->assertTrue(is_resource(Utils::debugResource()));
        $r = fopen('php://memory', 'r+');
        $this->assertSame($r, Utils::debugResource($r));
        fclose($r);
    }

    public function testUserAgentAndNormalize(): void
    {
        $this->assertStringStartsWith('GuzzleHttp/', Utils::defaultUserAgent());
        $this->assertSame(['host' => 'host'], Utils::normalizeHeaderKeys(['host' => 'a']));
        $this->assertSame(['x-a' => 'x-a'], Utils::normalizeHeaderKeys(['x-a' => 1]));
        $this->assertSame(['http', 'https'], Utils::normalizeProtocols(['http', 'https', 'http']));
    }

    public function testNormalizeProtocolsErrors(): void
    {
        foreach ([[], 'http', [1], ['ftp']] as $bad) {
            try {
                Utils::normalizeProtocols($bad);
                $this->fail('expected');
            } catch (InvalidArgumentException $e) {
                $this->assertStringContainsString('protocols', $e->getMessage());
            }
        }
    }

    public function testNoProxyHosts(): void
    {
        $this->assertTrue(Utils::isHostInNoProxy('example.com', ['*']));
        $this->assertTrue(Utils::isHostInNoProxy('example.com', ['example.com']));
        $this->assertTrue(Utils::isHostInNoProxy('foo.example.com', ['example.com']));
        $this->assertTrue(Utils::isHostInNoProxy('foo.example.com', ['.example.com']));
        $this->assertFalse(Utils::isHostInNoProxy('example.com', ['.example.com']));
        $this->assertFalse(Utils::isHostInNoProxy('notexample.com', ['example.com']));
        $this->assertTrue(Utils::isHostInNoProxy('EXAMPLE.com', ['example.COM']));
        $this->assertTrue(Utils::isHostInNoProxy('example.com:8080', ['example.com']));
        $this->assertFalse(Utils::isHostInNoProxy('example.com', ['other.org', 5]));
    }

    public function testNoProxyIpHosts(): void
    {
        $this->assertTrue(Utils::isHostInNoProxy('192.168.1.10', ['192.168.0.0/16']));
        $this->assertFalse(Utils::isHostInNoProxy('10.1.1.1', ['192.168.0.0/16']));
        $this->assertTrue(Utils::isHostInNoProxy('10.1.1.1', ['10.1.1.1']));
        $this->assertTrue(Utils::isHostInNoProxy('[fd00::1]', ['fd00::/8']));
        $this->assertFalse(Utils::isHostInNoProxy('[fe00::1]', ['fd00::/8']));
        $this->assertTrue(Utils::isHostInNoProxy('[::1]', ['::1']));
        $this->assertFalse(Utils::isHostInNoProxy('10.1.1.1', ['10.0.0.0/33']));
        $this->assertFalse(Utils::isHostInNoProxy('10.1.1.1', ['10.0.0.0/abc']));
        $this->assertFalse(Utils::isHostInNoProxy('10.1.1.1', ['']));
        $this->assertTrue(Utils::isHostInNoProxy('10.200.1.1', ['10.0.0.0/8']));
        $this->assertTrue(Utils::isHostInNoProxy('10.200.1.1', ['10.0.0.0/0']));
    }

    public function testNoProxyEmptyHost(): void
    {
        $this->expectException(InvalidArgumentException::class);
        Utils::isHostInNoProxy('', ['a']);
    }

    public function testNoProxyUri(): void
    {
        $u = new Uri('http://example.com:8080/x');
        $this->assertTrue(Utils::isUriInNoProxy($u, 'example.com:8080'));
        $this->assertFalse(Utils::isUriInNoProxy($u, 'example.com:9090'));
        $this->assertTrue(Utils::isUriInNoProxy($u, 'other.org, example.com'));
        $this->assertTrue(Utils::isUriInNoProxy(new Uri('http://example.com/'), 'example.com:80'));
        $this->assertTrue(Utils::isUriInNoProxy(new Uri('https://example.com/'), 'example.com:443'));
        $this->assertFalse(Utils::isUriInNoProxy(new Uri('https://example.com/'), 'example.com:80'));
        $this->assertTrue(Utils::isUriInNoProxy(new Uri('http://[::1]:8080/'), ['[::1]:8080']));
        $this->assertFalse(Utils::isUriInNoProxy(new Uri('http://[::1]:8080/'), ['[::1]:81']));
        $this->assertFalse(Utils::isUriInNoProxy($u, 5));
        $this->assertFalse(Utils::isUriInNoProxy(new Uri('/relative'), 'a'));
        $this->assertTrue(Utils::isUriInNoProxy($u, ['*']));
        $this->assertTrue(Utils::isUriInNoProxy(new Uri('http://10.0.0.5/'), '10.0.0.0/8'));
        $this->assertFalse(Utils::isUriInNoProxy(new Uri('ftp://example.com/'), 'example.com:21x'));
        $this->assertTrue(Utils::isUriInNoProxy(new Uri('http://example.com./'), 'example.com'));
    }

    public function testJson(): void
    {
        $this->assertSame(['a' => 1], Utils::jsonDecode('{"a":1}', true));
        $this->assertSame(1, Utils::jsonDecode('{"a":1}')->a);
        try {
            Utils::jsonDecode('{bad');
            $this->fail('expected');
        } catch (InvalidArgumentException $e) {
            $this->assertStringContainsString('json_decode error', $e->getMessage());
        }
        try {
            Utils::jsonEncode("\xB1\x31");
            $this->fail('expected');
        } catch (InvalidArgumentException $e) {
            $this->assertStringContainsString('json_encode error', $e->getMessage());
        }
        $this->assertSame(['x' => 1], \GuzzleHttp\json_decode('{"x":1}', true));
        $this->assertSame('[1]', \GuzzleHttp\json_encode([1]));
    }

    public function testCurrentTimeAndGetenv(): void
    {
        $this->assertGreaterThan(0, Utils::currentTime());
        putenv('GUZZLECOV_X=hello');
        $this->assertSame('hello', Utils::getenv('GUZZLECOV_X'));
        putenv('GUZZLECOV_X');
        $this->assertNull(Utils::getenv('GUZZLECOV_X'));
    }

    public function testIdnOption(): void
    {
        $this->assertNull(Utils::normalizeIdnConversionOption(false));
        $this->assertIsInt(Utils::normalizeIdnConversionOption(true));
        $this->assertSame(5, Utils::normalizeIdnConversionOption(5));
        try {
            Utils::normalizeIdnConversionOption('x');
            $this->fail('expected');
        } catch (InvalidArgumentException $e) {
            $this->assertStringContainsString('idn_conversion', $e->getMessage());
        }
        $u = new Uri('http://example.com/');
        $this->assertSame('example.com', Utils::idnUriConvert($u, 0)->getHost());
        $this->assertSame('example.com', Utils::idnUriConvert($u)->getHost());
        $this->assertSame('', Utils::idnUriConvert(new Uri('/p'))->getHost());
    }

    public function testChooseHandler(): void
    {
        $this->assertIsCallable(Utils::chooseHandler());
    }

    public function testChooseHandlerFunction(): void
    {
        $this->assertIsCallable(\GuzzleHttp\choose_handler());
    }

    public function testChooseHandlerFunctionsLegacy(): void
    {
        $this->assertSame(Utils::defaultUserAgent(), \GuzzleHttp\default_user_agent());
        $this->assertSame('int(1)', \GuzzleHttp\describe_type(1));
        $this->assertSame(['A' => ['1']], \GuzzleHttp\headers_from_lines(['A: 1']));
        $this->assertIsResource(\GuzzleHttp\debug_resource());
        $this->assertSame(['host' => 'host'], \GuzzleHttp\normalize_header_keys(['host' => 'a']));
        $this->assertTrue(\GuzzleHttp\is_host_in_noproxy('a.com', ['a.com']));
    }

    public function testChooseHandlerBadOptions(): void
    {
        foreach ([['max_host_connections' => 0], ['max_total_connections' => 'x']] as $o) {
            try {
                Utils::chooseHandler($o);
                $this->fail('expected');
            } catch (InvalidArgumentException $e) {
                $this->assertStringContainsString('positive integer', $e->getMessage());
            }
        }
    }

    public function testHostValidatorAccepts(): void
    {
        HostValidator::assertRequestHost(new Request('GET', 'http://example.com/'));
        HostValidator::assertRequestHost(new Request('GET', 'http://example.com:80/', ['Host' => 'example.com:80']));
        $this->assertTrue(true);
    }

    public function testHostValidatorAcceptsIpv6(): void
    {
        HostValidator::assertRequestHost(new Request('GET', 'http://[::1]:80/', ['Host' => 'example.com:80']));
        $this->assertTrue(true);
    }

    public function testHostValidatorRejects(): void
    {
        $bad = [
            new Request('GET', 'http://example.com/', ['Host' => "ex ample"]),
            new Request('GET', 'http://example.com/', ['Host' => "ex%41mple"]),
            new Request('GET', 'http://127.1./'),
            new Request('GET', 'http://0x7f.1../'),
        ];
        foreach ($bad as $req) {
            try {
                HostValidator::assertRequestHost($req);
                $this->fail('expected ' . $req->getUri());
            } catch (RequestException $e) {
                $this->assertStringContainsString('must', $e->getMessage());
            }
        }
    }

    public function testHostValidatorThirdPartyUri(): void
    {
        $mk = function (string $host): Request {
            $uri = new class($host) extends Uri {
                private $h;
                public function __construct($h) { $this->h = $h; }
                public function getHost(): string { return $this->h; }
            };
            return new Request('GET', 'http://x/') instanceof Request ? (new Request('GET', 'http://x/'))->withUri($uri, true) : null;
        };
        foreach (["a b", "a%41", "a/b", "a?b", "a@b", "a:b", "[a", "a]", "x[1]", "caf\xC3\xA9", "1.2.3.4.", "0x1.", "012."] as $h) {
            try {
                HostValidator::assertRequestHost($mk($h));
                $this->fail('expected ' . $h);
            } catch (RequestException $e) {
                $this->assertStringContainsString('request URI host', $e->getMessage());
            }
        }
        HostValidator::assertRequestHost($mk('[::1]'));
        HostValidator::assertRequestHost($mk('1.2.3.4'));
        HostValidator::assertRequestHost($mk('a.b.'));
        HostValidator::assertRequestHost($mk('1.2.3.4.5.'));
        HostValidator::assertRequestHost($mk('1..2.'));
        HostValidator::assertRequestHost($mk('09.'));
        HostValidator::assertRequestHost($mk('0x.'));
        $this->assertTrue(HostValidator::isNumericIpv4Host('1.2.3'));
        $this->assertTrue(HostValidator::isNumericIpv4Host('0x1F'));
        $this->assertTrue(HostValidator::isNumericIpv4Host('017'));
        $this->assertFalse(HostValidator::isNumericIpv4Host(''));
        $this->assertFalse(HostValidator::isNumericIpv4Host('1.2.3.4.5'));
        $this->assertFalse(HostValidator::isNumericIpv4Host('abc'));
        $this->assertFalse(HostValidator::isNumericIpv4Host('08'));
    }

    public function testHeaderProcessor(): void
    {
        [$v, $s, $r, $h] = HeaderProcessor::parseHeaders(['HTTP/1.1 200 OK', 'X-A: 1', 'x-a: 2']);
        $this->assertSame(['1.1', 200, 'OK', ['X-A' => ['1'], 'x-a' => ['2']]], [$v, $s, $r, $h]);
        [$v, $s, $r, $h] = HeaderProcessor::parseHeaders(['HTTP/1.1 100 Continue', '', 'HTTP/2 204']);
        $this->assertSame(['2', 204, null, []], [$v, $s, $r, $h]);
        foreach ([[], ['NOPE 200'], ['HTTP/1.1'], ['HTTP/1.1 20x'], ['HTTP/1.1 200 OK', 'bad line']] as $bad) {
            try {
                HeaderProcessor::parseHeaders($bad);
                $this->fail('expected');
            } catch (\RuntimeException $e) {
                $this->assertNotSame('', $e->getMessage());
            }
        }
        $this->assertTrue(HeaderProcessor::isStatusLineCandidate("HTTP/1.1 200 OK\r\n"));
        $this->assertFalse(HeaderProcessor::isStatusLineCandidate('X: y'));
        $this->assertTrue(HeaderProcessor::isValidHeaderFieldLine("X-A: b c\r\n"));
        $this->assertFalse(HeaderProcessor::isValidHeaderFieldLine('nocolon'));
        $this->assertFalse(HeaderProcessor::isValidHeaderFieldLine('bad name: v'));
        $this->assertFalse(HeaderProcessor::isValidHeaderFieldLine("X: a\x01b"));
    }

    public function testTlsVersion(): void
    {
        $this->assertSame(10, TlsVersion::ordinal('o', STREAM_CRYPTO_METHOD_TLSv1_0_CLIENT));
        $this->assertSame(11, TlsVersion::ordinal('o', STREAM_CRYPTO_METHOD_TLSv1_1_CLIENT));
        $this->assertSame(12, TlsVersion::ordinal('o', STREAM_CRYPTO_METHOD_TLSv1_2_CLIENT));
        $this->assertSame(13, TlsVersion::ordinal('o', STREAM_CRYPTO_METHOD_TLSv1_3_CLIENT));
        TlsVersion::assertRange(null, STREAM_CRYPTO_METHOD_TLSv1_2_CLIENT);
        TlsVersion::assertRange(STREAM_CRYPTO_METHOD_TLSv1_2_CLIENT, null);
        TlsVersion::assertRange(STREAM_CRYPTO_METHOD_TLSv1_2_CLIENT, STREAM_CRYPTO_METHOD_TLSv1_3_CLIENT);
        $this->assertIsInt(TlsVersion::streamProtocolVersion('o', STREAM_CRYPTO_METHOD_TLSv1_0_CLIENT));
        $this->assertIsInt(TlsVersion::streamProtocolVersion('o', STREAM_CRYPTO_METHOD_TLSv1_1_CLIENT));
        $this->assertIsInt(TlsVersion::streamProtocolVersion('o', STREAM_CRYPTO_METHOD_TLSv1_2_CLIENT));
        $this->assertIsInt(TlsVersion::streamProtocolVersion('o', STREAM_CRYPTO_METHOD_TLSv1_3_CLIENT));
    }

    public function testTlsVersionErrors(): void
    {
        try { TlsVersion::ordinal('o', 999999); $this->fail('x'); } catch (\InvalidArgumentException $e) { $this->assertStringContainsString('unknown version', $e->getMessage()); }
        try { TlsVersion::streamProtocolVersion('o', 999999); $this->fail('x'); } catch (\InvalidArgumentException $e) { $this->assertStringContainsString('unknown version', $e->getMessage()); }
        try { TlsVersion::assertRange(STREAM_CRYPTO_METHOD_TLSv1_3_CLIENT, STREAM_CRYPTO_METHOD_TLSv1_2_CLIENT); $this->fail('x'); } catch (\InvalidArgumentException $e) { $this->assertStringContainsString('maximum TLS', $e->getMessage()); }
    }

    public function testProxyWrappers(): void
    {
        $a = fn ($r, $o) => new FulfilledPromise('default');
        $b = fn ($r, $o) => new FulfilledPromise('other');
        $req = new Request('GET', 'http://x/');
        $this->assertSame('default', Proxy::wrapSync($a, $b)($req, [])->wait());
        $this->assertSame('other', Proxy::wrapSync($a, $b)($req, ['synchronous' => true])->wait());
        $this->assertSame('default', Proxy::wrapStreaming($a, $b)($req, [])->wait());
        $this->assertSame('other', Proxy::wrapStreaming($a, $b)($req, ['stream' => true])->wait());
        $this->assertSame('default', Proxy::wrapTlsFallback($a, $b)($req, [])->wait());
        $r = Proxy::wrapTlsFallback($a, $b)($req, ['crypto_method' => STREAM_CRYPTO_METHOD_TLSv1_2_CLIENT])->wait();
        $this->assertSame(CurlVersion::supportsTls12() ? 'default' : 'other', $r);
    }

    public function testProxyEnvironment(): void
    {
        foreach (['http_proxy', 'HTTP_PROXY', 'https_proxy', 'HTTPS_PROXY', 'all_proxy', 'ALL_PROXY', 'no_proxy', 'NO_PROXY'] as $n) {
            putenv($n);
        }
        $this->assertNull(ProxyEnvironment::getProxyForScheme('http'));
        $this->assertNull(ProxyEnvironment::getNoProxy());
        putenv('all_proxy=http://all:1');
        $this->assertSame('http://all:1', ProxyEnvironment::getProxyForScheme('http'));
        putenv('HTTPS_PROXY=http://s:2');
        $this->assertSame('http://s:2', ProxyEnvironment::getProxyForScheme('HTTPS'));
        putenv('https_proxy=http://sl:3');
        $this->assertSame('http://sl:3', ProxyEnvironment::getProxyForScheme('https'));
        putenv('NO_PROXY=a.com');
        $this->assertSame('a.com', ProxyEnvironment::getNoProxy());
        putenv('no_proxy=b.com');
        $this->assertSame('b.com', ProxyEnvironment::getNoProxy());
        foreach (['https_proxy', 'HTTPS_PROXY', 'all_proxy', 'no_proxy', 'NO_PROXY'] as $n) {
            putenv($n);
        }
    }

    public function testSplitNoProxy(): void
    {
        $this->assertSame(['a.com', 'b.org', 'c'], ProxyEnvironment::splitNoProxy('.a.com, b.org  c,,.'));
    }

    public function testCurlVersionHelpers(): void
    {
        $this->assertIsBool(CurlVersion::supportsCurlHandler());
        foreach (['supportsTls12', 'supportsTls13', 'supportsHttp2', 'supportsMultiplex', 'supportsHttpVersionReuseMatching', 'supportsConnectionCaps', 'supportsRequiredMultiplex', 'supportsHttpsProxy', 'supportsNtlm', 'supportsHandlerSharing', 'supportsSslSessionSharing', 'supportsShareConnectionCaches', 'supportsProxyTlsCredentialAwareConnectionReuse', 'supportsProxyCredentialAwareConnectionReuse', 'supportsSocksProxyCredentialAwareConnectionReuse', 'supportsProxyHeaderSeparation'] as $m) {
            $this->assertIsBool(CurlVersion::$m(), $m);
        }
        $v = CurlVersion::getVersion();
        $this->assertTrue($v === null || is_string($v));
    }

    public function testTransferStats(): void
    {
        $req = new Request('GET', 'http://x/');
        $s = new TransferStats($req, null, 0.5, 'err', ['total_time' => 1.5]);
        $this->assertSame($req, $s->getRequest());
        $this->assertNull($s->getResponse());
        $this->assertFalse($s->hasResponse());
        $this->assertSame('err', $s->getHandlerErrorData());
        $this->assertSame('http://x/', (string) $s->getEffectiveUri());
        $this->assertSame(0.5, $s->getTransferTime());
        $this->assertSame(['total_time' => 1.5], $s->getHandlerStats());
        $this->assertSame(1.5, $s->getHandlerStat('total_time'));
        $this->assertNull($s->getHandlerStat('nope'));
        $res = new Response(200);
        $s2 = new TransferStats($req, $res, null, null, []);
        $this->assertTrue($s2->hasResponse());
        $this->assertSame($res, $s2->getResponse());
        $this->assertNull($s2->getTransferTime());
    }

    public function testOnStatsCallback(): void
    {
        $got = null;
        $m = new MockHandler([new Response(200)], null, null);
        $c = new Client(['handler' => HandlerStack::create($m)]);
        $c->get('http://x/', ['on_stats' => function (TransferStats $s) use (&$got) { $got = $s; }]);
        $this->assertInstanceOf(TransferStats::class, $got);
        $this->assertSame(200, $got->getResponse()->getStatusCode());
    }
}
