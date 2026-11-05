<?php
use PHPUnit\Framework\TestCase;
use Symfony\Component\Routing\Route;
use Symfony\Component\Routing\RouteCollection;
use Symfony\Component\Routing\RequestContext;
use Symfony\Component\Routing\Matcher\UrlMatcher;
use Symfony\Component\Routing\Generator\UrlGenerator;
use Symfony\Component\Routing\Exception\ResourceNotFoundException;

final class RoutingEntryTest extends TestCase
{
    private function routes(): RouteCollection
    {
        $r = new RouteCollection();
        $r->add('home', new Route('/'));
        $r->add('post', new Route('/post/{id}', ['_controller' => 'show'], ['id' => '\d+']));
        $r->add('lang', new Route('/{lang}/about', [], ['lang' => 'en|fr']));
        return $r;
    }
    public function testMatch(): void
    {
        $m = new UrlMatcher($this->routes(), new RequestContext());
        $this->assertSame('home', $m->match('/')['_route']);
        $p = $m->match('/post/42');
        $this->assertSame(['_controller' => 'show', 'id' => '42', '_route' => 'post'], $p);
        $this->assertSame('fr', $m->match('/fr/about')['lang']);
    }
    public function testNoMatch(): void
    {
        $m = new UrlMatcher($this->routes(), new RequestContext());
        $this->expectException(ResourceNotFoundException::class);
        $m->match('/post/abc');
    }
    public function testGenerate(): void
    {
        $g = new UrlGenerator($this->routes(), new RequestContext());
        $this->assertSame('/post/7', $g->generate('post', ['id' => 7]));
        $this->assertSame('/post/7?x=1', $g->generate('post', ['id' => 7, 'x' => 1]));
    }
}
