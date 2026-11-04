<?php
use PHPUnit\Framework\TestCase;
use Psr\Log\{AbstractLogger, LogLevel, LoggerInterface, NullLogger, LoggerAwareTrait, InvalidArgumentException};

final class MemLogger extends AbstractLogger { public array $lines = []; public function log($level, string|\Stringable $message, array $context = []): void { $this->lines[] = $level . ':' . $message; } }
final class Svc { use LoggerAwareTrait; public function go(): void { $this->logger?->warning('careful'); } }

final class PsrLogTest extends TestCase
{
    public function testLevelsRouteToLog(): void
    {
        $l = new MemLogger;
        $l->info('a'); $l->error('b'); $l->debug('c'); $l->emergency('d');
        $this->assertSame(['info:a', 'error:b', 'debug:c', 'emergency:d'], $l->lines);
        $this->assertInstanceOf(LoggerInterface::class, $l);
    }
    public function testConstants(): void
    {
        $this->assertSame('critical', LogLevel::CRITICAL);
        $this->assertCount(8, (new ReflectionClass(LogLevel::class))->getConstants());
    }
    public function testAwareTrait(): void
    {
        $s = new Svc; $s->go();
        $l = new MemLogger; $s->setLogger($l); $s->go();
        $this->assertSame(['warning:careful'], $l->lines);
    }
    public function testNullLogger(): void
    {
        (new NullLogger)->alert('x');
        $this->assertTrue(true);
    }
}
