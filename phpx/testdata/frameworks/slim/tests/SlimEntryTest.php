<?php
use PHPUnit\Framework\TestCase;
use Psr\Http\Message\ResponseInterface as Response;
use Psr\Http\Message\ServerRequestInterface as Request;
use Slim\Factory\AppFactory;
use Slim\Psr7\Factory\ServerRequestFactory;
use Slim\Routing\RouteCollectorProxy;

final class SlimEntryTest extends TestCase
{
    private function app(): \Slim\App
    {
        $app = AppFactory::create();
        $app->get('/hello/{name}', function (Request $request, Response $response, array $args) {
            $response->getBody()->write('Hello, ' . $args['name']);
            return $response;
        });
        return $app;
    }
    private function get(\Slim\App $app, string $method, string $path): Response
    {
        return $app->handle((new ServerRequestFactory())->createServerRequest($method, $path));
    }
    public function testRoute(): void
    {
        $r = $this->get($this->app(), 'GET', '/hello/world');
        $this->assertSame(200, $r->getStatusCode());
        $this->assertSame('Hello, world', (string) $r->getBody());
    }
    public function testNotFoundAndMethodNotAllowed(): void
    {
        $app = $this->app();
        $app->addRoutingMiddleware();
        $app->addErrorMiddleware(false, false, false);
        $this->assertSame(404, $this->get($app, 'GET', '/nope')->getStatusCode());
        $this->assertSame(405, $this->get($app, 'POST', '/hello/x')->getStatusCode());
    }
    public function testMiddlewareOrder(): void
    {
        $app = $this->app();
        $log = [];
        $app->add(function (Request $req, $handler) use (&$log) { $log[] = 'outer-in'; $res = $handler->handle($req); $log[] = 'outer-out'; return $res; });
        $app->add(function (Request $req, $handler) use (&$log) { $log[] = 'inner-in'; $res = $handler->handle($req); $log[] = 'inner-out'; return $res; });
        $this->get($app, 'GET', '/hello/a');
        $this->assertSame(['inner-in', 'outer-in', 'outer-out', 'inner-out'], $log);
    }
    public function testGroupsAndPlaceholders(): void
    {
        $app = AppFactory::create();
        $app->group('/api', function (RouteCollectorProxy $g) {
            $g->get('/users/{id:[0-9]+}', fn (Request $q, Response $r, array $a) => (function () use ($r, $a) { $r->getBody()->write('user ' . $a['id']); return $r; })());
            $g->post('/users', function (Request $q, Response $r) { $r->getBody()->write('created'); return $r->withStatus(201); });
        });
        $this->assertSame('user 42', (string) $this->get($app, 'GET', '/api/users/42')->getBody());
        $this->assertSame(201, $this->get($app, 'POST', '/api/users')->getStatusCode());
    }
    public function testResponseHeaders(): void
    {
        $app = AppFactory::create();
        $app->get('/h', fn (Request $q, Response $r) => $r->withHeader('X-A', 'b')->withHeader('Content-Type', 'text/plain'));
        $r = $this->get($app, 'GET', '/h');
        $this->assertSame('b', $r->getHeaderLine('X-A'));
        $this->assertSame('text/plain', $r->getHeaderLine('Content-Type'));
    }
}
