<?php
use PHPUnit\Framework\TestCase;
use Psr\Log\{AbstractLogger, LogLevel, LoggerInterface, NullLogger, LoggerAwareInterface, LoggerAwareTrait, LoggerTrait, InvalidArgumentException};

final class TraitLogger implements LoggerInterface {
    use LoggerTrait;
    public array $seen = [];
    public function log($level, string|\Stringable $message, array $context = []): void { $this->seen[] = [$level, $this->interpolate((string) $message, $context)]; }
    private function interpolate(string $m, array $ctx): string
    {
        $r = [];
        foreach ($ctx as $k => $v) { if (is_null($v) || is_scalar($v) || $v instanceof \Stringable) { $r['{' . $k . '}'] = (string) $v; } }
        return strtr($m, $r);
    }
}
final class Named implements \Stringable { public function __toString(): string { return 'named'; } }
final class Aware implements LoggerAwareInterface { use LoggerAwareTrait; public function logger(): ?LoggerInterface { return $this->logger; } }

final class PsrLogMoreTest extends TestCase
{
    public function testLevelConstants(): void
    {
        $this->assertSame(['emergency', 'alert', 'critical', 'error', 'warning', 'notice', 'info', 'debug'], [LogLevel::EMERGENCY, LogLevel::ALERT, LogLevel::CRITICAL, LogLevel::ERROR, LogLevel::WARNING, LogLevel::NOTICE, LogLevel::INFO, LogLevel::DEBUG]);
        $this->assertSame(8, count((new ReflectionClass(LogLevel::class))->getConstants()));
    }
    public function testTraitDispatch(): void
    {
        $l = new TraitLogger();
        $l->emergency('e'); $l->alert('a'); $l->critical('c'); $l->error('r'); $l->warning('w'); $l->notice('n'); $l->info('i'); $l->debug('d');
        $this->assertSame(['emergency', 'alert', 'critical', 'error', 'warning', 'notice', 'info', 'debug'], array_column($l->seen, 0));
        $l->log(LogLevel::INFO, 'direct');
        $this->assertSame(['info', 'direct'], end($l->seen));
    }
    public function testInterpolationAndStringable(): void
    {
        $l = new TraitLogger();
        $l->info('hi {who}, {n} {obj} {arr}', ['who' => 'bob', 'n' => 3, 'obj' => new Named, 'arr' => [1]]);
        $this->assertSame('hi bob, 3 named {arr}', $l->seen[0][1]);
        $l->info(new Named);
        $this->assertSame('named', $l->seen[1][1]);
    }
    public function testNullLogger(): void
    {
        $n = new NullLogger();
        $this->assertNull($n->info('x'));
        $this->assertNull($n->log('whatever', 'y'));
        $this->assertInstanceOf(AbstractLogger::class, $n);
    }
    public function testAwareTrait(): void
    {
        $a = new Aware();
        $this->assertNull($a->logger());
        $a->setLogger($l = new TraitLogger());
        $this->assertSame($l, $a->logger());
        $this->assertInstanceOf(LoggerAwareInterface::class, $a);
        $this->assertSame(['setLogger'], array_map(fn ($m) => $m->getName(), (new ReflectionClass(LoggerAwareInterface::class))->getMethods()));
    }
    public function testExceptionContext(): void
    {
        $l = new TraitLogger();
        $e = new InvalidArgumentException('bad level');
        $l->error('failed', ['exception' => $e]);
        $this->assertSame('failed', $l->seen[0][1]);
        $this->assertInstanceOf(\InvalidArgumentException::class, $e);
    }
}
