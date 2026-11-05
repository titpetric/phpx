<?php
// SyslogUdpHandler needs ext-sockets in its constructor; here a subclass calls the ancestor constructor
// instead and swaps in a recording socket, so nothing is ever sent.
use Monolog\{Level, LogRecord};
use Monolog\Handler\{AbstractSyslogHandler, SyslogUdpHandler};
use Monolog\Handler\SyslogUdp\UdpSocket;
use Monolog\Formatter\LineFormatter;
use PHPUnit\Framework\TestCase;

foreach (['LOG_KERN' => 0, 'LOG_USER' => 8, 'LOG_MAIL' => 16, 'LOG_DAEMON' => 24, 'LOG_AUTH' => 32, 'LOG_SYSLOG' => 40, 'LOG_LPR' => 48, 'LOG_NEWS' => 56, 'LOG_UUCP' => 64, 'LOG_CRON' => 72, 'LOG_AUTHPRIV' => 80, 'LOG_LOCAL0' => 128, 'LOG_LOCAL1' => 136, 'LOG_LOCAL2' => 144, 'LOG_LOCAL3' => 152, 'LOG_LOCAL4' => 160, 'LOG_LOCAL5' => 168, 'LOG_LOCAL6' => 176, 'LOG_LOCAL7' => 184] as $name => $value) {
    if (!defined($name)) { define($name, $value); }
}

final class MonologCov13Socket extends UdpSocket
{
    public array $out = [];
    public bool $closed = false;
    protected function send(string $chunk): void { $this->out[] = $chunk; }
    public function close(): void { $this->closed = true; }
}

final class MonologCov13Syslog extends SyslogUdpHandler
{
    public function __construct(string|int $facility, int $rfc, string $ident = 'php', ?int $max = null)
    {
        AbstractSyslogHandler::__construct($facility, Level::Debug, true);
        $this->ident = $ident;
        $this->rfc = $rfc;
        $this->socket = new MonologCov13Socket('127.0.0.1', 514, $max);
    }
    public function out(): array { return $this->socket->out; }
    public function facility(): int { return $this->facility; }
}

final class MonologCov13Plain extends AbstractSyslogHandler
{
    protected function write(LogRecord $record): void {}
    public function facility(): int { return $this->facility; }
    public function priority(Level $l): int { return $this->toSyslogPriority($l); }
}

final class MonologCov13Test extends TestCase
{
    public function testAbstractSyslogFacilities(): void
    {
        foreach (['user' => 8, 'LOCAL0' => 128, 'auth' => 32, 'cron' => 72, 'daemon' => 24, 'local7' => 184, 'authpriv' => 80, 'uucp' => 64] as $name => $value) {
            $this->assertSame($value, (new MonologCov13Plain($name))->facility());
        }
        $this->assertSame(16, (new MonologCov13Plain(LOG_MAIL))->facility());
        $h = new MonologCov13Plain('user', Level::Error, false);
        $this->assertSame(Level::Error, $h->getLevel());
        $this->assertSame([7, 3, 0], [$h->priority(Level::Debug), $h->priority(Level::Error), $h->priority(Level::Emergency)]);
        $this->assertInstanceOf(LineFormatter::class, $h->getFormatter());
        try { new MonologCov13Plain('bogus-facility'); $this->fail(); } catch (\UnexpectedValueException $e) { $this->assertStringContainsString('Unknown facility value "bogus-facility"', $e->getMessage()); }
        try { new MonologCov13Plain(999); $this->fail(); } catch (\UnexpectedValueException $e) { $this->assertStringContainsString('Unknown facility value "999"', $e->getMessage()); }
    }

    private function rec(string $m = 'm', Level $l = Level::Info): LogRecord
    {
        return new LogRecord(new \DateTimeImmutable('2024-01-02 03:04:05', new \DateTimeZone('UTC')), 'chan', $l, $m);
    }

    public function testFacilityResolution(): void
    {
        foreach (['user' => 8, 'LOCAL0' => 128, 'auth' => 32, 'cron' => 72, 'daemon' => 24, 'local7' => 184] as $name => $value) {
            $this->assertSame($value, (new MonologCov13Syslog($name, SyslogUdpHandler::RFC5424))->facility());
        }
        $this->assertSame(16, (new MonologCov13Syslog(LOG_MAIL, SyslogUdpHandler::RFC5424))->facility());
        try { new MonologCov13Syslog('bogus-facility', SyslogUdpHandler::RFC5424); $this->fail(); } catch (\UnexpectedValueException $e) { $this->assertStringContainsString('Unknown facility value "bogus-facility"', $e->getMessage()); }
        try { new MonologCov13Syslog(999, SyslogUdpHandler::RFC5424); $this->fail(); } catch (\UnexpectedValueException $e) { $this->assertStringContainsString('Unknown facility value "999"', $e->getMessage()); }
    }

    public function testRfc5424Header(): void
    {
        $h = new MonologCov13Syslog('local3', SyslogUdpHandler::RFC5424, 'myapp', 200);
        $h->setFormatter(new LineFormatter("%message%", null, true));
        $h->handle($this->rec("line1\nline2", Level::Error));
        $out = $h->out();
        $this->assertCount(2, $out);
        $this->assertMatchesRegularExpression('/^<155>1 2024-01-02T03:04:05\+00:00 \S+ myapp \d+ - - line1$/', $out[0]);
        $this->assertStringEndsWith('line2', $out[1]);
        $this->assertInstanceOf(LineFormatter::class, $h->getFormatter());
        $h->close();
    }

    public function testRfc3164AndExtendedHeaders(): void
    {
        $h = new MonologCov13Syslog('user', SyslogUdpHandler::RFC3164);
        $h->setFormatter(new LineFormatter("%message%"));
        $h->handle($this->rec('legacy'));
        $this->assertMatchesRegularExpression('/^<14>Jan 02 03:04:05 \S+ php\[\d+\]: legacy$/', $h->out()[0]);
        $e = new MonologCov13Syslog('user', SyslogUdpHandler::RFC5424e);
        $e->setFormatter(new LineFormatter("%message%"));
        $e->handle($this->rec('extended'));
        $this->assertMatchesRegularExpression('/^<14>1 2024-01-02T03:04:05\.000\+00:00 /', $e->out()[0]);
    }

    public function testLongMessageIsCut(): void
    {
        $h = new MonologCov13Syslog('user', SyslogUdpHandler::RFC5424, 'php', 90);
        $h->setFormatter(new LineFormatter("%message%"));
        $h->handle($this->rec(str_repeat('x', 500)));
        $this->assertSame(90, strlen($h->out()[0]));
    }

    public function testSetSocketAndClose(): void
    {
        $h = new MonologCov13Syslog('user', SyslogUdpHandler::RFC5424);
        $sock = new MonologCov13Socket('127.0.0.1', 514);
        $this->assertSame($h, $h->setSocket($sock));
        $h->setFormatter(new LineFormatter("%message%"));
        $h->handle($this->rec('swapped'));
        $this->assertCount(1, $sock->out);
        $h->close();
        $this->assertTrue($sock->closed);
    }

    public function testUdpSocketAssemble(): void
    {
        $s = new MonologCov13Socket('127.0.0.1', 514, 20);
        UdpSocket::class;
        $parent = new class('127.0.0.1', 514, 20) extends UdpSocket {
            public array $out = [];
            protected function send(string $chunk): void { $this->out[] = $chunk; }
        };
        $parent->write('0123456789abcdefghij', 'HEADER: ');
        $parent->write('short');
        $this->assertSame(['HEADER: 0123456789ab', 'short'], $parent->out);
        $parent->close();
    }
}
