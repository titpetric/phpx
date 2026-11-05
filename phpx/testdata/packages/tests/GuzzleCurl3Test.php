<?php

use GuzzleHttp\Handler\CurlFactory;
use GuzzleHttp\Handler\CurlHandler;
use GuzzleHttp\Handler\CurlMultiHandler;
use GuzzleHttp\Multiplexing;
use GuzzleHttp\Psr7\Request;
use GuzzleHttp\TransportSharing;
use PHPUnit\Framework\TestCase;

/**
 * CurlMultiHandler / CurlHandler construction and request-level validation.
 * Nothing here reaches the network: pending transfers are cancelled.
 */
final class GuzzleCurl3Test extends TestCase
{
    /** @var list<string> */
    private array $deprecations = [];

    private function collect(callable $fn)
    {
        $this->deprecations = [];
        set_error_handler(function ($no, $msg) {
            $this->deprecations[] = $msg;

            return true;
        }, E_USER_DEPRECATED | E_USER_WARNING);
        try {
            return $fn();
        } finally {
            restore_error_handler();
        }
    }

    private function bad(callable $fn, string $needle, string $class = \InvalidArgumentException::class): void
    {
        try {
            $this->collect($fn);
            $this->fail('expected '.$class.' with '.$needle);
        } catch (\Throwable $e) {
            $this->assertInstanceOf($class, $e, $e->getMessage());
            $this->assertStringContainsString($needle, $e->getMessage());
        }
    }

    private function req(string $version = '1.1', array $headers = []): Request
    {
        return new Request('GET', 'http://127.0.0.1:1/x', $headers, null, $version);
    }

    public function testUnknownConstructorOption(): void
    {
        $this->collect(fn () => new CurlMultiHandler(['bogus' => 1]));
        $this->assertStringContainsString('"bogus" CurlMultiHandler constructor option is unknown', $this->deprecations[0]);
    }

    public function testMultiplexConstructorOption(): void
    {
        $this->bad(fn () => new CurlMultiHandler(['multiplex' => Multiplexing::EAGER]), 'only accepts Multiplexing::NONE');
        $this->bad(fn () => new CurlMultiHandler(['multiplex' => Multiplexing::REQUIRE_WAIT]), 'only accepts Multiplexing::NONE');
        $this->bad(fn () => new CurlMultiHandler(['multiplex' => 'weird']), 'must be null or Multiplexing::NONE; received string');
        $this->bad(fn () => new CurlMultiHandler(['multiplex' => Multiplexing::NONE, 'options' => 5]), 'options must be an array of cURL multi options when using the "multiplex" option');
        $this->bad(fn () => new CurlMultiHandler(['multiplex' => Multiplexing::NONE, 'options' => [CURLMOPT_PIPELINING => 2]]), 'multiplex conflicts with a CURLMOPT_PIPELINING entry');
        $h = $this->collect(fn () => new CurlMultiHandler(['multiplex' => Multiplexing::NONE]));
        $this->assertInstanceOf(CurlMultiHandler::class, $h);
    }

    public function testSelectTimeoutOption(): void
    {
        foreach ([0, 0.5, '2', 1] as $ok) {
            $this->collect(fn () => new CurlMultiHandler(['select_timeout' => $ok]));
            $this->assertSame([], $this->deprecations, (string) $ok);
        }
        $this->collect(fn () => new CurlMultiHandler(['select_timeout' => 'abc']));
        $this->assertStringContainsString('non-numeric "select_timeout"', $this->deprecations[0]);
        foreach ([-1, 0.0001, NAN, INF] as $bad) {
            $this->collect(fn () => new CurlMultiHandler(['select_timeout' => $bad]));
            $this->assertStringContainsString('not 0 or greater than or equal to 0.001', $this->deprecations[0]);
        }
    }

    public function testSelectTimeoutEnvironment(): void
    {
        putenv('GUZZLE_CURL_SELECT_TIMEOUT=3');
        try {
            $this->collect(fn () => new CurlMultiHandler());
            $this->assertStringContainsString('GUZZLE_CURL_SELECT_TIMEOUT environment variable is deprecated', $this->deprecations[0]);
        } finally {
            putenv('GUZZLE_CURL_SELECT_TIMEOUT');
        }
    }

    public function testConnectionCaps(): void
    {
        $this->bad(fn () => new CurlMultiHandler(['max_host_connections' => 0]), 'max_host_connections must be a positive integer');
        $this->bad(fn () => new CurlMultiHandler(['max_total_connections' => 'x']), 'max_total_connections must be a positive integer');
        $this->bad(fn () => new CurlMultiHandler(['max_host_connections' => 2, 'options' => [CURLMOPT_MAX_HOST_CONNECTIONS => 4]]), 'max_host_connections conflicts with a CURLMOPT_MAX_HOST_CONNECTIONS entry');
        $this->bad(fn () => new CurlMultiHandler(['max_total_connections' => 2, 'options' => 5]), 'options must be an array of cURL multi options when using connection cap options');
        $h = $this->collect(fn () => new CurlMultiHandler(['max_host_connections' => 2, 'max_total_connections' => 5]));
        $this->assertInstanceOf(CurlMultiHandler::class, $h);
        // a connection cap forbids a request-level CURLOPT_SHARE
        $this->bad(
            fn () => $h($this->req(), ['curl' => [CURLOPT_SHARE => curl_share_init()]]),
            'cannot be combined with CurlMultiHandler connection cap options'
        );
    }

    public function testMultiOptionsDeprecated(): void
    {
        $h = $this->collect(fn () => new CurlMultiHandler(['options' => [CURLMOPT_MAX_HOST_CONNECTIONS => 3, CURLMOPT_MAX_TOTAL_CONNECTIONS => 3, CURLMOPT_PIPELINING => 2]]));
        $this->assertCount(3, $this->deprecations);
        $this->assertStringContainsString('cURL multi handler "options" is deprecated', $this->deprecations[0]);
        $this->assertInstanceOf(CurlMultiHandler::class, $h);
    }

    public function testHandleFactoryAndSharing(): void
    {
        $f = new CurlFactory(5);
        $h = $this->collect(fn () => new CurlMultiHandler(['handle_factory' => $f]));
        $this->assertInstanceOf(CurlMultiHandler::class, $h);
        $this->bad(fn () => new CurlMultiHandler(['handle_factory' => $f, 'transport_sharing' => TransportSharing::HANDLER_REQUIRE]), 'transport_sharing');
        $h = $this->collect(fn () => new CurlMultiHandler(['transport_sharing' => TransportSharing::HANDLER_PREFER]));
        $this->assertInstanceOf(CurlMultiHandler::class, $h);
        $h = $this->collect(fn () => new CurlHandler(['transport_sharing' => TransportSharing::HANDLER_PREFER]));
        $this->assertInstanceOf(CurlHandler::class, $h);
        $this->bad(fn () => new CurlHandler(['transport_sharing' => 'weird']), 'transport_sharing');
    }

    public function testGetOnlyKnowsMh(): void
    {
        $h = new CurlMultiHandler();
        $this->bad(fn () => $h->nothing, "Can not get other property as '_mh'", \BadMethodCallException::class);
        $this->bad(fn () => (function () { return $this->nothing; })->call($h), "Can not get other property as '_mh'", \BadMethodCallException::class);
    }

    public function testRequestLevelMultiplexValidation(): void
    {
        $h = new CurlMultiHandler();
        $this->bad(fn () => $h($this->req(), ['multiplex' => Multiplexing::EAGER, 'curl' => [CURLOPT_PIPEWAIT => true]]), 'raw CURLOPT_PIPEWAIT');
        $this->bad(fn () => $h($this->req('2'), ['multiplex' => Multiplexing::NONE]), 'only be Multiplexing::NONE for an HTTP/1.x request');
        $this->bad(fn () => $h($this->req(), ['multiplex' => Multiplexing::NONE, 'curl' => [CURLOPT_HTTPAUTH => 1]]), 'raw CURLOPT_HTTPAUTH cURL option');
        $this->bad(fn () => $h($this->req('1.1', ['Expect' => '100-continue']), ['multiplex' => Multiplexing::NONE]), '"Expect: 100-continue"');
        $this->bad(fn () => $h($this->req(), ['multiplex' => Multiplexing::NONE, 'curl' => [CURLOPT_HTTPHEADER => ['A: b']]]), 'raw CURLOPT_HTTPHEADER');

        $none = new CurlMultiHandler(['multiplex' => Multiplexing::NONE]);
        $this->bad(fn () => $none($this->req('2'), ['multiplex' => Multiplexing::WAIT]), 'whose "multiplex" option is Multiplexing::NONE');
        if (\GuzzleHttp\Handler\CurlVersion::supportsRequiredMultiplex()) {
            $this->bad(fn () => $none($this->req('2'), ['multiplex' => Multiplexing::REQUIRE_WAIT]), 'Multiplexing::NONE');
        } else {
            $this->bad(fn () => $none($this->req('2'), ['multiplex' => Multiplexing::REQUIRE_WAIT]), 'Required multiplexing needs libcurl', \GuzzleHttp\Exception\ConnectException::class);
        }

        $piped = new CurlMultiHandler(['options' => [CURLMOPT_PIPELINING => 0]]);
        $this->bad(fn () => $piped($this->req('2'), ['multiplex' => Multiplexing::WAIT]), 'disables multiplexing');
        $this->bad(fn () => $piped($this->req(), ['multiplex' => Multiplexing::NONE]), 'raw CURLMOPT_PIPELINING');
        $ok = new CurlMultiHandler(['options' => [CURLMOPT_PIPELINING => 2]]);
        $this->collect(function () use ($ok) {
            $p = $ok($this->req('2'), ['multiplex' => Multiplexing::WAIT]);
            $p->cancel();
        });
        $weird = new CurlMultiHandler(['options' => [CURLMOPT_PIPELINING => [1]]]);
        $this->bad(fn () => $weird($this->req('2'), ['multiplex' => Multiplexing::WAIT]), 'PIPELINING option must be an integer');

        $custom = new CurlMultiHandler(['handle_factory' => new CurlFactory(5)]);
        $this->bad(fn () => $custom($this->req(), ['multiplex' => Multiplexing::NONE]), 'custom "handle_factory"');
    }

    public function testNoneMultiplexingCancel(): void
    {
        $h = new CurlMultiHandler();
        $this->collect(function () use ($h) {
            $p = $h($this->req(), ['multiplex' => Multiplexing::NONE]);
            $this->assertTrue($p->getState() === 'pending');
            $p->cancel();
            $this->assertSame('rejected', $p->getState());
        });
    }

    public function testCurlHandlerRejectsBadHost(): void
    {
        $h = new CurlHandler();
        $this->bad(fn () => $h(new Request('GET', 'ftp://127.0.0.1/x'), []), 'is not supported', \GuzzleHttp\Exception\RequestException::class);
    }
}
