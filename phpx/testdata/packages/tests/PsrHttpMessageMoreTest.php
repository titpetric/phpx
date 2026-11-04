<?php
use PHPUnit\Framework\TestCase;
use Psr\Http\Message\{MessageInterface, RequestInterface, ResponseInterface, ServerRequestInterface, StreamInterface, UriInterface, UploadedFileInterface, RequestFactoryInterface, ResponseFactoryInterface, StreamFactoryInterface, UriFactoryInterface};

final class PsrHttpMessageMoreTest extends TestCase
{
    private function methods(string $c): array
    {
        $m = array_map(fn ($x) => $x->getName(), (new ReflectionClass($c))->getMethods());
        sort($m);
        return $m;
    }
    public function testMessageContract(): void
    {
        $this->assertSame(['getBody', 'getHeader', 'getHeaderLine', 'getHeaders', 'getProtocolVersion', 'hasHeader', 'withAddedHeader', 'withBody', 'withHeader', 'withProtocolVersion', 'withoutHeader'], $this->methods(MessageInterface::class));
    }
    public function testHierarchy(): void
    {
        $this->assertTrue(is_subclass_of(ServerRequestInterface::class, RequestInterface::class));
        $this->assertTrue(is_subclass_of(RequestInterface::class, MessageInterface::class));
        $this->assertTrue(is_subclass_of(ResponseInterface::class, MessageInterface::class));
        $this->assertSame([MessageInterface::class], array_keys((new ReflectionClass(RequestInterface::class))->getInterfaces()));
    }
    public function testStreamContract(): void
    {
        $this->assertSame(['__toString', 'close', 'detach', 'eof', 'getContents', 'getMetadata', 'getSize', 'isReadable', 'isSeekable', 'isWritable', 'read', 'rewind', 'seek', 'tell', 'write'], $this->methods(StreamInterface::class));
        $this->assertTrue(is_subclass_of(StreamInterface::class, Stringable::class));
    }
    public function testUriContract(): void
    {
        $m = $this->methods(UriInterface::class);
        foreach (['getScheme', 'getAuthority', 'getUserInfo', 'getHost', 'getPort', 'getPath', 'getQuery', 'getFragment', 'withScheme', 'withUserInfo', 'withHost', 'withPort', 'withPath', 'withQuery', 'withFragment', '__toString'] as $n) {
            $this->assertContains($n, $m);
        }
        $p = (new ReflectionMethod(UriInterface::class, 'withUserInfo'))->getParameters();
        $this->assertSame(['user', 'password'], array_map(fn ($x) => $x->getName(), $p));
        $this->assertTrue($p[1]->isOptional());
        $this->assertNull($p[1]->getDefaultValue());
    }
    public function testReturnTypes(): void
    {
        $rt = fn ($c, $m) => (string) (new ReflectionMethod($c, $m))->getReturnType();
        $this->assertSame('int', $rt(ResponseInterface::class, 'getStatusCode'));
        $this->assertSame('string', $rt(RequestInterface::class, 'getMethod'));
        $this->assertSame('Psr\Http\Message\UriInterface', $rt(RequestInterface::class, 'getUri'));
        $this->assertSame('Psr\Http\Message\ResponseInterface', $rt(ResponseInterface::class, 'withStatus'));
        $this->assertSame('?int', $rt(UriInterface::class, 'getPort'));
        $this->assertSame('array', $rt(MessageInterface::class, 'getHeaders'));
    }
    public function testFactoriesAndUploads(): void
    {
        $this->assertTrue((new ReflectionClass(RequestFactoryInterface::class))->hasMethod('createRequest'));
        $this->assertTrue((new ReflectionClass(ResponseFactoryInterface::class))->hasMethod('createResponse'));
        $this->assertTrue((new ReflectionClass(StreamFactoryInterface::class))->hasMethod('createStreamFromFile'));
        $this->assertTrue((new ReflectionClass(UriFactoryInterface::class))->hasMethod('createUri'));
        $this->assertContains('moveTo', $this->methods(UploadedFileInterface::class));
        $this->assertContains('getAttributes', $this->methods(ServerRequestInterface::class));
    }
}
