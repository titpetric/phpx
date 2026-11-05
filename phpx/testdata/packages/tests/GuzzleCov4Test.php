<?php
use GuzzleHttp\{Client, HandlerStack, Middleware, MessageFormatter, Pool, BodySummarizer, RetryMiddleware, PrepareBodyMiddleware, RequestOptions};
use GuzzleHttp\Handler\MockHandler;
use GuzzleHttp\Exception\{RequestException, ConnectException, ClientException, ServerException, BadResponseException, TransferException, TooManyRedirectsException};
use GuzzleHttp\Promise\{Promise, FulfilledPromise, RejectedPromise};
use GuzzleHttp\Psr7\{Request, Response, Utils as PU7, Stream};
use PHPUnit\Framework\TestCase;
use Psr\Log\{AbstractLogger, LogLevel};

final class GuzzleCov4Logger extends AbstractLogger
{
    public array $records = [];
    public function log($level, $message, array $context = []): void
    {
        $this->records[] = [$level, (string) $message];
    }
}

final class GuzzleCov4Test extends TestCase
{
    private function ok(): callable
    {
        return fn ($r, $o) => new FulfilledPromise(new Response(200));
    }

    // --- HandlerStack ---
    public function testStackOrderAndToString(): void
    {
        $log = [];
        $mk = function (string $n) use (&$log) {
            return function (callable $h) use ($n, &$log) {
                return function ($r, $o) use ($h, $n, &$log) { $log[] = $n; return $h($r, $o); };
            };
        };
        $s = new HandlerStack($this->ok());
        $s->push($mk('b'), 'b');
        $s->unshift($mk('a'), 'a');
        $s->push($mk('d'), 'd');
        $s->before('d', $mk('c'), 'c');
        $s->after('d', $mk('e'), 'e');
        $s->after('a', $mk('a2'), 'a2');
        $s->before('a', $mk('a0'), 'a0');
        $s($this->req(), [])->wait();
        $this->assertSame(['a0', 'a', 'a2', 'b', 'c', 'd', 'e'], $log);
        $str = (string) $s;
        $this->assertStringContainsString("Name: 'a0'", $str);
        $this->assertStringContainsString('0) Handler: callable(', $str);
        $this->assertTrue($s->hasHandler());
    }

    private function req(): Request
    {
        return new Request('GET', 'http://x.test/');
    }

    public function testStackRemove(): void
    {
        $log = [];
        $mw = function (callable $h) use (&$log) { return function ($r, $o) use ($h, &$log) { $log[] = 'mw'; return $h($r, $o); }; };
        $s = new HandlerStack($this->ok());
        $s->push($mw, 'named');
        $s->push(fn (callable $h) => $h, 'other');
        $s->remove('named');
        $s($this->req(), [])->wait();
        $this->assertSame([], $log);
        $s->push($mw);
        $s->remove($mw);
        $s($this->req(), [])->wait();
        $this->assertSame([], $log);
        $s->remove('missing');
        $this->assertStringNotContainsString('mw', (string) $s);
    }

    public function testStackToStringCallableForms(): void
    {
        $s = new HandlerStack($this->ok());
        $s->push('GuzzleCov4Fn', 'str');
        $s->push([new GuzzleCov4Obj, 'mw'], 'arr');
        $s->push(['GuzzleCov4Obj', 'smw'], 'static');
        $str = (string) $s;
        $this->assertStringContainsString('callable(GuzzleCov4Fn)', $str);
        $this->assertStringContainsString("callable(['GuzzleCov4Obj', 'mw'])", $str);
        $this->assertStringContainsString('callable(GuzzleCov4Obj::smw)', $str);
    }

    public function testStackNoHandler(): void
    {
        $s = new HandlerStack();
        $this->assertFalse($s->hasHandler());
        $this->expectException(\LogicException::class);
        $s->resolve();
    }

    public function testStackSetHandlerAndCache(): void
    {
        $s = new HandlerStack();
        $s->setHandler($this->ok());
        $a = $s->resolve();
        $this->assertSame($a, $s->resolve());
        $s->push(fn (callable $h) => fn ($r, $o) => $h($r, $o));
        $this->assertNotSame($a, $s->resolve());
    }

    public function testStackMissingName(): void
    {
        $s = new HandlerStack($this->ok());
        $this->expectException(\InvalidArgumentException::class);
        $s->before('missing', fn ($h) => $h);
    }

    public function testStackCreateDefaults(): void
    {
        $s = HandlerStack::create($this->ok());
        $str = (string) $s;
        foreach (['http_errors', 'allow_redirects', 'cookies', 'prepare_body'] as $n) {
            $this->assertStringContainsString("'$n'", $str);
        }
    }

    // --- Middleware ---
    public function testTapMiddleware(): void
    {
        $log = [];
        $s = HandlerStack::create(new MockHandler([new Response(200)]));
        $s->push(Middleware::tap(function ($r, $o) use (&$log) { $log[] = 'before'; }, function ($r, $o, $p) use (&$log) { $log[] = 'after'; $this->assertInstanceOf(\GuzzleHttp\Promise\PromiseInterface::class, $p); }));
        (new Client(['handler' => $s]))->get('http://x.test/');
        $this->assertSame(['before', 'after'], $log);
        $s2 = HandlerStack::create(new MockHandler([new Response(200)]));
        $s2->push(Middleware::tap());
        $this->assertSame(200, (new Client(['handler' => $s2]))->get('http://x.test/')->getStatusCode());
    }

    public function testMapRequestAndResponse(): void
    {
        $h = [];
        $s = HandlerStack::create(new MockHandler([new Response(200, ['A' => 'b'])]));
        $s->push(Middleware::mapRequest(fn (Request $r) => $r->withHeader('X-Mapped', '1')));
        $s->push(Middleware::mapResponse(fn (Response $r) => $r->withHeader('X-Res', '2')));
        $s->push(Middleware::history($h));
        $r = (new Client(['handler' => $s]))->get('http://x.test/');
        $this->assertSame('2', $r->getHeaderLine('X-Res'));
        $this->assertSame('1', $h[0]['request']->getHeaderLine('X-Mapped'));
    }

    public function testHistoryRecordsRejections(): void
    {
        $h = [];
        $s = HandlerStack::create(new MockHandler([new ConnectException('boom', $this->req()), new Response(200)]));
        $s->push(Middleware::history($h));
        $c = new Client(['handler' => $s]);
        try { $c->get('http://x.test/'); $this->fail('expected'); } catch (ConnectException $e) {}
        $c->get('http://x.test/');
        $this->assertNull($h[0]['response']);
        $this->assertInstanceOf(ConnectException::class, $h[0]['error']);
        $this->assertNull($h[1]['error']);
        $this->assertSame(200, $h[1]['response']->getStatusCode());
    }

    public function testHistoryArrayAccess(): void
    {
        $h = new ArrayObject();
        $s = HandlerStack::create(new MockHandler([new Response(200)]));
        $s->push(Middleware::history($h));
        (new Client(['handler' => $s]))->get('http://x.test/');
        $this->assertCount(1, $h);
    }

    public function testHistoryRejectsBadContainer(): void
    {
        $this->expectException(\InvalidArgumentException::class);
        $x = 5;
        Middleware::history($x);
    }

    public function testLogMiddleware(): void
    {
        $logger = new GuzzleCov4Logger();
        $s = HandlerStack::create(new MockHandler([new Response(200), new Response(404), new ConnectException('down', $this->req())]));
        $s->push(Middleware::log($logger, new MessageFormatter('{method} {uri} {code} {error}')));
        $c = new Client(['handler' => $s, 'http_errors' => false]);
        $c->get('http://x.test/a');
        $c->get('http://x.test/b');
        try { $c->get('http://x.test/c'); } catch (ConnectException $e) {}
        $this->assertSame([
            ['info', 'GET http://x.test/a 200 NULL'],
            ['info', 'GET http://x.test/b 404 NULL'],
            ['error', 'GET http://x.test/c NULL down'],
        ], $logger->records);
    }

    public function testLogMiddlewareStringTemplateAndLevel(): void
    {
        $logger = new GuzzleCov4Logger();
        $s = HandlerStack::create(new MockHandler([new Response(200)]));
        $s->push(Middleware::log($logger, new MessageFormatter('{method}'), LogLevel::DEBUG));
        (new Client(['handler' => $s]))->get('http://x.test/a');
        $this->assertSame([['debug', 'GET']], $logger->records);
    }

    public function testLogMiddlewareBadFormatter(): void
    {
        $this->expectException(\LogicException::class);
        Middleware::log(new GuzzleCov4Logger(), 5);
    }

    public function testHttpErrorsMiddleware(): void
    {
        $s = HandlerStack::create(new MockHandler([new Response(400), new Response(502, [], str_repeat('x', 300))]));
        $s->remove('http_errors');
        $s->push(Middleware::httpErrors(new BodySummarizer(10)), 'he');
        $c = new Client(['handler' => $s]);
        try { $c->get('http://x.test/'); $this->fail('expected'); } catch (ClientException $e) { $this->assertSame(400, $e->getCode()); }
        try { $c->get('http://x.test/'); $this->fail('expected'); } catch (ServerException $e) { $this->assertStringContainsString('xxxxxxxxxx (truncated...)', $e->getMessage()); }
        $m = Middleware::httpErrors()(new MockHandler([new Response(500)]));
        $this->assertSame(500, $m($this->req(), ['http_errors' => false])->wait()->getStatusCode());
    }

    // --- Retry ---
    public function testRetryOnRejection(): void
    {
        $calls = 0;
        $s = HandlerStack::create(new MockHandler([new ConnectException('1', $this->req()), new ConnectException('2', $this->req()), new Response(200, [], 'done')]));
        $s->push(Middleware::retry(function ($n, $req, $res, $err) use (&$calls) { $calls++; return $err !== null && $n < 3; }, fn ($n) => 0));
        $c = new Client(['handler' => $s]);
        $this->assertSame('done', (string) $c->get('http://x.test/')->getBody());
        $this->assertSame(3, $calls);
    }

    public function testRetryGivesUp(): void
    {
        $s = HandlerStack::create(new MockHandler([new ConnectException('1', $this->req())]));
        $s->push(Middleware::retry(fn () => false));
        $this->expectException(ConnectException::class);
        (new Client(['handler' => $s]))->get('http://x.test/');
    }

    public function testRetryDelayOption(): void
    {
        $delays = [];
        $m = new MockHandler([new Response(500), new Response(200)]);
        $rm = new RetryMiddleware(fn ($n, $r, $res) => $res && $res->getStatusCode() >= 500, $m, function ($n) use (&$delays) { $delays[] = $n; return 0; });
        $this->assertSame(200, $rm($this->req(), [])->wait()->getStatusCode());
        $this->assertSame([1], $delays);
        $this->assertSame(0, $m->count());
    }

    // --- PrepareBody ---
    public function testPrepareBodyHeaders(): void
    {
        $seen = null;
        $h = fn (Request $r, $o) => new FulfilledPromise(new Response(200)) ?: null;
        $capture = function (Request $r, array $o) use (&$seen) { $seen = $r; return new FulfilledPromise(new Response(200)); };
        $pb = new PrepareBodyMiddleware($capture);

        $pb(new Request('PUT', 'http://x/', [], 'hello'), [])->wait();
        $this->assertSame('5', $seen->getHeaderLine('Content-Length'));
        $this->assertFalse($seen->hasHeader('Expect'));

        $pb(new Request('PUT', 'http://x/', [], ''), [])->wait();
        $this->assertFalse($seen->hasHeader('Content-Length'));

        $pb(new Request('PUT', 'http://x/', ['Content-Length' => '99'], 'hello'), [])->wait();
        $this->assertSame('99', $seen->getHeaderLine('Content-Length'));

        $pb(new Request('PUT', 'http://x/', [], 'hello'), ['expect' => true])->wait();
        $this->assertSame('100-Continue', $seen->getHeaderLine('Expect'));

        $pb(new Request('PUT', 'http://x/', [], 'hello'), ['expect' => false])->wait();
        $this->assertFalse($seen->hasHeader('Expect'));

        $pb(new Request('PUT', 'http://x/', [], 'hello'), ['expect' => 3])->wait();
        $this->assertSame('100-Continue', $seen->getHeaderLine('Expect'));

        $pb(new Request('PUT', 'http://x/', [], 'hello'), ['expect' => 10])->wait();
        $this->assertFalse($seen->hasHeader('Expect'));

        $pb((new Request('PUT', 'http://x/', [], 'hello'))->withProtocolVersion('1.0'), ['expect' => true])->wait();
        $this->assertFalse($seen->hasHeader('Expect'));

        $pb(new Request('PUT', 'http://x/', ['Expect' => 'custom'], 'hello'), ['expect' => true])->wait();
        $this->assertSame('custom', $seen->getHeaderLine('Expect'));

        $pb(new Request('PUT', 'http://x/', [], str_repeat('a', 1048577)), [])->wait();
        $this->assertSame('100-Continue', $seen->getHeaderLine('Expect'));
    }

    public function testPrepareBodyUnknownSizeAndMime(): void
    {
        $seen = null;
        $pb = new PrepareBodyMiddleware(function (Request $r, array $o) use (&$seen) { $seen = $r; return new FulfilledPromise(new Response(200)); });
        $stream = new \GuzzleHttp\Psr7\PumpStream(function ($n) { static $d = false; if ($d) { return false; } $d = true; return 'abc'; });
        $pb(new Request('POST', 'http://x/', [], $stream), [])->wait();
        $this->assertSame('chunked', $seen->getHeaderLine('Transfer-Encoding'));
        $this->assertFalse($seen->hasHeader('Content-Length'));
        $this->assertSame('100-Continue', $seen->getHeaderLine('Expect'));

        $f = sys_get_temp_dir() . '/guzzlecov4_' . getmypid() . '.json';
        file_put_contents($f, '{"a":1}');
        try {
            $fh = fopen($f, 'r');
            $pb(new Request('POST', 'http://x/', [], new Stream($fh)), [])->wait();
            $this->assertSame('application/json', $seen->getHeaderLine('Content-Type'));
            $this->assertSame('7', $seen->getHeaderLine('Content-Length'));
        } finally {
            unlink($f);
        }
    }

    // --- MockHandler ---
    public function testMockHandlerFeatures(): void
    {
        $got = [];
        $m = new MockHandler([
            new Response(200),
            fn (Request $r, array $o) => new Response(201),
            new RequestException('rx', $this->req()),
            new FulfilledPromise(new Response(202)),
        ], function ($v) use (&$got) { $got[] = 'ok'; }, function ($e) use (&$got) { $got[] = 'err'; });
        $this->assertCount(4, $m);
        $this->assertSame(200, $m($this->req(), [])->wait()->getStatusCode());
        $this->assertSame(201, $m($this->req(), [])->wait()->getStatusCode());
        try { $m($this->req(), [])->wait(); $this->fail('expected'); } catch (RequestException $e) {}
        $this->assertSame(202, $m($this->req(), [])->wait()->getStatusCode());
        $this->assertSame(['ok', 'ok', 'err', 'ok'], $got);
        $this->assertSame('http://x.test/', (string) $m->getLastRequest()->getUri());
        $this->assertSame([], $m->getLastOptions());
        $m->append(new Response(204));
        $m->reset();
        $this->assertCount(0, $m);
        $this->expectException(\OutOfBoundsException::class);
        $m($this->req(), []);
    }

    public function testMockHandlerAppendInvalid(): void
    {
        $m = new MockHandler();
        $this->expectException(\TypeError::class);
        $m->append('not callable or response');
    }

    public function testMockHandlerSinkAndOnHeaders(): void
    {
        $f = sys_get_temp_dir() . '/guzzlecov4sink_' . getmypid();
        try {
            $m = new MockHandler([new Response(200, [], 'to-file'), new Response(200, [], 'to-res'), new Response(200, [], 'to-stream')]);
            $m($this->req(), ['sink' => $f])->wait();
            $this->assertSame('to-file', file_get_contents($f));
            $res = fopen('php://memory', 'w+');
            $m($this->req(), ['sink' => $res])->wait();
            rewind($res);
            $this->assertSame('to-res', stream_get_contents($res));
            $st = PU7::streamFor('');
            $m($this->req(), ['sink' => $st])->wait();
            $this->assertSame('to-stream', (string) $st);
        } finally {
            if (is_file($f)) { unlink($f); }
        }
        $seen = null;
        $m = new MockHandler([new Response(200), new Response(200)]);
        $m($this->req(), ['on_headers' => function ($r) use (&$seen) { $seen = $r->getStatusCode(); }])->wait();
        $this->assertSame(200, $seen);
        try {
            $m($this->req(), ['on_headers' => function ($r) { throw new \Exception('hdr'); }])->wait();
            $this->fail('expected');
        } catch (RequestException $e) {
            $this->assertStringContainsString('on_headers', $e->getMessage());
        }
    }

    public function testMockHandlerOnHeadersNotCallable(): void
    {
        $m = new MockHandler([new Response(200)]);
        $this->expectException(\InvalidArgumentException::class);
        $m($this->req(), ['on_headers' => 'nope-not-a-function']);
    }

    public function testMockHandlerCreateWithMiddleware(): void
    {
        $stack = MockHandler::createWithMiddleware([new Response(201)]);
        $this->assertInstanceOf(HandlerStack::class, $stack);
        $this->assertSame(201, (new Client(['handler' => $stack]))->get('http://x.test/')->getStatusCode());
    }

    // --- MessageFormatter ---
    public function testMessageFormatterAll(): void
    {
        $req = new Request('POST', 'http://u:pw@x.test:81/p?q=1', ['Host' => 'x.test', 'X-Req' => 'rq'], 'req-body');
        $res = new Response(201, ['X-Res' => 'rs'], 'res-body');
        $f = new MessageFormatter('{req_header_X-Req}|{res_header_X-Res}|{req_headers}|{res_headers}|{req_body}|{res_body}|{method}|{target}|{version}|{req_version}|{res_version}|{host}|{code}|{phrase}|{uri}|{url}|{unknown}|{error}');
        $out = $f->format($req, $res, new \Exception('E!'));
        $this->assertStringStartsWith('rq|rs|', $out);
        $this->assertStringContainsString('|req-body|res-body|POST|/p?q=1|1.1|1.1|1.1|x.test|201|Created|http://u:pw@x.test:81/p?q=1|http://u:pw@x.test:81/p?q=1||E!', $out);
        $this->assertStringContainsString("Host: x.test\r\nX-Req: rq", $out);
        $this->assertSame('NULL|NULL|NULL|NULL|NULL|NULL', (new MessageFormatter('{res_header_A}|{res_body}|{code}|{phrase}|{res_version}|{error}'))->format($req));
    }

    public function testMessageFormatterFullMessages(): void
    {
        $req = new Request('GET', 'http://x.test/a');
        $res = new Response(200, [], 'b');
        $out = (new MessageFormatter(MessageFormatter::DEBUG))->format($req, $res);
        $this->assertStringContainsString("GET /a HTTP/1.1\r\n", $out);
        $this->assertStringContainsString('HTTP/1.1 200 OK', $out);
        $this->assertStringEndsWith('NULL', $out);
        $this->assertStringContainsString('"GET /a HTTP/1.1" 200', (new MessageFormatter(MessageFormatter::SHORT))->format($req, $res));
        $this->assertStringContainsString('"GET /a HTTP/1.1" 200', (new MessageFormatter())->format($req, $res));
        $clf = (new MessageFormatter('{ts} {date_iso_8601} {date_common_log} {hostname}'))->format($req, $res);
        $this->assertMatchesRegularExpression('/^\d{4}-\d\d-\d\dT[\d:+]+ \d{4}-\d\d-\d\dT[\d:+]+ /', $clf);
    }

    public function testMessageFormatterUnseekableBodies(): void
    {
        $body = new \GuzzleHttp\Psr7\NoSeekStream(PU7::streamFor('abc'));
        $req = new Request('POST', 'http://x.test/', [], $body);
        $res = new Response(200, [], new \GuzzleHttp\Psr7\NoSeekStream(PU7::streamFor('xyz')));
        $this->assertSame('abc|RESPONSE_NOT_LOGGEABLE', (new MessageFormatter('{req_body}|{res_body}'))->format($req, $res));
    }

    // --- Pool ---
    public function testPoolBatch(): void
    {
        $c = new Client(['handler' => HandlerStack::create(new MockHandler([new Response(200, [], 'a'), new Response(404), new Response(201, [], 'c')])), 'http_errors' => true]);
        $fulfilled = [];
        $res = Pool::batch($c, [new Request('GET', 'http://x/1'), new Request('GET', 'http://x/2'), fn ($opts) => $c->getAsync('http://x/3')], ['concurrency' => 1, 'fulfilled' => function ($r, $k) use (&$fulfilled) { $fulfilled[] = $k; }]);
        $this->assertCount(3, $res);
        $this->assertSame('a', (string) $res[0]->getBody());
        $this->assertInstanceOf(ClientException::class, $res[1]);
        $this->assertSame(201, $res[2]->getStatusCode());
        $this->assertSame([0, 2], $fulfilled);
    }

    public function testPoolRejectedCallbackAndOptions(): void
    {
        $c = new Client(['handler' => HandlerStack::create(new MockHandler([new Response(500)]))]);
        $rej = [];
        $pool = new Pool($c, new ArrayIterator([new Request('GET', 'http://x/1')]), ['options' => ['http_errors' => true], 'rejected' => function ($e, $i) use (&$rej) { $rej[] = get_class($e); }]);
        $pool->promise()->wait();
        $this->assertSame([ServerException::class], $rej);
    }

    public function testPoolInvalidEntry(): void
    {
        $c = new Client(['handler' => HandlerStack::create(new MockHandler([]))]);
        $this->expectException(\InvalidArgumentException::class);
        (new Pool($c, ['not a request']))->promise()->wait();
    }

    public function testPoolNonIterableDeprecated(): void
    {
        $c = new Client(['handler' => HandlerStack::create(new MockHandler([new Response(200)]))]);
        $r = Pool::batch($c, new Request('GET', 'http://x/1'));
        $this->assertCount(1, $r);
    }

    // --- Exceptions ---
    public function testRequestExceptionCreate(): void
    {
        $req = new Request('GET', 'http://user:secret@x.test/p');
        $e = RequestException::create($req);
        $this->assertSame('Error completing request', $e->getMessage());
        $this->assertFalse($e->hasResponse());
        $this->assertNull($e->getResponse());
        $this->assertSame([], $e->getHandlerContext());

        $e = RequestException::create($req, new Response(404, [], 'nf'), new \Exception('prev'), ['k' => 'v']);
        $this->assertInstanceOf(ClientException::class, $e);
        $this->assertStringContainsString('Client error: `GET http://user:***@x.test/p` resulted in a `404 Not Found` response', $e->getMessage());
        $this->assertStringContainsString("nf\n", $e->getMessage());
        $this->assertSame(404, $e->getCode());
        $this->assertTrue($e->hasResponse());
        $this->assertSame(['k' => 'v'], $e->getHandlerContext());
        $this->assertSame('prev', $e->getPrevious()->getMessage());

        $e = RequestException::create($req, new Response(503));
        $this->assertInstanceOf(ServerException::class, $e);
        $this->assertStringStartsWith('Server error:', $e->getMessage());

        $e = RequestException::create($req, new Response(302));
        $this->assertSame(RequestException::class, get_class($e));
        $this->assertStringStartsWith('Unsuccessful request:', $e->getMessage());
        $this->assertStringNotContainsString("\n", $e->getMessage());

        $e = RequestException::create($req, new Response(500, [], str_repeat('z', 500)), null, [], new BodySummarizer(3));
        $this->assertStringContainsString("zzz (truncated...)\n", $e->getMessage());
    }

    public function testRequestExceptionWrap(): void
    {
        $req = $this->req();
        $inner = new \RuntimeException('inner');
        $e = RequestException::wrapException($req, $inner);
        $this->assertSame('inner', $e->getMessage());
        $this->assertSame($e, RequestException::wrapException($req, $e));
    }

    public function testConnectException(): void
    {
        $e = new ConnectException('cannot', $this->req(), new \Exception('p'), ['errno' => 7]);
        $this->assertSame('GET', $e->getRequest()->getMethod());
        $this->assertSame(['errno' => 7], $e->getHandlerContext());
        $this->assertInstanceOf(TransferException::class, $e);
        $this->assertInstanceOf(\Psr\Http\Client\NetworkExceptionInterface::class, $e);
        $this->assertInstanceOf(\GuzzleHttp\Exception\GuzzleException::class, $e);
    }

    public function testBadResponseException(): void
    {
        $e = new BadResponseException('bad', $this->req(), new Response(400));
        $this->assertTrue($e->hasResponse());
        $this->assertSame(400, $e->getResponse()->getStatusCode());
        $t = new TooManyRedirectsException('loop', $this->req(), new Response(302));
        $this->assertInstanceOf(RequestException::class, $t);
    }

    public function testBodySummarizer(): void
    {
        $s = new BodySummarizer();
        $this->assertSame('hello', $s->summarize(new Response(500, [], 'hello')));
        $this->assertNull($s->summarize(new Response(500, [], '')));
        $big = str_repeat('a', 130);
        $this->assertSame(str_repeat('a', 120) . ' (truncated...)', $s->summarize(new Response(500, [], $big)));
        $this->assertNull($s->summarize(new Response(500, [], new \GuzzleHttp\Psr7\NoSeekStream(PU7::streamFor('abc')))));
        $this->assertSame('abc', (new BodySummarizer(3))->summarize(new Response(500, [], 'abc')));
    }

    public function testClientTrait(): void
    {
        $h = [];
        $s = HandlerStack::create(new MockHandler(array_fill(0, 12, new Response(200))));
        $s->push(Middleware::history($h));
        $c = new Client(['handler' => $s]);
        $c->get('http://x/', []);
        $c->head('http://x/');
        $c->put('http://x/');
        $c->post('http://x/');
        $c->patch('http://x/');
        $c->delete('http://x/');
        $c->getAsync('http://x/')->wait();
        $c->headAsync('http://x/')->wait();
        $c->putAsync('http://x/')->wait();
        $c->postAsync('http://x/')->wait();
        $c->patchAsync('http://x/')->wait();
        $c->deleteAsync('http://x/')->wait();
        $this->assertSame(['GET', 'HEAD', 'PUT', 'POST', 'PATCH', 'DELETE', 'GET', 'HEAD', 'PUT', 'POST', 'PATCH', 'DELETE'], array_map(fn ($e) => $e['request']->getMethod(), $h));
    }

    public function testLowercaseMethodNormalised(): void
    {
        $h = [];
        $s = HandlerStack::create(new MockHandler([new Response(200)]));
        $s->push(Middleware::history($h));
        $c = new Client(['handler' => $s]);
        set_error_handler(fn () => true, E_USER_DEPRECATED);
        try { $c->request('get', 'http://x/'); } finally { restore_error_handler(); }
        $this->assertSame('GET', $h[0]['request']->getMethod());
    }
}

function GuzzleCov4Fn(callable $h) { return $h; }

final class GuzzleCov4Obj
{
    public function mw(callable $h) { return $h; }
    public static function smw(callable $h) { return $h; }
}
