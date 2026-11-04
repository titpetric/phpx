<?php
use PHPUnit\Framework\TestCase;
use Psr\Container\{ContainerInterface, NotFoundExceptionInterface, ContainerExceptionInterface};

final class Missing extends RuntimeException implements NotFoundExceptionInterface {}
final class Broken extends LogicException implements ContainerExceptionInterface {}
final class Lazy implements ContainerInterface {
    private array $built = [];
    public int $builds = 0;
    public function __construct(private array $defs) {}
    public function get(string $id): mixed
    {
        if (isset($this->built[$id])) { return $this->built[$id]; }
        if (!isset($this->defs[$id])) { throw new Missing(sprintf('Service "%s" not found', $id)); }
        $this->builds++;
        try { return $this->built[$id] = ($this->defs[$id])($this); }
        catch (Missing $e) { throw new Broken("cannot build $id", 0, $e); }
    }
    public function has(string $id): bool { return isset($this->defs[$id]); }
}
final class PsrContainerMoreTest extends TestCase
{
    public function testSharedInstances(): void
    {
        $c = new Lazy(['o' => fn () => new stdClass]);
        $this->assertSame($c->get('o'), $c->get('o'));
        $this->assertSame(1, $c->builds);
    }
    public function testDependencyChain(): void
    {
        $c = new Lazy(['a' => fn () => 1, 'b' => fn ($c) => $c->get('a') + 1, 'c' => fn ($c) => $c->get('b') * 10]);
        $this->assertSame(20, $c->get('c'));
        $this->assertSame(3, $c->builds);
    }
    public function testWrappedExceptionKeepsPrevious(): void
    {
        $c = new Lazy(['x' => fn ($c) => $c->get('nope')]);
        try { $c->get('x'); $this->fail('no throw'); }
        catch (ContainerExceptionInterface $e) {
            $this->assertInstanceOf(Broken::class, $e);
            $this->assertInstanceOf(NotFoundExceptionInterface::class, $e->getPrevious());
            $this->assertSame('Service "nope" not found', $e->getPrevious()->getMessage());
        }
    }
    public function testInterfaceHierarchy(): void
    {
        $this->assertTrue(is_subclass_of(NotFoundExceptionInterface::class, ContainerExceptionInterface::class));
        $this->assertTrue(is_subclass_of(NotFoundExceptionInterface::class, Throwable::class));
        $this->assertTrue((new ReflectionClass(ContainerInterface::class))->hasMethod('has'));
        $this->assertInstanceOf(Throwable::class, new Missing('x'));
        $this->assertSame(['get', 'has'], array_map(fn ($m) => $m->getName(), (new ReflectionClass(ContainerInterface::class))->getMethods()));
    }
    public function testClosureFactoriesAndTypes(): void
    {
        $c = new Lazy(['n' => fn () => null, 'arr' => fn () => [1, 2], 'f' => fn () => fn ($x) => $x * 2]);
        $this->assertTrue($c->has('n'));
        $this->assertNull($c->get('n'));
        $this->assertSame([1, 2], $c->get('arr'));
        $this->assertSame(8, $c->get('f')(4));
        $this->assertFalse($c->has('zzz'));
    }
}
