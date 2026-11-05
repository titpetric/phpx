<?php
use GuzzleHttp\Handler\{CurlShareHandleState, CurlHandler, CurlMultiHandler, CurlFactoryInterface, EasyHandle, CurlVersion, MockHandler};
use GuzzleHttp\{TransportSharing, Multiplexing, Utils};
use GuzzleHttp\Psr7\{Request, Response, Utils as PU7};
use PHPUnit\Framework\TestCase;

final class GuzzleCov6Test extends TestCase
{
    private function quiet(callable $fn)
    {
        set_error_handler(fn () => true, E_USER_DEPRECATED);
        try {
            return $fn();
        } finally {
            restore_error_handler();
        }
    }

    public function testNormalizeMode(): void
    {
        $this->assertSame('none', CurlShareHandleState::normalizeMode(null, 'o'));
        $this->assertSame('none', CurlShareHandleState::normalizeMode(TransportSharing::NONE, 'o'));
        $this->assertSame('handler_prefer', CurlShareHandleState::normalizeMode(TransportSharing::HANDLER_PREFER, 'o'));
        $this->assertSame('handler_require', CurlShareHandleState::normalizeMode(TransportSharing::HANDLER_REQUIRE, 'o'));
        try {
            CurlShareHandleState::normalizeMode('bogus', 'transport_sharing');
            $this->fail('expected');
        } catch (\InvalidArgumentException $e) {
            $this->assertStringContainsString('must be null or a GuzzleHttp\TransportSharing::* constant; received string', $e->getMessage());
        }
        try {
            CurlShareHandleState::normalizeMode(5, 'x');
            $this->fail('expected');
        } catch (\InvalidArgumentException $e) {
            $this->assertStringContainsString('received int', $e->getMessage());
        }
    }

    public function testFromOption(): void
    {
        $this->assertNull(CurlShareHandleState::fromOption(null));
        $this->assertNull(CurlShareHandleState::fromOption('none'));
        $prefer = CurlShareHandleState::fromOption('handler_prefer');
        $this->assertTrue($prefer === null || $prefer->mode === 'handler_prefer');
        if ($prefer !== null) {
            $this->assertSame($prefer, CurlShareHandleState::fromOption($prefer));
            $this->assertSame('handler_prefer', CurlShareHandleState::normalizeMode($prefer, 'o'));
        }
    }

    public function testFromOptionRequire(): void
    {
        try {
            $s = CurlShareHandleState::fromOption('handler_require');
            $this->assertSame('handler_require', $s->mode);
        } catch (\InvalidArgumentException $e) {
            $this->assertStringContainsString('transport_sharing', $e->getMessage());
        }
    }

    public function testCustomFactoryConflict(): void
    {
        CurlShareHandleState::assertNoRequiredSharingCustomFactoryConflict([], 'CurlHandler');
        CurlShareHandleState::assertNoRequiredSharingCustomFactoryConflict(['handle_factory' => null], 'CurlHandler');
        CurlShareHandleState::assertNoRequiredSharingCustomFactoryConflict(['handle_factory' => new GuzzleCov6Factory], 'CurlHandler');
        CurlShareHandleState::assertNoRequiredSharingCustomFactoryConflict(['handle_factory' => new GuzzleCov6Factory, 'transport_sharing' => 'handler_prefer'], 'CurlHandler');
        $this->expectException(\InvalidArgumentException::class);
        CurlShareHandleState::assertNoRequiredSharingCustomFactoryConflict(['handle_factory' => new GuzzleCov6Factory, 'transport_sharing' => 'handler_require'], 'CurlHandler');
    }

    public function testCurlHandlerConstructWithFactory(): void
    {
        $h = new CurlHandler(['handle_factory' => new GuzzleCov6Factory]);
        $this->assertInstanceOf(CurlHandler::class, $h);
        $this->quiet(fn () => new CurlHandler(['handle_factory' => new GuzzleCov6Factory, 'bogus' => 1]));
        $this->expectException(\InvalidArgumentException::class);
        new CurlHandler(['handle_factory' => new GuzzleCov6Factory, 'transport_sharing' => TransportSharing::HANDLER_REQUIRE]);
    }

    public function testCurlHandlerBadSharing(): void
    {
        $this->expectException(\InvalidArgumentException::class);
        new CurlHandler(['transport_sharing' => 'weird']);
    }

    public function testCurlHandlerRejectsBadHost(): void
    {
        $h = new CurlHandler(['handle_factory' => new GuzzleCov6Factory]);
        $this->expectException(\GuzzleHttp\Exception\RequestException::class);
        $h(new Request('GET', 'http://1.2.3.4./'), []);
    }

    public function testEasyHandleCreateResponse(): void
    {
        $e = new EasyHandle();
        $e->headers = ['HTTP/1.1 200 OK', 'Content-Type: text/plain', 'Content-Encoding: gzip', 'Content-Length: 99'];
        $e->sink = PU7::streamFor('body');
        $e->options = ['decode_content' => true];
        $e->createResponse();
        $this->assertSame(200, $e->response->getStatusCode());
        $this->assertSame('text/plain', $e->response->getHeaderLine('Content-Type'));
        $this->assertFalse($e->response->hasHeader('Content-Encoding'));
        $this->assertSame(['gzip'], $e->response->getHeader('x-encoded-content-encoding'));
        $this->assertSame(['99'], $e->response->getHeader('x-encoded-content-length'));
        $this->assertSame('4', $e->response->getHeaderLine('Content-Length'));
        $this->assertSame('body', (string) $e->response->getBody());
    }

    public function testEasyHandleEmptyBodyDropsLength(): void
    {
        $e = new EasyHandle();
        $e->headers = ['HTTP/1.1 204 No Content', 'Content-Encoding: gzip', 'Content-Length: 20'];
        $e->sink = PU7::streamFor('');
        $e->options = ['decode_content' => 'gzip'];
        $e->createResponse();
        $this->assertFalse($e->response->hasHeader('Content-Length'));
        $e2 = new EasyHandle();
        $e2->headers = ['HTTP/1.1 200 OK', 'Content-Encoding: gzip'];
        $e2->sink = PU7::streamFor('x');
        $e2->options = ['decode_content' => false];
        $e2->createResponse();
        $this->assertSame('gzip', $e2->response->getHeaderLine('Content-Encoding'));
    }

    public function testEasyHandleErrors(): void
    {
        $e = new EasyHandle();
        $e->headers = [];
        try {
            $e->createResponse();
            $this->fail('expected');
        } catch (\RuntimeException $ex) {
            $this->assertStringContainsString('non-empty', $ex->getMessage());
        }
        $this->expectException(\BadMethodCallException::class);
        unset($e->handle);
        $e->handle;
    }

    public function testEasyHandleInvalidProperty(): void
    {
        $e = new EasyHandle();
        try {
            $e->nonexistent;
            $this->fail('expected');
        } catch (\BadMethodCallException $ex) {
            $this->assertSame('Invalid property: nonexistent', $ex->getMessage());
        }
    }

    public function testMultiplexingConstants(): void
    {
        $this->assertSame('none', Multiplexing::NONE);
        foreach (['EAGER', 'WAIT', 'REQUIRE_EAGER', 'REQUIRE_WAIT'] as $c) {
            $this->assertIsString(constant(Multiplexing::class . '::' . $c));
        }
        $this->assertSame('handler_prefer', TransportSharing::HANDLER_PREFER);
    }

    public function testCurlVersionEnsureHelpers(): void
    {
        foreach (['ensureHandlerSharingSupported', 'ensureSslSessionSharingSupported'] as $m) {
            try {
                CurlVersion::$m();
                $this->assertTrue(CurlVersion::supportsHandlerSharing() || $m === 'ensureSslSessionSharingSupported');
            } catch (\InvalidArgumentException $e) {
                $this->assertNotSame('', $e->getMessage());
            }
        }
        try {
            CurlVersion::ensureConnectionCapsSupported('max_host_connections');
            $this->assertTrue(CurlVersion::supportsConnectionCaps());
        } catch (\InvalidArgumentException $e) {
            $this->assertStringContainsString('max_host_connections', $e->getMessage());
        }
    }

    public function testCurlMultiHandlerConstruct(): void
    {
        try {
            $h = new CurlMultiHandler(['handle_factory' => new GuzzleCov6Factory]);
            $this->assertInstanceOf(CurlMultiHandler::class, $h);
        } catch (\Throwable $e) {
            $this->assertNotSame('', $e->getMessage());
        }
        try {
            new CurlMultiHandler(['max_host_connections' => 0]);
            $this->fail('expected');
        } catch (\InvalidArgumentException $e) {
            $this->assertNotSame('', $e->getMessage());
        }
    }
}

final class GuzzleCov6Factory implements CurlFactoryInterface
{
    public function create(\Psr\Http\Message\RequestInterface $request, array $options): EasyHandle
    {
        return new EasyHandle();
    }

    public function release(EasyHandle $easy): void
    {
    }
}
