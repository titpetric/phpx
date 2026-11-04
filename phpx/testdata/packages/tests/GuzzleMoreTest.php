<?php
use GuzzleHttp\{Client, HandlerStack, Middleware, RequestOptions, Utils, BodySummarizer, RetryMiddleware, Pool, MessageFormatter, ClientInterface};
use GuzzleHttp\Handler\MockHandler;
use GuzzleHttp\Promise as P;
use GuzzleHttp\Promise\{Promise, FulfilledPromise, RejectedPromise, Utils as PU, Create, Each, Coroutine};
use GuzzleHttp\Psr7\{Response, Request};
use GuzzleHttp\Exception\{ClientException, ServerException, RequestException, ConnectException, TransferException};
use GuzzleHttp\Cookie\{CookieJar, SetCookie};
use PHPUnit\Framework\TestCase;

final class GuzzleMoreTest extends TestCase
{
    public function testPromiseChain(): void
    {
        $p = new Promise();
        $p->resolve(2);
        $r = $p->then(fn ($v) => $v * 3)->then(fn ($v) => $v + 1);
        $this->assertSame(7, $r->wait());
        $this->assertSame('fulfilled', $r->getState());
    }
    public function testPromiseRejection(): void
    {
        $p = new RejectedPromise('bad');
        $r = $p->otherwise(fn ($e) => "recovered $e");
        $this->assertSame('recovered bad', $r->wait());
        $this->expectException(P\RejectionException::class);
        (new RejectedPromise('nope'))->wait();
    }
    public function testPromiseAll(): void
    {
        $all = PU::all([new FulfilledPromise(1), new FulfilledPromise(2), Create::promiseFor(3)])->wait();
        $this->assertSame([1, 2, 3], $all);
        $settled = PU::settle([new FulfilledPromise('a'), new RejectedPromise('b')])->wait();
        $this->assertSame('fulfilled', $settled[0]['state']);
        $this->assertSame('rejected', $settled[1]['state']);
        $this->assertSame('b', $settled[1]['reason']);
    }
    public function testQueueRuns(): void
    {
        $log = [];
        $p = new Promise(function () use (&$p, &$log) { $log[] = 'wait'; $p->resolve('done'); });
        $p->then(function ($v) use (&$log) { $log[] = $v; });
        $p->wait();
        P\Utils::queue()->run();
        $this->assertSame(['wait', 'done'], $log);
    }
    public function testErrors(): void
    {
        $mock = new MockHandler([new Response(404, [], 'nf'), new Response(500, [], 'srv'), new ConnectException('down', new Request('GET', 'http://x/'))]);
        $c = new Client(['handler' => HandlerStack::create($mock)]);
        try { $c->get('http://x/a'); $this->fail('no exception'); } catch (ClientException $e) {
            $this->assertSame(404, $e->getResponse()->getStatusCode());
            $this->assertSame('GET', $e->getRequest()->getMethod());
            $this->assertStringContainsString('404 Not Found', $e->getMessage());
        }
        try { $c->get('http://x/b'); $this->fail('no exception'); } catch (ServerException $e) {
            $this->assertSame(500, $e->getResponse()->getStatusCode());
        }
        $this->expectException(ConnectException::class);
        $c->get('http://x/c');
    }
    public function testHttpErrorsOff(): void
    {
        $c = new Client(['handler' => HandlerStack::create(new MockHandler([new Response(418)])), 'http_errors' => false]);
        $this->assertSame(418, $c->get('http://x/')->getStatusCode());
    }
    public function testRequestOptions(): void
    {
        $h = [];
        $stack = HandlerStack::create(new MockHandler([new Response(200), new Response(200), new Response(200)]));
        $stack->push(Middleware::history($h));
        $c = new Client(['handler' => $stack, 'base_uri' => 'http://api.test/v1/', 'headers' => ['X-Def' => 'd']]);
        $c->post('items', [RequestOptions::JSON => ['a' => 1]]);
        $c->post('form', [RequestOptions::FORM_PARAMS => ['x' => 'y z']]);
        $c->get('q', [RequestOptions::QUERY => ['a' => [1, 2]], RequestOptions::HEADERS => ['X-Req' => 'r']]);
        $this->assertSame('http://api.test/v1/items', (string) $h[0]['request']->getUri());
        $this->assertSame('{"a":1}', (string) $h[0]['request']->getBody());
        $this->assertSame('application/json', $h[0]['request']->getHeaderLine('Content-Type'));
        $this->assertSame('x=y+z', (string) $h[1]['request']->getBody());
        $this->assertSame('application/x-www-form-urlencoded', $h[1]['request']->getHeaderLine('Content-Type'));
        $this->assertSame('a%5B0%5D=1&a%5B1%5D=2', $h[2]['request']->getUri()->getQuery());
        $this->assertSame('d', $h[2]['request']->getHeaderLine('X-Def'));
        $this->assertSame('r', $h[2]['request']->getHeaderLine('X-Req'));
    }
    public function testRetryMiddleware(): void
    {
        $mock = new MockHandler([new Response(503), new Response(503), new Response(200, [], 'ok')]);
        $stack = HandlerStack::create($mock);
        $stack->push(Middleware::retry(fn ($n, $req, $res) => $n < 3 && $res && $res->getStatusCode() >= 500, fn () => 0));
        $c = new Client(['handler' => $stack, 'http_errors' => false]);
        $this->assertSame('ok', (string) $c->get('http://x/')->getBody());
    }
    public function testAsyncAndPool(): void
    {
        $c = new Client(['handler' => HandlerStack::create(new MockHandler([new Response(200, [], 'a'), new Response(200, [], 'b')]))]);
        $ps = [$c->getAsync('http://x/1'), $c->getAsync('http://x/2')];
        $rs = PU::unwrap($ps);
        $this->assertSame(['a', 'b'], array_map(fn ($r) => (string) $r->getBody(), $rs));
        $c2 = new Client(['handler' => HandlerStack::create(new MockHandler([new Response(200), new Response(201)]))]);
        $seen = [];
        $pool = new Pool($c2, [new Request('GET', 'http://x/1'), new Request('GET', 'http://x/2')], ['concurrency' => 2, 'fulfilled' => function ($r, $i) use (&$seen) { $seen[$i] = $r->getStatusCode(); }]);
        $pool->promise()->wait();
        ksort($seen);
        $this->assertSame([200, 201], array_values($seen));
    }
    public function testMessageFormatterAndCookies(): void
    {
        $f = new MessageFormatter('{method} {uri} {code}');
        $this->assertSame('GET http://x/p 200', $f->format(new Request('GET', 'http://x/p'), new Response(200)));
        $jar = new CookieJar();
        $jar->setCookie(new SetCookie(['Name' => 'sid', 'Value' => 'abc', 'Domain' => 'x.test', 'Path' => '/']));
        $this->assertCount(1, $jar);
        $this->assertSame('sid=abc', $jar->getCookieByName('sid') ? 'sid=' . $jar->getCookieByName('sid')->getValue() : '');
        $sc = SetCookie::fromString('a=b; Domain=d.test; Path=/p; Secure; HttpOnly');
        $this->assertSame('d.test', $sc->getDomain());
        $this->assertTrue($sc->getSecure());
        $this->assertTrue($sc->getHttpOnly());
    }
    public function testUtilsAndSummarizer(): void
    {
        $this->assertSame('GET', Utils::describeType('GET') === 'string(3) "GET"' ? 'GET' : 'x');
        $this->assertSame('GuzzleHttp', explode('\\', Client::class)[0]);
        $this->assertSame('x.test', Utils::idnUriConvert(new \GuzzleHttp\Psr7\Uri('http://x.test/'))->getHost());
        $this->assertSame(['a', 'b'], Utils::jsonDecode('["a","b"]', true));
        $this->assertSame('{"a":1}', Utils::jsonEncode(['a' => 1]));
        $s = new BodySummarizer(5);
        $this->assertSame('abcde (truncated...)', $s->summarize(new Response(500, [], 'abcdefghij')));
    }
}
