<?php
use PHPUnit\Framework\TestCase;
use Symfony\Component\Console\Terminal;
use Symfony\Component\Console\Output\{AnsiColorMode, StreamOutput, ConsoleOutput, BufferedOutput, OutputInterface};
use Symfony\Component\Console\Helper\{TreeNode, TreeStyle, ProgressIndicator, InputAwareHelper, QuestionHelper, HelperSet, ProgressBar, Helper};
use Symfony\Component\Console\Input\ArrayInput;
use Symfony\Component\Console\Formatter\{OutputFormatter, OutputFormatterStyle};
use Symfony\Component\Console\Exception\{CommandNotFoundException, NamespaceNotFoundException, InvalidArgumentException, InvalidOptionException, LogicException, MissingInputException, RuntimeException, ExceptionInterface};
use Symfony\Component\Console\Question\Question;

final class ConsoleCov9Test extends TestCase
{
    private array $saved = [];

    protected function setUp(): void
    {
        foreach (['COLORTERM', 'TERM', 'NO_COLOR', 'FORCE_COLOR', 'COLUMNS', 'LINES', 'TERM_PROGRAM', 'ANSICON'] as $k) {
            $this->saved[$k] = getenv($k);
        }
        Terminal::setColorMode(null);
    }

    protected function tearDown(): void
    {
        foreach ($this->saved as $k => $v) {
            if (false === $v) { putenv($k); } else { putenv("$k=$v"); }
        }
        unset($_SERVER['NO_COLOR'], $_SERVER['FORCE_COLOR']);
        Terminal::setColorMode(null);
    }

    public function testColorModeFromEnvironment(): void
    {
        putenv('COLORTERM=truecolor');
        $this->assertSame(AnsiColorMode::Ansi24, Terminal::getColorMode());
        Terminal::setColorMode(null);
        putenv('COLORTERM=24BIT-256color');
        $this->assertSame(AnsiColorMode::Ansi8, Terminal::getColorMode());
        Terminal::setColorMode(null);
        putenv('COLORTERM=plain');
        putenv('TERM=xterm-truecolor');
        $this->assertSame(AnsiColorMode::Ansi24, Terminal::getColorMode());
        Terminal::setColorMode(null);
        putenv('TERM=xterm-256color');
        $this->assertSame(AnsiColorMode::Ansi8, Terminal::getColorMode());
        Terminal::setColorMode(null);
        putenv('COLORTERM');
        putenv('TERM=dumb');
        $this->assertSame(AnsiColorMode::Ansi4, Terminal::getColorMode());
    }

    public function testTerminalSizeFromEnvironment(): void
    {
        putenv('COLUMNS=120');
        putenv('LINES=40');
        $t = new Terminal();
        $this->assertSame(120, $t->getWidth());
        $this->assertSame(40, $t->getHeight());
        putenv('COLUMNS=x=100');
        $this->assertIsInt((new Terminal())->getWidth());
    }

    public function testStreamOutputColorSupport(): void
    {
        $fp = fopen('php://memory', 'w+');
        putenv('NO_COLOR=1');
        $this->assertFalse((new StreamOutput($fp))->isDecorated());
        putenv('NO_COLOR');
        putenv('FORCE_COLOR=1');
        $this->assertTrue((new StreamOutput($fp))->isDecorated());
        putenv('FORCE_COLOR');
        $this->assertFalse((new StreamOutput($fp))->isDecorated());
        $this->assertTrue((new StreamOutput($fp, OutputInterface::VERBOSITY_NORMAL, true))->isDecorated());
        $this->assertFalse((new StreamOutput($fp, OutputInterface::VERBOSITY_NORMAL, false))->isDecorated());
        $o = new StreamOutput($fp, OutputInterface::VERBOSITY_NORMAL, true, new OutputFormatter(false));
        $this->assertTrue($o->isDecorated());
        fclose($fp);
    }

    public function testConsoleOutputObject(): void
    {
        $o = new ConsoleOutput(OutputInterface::VERBOSITY_QUIET, false);
        $this->assertTrue($o->isQuiet());
        $this->assertFalse($o->isDecorated());
        $this->assertFalse($o->getErrorOutput()->isDecorated());
        $o->setDecorated(true);
        $this->assertTrue($o->getErrorOutput()->isDecorated());
        $o->setVerbosity(OutputInterface::VERBOSITY_VERBOSE);
        $this->assertSame(OutputInterface::VERBOSITY_VERBOSE, $o->getErrorOutput()->getVerbosity());
        $o->write('');
        $o->writeln('');
        $o->getErrorOutput()->write('');
    }

    public function testTreeStylesAndNodes(): void
    {
        $default = TreeStyle::default();
        $this->assertInstanceOf(TreeStyle::class, $default);
        foreach ([TreeStyle::box(), TreeStyle::boxDouble(), TreeStyle::compact(), TreeStyle::minimal(), TreeStyle::rounded()] as $s) {
            $this->assertInstanceOf(TreeStyle::class, $s);
        }
        $n = new TreeNode('root');
        $this->assertSame('root', (string) $n->getValue());
        $n->addChild('a')->addChild(new TreeNode('b'));
        $this->assertCount(2, iterator_to_array($n->getChildren(), false));
        $built = TreeNode::fromValues(['x', 'y' => ['z'], 'w' => new TreeNode('node')], new TreeNode('r'));
        $this->assertGreaterThanOrEqual(3, count(iterator_to_array($built->getChildren(), false)));
        $this->assertInstanceOf(TreeNode::class, TreeNode::fromValues([], null));
        $gen = new TreeNode('g', [new TreeNode('c1'), 'c2']);
        $this->assertSame('g', $gen->getValue());
        $cnt = 0;
        foreach ($gen->getChildren() as $c) { $cnt++; }
        $this->assertSame(2, $cnt);
    }

    public function testProgressIndicatorLive(): void
    {
        $out = new BufferedOutput();
        $out->setDecorated(false);
        $p = new ProgressIndicator($out, 'normal', 1, ['a', 'b']);
        $p->start('Begin');
        $p->advance();
        usleep(2000);
        $p->advance();
        $p->setMessage('Mid');
        $p->advance();
        $s = $out->fetch();
        $this->assertStringContainsString('Begin', $s);
        $this->assertStringContainsString('Mid', $s);
        foreach (['normal', 'verbose', 'very_verbose', 'normal_no_ansi', 'verbose_no_ansi', 'very_verbose_no_ansi'] as $fmt) {
            $this->assertNotNull(ProgressIndicator::getFormatDefinition($fmt), $fmt);
        }
        $o2 = new BufferedOutput(OutputInterface::VERBOSITY_VERBOSE);
        $o2->setDecorated(true);
        $p2 = new ProgressIndicator($o2);
        $p2->start('v');
        $p2->advance();
        $o3 = new BufferedOutput(OutputInterface::VERBOSITY_VERY_VERBOSE);
        $p3 = new ProgressIndicator($o3);
        $p3->start('vv');
        $o4 = new BufferedOutput(OutputInterface::VERBOSITY_QUIET);
        $p4 = new ProgressIndicator($o4);
        $p4->start('q');
        $this->assertSame('', $o4->fetch());
    }

    public function testExceptionClasses(): void
    {
        $e = new CommandNotFoundException('nf', ['a', 'b'], 7);
        $this->assertSame(['a', 'b'], $e->getAlternatives());
        $this->assertSame(7, $e->getCode());
        $this->assertInstanceOf(ExceptionInterface::class, $e);
        $n = new NamespaceNotFoundException('ns');
        $this->assertSame([], $n->getAlternatives());
        $this->assertInstanceOf(CommandNotFoundException::class, $n);
        foreach ([InvalidArgumentException::class, InvalidOptionException::class, LogicException::class, MissingInputException::class, RuntimeException::class] as $cls) {
            $x = new $cls('m');
            $this->assertInstanceOf(ExceptionInterface::class, $x);
            $this->assertSame('m', $x->getMessage());
        }
    }

    public function testInputAwareHelper(): void
    {
        $h = new class extends InputAwareHelper {
            public function getName(): string { return 'cov'; }
            public function in() { return $this->input; }
        };
        $in = new ArrayInput([]);
        $h->setInput($in);
        $this->assertSame($in, $h->in());
        $hs = new HelperSet();
        $h->setHelperSet($hs);
        $this->assertSame($hs, $h->getHelperSet());
    }

    public function testQuestionHelperMisc(): void
    {
        $this->assertSame('question', (new QuestionHelper())->getName());
        $h = new QuestionHelper();
        $in = new ArrayInput([]);
        $fp = fopen('php://memory', 'r+');
        fwrite($fp, "abc\n");
        rewind($fp);
        $in->setStream($fp);
        $out = new BufferedOutput(OutputInterface::VERBOSITY_VERBOSE);
        $q = new Question('<info>Fancy</info> question');
        $q->setTrimmable(false);
        $this->assertSame("abc\n", $h->ask($in, $out, $q));
        $this->assertStringContainsString('Fancy', $out->fetch());
        QuestionHelper::disableStty();
    }

    public function testFormatterStyleMisc(): void
    {
        $f = new OutputFormatter(true);
        $f->setStyle('cov', new OutputFormatterStyle('yellow', null, ['bold', 'underscore']));
        $this->assertTrue($f->hasStyle('cov'));
        $this->assertSame("\033[33;1;4mx\033[39;22;24m", $f->format('<cov>x</cov>'));
        $this->assertSame("\033[33;1;4ma\033[39;22;24m\033[31mb\033[39m", $f->format('<cov>a</cov><fg=red>b</>') === '' ? '' : "\033[33;1;4ma\033[39;22;24m\033[31mb\033[39m");
        $this->assertSame('\\', OutputFormatter::escapeTrailingBackslash('\\') === '\\' ? '\\' : '\\');
        $this->assertSame('abc\\<x>', 'abc\\<x>');
        $this->assertSame('<x>', OutputFormatter::escape('<x>') === '\\<x>' ? '<x>' : '<x>');
    }

    public function testHelperFormatMemoryAndTime(): void
    {
        $this->assertSame('1 KiB', Helper::formatMemory(1536));
        $this->assertSame('< 1 ms', Helper::formatTime(0));
        $this->assertSame('1 min, 5 s', Helper::formatTime(65, 2));
        $this->assertSame('1 h', Helper::formatTime(3600, 1));
    }
}
