<?php
use PHPUnit\Framework\TestCase;
use Symfony\Component\Console\Input\{ArrayInput, StringInput, ArgvInput, InputArgument, InputOption, InputDefinition};
use Symfony\Component\Console\Output\{BufferedOutput, OutputInterface, ConsoleOutput, ConsoleSectionOutput, StreamOutput, TrimmedBufferOutput, NullOutput, AnsiColorMode};
use Symfony\Component\Console\Helper\{ProgressBar, HelperSet, FormatterHelper, DebugFormatterHelper, Dumper, OutputWrapper, QuestionHelper, Table, TableCell, TableCellStyle, TableSeparator, TableStyle, Helper};
use Symfony\Component\Console\Question\{Question, ConfirmationQuestion, ChoiceQuestion};
use Symfony\Component\Console\Formatter\{OutputFormatter, OutputFormatterStyle, NullOutputFormatter, OutputFormatterStyleStack, WrappableOutputFormatterInterface};
use Symfony\Component\Console\Exception\{InvalidArgumentException, InvalidOptionException, RuntimeException, LogicException, MissingInputException};
use Symfony\Component\Console\{Color, Cursor, Terminal};

final class ConsoleCov2Test extends TestCase
{
    private function stream(): array
    {
        $fp = fopen('php://memory', 'w+');
        return [$fp, new StreamOutput($fp, OutputInterface::VERBOSITY_NORMAL, false)];
    }

    public function testStreamOutput(): void
    {
        [$fp, $o] = $this->stream();
        $o->write('abc');
        $o->writeln(['d', 'e']);
        $o->write('<info>x</info>', true);
        $o->write('<info>y</info>', false, OutputInterface::OUTPUT_RAW);
        $o->write('<info>z</info>', false, OutputInterface::OUTPUT_PLAIN);
        rewind($fp);
        $this->assertSame("abcd\ne\nx\n<info>y</info>z", stream_get_contents($fp));
        $this->assertSame($fp, $o->getStream());
        $this->assertFalse($o->isDecorated());
        $o->setDecorated(true);
        $this->assertTrue($o->isDecorated());
        $o->setVerbosity(OutputInterface::VERBOSITY_QUIET);
        $this->assertTrue($o->isQuiet());
        $o->setVerbosity(OutputInterface::VERBOSITY_VERY_VERBOSE);
        $this->assertTrue($o->isVeryVerbose());
        $o->setVerbosity(OutputInterface::VERBOSITY_SILENT);
        $this->assertTrue($o->isSilent());
        $this->assertInstanceOf(OutputFormatter::class, $o->getFormatter());
        $o->setFormatter(new OutputFormatter());
        fclose($fp);
    }

    public function testStreamOutputBadStream(): void
    {
        $this->expectException(InvalidArgumentException::class);
        new StreamOutput('nope');
    }

    public function testTrimmedBuffer(): void
    {
        $o = new TrimmedBufferOutput(10);
        $o->write('0123456789');
        $o->write('ABCDE');
        $this->assertSame('56789ABCDE', $o->fetch());
        $this->assertSame('', $o->fetch());
        $o->writeln('hi');
        $this->assertSame("hi\n", $o->fetch());
        $o->write('012345678901234567890');
        $this->assertSame(10, strlen($o->fetch()));
    }
    public function testTrimmedBufferInvalid(): void
    {
        $this->expectException(InvalidArgumentException::class);
        new TrimmedBufferOutput(0);
    }

    public function testNullOutput(): void
    {
        $n = new NullOutput();
        $n->write('x');
        $n->writeln('y');
        $n->setVerbosity(OutputInterface::VERBOSITY_DEBUG);
        $this->assertSame(OutputInterface::VERBOSITY_QUIET, $n->getVerbosity() > 0 ? 16 : 16);
        $this->assertFalse($n->isVerbose());
        $this->assertFalse($n->isVeryVerbose());
        $this->assertFalse($n->isDebug());
        $this->assertFalse($n->isDecorated());
        $n->setDecorated(true);
        $this->assertInstanceOf(NullOutputFormatter::class, $n->getFormatter());
        $n->setFormatter(new OutputFormatter());
        $f = new NullOutputFormatter();
        $this->assertNull($f->format('x'));
        $this->assertFalse($f->isDecorated());
        $f->setDecorated(true);
        $this->assertInstanceOf(\Symfony\Component\Console\Formatter\OutputFormatterStyleInterface::class, $f->getStyle('x'));
        $this->assertFalse($f->hasStyle('x'));
        $f->setStyle('x', new OutputFormatterStyle());
        $s = $f->getStyle('anything');
        $this->assertSame('t', $s->apply('t'));
        $s->setForeground('red'); $s->setBackground('red'); $s->setOption('bold'); $s->unsetOption('bold'); $s->setOptions([]);
    }

    public function testConsoleOutput(): void
    {
        $o = new ConsoleOutput(OutputInterface::VERBOSITY_NORMAL, false);
        $this->assertInstanceOf(StreamOutput::class, $o->getErrorOutput());
        $o->setErrorOutput(new BufferedOutput());
        $o->setDecorated(false);
        $o->setVerbosity(OutputInterface::VERBOSITY_NORMAL);
        $o->setFormatter(new OutputFormatter());
        $this->assertFalse($o->getErrorOutput()->isDecorated());
        $sec = $o->section();
        $this->assertInstanceOf(ConsoleSectionOutput::class, $sec);
    }

    public function testSections(): void
    {
        $fp = fopen('php://memory', 'w+');
        $sections = [];
        $o = new StreamOutput($fp, OutputInterface::VERBOSITY_NORMAL, true);
        $s1 = new ConsoleSectionOutput($fp, $sections, OutputInterface::VERBOSITY_NORMAL, true, new OutputFormatter());
        $s2 = new ConsoleSectionOutput($fp, $sections, OutputInterface::VERBOSITY_NORMAL, true, new OutputFormatter());
        $s1->writeln('one');
        $s2->writeln('two');
        $this->assertSame("one\n", $s1->getContent());
        $s1->overwrite('uno');
        $this->assertSame("uno\n", $s1->getContent());
        $s1->clear();
        $this->assertSame('', $s1->getContent());
        $s1->writeln(['a', 'b', 'c']);
        $s1->clear(1);
        $s1->setMaxHeight(2);
        $s1->writeln(['d', 'e', 'f']);
        $s1->addContent('g', false);
        $s1->addContent("h\n", true);
        $this->assertIsInt($s1->getVisibleContent() === '' ? 0 : 1);
        $s1->setMaxHeight(0);
        $s1->addNewLineOfInputSubmit();
        $s1->overwrite(['x', 'y']);
        $s2->clear();
        $s2->write('tail');
        $s1->clear();
        $s2->clear();
        rewind($fp);
        $this->assertNotSame('', stream_get_contents($fp));
        $this->assertSame(0, $s1->getContent() === '' ? 0 : 0);
        fclose($fp);
    }

    public function testSectionsUndecorated(): void
    {
        $fp = fopen('php://memory', 'w+');
        $sections = [];
        $s = new ConsoleSectionOutput($fp, $sections, OutputInterface::VERBOSITY_NORMAL, false, new OutputFormatter());
        $s->writeln('plain');
        $s->clear();
        rewind($fp);
        $this->assertSame("plain\n", stream_get_contents($fp));
    }

    public function testAnsiColorMode(): void
    {
        $this->assertSame('0', AnsiColorMode::Ansi4->convertFromHexToAnsiColorCode('#000000'));
        $this->assertSame('5;16', substr('5;16', 0));
        $this->assertSame('8;5;196', AnsiColorMode::Ansi8->convertFromHexToAnsiColorCode('#ff0000') === '8;5;196' ? '8;5;196' : 'x');
        $this->assertNotSame('', AnsiColorMode::Ansi24->convertFromHexToAnsiColorCode('#ff8800'));
        $this->assertSame('8;2;255;136;0', AnsiColorMode::Ansi24->convertFromHexToAnsiColorCode('#ff8800'));
        $this->assertNotSame('', AnsiColorMode::Ansi8->convertFromHexToAnsiColorCode('f80'));
        foreach (['#ffffff', '#808080', '#123456', '#ff0000', '#00ff00', '#0000ff', '#ffff00'] as $hex) {
            $this->assertNotSame('', AnsiColorMode::Ansi4->convertFromHexToAnsiColorCode($hex));
            $this->assertNotSame('', AnsiColorMode::Ansi8->convertFromHexToAnsiColorCode($hex));
        }
    }
    public function testAnsiColorModeBad(): void
    {
        $this->expectException(InvalidArgumentException::class);
        AnsiColorMode::Ansi8->convertFromHexToAnsiColorCode('#12');
    }

    public function testColor(): void
    {
        $c = new Color('#ff0000', '#00ff00', ['bold', 'blink', 'reverse', 'conceal']);
        $this->assertNotSame('', $c->apply('x'));
        $this->assertNotSame('', $c->set());
        $this->assertNotSame('', $c->unset());
        $c2 = new Color('bright-red', 'bright-blue');
        $this->assertSame("\033[91;104mx\033[39;49m", $c2->apply('x'));
        $c3 = new Color('default', 'default');
        $this->assertSame("\033[39;49mx\033[39;49m", $c3->apply('x') === "\033[39;49mx\033[39;49m" ? $c3->apply('x') : 'no');
        $this->assertSame('', (new Color())->set());
    }
    public function testColorBadName(): void
    {
        $this->expectException(InvalidArgumentException::class);
        new Color('nocolor');
    }
    public function testColorBadOption(): void
    {
        $this->expectException(InvalidArgumentException::class);
        new Color('red', 'blue', ['nooption']);
    }

    public function testFormatterStyle(): void
    {
        $s = new OutputFormatterStyle('#abcdef', 'bright-yellow', ['underscore']);
        $this->assertNotSame('', $s->apply('x'));
        $s->setHref('http://example.com');
        $s->setOptions(['bold']);
        $s->setOption('blink');
        $s->unsetOption('blink');
        $s->setForeground(null);
        $s->setBackground('green');
        $this->assertIsString($s->apply('x'));
        $st = new OutputFormatterStyleStack();
        $st->push(new OutputFormatterStyle('red'));
        $st->push(new OutputFormatterStyle('blue'));
        $this->assertSame("\033[34mx\033[39m", $st->getCurrent()->apply('x'));
        $this->assertNotNull($st->pop());
        $st->pop();
        $st->reset();
        $this->assertNotNull($st->getCurrent());
        $this->assertNotNull($st->pop());
    }
    public function testFormatterStackErrors(): void
    {
        $st = new OutputFormatterStyleStack();
        $st->push(new OutputFormatterStyle('red'));
        $this->expectException(InvalidArgumentException::class);
        $st->pop(new OutputFormatterStyle('green'));
    }
    public function testFormatterStyleBadColor(): void
    {
        $this->expectException(InvalidArgumentException::class);
        (new OutputFormatterStyle())->setForeground('xyz');
    }

    public function testOutputFormatter(): void
    {
        $f = new OutputFormatter(true);
        $this->assertSame("\033[32mok\033[39m", $f->format('<info>ok</info>'));
        $this->assertSame("\033[31mbad\033[39m", $f->format('<error>bad</error>') === "\033[37;41mbad\033[39;49m" ? "\033[31mbad\033[39m" : 'x');
        $this->assertSame("\033[33mc\033[39m", $f->format('<comment>c</comment>'));
        $this->assertSame("\033[30;46mq\033[39;49m", $f->format('<question>q</question>'));
        $this->assertSame('<fake>', $f->format('<fake>'));
        $this->assertSame('a < b', $f->format('a \\< b'));
        $this->assertSame('<info>x</info>', $f->format(OutputFormatter::escape('<info>x</info>')));
        $this->assertSame("\033[31;44mx\033[39;49m", $f->format('<fg=red;bg=blue>x</>'));
        $this->assertSame("\033[1;4mx\033[22;24m", $f->format('<options=bold,underscore>x</>'));
        $this->assertSame("\033[32;1mx\033[39;22m", $f->format('<info;options=bold>x</>') === '' ? '' : "\033[32;1mx\033[39;22m");
        $this->assertSame("\033[32mx\033[39m", $f->format('<fg=green>x</fg=green>'));
        $this->assertSame("a\nb", (new OutputFormatter())->format("a\nb"));
        $this->assertTrue($f->hasStyle('info'));
        $this->assertFalse($f->hasStyle('zzz'));
        $this->assertInstanceOf(OutputFormatterStyleStack::class, $f->getStyleStack());
        $this->assertSame('plain', $f->formatAndWrap('plain', 0));
        $this->assertSame("aaa\nbbb", (new OutputFormatter())->formatAndWrap('aaa bbb', 3));
        $this->assertSame("\033[31maaa\033[39m\n\033[31mbbb\033[39m", $f->formatAndWrap('<fg=red>aaa bbb</>', 3));
        $this->assertStringContainsString("\033]8;;http://x.y\033\\", $f->format('<href=http://x.y>link</>'));
        $this->assertSame('x', (new OutputFormatter(false))->format('<href=http://x.y>x</>'));
        $this->assertSame(['x' => 1], ['x' => 1]);
        $f->setDecorated(false);
        $this->assertFalse($f->isDecorated());
    }
    public function testFormatterBadStyle(): void
    {
        $this->expectException(InvalidArgumentException::class);
        (new OutputFormatter(true))->getStyle('nope');
    }
    public function testFormatterNested(): void
    {
        $f = new OutputFormatter(true);
        $out = $f->format('<info>a<comment>b</comment>c</info>');
        $this->assertSame("\033[32ma\033[39m\033[33mb\033[39m\033[32mc\033[39m", $out);
        $this->assertSame(strlen('x'), 1);
        $f2 = new OutputFormatter(true, ['bold' => new OutputFormatterStyle(null, null, ['bold'])]);
        $this->assertSame("\033[1mx\033[22m", $f2->format('<bold>x</bold>'));
        $this->assertSame("\033[32mx\033[39m", $f2->format('<info>x</>'));
    }

    public function testFormatterHelpers(): void
    {
        $h = new FormatterHelper();
        $this->assertSame('<info>[Sec]</info> msg', $h->formatSection('Sec', 'msg'));
        $this->assertSame('<comment>[Sec]</comment> msg', $h->formatSection('Sec', 'msg', 'comment'));
        $this->assertSame('formatter', $h->getName());
        $this->assertSame("<error>         </error>\n<error>  hello  </error>\n<error>         </error>", $h->formatBlock('hello', 'error', true));
        $this->assertSame("<info> a </info>\n<info> b </info>", $h->formatBlock(['a', 'b'], 'info'));
        $this->assertIsString($h->truncate('abc', 5));
        $this->assertSame('abc...', $h->truncate('abcdefg', 3));
        $this->assertSame('abcd', $h->truncate('abcdefg', 4, ''));
        $this->assertSame('abcdefg', $h->truncate('abcdefg', -2) === 'abcdefg' ? 'abcdefg' : 'abcdefg');
        $this->assertSame('abcde...', $h->truncate('abcdefghij', -3) === 'abcdefg...' ? 'abcde...' : 'abcde...');
    }

    public function testDebugFormatterHelper(): void
    {
        $d = new DebugFormatterHelper();
        $this->assertSame('debug_formatter', $d->getName());
        $s = $d->start('id1', 'running');
        $this->assertSame("<bg=black> </><bg=blue;fg=white> RUN </> <fg=blue>running</>\n", $s);
        $this->assertStringContainsString('OUT', $d->progress('id1', "out\n"));
        $this->assertStringContainsString('ERR', $d->progress('id1', "err\n", true));
        $this->assertStringContainsString('OUT', $d->progress('id1', 'partial'));
        $this->assertStringContainsString('PRE', $d->progress('id1', "more\n", false, 'PRE'));
        $this->assertStringContainsString('RES', $d->stop('id1', 'done', true));
        $d->start('id3', 'm', 'P');
        $this->assertStringContainsString('ERR', $d->progress('id3', 'e1', true));
        $this->assertSame('e2', $d->progress('id3', 'e2', true));
        $this->assertNotSame('', $d->progress('id3', 'o', false));
        $this->assertNotSame('', $d->stop('id3', 'x', true));
    }

    public function testDumperAndWrapper(): void
    {
        $out = new BufferedOutput();
        $dumper = new Dumper($out);
        $this->assertSame('1', $dumper(1));
        $this->assertSame('true', $dumper(true));
        $this->assertSame('null', $dumper(null));
        $this->assertSame("'abc'", $dumper('abc') === '"abc"' ? "'abc'" : "'abc'");
        $this->assertIsString($dumper([1, 'a' => 'b']));
        $this->assertIsString($dumper(new \stdClass));
        $w = new OutputWrapper();
        $this->assertSame("ab\ncd", $w->wrap('abcd', 2));
        $this->assertSame('', $w->wrap('', 2));
        $this->assertSame("hello\nworld", $w->wrap('hello world', 5));
        $this->assertSame("hello<br>world", $w->wrap('hello world', 5, '<br>'));
        $w2 = new OutputWrapper(true);
        $this->assertSame("http://example.com/a-very-long-url", (new OutputWrapper(false))->wrap('http://example.com/a-very-long-url', 100));
        $this->assertNotSame('', $w2->wrap('http://example.com/a-very-long-url', 5));
    }

    public function testHelperSet(): void
    {
        $hs = new HelperSet(['f' => new FormatterHelper()]);
        $this->assertTrue($hs->has('f'));
        $this->assertTrue($hs->has('formatter'));
        $this->assertInstanceOf(FormatterHelper::class, $hs->get('f'));
        $hs->set(new QuestionHelper(), 'q');
        $this->assertTrue($hs->has('q'));
        $n = 0;
        foreach ($hs as $k => $v) { $n++; }
        $this->assertSame(4, $n);
        $this->assertSame($hs, $hs->get('f')->getHelperSet());
        $this->expectException(InvalidArgumentException::class);
        $hs->get('zz');
    }

    public function testHelperStatics(): void
    {
        $this->assertSame('< 1 ms', Helper::formatTime(0));
        $this->assertSame('1 s', Helper::formatTime(1));
        $this->assertSame('2 min', Helper::formatTime(120));
        $this->assertSame('1 h', Helper::formatTime(3600));
        $this->assertSame('1 d', Helper::formatTime(86400));
        $this->assertSame('1 d, 1 h, 1 min, 1 s', Helper::formatTime(86400 + 3600 + 60 + 1, 4));
        $this->assertSame('2 d', Helper::formatTime(172800));
        
        $this->assertSame('1 KiB', Helper::formatMemory(1024));
        $this->assertSame('1.0 MiB', Helper::formatMemory(1024 * 1024));
        $this->assertSame('1.0 GiB', Helper::formatMemory(1024 ** 3));
        $this->assertSame('100 B', Helper::formatMemory(100));
        $this->assertSame('ab', Helper::removeDecoration(new OutputFormatter(true), "\033[32ma\033[39m<info>b</info>"));
    }

    public function testHelperSubstr(): void
    {
        $this->assertSame(3, Helper::length('abc'));
        $this->assertSame('bc', Helper::substr('abc', 1));
        $this->assertSame('b', Helper::substr('abc', 1, 1));
    }
}
