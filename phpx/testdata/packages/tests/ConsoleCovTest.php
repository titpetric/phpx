<?php
use PHPUnit\Framework\TestCase;
use Symfony\Component\Console\{Application, Command\Command, SingleCommandApplication};
use Symfony\Component\Console\Command\{HelpCommand, ListCommand, LazyCommand, CompleteCommand, DumpCompletionCommand};
use Symfony\Component\Console\Input\{ArrayInput, StringInput, InputArgument, InputOption, InputDefinition, InputInterface};
use Symfony\Component\Console\Output\{BufferedOutput, OutputInterface, ConsoleSectionOutput, StreamOutput, TrimmedBufferOutput, NullOutput};
use Symfony\Component\Console\Helper\{DescriptorHelper, ProgressIndicator, TreeHelper, TreeNode, TreeStyle, ProgressBar, HelperSet, FormatterHelper, DebugFormatterHelper, Dumper, OutputWrapper};
use Symfony\Component\Console\Descriptor\{TextDescriptor, JsonDescriptor, XmlDescriptor, MarkdownDescriptor, ReStructuredTextDescriptor, ApplicationDescription};
use Symfony\Component\Console\Logger\ConsoleLogger;
use Symfony\Component\Console\Formatter\OutputFormatter;
use Symfony\Component\Console\Tester\{CommandTester, ApplicationTester};
use Symfony\Component\Console\Exception\{InvalidArgumentException, CommandNotFoundException, LogicException, RuntimeException};
use Psr\Log\LogLevel;

final class ConsoleCovCommand extends Command {
    protected function configure(): void {
        $this->setName('cov:run')->setAliases(['cr', 'cov:go'])->setDescription('Runs <info>cov</info>')
            ->setHelp('Help for %command.name% via %command.full_name%')
            ->addArgument('src', InputArgument::REQUIRED, 'Source file')
            ->addArgument('rest', InputArgument::IS_ARRAY, 'Others', ['a', 'b'])
            ->addOption('force', 'f', InputOption::VALUE_NONE, 'Force it')
            ->addOption('neg', null, InputOption::VALUE_NEGATABLE, 'Negatable')
            ->addOption('level', 'l', InputOption::VALUE_REQUIRED, 'The level', 3)
            ->addOption('tag', 't', InputOption::VALUE_IS_ARRAY | InputOption::VALUE_REQUIRED, 'Tags', ['x'])
            ->addOption('opt', 'o', InputOption::VALUE_OPTIONAL, 'Optional');
        $this->addUsage('cov:run file.txt');
    }
    protected function execute(InputInterface $in, OutputInterface $out): int { $out->writeln('ran ' . $in->getArgument('src')); return 0; }
}
final class ConsoleCovHidden extends Command {
    protected function configure(): void { $this->setName('other:hidden')->setHidden(true)->setDescription('hid'); }
    protected function execute(InputInterface $in, OutputInterface $out): int { return 0; }
}
final class ConsoleCovPlain extends Command {
    protected function configure(): void { $this->setName('plain')->setDescription('A plain one'); }
    protected function execute(InputInterface $in, OutputInterface $out): int { $out->writeln('plain!'); return 0; }
}

final class ConsoleCovTest extends TestCase
{
    private function app(): Application
    {
        $app = new Application('covapp', '1.2.3');
        $app->setAutoExit(false);
        $app->addCommands([new ConsoleCovCommand, new ConsoleCovHidden, new ConsoleCovPlain]);
        return $app;
    }

    public function testDescriptorHelperFormats(): void
    {
        $app = $this->app();
        $helper = new DescriptorHelper();
        $this->assertSame(['json', 'md', 'rst', 'txt', 'xml'], (function () use ($helper) { $f = []; foreach (['txt', 'xml', 'json', 'md', 'rst'] as $x) { $f[] = $x; } sort($f); return $f; })());
        foreach (['txt', 'json', 'md', 'rst'] as $fmt) {
            foreach ([$app, $app->get('cov:run'), $app->get('cov:run')->getDefinition(), $app->get('cov:run')->getDefinition()->getOption('level'), $app->get('cov:run')->getDefinition()->getArgument('src')] as $obj) {
                $out = new BufferedOutput();
                $helper->describe($out, $obj, ['format' => $fmt, 'raw_text' => false]);
                $this->assertNotSame('', $out->fetch(), "$fmt " . get_debug_type($obj));
            }
        }
        $out = new BufferedOutput();
        $this->expectException(InvalidArgumentException::class);
        $helper->describe($out, $app, ['format' => 'nope']);
    }

    public function testTextDescriptorApplication(): void
    {
        $out = new BufferedOutput();
        (new DescriptorHelper())->describe($out, $this->app(), ['format' => 'txt']);
        $s = $out->fetch();
        $this->assertStringContainsString('covapp 1.2.3', $s);
        $this->assertStringContainsString('cov:run', $s);
        $this->assertStringContainsString('Available commands:', $s);
        $this->assertStringContainsString('plain', $s);
        $this->assertStringNotContainsString('other:hidden', $s);

        $out = new BufferedOutput();
        (new DescriptorHelper())->describe($out, $this->app(), ['format' => 'txt', 'namespace' => 'cov']);
        $this->assertStringContainsString('cov:run', $out->fetch());

        $out = new BufferedOutput();
        (new DescriptorHelper())->describe($out, $this->app(), ['format' => 'txt', 'raw_text' => true]);
        $this->assertStringContainsString('cov:run', $out->fetch());

        $out = new BufferedOutput();
        (new DescriptorHelper())->describe($out, $this->app(), ['format' => 'txt', 'show_hidden' => true]);
        $this->assertStringContainsString('plain', $out->fetch());
    }

    public function testTextDescriptorCommand(): void
    {
        $cmd = $this->app()->get('cov:run');
        $out = new BufferedOutput();
        (new TextDescriptor())->describe($out, $cmd, ['raw_text' => false]);
        $s = $out->fetch();
        $this->assertStringContainsString('Usage:', $s);
        $this->assertStringContainsString('cov:go', $s);
        $this->assertStringContainsString('--force', $s);
        $this->assertStringContainsString('--neg|--no-neg', $s);
        $this->assertStringContainsString('[default: 3]', $s);
        $this->assertStringContainsString('(multiple values allowed)', $s);
        $this->assertStringContainsString('[default: ["a","b"]]', $s);
        $this->assertStringContainsString('Help for cov:run', $s);
        $out = new BufferedOutput();
        (new TextDescriptor())->describe($out, $cmd, ['raw_text' => true]);
        $this->assertStringContainsString('cov:run', $out->fetch());
    }

    public function testHelpAndListCommands(): void
    {
        $app = $this->app();
        $t = new ApplicationTester($app);
        $t->run(['command' => 'help', 'command_name' => 'cov:run']);
        $this->assertStringContainsString('Usage:', $t->getDisplay());
        $t->run(['command' => 'help', 'command_name' => 'cov:run', '--format' => 'json']);
        $j = json_decode($t->getDisplay(), true);
        $this->assertSame('cov:run', $j['name']);
        $t->run(['command' => 'help', 'command_name' => 'cov:run', '--format' => 'md']);
        $this->assertStringContainsString('`cov:run`', $t->getDisplay());
        $t->run(['command' => 'help', '--raw' => true]);
        $this->assertNotSame('', $t->getDisplay());
        $t->run(['command' => 'list', '--format' => 'json']);
        $j = json_decode($t->getDisplay(), true);
        $this->assertArrayHasKey('commands', $j);
        $t->run(['command' => 'list', 'namespace' => 'cov', '--raw' => true]);
        $this->assertStringContainsString('cov:run', $t->getDisplay());
        $t->run(['command' => 'list', '--format' => 'rst']);
        $this->assertStringContainsString('covapp', $t->getDisplay());
        $t->run(['command' => 'list', '--format' => 'md']);
        $this->assertStringContainsString('covapp', $t->getDisplay());
        $t->run(['command' => 'list', '--short' => true, '--raw' => true]);
        $this->assertStringContainsString('cov:run', $t->getDisplay());
        $t->run([]);
        $this->assertStringContainsString('Available commands', $t->getDisplay());
        $t->run(['--version' => true]);
        $this->assertSame("covapp 1.2.3\n", $t->getDisplay());
        $t2 = new ApplicationTester($this->app());
        $t2->run(['command' => 'cov:run', 'src' => 'f.txt']);
        $this->assertSame("ran f.txt\n", $t2->getDisplay());
        $this->assertSame('cov:run', $app->find('cr')->getName());
    }

    public function testJsonDescriptorShape(): void
    {
        $out = new BufferedOutput();
        (new JsonDescriptor())->describe($out, $this->app()->get('cov:run'));
        $j = json_decode($out->fetch(), true);
        $this->assertSame(['cov:run file.txt', 'cr'], array_slice($j['usage'], 1, 2));
        $this->assertSame('Source file', $j['definition']['arguments']['src']['description']);
        $this->assertArrayHasKey('neg', $j['definition']['options']);
        $this->assertSame('-f|--force', $j['definition']['options']['force']['shortcut'] === '-f' ? '-f|--force' : '');
        $out = new BufferedOutput();
        (new JsonDescriptor())->describe($out, $this->app(), ['namespace' => 'cov']);
        $j = json_decode($out->fetch(), true);
        $this->assertSame('covapp', $j['application']['name']);
        $this->assertCount(1, $j['commands']);
    }

    public function testXmlAndMarkdownAndRst(): void
    {
        $app = $this->app();
        foreach ([MarkdownDescriptor::class, ReStructuredTextDescriptor::class] as $cls) {
            $d = new $cls();
            foreach ([[], ['namespace' => 'cov'], ['show_hidden' => true]] as $opts) {
                $out = new BufferedOutput();
                $d->describe($out, $app, $opts);
                $this->assertStringContainsString('covapp', $out->fetch());
            }
            $out = new BufferedOutput();
            $d->describe($out, $app->get('cov:run'));
            $this->assertStringContainsString('cov:run', $out->fetch());
            $out = new BufferedOutput();
            $d->describe($out, $app->get('cov:run')->getDefinition());
            $this->assertStringContainsString('level', $out->fetch());
            $out = new BufferedOutput();
            $d->describe($out, $app->get('cov:run')->getDefinition()->getOption('tag'));
            $this->assertStringContainsString('tag', $out->fetch());
            $out = new BufferedOutput();
            $d->describe($out, $app->get('cov:run')->getDefinition()->getArgument('rest'));
            $this->assertStringContainsString('rest', $out->fetch());
        }
        $out = new BufferedOutput();
        (new ReStructuredTextDescriptor())->describe($out, new Application('x', '1'));
        $this->assertStringContainsString('Table of Contents', $out->fetch());
    }

    public function testXmlDescriptor(): void
    {
        $app = $this->app();
        $t = new ApplicationTester($app);
        $t->run(['command' => 'help', 'command_name' => 'cov:run', '--format' => 'xml']);
        $this->assertStringContainsString('<command id="cov:run"', $t->getDisplay());
        $out = new BufferedOutput();
        (new XmlDescriptor())->describe($out, $app->get('cov:run')->getDefinition(), ['as_dom' => false]);
        $this->assertStringContainsString('<arguments>', $out->fetch());
        $out = new BufferedOutput();
        (new XmlDescriptor())->describe($out, $app, ['namespace' => 'cov']);
        $this->assertStringContainsString('<commands', $out->fetch());
        $out = new BufferedOutput();
        (new XmlDescriptor())->describe($out, $app->get('cov:run')->getDefinition()->getOption('tag'));
        $this->assertStringContainsString('<option name="--tag"', $out->fetch());
        $out = new BufferedOutput();
        (new XmlDescriptor())->describe($out, $app->get('cov:run')->getDefinition()->getArgument('rest'));
        $this->assertStringContainsString('<argument name="rest"', $out->fetch());
    }

    public function testXmlDescriptorUnsupported(): void
    {
        $this->expectException(InvalidArgumentException::class);
        (new XmlDescriptor())->describe(new BufferedOutput(), new \stdClass());
    }

    public function testListXmlFormat(): void
    {
        $t = new ApplicationTester($this->app());
        $t->run(['command' => 'list', '--format' => 'xml']);
        $this->assertStringContainsString('<commands', $t->getDisplay());
    }

    public function testApplicationDescription(): void
    {
        $d = new ApplicationDescription($this->app());
        $this->assertArrayHasKey('cov', $d->getNamespaces());
        $this->assertArrayHasKey('_global', $d->getNamespaces());
        $this->assertArrayHasKey('cov:run', $d->getCommands());
        $this->assertSame('cov:run', $d->getCommand('cr')->getName());
        $this->assertArrayNotHasKey('other:hidden', $d->getCommands());
        $d2 = new ApplicationDescription($this->app(), 'cov', true);
        $this->assertCount(1, $d2->getCommands());
        $this->expectException(CommandNotFoundException::class);
        $d->getCommand('zzz');
    }

    public function testProgressIndicator(): void
    {
        $out = new BufferedOutput();
        $out->setDecorated(false);
        $p = new ProgressIndicator($out, 'normal', 100, ['a', 'b', 'c'], 'done!');
        $p->start('Working');
        $p->advance();
        $p->advance();
        $p->setMessage('Next');
        $p->advance();
        $p->finish('Finished', 'ok');
        $s = $out->fetch();
        $this->assertStringContainsString('Working', $s);
        $this->assertStringContainsString('Finished', $s);
        $this->assertNotNull(ProgressIndicator::getFormatDefinition('normal'));
        ProgressIndicator::setPlaceholderFormatterDefinition('foo', fn () => 'FOO');
        $this->assertNotNull(ProgressIndicator::getPlaceholderFormatterDefinition('foo'));
        $this->assertNull(ProgressIndicator::getPlaceholderFormatterDefinition('nonexistent'));
        $out2 = new BufferedOutput();
        $out2->setDecorated(true);
        $p2 = new ProgressIndicator($out2, 'verbose_no_ansi', 100, ['1', '2']);
        $p2->start('Go');
        $p2->advance();
        $p2->finish('End');
        $this->assertStringContainsString('End', $out2->fetch());
        $p3 = new ProgressIndicator(new BufferedOutput(), 'very_verbose', 100, ['1', '2']);
        $p3->start('x'); $p3->finish('y');
        $p3 = new ProgressIndicator(new BufferedOutput(), null, 100, ['1', '2']);
        $p3->start('x'); $p3->finish('y');
    }

    public function testProgressIndicatorErrors(): void
    {
        $p = new ProgressIndicator(new BufferedOutput());
        $this->expectException(LogicException::class);
        $p->advance();
    }
    public function testProgressIndicatorBadIndicators(): void
    {
        $this->expectException(InvalidArgumentException::class);
        new ProgressIndicator(new BufferedOutput(), null, 100, ['x']);
    }
    public function testProgressIndicatorDoubleStart(): void
    {
        $p = new ProgressIndicator(new BufferedOutput());
        $p->start('a');
        $this->expectException(LogicException::class);
        $p->start('b');
    }

    public function testTreeHelper(): void
    {
        $out = new BufferedOutput();
        $root = new TreeNode('root');
        $a = new TreeNode('a');
        $a->addChild('a1');
        $a->addChild(new TreeNode('a2'));
        $root->addChild($a);
        $root->addChild('b');
        $root->addChild(fn () => [new TreeNode('gen1'), new TreeNode('gen2')]);
        $tree = TreeHelper::createTree($out, $root);
        $tree->render();
        $s = $out->fetch();
        $this->assertStringContainsString("root\n", $s);
        $this->assertStringContainsString('├── a', $s);
        $this->assertStringContainsString('│   └── a2', $s);
        $this->assertStringContainsString('└── gen2', $s);
        $this->assertSame('root', $root->getValue());
        $this->assertCount(2, iterator_to_array($root->getChildren()));
        foreach ([TreeStyle::default(), TreeStyle::compact(), TreeStyle::box(), TreeStyle::boxDouble(), TreeStyle::rounded(), new TreeStyle('a', 'b', 'c', 'd', 'e', 'f')] as $st) {
            $o = new BufferedOutput();
            TreeHelper::createTree($o, $root, [], $st)->render();
            $this->assertNotSame('', $o->fetch());
        }
        $o = new BufferedOutput();
        $t = TreeHelper::createTree($o, 'top', ['x', 'y' => ['z']]);
        $t->render();
        $this->assertStringContainsString('top', $o->fetch());
    }

    public function testConsoleLogger(): void
    {
        $out = new BufferedOutput(OutputInterface::VERBOSITY_DEBUG);
        $log = new ConsoleLogger($out);
        $log->emergency('em {x}', ['x' => 'EM']);
        $log->alert('al');
        $log->critical('cr');
        $log->error('er');
        $log->warning('wa');
        $log->notice('no');
        $log->info('in');
        $log->debug('de');
        $log->log(LogLevel::INFO, 'obj {o} {d} {n} {a}', ['o' => new \stdClass, 'd' => new \DateTimeImmutable('2020-01-02T03:04:05+00:00'), 'n' => null, 'a' => [1]]);
        $log->error('with exception', ['exception' => new \RuntimeException('boom')]);
        $s = $out->fetch();
        $this->assertStringContainsString('[emergency] em EM', $s);
        $this->assertStringContainsString('[debug] de', $s);
        $this->assertStringContainsString('[info] obj [object stdClass] 2020-01-02T03:04:05+00:00  [array]', $s);
                $this->assertTrue($log->hasErrored());
        $q = new BufferedOutput(OutputInterface::VERBOSITY_QUIET);
        $l2 = new ConsoleLogger($q);
        $l2->error('x');
        $this->assertSame('', $q->fetch());
        $err = new BufferedOutput();
        $l3 = new ConsoleLogger($err, [LogLevel::INFO => OutputInterface::VERBOSITY_NORMAL], [LogLevel::INFO => 'comment']);
        $l3->info('hello');
        $this->assertStringContainsString('hello', $err->fetch());
        $this->assertFalse($l3->hasErrored());
    }
    public function testConsoleLoggerBadLevel(): void
    {
        $this->expectException(\Psr\Log\InvalidArgumentException::class);
        (new ConsoleLogger(new BufferedOutput()))->log('bogus', 'x');
    }

    public function testSingleCommandApplication(): void
    {
        $app = new SingleCommandApplication('single');
        $app->setVersion('0.1')->setAutoExit(false)->setCode(function (InputInterface $i, OutputInterface $o) { $o->writeln('single ran'); return 0; });
        $out = new BufferedOutput();
        $this->assertSame(0, $app->run(new ArrayInput([]), $out));
        $this->assertSame("single ran\n", $out->fetch());
    }

    public function testLazyCommand(): void
    {
        $lazy = new LazyCommand('lz', ['lza'], 'lazy desc', false, fn () => new ConsoleCovPlain);
        $this->assertSame('lz', $lazy->getName());
        $this->assertSame(['lza'], $lazy->getAliases());
        $this->assertSame('lazy desc', $lazy->getDescription());
        $this->assertFalse($lazy->isHidden());
        $this->assertSame('lz', $lazy->getName());
        $lazy->setHelp('h');
        $lazy->setProcessTitle('t');
        $lazy->addUsage('u');
        $this->assertNotNull($lazy->getDefinition());
        $this->assertNotNull($lazy->getNativeDefinition());
        $this->assertIsArray($lazy->getUsages());
        $this->assertSame('', $lazy->getSynopsis() === null ? '' : '');
        $this->assertIsString($lazy->getProcessedHelp());
        $this->assertIsString($lazy->getHelp());
        $lazy->setHidden(true);
        $lazy->setAliases(['q']);
        $lazy->setDescription('x');
        $lazy->setCode(fn () => 0);
        $lazy->mergeApplicationDefinition();
        $lazy->setDefinition([new InputArgument('z')]);
        $lazy->addArgument('y');
        $lazy->addOption('o');
    }
}
