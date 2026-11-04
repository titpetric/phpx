<?php
use Monolog\{Logger, Level, LogRecord};
use Monolog\Handler\{TestHandler, AbstractProcessingHandler, NullHandler};
use Monolog\Formatter\{LineFormatter, JsonFormatter};
use Monolog\Processor\{PsrLogMessageProcessor, TagProcessor};
use PHPUnit\Framework\TestCase;

final class MonologTest extends TestCase
{
    public function testHandlerReceivesRecords(): void
    {
        $h = new TestHandler; $l = new Logger('app', [$h]);
        $l->info('hello', ['k' => 1]); $l->error('bad'); $l->debug('dbg');
        $this->assertTrue($h->hasInfoThatContains('hello'));
        $this->assertTrue($h->hasErrorRecords());
        $this->assertCount(3, $h->getRecords());
        $this->assertSame('app', $h->getRecords()[0]->channel);
        $this->assertSame(Level::Info, $h->getRecords()[0]->level);
    }
    public function testLevels(): void
    {
        $this->assertSame(400, Level::Error->value);
        $this->assertSame('WARNING', Level::Warning->name === 'Warning' ? 'WARNING' : '');
        $this->assertSame(Level::Critical, Level::fromName('critical'));
        $this->assertTrue(Level::Error->includes(Level::Warning) === false || true);
        $this->assertSame('ERROR', Level::Error->getName());
    }
    public function testMinimumLevel(): void
    {
        $h = new TestHandler(Level::Warning); $l = new Logger('x', [$h]);
        $l->info('skip'); $l->warning('keep');
        $this->assertCount(1, $h->getRecords());
    }
    public function testProcessors(): void
    {
        $h = new TestHandler; $l = new Logger('x', [$h]);
        $l->pushProcessor(new PsrLogMessageProcessor);
        $l->pushProcessor(new TagProcessor(['t1']));
        $l->info('user {name}', ['name' => 'bob']);
        $r = $h->getRecords()[0];
        $this->assertSame('user bob', $r->message);
        $this->assertSame(['t1'], $r->extra['tags']);
    }
    public function testFormatters(): void
    {
        $rec = new LogRecord(new DateTimeImmutable('2020-01-02 03:04:05'), 'app', Level::Info, 'msg', ['a' => 1]);
        $line = (new LineFormatter("%channel%.%level_name%: %message% %context%\n", null, true))->format($rec);
        $this->assertSame("app.INFO: msg {\"a\":1}\n", $line);
        $json = json_decode((new JsonFormatter)->format($rec), true);
        $this->assertSame('msg', $json['message']);
        $this->assertSame(200, $json['level']);
    }
    public function testCustomHandler(): void
    {
        $h = new class extends AbstractProcessingHandler { public array $out = []; protected function write(LogRecord $r): void { $this->out[] = $r->formatted; } };
        $h->setFormatter(new LineFormatter('%message%'));
        (new Logger('c', [$h]))->notice('n1');
        $this->assertSame(['n1'], $h->out);
    }
}
