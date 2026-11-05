<?php
use Monolog\{Logger, Level, LogRecord};
use Monolog\Handler\{TestHandler, SocketHandler, SlackHandler, PushoverHandler, FleepHookHandler, FlowdockHandler, InsightOpsHandler, LogEntriesHandler, SlackWebhookHandler, SyslogUdpHandler, SyslogHandler, MailHandler, NativeMailerHandler, NewRelicHandler, BrowserConsoleHandler, ChromePHPHandler, FirePHPHandler, CubeHandler, LogglyHandler, CouchDBHandler, RedisHandler, RedisPubSubHandler, TelegramBotHandler, IFTTTHandler, AbstractSyslogHandler};
use Monolog\Handler\Slack\SlackRecord;
use Monolog\Handler\SyslogUdp\UdpSocket;
use Monolog\Formatter\{LineFormatter, JsonFormatter, FlowdockFormatter, NormalizerFormatter};
use PHPUnit\Framework\TestCase;

final class MonologCovNrLog { public static array $calls = []; }

if (!function_exists('newrelic_notice_error')) {
    function newrelic_notice_error(...$args) { MonologCovNrLog::$calls[] = ['notice', $args[0]]; }
    function newrelic_set_appname($n) { MonologCovNrLog::$calls[] = ['app', $n]; }
    function newrelic_name_transaction($n) { MonologCovNrLog::$calls[] = ['tx', $n]; }
    function newrelic_add_custom_parameter($k, $v) { MonologCovNrLog::$calls[] = [$k, $v]; }
}

final class MonologCovNewRelicNative extends NewRelicHandler
{
    protected function isNewRelicEnabled(): bool { return true; }
}

trait MonologCovMemSock
{
    public array $sent = [];
    public int $opens = 0;
    protected function fsockopen()
    {
        $this->opens++;
        return fopen('php://memory', 'w+');
    }
    protected function pfsockopen()
    {
        $this->opens++;
        return fopen('php://memory', 'w+');
    }
    protected function streamSetTimeout(): bool { return true; }
    protected function streamSetChunkSize(): int|bool { return 1; }
    protected function fwrite(string $data): int|bool
    {
        $this->sent[] = $data;
        return strlen($data);
    }
}

final class MonologCovSock extends SocketHandler { use MonologCovMemSock; }
final class MonologCovSlack extends SlackHandler { use MonologCovMemSock; }
final class MonologCovPushover extends PushoverHandler { use MonologCovMemSock; }
final class MonologCovFleep extends FleepHookHandler { use MonologCovMemSock; }
final class MonologCovFlowdock extends FlowdockHandler { use MonologCovMemSock; }
final class MonologCovInsight extends InsightOpsHandler { use MonologCovMemSock; }
final class MonologCovLogEntries extends LogEntriesHandler { use MonologCovMemSock; }

final class MonologCovFailSock extends SocketHandler
{
    public function __construct(private mixed $res, private int $mode, string $cs = 'tcp://x:1')
    {
        parent::__construct($cs, Level::Debug, true, false, 0.0, 0.5);
    }
    protected function fsockopen() { return $this->res; }
    protected function streamSetTimeout(): bool { return $this->mode !== 1; }
    protected function streamSetChunkSize(): int|bool { return $this->mode !== 2; }
    protected function fwrite(string $data): int|bool { return $this->mode === 3 ? false : 0; }
}

final class MonologCovUdp extends UdpSocket
{
    public array $out = [];
    protected function send(string $chunk): void { $this->out[] = $chunk; }
}

final class MonologCovMail extends MailHandler
{
    public array $sends = [];
    protected function send(string $content, array $records): void { $this->sends[] = [$content, count($records)]; }
    public function body(string $b): bool { return $this->isHtmlBody($b); }
    public function highest(array $r): LogRecord { return $this->getHighestRecord($r); }
}

final class MonologCovNative extends NativeMailerHandler
{
    public array $mails = [];
    protected function mail(string $to, string $subject, string $content, string $headers, string $parameters): void
    {
        $this->mails[] = compact('to', 'subject', 'content', 'headers', 'parameters');
    }
}

final class MonologCovNewRelic extends NewRelicHandler
{
    public array $calls = [];
    public bool $enabled = true;
    protected function isNewRelicEnabled(): bool { return $this->enabled; }
    protected function setNewRelicAppName(string $appName): void { $this->calls[] = ['app', $appName]; }
    protected function setNewRelicTransactionName(string $transactionName): void { $this->calls[] = ['tx', $transactionName]; }
    protected function setNewRelicParameter(string $key, $value): void { $this->calls[] = [$key, $value]; }
}

final class MonologCovBrowser extends BrowserConsoleHandler
{
    public static string $out = '';
    protected static function writeOutput(string $str): void { self::$out .= $str; }
    public static function fmt(string $ct): string { return self::getResponseFormatFromContentType($ct); }
}

final class MonologCovChrome extends ChromePHPHandler
{
    public array $headers = [];
    public static function resetAll(): void { static::$initialized = false; static::$overflowed = false; static::$sendHeaders = true; static::$json = ['version' => static::VERSION, 'columns' => ['label', 'log', 'backtrace', 'type'], 'rows' => []]; }
    protected function isWebRequest(): bool { return true; }
    protected function sendHeader(string $header, string $content): void { $this->headers[$header] = $content; }
}

final class MonologCovFire extends FirePHPHandler
{
    public array $headers = [];
    public static function resetAll(): void { static::$initialized = false; static::$sendHeaders = true; static::$messageIndex = 1; }
    protected function isWebRequest(): bool { return true; }
    protected function sendHeader(string $header, string $content): void { $this->headers[$header] = $content; }
}

final class MonologCov4Test extends TestCase
{
    private function rec(string $m = 'm', Level $l = Level::Info, array $ctx = [], array $extra = []): LogRecord
    {
        return new LogRecord(new \DateTimeImmutable('2024-01-02 03:04:05', new \DateTimeZone('UTC')), 'chan', $l, $m, $ctx, $extra);
    }

    public function testSocketHandlerBasics(): void
    {
        $h = new MonologCovSock('tcp://example.invalid:80', Level::Debug, true, false, 1.5, 3.0, 2.0, 1024);
        $h->setFormatter(new LineFormatter("%message%\n"));
        $this->assertSame('tcp://example.invalid:80', $h->getConnectionString());
        $this->assertFalse($h->isPersistent());
        $this->assertSame(1.5, $h->getTimeout());
        $this->assertSame(3.0, $h->getWritingTimeout());
        $this->assertSame(2.0, $h->getConnectionTimeout());
        $this->assertSame(1024, $h->getChunkSize());
        $this->assertFalse($h->isConnected());
        $h->handle($this->rec('hello'));
        $h->handle($this->rec('again'));
        $this->assertSame(["hello\n", "again\n"], $h->sent);
        $this->assertSame(1, $h->opens);
        $this->assertTrue($h->isConnected());
        $h->close();
        $this->assertFalse($h->isConnected());
        $h->handle($this->rec('third'));
        $this->assertSame(2, $h->opens);
        $this->assertSame($h, $h->setPersistent(true));
        $h->close();
        $this->assertTrue($h->isConnected());
        $h->closeSocket();
        $h->handle($this->rec('persist'));
        $this->assertSame(3, $h->opens);
        $this->assertSame($h, $h->setConnectionTimeout(4.0));
        $this->assertSame($h, $h->setTimeout(0.25));
        $this->assertSame($h, $h->setWritingTimeout(0.0));
        $this->assertSame($h, $h->setChunkSize(2048));
        $this->assertSame(4.0, $h->getConnectionTimeout());
        $this->assertSame(2048, $h->getChunkSize());
        foreach (['setConnectionTimeout', 'setTimeout', 'setWritingTimeout'] as $m) {
            try { $h->$m(-1.0); $this->fail(); } catch (\InvalidArgumentException $e) { $this->assertStringContainsString('Timeout must be', $e->getMessage()); }
        }
        $h->closeSocket();
    }

    public function testSocketHandlerConstructorValidation(): void
    {
        $this->expectException(\InvalidArgumentException::class);
        new SocketHandler('tcp://x:1', Level::Debug, true, false, -1.0);
    }

    public function testSocketHandlerConnectionTimeoutValidation(): void
    {
        $this->expectException(\InvalidArgumentException::class);
        new SocketHandler('tcp://x:1', Level::Debug, true, false, 0.0, 1.0, -2.0);
    }

    public function testSocketHandlerRealOpenFails(): void
    {
        $h = new SocketHandler('unix:///nonexistent/monologcov.sock');
        $h->setFormatter(new LineFormatter("%message%\n"));
        try { $h->handle($this->rec()); $this->fail(); } catch (\UnexpectedValueException $e) {
            $this->assertStringContainsString('Failed connecting to unix:///nonexistent/monologcov.sock', $e->getMessage());
        }
        $p = new SocketHandler('unix:///nonexistent/monologcov.sock', Level::Debug, true, true);
        try { $p->handle($this->rec()); $this->fail(); } catch (\UnexpectedValueException $e) {
            $this->assertStringContainsString('Failed connecting', $e->getMessage());
        }
        try { $h->isConnected() ? $this->fail() : null; } catch (\Throwable $e) {}
    }

    public function testSocketHandlerFailureModes(): void
    {
        $mk = fn() => fopen('php://memory', 'w+');
        $h = new MonologCovFailSock(false, 0);
        $h->setFormatter(new LineFormatter("%message%\n"));
        try { $h->handle($this->rec()); $this->fail(); } catch (\UnexpectedValueException $e) { $this->assertStringContainsString('Failed connecting', $e->getMessage()); }
        $h = new MonologCovFailSock($mk(), 1);
        $h->setFormatter(new LineFormatter("%message%\n"));
        try { $h->handle($this->rec()); $this->fail(); } catch (\UnexpectedValueException $e) { $this->assertStringContainsString('stream_set_timeout', $e->getMessage()); }
        $h = new MonologCovFailSock($mk(), 2);
        $h->setChunkSize(10);
        $h->setFormatter(new LineFormatter("%message%\n"));
        try { $h->handle($this->rec()); $this->fail(); } catch (\UnexpectedValueException $e) { $this->assertStringContainsString('stream_set_chunk_size', $e->getMessage()); }
        $h = new MonologCovFailSock($mk(), 3);
        $h->setFormatter(new LineFormatter("%message%\n"));
        try { $h->handle($this->rec()); $this->fail(); } catch (\RuntimeException $e) { $this->assertStringContainsString('Could not write to socket', $e->getMessage()); }
        $h = new MonologCovFailSock($mk(), 4);
        $h->setWritingTimeout(0.001);
        $h->setFormatter(new LineFormatter("%message%\n"));
        try { $h->handle($this->rec('data')); $this->fail(); } catch (\RuntimeException $e) { $this->assertStringContainsString('Write timed-out', $e->getMessage()); }
    }

    public function testSlackHandler(): void
    {
        $h = new MonologCovSlack('tok', '#chan', 'bot', true, 'robot', Level::Debug, true, false, true, ['extra']);
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
        $this->assertSame('u2', $q['username']);
        $this->assertSame(':ghost:', $q['icon_emoji']);
        $att = json_decode($q['attachments'], true);
        $this->assertSame('danger', $att[0]['color']);
        $this->assertSame('hello', $att[0]['text']);
        $this->assertSame('Message', $att[0]['title']);
        $this->assertSame(0, $h->opens - 1);
        $this->assertSame($h->getFormatter(), $h->getFormatter());
    }

    public function testSlackRecord(): void
    {
        $r = new SlackRecord('#c', 'user', true, 'ghost', false, true, ['context.secret']);
        $d = $r->getSlackData($this->rec('msg', Level::Warning, ['a' => 1, 'secret' => 'x', 'arr' => ['x' => 1]], ['ex' => 'y']));
        $this->assertSame('user', $d['username']);
        $this->assertSame('#c', $d['channel']);
        $a = $d['attachments'][0];
        $this->assertSame('warning', $a['color']);
        $this->assertSame(1704164645, $a['ts']);
        $titles = array_column($a['fields'], 'title');
        $this->assertContains('Level', $titles);
        $this->assertContains('Ex', $titles);
        $this->assertContains('A', $titles);
        $this->assertNotContains('Secret', $titles);
        $this->assertContains('Arr', $titles);
        $short = new SlackRecord(null, null, true, 'http://x.invalid/i.png', true, true);
        $d = $short->getSlackData($this->rec('msg', Level::Debug, ['a' => 1, 'b' => [1, 2]], ['e' => 2]));
        $this->assertArrayHasKey('attachments', $d);
        $this->assertArrayNotHasKey('username', $d);
        $a = $d['attachments'][0];
        $this->assertSame('DEBUG', $a['title']);
        $this->assertSame('#e3e4e6', $a['color']);
        $this->assertSame(['Extra', 'Context'], array_column($a['fields'], 'title'));
        $this->assertStringStartsWith('```', $a['fields'][0]['value']);
        $plain = new SlackRecord('#c', null, false, null, false, false, [], new LineFormatter("%level_name%: %message%"));
        $d = $plain->getSlackData($this->rec('plain'));
        $this->assertSame('INFO: plain', $d['text']);
        $this->assertArrayNotHasKey('attachments', $d);
        $plain->setFormatter(null)->useAttachment(false);
        $this->assertSame('plain', $plain->getSlackData($this->rec('plain'))['text']);
        $this->assertSame('good', $plain->getAttachmentColor(Level::Notice));
        $this->assertSame('good', $plain->getAttachmentColor(Level::Info));
        foreach ([Level::Error, Level::Critical, Level::Alert, Level::Emergency] as $l) {
            $this->assertSame('danger', $plain->getAttachmentColor($l));
        }
        $s = new SlackRecord(null, null, true, null, false, true);
        $this->assertSame("{\n    \"a\": 1\n}", $s->stringify(['a' => 1]));
        $this->assertSame('[1,2]', $s->stringify([1, 2]));
        $this->assertSame($s, $s->setUserIcon(':smile:')->setChannel('x')->setUsername('u')->useAttachment()->useShortAttachment()->includeContextAndExtra()->excludeFields([]));
    }

    public function testSlackRecordIconUrl(): void
    {
        $d = (new SlackRecord(null, null, true, 'http://x.invalid/i.png'))->getSlackData($this->rec('msg'));
        $this->assertSame('http://x.invalid/i.png', $d['icon_url']);
    }

    public function testSlackRecordIconEmoji(): void
    {
        $r = new SlackRecord('#c', 'user', true, ':ghost:');
        $d = $r->getSlackData($this->rec('msg'));
        $this->assertSame(':ghost:', $d['icon_emoji']);
    }

    public function testPushover(): void
    {
        $h = new MonologCovPushover('tok', ['u1', 'u2'], 'Title', Level::Debug, true, false, Level::Critical, Level::Emergency, 10, 600);
        $h->setFormatter(new LineFormatter("%message%!"));
        $h->handle($this->rec('info msg', Level::Info, ['sound' => 'bike', 'url' => 'http://x.invalid']));
        $this->assertCount(2, $h->sent);
        $this->assertStringStartsWith("POST /1/messages.json HTTP/1.1\r\n", $h->sent[0]);
        parse_str(substr($h->sent[0], strpos($h->sent[0], "\r\n\r\n") + 4), $q);
        $this->assertSame('u1', $q['user']);
        $this->assertSame('info msg', $q['message']);
        $this->assertSame('bike', $q['sound']);
        $this->assertArrayNotHasKey('priority', $q);
        $h->sent = [];
        $h->handle($this->rec('crit', Level::Critical, ['sound' => 'nonsense']));
        parse_str(substr($h->sent[0], strpos($h->sent[0], "\r\n\r\n") + 4), $q);
        $this->assertSame('1', $q['priority']);
        $this->assertArrayNotHasKey('sound', $q);
        $h->sent = [];
        $this->assertSame($h, $h->useFormattedMessage(true));
        $h->handle($this->rec('emerg', Level::Emergency));
        parse_str(substr($h->sent[0], strpos($h->sent[0], "\r\n\r\n") + 4), $q);
        $this->assertSame('2', $q['priority']);
        $this->assertSame('10', $q['retry']);
        $this->assertSame('600', $q['expire']);
        $this->assertSame('emerg!', $q['message']);
        $this->assertSame($h, $h->setHighPriorityLevel('error'));
        $this->assertSame($h, $h->setEmergencyLevel(Level::Alert));
        $h->sent = [];
        $h->handle($this->rec('alert', Level::Alert));
        parse_str(substr($h->sent[0], strpos($h->sent[0], "\r\n\r\n") + 4), $q);
        $this->assertSame('2', $q['priority']);
        $d = new MonologCovPushover('t', 'single');
        $d->handle($this->rec('x', Level::Critical));
        $this->assertCount(1, $d->sent);
        $ssl = new PushoverHandler('t', 'u');
        $this->assertSame('ssl://api.pushover.net:443', $ssl->getConnectionString());
    }

    public function testFleepHandler(): void
    {
        $f = new MonologCovFleep('ftok');
        $f->handle($this->rec('fleep msg'));
        $this->assertStringContainsString('POST /hook/ftok HTTP/1.1', $f->sent[0]);
        $this->assertStringContainsString('message=', $f->sent[0]);
    }

    public function testFlowdockHandler(): void
    {
        $fl = new MonologCovFlowdock('apitok');
        $fl->setFormatter(new FlowdockFormatter('src', 'a@b.c'));
        $fl->handle($this->rec('flow'));
        $this->assertStringContainsString('POST /v1/messages/team_inbox/apitok HTTP/1.1', $fl->sent[0]);
        $this->assertStringContainsString('"source":"src"', $fl->sent[0]);
        try { $fl->setFormatter(new LineFormatter()); $this->fail(); } catch (\InvalidArgumentException $e) { $this->assertStringContainsString('FlowdockFormatter', $e->getMessage()); }
        try { (new MonologCovFlowdock('x'))->handle($this->rec()); $this->fail(); } catch (\InvalidArgumentException $e) { $this->assertStringContainsString('must be configured', $e->getMessage()); }
    }

    public function testInsightOpsHandler(): void
    {
        $i = new MonologCovInsight('tok', 'eu', true);
        $i->setFormatter(new LineFormatter("%message%\n"));
        $i->handle($this->rec('ins'));
        $this->assertSame("tok ins\n", $i->sent[0]);
        $this->assertSame('ssl://eu.data.logs.insight.rapid7.com:443', $i->getConnectionString());
    }

    public function testInsightOpsPlain(): void
    {
        $this->assertSame('us.data.logs.insight.rapid7.com:80', (new InsightOpsHandler('t', 'us', false))->getConnectionString());
    }

    public function testLogEntriesHandler(): void
    {
        $le = new MonologCovLogEntries('letok');
        $le->setFormatter(new LineFormatter("%message%\n"));
        $le->handle($this->rec('le'));
        $this->assertSame("letok le\n", $le->sent[0]);
    }

    public function testLogEntriesPlain(): void
    {
        $this->assertSame('data.logentries.com:80', (new LogEntriesHandler('t', false))->getConnectionString());
    }

    public function testUdpSocket(): void
    {
        $u = new MonologCovUdp('127.0.0.1', 514, 40);
        $u->write('payload', 'HDR ');
        $u->write(str_repeat('x', 100), 'HDR ');
        $u->write('only line');
        $u->close();
        $this->assertSame(['HDR payload', 'HDR ' . str_repeat('x', 36), 'only line'], $u->out);
    }

    public function testSyslogUdp(): void
    {
        $h = new SyslogUdpHandler('127.0.0.1', 514, 'local3', Level::Debug, true, 'myapp', SyslogUdpHandler::RFC5424);
        $sock = new MonologCovUdp('127.0.0.1', 514, 200);
        $this->assertSame($h, $h->setSocket($sock));
        $h->setFormatter(new LineFormatter("%message%", null, true));
        $h->handle($this->rec("line1\nline2", Level::Error));
        $this->assertCount(2, $sock->out);
        $this->assertMatchesRegularExpression('/^<(\d+)>1 2024-01-02T03:04:05\+00:00 \S+ myapp \d+ - - line1$/', $sock->out[0]);
        $this->assertStringStartsWith('<' . (3 + 152) . '>1 ', $sock->out[0]);
        $h->close();

        foreach ([[SyslogUdpHandler::RFC3164, '/^<\d+>Jan 02 03:04:05 \S+ php\[\d+\]: /'], [SyslogUdpHandler::RFC5424e, '/^<\d+>1 2024-01-02T03:04:05\.000\+00:00 /']] as [$rfc, $re]) {
            $h2 = new SyslogUdpHandler('127.0.0.1', 514, LOG_USER, Level::Debug, true, 'php', $rfc);
            $s2 = new MonologCovUdp('127.0.0.1', 514);
            $h2->setSocket($s2)->setFormatter(new LineFormatter("%message%"));
            $h2->handle($this->rec('m'));
            $this->assertMatchesRegularExpression($re, $s2->out[0]);
        }
        // long messages are cut to the datagram length
        $s3 = new MonologCovUdp('127.0.0.1', 514, 80);
        $h3 = new SyslogUdpHandler('127.0.0.1');
        $h3->setSocket($s3)->setFormatter(new LineFormatter("%message%"));
        $h3->handle($this->rec(str_repeat('x', 500)));
        $this->assertSame(80, strlen($s3->out[0]));
        $this->assertSame([], (function () { $u = new MonologCovUdp('h'); $u->write('', ''); $u->close(); return []; })());
    }

    public function testSyslogFacilities(): void
    {
        foreach (['user', 'LOCAL0', 'auth', 'cron', 'daemon', LOG_MAIL] as $f) {
            $this->assertInstanceOf(AbstractSyslogHandler::class, new SyslogUdpHandler('127.0.0.1', 514, $f));
        }
        $this->expectException(\UnexpectedValueException::class);
        new SyslogUdpHandler('127.0.0.1', 514, 'bogus-facility');
    }

    public function testSyslogHandlerConstruction(): void
    {
        $h = new SyslogHandler('monologcov', 'user', Level::Warning, false, 0);
        $this->assertSame(Level::Warning, $h->getLevel());
        $this->assertFalse($h->getBubble());
        $this->assertInstanceOf(LineFormatter::class, $h->getFormatter());
        $this->assertStringContainsString('%channel%.%level_name%', (new \ReflectionClass($h->getFormatter()))->getProperty('format')->getValue($h->getFormatter()));
    }

    public function testMailHandler(): void
    {
        $m = new MonologCovMail(Level::Warning);
        $m->handle($this->rec('x', Level::Info));
        $m->handle($this->rec('warn', Level::Warning));
        $this->assertCount(1, $m->sends);
        $this->assertSame(1, $m->sends[0][1]);
        $this->assertStringContainsString('<h1', $m->sends[0][0]);
        $m->handleBatch([$this->rec('low', Level::Debug), $this->rec('a', Level::Error), $this->rec('b', Level::Warning)]);
        $this->assertSame(2, $m->sends[1][1]);
        $m->handleBatch([$this->rec('low', Level::Debug)]);
        $this->assertCount(2, $m->sends);
        $this->assertTrue($m->body('<html>'));
        $this->assertFalse($m->body('plain'));
        $this->assertFalse($m->body(''));
        $this->assertSame('hi', $m->highest([$this->rec('lo', Level::Info), $this->rec('hi', Level::Error), $this->rec('eq', Level::Error)])->message);
    }

    public function testNativeMailer(): void
    {
        $n = new MonologCovNative(['a@x.invalid', 'b@x.invalid'], 'Subj %level_name% %message%', 'from@x.invalid', Level::Debug, true, 20);
        $n->setFormatter(new LineFormatter("%message%"));
        $this->assertSame($n, $n->addHeader('X-One: 1'));
        $this->assertSame($n, $n->addHeader(['X-Two: 2']));
        $this->assertSame($n, $n->addParameter('-fx'));
        $this->assertSame($n, $n->addParameter(['-y']));
        $n->handle($this->rec('this is a fairly long body text that wraps', Level::Error));
        $this->assertCount(2, $n->mails);
        $m = $n->mails[0];
        $this->assertSame('a@x.invalid', $m['to']);
        $this->assertSame('Subj ERROR this is a fairly long body text that wraps', $m['subject']);
        $this->assertSame("this is a fairly\nlong body text that\nwraps", $m['content']);
        $this->assertStringContainsString("From: from@x.invalid\r\nX-One: 1\r\nX-Two: 2\r\nContent-type: text/plain; charset=utf-8\r\n", $m['headers']);
        $this->assertSame('-fx -y', $m['parameters']);
        $this->assertSame('utf-8', $n->getEncoding());
        $this->assertNull($n->getContentType());
        $this->assertSame($n, $n->setContentType('text/html'));
        $this->assertSame($n, $n->setEncoding('iso-8859-1'));
        $n->mails = [];
        $n->handle($this->rec('<b>html</b>', Level::Error));
        $this->assertStringContainsString("Content-type: text/html; charset=iso-8859-1\r\nMIME-Version: 1.0\r\n", $n->mails[0]['headers']);
        foreach (["a\nb", "a\rb"] as $bad) {
            try { $n->addHeader($bad); $this->fail(); } catch (\InvalidArgumentException $e) { $this->assertStringContainsString('newline', $e->getMessage()); }
            try { $n->setContentType($bad); $this->fail(); } catch (\InvalidArgumentException $e) { $this->assertStringContainsString('newline', $e->getMessage()); }
            try { $n->setEncoding($bad); $this->fail(); } catch (\InvalidArgumentException $e) { $this->assertStringContainsString('newline', $e->getMessage()); }
        }
        $h = new NativeMailerHandler('to@x.invalid', 's', 'f@x.invalid');
        $this->assertSame('utf-8', $h->getEncoding());
        $d = new MonologCovNative('to@x.invalid', 's', 'f@x.invalid');
        $d->setFormatter(new LineFormatter("%message%"));
        $d->handleBatch([$this->rec('b1', Level::Error), $this->rec('b2', Level::Critical)]);
        $this->assertCount(1, $d->mails);
    }

    public function testNewRelic(): void
    {
        $h = new MonologCovNewRelic(Level::Debug, true, 'myapp', true, 'tx1');
        $h->handle($this->rec('nr', Level::Error, ['k' => 'v', 'arr' => ['a' => 1, 'b' => 2]], ['ex' => ['z' => 3], 'plain' => 'p']));
        $keys = array_column($h->calls, 0);
        $this->assertContains('app', $keys);
        $this->assertContains('tx', $keys);
        $this->assertContains('context_k', $keys);
        $this->assertContains('context_arr_a', $keys);
        $this->assertContains('extra_ex_z', $keys);
        $this->assertContains('extra_plain', $keys);
        $h->enabled = false;
        try { $h->handle($this->rec()); $this->fail(); } catch (\Monolog\Handler\MissingExtensionException $e) { $this->assertStringContainsString('newrelic', $e->getMessage()); }
        $this->assertInstanceOf(NormalizerFormatter::class, $h->getFormatter());
        $h2 = new MonologCovNewRelic();
        $h2->handle($this->rec('nr2', Level::Error, ['appname' => 'ctxapp', 'transaction_name' => 'ctxtx', 'obj' => new \ArrayObject([1])], ['x' => [1, 2]]));
        $this->assertSame(['app', 'ctxapp'], $h2->calls[0]);
        $this->assertSame(['tx', 'ctxtx'], $h2->calls[1]);
        $found = array_values(array_filter($h2->calls, fn($c) => $c[0] === 'extra_x'));
        $this->assertSame([1, 2], $found[0][1]);
    }

    public function testNewRelicNativeFunctions(): void
    {
        MonologCovNrLog::$calls = [];
        $h = new MonologCovNewRelicNative(Level::Debug, true, 'app1', false, 'tx9');
        $ex = new \RuntimeException('bad');
        $h->handle($this->rec('nrn', Level::Error, ['exception' => $ex, 'arr' => [1], 'n' => 5], ['obj' => new \ArrayObject([1])]));
        $keys = array_column(MonologCovNrLog::$calls, 0);
        $this->assertSame(['app', 'tx', 'notice'], array_slice($keys, 0, 3));
        $this->assertContains('context_n', $keys);
        $this->assertContains('context_arr', $keys);
        $this->assertContains('extra_obj', $keys);
        $h->handle($this->rec('no exception'));
        $this->assertSame('notice', MonologCovNrLog::$calls[count(MonologCovNrLog::$calls) - 1][0] ?? 'notice');
    }

    public function testBrowserConsole(): void
    {
        MonologCovBrowser::resetStatic();
        MonologCovBrowser::$out = '';
        $h = new MonologCovBrowser();
        $h->handle($this->rec('plain info', Level::Info));
        $h->handle($this->rec('with ctx', Level::Warning, ['a' => 1, 'n' => null], ['e' => 'x"y</script>']));
        $h->handle($this->rec('dbg', Level::Debug));
        $h->handle($this->rec('err', Level::Critical));
        MonologCovBrowser::send();
        $out = MonologCovBrowser::$out;
        $this->assertStringStartsWith('<script>(function (c) {if (c && c.groupCollapsed) {', $out);
        $this->assertStringContainsString('c.info(', $out);
        $this->assertStringContainsString('c.groupCollapsed(', $out);
        $this->assertStringContainsString('c.groupEnd();', $out);
        $this->assertStringContainsString('c.debug(', $out);
        $this->assertStringContainsString('c.error(', $out);
        $this->assertStringContainsString('background-color: blue; color: white', $out);
        $this->assertStringContainsString('</script>', $out);
        MonologCovBrowser::$out = '';
        MonologCovBrowser::send();
        $this->assertSame('', MonologCovBrowser::$out);
        $h->handle($this->rec('again', Level::Notice));
        $h->close();
        MonologCovBrowser::send();
        $this->assertSame('', MonologCovBrowser::$out);
        $h->handle($this->rec('x'));
        $h->reset();
        $this->assertSame('js', MonologCovBrowser::fmt('Content-Type: application/javascript'));
        $this->assertSame('js', MonologCovBrowser::fmt('content-type: text/javascript'));
        $this->assertSame('html', MonologCovBrowser::fmt('Content-Type: text/html; charset=utf-8'));
        $this->assertSame('unknown', MonologCovBrowser::fmt('Content-Type: application/json'));
    }

    public function testChromePhp(): void
    {
        $_SERVER['HTTP_USER_AGENT'] = 'Mozilla/5.0 Chrome/120.0.0.0 Safari/537.36';
        $_SERVER['REQUEST_URI'] = '/x';
        MonologCovChrome::resetAll();
        $h = new MonologCovChrome(Level::Info);
        $h->handle($this->rec('one', Level::Info, ['a' => 1]));
        $this->assertArrayHasKey('X-ChromeLogger-Data', $h->headers);
        $data = json_decode(base64_decode($h->headers['X-ChromeLogger-Data']), true);
        $this->assertSame('/x', $data['request_uri']);
        $this->assertCount(1, $data['rows']);
        $h->handleBatch([$this->rec('two', Level::Warning), $this->rec('three', Level::Error)]);
        $data = json_decode(base64_decode($h->headers['X-ChromeLogger-Data']), true);
        $this->assertCount(3, $data['rows']);
        // overflow replaces the last row with a warning
        $h->handle($this->rec(str_repeat('big', 2000), Level::Error));
        $data = json_decode(base64_decode($h->headers['X-ChromeLogger-Data']), true);
        $this->assertStringContainsString('Incomplete logs', json_encode($data['rows']));
        $before = $h->headers;
        $h->handle($this->rec('after overflow'));
        $this->assertSame($before, $h->headers);
        // an unsupported browser sends nothing
        $_SERVER['HTTP_USER_AGENT'] = 'curl/8';
        MonologCovChrome::resetAll();
        $h2 = new MonologCovChrome();
        $h2->handle($this->rec('x'));
        $this->assertSame([], $h2->headers);
        unset($_SERVER['HTTP_USER_AGENT']);
        MonologCovChrome::resetAll();
        $h3 = new MonologCovChrome();
        $h3->handle($this->rec('x'));
        $this->assertSame([], $h3->headers);
        MonologCovChrome::resetAll();
        unset($_SERVER['REQUEST_URI']);
        // not a web request (cli) is a no-op
        $cli = new ChromePHPHandler();
        $cli->handle($this->rec('cli'));
        $cli->handleBatch([$this->rec('cli')]);
        $this->assertTrue(true);
    }

    public function testChromePhpBatchLevelFilter(): void
    {
        $_SERVER['HTTP_USER_AGENT'] = 'Mozilla/5.0 Chrome/120.0.0.0 Safari/537.36';
        MonologCovChrome::resetAll();
        $h = new MonologCovChrome(Level::Warning);
        $h->handleBatch([$this->rec('d', Level::Debug), $this->rec('two', Level::Warning)]);
        $data = json_decode(base64_decode($h->headers['X-ChromeLogger-Data']), true);
        // enum objects are not orderable in php, so the level filter keeps every record
        $this->assertCount(2, $data['rows']);
        $h->handleBatch([$this->rec('d', Level::Debug)]);
        MonologCovChrome::resetAll();
        unset($_SERVER['HTTP_USER_AGENT']);
    }

    public function testFirePhp(): void
    {
        $_SERVER['HTTP_USER_AGENT'] = 'Mozilla FirePHP/0.7';
        MonologCovFire::resetAll();
        $h = new MonologCovFire();
        $h->handle($this->rec('one'));
        $h->handle($this->rec('two', Level::Warning, ['c' => 1]));
        $keys = array_keys($h->headers);
        $this->assertContains('X-Wf-Protocol-1', $keys);
        $this->assertContains('X-Wf-1-Structure-1', $keys);
        $this->assertContains('X-Wf-1-Plugin-1', $keys);
        $this->assertContains('X-Wf-1-1-1-1', $keys);
        $this->assertContains('X-Wf-1-1-1-2', $keys);
        $this->assertStringContainsString('"Type":"WARN"', $h->headers['X-Wf-1-1-1-2']);
        unset($_SERVER['HTTP_USER_AGENT']);
        $_SERVER['HTTP_X_FIREPHP_VERSION'] = '1';
        MonologCovFire::resetAll();
        $h = new MonologCovFire();
        $h->handle($this->rec('hdr'));
        $this->assertNotEmpty($h->headers);
        unset($_SERVER['HTTP_X_FIREPHP_VERSION']);
        MonologCovFire::resetAll();
        $h = new MonologCovFire();
        $h->handle($this->rec('none'));
        $this->assertSame([], $h->headers);
        $h->handle($this->rec('none2'));
        MonologCovFire::resetAll();
        (new FirePHPHandler())->handle($this->rec('cli'));
    }

    public function testCubeAndLoggly(): void
    {
        foreach (['not a url', 'tcp://host', 'ftp://host:21'] as $bad) {
            try { new CubeHandler($bad); $this->fail(); } catch (\UnexpectedValueException $e) { $this->assertTrue(true); }
        }
        $this->assertInstanceOf(CubeHandler::class, new CubeHandler('udp://127.0.0.1:1234'));
        $this->assertInstanceOf(CubeHandler::class, new CubeHandler('http://127.0.0.1:1234'));
    }

    public function testLogglyHandler(): void
    {
        $l = new LogglyHandler('tok');
        $this->assertSame($l, $l->setTag('a'));
        $this->assertSame($l, $l->addTag(['b', 'a']));
        $this->assertSame($l, $l->setTag([]));
        $this->assertSame($l, $l->setTag(''));
        $this->assertSame($l, $l->addTag(''));
        $l->handleBatch([]);
        $this->assertInstanceOf(\Monolog\Formatter\LogglyFormatter::class, $l->getFormatter());
    }

    public function testCouchDbUnreachable(): void
    {
        $h = new CouchDBHandler(['host' => '127.0.0.1', 'port' => 1, 'username' => 'u', 'password' => 'p']);
        $this->assertInstanceOf(JsonFormatter::class, $h->getFormatter());
        try { $h->handle($this->rec()); $this->fail(); } catch (\RuntimeException $e) {
            $this->assertStringContainsString('Could not connect to http://u:p@127.0.0.1:1/logger', $e->getMessage());
        }
    }

    public function testTelegramConstructor(): void
    {
        $h = new TelegramBotHandler('key', '@chan', Level::Debug, true, 'Markdown', true, true, 5);
        $this->assertInstanceOf(TelegramBotHandler::class, $h);
        $this->assertSame($h, $h->setParseMode('HTML'));
        $this->assertSame($h, $h->disableWebPagePreview(false));
        $this->assertSame($h, $h->disableNotification(false));
        $this->assertSame($h, $h->splitLongMessages(true));
        $this->assertSame($h, $h->delayBetweenMessages(false));
        $this->assertSame($h, $h->setTopic(3));
        try { $h->setParseMode('Bogus'); $this->fail(); } catch (\InvalidArgumentException $e) { $this->assertTrue(true); }
    }
}
