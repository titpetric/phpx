<?php
use PHPUnit\Framework\TestCase;

interface Repo { public function find(int $id): ?string; public function save(string $s): void; public function all(): array; }
class Mailer { public function send(string $to, string $body): bool { return true; } public function name(): string { return 'real'; } }
final class Service {
    public function __construct(private Repo $repo, private ?Mailer $mailer = null) {}
    public function name(int $id): string { return strtoupper($this->repo->find($id) ?? 'none'); }
    public function notify(int $id): bool { $n = $this->repo->find($id); return $n !== null && $this->mailer->send($n, 'hi'); }
}
final class MockTest extends TestCase
{
    public function testStub(): void
    {
        $repo = $this->createStub(Repo::class);
        $repo->method('find')->willReturn('bob');
        $this->assertSame('BOB', (new Service($repo))->name(1));
    }
    public function testStubDefaults(): void
    {
        $repo = $this->createStub(Repo::class);
        $this->assertNull($repo->find(1));
        $this->assertSame([], $repo->all());
    }
    public function testMockExpectation(): void
    {
        $repo = $this->createMock(Repo::class);
        $repo->expects($this->once())->method('save')->with('x');
        $repo->save('x');
    }
    public function testReturnMapAndCallback(): void
    {
        $repo = $this->createStub(Repo::class);
        $repo->method('find')->willReturnMap([[1, 'one'], [2, 'two']]);
        $this->assertSame('two', $repo->find(2));
        $r2 = $this->createStub(Repo::class);
        $r2->method('find')->willReturnCallback(fn(int $id) => "id$id");
        $this->assertSame('id9', $r2->find(9));
        $r3 = $this->createStub(Repo::class);
        $r3->method('find')->willReturnOnConsecutiveCalls('a', 'b');
        $this->assertSame('a', $r3->find(1));
        $this->assertSame('b', $r3->find(1));
    }
    public function testThrows(): void
    {
        $repo = $this->createStub(Repo::class);
        $repo->method('find')->willThrowException(new RuntimeException('db down'));
        $this->expectException(RuntimeException::class);
        $this->expectExceptionMessage('db down');
        $repo->find(1);
    }
    public function testClassMock(): void
    {
        $mailer = $this->createMock(Mailer::class);
        $mailer->expects($this->exactly(2))->method('send')->willReturn(true);
        $repo = $this->createStub(Repo::class);
        $repo->method('find')->willReturn('ann');
        $svc = new Service($repo, $mailer);
        $this->assertTrue($svc->notify(1));
        $this->assertTrue($svc->notify(2));
    }
    public function testNever(): void
    {
        $repo = $this->createMock(Repo::class);
        $repo->expects($this->never())->method('save');
        $this->assertSame('NONE', (new Service($repo))->name(5));
    }
    public function testExpectationFails(): void
    {
        $repo = $this->createMock(Repo::class);
        $repo->expects($this->once())->method('save')->with('expected');
        $repo->save('actual');
    }
    public function testMockBuilder(): void
    {
        $m = $this->getMockBuilder(Mailer::class)->onlyMethods(['send'])->getMock();
        $m->method('send')->willReturn(false);
        $this->assertFalse($m->send('a', 'b'));
        $this->assertSame('real', $m->name());
    }
}
