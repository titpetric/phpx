<?php

use GuzzleHttp\Exception\ConnectException;
use GuzzleHttp\Exception\RequestException;
use GuzzleHttp\Handler\CurlFactory;
use GuzzleHttp\Handler\CurlShareHandleState;
use GuzzleHttp\TransportSharing;

require_once __DIR__.'/GuzzleCurlFactoryBase.php';

/**
 * Proxy, share-handle and header-list handling of CurlFactory::create().
 */
final class GuzzleCurl4Test extends GuzzleCurlFactoryBase
{
    private const HTTPS = ['uri' => 'https://example.test/secure'];

    public function testProxyStrings(): void
    {
        foreach ([
            'http://127.0.0.1:9', '127.0.0.1:9', 'http://user:pa%20ss@127.0.0.1:9', 'https://127.0.0.1:9',
            'socks5://127.0.0.1:9', 'socks5h://u:p@127.0.0.1:9', 'socks4://127.0.0.1:9', 'socks4a://127.0.0.1:9',
            'socks://127.0.0.1:9', 'ftp://127.0.0.1:9', 'HTTP://127.0.0.1:9', '',
        ] as $proxy) {
            $this->ok(['proxy' => $proxy]);
            $this->ok(['proxy' => $proxy], self::HTTPS);
        }
        $this->fails(RequestException::class, 'The proxy URL is malformed.', ['proxy' => '1bad_scheme!://127.0.0.1:9']);
    }

    public function testProxyArrays(): void
    {
        $proxy = ['http' => 'http://127.0.0.1:9', 'https' => 'http://127.0.0.1:10', 'no' => ['example.test']];
        $this->ok(['proxy' => $proxy]);
        $this->ok(['proxy' => $proxy], self::HTTPS);
        $this->ok(['proxy' => ['http' => 'http://127.0.0.1:9', 'no' => ['other.test']]]);
        $this->ok(['proxy' => ['https' => 'http://127.0.0.1:9']]);
        $this->ok(['proxy' => ['http' => 'http://127.0.0.1:9', 'no' => ['.test']]], ['uri' => 'http://sub.example.test/']);
        $this->ok(['proxy' => ['http' => 'http://127.0.0.1:9', 'no' => ['*']]]);
    }

    public function testProxyFromEnvironment(): void
    {
        $saved = [];
        foreach (['http_proxy', 'HTTP_PROXY', 'https_proxy', 'HTTPS_PROXY', 'no_proxy', 'NO_PROXY'] as $k) {
            $saved[$k] = getenv($k);
            putenv($k);
        }
        try {
            putenv('http_proxy=http://127.0.0.1:9');
            putenv('https_proxy=http://127.0.0.1:10');
            $this->ok();
            $this->ok([], self::HTTPS);
            putenv('no_proxy=example.test,.other');
            $this->ok();
            $this->ok([], self::HTTPS);
            $this->ok([], ['uri' => 'http://elsewhere.test/']);
        } finally {
            foreach ($saved as $k => $v) {
                $v === false ? putenv($k) : putenv("$k=$v");
            }
        }
    }

    public function testFinalProxyOptionTypes(): void
    {
        $this->fails(\InvalidArgumentException::class, 'CURLOPT_PROXYTYPE must be an integer.', ['curl' => [CURLOPT_PROXYTYPE => 'socks']]);
        $this->fails(\InvalidArgumentException::class, 'CURLOPT_PROXY must be a string.', ['curl' => [CURLOPT_PROXY => 5]]);
        $this->fails(\InvalidArgumentException::class, 'CURLOPT_NOPROXY must be a string.', ['curl' => [CURLOPT_NOPROXY => 5]]);
        $this->fails(\InvalidArgumentException::class, 'CURLOPT_PRE_PROXY must be a string.', ['curl' => [CURLOPT_PRE_PROXY => 5]]);
        $this->ok(['curl' => [CURLOPT_PROXY => 'http://127.0.0.1:9', CURLOPT_PROXYTYPE => CURLPROXY_SOCKS5]]);
        $this->ok(['curl' => [CURLOPT_PROXY => '127.0.0.1:9', CURLOPT_PROXYTYPE => CURLPROXY_HTTP]]);
        $this->ok(['curl' => [CURLOPT_PROXY => '127.0.0.1:9', CURLOPT_PROXYTYPE => CURLPROXY_SOCKS4A]]);
        $this->ok(['curl' => [CURLOPT_PRE_PROXY => 'socks5://127.0.0.1:9']]);
        $this->ok(['curl' => [CURLOPT_PRE_PROXY => '']]);
    }

    public function testStringableProxyCredentials(): void
    {
        $s = new class() {
            public function __toString(): string
            {
                return 'secret';
            }
        };
        $this->ok(['proxy' => 'http://127.0.0.1:9', 'curl' => [CURLOPT_PROXYUSERPWD => $s, CURLOPT_PROXYUSERNAME => $s, CURLOPT_PROXYPASSWORD => $s]]);
        $bad = new class() {
            public function __toString(): string
            {
                throw new \RuntimeException('no string');
            }
        };
        $this->fails(\InvalidArgumentException::class, 'Unable to set cURL option', ['proxy' => 'http://127.0.0.1:9', 'curl' => [CURLOPT_PROXYUSERPWD => $bad]]);
        $this->fails(\InvalidArgumentException::class, 'Unable to set cURL option', ['proxy' => 'http://127.0.0.1:9', 'curl' => [CURLOPT_PROXYUSERPWD => new \stdClass()]]);
    }

    public function testTunnelOwnershipAcrossRequests(): void
    {
        $f = new CurlFactory(5);
        $a = $this->ok(['proxy' => 'http://127.0.0.1:9'], self::HTTPS, $f);
        $this->assertNotNull($a->effectiveProxy);
        $f->release($a);
        $b = $this->ok(['proxy' => 'http://u:p@127.0.0.1:9'], self::HTTPS, $f);
        $c = $this->ok(['proxy' => 'http://u:p@127.0.0.1:9', 'curl' => [CURLOPT_PROXYUSERNAME => 'x']], self::HTTPS, $f);
        $f->release($b);
        $f->release($c);
        $d = $this->ok(['proxy' => 'socks5://127.0.0.1:9'], self::HTTPS, $f);
        $f->release($d);
        $e = $this->ok(['proxy' => 'http://127.0.0.1:9', 'curl' => [CURLOPT_PROXY_SSLCERT => 'x', CURLOPT_PROXY_TLSAUTH_USERNAME => 'x', CURLOPT_PROXY_SSLKEY => 'k', CURLOPT_PROXY_KEYPASSWD => 'p', CURLOPT_PROXY_SSLVERSION => 1]], self::HTTPS, $f);
        $f->release($e);
        $g = $this->ok(['proxy' => 'http://127.0.0.1:9', 'curl' => [CURLOPT_HTTPPROXYTUNNEL => true]], [], $f);
        $f->release($g);
        $h = $this->ok(['proxy' => 'http://127.0.0.1:9', 'curl' => [CURLOPT_CONNECT_TO => ['example.test:80:127.0.0.1:80']]], [], $f);
        $f->release($h);
        $i = $this->ok(['proxy' => 'http://127.0.0.1:9', 'curl' => [CURLOPT_CONNECT_TO => []]], [], $f);
        $f->release($i);
        $this->assertTrue(true);
    }

    public function testProxyAuthorizationHeaders(): void
    {
        $p = ['proxy' => 'http://127.0.0.1:9'];
        $this->ok($p, ['headers' => ['Proxy-Authorization' => 'Basic abc']]);
        $this->ok($p, ['headers' => ['Proxy-Authorization' => ['Basic abc', 'Bearer x']]]);
        $this->ok($p, ['headers' => ['Proxy-Authorization' => '']]);
        $this->ok($p + ['curl' => [CURLOPT_PROXYHEADER => ['X-P: 1']]], ['headers' => ['Proxy-Authorization' => 'Basic abc']]);
        $this->fails(\InvalidArgumentException::class, 'CURLOPT_PROXYHEADER must be an array', $p + ['curl' => [CURLOPT_PROXYHEADER => 'X-P: 1']], ['headers' => ['Proxy-Authorization' => 'Basic abc']]);
        $this->ok($p + ['curl' => [CURLOPT_HTTPHEADER => ['Proxy-Authorization: Basic xyz', 'X-A: b', 'No-Colon', 'proxy-authorization;']]]);
        $this->ok($p + ['curl' => [CURLOPT_HTTPHEADER => ['X-A: b']]], ['headers' => ['Proxy-Authorization' => 'Basic abc']]);
        $this->ok($p + ['curl' => [CURLOPT_PROXYHEADER => ['Proxy-Authorization: Basic zzz']]], self::HTTPS);
        $this->ok(['proxy' => 'socks5://127.0.0.1:9'], ['headers' => ['Proxy-Authorization' => 'Basic abc']]);
        $this->ok(['proxy' => 'socks5://127.0.0.1:9', 'curl' => [CURLOPT_HTTPHEADER => ['Proxy-Authorization: x']]]);
        $this->ok([], ['headers' => ['Proxy-Authorization' => 'Basic abc']]);
        $this->ok($p, self::HTTPS + ['headers' => ['X-A' => 'b']]);
    }

    public function testCurlHeaderListNormalization(): void
    {
        $s = new class() {
            public function __toString(): string
            {
                return 'X-S: stringable';
            }
        };
        $this->ok(['curl' => [CURLOPT_HTTPHEADER => [$s, 5, 1.5, true, INF, -INF, NAN, 'A: b']]]);
        $this->fails(\InvalidArgumentException::class, 'CURLOPT_HTTPHEADER entries must be strings', ['curl' => [CURLOPT_HTTPHEADER => [['x']]]]);
        $this->fails(\InvalidArgumentException::class, 'CURLOPT_HTTPHEADER entries must be strings', ['curl' => [CURLOPT_HTTPHEADER => [null]]]);
        $this->fails(\InvalidArgumentException::class, 'must not contain a carriage return or line feed', ['curl' => [CURLOPT_HTTPHEADER => ["A: b\r\nX: y"]]]);
        $this->fails(\InvalidArgumentException::class, 'CURLOPT_PROXYHEADER entries must not contain a carriage return', ['proxy' => 'http://127.0.0.1:9', 'curl' => [CURLOPT_PROXYHEADER => ["A: b\nX: y"]]]);
        $this->ok(['curl' => [CURLOPT_HTTPHEADER => []]]);
        $this->fails(\InvalidArgumentException::class, 'Unable to set cURL option', ['curl' => [CURLOPT_HTTPHEADER => 'not-an-array']]);
    }

    public function testOptionNamesInErrors(): void
    {
        $this->fails(\InvalidArgumentException::class, 'Unable to set cURL option CURLOPT_HTTPHEADER (10023)', ['curl' => [CURLOPT_HTTPHEADER => 'not-an-array']]);
        $this->fails(\InvalidArgumentException::class, 'Unable to set cURL option CURLOPT_PROXYUSERPWD (10006)', ['proxy' => 'http://127.0.0.1:9', 'curl' => [CURLOPT_PROXYUSERPWD => new \stdClass()]]);
    }

    public function testConstructorValidation(): void
    {
        $f = fn (...$a) => new CurlFactory(...$a);
        $this->assertInstanceOf(CurlFactory::class, $f(1));
        $this->expectException(\InvalidArgumentException::class);
        $this->expectExceptionMessage('A cURL share handle is required');
        $f(1, TransportSharing::HANDLER_PREFER);
    }

    public function testConstructorValidationNone(): void
    {
        $this->expectException(\InvalidArgumentException::class);
        $this->expectExceptionMessage('cannot be provided when transport sharing is disabled');
        new CurlFactory(1, TransportSharing::NONE, curl_share_init());
    }

    public function testConstructorValidationType(): void
    {
        $this->expectException(\InvalidArgumentException::class);
        $this->expectExceptionMessage('must be an instance of CurlShareHandle');
        new CurlFactory(1, TransportSharing::HANDLER_PREFER, 'not a share handle');
    }

    public function testConstructorValidationStateMode(): void
    {
        $state = CurlShareHandleState::fromOption(TransportSharing::HANDLER_PREFER);
        $this->assertNotNull($state);
        $this->expectException(\InvalidArgumentException::class);
        $this->expectExceptionMessage('does not match the configured transport sharing mode');
        new CurlFactory(1, TransportSharing::HANDLER_REQUIRE, $state);
    }

    public function testShareHandleFactories(): void
    {
        $state = CurlShareHandleState::fromOption(TransportSharing::HANDLER_PREFER);
        $managed = new CurlFactory(3, TransportSharing::HANDLER_PREFER, $state);
        $this->ok([], [], $managed);
        $this->ok(['proxy' => 'http://u:p@127.0.0.1:9'], self::HTTPS, $managed);
        $this->ok(['proxy' => 'socks5://u:p@127.0.0.1:9'], [], $managed);
        $this->fails(\InvalidArgumentException::class, 'request-level CURLOPT_SHARE cURL option cannot be combined with configured transport sharing', ['curl' => [CURLOPT_SHARE => curl_share_init()]], [], $managed);

        $external = new CurlFactory(3, TransportSharing::HANDLER_PREFER, curl_share_init());
        $this->ok([], [], $external);
        $this->ok(['proxy' => 'http://127.0.0.1:9'], self::HTTPS, $external);
        $this->ok(['proxy' => 'http://u:p@127.0.0.1:9'], self::HTTPS, $external);
        $this->ok(['proxy' => 'http://127.0.0.1:9', 'curl' => [CURLOPT_PROXYHEADER => ['Proxy-Authorization: x']]], self::HTTPS, $external);
        $this->ok(['proxy' => 'socks5://127.0.0.1:9'], self::HTTPS, $external);
    }

    public function testRequestLevelShareWithProxies(): void
    {
        $share = ['curl' => [CURLOPT_SHARE => curl_share_init()]];
        $this->ok($share);
        $this->ok(['proxy' => 'http://127.0.0.1:9'] + $share);
        $this->ok(['proxy' => 'http://127.0.0.1:9'] + $share, self::HTTPS === [] ? [] : []);
        $this->fails(\InvalidArgumentException::class, 'authenticated SOCKS proxy', ['proxy' => 'socks5://u:p@127.0.0.1:9'] + $share);
        $this->fails(\InvalidArgumentException::class, 'authenticated SOCKS proxy', ['proxy' => 'socks5://127.0.0.1:9', 'curl' => [CURLOPT_SHARE => curl_share_init(), CURLOPT_PROXYUSERPWD => 'u:p']]);
        $this->ok(['proxy' => 'socks5://127.0.0.1:9'] + $share);
        $this->fails(\InvalidArgumentException::class, 'authenticated HTTP/HTTPS proxy tunnel', ['proxy' => 'http://u:p@127.0.0.1:9'] + $share, self::HTTPS);
        $this->fails(\InvalidArgumentException::class, 'HTTP/HTTPS proxy tunnel configuration on libcurl', ['proxy' => 'http://127.0.0.1:9'] + $share, self::HTTPS);
        $this->fails(\InvalidArgumentException::class, 'authenticated HTTP/HTTPS proxy tunnel', ['proxy' => 'http://127.0.0.1:9', 'curl' => [CURLOPT_SHARE => curl_share_init(), CURLOPT_PROXY_SSLCERT => 'x']], self::HTTPS);
        $this->fails(\InvalidArgumentException::class, 'authenticated HTTP/HTTPS proxy tunnel', ['proxy' => 'http://127.0.0.1:9', 'curl' => [CURLOPT_SHARE => curl_share_init(), CURLOPT_HTTPHEADER => ['Proxy-Authorization: x']]], self::HTTPS);
        $this->ok(['proxy' => 'http://127.0.0.1:9'] + $share);
    }

    public function testRequestSchemeProxyEdge(): void
    {
        $this->fails(ConnectException::class, 'HTTP/3 is not supported', ['proxy' => 'http://127.0.0.1:9'], ['version' => '3']);
        $this->ok(['proxy' => 'http://127.0.0.1:9', 'curl' => [CURLOPT_NOPROXY => '*']]);
        $this->ok(['proxy' => 'http://127.0.0.1:9', 'curl' => [CURLOPT_NOPROXY => 'example.test']]);
    }
}
