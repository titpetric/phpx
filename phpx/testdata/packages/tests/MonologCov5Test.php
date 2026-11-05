<?php
require_once __DIR__ . '/MonologCovStubGlobal.php';
require_once __DIR__ . '/MonologCovStubPredis.php';
require_once __DIR__ . '/MonologCovStubAmqpMsg.php';
require_once __DIR__ . '/MonologCovStubAmqpChan.php';
require_once __DIR__ . '/MonologCovStubRollbar.php';
require_once __DIR__ . '/MonologCovStubSqs.php';

use Monolog\{Logger, Level, LogRecord};
use Monolog\Handler\{TestHandler, RedisHandler, RedisPubSubHandler, AmqpHandler, RollbarHandler, SqsHandler, LogMonsterHandler, TelegramBotHandler, ZendMonitorHandler, FrankenPhpHandler, PHPConsoleHandler, GelfHandler, MissingExtensionException, DoctrineCouchDBHandler, MongoDBHandler};
use Monolog\Formatter\{LineFormatter, JsonFormatter};
use Monolog\Processor\TagProcessor;
use PHPUnit\Framework\TestCase;

final class MonologCovTelegram extends TelegramBotHandler
{
public array $messages = [];
protected function sendCurl(string $message): void { $this->messages[] = $message; }
public function callSend(string $m): void { $this->send($m); }
public function validate(string|bool $r): void { $this->validateApiResponse($r); }
public function url(): string { return $this->getBotApiUrl(); }
}

final class MonologCov5Test extends TestCase
{
private function rec(string $m = 'm', Level $l = Level::Info, array $ctx = [], array $extra = []): LogRecord
{
    return new LogRecord(new \DateTimeImmutable('2024-01-02 03:04:05', new \DateTimeZone('UTC')), 'chan', $l, $m, $ctx, $extra);
}

public function testRedisHandler(): void
{
    $r = new \Redis();
    $h = new RedisHandler($r, 'key', Level::Debug, true, 0);
    $h->setFormatter(new LineFormatter("%message%"));
    $h->handle($this->rec('a'));
    $this->assertSame([['rpush', 'key', 'a']], $r->calls);
    $this->assertInstanceOf(LineFormatter::class, (new RedisHandler($r, 'k'))->getFormatter());
    $r2 = new \Redis();
    $h = new RedisHandler($r2, 'key', Level::Debug, true, 5);
    $h->setFormatter(new LineFormatter("%message%"));
    $h->handle($this->rec('b'));
    $this->assertSame([['multi', 1], ['rpush', 'key', 'b'], ['ltrim', 'key', -5, -1], ['exec']], $r2->calls);
    $p = new \Predis\Client();
    $h = new RedisHandler($p, 'pk', Level::Debug, true, 3);
    $h->setFormatter(new LineFormatter("%message%"));
    $h->handle($this->rec('c'));
    $this->assertSame([['transaction'], ['rpush', 'pk', 'c'], ['ltrim', 'pk', -3, -1]], $p->calls);
}

public function testRedisPubSub(): void
{
    $r = new \Redis();
    $h = new RedisPubSubHandler($r, 'chan');
    $h->setFormatter(new LineFormatter("%message%"));
    $h->handle($this->rec('pub'));
    $this->assertSame([['publish', 'chan', 'pub']], $r->calls);
    $this->assertInstanceOf(LineFormatter::class, (new RedisPubSubHandler(new \Predis\Client(), 'c'))->getFormatter());
}

public function testAmqpExchange(): void
{
    $ex = new \AMQPExchange();
    $h = new AmqpHandler($ex, null, Level::Debug);
    $this->assertSame([], $h->getExtraAttributes());
    $this->assertSame($h, $h->setExtraAttributes(['priority' => 5]));
    $h->handle($this->rec('x', Level::Warning));
    $this->assertSame('warning.chan', $ex->published[0][1]);
    $this->assertSame(['delivery_mode' => 2, 'content_type' => 'application/json', 'priority' => 5], $ex->published[0][3]);
    $this->assertSame('x', json_decode($ex->published[0][0], true)['message']);
    $h->setExtraAttributes([]);
    $h->handle($this->rec('y'));
    $this->assertSame(['delivery_mode' => 2, 'content_type' => 'application/json'], $ex->published[1][3]);
    $h->handleBatch([$this->rec('b1'), $this->rec('b2')]);
    $this->assertCount(4, $ex->published);
    $this->assertInstanceOf(JsonFormatter::class, $h->getFormatter());
}

public function testAmqpChannel(): void
{
    $ch = new \PhpAmqpLib\Channel\AMQPChannel();
    $h = new AmqpHandler($ch, 'logs', Level::Info);
    $h->setExtraAttributes(['x' => 1]);
    $h->handle($this->rec('x'));
    $this->assertSame('logs', $ch->published[0][1]);
    $this->assertSame('info.chan', $ch->published[0][2]);
    $this->assertSame(['delivery_mode' => 2, 'content_type' => 'application/json', 'x' => 1], $ch->published[0][0]->props);
    $h->handleBatch([$this->rec('dbg', Level::Debug), $this->rec('b', Level::Error)]);
    $this->assertCount(1, $ch->batch);
    $this->assertSame('error.chan', $ch->batch[0][2]);
    $this->assertSame(1, $ch->flushed);
    $h2 = new AmqpHandler($ch);
    $h2->handle($this->rec('y'));
    $this->assertSame('', $ch->published[1][1]);
}

public function testRollbar(): void
{
    $rb = new \Rollbar\RollbarLogger();
    $h = new RollbarHandler($rb, Level::Debug);
    $h->handle($this->rec('r1', Level::Notice, ['k' => 1], ['e' => 2]));
    $h->handle($this->rec('r2', Level::Emergency, ['exception' => new \RuntimeException('boom')]));
    $this->assertSame('info', $rb->logs[0][0]);
    $this->assertSame('r1', $rb->logs[0][1]);
    $this->assertSame('NOTICE', $rb->logs[0][2]['monolog_level']);
    $this->assertSame('critical', $rb->logs[1][0]);
    $this->assertInstanceOf(\RuntimeException::class, $rb->logs[1][1]);
    $this->assertArrayNotHasKey('exception', $rb->logs[1][2]);
    $h->flush();
    $h->flush();
    $this->assertSame(1, $rb->flushes);
    $h->handle($this->rec('r3'));
    $h->close();
    $h->handle($this->rec('r4'));
    $h->reset();
    $this->assertSame(3, $rb->flushes);
    foreach ([[Level::Debug, 'debug'], [Level::Info, 'info'], [Level::Warning, 'warning'], [Level::Error, 'error'], [Level::Critical, 'critical'], [Level::Alert, 'critical']] as [$l, $e]) {
        $h->handle($this->rec('l', $l));
        $this->assertSame($e, $rb->logs[count($rb->logs) - 1][0]);
    }
}

public function testSqs(): void
{
    $c = new \Aws\Sqs\SqsClient();
    $h = new SqsHandler($c, 'https://q.invalid/1', Level::Debug);
    $h->setFormatter(new LineFormatter("%message%"));
    $h->handle($this->rec('small'));
    $this->assertSame(['QueueUrl' => 'https://q.invalid/1', 'MessageBody' => 'small'], $c->sent[0]);
    $h->handle($this->rec(str_repeat('a', 300000)));
    $this->assertSame(102400, strlen($c->sent[1]['MessageBody']));
    $arr = new SqsHandler($c, 'q');
    $arr->setFormatter(new \Monolog\Formatter\NormalizerFormatter());
    $this->expectException(\InvalidArgumentException::class);
    $arr->handle($this->rec());
}

public function testLogMonster(): void
{
    $t = new TestHandler();
    $m = new LogMonsterHandler($t, 3, Level::Warning, 'monster');
    $this->assertTrue($m->isHandling($this->rec()));
    $this->assertFalse($m->handle($this->rec()));
    $m->handle($this->rec());
    $m->close();
    $this->assertCount(1, $t->getRecords());
    $this->assertStringContainsString('ate 2 of 3', $t->getRecords()[0]->message);
    $this->assertSame(Level::Warning, $t->getRecords()[0]->level);
    $this->assertSame('monster', $t->getRecords()[0]->channel);
    $t->clear();
    $m->feed();
    $m->close();
    $this->assertCount(0, $t->getRecords());
    $m->reset();
    $this->assertSame($t, $m->getHandler());
    $m->pushProcessor(new TagProcessor(['mon']));
    $m->close();
    $this->assertSame(['tags' => ['mon']], $t->getRecords()[0]->extra);
    $chips = new LogMonsterHandler(fn($r, $h) => new TestHandler(), 2, 'error', 'c', true);
    $chips->handle($this->rec('no ctx'));
    $chips->handle($this->rec('ctx', Level::Info, ['a' => 1]));
    $chips->close();
    $this->assertInstanceOf(TestHandler::class, $chips->getHandler());
    $this->assertTrue($chips->getHandler()->hasErrorRecords());
    $f = new LineFormatter();
    $this->assertSame($m, $m->setFormatter($f));
    $this->assertSame($f, $m->getFormatter());
    $bad = new LogMonsterHandler(fn() => 1, 1);
    try { $bad->getHandler(); $this->fail(); } catch (\RuntimeException $e) { $this->assertStringContainsString('should return a HandlerInterface', $e->getMessage()); }
    $nf = new LogMonsterHandler(new \Monolog\Handler\NoopHandler(), 1);
    try { $nf->setFormatter($f); $this->fail(); } catch (\UnexpectedValueException $e) { $this->assertStringContainsString('does not support formatters', $e->getMessage()); }
    try { $nf->getFormatter(); $this->fail(); } catch (\UnexpectedValueException $e) { $this->assertTrue(true); }
}

public function testTelegram(): void
{
    $h = new MonologCovTelegram('KEY', '@chan', Level::Debug, true, null, null, null, false, false, null);
    $h->setFormatter(new LineFormatter("%message%"));
    $this->assertSame('https://api.telegram.org/bot', $h->url());
    $h->handle($this->rec('hello'));
    $this->assertSame(['hello'], $h->messages);
    $h->messages = [];
    $h->handleBatch([$this->rec('a', Level::Debug), $this->rec('b')]);
    $this->assertCount(1, $h->messages);
    $h->messages = [];
    $h->handleBatch([]);
    $this->assertSame([], $h->messages);
    $h->pushProcessor(new TagProcessor(['t']));
    $h->handleBatch([$this->rec('x')]);
    $long = str_repeat('a', 5000);
    $h->messages = [];
    $h->callSend($long);
    $this->assertSame(1, count($h->messages));
    $this->assertStringEndsWith('(…truncated)', $h->messages[0]);
    $this->assertSame(4096, strlen($h->messages[0]));
    $h->splitLongMessages(true);
    $h->messages = [];
    $h->callSend($long);
    $this->assertSame([4096, 904], array_map('strlen', $h->messages));
    $h->delayBetweenMessages(false);
    $h->setParseMode('HTML');
    $h->messages = [];
    $h->splitLongMessages(false);
    $h->callSend('<b>' . str_repeat('x', 5000) . '</b>');
    $this->assertStringContainsString('</b>', $h->messages[0]);
    $this->assertStringEndsWith('(…truncated)', $h->messages[0]);
    $h->splitLongMessages(true);
    $h->messages = [];
    $h->callSend('<b><i>' . str_repeat('y', 6000) . '</i></b>');
    $this->assertCount(2, $h->messages);
    $this->assertStringEndsWith('</i></b>', $h->messages[0]);
    $this->assertStringStartsWith('<b><i>', $h->messages[1]);
}

public function testTelegramApiResponse(): void
{
    $h = new MonologCovTelegram('KEY', '@chan');
    $h->validate('{"ok":true}');
    foreach ([[false, 'No response'], ['not json', 'non-JSON'], ['{"ok":false,"description":"bad"}', 'bad'], ['{"ok":false}', 'Unknown error']] as [$in, $msg]) {
        try { $h->validate($in); $this->fail(); } catch (\RuntimeException $e) { $this->assertStringContainsString($msg, $e->getMessage()); }
    }
}

public function testMissingExtensionHandlers(): void
{
    foreach ([ZendMonitorHandler::class, FrankenPhpHandler::class] as $cls) {
        if (function_exists('frankenphp_log') || function_exists('zend_monitor_custom_event')) { continue; }
        try { new $cls(); $this->fail($cls); } catch (MissingExtensionException $e) { $this->assertNotSame('', $e->getMessage()); }
    }
    if (!class_exists('PhpConsole\\Connector')) {
        try { new PHPConsoleHandler(); $this->fail(); } catch (\RuntimeException $e) { $this->assertStringContainsString('PHP Console library not found', $e->getMessage()); }
    }
    $this->assertInstanceOf(\Throwable::class, new MissingExtensionException('x'));
}
}

