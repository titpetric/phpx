<?php
use Monolog\{Level, LogRecord};
use Monolog\Handler\{FrankenPhpHandler, ZendMonitorHandler, LogmaticHandler};
use Monolog\Formatter\{NormalizerFormatter, LogmaticFormatter, LineFormatter};
use PHPUnit\Framework\TestCase;

final class MonologCov9Log { public static array $franken = []; }

final class MonologCov9Franken extends FrankenPhpHandler
{
    public function lvl(Level $l): int { return $this->toFrankenPhpLevel($l); }
}

final class MonologCov9Zend extends ZendMonitorHandler
{
    public array $events = [];
    protected function writeZendMonitorCustomEvent(string $type, string $message, array $formatted, int $severity): void { $this->events[] = [$type, $message, $severity]; }
    public function lvl(Level $l): int { return $this->toZendMonitorLevel($l); }
}

final class MonologCov9Logmatic extends LogmaticHandler
{
    public array $sent = [];
    protected function fsockopen() { return fopen('php://memory', 'w+'); }
    protected function streamSetTimeout(): bool { return true; }
    protected function fwrite(string $data): int|bool { $this->sent[] = $data; return strlen($data); }
    public function isConnected(): bool { return is_resource($this->getResource()); }
}

final class MonologCov9Test extends TestCase
{
    private function rec(string $m = 'm', Level $l = Level::Info, array $ctx = [], array $extra = []): LogRecord
    {
        return new LogRecord(new \DateTimeImmutable('2024-01-02 03:04:05', new \DateTimeZone('UTC')), 'chan', $l, $m, $ctx, $extra);
    }

    public function testFrankenPhpHandler(): void
    {
        foreach (['FRANKENPHP_LOG_LEVEL_DEBUG' => -4, 'FRANKENPHP_LOG_LEVEL_INFO' => 0, 'FRANKENPHP_LOG_LEVEL_WARN' => 4, 'FRANKENPHP_LOG_LEVEL_ERROR' => 8] as $c => $v) {
            if (!defined($c)) { define($c, $v); }
        }
        if (!function_exists('frankenphp_log')) {
            function frankenphp_log($message, $level, $context) { MonologCov9Log::$franken[] = [$message, $level, $context]; }
        }
        $h = new MonologCov9Franken();
        $h->handle($this->rec('fp', Level::Warning, ['k' => 'v']));
        $this->assertSame('fp', MonologCov9Log::$franken[0][0]);
        $this->assertSame(4, MonologCov9Log::$franken[0][1]);
        $this->assertArrayNotHasKey('message', MonologCov9Log::$franken[0][2]);
        $this->assertArrayNotHasKey('level', MonologCov9Log::$franken[0][2]);
        $this->assertArrayNotHasKey('datetime', MonologCov9Log::$franken[0][2]);
        $this->assertSame('WARNING', MonologCov9Log::$franken[0][2]['level_name']);
        $this->assertSame([-4, 0, 1, 4, 8, 9, 10, 12], array_map(fn($l) => $h->lvl($l), [Level::Debug, Level::Info, Level::Notice, Level::Warning, Level::Error, Level::Critical, Level::Alert, Level::Emergency]));
        $this->assertInstanceOf(NormalizerFormatter::class, $h->getDefaultFormatter());
        (new FrankenPhpHandler())->handle($this->rec('real write'));
        $this->assertCount(2, MonologCov9Log::$franken);
    }

    public function testZendMonitorHandler(): void
    {
        foreach (['ZEND_MONITOR_EVENT_SEVERITY_INFO' => 0, 'ZEND_MONITOR_EVENT_SEVERITY_WARNING' => 1, 'ZEND_MONITOR_EVENT_SEVERITY_ERROR' => 2] as $c => $v) {
            if (!defined($c)) { define($c, $v); }
        }
        if (!function_exists('zend_monitor_custom_event')) {
            function zend_monitor_custom_event($type, $message, $formatted, $severity) {}
        }
        $h = new MonologCov9Zend();
        $h->handle($this->rec('zm', Level::Error));
        $this->assertSame(['ERROR', 'zm', 2], $h->events[0]);
        $this->assertSame([0, 0, 0, 1, 2, 2, 2, 2], array_map(fn($l) => $h->lvl($l), Level::cases()));
        $this->assertInstanceOf(NormalizerFormatter::class, $h->getDefaultFormatter());
        (new ZendMonitorHandler())->handle($this->rec('real write', Level::Warning));
    }

    public function testLogmaticHandler(): void
    {
        $h = new MonologCov9Logmatic('tok', 'host1', 'app1', false, Level::Debug, true, false, 1.0, 2.0, 3.0, 512);
        $this->assertSame('api.logmatic.io:10514/v1/', $h->getConnectionString());
        $this->assertInstanceOf(LogmaticFormatter::class, $h->getFormatter());
        $h->handle($this->rec('lm'));
        $this->assertStringStartsWith('tok {', $h->sent[0]);
        $this->assertStringContainsString('"hostname":"host1"', $h->sent[0]);
        $this->assertStringContainsString('"appname":"app1"', $h->sent[0]);
        $plain = new MonologCov9Logmatic('tok2', '', '', false);
        $f = $plain->getFormatter();
        $plain->handle($this->rec('lm2'));
        $this->assertStringNotContainsString('hostname', $plain->sent[0]);
        $this->assertSame('api.logmatic.io:10514/v1/', $plain->getConnectionString());
    }
}
