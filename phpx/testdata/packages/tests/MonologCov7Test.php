<?php
foreach (['PhpConsole', 'Dynamo', 'Mime', 'Mailer', 'Swift'] as $stub) {
    require_once __DIR__ . '/MonologCovStub' . $stub . '.php';
}

use Monolog\{Logger, Level, LogRecord};
use Monolog\Handler\{PHPConsoleHandler, DynamoDbHandler, SymfonyMailerHandler, MandrillHandler};
use Monolog\Formatter\{LineFormatter, ScalarFormatter};
use PhpConsole\{Connector, Helper, Handler as PcHandler};
use Symfony\Component\Mime\Email;
use PHPUnit\Framework\TestCase;

final class MonologCov7Test extends TestCase
{
    private function rec(string $m = 'm', Level $l = Level::Info, array $ctx = [], array $extra = []): LogRecord
    {
        return new LogRecord(new \DateTimeImmutable('2024-01-02 03:04:05', new \DateTimeZone('UTC')), 'chan', $l, $m, $ctx, $extra);
    }

    public function testPhpConsoleHandler(): void
    {
        $c = new Connector();
        $h = new PHPConsoleHandler(['dumperLevelLimit' => 7, 'sourcesBasePath' => '/src', 'serverEncoding' => 'utf-8', 'password' => 'pw', 'enableSslOnlyMode' => true, 'ipMasks' => ['127.0.0.1'], 'headersLimit' => 100, 'detectDumpTraceAndSource' => true, 'enableEvalListener' => true, 'useOwnErrorsHandler' => true, 'useOwnExceptionsHandler' => true], $c);
        $this->assertSame($c, $h->getConnector());
        $this->assertSame(7, $h->getOptions()['dumperLevelLimit']);
        $this->assertSame(7, $c->dumper->levelLimit);
        $this->assertTrue($c->debug->detectTraceAndSource);
        $this->assertTrue(Helper::isRegistered());
        $this->assertSame([['errors', true], ['exceptions', true], ['start']], PcHandler::getInstance()->calls);
        $this->assertSame(['base', 'enc', 'pw', 'ssl', 'ips', 'limit', 'eval'], array_column($c->calls, 0));

        $h->handle($this->rec('debug msg', Level::Debug, ['tag' => 'mytag', 'k' => 'v']));
        $h->handle($this->rec('with first', Level::Info, ['first-tag-value']));
        $h->handle($this->rec('no ctx', Level::Info));
        $this->assertSame('mytag', $c->debug->debugs[0][1]);
        $this->assertStringContainsString('debug msg {"dumped":{"k":"v"}}', $c->debug->debugs[0][0]);
        $this->assertSame('info', $c->debug->debugs[2][1]);
        $h->handle($this->rec('note', Level::Notice, ['code' => 3, 'file' => 'a.php', 'line' => 9]));
        $h->handle($this->rec('warn', Level::Warning));
        $this->assertSame([3, 'note', 'a.php', 9], $c->errors->errors[0]);
        $this->assertSame([null, 'warn', null, null], $c->errors->errors[1]);
        $ex = new \RuntimeException('x');
        $h->handle($this->rec('exc', Level::Error, ['exception' => $ex]));
        $this->assertSame($ex, $c->errors->exceptions[0]);
        $c->active = false;
        $this->assertFalse($h->handle($this->rec('inactive')));
        $this->assertInstanceOf(LineFormatter::class, $h->getFormatter());
    }

    public function testPhpConsoleDefaultsAndErrors(): void
    {
        Connector::$instance = null;
        $c = new Connector();
        Connector::$instance = $c;
        $h = new PHPConsoleHandler();
        $this->assertSame($c, $h->getConnector());
        $this->assertSame([], $c->calls);
        try { new PHPConsoleHandler(['bogus' => 1, 'other' => 2]); $this->fail(); } catch (\RuntimeException $e) { $this->assertSame('Unknown options: bogus, other', $e->getMessage()); }
        $off = new PHPConsoleHandler(['enabled' => false], $c);
        $this->assertTrue($off->handle($this->rec()) === false);
    }

    public function testDynamoDb(): void
    {
        $client = new \Aws\DynamoDb\DynamoDbClient();
        $h = new DynamoDbHandler($client, 'logs');
        $h->handle($this->rec('dyn', Level::Info, ['a' => 1], []));
        $this->assertSame('logs', $client->items[0]['TableName']);
        $this->assertArrayHasKey('message', $client->items[0]['Item']);
        $this->assertInstanceOf(ScalarFormatter::class, $h->getFormatter());
    }

    public function testSymfonyMailer(): void
    {
        $mailer = new class implements \Symfony\Component\Mailer\MailerInterface { public array $sent = []; public function send($m) { $this->sent[] = $m; } };
        $tpl = new Email();
        $tpl->subject('Alert %level_name% %message%');
        $h = new SymfonyMailerHandler($mailer, $tpl, Level::Debug);
        $h->setFormatter(new LineFormatter("%message%"));
        $h->handle($this->rec('plain body', Level::Error));
        $this->assertCount(1, $mailer->sent);
        $m = $mailer->sent[0];
        $this->assertNotSame($tpl, $m);
        $this->assertSame('Alert ERROR plain body', $m->subject);
        $this->assertSame('plain body', $m->text);
        $this->assertNull($m->html);
        $this->assertInstanceOf(\DateTimeInterface::class, $m->date);
        $tpl->htmlCharset = 'iso-8859-1';
        $tpl->textCharset = 'ascii';
        $h2 = new SymfonyMailerHandler($mailer, $tpl, Level::Debug);
        $h2->setFormatter(new LineFormatter("<b>%message%</b>"));
        $h2->handle($this->rec('html', Level::Warning));
        $this->assertSame('<b>html</b>', $mailer->sent[1]->html);
        $h3 = new SymfonyMailerHandler($mailer, function (string $content, array $records) { return (new Email())->subject('from closure'); }, Level::Debug);
        $h3->setFormatter(new LineFormatter("%message%"));
        $h3->handleBatch([$this->rec('c1', Level::Error), $this->rec('c2', Level::Critical)]);
        $this->assertStringStartsWith('from closure', $mailer->sent[2]->subject);
        $this->assertSame("c1c2", $mailer->sent[2]->text);
        $h4 = new SymfonyMailerHandler($mailer, fn() => null, Level::Debug);
        $h4->setFormatter(new LineFormatter("%message%"));
        try { $h4->handle($this->rec('bad')); $this->fail(); } catch (\InvalidArgumentException $e) {
            $this->assertStringContainsString('Could not resolve message', $e->getMessage());
            $this->assertStringContainsString('bad', $e->getMessage());
        }
    }

    public function testMandrillConstructor(): void
    {
        $msg = new \Swift_Message();
        $h = new MandrillHandler('key', $msg);
        $this->assertInstanceOf(MandrillHandler::class, $h);
        $h2 = new MandrillHandler('key', fn() => new \Swift_Message(), Level::Debug, false);
        $this->assertInstanceOf(MandrillHandler::class, $h2);
        $this->expectException(\InvalidArgumentException::class);
        new MandrillHandler('key', fn() => 'not a message');
    }
}
