<?php
use GuzzleHttp\{Client, HandlerStack, Middleware, RequestOptions};
use GuzzleHttp\Handler\MockHandler;
use GuzzleHttp\Psr7\{Response, Request};
use GuzzleHttp\Exception\{ClientException, RequestException};
use PHPUnit\Framework\TestCase;

final class GuzzleTest extends TestCase
{
    private function client(array $queue, array &$history = []): Client
    {
        $stack = HandlerStack::create(new MockHandler($queue));
        $stack->push(Middleware::history($history));
        return new Client(['handler' => $stack, 'base_uri' => 'http://api.test/']);
    }
    public function testGetWithMock(): void
    {
        $c = $this->client([new Response(200, ['X-A' => '1'], '{"ok":true}')]);
        $r = $c->get('items', ['query' => ['page' => 2]]);
        $this->assertSame(200, $r->getStatusCode());
        $this->assertSame(['ok' => true], json_decode((string) $r->getBody(), true));
        $this->assertSame('1', $r->getHeaderLine('X-A'));
    }
    public function testHistoryRecordsRequests(): void
    {
        $h = [];
        $c = $this->client([new Response(200), new Response(201)], $h);
        $c->post('a', [RequestOptions::JSON => ['k' => 'v']]);
        $c->request('PUT', 'b');
        $this->assertCount(2, $h);
        $this->assertSame('POST', $h[0]['request']->getMethod());
        $this->assertSame('http://api.test/a', (string) $h[0]['request']->getUri());
        $this->assertSame('{"k":"v"}', (string) $h[0]['request']->getBody());
        $this->assertSame('application/json', $h[0]['request']->getHeaderLine('Content-Type'));
    }
    public function testClientErrorThrows(): void
    {
        $c = $this->client([new Response(404, [], 'gone')]);
        $this->expectException(ClientException::class);
        $c->get('missing');
    }
    public function testHttpErrorsOff(): void
    {
        $c = $this->client([new Response(500)]);
        $this->assertSame(500, $c->get('x', ['http_errors' => false])->getStatusCode());
    }
    public function testQueueExhausted(): void
    {
        $c = $this->client([]);
        $this->expectException(OutOfBoundsException::class);
        $c->get('x');
    }
}
