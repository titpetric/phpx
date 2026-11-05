<?php
use PHPUnit\Framework\TestCase;
use Ramsey\Uuid\Uuid;

final class UuidEntryTest extends TestCase
{
    public function testV4Shape(): void
    {
        $u = Uuid::uuid4();
        $this->assertMatchesRegularExpression('/^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/', $u->toString());
        $this->assertSame(4, $u->getVersion());
    }
    public function testNameBased(): void
    {
        $this->assertSame('6ba7b810-9dad-11d1-80b4-00c04fd430c8', Uuid::NAMESPACE_DNS);
        $this->assertSame('886313e1-3b8a-5372-9b90-0c9aee199e5d', Uuid::uuid5(Uuid::NAMESPACE_DNS, 'python.org')->toString());
        $this->assertSame('6fa459ea-ee8a-3ca4-894e-db77e160355e', Uuid::uuid3(Uuid::NAMESPACE_DNS, 'python.org')->toString());
    }
    public function testParseAndBytes(): void
    {
        $u = Uuid::fromString('886313e1-3b8a-5372-9b90-0c9aee199e5d');
        $this->assertSame(16, strlen($u->getBytes()));
        $this->assertSame('886313e13b8a53729b900c9aee199e5d', $u->getHex()->toString());
        $this->assertTrue(Uuid::isValid('886313e1-3b8a-5372-9b90-0c9aee199e5d'));
        $this->assertFalse(Uuid::isValid('nope'));
        $this->assertTrue($u->equals(Uuid::fromBytes($u->getBytes())));
    }
}
