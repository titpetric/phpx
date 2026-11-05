<?php

use GuzzleHttp\Exception\ConnectException;
use GuzzleHttp\Exception\RequestException;
use GuzzleHttp\Handler\CurlFactory;
use GuzzleHttp\Handler\CurlVersion;
use GuzzleHttp\Multiplexing;
use GuzzleHttp\Psr7\Request;
use GuzzleHttp\Psr7\Utils as PU7;
use PHPUnit\Framework\Attributes\DataProvider;
use PHPUnit\Framework\TestCase;

/**
 * CurlFactory::create() and its option mapping, without a transfer.
 */
final class GuzzleCurl2Test extends TestCase
{
    /** @var list<string> */
    private array $deprecations = [];
    /** @var list<string> */
    private array $tmp = [];

    protected function tearDown(): void
    {
        foreach ($this->tmp as $p) {
            if (is_link($p) || is_file($p)) {
                unlink($p);
            } elseif (is_dir($p)) {
                rmdir($p);
            }
        }
        $this->tmp = [];
    }

    private function tmpFile(string $suffix = '.pem', string $data = 'x'): string
    {
        $p = __DIR__.'/gcurl2-'.getmypid().'-'.count($this->tmp).$suffix;
        file_put_contents($p, $data);
        $this->tmp[] = $p;

        return $p;
    }

    /** @return array{0: ?\GuzzleHttp\Handler\EasyHandle, 1: ?\Throwable} */
    private function make(array $options = [], array $req = [], ?CurlFactory $f = null): array
    {
        $this->deprecations = [];
        set_error_handler(function ($no, $msg) {
            $this->deprecations[] = $msg;

            return true;
        }, E_USER_DEPRECATED);
        try {
            $request = new Request(
                $req['method'] ?? 'GET',
                $req['uri'] ?? 'http://example.test/path',
                $req['headers'] ?? [],
                $req['body'] ?? null,
                $req['version'] ?? '1.1'
            );

            return [($f ?? new CurlFactory(3))->create($request, $options), null];
        } catch (\Throwable $e) {
            return [null, $e];
        } finally {
            restore_error_handler();
        }
    }

    private function ok(array $options = [], array $req = [], ?CurlFactory $f = null): \GuzzleHttp\Handler\EasyHandle
    {
        [$easy, $e] = $this->make($options, $req, $f);
        $this->assertNull($e, $e ? get_class($e).': '.$e->getMessage() : '');

        return $easy;
    }

    private function fails(string $class, string $needle, array $options = [], array $req = [], ?CurlFactory $f = null): \Throwable
    {
        [$easy, $e] = $this->make($options, $req, $f);
        $this->assertNotNull($e, 'expected '.$class.' with '.$needle);
        $this->assertInstanceOf($class, $e, $e->getMessage());
        $this->assertStringContainsString($needle, $e->getMessage());

        return $e;
    }

    public function testSchemeAndHost(): void
    {
        $this->fails(RequestException::class, 'is not allowed by the protocols', ['protocols' => ['https']]);
        $this->fails(RequestException::class, 'URI must include a scheme', [], ['uri' => '/relative']);
        $this->fails(RequestException::class, 'is not supported', [], ['uri' => 'ftp://example.test/x']);
        $this->ok(['protocols' => ['http', 'https']]);
        $this->ok([], ['uri' => 'https://example.test/']);
    }

    public function testOnTrailersMustBeCallable(): void
    {
        $this->fails(\InvalidArgumentException::class, 'on_trailers must be callable', ['on_trailers' => 'nope-not-callable']);
        $this->fails(\InvalidArgumentException::class, 'on_headers must be callable', ['on_headers' => 'nope-not-callable']);
        $this->ok(['on_trailers' => fn () => null, 'on_headers' => fn () => null]);
    }

    public function testProtocolVersions(): void
    {
        $this->ok([], ['version' => '1.0']);
        $this->ok([], ['version' => '2']);
        $this->ok([], ['version' => '2.0']);
        $this->fails(ConnectException::class, 'HTTP/3 is not supported', [], ['version' => '3']);
        $this->ok([], ['version' => '']);
        $this->assertNotEmpty($this->deprecations);
        $this->assertStringContainsString('empty protocol version', $this->deprecations[0]);
    }

    public function testMultiplexOption(): void
    {
        $this->fails(\InvalidArgumentException::class, 'must be null or a GuzzleHttp\\Multiplexing::* constant; received string', ['multiplex' => 'bogus']);
        $this->fails(\InvalidArgumentException::class, 'received int', ['multiplex' => 5]);
        foreach ([Multiplexing::NONE, Multiplexing::EAGER, Multiplexing::WAIT] as $m) {
            $this->ok(['multiplex' => $m], ['version' => '2']);
            $this->ok(['multiplex' => $m], ['version' => '1.1']);
        }
        foreach ([Multiplexing::REQUIRE_EAGER, Multiplexing::REQUIRE_WAIT] as $m) {
            $this->fails(ConnectException::class, 'cannot be required for HTTP/1.1', ['multiplex' => $m], ['version' => '1.1']);
            $this->fails(ConnectException::class, 'cannot be required for HTTP/1.0', ['multiplex' => $m], ['version' => '1.0']);
            $this->fails(\InvalidArgumentException::class, 'raw CURLOPT_HTTP_VERSION cURL option', ['multiplex' => $m, 'curl' => [CURLOPT_HTTP_VERSION => 2]], ['version' => '2']);
            $this->fails(\InvalidArgumentException::class, 'raw CURLOPT_URL cURL option', ['multiplex' => $m, 'curl' => [CURLOPT_URL => 'http://x/']], ['version' => '2']);
            $this->fails(\InvalidArgumentException::class, 'raw CURLOPT_FOLLOWLOCATION cURL option', ['multiplex' => $m, 'curl' => [CURLOPT_FOLLOWLOCATION => true]], ['version' => '2']);
            if (!CurlVersion::supportsRequiredMultiplex()) {
                $this->fails(ConnectException::class, 'Required multiplexing needs libcurl 8.14.0', ['multiplex' => $m], ['version' => '2']);

                continue;
            }
            $this->ok(['multiplex' => $m], ['version' => '2']);
            $this->ok(['multiplex' => $m], ['version' => '2', 'uri' => 'https://example.test/']);
            $this->ok(['multiplex' => $m, 'proxy' => 'http://127.0.0.1:9', 'curl' => [CURLOPT_NOPROXY => '*']], ['version' => '2']);
            $this->fails(ConnectException::class, 'through a proxy', ['multiplex' => $m, 'proxy' => 'http://127.0.0.1:9'], ['version' => '2']);
            $this->fails(\InvalidArgumentException::class, 'final CURLOPT_PROXY cURL option value is not a string', ['multiplex' => $m, 'curl' => [CURLOPT_PROXY => 5]], ['version' => '2']);
            $this->fails(\InvalidArgumentException::class, 'final CURLOPT_NOPROXY cURL option value is not a string', ['multiplex' => $m, 'curl' => [CURLOPT_NOPROXY => 5]], ['version' => '2']);
            $this->fails(ConnectException::class, 'through a proxy', ['multiplex' => $m, 'curl' => [CURLOPT_PRE_PROXY => 'socks5://127.0.0.1:9']], ['version' => '2']);
            $this->fails(\InvalidArgumentException::class, 'permits NTLM', ['multiplex' => $m, 'curl' => [CURLOPT_HTTPAUTH => CURLAUTH_NTLM]], ['version' => '2']);
            $this->fails(\InvalidArgumentException::class, 'permits NTLM', ['multiplex' => $m, 'curl' => [CURLOPT_HTTPAUTH => CURLAUTH_ANY]], ['version' => '2']);
            $this->fails(\InvalidArgumentException::class, 'not an integer', ['multiplex' => $m, 'curl' => [CURLOPT_HTTPAUTH => [1]]], ['version' => '2']);
            $this->ok(['multiplex' => $m, 'curl' => [CURLOPT_HTTPAUTH => CURLAUTH_BASIC]], ['version' => '2']);
        }
        $this->ok(['multiplex' => Multiplexing::WAIT], ['version' => '2']);
    }

    public function testInvalidCurlOptions(): void
    {
        $this->fails(\InvalidArgumentException::class, 'Invalid cURL option "foo"', ['curl' => ['foo' => 1]]);
        $e = $this->fails(\InvalidArgumentException::class, 'Unable to set cURL option', ['curl' => [999999 => 1]]);
        $this->assertNotNull($e);
        // a raw option outside the allow-list and a conflicting one only deprecate
        $this->ok(['curl' => [CURLOPT_BUFFERSIZE => 4096]]);
        $this->assertStringContainsString("allow-list", implode("\n", $this->deprecations));
        $this->ok(['curl' => [CURLOPT_TIMEOUT => 5]]);
        $this->assertStringContainsString('"timeout" request option', implode("\n", $this->deprecations));
        $this->ok(['stream_context' => ['http' => ['method' => 'GET']]]);
        $this->assertStringContainsString('stream_context', implode("\n", $this->deprecations));
    }

    public static function conflictingOptions(): array
    {
        $names = [
            'CURLOPT_PORT' => 80, 'CURLOPT_HTTPGET' => true, 'CURLOPT_NOBODY' => false, 'CURLOPT_NOSIGNAL' => true,
            'CURLOPT_VERBOSE' => false, 'CURLOPT_MAXREDIRS' => 3, 'CURLOPT_POSTREDIR' => 1, 'CURLOPT_PROTOCOLS' => 3,
            'CURLOPT_PIPEWAIT' => true, 'CURLOPT_IPRESOLVE' => 1, 'CURLOPT_SSL_VERIFYPEER' => true, 'CURLOPT_SSL_VERIFYHOST' => 2,
            'CURLOPT_SSLVERSION' => 0, 'CURLOPT_COOKIE' => 'a=b', 'CURLOPT_COOKIESESSION' => true, 'CURLOPT_PROXYTYPE' => 0,
            'CURLOPT_USERAGENT' => 'ua', 'CURLOPT_REFERER' => 'http://r/', 'CURLOPT_CONNECTTIMEOUT' => 3, 'CURLOPT_TIMEOUT_MS' => 3000,
            'CURLOPT_CONNECTTIMEOUT_MS' => 3000, 'CURLOPT_PUT' => false, 'CURLOPT_UPLOAD' => false, 'CURLOPT_POST' => false,
            'CURLOPT_POSTFIELDS' => 'a', 'CURLOPT_CUSTOMREQUEST' => 'GET', 'CURLOPT_INFILESIZE' => 1, 'CURLOPT_CAPATH' => '/',
            'CURLOPT_SSLCERTTYPE' => 'PEM', 'CURLOPT_SSLKEYTYPE' => 'PEM', 'CURLOPT_COOKIELIST' => 'ALL',
        ];
        $rows = [];
        foreach ($names as $n => $v) {
            if (defined($n)) {
                $rows[$n] = [$n, constant($n), $v];
            }
        }

        return $rows;
    }

    #[DataProvider('conflictingOptions')]
    public function testConflictingOptionDeprecated(string $name, int $opt, $value): void
    {
        $this->make(['curl' => [$opt => $value]]);
        $this->assertNotEmpty($this->deprecations, $name);
        $this->assertStringContainsString('request option is deprecated; guzzlehttp/guzzle 8.0 will reject', implode("\n", $this->deprecations));
    }

    public function testConflictingOptionNamedInMessage(): void
    {
        foreach (self::conflictingOptions() as [$name, $opt, $value]) {
            $this->make(['curl' => [$opt => $value]]);
            $this->assertStringContainsString('Passing '.$name.' ('.$opt.') in the "curl" request option', implode("\n", $this->deprecations), $name);
        }
    }

    public function testVerifyOption(): void
    {
        $this->ok(['verify' => false]);
        $this->ok(['verify' => true]);
        $f = $this->tmpFile('.crt');
        $this->ok(['verify' => $f]);
        $this->ok(['verify' => __DIR__]);
        $this->fails(\InvalidArgumentException::class, 'SSL CA bundle not found', ['verify' => '/nonexistent/ca.pem']);
        $link = __DIR__.'/gcurl2-'.getmypid().'-link';
        if (@symlink('.', $link)) {
            $this->tmp[] = $link;
            $this->ok(['verify' => $link]);
        }
    }

    public function testCertOption(): void
    {
        $pem = $this->tmpFile('.pem');
        $der = $this->tmpFile('.der');
        $p12 = $this->tmpFile('.p12');
        $this->ok(['cert' => $pem]);
        $this->ok(['cert' => [$pem, 'secret']]);
        $this->ok(['cert' => $der]);
        $this->ok(['cert' => $p12]);
        $this->ok(['cert' => $der, 'cert_type' => 'pem']);
        $this->fails(\InvalidArgumentException::class, 'Invalid cert request option', ['cert' => 5]);
        $this->fails(\InvalidArgumentException::class, 'Invalid cert request option', ['cert' => []]);
        $this->fails(\InvalidArgumentException::class, 'Invalid cert request option', ['cert' => [$pem, 5]]);
        $this->fails(\InvalidArgumentException::class, 'SSL certificate not found', ['cert' => '/nonexistent/cert.pem']);
        $this->fails(\InvalidArgumentException::class, 'cert_type must be a non-empty string', ['cert_type' => '']);
        $this->fails(\InvalidArgumentException::class, 'cert_type must be a non-empty string', ['cert_type' => 5]);
    }

    public function testSslKeyOption(): void
    {
        $key = $this->tmpFile('.key');
        $this->ok(['ssl_key' => $key]);
        $this->ok(['ssl_key' => [$key, 'pw']]);
        $this->ok(['ssl_key' => '/nonexistent/engine-key', 'ssl_key_type' => 'eng']);
        $this->ok(['ssl_key' => 'pkcs11:id=1', 'ssl_key_type' => 'PROV']);
        $this->fails(\InvalidArgumentException::class, 'SSL private key not found', ['ssl_key' => '/nonexistent/key.pem']);
        $this->fails(\InvalidArgumentException::class, 'Invalid ssl_key request option', ['ssl_key' => 5]);
        $this->fails(\InvalidArgumentException::class, 'Invalid ssl_key request option', ['ssl_key' => []]);
        $this->fails(\InvalidArgumentException::class, 'Invalid ssl_key request option', ['ssl_key' => [$key, 5]]);
        $this->fails(\InvalidArgumentException::class, 'ssl_key_type must be a non-empty string', ['ssl_key_type' => '']);
    }

    public function testCryptoMethods(): void
    {
        $t10 = STREAM_CRYPTO_METHOD_TLSv1_0_CLIENT;
        $t11 = STREAM_CRYPTO_METHOD_TLSv1_1_CLIENT;
        $t12 = STREAM_CRYPTO_METHOD_TLSv1_2_CLIENT;
        $t13 = STREAM_CRYPTO_METHOD_TLSv1_3_CLIENT;
        foreach ([$t10, $t11, $t12, $t13] as $min) {
            $this->ok(['crypto_method' => $min]);
            $this->ok(['crypto_method_max' => $min]);
        }
        $this->ok(['crypto_method' => $t11, 'crypto_method_max' => $t13]);
        $this->ok(['crypto_method' => $t12, 'crypto_method_max' => $t12]);
        $this->ok(['crypto_method' => $t10], ['version' => '2']);
        $this->ok(['crypto_method_max' => $t13], ['version' => '2']);
        $this->fails(\InvalidArgumentException::class, 'HTTP/2 requires TLS 1.2 or higher', ['crypto_method_max' => $t11], ['version' => '2']);
        $this->fails(\InvalidArgumentException::class, '', ['crypto_method' => $t13, 'crypto_method_max' => $t12]);
        $this->fails(\InvalidArgumentException::class, '', ['crypto_method' => 12345]);
        $this->fails(\InvalidArgumentException::class, '', ['crypto_method_max' => 12345]);
    }

    public function testProgressAndDebug(): void
    {
        $this->fails(\InvalidArgumentException::class, 'progress client option must be callable', ['progress' => 'nope-not-callable']);
        $this->ok(['progress' => fn () => null]);
        $this->ok(['debug' => true]);
        $this->ok(['debug' => fopen('php://memory', 'w+')]);
        $this->ok(['debug' => false]);
    }

    public function testBodiesAndMethods(): void
    {
        $e = $this->ok([], ['method' => 'HEAD', 'headers' => ['Content-Length' => '5', 'Transfer-Encoding' => 'chunked', 'Expect' => '100-continue']]);
        $this->assertSame('HEAD', $e->request->getMethod());
        $this->ok([], ['method' => 'HEAD', 'headers' => ['Content-Length' => '0']]);
        $this->ok([], ['method' => 'POST', 'body' => '']);
        $this->ok([], ['method' => 'PUT', 'body' => '']);
        $this->ok([], ['method' => 'PUT', 'body' => '', 'headers' => ['Content-Length' => '0']]);
        $this->ok([], ['method' => 'POST', 'body' => 'abc', 'headers' => ['Content-Length' => '3']]);
        $this->ok([], ['method' => 'POST', 'body' => 'abc']);
        $this->ok(['curl' => ['body_as_string' => true]], ['method' => 'POST', 'body' => 'abc']);
        $this->ok([], ['method' => 'POST', 'body' => 'abc', 'headers' => ['Content-Length' => '2000000', 'Expect' => '100-continue', 'Content-Type' => 'text/plain']]);
        $this->ok([], ['method' => 'POST', 'body' => PU7::streamFor(new \GuzzleHttp\Psr7\PumpStream(fn () => ''))]);
        $this->ok([], ['method' => 'POST', 'body' => 'abc', 'headers' => ['Accept' => 'text/plain', 'X-Empty' => '']]);
    }

    public function testReadFunctionStreamsBody(): void
    {
        $body = PU7::streamFor(str_repeat('z', 50));
        $e = $this->ok([], ['method' => 'POST', 'body' => $body, 'headers' => ['Content-Length' => '2000000']]);
        $this->assertSame(0, $body->tell());
        $this->assertNotNull($e->handle);
    }

    public function testSinkOption(): void
    {
        $this->fails(\RuntimeException::class, 'Directory /nonexistent-dir-gc does not exist for sink', ['sink' => '/nonexistent-dir-gc/out']);
        $e = $this->ok(['sink' => __DIR__.'/gcurl2-sink-'.getmypid()]);
        $this->assertInstanceOf(\GuzzleHttp\Psr7\LazyOpenStream::class, $e->sink);
        $e = $this->ok(['sink' => fopen('php://memory', 'w+')]);
        $this->assertNotNull($e->sink);
    }

    public function testDecodeContentAndTimeouts(): void
    {
        $this->ok(['decode_content' => true]);
        $this->ok(['decode_content' => 'gzip'], ['headers' => ['Accept-Encoding' => 'gzip']]);
        $this->ok(['decode_content' => false]);
        $this->ok(['timeout' => 0.5, 'connect_timeout' => 0.2]);
        $this->ok(['timeout' => 3, 'connect_timeout' => 2]);
        $this->ok(['force_ip_resolve' => 'v4']);
        $this->ok(['force_ip_resolve' => 'v6']);
        $this->ok(['force_ip_resolve' => 'other']);
    }

    public function testHandlePoolAndRelease(): void
    {
        $f = new CurlFactory(1);
        $a = $this->ok([], [], $f);
        $b = $this->ok([], [], $f);
        $f->release($a);
        $f->release($b); // pool is full: dropped
        $c = $this->ok([], [], $f);
        $this->assertNotNull($c->handle);
        $f->release($c);
        unset($f);
        $this->assertTrue(true);
    }
}
