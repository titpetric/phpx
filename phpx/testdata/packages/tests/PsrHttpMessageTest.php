<?php
use PHPUnit\Framework\TestCase;
use Psr\Http\Message\{MessageInterface, RequestInterface, ResponseInterface, StreamInterface, UriInterface};

final class PsrHttpMessageTest extends TestCase
{
    public function testInterfacesDeclareTheContract(): void
    {
        $r = new ReflectionClass(ResponseInterface::class);
        $this->assertTrue($r->isInterface());
        $this->assertTrue($r->hasMethod('getStatusCode'));
        $this->assertTrue($r->implementsInterface(MessageInterface::class));
        $this->assertTrue((new ReflectionClass(RequestInterface::class))->hasMethod('getRequestTarget'));
    }
    public function testMethodSignatures(): void
    {
        $m = new ReflectionMethod(StreamInterface::class, 'read');
        $this->assertSame(1, $m->getNumberOfParameters());
        $this->assertSame('length', $m->getParameters()[0]->getName());
        $this->assertSame('string', $m->getReturnType()->getName());
        $this->assertTrue((new ReflectionMethod(UriInterface::class, 'withPort'))->getParameters()[0]->allowsNull());
    }
    public function testStringableUri(): void
    {
        $this->assertTrue((new ReflectionClass(UriInterface::class))->hasMethod('__toString'));
    }
}
