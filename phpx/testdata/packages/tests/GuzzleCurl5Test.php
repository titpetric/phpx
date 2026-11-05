<?php

use GuzzleHttp\Client;
use GuzzleHttp\Exception\ConnectException;
use GuzzleHttp\Exception\RequestException;
use GuzzleHttp\Handler\CurlFactory;
use GuzzleHttp\Handler\CurlHandler;
use GuzzleHttp\Handler\CurlMultiHandler;
use GuzzleHttp\HandlerStack;
use GuzzleHttp\Pool;
use GuzzleHttp\Promise\Utils as PromiseUtils;
use GuzzleHttp\Psr7\Request;
use GuzzleHttp\Psr7\Utils as PU7;
use PHPUnit\Framework\Attributes\DataProvider;

require_once __DIR__.'/GuzzleCurlServer.php';

/**
 * CurlMultiHandler scheduling and CurlHandler against a live server.
 */
final class GuzzleCurl5Test extends GuzzleCurlBase
{
    private function req(string $path = '/text?t=x', string $method = 'GET', array $headers = [], $body = null): Request
    {
        return new Request($method, self::url($path), $headers, $body);
    }

    public static function kinds(): array
    {
        return [['curl'], ['multi']];
    }

    private function client(string $kind, array $extra = []): Client
    {
        $h = $kind === 'multi' ? new CurlMultiHandler() : new CurlHandler();

        return new Client($extra + ['handler' => HandlerStack::create($h)]);
    }

    public function testDelayedRequestsAndExecute(): void
    {
        $h = new CurlMultiHandler();
        $t0 = microtime(true);
        $p1 = $h($this->req('/text?t=late'), ['delay' => 700]);
        $p2 = $h($this->req('/text?t=soon'), []);
        $order = [];
        $p1->then(function ($r) use (&$order) {
            $order[] = (string) $r->getBody();
        });
        $p2->then(function ($r) use (&$order) {
            $order[] = (string) $r->getBody();
        });
        $h->execute();
        sort($order);
        $this->assertSame(['late', 'soon'], $order);
        $this->assertGreaterThanOrEqual(0.6, microtime(true) - $t0);
    }

    public function testWaitOnDelayed(): void
    {
        $h = new CurlMultiHandler(['select_timeout' => 0.05]);
        $p = $h($this->req('/text?t=waited'), ['delay' => 100]);
        $this->assertSame('waited', (string) $p->wait()->getBody());
    }

    public function testManualTick(): void
    {
        $h = new CurlMultiHandler();
        $p = $h($this->req('/text?t=ticked'), []);
        for ($i = 0; $i < 400 && $p->getState() === 'pending'; $i++) {
            $h->tick();
        }
        $this->assertSame('fulfilled', $p->getState());
        $this->assertSame('ticked', (string) $p->wait()->getBody());
    }

    public function testCancelAttachedAndDelayed(): void
    {
        $h = new CurlMultiHandler();
        $a = $h($this->req('/slow?us=500000'), []);
        $b = $h($this->req('/text'), ['delay' => 5000]);
        $c = $h($this->req('/text?t=kept'), []);
        $a->cancel();
        $b->cancel();
        $this->assertSame('rejected', $a->getState());
        $this->assertSame('rejected', $b->getState());
        $this->assertSame('kept', (string) $c->wait()->getBody());
        $h->execute();
        $this->expectException(\GuzzleHttp\Promise\CancellationException::class);
        $a->wait();
    }

    public function testConcurrentSlowRequestsOverlap(): void
    {
        // two servers: one php -S process can queue both connections on one worker
        $second = GuzzleCurlServer::start();
        try {
            $c = $this->client('multi');
            $t0 = microtime(true);
            $ps = [$c->getAsync(self::url('/slow?us=400000')), $c->getAsync($second->url('/slow?us=400000'))];
            $res = PromiseUtils::unwrap($ps);
            $this->assertCount(2, $res);
            $this->assertLessThan(0.75, microtime(true) - $t0, 'two 0.4s transfers overlap');
        } finally {
            $second->stop();
        }
    }

    public function testPoolWithConcurrencyAndMaxConnections(): void
    {
        $h = new CurlMultiHandler(['max_total_connections' => 2, 'max_host_connections' => 2, 'select_timeout' => 0.1]);
        $c = new Client(['handler' => HandlerStack::create($h)]);
        $done = [];
        $reqs = function () {
            for ($i = 0; $i < 6; $i++) {
                yield new Request('GET', self::url('/text?t=p'.$i));
            }
        };
        $pool = new Pool($c, $reqs(), ['concurrency' => 3, 'fulfilled' => function ($r, $i) use (&$done) {
            $done[$i] = (string) $r->getBody();
        }]);
        $pool->promise()->wait();
        ksort($done);
        $this->assertSame(['p0', 'p1', 'p2', 'p3', 'p4', 'p5'], array_values($done));
    }

    public function testCallbacksOnTheSameMultiHandler(): void
    {
        $h = new CurlMultiHandler();
        $other = null;
        $nestedWait = null;
        $executeError = null;
        $added = null;
        $c = new Client(['handler' => HandlerStack::create($h)]);
        $slow = GuzzleCurlServer::start(); // keeps the slow transfers off a's worker
        $b = $c->getAsync($slow->url('/slow?us=1000000'));
        $cancelMe = $c->getAsync($slow->url('/slow?us=1000000'));
        $a = $c->getAsync(self::url('/text?t=a'), ['on_headers' => function () use (&$nestedWait, &$executeError, &$added, $b, $cancelMe, $c, $h) {
            try {
                $b->wait();
            } catch (\Throwable $e) {
                $nestedWait = $e;
            }
            try {
                $h->execute();
            } catch (\LogicException $e) {
                $executeError = $e;
            }
            $cancelMe->cancel();
            $added = $c->getAsync(self::url('/text?t=added'));
        }]);
        $this->assertSame('a', (string) $a->wait()->getBody());
        $this->assertInstanceOf(RequestException::class, $nestedWait);
        $this->assertStringContainsString('Cannot synchronously wait for a transfer from inside a cURL callback', $nestedWait->getMessage());
        $this->assertInstanceOf(\LogicException::class, $executeError);
        $this->assertSame('added', (string) $added->wait()->getBody());
        $this->assertSame('rejected', $cancelMe->getState());
        $h->execute();
        $slow->stop();
    }

    public function testOnStatsFailureRejectsOnlyThatRequest(): void
    {
        $h = new CurlMultiHandler();
        $c = new Client(['handler' => HandlerStack::create($h)]);
        $bad = $c->getAsync(self::url('/text?t=bad'), ['on_stats' => function () {
            throw new \DomainException('stats broke');
        }]);
        $good = $c->getAsync(self::url('/text?t=good'));
        $this->assertSame('good', (string) $good->wait()->getBody());
        try {
            $bad->wait();
            $this->fail('expected');
        } catch (\DomainException $e) {
            $this->assertSame('stats broke', $e->getMessage());
        }
    }

    #[DataProvider('kinds')]
    public function testOnStatsFailureDirect(string $k): void
    {
        try {
            $this->client($k)->get(self::url('/text'), ['on_stats' => function () {
                throw new \DomainException('stats broke');
            }]);
            $this->fail('expected');
        } catch (\DomainException $e) {
            $this->assertSame('stats broke', $e->getMessage());
        }
    }

    public function testDisplacedAndWaitOnSettled(): void
    {
        $h = new CurlMultiHandler();
        $p = $h($this->req('/text?t=once'), []);
        $this->assertSame('once', (string) $p->wait()->getBody());
        $this->assertSame('once', (string) $p->wait()->getBody());
        $q = $h($this->req('/text?t=twice'), []);
        $this->assertSame('twice', (string) $q->wait()->getBody());
    }

    #[DataProvider('kinds')]
    public function testResolveOption(string $k): void
    {
        $port = parse_url(self::url(), PHP_URL_PORT);
        $r = $this->client($k)->get('http://resolve.test:'.$port.'/echo', ['curl' => [CURLOPT_RESOLVE => ['resolve.test:'.$port.':127.0.0.1']]]);
        $j = json_decode((string) $r->getBody(), true);
        $this->assertSame('resolve.test:'.$port, $j['headers']['Host']);
    }

    #[DataProvider('kinds')]
    public function testProxyTransfer(string $k): void
    {
        $proxy = self::url();
        $r = $this->client($k)->get('http://proxied.test/echo', ['proxy' => $proxy]);
        $j = json_decode((string) $r->getBody(), true);
        $this->assertSame('/echo', $j['path']);
        $this->assertSame('proxied.test', $j['headers']['Host']);

        $r = $this->client($k)->get('http://proxied.test/echo', ['proxy' => 'http://pu:pp@127.0.0.1:'.parse_url($proxy, PHP_URL_PORT)]);
        $j = json_decode((string) $r->getBody(), true);
        $this->assertSame('Basic '.base64_encode('pu:pp'), $j['headers']['Proxy-Authorization']);

        // no_proxy bypasses an unusable proxy
        $r = $this->client($k)->get(self::url('/text?t=direct'), ['proxy' => ['http' => 'http://127.0.0.1:1', 'no' => ['127.0.0.1']]]);
        $this->assertSame('direct', (string) $r->getBody());

        // an unusable proxy is a connect error
        $this->expectException(ConnectException::class);
        $this->client($k)->get(self::url('/text'), ['proxy' => 'http://127.0.0.1:1', 'connect_timeout' => 2]);
    }

    #[DataProvider('kinds')]
    public function testTlsToPlainServerFails(string $k): void
    {
        $url = 'https://127.0.0.1:'.parse_url(self::url(), PHP_URL_PORT).'/echo';
        foreach ([true, false] as $verify) {
            try {
                $this->client($k)->get($url, ['verify' => $verify, 'connect_timeout' => 3, 'timeout' => 5]);
                $this->fail('expected');
            } catch (ConnectException $e) {
                $this->assertStringContainsString('cURL error 35', $e->getMessage());
                $this->assertNotNull($e->getHandlerContext());
            }
        }
    }

    #[DataProvider('kinds')]
    public function testLargeUploadFromStream(string $k): void
    {
        $file = __DIR__.'/gcurl5-upload-'.getmypid();
        file_put_contents($file, str_repeat('u', 1200000));
        try {
            $fh = fopen($file, 'r');
            $r = $this->client($k)->post(self::url('/echo'), ['body' => $fh, 'headers' => ['Content-Length' => '1200000']]);
            $j = json_decode((string) $r->getBody(), true);
            $this->assertSame(1200000, strlen($j['body']));
            $r = $this->client($k)->put(self::url('/echo'), ['body' => PU7::streamFor(fopen($file, 'r'))]);
            $this->assertSame(1200000, strlen(json_decode((string) $r->getBody(), true)['body']));
        } finally {
            unlink($file);
        }
    }

    #[DataProvider('kinds')]
    public function testHandlerReuseAcrossRequests(string $k): void
    {
        $f = new CurlFactory(1);
        $h = $k === 'multi' ? new CurlMultiHandler(['handle_factory' => $f]) : new CurlHandler(['handle_factory' => $f]);
        $c = new Client(['handler' => HandlerStack::create($h)]);
        for ($i = 0; $i < 3; $i++) {
            $this->assertSame('r'.$i, (string) $c->get(self::url('/text?t=r'.$i))->getBody());
        }
        $this->assertSame(200, $c->get(self::url('/status?code=200'))->getStatusCode());
    }

    #[DataProvider('kinds')]
    public function testSharingModes(string $k): void
    {
        foreach (['handler_prefer', 'handler_require'] as $mode) {
            try {
                $h = $k === 'multi' ? new CurlMultiHandler(['transport_sharing' => $mode]) : new CurlHandler(['transport_sharing' => $mode]);
            } catch (\InvalidArgumentException $e) {
                $this->assertFalse(\GuzzleHttp\Handler\CurlVersion::supportsSslSessionSharing());
                $this->assertStringContainsString('requires libcurl 8.6.0 or higher', $e->getMessage());

                continue;
            }
            $c = new Client(['handler' => HandlerStack::create($h)]);
            $this->assertSame('shared', (string) $c->get(self::url('/text?t=shared'))->getBody());
            $this->assertSame('shared', (string) $c->get(self::url('/text?t=shared'))->getBody());
        }
    }
}
