<?php

use GuzzleHttp\Client;
use GuzzleHttp\Exception\ConnectException;
use GuzzleHttp\Exception\RequestException;
use GuzzleHttp\Handler\CurlHandler;
use GuzzleHttp\Handler\CurlMultiHandler;
use GuzzleHttp\HandlerStack;
use GuzzleHttp\Psr7\Request;
use PHPUnit\Framework\Attributes\DataProvider;

require_once __DIR__.'/GuzzleCurlServer.php';

/**
 * Trailers, response and rewind failures, error redaction and proxy tunnel
 * ownership of the cURL handlers against a live server.
 */
final class GuzzleCurl6Test extends GuzzleCurlBase
{
    public static function kinds(): array
    {
        return [['curl'], ['multi']];
    }

    private function handler(string $kind)
    {
        return $kind === 'multi' ? new CurlMultiHandler() : new CurlHandler();
    }

    private function client(string $kind, array $extra = []): Client
    {
        return new Client($extra + ['handler' => HandlerStack::create($this->handler($kind))]);
    }

    #[DataProvider('kinds')]
    public function testTrailers(string $k): void
    {
        $got = null;
        $r = $this->client($k)->get(self::url('/trailers'), ['on_trailers' => function (array $trailers, $response) use (&$got) {
            $got = [$trailers, $response->getStatusCode()];
        }]);
        $this->assertSame('hello', (string) $r->getBody());
        $this->assertSame([['x-trailer-one' => ['one'], 'x-trailer-two' => ['two']], 200], $got);
    }

    #[DataProvider('kinds')]
    public function testTrailersWithoutCallbackAndFailingCallback(string $k): void
    {
        $this->assertSame('hello', (string) $this->client($k)->get(self::url('/trailers'))->getBody());
        $this->assertSame('hello', (string) $this->client($k)->get(self::url('/text?t=hello'), ['on_trailers' => fn () => null])->getBody());
        try {
            $this->client($k)->get(self::url('/trailers'), ['on_trailers' => function () {
                throw new \RuntimeException('trailers broke');
            }]);
            $this->fail('expected');
        } catch (RequestException $e) {
            $this->assertStringContainsString('An error was encountered during the on_trailers event', $e->getMessage());
            $this->assertSame('trailers broke', $e->getPrevious()->getMessage());
        }
    }

    #[DataProvider('kinds')]
    public function testInvalidResponseHeaderValue(string $k): void
    {
        try {
            $this->client($k)->get(self::url('/ctl'));
            $this->fail('expected');
        } catch (RequestException $e) {
            $this->assertStringContainsString('An error was encountered while creating the response', $e->getMessage());
            $this->assertInstanceOf(\InvalidArgumentException::class, $e->getPrevious());
        }
    }

    #[DataProvider('kinds')]
    public function testRewindRetries(string $k): void
    {
        $fh = fopen('php://temp', 'w+');
        fwrite($fh, str_repeat('x', 1200000));
        rewind($fh);
        try {
            $this->client($k)->post(self::url('/redirect?code=307'), [
                'body' => $fh,
                'headers' => ['Content-Length' => '1200000'],
                'allow_redirects' => false,
                'curl' => [CURLOPT_FOLLOWLOCATION => true],
            ]);
            $this->fail('expected');
        } catch (RequestException $e) {
            $this->assertStringContainsString('cURL error 65', $e->getMessage());
            $this->assertStringContainsString('was retried 3 times and did not succeed', $e->getMessage());
        }
    }

    #[DataProvider('kinds')]
    public function testRewindFailureOfUnseekableBody(string $k): void
    {
        $body = \GuzzleHttp\Psr7\FnStream::decorate(\GuzzleHttp\Psr7\Utils::streamFor(str_repeat('x', 1200000)), [
            'rewind' => function () {
                throw new \RuntimeException('cannot rewind');
            },
            'isSeekable' => fn () => false,
        ]);
        try {
            $this->client($k)->post(self::url('/redirect?code=307'), [
                'body' => $body,
                'headers' => ['Content-Length' => '1200000'],
                'allow_redirects' => false,
                'curl' => [CURLOPT_FOLLOWLOCATION => true],
            ]);
            $this->fail('expected');
        } catch (RequestException $e) {
            $this->assertStringContainsString('attempting to rewind the request body failed', $e->getMessage());
        }
    }

    #[DataProvider('kinds')]
    public function testCredentialsAreRedactedFromErrors(string $k): void
    {
        foreach ([
            ['http://127.0.0.1:1/path?token=abc', ['proxy' => 'http://pxuser:pxsecret@127.0.0.1:1']],
            ['http://uuser:usecret@127.0.0.1:1/path', []],
            ['http://127.0.0.1:1/path', ['proxy' => 'http://pxuser:pxsecret@127.0.0.1:1/']],
            ['http://127.0.0.1:1/path', ['proxy' => 'pxuser:pxsecret@127.0.0.1:1']],
            ['http://127.0.0.1:1/path', ['proxy' => 'http://pxuser:p@xsecret@127.0.0.1:1']],
            ['http://127.0.0.1:1/path', ['proxy' => 'http://only@127.0.0.1:1']],
        ] as [$url, $opts]) {
            try {
                $this->client($k)->get($url, $opts + ['connect_timeout' => 2]);
                $this->fail('expected');
            } catch (\GuzzleHttp\Exception\TransferException $e) {
                $this->assertStringNotContainsString('pxsecret', $e->getMessage());
                $this->assertStringNotContainsString('usecret', $e->getMessage());
                $this->assertStringNotContainsString('p@xsecret', $e->getMessage());
                $this->assertStringContainsString('cURL error', $e->getMessage());
            }
        }
    }

    public function testProxyTunnelOwnershipOnTheMultiHandler(): void
    {
        $h = new CurlMultiHandler();
        $req = new Request('GET', 'https://tunnel.test/x');
        $opts = fn (string $user) => ['proxy' => 'http://'.$user.':pw@127.0.0.1:1', 'connect_timeout' => 2];
        $set_error = set_error_handler(fn () => true);
        try {
            $p1 = $h($req, $opts('first'));
            $p1b = $h($req, $opts('first'));
            $p2 = $h($req, $opts('second'));
            $outcomes = [];
            foreach ([$p1, $p1b, $p2] as $p) {
                try {
                    $p->wait();
                } catch (\Throwable $e) {
                    $outcomes[] = get_class($e);
                }
            }
            $this->assertSame([ConnectException::class, ConnectException::class, ConnectException::class], $outcomes);
            // the pool is idle now: a third owner replaces the multi handle
            $p3 = $h($req, $opts('third'));
            try {
                $p3->wait();
                $this->fail('expected');
            } catch (ConnectException $e) {
                $this->assertStringContainsString('cURL error 7', $e->getMessage());
            }
            // and an anonymous direct request still works afterwards
            $ok = $h(new Request('GET', self::url('/text?t=after')), []);
            $this->assertSame('after', (string) $ok->wait()->getBody());
        } finally {
            if ($set_error !== null) {
                restore_error_handler();
            }
        }
    }

    public function testProxyTunnelOwnershipOnTheCurlHandler(): void
    {
        $h = new CurlHandler();
        $req = new Request('GET', 'https://tunnel.test/x');
        foreach (['first', 'first', 'second', 'third'] as $user) {
            try {
                $h($req, ['proxy' => 'http://'.$user.':pw@127.0.0.1:1', 'connect_timeout' => 2])->wait();
                $this->fail('expected');
            } catch (ConnectException $e) {
                $this->assertStringContainsString('cURL error 7', $e->getMessage());
            }
        }
        $this->assertSame('plain', (string) $h(new Request('GET', self::url('/text?t=plain')), [])->wait()->getBody());
    }

    public function testConnectionRefusedStats(): void
    {
        $h = new CurlHandler();
        $stats = null;
        try {
            $h(new Request('GET', 'http://127.0.0.1:1/'), ['connect_timeout' => 2, 'on_stats' => function ($s) use (&$stats) {
                $stats = $s;
            }])->wait();
            $this->fail('expected');
        } catch (ConnectException $e) {
            $this->assertInstanceOf(\GuzzleHttp\TransferStats::class, $stats);
            $this->assertTrue($stats->hasResponse() === false);
            $this->assertSame(7, $stats->getHandlerErrorData());
        }
    }
}
