<?php
// Handlers whose constructor insists on a php extension (curl, openssl) are subclassed here with a constructor
// that calls the ancestor constructor directly, so everything after the extension check runs.
use Monolog\{Level, LogRecord};
use Monolog\Handler\{AbstractHandler, AbstractProcessingHandler, LogglyHandler, SlackWebhookHandler, SocketHandler, SlackHandler, FleepHookHandler, FlowdockHandler, TelegramBotHandler, CubeHandler, SyslogUdpHandler};
use Monolog\Handler\Slack\SlackRecord;
use Monolog\Formatter\{LineFormatter, FlowdockFormatter};
use Monolog\Processor\TagProcessor;
use PHPUnit\Framework\TestCase;

trait MonologCov12Sock
{
    public array $sent = [];
    protected function fsockopen() { return fopen('php://memory', 'w+'); }
    protected function pfsockopen() { return fopen('php://memory', 'w+'); }
    protected function streamSetTimeout(): bool { return true; }
    protected function streamSetChunkSize(): int|bool { return 1; }
    protected function fwrite(string $data): int|bool { $this->sent[] = $data; return strlen($data); }
    public function isConnected(): bool { return is_resource($this->getResource()); }
}

trait MonologCov12Boot
{
    private function setPrivate(string $owner, string $name, mixed $value): void
    {
        (new \ReflectionProperty($owner, $name))->setValue($this, $value);
    }
}

final class MonologCov12Slack extends SlackHandler
{
    use MonologCov12Sock, MonologCov12Boot;
    public function __construct()
    {
        SocketHandler::__construct('ssl://slack.com:443', Level::Debug, true, false, 0.0, 10.0, null, null);
        $this->setPrivate(SlackHandler::class, 'token', 'tok');
        $this->setPrivate(SlackHandler::class, 'slackRecord', new SlackRecord('#chan', 'bot', true, 'robot', false, true, ['extra']));
    }
}

final class MonologCov12Fleep extends FleepHookHandler
{
    use MonologCov12Sock;
    public function __construct()
    {
        SocketHandler::__construct('ssl://fleep.io:443', Level::Debug, true, false, 0.0, 10.0, null, null);
        $this->token = 'ftok';
    }
}

final class MonologCov12Flowdock extends FlowdockHandler
{
    use MonologCov12Sock;
    public function __construct()
    {
        SocketHandler::__construct('ssl://api.flowdock.com:443', Level::Debug, true, false, 0.0, 10.0, null, null);
        $this->apiToken = 'apitok';
    }
}

final class MonologCov12Telegram extends TelegramBotHandler
{
    use MonologCov12Boot;
    public array $messages = [];
    public function __construct()
    {
        AbstractHandler::__construct(Level::Debug, true);
        $this->setPrivate(TelegramBotHandler::class, 'apiKey', 'KEY');
        $this->setPrivate(TelegramBotHandler::class, 'channel', '@chan');
        $this->setParseMode(null)->disableWebPagePreview(null)->disableNotification(null)->splitLongMessages(false)->delayBetweenMessages(false)->setTopic(null);
    }
    protected function sendCurl(string $message): void { $this->messages[] = $message; }
    public function callSend(string $m): void { $this->send($m); }
    public function validate(string|bool $r): void { $this->validateApiResponse($r); }
    public function url(): string { return $this->getBotApiUrl(); }
    public function headers(): array { return $this->getCurlHeaders(); }
}

final class MonologCov12Loggly extends LogglyHandler
{
    public array $posts = [];
    public function __construct()
    {
        AbstractProcessingHandler::__construct(Level::Debug, true);
        $this->token = 'ltok';
    }
    protected function send(string $data, string $endpoint): void { $this->posts[] = [$endpoint, $data, $this->tag]; }
}

final class MonologCov12Webhook extends SlackWebhookHandler
{
    use MonologCov12Boot;
    public function __construct()
    {
        AbstractProcessingHandler::__construct(Level::Critical, true);
        $this->setPrivate(SlackWebhookHandler::class, 'webhookUrl', 'https://hooks.invalid/x');
        $this->setPrivate(SlackWebhookHandler::class, 'slackRecord', new SlackRecord('#c', 'u', true, null, true, false));
    }
}

final class MonologCov12Cube extends CubeHandler
{
    protected function connectUdp(): void {}
    protected function connectHttp(): void {}
}

final class MonologCov12Test extends TestCase
{
    private function rec(string $m = 'm', Level $l = Level::Info, array $ctx = [], array $extra = []): LogRecord
    {
        return new LogRecord(new \DateTimeImmutable('2024-01-02 03:04:05', new \DateTimeZone('UTC')), 'chan', $l, $m, $ctx, $extra);
    }

    public function testSlackHandlerWithoutConstructor(): void
    {
        $h = new MonologCov12Slack();
        $h->setFormatter(new LineFormatter("%message%"));
        $this->assertSame('tok', $h->getToken());
        $this->assertInstanceOf(SlackRecord::class, $h->getSlackRecord());
        $this->assertSame($h, $h->setChannel('#other'));
        $this->assertSame($h, $h->setUsername('u2'));
        $this->assertSame($h, $h->setIconEmoji(':ghost:'));
        $this->assertSame($h, $h->useShortAttachment(false));
        $this->assertSame($h, $h->includeContextAndExtra(true));
        $this->assertSame($h, $h->excludeFields(['context.secret']));
        $this->assertSame($h, $h->useAttachment(true));
        $h->handle($this->rec('hello', Level::Error, ['k' => 'v', 'secret' => 's'], ['e' => 1]));
        $this->assertCount(1, $h->sent);
        $this->assertStringStartsWith("POST /api/chat.postMessage HTTP/1.1\r\nHost: slack.com\r\n", $h->sent[0]);
        parse_str(substr($h->sent[0], strpos($h->sent[0], "\r\n\r\n") + 4), $q);
        $this->assertSame('tok', $q['token']);
        $this->assertSame('#other', $q['channel']);
        $att = json_decode($q['attachments'], true);
        $this->assertSame('danger', $att[0]['color']);
        $this->assertSame('hello', $att[0]['text']);
        $this->assertSame($h->getFormatter(), $h->getFormatter());
        $h->useAttachment(false);
        $h->handle($this->rec('plain', Level::Warning));
        parse_str(substr($h->sent[1], strpos($h->sent[1], "\r\n\r\n") + 4), $q);
        $this->assertSame('plain', $q['text']);
        $this->assertArrayNotHasKey('attachments', $q);
        $this->assertFalse($h->isConnected());
    }

    public function testFleepHandlerWithoutConstructor(): void
    {
        $h = new MonologCov12Fleep();
        $this->assertInstanceOf(LineFormatter::class, $h->getFormatter());
        $h->handle($this->rec('fleep msg'));
        $this->assertStringContainsString("POST /hook/ftok HTTP/1.1\r\nHost: fleep.io\r\n", $h->sent[0]);
        $this->assertStringContainsString('message=', $h->sent[0]);
        $this->assertFalse($h->isConnected());
    }

    public function testFlowdockHandlerWithoutConstructor(): void
    {
        $h = new MonologCov12Flowdock();
        try { $h->handle($this->rec()); $this->fail(); } catch (\InvalidArgumentException $e) { $this->assertStringContainsString('must be configured', $e->getMessage()); }
        try { $h->setFormatter(new LineFormatter()); $this->fail(); } catch (\InvalidArgumentException $e) { $this->assertStringContainsString('FlowdockFormatter', $e->getMessage()); }
        $this->assertSame($h, $h->setFormatter(new FlowdockFormatter('src', 'a@b.c')));
        $h->handle($this->rec('flow'));
        $this->assertStringContainsString('POST /v1/messages/team_inbox/apitok HTTP/1.1', $h->sent[0]);
        $this->assertStringContainsString('"source":"src"', $h->sent[0]);
    }

    private function telegram(): MonologCov12Telegram
    {
        $h = new MonologCov12Telegram();
        $h->setFormatter(new LineFormatter("%message%"));
        return $h;
    }

    public function testTelegramWithoutConstructor(): void
    {
        $h = $this->telegram();
        $this->assertSame('https://api.telegram.org/bot', $h->url());
        $this->assertSame([], $h->headers());
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
        $this->assertCount(1, $h->messages);
        $h->popProcessor();
        $h->messages = [];
        $h->callSend('   ');
        $this->assertSame(['   '], $h->messages);
    }

    public function testTelegramLongMessages(): void
    {
        $h = $this->telegram();
        $long = str_repeat('a', 5000);
        $h->callSend($long);
        $this->assertSame(1, count($h->messages));
        $this->assertSame(4096, strlen($h->messages[0]));
        $this->assertStringEndsWith('(…truncated)', $h->messages[0]);
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
        $h->messages = [];
        $h->callSend('<p>' . str_repeat('z', 4090) . '</p>tail');
        $this->assertGreaterThanOrEqual(2, count($h->messages));
        $h->messages = [];
        $h->splitLongMessages(false);
        $h->callSend('<b>trunc </b>' . str_repeat('<u>', 1400));
        $this->assertStringEndsWith('(…truncated)', $h->messages[0]);
    }

    public function testTelegramSettersAndApiResponse(): void
    {
        $h = $this->telegram();
        $this->assertSame($h, $h->setParseMode('MarkdownV2'));
        $this->assertSame($h, $h->disableWebPagePreview(true));
        $this->assertSame($h, $h->disableNotification(true));
        $this->assertSame($h, $h->splitLongMessages(true));
        $this->assertSame($h, $h->delayBetweenMessages(true));
        $this->assertSame($h, $h->setTopic(7));
        try { $h->setParseMode('Bogus'); $this->fail(); } catch (\InvalidArgumentException $e) { $this->assertStringContainsString('Unknown parseMode', $e->getMessage()); }
        $h->validate('{"ok":true}');
        foreach ([[false, 'No response'], ['not json', 'non-JSON'], ['{"ok":false,"description":"bad"}', 'bad'], ['{"ok":false}', 'Unknown error']] as [$in, $msg]) {
            try { $h->validate($in); $this->fail(); } catch (\RuntimeException $e) { $this->assertStringContainsString($msg, $e->getMessage()); }
        }
    }

    public function testCubeHandlerWriteBeforeConnect(): void
    {
        $udp = new MonologCov12Cube('udp://127.0.0.1:1234');
        try { $udp->handle($this->rec('cube', Level::Info, ['type' => 'custom', 'k' => 'v'])); $this->fail(); } catch (\LogicException $e) { $this->assertStringContainsString('No UDP socket could be opened', $e->getMessage()); }
        try { $udp->handle($this->rec('cube2')); $this->fail(); } catch (\LogicException $e) { $this->assertStringContainsString('No UDP socket', $e->getMessage()); }
        $http = new MonologCov12Cube('http://127.0.0.1:1234');
        try { $http->handle($this->rec('cube3')); $this->fail(); } catch (\LogicException $e) { $this->assertStringContainsString('No connection could be established', $e->getMessage()); }
    }

    public function testLogglyWithoutConstructor(): void
    {
        $l = new MonologCov12Loggly();
        $this->assertSame($l, $l->setTag('a'));
        $this->assertSame($l, $l->addTag(['b', 'a']));
        $l->handle($this->rec('single'));
        $this->assertSame('inputs', $l->posts[0][0]);
        $this->assertSame(['a', 'b'], array_values($l->posts[0][2]));
        $this->assertStringContainsString('"message":"single"', $l->posts[0][1]);
        $l->setTag([]);
        $this->assertSame($l, $l->setTag(''));
        $this->assertSame($l, $l->addTag(''));
        $this->assertSame($l, $l->addTag('only'));
        $l->handleBatch([]);
        $this->assertCount(1, $l->posts);
        $l->handleBatch([$this->rec('b1'), $this->rec('b2', Level::Error)]);
        $this->assertSame('bulk', $l->posts[1][0]);
        $this->assertSame(['only'], $l->posts[1][2]);
        $this->assertInstanceOf(\Monolog\Formatter\LogglyFormatter::class, $l->getFormatter());
    }

    public function testSlackWebhookWithoutConstructor(): void
    {
        $h = new MonologCov12Webhook();
        $this->assertSame('https://hooks.invalid/x', $h->getWebhookUrl());
        $this->assertInstanceOf(SlackRecord::class, $h->getSlackRecord());
        $f = new LineFormatter("%message%");
        $this->assertSame($h, $h->setFormatter($f));
        $this->assertSame($f, $h->getFormatter());
        $this->assertFalse($h->isHandling($this->rec('x', Level::Info)));
        $this->assertTrue($h->isHandling($this->rec('x', Level::Critical)));
    }

    public function testCubeUrlValidation(): void
    {
        foreach (['not a url', 'tcp://host', 'ftp://host:21'] as $bad) {
            try { new CubeHandler($bad); $this->fail(); } catch (\UnexpectedValueException $e) { $this->assertNotSame('', $e->getMessage()); }
        }
    }
}
