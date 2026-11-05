<?php

use GuzzleHttp\Client;
use GuzzleHttp\Exception\ConnectException;
use GuzzleHttp\Exception\RequestException;
use GuzzleHttp\Handler\CurlHandler;
use GuzzleHttp\Handler\CurlMultiHandler;
use GuzzleHttp\HandlerStack;
use GuzzleHttp\Psr7\Request;
use GuzzleHttp\Psr7\Utils as PU7;
use PHPUnit\Framework\Attributes\DataProvider;

require_once __DIR__.'/GuzzleCurlServer.php';

final class GuzzleCurlTest extends GuzzleCurlBase
{
    private function client(string $kind = 'curl', array $extra = []): Client
    {
        $h = $kind === 'multi' ? new CurlMultiHandler() : new CurlHandler();

        return new Client($extra + ['handler' => HandlerStack::create($h), 'base_uri' => self::url()]);
    }

    public static function kinds(): array
    {
        return [['curl'], ['multi']];
    }

    #[DataProvider('kinds')]
    public function testGetEcho(string $k): void
    {
        $r = $this->client($k)->get('/echo?x=1', ['headers' => ['X-A' => 'b']]);
        $j = json_decode((string) $r->getBody(), true);
        $this->assertSame('GET', $j['method']);
        $this->assertSame('x=1', $j['query']);
        $this->assertSame('b', $j['headers']['X-A']);
        $this->assertSame('application/json', $r->getHeaderLine('Content-Type'));
    }

    #[DataProvider('kinds')]
    public function testPostBodies(string $k): void
    {
        $c = $this->client($k);
        $j = json_decode((string) $c->post('/echo', ['body' => 'raw'])->getBody(), true);
        $this->assertSame('raw', $j['body']);
        $this->assertSame('POST', $j['method']);

        $j = json_decode((string) $c->post('/echo', ['json' => ['a' => 1]])->getBody(), true);
        $this->assertSame('{"a":1}', $j['body']);
        $this->assertSame('application/json', $j['headers']['Content-Type']);

        $j = json_decode((string) $c->post('/echo', ['form_params' => ['a' => 'b c', 'd' => ['e']]])->getBody(), true);
        $this->assertSame('b c', $j['post']['a']);
        $this->assertSame(['e'], $j['post']['d']);

        $j = json_decode((string) $c->post('/echo', ['body' => PU7::streamFor('streamed body')])->getBody(), true);
        $this->assertSame('streamed body', $j['body']);

        $fh = fopen('php://temp', 'w+');
        fwrite($fh, 'resource body');
        rewind($fh);
        $j = json_decode((string) $c->post('/echo', ['body' => $fh])->getBody(), true);
        $this->assertSame('resource body', $j['body']);
    }

    #[DataProvider('kinds')]
    public function testMultipart(string $k): void
    {
        $j = json_decode((string) $this->client($k)->post('/echo', ['multipart' => [
            ['name' => 'field', 'contents' => 'value'],
            ['name' => 'file', 'contents' => 'file data', 'filename' => 'f.txt'],
        ]])->getBody(), true);
        $this->assertSame('value', $j['post']['field']);
        $this->assertSame(['file' => 'f.txt'], $j['files']);
    }

    #[DataProvider('kinds')]
    public function testMethods(string $k): void
    {
        $c = $this->client($k);
        foreach (['PUT', 'PATCH', 'DELETE', 'OPTIONS'] as $m) {
            $j = json_decode((string) $c->request($m, '/echo', ['body' => 'x'])->getBody(), true);
            $this->assertSame($m, $j['method']);
        }
        $r = $c->head('/head');
        $this->assertSame('1', $r->getHeaderLine('X-Head'));
        $this->assertSame('', (string) $r->getBody());
        $this->assertSame('hello', (string) $c->get('/head')->getBody());
        $this->assertSame(204, $c->get('/empty')->getStatusCode());
    }

    #[DataProvider('kinds')]
    public function testQueryAndAuth(string $k): void
    {
        $c = $this->client($k);
        $j = json_decode((string) $c->get('/echo', ['query' => ['a' => 'b', 'c' => 'd e']])->getBody(), true);
        $this->assertSame('a=b&c=d%20e', $j['query']);
        $this->assertSame('u:p', (string) $c->get('/auth', ['auth' => ['u', 'p']])->getBody());
        $this->assertSame('u:p', (string) $c->get('/auth', ['auth' => ['u', 'p', 'basic']])->getBody());
        $r = $c->get('/auth', ['auth' => ['u', 'p', 'digest'], 'http_errors' => false]);
        $this->assertSame(401, $r->getStatusCode());
        $r = $c->get('/auth', ['auth' => ['u', 'p', 'ntlm'], 'http_errors' => false]);
        $this->assertSame(401, $r->getStatusCode());
        $this->assertSame(401, $c->get('/auth', ['http_errors' => false])->getStatusCode());
    }

    #[DataProvider('kinds')]
    public function testStatusErrors(string $k): void
    {
        $c = $this->client($k);
        $this->assertSame(200, $c->get('/status?code=200')->getStatusCode());
        $this->assertSame(500, $c->get('/status?code=500', ['http_errors' => false])->getStatusCode());
        try {
            $c->get('/status?code=404');
            $this->fail('no exception');
        } catch (RequestException $e) {
            $this->assertSame(404, $e->getResponse()->getStatusCode());
        }
        try {
            $c->get('/status?code=503');
            $this->fail('no exception');
        } catch (\GuzzleHttp\Exception\ServerException $e) {
            $this->assertSame(503, $e->getResponse()->getStatusCode());
        }
    }

    #[DataProvider('kinds')]
    public function testRedirects(string $k): void
    {
        $c = $this->client($k);
        foreach ([301, 302, 303, 307, 308] as $code) {
            $r = $c->post("/redirect?code=$code&n=2", ['body' => 'payload']);
            $j = json_decode((string) $r->getBody(), true);
            $this->assertSame('/echo', $j['path']);
            $this->assertSame(in_array($code, [301, 302, 303], true) ? 'GET' : 'POST', $j['method']);
        }
        $r = $c->get('/redirect?n=3', ['allow_redirects' => ['track_redirects' => true, 'referer' => true]]);
        $this->assertCount(3, $r->getHeader('X-Guzzle-Redirect-History'));
        $this->assertSame(200, $r->getStatusCode());
        $j = json_decode((string) $r->getBody(), true);
        $this->assertStringContainsString('/redirect', $j['headers']['Referer']);

        $r = $c->post('/redirect?code=302', ['body' => 'p', 'allow_redirects' => ['strict' => true]]);
        $this->assertSame('POST', json_decode((string) $r->getBody(), true)['method']);

        $r = $c->get('/redirect', ['allow_redirects' => false]);
        $this->assertSame(302, $r->getStatusCode());
        $this->assertSame('/echo', $r->getHeaderLine('Location'));

        try {
            $c->get('/loop', ['allow_redirects' => ['max' => 3]]);
            $this->fail('no exception');
        } catch (\GuzzleHttp\Exception\TooManyRedirectsException $e) {
            $this->assertStringContainsString('Will not follow more than 3 redirects', $e->getMessage());
        }
        try {
            $c->get('/redirect?to=ftp://127.0.0.1/x', ['allow_redirects' => ['protocols' => ['http']]]);
            $this->fail('no exception');
        } catch (RequestException $e) {
            $this->assertStringContainsString('allowed redirect protocols', $e->getMessage());
        }
        $seen = [];
        $c->get('/redirect?n=2', ['allow_redirects' => ['on_redirect' => function ($rq, $rs, $uri) use (&$seen) {
            $seen[] = (string) $uri;
        }]]);
        $this->assertNotEmpty($seen);
    }

    #[DataProvider('kinds')]
    public function testContentEncodings(string $k): void
    {
        $c = $this->client($k);
        $this->assertSame('zipped body', (string) $c->get('/gzip')->getBody());
        $this->assertSame('zipped body', (string) $c->get('/gzip', ['decode_content' => true])->getBody());
        $this->assertSame('zipped body', (string) $c->get('/gzip', ['decode_content' => 'gzip'])->getBody());
        $r = $c->get('/gzip', ['decode_content' => false]);
        $this->assertSame('zipped body', gzdecode((string) $r->getBody()));
        $this->assertSame('gzip', $r->getHeaderLine('Content-Encoding'));
        $this->assertSame('deflated body', (string) $c->get('/deflate', ['decode_content' => true])->getBody());
        $j = json_decode((string) $c->get('/echo', ['decode_content' => 'gzip'])->getBody(), true);
        $this->assertSame('gzip', $j['headers']['Accept-Encoding']);
    }

    #[DataProvider('kinds')]
    public function testBigAndChunked(string $k): void
    {
        $c = $this->client($k);
        $this->assertSame(300000 - 300000 % 16, strlen((string) $c->get('/large')->getBody()));
        $this->assertSame("chunk0\nchunk1\nchunk2\nchunk3\nchunk4\n", (string) $c->get('/chunked')->getBody());
        $j = json_decode((string) $c->post('/echo', ['body' => str_repeat('x', 200000)])->getBody(), true);
        $this->assertSame(200000, strlen($j['body']));
    }

    #[DataProvider('kinds')]
    public function testHeadersRepeatedAndCookies(string $k): void
    {
        $c = $this->client($k);
        $r = $c->get('/multi');
        $this->assertSame(['one', 'two', 'three'], $r->getHeader('X-Dup'));

        $jar = new \GuzzleHttp\Cookie\CookieJar();
        $c->get('/cookie', ['cookies' => $jar]);
        $this->assertCount(2, $jar);
        $j = json_decode((string) $c->get('/echo', ['cookies' => $jar])->getBody(), true);
        $this->assertSame(['a' => '1', 'b' => '2'], $j['cookies']);
    }

    #[DataProvider('kinds')]
    public function testSinkAndStream(string $k): void
    {
        $c = $this->client($k);
        $path = __DIR__.'/gcurl-sink-'.getmypid();
        $c->get('/text?t=sinkme', ['sink' => $path]);
        $this->assertSame('sinkme', file_get_contents($path));
        unlink($path);

        $fh = fopen($path, 'w+');
        $c->get('/text?t=toresource', ['sink' => $fh]);
        $this->assertSame('toresource', file_get_contents($path));
        unlink($path);

        $s = PU7::streamFor(fopen('php://temp', 'w+'));
        $c->get('/text?t=tostream', ['sink' => $s]);
        $this->assertSame('tostream', (string) $s);

        $r = $c->get('/text?t=streamed', ['stream' => true]);
        $this->assertSame('streamed', (string) $r->getBody());
    }

    #[DataProvider('kinds')]
    public function testCallbacks(string $k): void
    {
        $c = $this->client($k);
        $stats = null;
        $headers = null;
        $progress = 0;
        $c->get('/large', [
            'on_stats' => function (\GuzzleHttp\TransferStats $s) use (&$stats) {
                $stats = $s;
            },
            'on_headers' => function ($r) use (&$headers) {
                $headers = $r->getStatusCode();
            },
            'progress' => function () use (&$progress) {
                $progress++;
            },
        ]);
        $this->assertInstanceOf(\GuzzleHttp\TransferStats::class, $stats);
        $this->assertSame(200, $stats->getResponse()->getStatusCode());
        $this->assertSame(200, $headers);
        $this->assertGreaterThan(0, $progress);
        $this->assertArrayHasKey('total_time', $stats->getHandlerStats());

        try {
            $c->get('/text', ['on_headers' => function () {
                throw new \RuntimeException('abort');
            }]);
            $this->fail('no exception');
        } catch (RequestException $e) {
            $this->assertStringContainsString('An error was encountered during the on_headers', $e->getMessage());
        }
    }

    #[DataProvider('kinds')]
    public function testExpectAndVersions(string $k): void
    {
        $c = $this->client($k);
        $j = json_decode((string) $c->post('/echo', ['body' => str_repeat('x', 2000), 'expect' => true])->getBody(), true);
        $this->assertSame(2000, strlen($j['body']));
        $j = json_decode((string) $c->post('/echo', ['body' => 'x', 'expect' => false])->getBody(), true);
        $this->assertArrayNotHasKey('Expect', $j['headers']);
        $j = json_decode((string) $c->post('/echo', ['body' => str_repeat('x', 5000), 'expect' => 100])->getBody(), true);
        $this->assertSame(5000, strlen($j['body']));
        foreach (['1.0', '1.1', '2', '2.0'] as $v) {
            $r = $c->get('/echo', ['version' => $v]);
            $this->assertSame(200, $r->getStatusCode());
        }
        $r = $c->get('/echo', ['force_ip_resolve' => 'v4']);
        $this->assertSame(200, $r->getStatusCode());
        try {
            $c->get('/echo', ['force_ip_resolve' => 'v6', 'connect_timeout' => 1]);
            $this->fail('no exception');
        } catch (ConnectException $e) {
            $this->assertStringContainsString('cURL error 7', $e->getMessage());
        }
    }

    #[DataProvider('kinds')]
    public function testRawCurlOptions(string $k): void
    {
        $c = $this->client($k);
        $j = json_decode((string) $c->get('/echo', ['curl' => [CURLOPT_COOKIE => 'k=v']])->getBody(), true);
        $this->assertSame(['k' => 'v'], $j['cookies']);
        $j = json_decode((string) $c->get('/echo', ['curl' => [CURLOPT_HTTPHEADER => ['X-Raw: 1']]])->getBody(), true);
        $this->assertSame('1', $j['headers']['X-Raw']);
    }

    #[DataProvider('kinds')]
    public function testConnectionRefused(string $k): void
    {
        $c = $this->client($k);
        try {
            $c->get('http://127.0.0.1:1/', ['connect_timeout' => 2]);
            $this->fail('no exception');
        } catch (ConnectException $e) {
            $this->assertNotNull($e->getHandlerContext());
        }
        try {
            $c->get('ftp://127.0.0.1/x');
            $this->fail('no exception');
        } catch (\Throwable $e) {
            $this->assertNotNull($e->getMessage());
        }
    }

    #[DataProvider('kinds')]
    public function testTimeout(string $k): void
    {
        $c = $this->client($k);
        try {
            $c->get('/slow?us=1500000', ['timeout' => 0.3]);
            $this->fail('no exception');
        } catch (ConnectException $e) {
            $this->assertStringContainsString('cURL error 28', $e->getMessage());
        }
        try {
            $c->get('/slow?us=1500000', ['read_timeout' => 0.3]);
        } catch (\Throwable $e) {
            $this->assertNotNull($e);
        }
        $this->assertSame('slow', (string) $c->get('/slow?us=100000', ['timeout' => 5])->getBody());
    }

    #[DataProvider('kinds')]
    public function testPrematureClose(string $k): void
    {
        $c = $this->client($k);
        try {
            $c->get('/close');
            $this->fail('no exception');
        } catch (RequestException $e) {
            $this->assertNotEmpty($e->getMessage());
        }
    }

    public function testAsyncConcurrent(): void
    {
        $c = $this->client('multi');
        $ps = [];
        foreach ([1, 2, 3, 4] as $i) {
            $ps[$i] = $c->getAsync("/text?t=r$i");
        }
        $res = \GuzzleHttp\Promise\Utils::unwrap($ps);
        foreach ($res as $i => $r) {
            $this->assertSame("r$i", (string) $r->getBody());
        }
    }
}
