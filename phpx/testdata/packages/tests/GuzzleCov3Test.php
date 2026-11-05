<?php
use GuzzleHttp\{Client, HandlerStack, Middleware, MessageFormatter, Pool, BodySummarizer, RetryMiddleware, PrepareBodyMiddleware, RequestOptions};
use GuzzleHttp\Handler\MockHandler;
use GuzzleHttp\Cookie\{CookieJar, SetCookie, FileCookieJar, SessionCookieJar};
use GuzzleHttp\Exception\{RequestException, ConnectException, ClientException, ServerException, BadResponseException, TransferException, TooManyRedirectsException};
use GuzzleHttp\Promise\{FulfilledPromise, RejectedPromise};
use GuzzleHttp\Psr7\{Request, Response, Utils as PU7};
use PHPUnit\Framework\TestCase;

final class GuzzleCov3Test extends TestCase
{
    private array $files = [];

    protected function tearDown(): void
    {
        foreach ($this->files as $f) {
            if (is_file($f)) {
                unlink($f);
            }
        }
        $this->files = [];
        unset($_SESSION['guzzlecov']);
    }

    private function tmp(string $suffix): string
    {
        $f = sys_get_temp_dir() . '/guzzlecov_' . getmypid() . '_' . $suffix;
        $this->files[] = $f;
        return $f;
    }

    // --- SetCookie ---
    public function testSetCookieFromString(): void
    {
        $c = SetCookie::fromString('sid=abc; Domain=.Example.com; Path=/a; Max-Age=100; Secure; HttpOnly; Discard; Custom=1; HostOnly=1');
        $this->assertSame('sid', $c->getName());
        $this->assertSame('abc', $c->getValue());
        $this->assertSame('.Example.com', $c->getDomain());
        $this->assertSame('/a', $c->getPath());
        $this->assertSame(100, $c->getMaxAge());
        $this->assertTrue($c->getSecure());
        $this->assertTrue($c->getHttpOnly());
        $this->assertTrue($c->getDiscard());
        $this->assertFalse($c->getHostOnly());
        $this->assertSame('1', $c->toArray()['Custom']);
        $this->assertIsInt($c->getExpires());
        $this->assertStringContainsString('Max-Age=100', (string) $c);
        $this->assertStringContainsString('Expires=', (string) $c);
    }

    public function testSetCookieFromStringInvalid(): void
    {
        $this->assertNull(SetCookie::fromString('')->getName());
        $this->assertNull(SetCookie::fromString('novalue')->getName());
        $c = SetCookie::fromString('a=b; Max-Age=notnum');
        $this->assertNull($c->getMaxAge());
    }

    public function testSetCookieSetters(): void
    {
        $c = new SetCookie(['Name' => 'n', 'Value' => 'v', 'Expires' => 'Wed, 01 Jan 2031 00:00:00 GMT', 'HostOnly' => true, 'Domain' => 'a.test']);
        $this->assertSame(1924992000, $c->getExpires());
        $this->assertTrue($c->getHostOnly());
        $this->assertFalse($c->isExpired());
        $this->assertStringNotContainsString('Domain=', (string) $c);
        $this->assertTrue($c->toArray()['HostOnly']);
        $c->setExpires(1);
        $this->assertTrue($c->isExpired());
        $c->setMaxAge(0);
        $this->assertSame(0, $c->getMaxAge());
        $c->setSecure(true);
        $c->setDiscard(true);
        $c->setHttpOnly(true);
        $c->setPath('/p');
        $c->setDomain('b.test');
        $c->setName('nn');
        $c->setValue('vv');
        $c->setHostOnly(false);
        $this->assertSame('nn=vv; Domain=b.test; Path=/p; Max-Age=0; Expires=Thu, 01 Jan 1970 00:00:01 GMT; Secure; Discard; HttpOnly', (string) $c);
    }

    public function testSetCookieHostOnlyMustBeBool(): void
    {
        $this->expectException(\InvalidArgumentException::class);
        new SetCookie(['HostOnly' => 'yes']);
    }

    public function testSetCookieMaxAgeLarge(): void
    {
        $c = new SetCookie(['Name' => 'a', 'Value' => 'b', 'Max-Age' => PHP_INT_MAX]);
        $this->assertSame(PHP_INT_MAX, $c->getExpires());
        $c = new SetCookie(['Name' => 'a', 'Value' => 'b', 'Max-Age' => -5]);
        $this->assertLessThan(time(), $c->getExpires());
    }

    public function testSetCookieMatchesPath(): void
    {
        $c = new SetCookie(['Name' => 'a', 'Value' => 'b', 'Path' => '/foo']);
        $this->assertTrue($c->matchesPath('/foo'));
        $this->assertTrue($c->matchesPath('/foo/bar'));
        $this->assertFalse($c->matchesPath('/foobar'));
        $this->assertFalse($c->matchesPath('/other'));
        $d = new SetCookie(['Name' => 'a', 'Value' => 'b', 'Path' => '/foo/']);
        $this->assertTrue($d->matchesPath('/foo/x'));
        $this->assertTrue((new SetCookie(['Name' => 'a', 'Path' => '/']))->matchesPath('/anything'));
    }

    public function testSetCookieMatchesDomainExact(): void
    {
        $mk = fn (string $d, bool $host = false) => new SetCookie(['Name' => 'a', 'Value' => 'b', 'Domain' => $d, 'HostOnly' => $host]);
        $this->assertTrue($mk('example.com')->matchesDomain('example.com'));
        $this->assertFalse($mk('example.com', true)->matchesDomain('www.example.com'));
        $this->assertTrue($mk('example.com', true)->matchesDomain('EXAMPLE.com'));
        $this->assertFalse($mk('.')->matchesDomain('example.com'));
        $this->assertFalse($mk('exa%6Dple.com')->matchesDomain('www.exa%6Dple.com'));
        $nodomain = new SetCookie(['Name' => 'a', 'Value' => 'b']);
        $this->assertTrue($nodomain->matchesDomain('x.com'));
        $nodomain->setHostOnly(true);
        $this->assertFalse($nodomain->matchesDomain('x.com'));
    }

    public function testSetCookieMatchesDomainSubdomains(): void
    {
        $mk = fn (string $d, bool $host = false) => new SetCookie(['Name' => 'a', 'Value' => 'b', 'Domain' => $d, 'HostOnly' => $host]);
        $this->assertTrue($mk('example.com')->matchesDomain('www.example.com'));
        $this->assertTrue($mk('.Example.com')->matchesDomain('WWW.example.com'));
        $this->assertFalse($mk('example.com')->matchesDomain('badexample.com'));
    }

    public function testSetCookieMatchesDomainNumeric(): void
    {
        $mk = fn (string $d, bool $host = false) => new SetCookie(['Name' => 'a', 'Value' => 'b', 'Domain' => $d, 'HostOnly' => $host]);
        $this->assertTrue($mk('10.0.0.1')->matchesDomain('10.0.0.1'));
        $this->assertFalse($mk('0.1')->matchesDomain('10.0.0.1'));
        $this->assertFalse($mk('example.123')->matchesDomain('www.example.123'));
        $this->assertFalse($mk('1.2.3')->matchesDomain('x.1.2.3'));
        $this->assertFalse($mk('ex.com')->matchesDomain('10.0.0.1'));
    }

    public function testSetCookieValidate(): void
    {
        $this->assertSame('The cookie name must not be empty', (new SetCookie(['Name' => '', 'Value' => 'v', 'Domain' => 'a']))->validate());
        $this->assertStringContainsString('invalid characters', (new SetCookie(['Name' => 'a b', 'Value' => 'v', 'Domain' => 'a']))->validate());
        $this->assertSame('The cookie value must not be empty', (new SetCookie(['Name' => 'ab', 'Domain' => 'a']))->validate());
        $this->assertSame('The cookie domain must not be empty', (new SetCookie(['Name' => 'ab', 'Value' => '1', 'Domain' => '..']))->validate());
        $this->assertSame('The cookie domain must not be empty', (new SetCookie(['Name' => 'ab', 'Value' => '1', 'Domain' => '']))->validate());
        $this->assertTrue((new SetCookie(['Name' => 'ab', 'Value' => '0', 'Domain' => 'a.com']))->validate());
        $this->assertTrue((new SetCookie(['Name' => 'ab', 'Value' => '0']))->validate());
    }

    // --- CookieJar ---
    private function jar(): CookieJar
    {
        $j = new CookieJar();
        $j->setCookie(new SetCookie(['Name' => 'a', 'Value' => '1', 'Domain' => 'x.test', 'Path' => '/', 'Expires' => time() + 1000]));
        $j->setCookie(new SetCookie(['Name' => 'b', 'Value' => '2', 'Domain' => 'x.test', 'Path' => '/sub', 'Discard' => true]));
        $j->setCookie(new SetCookie(['Name' => 'c', 'Value' => '3', 'Domain' => 'y.test', 'Path' => '/', 'Secure' => true]));
        return $j;
    }

    public function testCookieJarBasics(): void
    {
        $j = $this->jar();
        $this->assertCount(3, $j);
        $this->assertSame('2', $j->getCookieByName('B')->getValue());
        $this->assertNull($j->getCookieByName('zzz'));
        $this->assertCount(3, $j->toArray());
        $this->assertSame(['a', 'b', 'c'], array_map(fn ($c) => $c->getName(), iterator_to_array($j)));
        $this->assertFalse($j->setCookie(new SetCookie(['Name' => '', 'Value' => 'x', 'Domain' => 'q'])));
        $this->assertFalse($j->setCookie(new SetCookie(['Name' => 'a', 'Value' => '1', 'Domain' => 'x.test', 'Path' => '/', 'Expires' => $j->getCookieByName('a')->getExpires()])));
        $this->assertTrue($j->setCookie(new SetCookie(['Name' => 'a', 'Value' => 'changed', 'Domain' => 'x.test', 'Path' => '/', 'Expires' => time() + 1000])));
        $this->assertSame('changed', $j->getCookieByName('a')->getValue());
        $this->assertCount(3, $j);
        $this->assertTrue($j->setCookie(new SetCookie(['Name' => '0', 'Value' => 'zero', 'Domain' => 'x.test'])));
    }

    public function testCookieJarStrict(): void
    {
        $j = new CookieJar(true);
        $this->expectException(\RuntimeException::class);
        $j->setCookie(new SetCookie(['Name' => 'a b', 'Value' => '1', 'Domain' => 'x']));
    }

    public function testCookieJarConstructorArray(): void
    {
        $j = new CookieJar(false, [new SetCookie(['Name' => 'a', 'Value' => '1', 'Domain' => 'x.test']), ['Name' => 'b', 'Value' => '2', 'Domain' => 'x.test']]);
        $this->assertCount(2, $j);
    }

    public function testCookieJarFromArray(): void
    {
        $j = CookieJar::fromArray(['k' => 'v', 'k2' => 'v2'], 'z.test');
        $this->assertCount(2, $j);
        $this->assertTrue($j->getCookieByName('k')->getDiscard());
        $req = $j->withCookieHeader(new Request('GET', 'http://z.test/'));
        $this->assertSame('k=v; k2=v2', $req->getHeaderLine('Cookie'));
    }

    public function testCookieJarClear(): void
    {
        $j = $this->jar();
        $j->clear('x.test', '/sub', 'b');
        $this->assertCount(2, $j);
        $j = $this->jar();
        $j->clear('x.test', '/sub');
        $this->assertCount(1, $j);
        $j = $this->jar();
        $j->clear('x.test');
        $this->assertCount(1, $j);
        $j->clear();
        $this->assertCount(0, $j);
    }

    public function testCookieJarClearSession(): void
    {
        $j = $this->jar();
        $j->clearSessionCookies();
        $this->assertSame(['a'], array_map(fn ($c) => $c->getName(), array_values(iterator_to_array($j))));
    }

    public function testShouldPersist(): void
    {
        $this->assertTrue(CookieJar::shouldPersist(new SetCookie(['Name' => 'a', 'Value' => 'b', 'Expires' => time() + 99])));
        $this->assertFalse(CookieJar::shouldPersist(new SetCookie(['Name' => 'a', 'Value' => 'b'])));
        $this->assertTrue(CookieJar::shouldPersist(new SetCookie(['Name' => 'a', 'Value' => 'b']), true));
        $this->assertFalse(CookieJar::shouldPersist(new SetCookie(['Name' => 'a', 'Value' => 'b', 'Discard' => true]), true));
    }

    public function testExtractCookies(): void
    {
        $j = new CookieJar();
        $req = new Request('GET', 'http://www.x.test/dir/page');
        $res = new Response(200, ['Set-Cookie' => [
            'a=1',
            'b=2; Domain=x.test; Path=/',
            'c=3; Domain=other.test',
            'd=4; Path=rel',
            'e=5; Domain=www.x.test.',
            'f=6; Max-Age=0',
            str_repeat('g', 9000) . '=1',
        ]]);
        $j->extractCookies($req, $res);
        $names = array_map(fn ($c) => $c->getName(), iterator_to_array($j));
        $this->assertSame(['a', 'b', 'd', 'e'], $names);
        $this->assertTrue($j->getCookieByName('a')->getHostOnly());
        $this->assertSame('/dir', $j->getCookieByName('a')->getPath() === '/' ? '/dir' : $j->getCookieByName('d')->getPath());
        $this->assertFalse($j->getCookieByName('b')->getHostOnly());
        $this->assertSame('/dir', $j->getCookieByName('d')->getPath());
    }

    public function testExtractCookiesRemovesExpired(): void
    {
        $j = new CookieJar();
        $req = new Request('GET', 'http://x.test/');
        $j->extractCookies($req, new Response(200, ['Set-Cookie' => 'a=1; Domain=x.test']));
        $this->assertCount(1, $j);
        $j->extractCookies($req, new Response(200, ['Set-Cookie' => 'a=1; Domain=x.test; Max-Age=0']));
        $this->assertCount(0, $j);
        $j->extractCookies($req, new Response(200, ['Set-Cookie' => 'a=1; Domain=x.test']));
        $j->extractCookies($req, new Response(200, ['Set-Cookie' => 'a=; Domain=x.test']));
        $this->assertCount(1, $j);
        $this->assertSame('', $j->getCookieByName('a')->getValue());
    }

    public function testExtractCookiesPathVariants(): void
    {
        $j = new CookieJar();
        $j->extractCookies(new Request('GET', 'http://x.test'), new Response(200, ['Set-Cookie' => 'a=1; Path=nope']));
        $j->extractCookies(new Request('GET', 'http://x.test/'), new Response(200, ['Set-Cookie' => 'b=1; Path=nope']));
        $j->extractCookies(new Request('GET', 'http://x.test/one'), new Response(200, ['Set-Cookie' => 'c=1; Path=nope']));
        $j->extractCookies(new Request('GET', 'http://x.test/one/two/'), new Response(200, ['Set-Cookie' => 'd=1; Path=nope']));
        $this->assertSame('/', $j->getCookieByName('a')->getPath());
        $this->assertSame('/', $j->getCookieByName('b')->getPath());
        $this->assertSame('/', $j->getCookieByName('c')->getPath());
        $this->assertSame('/one/two', $j->getCookieByName('d')->getPath());
    }

    public function testWithCookieHeader(): void
    {
        $j = $this->jar();
        $r = $j->withCookieHeader(new Request('GET', 'http://x.test/sub/page'));
        $this->assertSame('a=1; b=2', $r->getHeaderLine('Cookie'));
        $this->assertFalse($j->withCookieHeader(new Request('GET', 'http://y.test/'))->hasHeader('Cookie'));
        $this->assertSame('c=3', $j->withCookieHeader(new Request('GET', 'https://y.test/'))->getHeaderLine('Cookie'));
        $this->assertFalse($j->withCookieHeader(new Request('GET', 'http://none.test/'))->hasHeader('Cookie'));
    }

    public function testWithCookieHeaderLimits(): void
    {
        $j = new CookieJar();
        for ($i = 0; $i < 200; $i++) {
            $j->setCookie(new SetCookie(['Name' => 'n' . $i, 'Value' => str_repeat('v', 100), 'Domain' => 'x.test']));
        }
        $h = $j->withCookieHeader(new Request('GET', 'http://x.test/'))->getHeaderLine('Cookie');
        $this->assertLessThanOrEqual(8192, strlen($h) + 8);
        $this->assertGreaterThan(10, substr_count($h, ';'));
        $j2 = new CookieJar();
        for ($i = 0; $i < 300; $i++) {
            $j2->setCookie(new SetCookie(['Name' => 'n' . $i, 'Value' => '1', 'Domain' => 'x.test']));
        }
        $this->assertLessThanOrEqual(300, substr_count($j2->withCookieHeader(new Request('GET', 'http://x.test/'))->getHeaderLine('Cookie'), ';') + 1);
    }

    public function testCookiesMiddlewareRoundTrip(): void
    {
        $jar = new CookieJar();
        $h = [];
        $stack = HandlerStack::create(new MockHandler([
            new Response(200, ['Set-Cookie' => 'sid=xyz; Path=/']),
            new Response(200),
        ]));
        $stack->push(Middleware::history($h));
        $c = new Client(['handler' => $stack, 'cookies' => $jar, 'base_uri' => 'http://site.test/']);
        $c->get('login');
        $c->get('next');
        $this->assertSame('sid=xyz', $h[1]['request']->getHeaderLine('Cookie'));
        $c2 = new Client(['handler' => HandlerStack::create(new MockHandler([new Response(200)])), 'cookies' => true]);
        $c2->get('http://site.test/');
        $this->assertInstanceOf(CookieJar::class, $c2->getConfig('cookies'));
    }

    public function testFileCookieJar(): void
    {
        $f = $this->tmp('jar.json');
        $j = new FileCookieJar($f);
        $j->setCookie(new SetCookie(['Name' => 'keep', 'Value' => '1', 'Domain' => 'x.test', 'Expires' => time() + 1000]));
        $j->setCookie(new SetCookie(['Name' => 'sess', 'Value' => '2', 'Domain' => 'x.test']));
        $j->save($f);
        $data = json_decode(file_get_contents($f), true);
        $this->assertCount(1, $data);
        $this->assertSame('keep', $data[0]['Name']);
        $j2 = new FileCookieJar($f);
        $this->assertCount(1, $j2);
        $this->assertSame('1', $j2->getCookieByName('keep')->getValue());
        $j3 = new FileCookieJar($f, true);
        $j3->setCookie(new SetCookie(['Name' => 'sess', 'Value' => '2', 'Domain' => 'x.test']));
        $j3->save($f);
        $this->assertCount(2, json_decode(file_get_contents($f), true));
        unset($j, $j2, $j3);
    }

    public function testFileCookieJarLoadErrors(): void
    {
        $f = $this->tmp('bad.json');
        $j = new FileCookieJar($f);
        file_put_contents($f, '');
        $j->load($f);
        $this->assertCount(0, $j);
        file_put_contents($f, '{not json');
        try { $j->load($f); $this->fail('expected'); } catch (\InvalidArgumentException $e) { $this->assertStringContainsString('json_decode error', $e->getMessage()); }
        file_put_contents($f, '[{"Name":"a"}]');
        try { $j->load($f); $this->fail('expected'); } catch (\RuntimeException $e) { $this->assertStringContainsString('Invalid cookie file', $e->getMessage()); }
        file_put_contents($f, '"scalar"');
        try { $j->load($f); $this->fail('expected'); } catch (\RuntimeException $e) { $this->assertStringContainsString('Invalid cookie file', $e->getMessage()); }
        file_put_contents($f, '[]');
        $j->load($f);
        unset($j);
    }

    public function testSessionCookieJar(): void
    {
        $j = new SessionCookieJar('guzzlecov');
        $j->setCookie(new SetCookie(['Name' => 'a', 'Value' => '1', 'Domain' => 'x.test', 'Expires' => time() + 1000, 'HostOnly' => false]));
        $j->save();
        $this->assertIsString($_SESSION['guzzlecov']);
        $j2 = new SessionCookieJar('guzzlecov');
        $this->assertSame('1', $j2->getCookieByName('a')->getValue());
        $_SESSION['guzzlecov'] = 5;
        try { new SessionCookieJar('guzzlecov'); $this->fail('expected'); } catch (\RuntimeException $e) { $this->assertSame('Invalid cookie data', $e->getMessage()); }
        $_SESSION['guzzlecov'] = '[{"Name":"a"}]';
        try { new SessionCookieJar('guzzlecov'); $this->fail('expected'); } catch (\RuntimeException $e) { $this->assertSame('Invalid cookie data', $e->getMessage()); }
        $_SESSION['guzzlecov'] = '"scalar"';
        try { new SessionCookieJar('guzzlecov'); $this->fail('expected'); } catch (\RuntimeException $e) { $this->assertSame('Invalid cookie data', $e->getMessage()); }
        $_SESSION['guzzlecov'] = '[]';
        $this->assertCount(0, new SessionCookieJar('guzzlecov', true));
        unset($_SESSION['guzzlecov']);
        $this->assertCount(0, new SessionCookieJar('guzzlecov'));
    }
}
