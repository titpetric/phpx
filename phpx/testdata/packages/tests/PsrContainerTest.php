<?php
use PHPUnit\Framework\TestCase;
use Psr\Container\{ContainerInterface, NotFoundExceptionInterface, ContainerExceptionInterface};

final class NotFound extends RuntimeException implements NotFoundExceptionInterface {}
final class Box implements ContainerInterface {
    public function __construct(private array $items) {}
    public function get(string $id): mixed { if (!$this->has($id)) { throw new NotFound("no $id"); } $v = $this->items[$id]; return is_callable($v) ? $v($this) : $v; }
    public function has(string $id): bool { return array_key_exists($id, $this->items); }
}
final class PsrContainerTest extends TestCase
{
    public function testGetHas(): void
    {
        $c = new Box(['a' => 1, 'b' => fn(ContainerInterface $c) => $c->get('a') + 1]);
        $this->assertTrue($c->has('a'));
        $this->assertSame(2, $c->get('b'));
        $this->assertFalse($c->has('z'));
    }
    public function testNotFoundIsAContainerException(): void
    {
        $this->expectException(ContainerExceptionInterface::class);
        (new Box([]))->get('x');
    }
    public function testCatchByInterface(): void
    {
        try { (new Box([]))->get('q'); $this->fail('no throw'); }
        catch (NotFoundExceptionInterface $e) { $this->assertSame('no q', $e->getMessage()); }
    }
}
