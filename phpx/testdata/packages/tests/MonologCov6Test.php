<?php
foreach (['Gelf', 'MongoBson', 'MongoDriver', 'Mongo', 'Couch', 'Elastica', 'Elasticsearch', 'ElasticsearchEx', 'Elastic8'] as $stub) {
    require_once __DIR__ . '/MonologCovStub' . $stub . '.php';
}

use Monolog\{Logger, Level, LogRecord, ErrorHandler};
use Monolog\Handler\{GelfHandler, MongoDBHandler, DoctrineCouchDBHandler, ElasticaHandler, ElasticsearchHandler, SocketHandler, PushoverHandler, InsightOpsHandler, LogEntriesHandler, BrowserConsoleHandler, NativeMailerHandler, TestHandler};
use Monolog\Formatter\{GelfMessageFormatter, MongoDBFormatter, ElasticaFormatter, ElasticsearchFormatter, LineFormatter, JsonFormatter, NormalizerFormatter};
use Psr\Log\LogLevel;
use PHPUnit\Framework\TestCase;

trait MonologCovConnSock
{
    public array $sent = [];
    protected function fsockopen() { return fopen('php://memory', 'w+'); }
    protected function pfsockopen() { return fopen('php://memory', 'w+'); }
    protected function streamSetTimeout(): bool { return true; }
    protected function streamSetChunkSize(): int|bool { return 1; }
    protected function fwrite(string $data): int|bool { $this->sent[] = $data; return strlen($data); }
    public function isConnected(): bool { return is_resource($this->getResource()); }
}

final class MonologCovConnSocket extends SocketHandler { use MonologCovConnSock; }
final class MonologCovConnPushover extends PushoverHandler { use MonologCovConnSock; }
final class MonologCovConnInsight extends InsightOpsHandler { use MonologCovConnSock; }
final class MonologCovConnEntries extends LogEntriesHandler { use MonologCovConnSock; }

final class MonologCovJsBrowser extends BrowserConsoleHandler
{
    public static string $out = '';
    public static string $fmt = 'html';
    protected static function writeOutput(string $str): void { self::$out .= $str; }
    protected static function getResponseFormat(): string { return self::$fmt; }
}

final class MonologCov6Test extends TestCase
{
    private int $oldReporting = 0;

    protected function setUp(): void { $this->oldReporting = error_reporting(E_ALL); }

    protected function tearDown(): void { error_reporting($this->oldReporting); }

    private function rec(string $m = 'm', Level $l = Level::Info, array $ctx = [], array $extra = [], string $ch = 'chan'): LogRecord
    {
        return new LogRecord(new \DateTimeImmutable('2024-01-02 03:04:05.250000', new \DateTimeZone('UTC')), $ch, $l, $m, $ctx, $extra);
    }

    public function testGelfFormatterAndHandler(): void
    {
        $f = new GelfMessageFormatter('sys', 'ex_', 'cx_', 200);
        $msg = $f->format($this->rec('short', Level::Error, ['a' => 1, 'b' => true, 'c' => [1, 2], 'k y!' => 'v', 'n' => null, 'exception' => ['file' => '/x/y.php:42']], ['e' => 'x', 'f' => false, 'g' => ['q' => 1]]));
        $this->assertSame('short', $msg->data['short']);
        $this->assertSame('sys', $msg->data['host']);
        $this->assertSame(3, $msg->data['level']);
        $this->assertSame('chan', $msg->additional['facility']);
        $this->assertSame(1, $msg->additional['cx_a']);
        $this->assertSame(1, $msg->additional['cx_b']);
        $this->assertSame('[1,2]', $msg->additional['cx_c']);
        $this->assertSame('v', $msg->additional['cx_k-y-']);
        $this->assertSame('x', $msg->additional['ex_e']);
        $this->assertSame(0, $msg->additional['ex_f']);
        $this->assertSame('/x/y.php', $msg->additional['file']);
        $this->assertSame('42', $msg->additional['line']);
        $this->assertArrayHasKey('cx_n', $msg->additional);
        $long = $f->format($this->rec(str_repeat('m', 300), Level::Debug, ['big' => str_repeat('b', 500)], ['bigx' => str_repeat('x', 500)], ''));
        $this->assertSame(200, strlen($long->data['short']));
        $this->assertArrayNotHasKey('facility', $long->additional);
        $this->assertSame(200, strlen($long->additional['cx_big']));
        $this->assertSame(200, strlen($long->additional['ex_bigx']));
        $d = new GelfMessageFormatter();
        foreach ([[Level::Debug, 7], [Level::Info, 6], [Level::Notice, 5], [Level::Warning, 4], [Level::Critical, 2], [Level::Alert, 1], [Level::Emergency, 0]] as [$l, $p]) {
            $this->assertSame($p, $d->format($this->rec('m', $l))->data['level']);
        }
        $this->assertSame(gethostname(), $d->format($this->rec())->data['host']);

        $pub = new class implements \Gelf\PublisherInterface {
            public array $got = [];
            public function publish(\Gelf\Message $m) { $this->got[] = $m; }
        };
        $h = new GelfHandler($pub);
        $h->handle($this->rec('to gelf'));
        $this->assertSame('to gelf', $pub->got[0]->data['short']);
        $this->assertInstanceOf(GelfMessageFormatter::class, $h->getFormatter());
    }

    public function testMongoFormatterAndHandler(): void
    {
        $f = new MongoDBFormatter(4, true);
        $e = new \RuntimeException('boom', 5);
        $out = $f->format($this->rec('m', Level::Info, ['ex' => $e, 'obj' => new \ArrayObject([1]), 'nested' => ['a' => ['b' => ['c' => 1]]], 'dt' => new \DateTimeImmutable('2020-01-01 00:00:00.500', new \DateTimeZone('UTC'))]));
        $this->assertSame('RuntimeException', $out['context']['ex']['class']);
        $this->assertIsString($out['context']['ex']['trace']);
        $this->assertSame('[...]', (new MongoDBFormatter(2))->format($this->rec('m', Level::Info, ['nested' => ['a' => ['b' => ['c' => 1]]]]))['context']['nested']['a']);
        $this->assertInstanceOf(\MongoDB\BSON\UTCDateTime::class, $out['context']['dt']);
        $this->assertSame(1577836800500, $out['context']['dt']->ms);
        $this->assertInstanceOf(\MongoDB\BSON\UTCDateTime::class, $out['datetime']);
        $this->assertSame('ArrayObject', $out['context']['obj']['class']);
        $f2 = new MongoDBFormatter(0, false);
        $out = $f2->format($this->rec('m', Level::Info, ['ex' => $e]));
        $this->assertIsArray($out['context']['ex']['trace']);
        $this->assertCount(2, $f2->formatBatch([$this->rec('a'), $this->rec('b')]));

        $client = new \MongoDB\Client();
        $h = new MongoDBHandler($client, 'db', 'coll');
        $h->handle($this->rec('to mongo'));
        $this->assertSame('to mongo', $client->last->docs[0]['message']);
        $mgr = new \MongoDB\Driver\Manager();
        $h2 = new MongoDBHandler($mgr, 'db', 'coll');
        $h2->handle($this->rec('via manager'));
        $this->assertSame('db.coll', $mgr->writes[0][0]);
        $this->assertSame('via manager', $mgr->writes[0][1][0]['message']);
        $this->assertInstanceOf(MongoDBFormatter::class, $h2->getFormatter());
    }

    public function testCouchDoctrineHandler(): void
    {
        $c = new \Doctrine\CouchDB\CouchDBClient();
        $h = new DoctrineCouchDBHandler($c);
        $h->handle($this->rec('couch', Level::Info, ['a' => 1]));
        $this->assertSame('couch', $c->docs[0]['message']);
        $this->assertInstanceOf(NormalizerFormatter::class, $h->getFormatter());
    }

    public function testElastica(): void
    {
        $c = new \Elastica\Client();
        $h = new ElasticaHandler($c, ['index' => 'idx', 'type' => 'typ', 'ignore_error' => false]);
        $this->assertSame('idx', $h->getOptions()['index']);
        $h->handle($this->rec('es', Level::Info, ['a' => 1]));
        $this->assertSame('idx', $c->added[0][0]->index);
        $this->assertSame('es', $c->added[0][0]->data['message']);
        $h->handleBatch([$this->rec('a'), $this->rec('b')]);
        $this->assertCount(2, $c->added[1]);
        $this->assertSame($h, $h->setFormatter(new ElasticaFormatter('i2', 't2')));
        $f = new ElasticaFormatter('i', null);
        $this->assertSame('i', $f->getIndex());
        try { $h->setFormatter(new LineFormatter()); $this->fail(); } catch (\InvalidArgumentException $e) { $this->assertStringContainsString('ElasticaFormatter', $e->getMessage()); }
        $c->fail = new \RuntimeException('transport');
        $c->fail = null;
        $default = new ElasticaHandler($c);
        $this->assertSame('monolog', $default->getOptions()['index']);
        $this->assertSame('record', $default->getOptions()['type']);
    }

    public function testElasticsearch(): void
    {
        $c = new \Elasticsearch\Client();
        $h = new ElasticsearchHandler($c, ['index' => 'logs', 'op_type' => 'create']);
        $this->assertSame('_doc', $h->getOptions()['type']);
        $h->handle($this->rec('es', Level::Info));
        $this->assertSame(['create' => ['_index' => 'logs']], $c->bulks[0]['body'][0]);
        $this->assertSame('es', $c->bulks[0]['body'][1]['message']);
        $this->assertArrayNotHasKey('_index', $c->bulks[0]['body'][1]);
        $h->handleBatch([$this->rec('a'), $this->rec('b')]);
        $this->assertCount(4, $c->bulks[1]['body']);
        $this->assertSame($h, $h->setFormatter(new ElasticsearchFormatter('i', 't')));
        try { $h->setFormatter(new LineFormatter()); $this->fail(); } catch (\InvalidArgumentException $e) { $this->assertStringContainsString('ElasticsearchFormatter', $e->getMessage()); }

        $c->response = ['errors' => true, 'items' => [['index' => ['error' => ['type' => 'mapper', 'reason' => 'bad', 'caused_by' => ['type' => 'inner', 'reason' => 'deep']]]]]];
        try { $h->handle($this->rec('fail')); $this->fail(); } catch (\RuntimeException $e) {
            $this->assertSame('Error sending messages to Elasticsearch', $e->getMessage());
            $this->assertSame('mapper: bad', $e->getPrevious()->getMessage());
            $this->assertSame('inner: deep', $e->getPrevious()->getPrevious()->getMessage());
        }
        $c->response = ['errors' => true, 'items' => []];
        try { $h->handle($this->rec('fail2')); $this->fail(); } catch (\RuntimeException $e) {
            $this->assertStringContainsString('failed to index', $e->getPrevious()->getMessage());
        }
        $quiet = new ElasticsearchHandler($c, ['ignore_error' => true]);
        $quiet->handle($this->rec('ignored'));
        $this->assertTrue(true);

        $c8 = new \Elastic\Elasticsearch\Client();
        $h8 = new ElasticsearchHandler($c8, ['type' => 'ignored']);
        $this->assertSame('_doc', $h8->getOptions()['type']);
        $h8->handle($this->rec('v8'));
        $this->assertSame(['index' => ['_index' => 'monolog']], $c8->bulks[0]['body'][0]);
        $f = new ElasticsearchFormatter('i', 't');
        $this->assertSame('i', $f->getIndex());
    }

    public function testConnectedSocketHandlers(): void
    {
        $h = new MonologCovConnSocket('tcp://example.invalid:80', Level::Debug, true, false, 1.5, 3.0, 2.0, 1024);
        $h->setFormatter(new LineFormatter("%message%\n"));
        $h->handle($this->rec('hello'));
        $h->handle($this->rec('again'));
        $this->assertSame(["hello\n", "again\n"], $h->sent);
        $this->assertTrue($h->isConnected());
        $h->close();
        $this->assertFalse($h->isConnected());
        $h->setPersistent(true);
        $h->handle($this->rec('persist'));
        $h->close();
        $this->assertTrue($h->isConnected());
        $h->closeSocket();

        $p = new MonologCovConnPushover('tok', ['u1', 'u2'], 'T', Level::Debug, true, false, Level::Critical, Level::Emergency, 10, 600);
        $p->setFormatter(new LineFormatter("%message%!"));
        $p->handle($this->rec('info msg', Level::Info, ['sound' => 'bike']));
        $this->assertCount(2, $p->sent);
        parse_str(substr($p->sent[0], strpos($p->sent[0], "\r\n\r\n") + 4), $q);
        $this->assertSame('bike', $q['sound']);
        $this->assertSame('u1', $q['user']);
        $p->sent = [];
        $p->handle($this->rec('crit', Level::Critical, ['sound' => 'nope']));
        parse_str(substr($p->sent[0], strpos($p->sent[0], "\r\n\r\n") + 4), $q);
        $this->assertSame('1', $q['priority']);
        $this->assertArrayNotHasKey('sound', $q);
        $p->sent = [];
        $p->useFormattedMessage(true)->handle($this->rec('emerg', Level::Emergency));
        parse_str(substr($p->sent[0], strpos($p->sent[0], "\r\n\r\n") + 4), $q);
        $this->assertSame(['2', '10', '600', 'emerg!'], [$q['priority'], $q['retry'], $q['expire'], $q['message']]);
        $p->setHighPriorityLevel('error')->setEmergencyLevel(Level::Alert);
        $p->sent = [];
        $p->handle($this->rec('alert', Level::Alert));
        parse_str(substr($p->sent[0], strpos($p->sent[0], "\r\n\r\n") + 4), $q);
        $this->assertSame('2', $q['priority']);

        $i = new MonologCovConnInsight('tok', 'eu', false);
        $i->setFormatter(new LineFormatter("%message%\n"));
        $i->handle($this->rec('ins'));
        $this->assertSame("tok ins\n", $i->sent[0]);
        $this->assertSame('eu.data.logs.insight.rapid7.com:80', $i->getConnectionString());
        $l = new MonologCovConnEntries('letok', false);
        $l->setFormatter(new LineFormatter("%message%\n"));
        $l->handle($this->rec('le'));
        $this->assertSame("letok le\n", $l->sent[0]);
    }

    public function testBrowserConsoleWithFormatOverride(): void
    {
        MonologCovJsBrowser::resetStatic();
        MonologCovJsBrowser::$out = '';
        MonologCovJsBrowser::$fmt = 'html';
        $h = new MonologCovJsBrowser();
        $h->handle($this->rec('plain info', Level::Info));
        $h->handle($this->rec('with ctx', Level::Warning, ['a' => 1, 'n' => null], ['e' => 'x"y</script>']));
        $h->handle($this->rec('dbg', Level::Debug));
        $h->handle($this->rec('err', Level::Critical));
        $h->handle($this->rec('notice', Level::Notice));
        MonologCovJsBrowser::send();
        $out = MonologCovJsBrowser::$out;
        $this->assertStringStartsWith('<script>(function (c) {if (c && c.groupCollapsed) {', $out);
        foreach (['c.info(', 'c.groupCollapsed(', 'c.groupEnd();', 'c.debug(', 'c.error('] as $needle) {
            $this->assertStringContainsString($needle, $out);
        }
        $this->assertStringContainsString('background-color: ', $out);
        $this->assertStringContainsString('</script>', $out);
        MonologCovJsBrowser::$out = '';
        MonologCovJsBrowser::send();
        $this->assertSame('', MonologCovJsBrowser::$out);
        MonologCovJsBrowser::$fmt = 'js';
        $h->handle($this->rec('js out', Level::Warning));
        MonologCovJsBrowser::send();
        $this->assertStringStartsWith('(function (c)', MonologCovJsBrowser::$out);
        MonologCovJsBrowser::$out = '';
        MonologCovJsBrowser::$fmt = 'unknown';
        $h->handle($this->rec('skipped'));
        MonologCovJsBrowser::send();
        $this->assertSame('', MonologCovJsBrowser::$out);
        $h->close();
        $h->reset();
    }

    public function testNativeMailerHtmlBody(): void
    {
        $n = new class('a@x.invalid', 'S %level_name%', 'f@x.invalid') extends NativeMailerHandler {
            public array $mails = [];
            protected function mail(string $to, string $subject, string $content, string $headers, string $parameters): void { $this->mails[] = [$to, $subject, $headers]; }
        };
        $n->setContentType('text/html');
        $n->handle($this->rec('html', Level::Error));
        $this->assertSame('a@x.invalid', $n->mails[0][0]);
        $this->assertSame('S ERROR', $n->mails[0][1]);
        $this->assertStringContainsString('MIME-Version: 1.0', $n->mails[0][2]);
        $n->handleBatch([$this->rec('b', Level::Error)]);
        $this->assertCount(2, $n->mails);
    }

    public function testErrorHandlerInternals(): void
    {
        $h = new TestHandler();
        $eh = new ErrorHandler(new Logger('int', [$h]));
        $eh->captureStackTraces(true, E_USER_WARNING);
        $eh->registerErrorHandler([], false);
        restore_error_handler();
        $call = new \ReflectionMethod($eh, 'handleError');
        $old = error_reporting(E_ALL);
        try {
            $this->assertTrue($call->invoke($eh, E_USER_WARNING, 'warned', 'f.php', 3));
            $this->assertTrue($call->invoke($eh, E_USER_NOTICE, 'noticed', 'f.php', 4));
            error_reporting(0);
            $this->assertFalse($call->invoke($eh, E_USER_WARNING, 'silenced', 'f.php', 5));
        } finally {
            error_reporting($old);
        }
        $recs = $h->getRecords();
        $this->assertCount(2, $recs);
        $this->assertSame('E_USER_WARNING: warned', $recs[0]->message);
        $this->assertSame(Level::Warning, $recs[0]->level);
        $this->assertInstanceOf(\ErrorException::class, $recs[0]->context['exception']);
        $this->assertSame('E_USER_NOTICE: noticed', $recs[1]->message);
        $this->assertSame(['code' => E_USER_NOTICE, 'message' => 'noticed', 'file' => 'f.php', 'line' => 4], $recs[1]->context);
        foreach ([E_ERROR => 'E_ERROR', E_NOTICE => 'E_NOTICE', E_PARSE => 'E_PARSE', E_DEPRECATED => 'E_DEPRECATED', E_RECOVERABLE_ERROR => 'E_RECOVERABLE_ERROR', 12345 => 'Unknown PHP error'] as $code => $name) {
            $h->clear();
            $call->invoke($eh, $code, 'x', 'f.php', 1);
            $this->assertSame($name . ': x', $h->getRecords()[0]->message);
        }
    }

    public function testErrorHandlerFatalPath(): void
    {
        $h = new TestHandler();
        $eh = new ErrorHandler(new Logger('fatal', [$h]));
        $this->assertSame($eh, $eh->registerFatalHandler(LogLevel::EMERGENCY, 1));
        $err = new \ReflectionMethod($eh, 'handleError');
        $this->assertTrue($err->invoke($eh, E_USER_ERROR, 'died', 'f.php', 9));
        $this->assertCount(0, $h->getRecords());
        $eh->handleFatalError();
        $this->assertCount(1, $h->getRecords());
        $r = $h->getRecords()[0];
        $this->assertSame(Level::Emergency, $r->level);
        $this->assertSame('Fatal Error (E_USER_ERROR): died', $r->message);
        $this->assertSame('f.php', $r->context['file']);
        $this->assertInstanceOf(\ErrorException::class, $r->context['exception']);
    }

    public function testErrorHandlerInternalsUnfiltered(): void
    {
        $h = new TestHandler();
        $eh = new ErrorHandler(new Logger('int', [$h]));
        $eh->registerErrorHandler([], false, -1, false);
        restore_error_handler();
        $call = new \ReflectionMethod($eh, 'handleError');
        $old = error_reporting(E_ALL);
        try {
            $this->assertTrue($call->invoke($eh, E_USER_WARNING, 'warned', 'f.php', 3));
            $this->assertTrue($call->invoke($eh, E_USER_NOTICE, 'noticed', 'f.php', 4));
        } finally {
            error_reporting($old);
        }
        $recs = $h->getRecords();
        $this->assertCount(2, $recs);
        $this->assertSame('E_USER_WARNING: warned', $recs[0]->message);
        $this->assertSame(Level::Warning, $recs[0]->level);
        $this->assertArrayNotHasKey('exception', $recs[0]->context);
        $this->assertSame('E_USER_NOTICE: noticed', $recs[1]->message);
        $this->assertSame(['code' => E_USER_NOTICE, 'message' => 'noticed', 'file' => 'f.php', 'line' => 4], $recs[1]->context);
        foreach ([E_ERROR => 'E_ERROR', E_NOTICE => 'E_NOTICE', E_PARSE => 'E_PARSE', E_DEPRECATED => 'E_DEPRECATED', E_RECOVERABLE_ERROR => 'E_RECOVERABLE_ERROR', 12345 => 'Unknown PHP error'] as $code => $name) {
            $h->clear();
            $call->invoke($eh, $code, 'x', 'f.php', 1);
            $this->assertSame($name . ': x', $h->getRecords()[0]->message);
        }
    }

    public function testErrorHandlerFatalPathUnfiltered(): void
    {
        $h = new TestHandler();
        $eh = new ErrorHandler(new Logger('fatal2', [$h]));
        $eh->registerFatalHandler(LogLevel::EMERGENCY, 1);
        $eh->registerErrorHandler([], false, -1, false);
        restore_error_handler();
        $err = new \ReflectionMethod($eh, 'handleError');
        $this->assertTrue($err->invoke($eh, E_USER_ERROR, 'died', 'f.php', 9));
        $this->assertCount(0, $h->getRecords());
        $prop = new \ReflectionProperty($eh, 'lastFatalData');
        $this->assertSame('died', $prop->getValue($eh)['message']);
    }

    public function testErrorHandlerRegistrationReturns(): void
    {
        $eh = new ErrorHandler(new Logger('reg', [new TestHandler()]));
        $this->assertSame($eh, $eh->registerExceptionHandler(['LogicException' => LogLevel::WARNING], false));
        restore_exception_handler();
        $this->assertSame($eh, $eh->registerErrorHandler([E_USER_NOTICE => LogLevel::ALERT], false, E_USER_NOTICE, false));
        restore_error_handler();
        $this->assertSame($eh, $eh->captureStackTraces(false));
    }
}
