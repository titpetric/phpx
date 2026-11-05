<?php
use Monolog\{Logger, Level, LogRecord};
use Monolog\Formatter\{LineFormatter, JsonFormatter, NormalizerFormatter, ScalarFormatter, HtmlFormatter, RedactingFormatter, WildfireFormatter, FluentdFormatter, FlowdockFormatter, ChromePHPFormatter, SyslogFormatter, LogglyFormatter, LogmaticFormatter, LogstashFormatter, GoogleCloudLoggingFormatter, ElasticsearchFormatter};
use PHPUnit\Framework\TestCase;

final class MonologCovSecret
{
    public function __construct(
        #[\SensitiveParameter]
        private string $token, public string $name = 'n') {}
}

final class MonologCovStr implements \Stringable
{
    public function __construct(private string $s) {}
    public function __toString(): string { return $this->s; }
}

final class MonologCovThrowStr implements \Stringable
{
    public function __toString(): string { throw new \RuntimeException('no'); }
}

final class MonologCovJs implements \JsonSerializable
{
    public function jsonSerialize(): mixed { return ['js' => 1]; }
}

final class MonologCovTest extends TestCase
{
    private function rec(string $m = 'm', Level $l = Level::Info, array $ctx = [], array $extra = [], string $ch = 'chan'): LogRecord
    {
        return new LogRecord(new \DateTimeImmutable('2024-01-02 03:04:05.123456', new \DateTimeZone('UTC')), $ch, $l, $m, $ctx, $extra);
    }

    public function testRedactingKeys(): void
    {
        $f = new RedactingFormatter(new JsonFormatter(), ['Password', 'secret']);
        $out = $f->format($this->rec('x', Level::Info, ['password' => 'hunter22', 'ok' => 1, 'n' => ['SECRET' => 'abc', 'deep' => ['x' => 2]]], ['secret' => 'zzzzzz']));
        $d = json_decode($out, true);
        $this->assertSame('[REDACTED]', $d['context']['password']);
        $this->assertSame('[REDACTED]', $d['context']['n']['SECRET']);
        $this->assertSame(2, $d['context']['n']['deep']['x']);
        $this->assertSame('[REDACTED]', $d['extra']['secret']);
        $this->assertSame(1, $d['context']['ok']);
        $this->assertInstanceOf(JsonFormatter::class, $f->getWrappedFormatter());
    }

    public function testRedactingSweepsValues(): void
    {
        $f = new RedactingFormatter(new LineFormatter("%message% %context%\n"), ['token']);
        $out = $f->format($this->rec('leak hunter22 here', Level::Info, ['token' => 'hunter22']));
        $this->assertStringNotContainsString('hunter22', $out);
        $this->assertStringContainsString('[REDACTED]', $out);
        // nested array under a sensitive key
        $out = $f->format($this->rec('a longsecret b other', Level::Info, ['token' => ['x' => 'longsecret', 'y' => ['z' => 'otherval']]]));
        $this->assertStringNotContainsString('longsecret', $out);
        // short secrets stay
        $out = $f->format($this->rec('ab cd', Level::Info, ['token' => 'ab']));
        $this->assertStringContainsString('ab cd', $out);
        // escaped forms
        $out = $f->format($this->rec('say "q\\uote" <b>&', Level::Info, ['token' => 'say "q\\uote" <b>&']));
        $this->assertStringNotContainsString('uote', $out);
        // ints left alone
        $out = $f->format($this->rec('12345678', Level::Info, ['token' => 12345678]));
        $this->assertStringContainsString('12345678', $out);
    }

    public function testRedactingSensitiveParameterObject(): void
    {
        $f = new RedactingFormatter(new LineFormatter("%message% %context%\n"), ['token']);
        $obj = new MonologCovSecret('supersecretvalue');
        $out = $f->format($this->rec('msg supersecretvalue', Level::Info, ['o' => $obj, 'again' => $obj]));
        $this->assertStringNotContainsString('supersecretvalue', $out);
    }

    public function testRedactingPlainKeysAndStringable(): void
    {
        $f = new RedactingFormatter(new LineFormatter("%message% %context%\n"), ['token']);
        $out = $f->format($this->rec('msg strsecret1', Level::Info, ['tok' => 1], []));
        $this->assertStringContainsString('strsecret1', $out);
        $out = $f->format($this->rec('msg strsecret1', Level::Info, ['token' => new MonologCovStr('strsecret1'), 'bad' => new MonologCovThrowStr()]));
        $this->assertStringNotContainsString('strsecret1', $out);
    }

    public function testRedactingSensitiveParameterValue(): void
    {
        $f = new RedactingFormatter(new LineFormatter("%message% %context%\n"), ['token']);
        $out = $f->format($this->rec('msg spvalue1', Level::Info, ['x' => new \SensitiveParameterValue('spvalue1')]));
        $this->assertStringNotContainsString('spvalue1', $out);
    }

    public function testRedactingWithoutParameterScan(): void
    {
        $obj = new MonologCovSecret('supersecretvalue');
        $f2 = new RedactingFormatter(new LineFormatter("%message% %context%\n"), ['token'], [], '***', false);
        $out = $f2->format($this->rec('msg supersecretvalue', Level::Info, ['o' => $obj]));
        $this->assertStringContainsString('supersecretvalue', $out);
        $out = $f2->format($this->rec('p', Level::Info, ['password' => 'x', 'token' => 'abcdefgh']));
        $this->assertStringContainsString('***', $out);
    }

    public function testRedactingPatternsSingle(): void
    {
        $f = new RedactingFormatter(new LineFormatter("%message%\n"), [], ['/\d{3}-\d{2}/']);
        $this->assertSame("a [REDACTED]\n", $f->format($this->rec('a 111-22')));
    }

    public function testRedactingPatternsArray(): void
    {
        $f = new RedactingFormatter(new LineFormatter("%message%\n"), [], [RedactingFormatter::TOKEN_PATTERN, '/\d{3}-\d{2}/']);
        $out = $f->format($this->rec('tok abcdefghijklmnopqrstuvwxyz0123456789 and 123-45 end'));
        $this->assertSame("tok [REDACTED] and [REDACTED] end\n", $out);
    }

    public function testRedactingBatch(): void
    {
        $f = new RedactingFormatter(new LineFormatter("%message%\n"), ['pw']);
        $out = $f->formatBatch([$this->rec('a abcdefgh', Level::Info, ['pw' => 'abcdefgh']), $this->rec('b')]);
        $this->assertSame("a [REDACTED]\nb\n", $out);
    }

    public function testRedactingInvalidPattern(): void
    {
        $this->expectException(\InvalidArgumentException::class);
        new RedactingFormatter(new LineFormatter(), [], ['/(unclosed']);
    }

    public function testRedactingArrayOutputAndDepth(): void
    {
        $f = new RedactingFormatter(new NormalizerFormatter(), ['pw']);
        $out = $f->format($this->rec('with abcdefgh', Level::Info, ['pw' => 'abcdefgh']));
        $this->assertSame('with [REDACTED]', $out['message']);
        $deep = ['pw' => 'secret99'];
        for ($i = 0; $i < 12; $i++) { $deep = ['k' => $deep]; }
        $out = $f->formatBatch([$this->rec('x', Level::Info, $deep)]);
        $this->assertIsArray($out);
        $f3 = new RedactingFormatter(new JsonFormatter(), ['pw'], ['/x/']);
        $this->assertIsString($f3->format($this->rec('x', Level::Info, $deep)));
    }

    public function testHtmlFormatter(): void
    {
        $f = new HtmlFormatter('Y-m-d');
        foreach (Level::cases() as $lv) {
            $o = $f->format($this->rec('a<b', $lv));
            $this->assertStringContainsString('a&lt;b', $o);
            $this->assertStringContainsString($lv->getName(), $o);
        }
        $o = $f->format($this->rec('m', Level::Error, ['a' => 1, 'b' => ['x' => 'y'], 'n' => null, 's' => 'str'], ['e' => true]));
        $this->assertStringContainsString('Context', $o);
        $this->assertStringContainsString('Extra', $o);
        $this->assertStringContainsString('2024-01-02', $o);
        $this->assertStringContainsString('#FD7E14', $o);
        $b = $f->formatBatch([$this->rec('1'), $this->rec('2')]);
        $this->assertSame(2, substr_count($b, '<h1'));
    }

    public function testHtmlFormatterNestedContext(): void
    {
        $o = (new HtmlFormatter())->format($this->rec('m', Level::Error, ['b' => ['x' => 'y']]));
        $this->assertStringContainsString('"x": "y"', $o);
    }

    public function testWildfire(): void
    {
        $f = new WildfireFormatter();
        $o = $f->format($this->rec('hello', Level::Debug));
        $this->assertSame('59|[{"Type":"LOG","File":"","Line":"","Label":"chan"},"hello"]|', $o);
        $o = $f->format($this->rec('hello', Level::Warning, ['a' => 1], ['file' => 'f.php', 'line' => 12, 'k' => 'v']));
        $this->assertStringContainsString('"Type":"WARN","File":"f.php","Line":12', $o);
        $this->assertStringContainsString('"extra":{"k":"v"}', $o);
        foreach ([[Level::Info, 'INFO'], [Level::Notice, 'INFO'], [Level::Error, 'ERROR'], [Level::Critical, 'ERROR'], [Level::Alert, 'ERROR'], [Level::Emergency, 'ERROR']] as [$lv, $t]) {
            $this->assertStringContainsString('"Type":"' . $t . '"', $f->format($this->rec('x', $lv)));
        }
        $o = $f->format($this->rec('tbl', Level::Info, ['table' => [['a', 'b'], [1, 2]]]));
        $this->assertStringContainsString('"Type":"TABLE"', $o);
        $this->assertStringContainsString('"Label":"chan: tbl"', $o);
        $o = $f->format($this->rec('obj', Level::Info, ['d' => new \DateTimeImmutable('2020-01-01 00:00:00', new \DateTimeZone('UTC'))]));
        $this->assertStringContainsString('2020-01-01', $o);
        $this->expectException(\BadMethodCallException::class);
        $f->formatBatch([$this->rec()]);
    }

    public function testFluentd(): void
    {
        $f = new FluentdFormatter();
        $this->assertFalse($f->isUsingLevelsInTag());
        $o = $f->format($this->rec('m', Level::Info, ['a' => 1]));
        $d = json_decode($o, true);
        $this->assertSame('chan', $d[0]);
        $this->assertSame(1704164645, $d[1]);
        $this->assertSame('m', $d[2]['message']);
        $f = new FluentdFormatter(true);
        $this->assertTrue($f->isUsingLevelsInTag());
        $d = json_decode($f->format($this->rec('m', Level::Error)), true);
        $this->assertSame('chan.error', $d[0]);
        $b = $f->formatBatch([$this->rec('a'), $this->rec('b')]);
        $this->assertSame(2, substr_count($b, 'chan.info'));
    }

    public function testFlowdock(): void
    {
        $f = new FlowdockFormatter('src', 'a@b.c');
        $o = $f->format($this->rec('short', Level::Error, [], ['x' => 'y']));
        $this->assertSame('in src: ERROR - short', $o['subject']);
        $this->assertSame(['#logs', '#error', '#chan', '#y'], $o['tags']);
        $this->assertSame('a@b.c', $o['from_address']);
        $this->assertCount(2, $f->formatBatch([$this->rec('a'), $this->rec('b')]));
        $this->assertSame(str_repeat('a', 41) . ' ...', $f->getShortMessage(str_repeat('a', 60)));
        $this->assertSame('abc', $f->getShortMessage('abc'));
    }

    public function testChromePhp(): void
    {
        $f = new ChromePHPFormatter();
        $o = $f->format($this->rec('m', Level::Debug, ['a' => 1], ['file' => 'x.php', 'line' => 3]));
        $this->assertSame('log', $o[3]);
        $this->assertSame('x.php : 3', $o[2]);
        $this->assertSame('chan', $o[0]);
        $o = $f->format($this->rec('m', Level::Error));
        $this->assertSame('error', $o[3]);
        $this->assertSame('unknown', $o[2]);
        foreach ([[Level::Info, 'info'], [Level::Notice, 'info'], [Level::Warning, 'warn'], [Level::Critical, 'error'], [Level::Alert, 'error'], [Level::Emergency, 'error']] as [$lv, $t]) {
            $this->assertSame($t, $f->format($this->rec('x', $lv))[3]);
        }
        $this->assertCount(2, $f->formatBatch([$this->rec('a'), $this->rec('b', Level::Warning, [], ['file' => 'a', 'line' => 1])]));
    }

    public function testSyslogFormatter(): void
    {
        $f = new SyslogFormatter('app');
        $o = $f->format($this->rec('hello', Level::Error, ['a' => 1], ['b' => 2]));
        $this->assertStringContainsString(' app ', $o);
        $this->assertStringContainsString('hello', $o);
        $this->assertStringContainsString('2024-01-02T03:04:05.123456+00:00', $o);
    }

    public function testSyslogFormatterPriority(): void
    {
        $o = (new SyslogFormatter('app'))->format($this->rec('hello', Level::Error, ['a' => 1], ['b' => 2]));
        $this->assertStringStartsWith('<11>1 2024-01-02T03:04:05.123456+00:00 ', $o);
        $o = (new SyslogFormatter())->format($this->rec('hi'));
        $this->assertStringStartsWith('<14>1 ', $o);
    }

    public function testLoggly(): void
    {
        $f = new LogglyFormatter();
        $d = json_decode($f->format($this->rec('m', Level::Info, ['a' => 1])), true);
        $this->assertArrayHasKey('timestamp', $d);
        $this->assertArrayNotHasKey('datetime', $d);
        $f = new LogglyFormatter(LogglyFormatter::BATCH_MODE_JSON, true);
        $this->assertStringEndsWith("\n", $f->format($this->rec()));
    }

    public function testLogmatic(): void
    {
        $f = new LogmaticFormatter();
        $this->assertSame($f, $f->setHostname('h'));
        $this->assertSame($f, $f->setAppName('app'));
        $d = json_decode($f->format($this->rec('m')), true);
        $this->assertSame('h', $d['hostname']);
        $this->assertSame('app', $d['appname']);
        $this->assertSame(['sourcecode', 'php'], $d['@marker']);
    }

    public function testLogstash(): void
    {
        $f = new LogstashFormatter('app', 'sys', 'ex', 'cx');
        $d = json_decode($f->format($this->rec('m', Level::Info, ['a' => 1], ['b' => 2])), true);
        $this->assertSame('sys', $d['host']);
        $this->assertSame('app', $d['type']);
        $this->assertSame(['a' => 1], $d['cx']);
        $this->assertSame(['b' => 2], $d['ex']);
        $this->assertSame('chan', $d['channel']);
        $d = json_decode((new LogstashFormatter('app'))->format($this->rec('m')), true);
        $this->assertArrayHasKey('@timestamp', $d);
        $this->assertArrayNotHasKey('context', $d);
    }

    public function testGoogleCloud(): void
    {
        $f = new GoogleCloudLoggingFormatter();
        $d = json_decode($f->format($this->rec('m', Level::Warning, ['a' => 1])), true);
        $this->assertSame('WARNING', $d['severity']);
        $this->assertSame('m', $d['message']);
        $this->assertArrayNotHasKey('level', $d);
        $this->assertArrayNotHasKey('level_name', $d);
        $this->assertArrayHasKey('time', $d);
    }

    public function testElasticsearch(): void
    {
        $f = new ElasticsearchFormatter('idx', 'typ');
        $this->assertSame('idx', $f->getIndex());
        $this->assertSame('typ', $f->getType());
        $o = $f->format($this->rec('m', Level::Info, ['a' => 1]));
        $this->assertSame('idx', $o['_index']);
        $this->assertSame('typ', $o['_type']);
        $this->assertSame('m', $o['message']);
    }

    public function testScalarAndLineMore(): void
    {
        $f = new ScalarFormatter();
        $o = $f->format($this->rec('m', Level::Info, ['a' => [1, 2], 'o' => new MonologCovJs()]));
        $this->assertSame('{"a":[1,2],"o":{"MonologCovJs":{"js":1}}}', $o['context']);
        $this->assertSame('m', $o['message']);
        $l = new LineFormatter(null, null, false, true);
        $ex = new \RuntimeException('boom', 3, new \LogicException('inner'));
        $o = $l->format($this->rec('m', Level::Error, ['exception' => $ex]));
        $this->assertStringContainsString('[object] (RuntimeException(code: 3): boom at', $o);
        $this->assertStringContainsString('[previous exception]', $o);
        $l->includeStacktraces(true);
        $this->assertStringContainsString('[stacktrace]', $l->format($this->rec('m', Level::Error, ['exception' => $ex])));
        $l->includeStacktraces(true, fn(string $s) => 'ST:' . strlen($s) > 0 ? 'ST' : '');
        $this->assertStringContainsString('ST', $l->format($this->rec('m', Level::Error, ['exception' => $ex])));
        $l->setMaxLevelNameLength(3);
        $this->assertStringContainsString('ERR', $l->format($this->rec('m', Level::Error)));
        $l->setMaxLevelNameLength(null);
        $l->allowInlineLineBreaks(false);
        $l->ignoreEmptyContextAndExtra(false);
        $l->setJsonPrettyPrint(false);
        $this->assertStringContainsString('[] []', $l->format($this->rec('m')));
        $l2 = new LineFormatter("%message% %extra.a% %context.b% %datetime%\n", 'Y');
        $this->assertSame("m 1 2 2024\n", $l2->format($this->rec('m', Level::Info, ['b' => 2], ['a' => 1])));
        $this->assertSame("m  2 2024\n", $l2->format($this->rec('m', Level::Info, ['b' => 2])));
        $this->assertSame("m 1 2 2024\n", $l2->stringify("") === '' ? "m 1 2 2024\n" : '');
    }

    public function testNormalizerMore(): void
    {
        $n = new NormalizerFormatter();
        $n->setJsonPrettyPrint(true);
        $n->setMaxNormalizeDepth(2);
        $n->setMaxNormalizeItemCount(2);
        $this->assertSame(2, $n->getMaxNormalizeDepth());
        $this->assertSame(2, $n->getMaxNormalizeItemCount());
        $o = $n->format($this->rec('m', Level::Info, ['a' => ['b' => ['c' => ['d' => 1]]], 'many' => [1, 2, 3, 4]]));
        $this->assertSame('Over 2 items (4 total), aborting normalization', $o['context']['many']['...']);
        $n2 = new NormalizerFormatter();
        $o = $n2->format($this->rec('m', Level::Info, [
            'inf' => INF, 'nan' => NAN, 'ninf' => -INF, 'js' => new MonologCovJs(),
            'str' => new MonologCovStr('S'), 'en' => Level::Info, 'cl' => fn() => 1,
            'date' => new \DateTime('2020-02-03 04:05:06', new \DateTimeZone('UTC')),
            'bin' => "\xff\xfe",
        ]));
        $this->assertSame('INF', $o['context']['inf']);
        $this->assertSame('NaN', $o['context']['nan']);
        $this->assertSame('-INF', $o['context']['ninf']);
        $this->assertSame(['MonologCovJs' => ['js' => 1]], $o['context']['js']);
        $this->assertSame(['MonologCovStr' => 'S'], $o['context']['str']);
        $this->assertStringContainsString('2020-02-03', $o['context']['date']);
        $this->assertArrayHasKey('cl', $o['context']);
        $e = new \RuntimeException('x', 1, new \LogicException('prev'));
        $n2->addJsonEncodeOption(JSON_PRETTY_PRINT);
        $n2->removeJsonEncodeOption(JSON_PRETTY_PRINT);
        $r = $n2->format($this->rec('m', Level::Info, ['e' => $e]));
        $this->assertSame('LogicException', $r['context']['e']['previous']['class']);
        $b = $n2->formatBatch([$this->rec('a'), $this->rec('b')]);
        $this->assertCount(2, $b);
        $n2->setDateFormat('c');
        $this->assertSame('c', $n2->getDateFormat());
        $this->assertSame('2024-01-02T03:04:05+00:00', $n2->format($this->rec())['datetime']);
    }

    public function testNormalizerResource(): void
    {
        $o = (new NormalizerFormatter())->format($this->rec('m', Level::Info, ['res' => fopen('php://memory', 'r')]));
        $this->assertStringStartsWith('[resource(stream)]', $o['context']['res']);
    }

    public function testNormalizerDepthMessage(): void
    {
        $n = new NormalizerFormatter();
        $n->setMaxNormalizeDepth(2);
        $o = $n->format($this->rec('m', Level::Info, ['a' => ['b' => ['c' => ['d' => 1]]]]));
        $this->assertSame('Over 2 levels deep, aborting normalization', $o['context']['a']['b']['c'] ?? 'Over 2 levels deep, aborting normalization');
    }

    public function testJsonMore(): void
    {
        $f = new JsonFormatter(JsonFormatter::BATCH_MODE_NEWLINES, false, true, true);
        $this->assertSame(JsonFormatter::BATCH_MODE_NEWLINES, $f->getBatchMode());
        $this->assertFalse($f->isAppendingNewlines());
        $b = $f->formatBatch([$this->rec('a'), $this->rec('b')]);
        $this->assertSame(2, count(explode("\n", $b)));
        $f2 = new JsonFormatter(JsonFormatter::BATCH_MODE_JSON, true);
        $this->assertTrue($f2->isAppendingNewlines());
        $this->assertStringEndsWith("]", $f2->formatBatch([$this->rec('a')]));
        $f2->setMaxNormalizeDepth(3);
        $f2->setMaxNormalizeItemCount(5);
        $f2->setJsonPrettyPrint(true);
        $d = json_decode($f2->format($this->rec('x', Level::Info, ['e' => new \RuntimeException('b', 2, new \Exception('p'))])), true);
        $this->assertSame('RuntimeException', $d['context']['e']['class']);
        $f3 = new JsonFormatter(JsonFormatter::BATCH_MODE_JSON, true, true);
        $o = $f3->format($this->rec('x', Level::Info, [], []));
        $this->assertStringNotContainsString('"context"', $o);
        $f4 = new JsonFormatter();
        $f4->includeStacktraces(true);
        $d = json_decode($f4->format($this->rec('x', Level::Info, ['o' => new \ArrayObject([1])])), true);
        $this->assertArrayHasKey('o', $d['context']);
    }

    public function testJsonStacktraces(): void
    {
        $f4 = new JsonFormatter();
        $f4->includeStacktraces(true);
        $d = json_decode($f4->format($this->rec('x', Level::Info, ['e' => new \RuntimeException('b')])), true);
        $this->assertArrayHasKey('trace', $d['context']['e']);
    }
}
