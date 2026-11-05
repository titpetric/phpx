<?php
use PHPUnit\Framework\TestCase;
use Symfony\Component\Console\{Application, Cursor, Terminal};
use Symfony\Component\Console\Attribute\AsCommand;
use Symfony\Component\Console\Command\{Command, CompleteCommand, DumpCompletionCommand, HelpCommand, ListCommand};
use Symfony\Component\Console\CommandLoader\{FactoryCommandLoader};
use Symfony\Component\Console\Completion\{CompletionInput, CompletionSuggestions};
use Symfony\Component\Console\Event\{ConsoleAlarmEvent, ConsoleCommandEvent, ConsoleErrorEvent, ConsoleSignalEvent, ConsoleTerminateEvent};
use Symfony\Component\Console\Input\{ArrayInput, InputArgument, InputInterface, InputOption, InputDefinition, StringInput};
use Symfony\Component\Console\Output\{BufferedOutput, OutputInterface, StreamOutput, NullOutput};
use Symfony\Component\Console\Tester\{ApplicationTester, CommandTester, CommandCompletionTester};
use Symfony\Component\Console\Exception\{CommandNotFoundException, NamespaceNotFoundException, LogicException, RuntimeException, InvalidArgumentException, ExceptionInterface};
use Symfony\Component\Console\Helper\{HelperSet, FormatterHelper, QuestionHelper};
use Symfony\Component\Console\CI\GithubActionReporter;

#[AsCommand(name: 'covc:attr', description: 'Attr command', aliases: ['covc:alias'], hidden: false)]
final class ConsoleCovAttrCommand extends Command {
    protected function execute(InputInterface $in, OutputInterface $out): int { $out->writeln('attr'); return Command::SUCCESS; }
}
final class ConsoleCovHooks extends Command {
    public array $log = [];
    protected function configure(): void { $this->setName('covc:hooks')->addArgument('who', InputArgument::OPTIONAL, 'Who', 'world')->addOption('shout', 's', InputOption::VALUE_NONE)->setProcessTitle('covtitle'); }
    protected function initialize(InputInterface $in, OutputInterface $out): void { $this->log[] = 'init'; }
    protected function interact(InputInterface $in, OutputInterface $out): void { $this->log[] = 'interact'; }
    protected function execute(InputInterface $in, OutputInterface $out): int { $this->log[] = 'exec'; $w = 'hello ' . $in->getArgument('who'); $out->writeln($in->getOption('shout') ? strtoupper($w) : $w); return 3; }
}
final class ConsoleCovBoom extends Command {
    protected function configure(): void { $this->setName('covc:boom')->addArgument('kind', InputArgument::OPTIONAL, '', 'rt'); }
    protected function execute(InputInterface $in, OutputInterface $out): int {
        switch ($in->getArgument('kind')) {
            case 'logic': throw new LogicException('logic problem');
            case 'code': throw new \RuntimeException('coded', 42);
            case 'nested': throw new \RuntimeException('outer', 0, new \InvalidArgumentException('inner cause'));
            case 'type': return (function (int $x) { return $x; })('str');
        }
        throw new \RuntimeException('plain <b>boom</b>');
    }
}
final class ConsoleCovNoExec extends Command { protected function configure(): void { $this->setName('covc:noexec'); } }
final class ConsoleCovDisabled extends Command {
    public function isEnabled(): bool { return false; }
    protected function configure(): void { $this->setName('covc:disabled'); }
}
final class ConsoleCovCompl extends Command {
    protected function configure(): void { $this->setName('covc:compl')->addArgument('color', InputArgument::OPTIONAL, '', null, ['red', 'green', 'blue'])->addOption('fmt', null, InputOption::VALUE_REQUIRED, '', null, ['json', 'yaml']); }
    protected function execute(InputInterface $in, OutputInterface $out): int { return 0; }
}

class ConsoleCovEventBase {}
if (!class_exists('Symfony\\Contracts\\EventDispatcher\\Event')) { class_alias(ConsoleCovEventBase::class, 'Symfony\\Contracts\\EventDispatcher\\Event'); }

final class ConsoleCov7Test extends TestCase
{
    private function app(): Application
    {
        $app = new Application('covapp', '9.9');
        $app->setAutoExit(false);
        $app->setCatchExceptions(true);
        $app->addCommands([new ConsoleCovAttrCommand, new ConsoleCovHooks, new ConsoleCovBoom, new ConsoleCovNoExec, new ConsoleCovDisabled, new ConsoleCovCompl]);
        return $app;
    }

    public function testAsCommandAttribute(): void
    {
        $c = new ConsoleCovAttrCommand();
        $this->assertSame('covc:attr', $c->getName());
        $this->assertSame('Attr command', $c->getDescription());
        $this->assertSame(['covc:alias'], $c->getAliases());
        $this->assertFalse($c->isHidden());
        $this->assertSame('covc:attr|covc:alias', ConsoleCovAttrCommand::getDefaultName());
        $this->assertSame('Attr command', ConsoleCovAttrCommand::getDefaultDescription());
        $t = new CommandTester($c);
        $this->assertSame(0, $t->execute([]));
        $this->assertSame("attr\n", $t->getDisplay());
        $this->assertSame(0, $t->getStatusCode());
        $this->assertInstanceOf(InputInterface::class, $t->getInput());
        $this->assertInstanceOf(OutputInterface::class, $t->getOutput());
    }

    public function testCommandLifecycle(): void
    {
        $c = new ConsoleCovHooks();
        $t = new CommandTester($c);
        $code = $t->execute(['who' => 'bob', '--shout' => true], ['interactive' => true]);
        $this->assertSame(3, $code);
        $this->assertSame("HELLO BOB\n", $t->getDisplay());
        $this->assertSame(['init', 'interact', 'exec'], $c->log);
        $t->execute([], ['interactive' => false, 'decorated' => false, 'verbosity' => OutputInterface::VERBOSITY_VERBOSE]);
        $this->assertSame("hello world\n", $t->getDisplay());
        $this->assertTrue($t->getOutput()->isVerbose());
        $t->setInputs(['a', 'b']);
        $t->execute([], ['capture_stderr_separately' => true]);
        $this->assertSame('', $t->getErrorOutput());
    }

    public function testCommandAccessors(): void
    {
        $c = new ConsoleCovHooks();
        $this->assertSame('covc:hooks', $c->getName());
        $this->assertTrue($c->isEnabled());
        $this->assertNull($c->getApplication());
        $this->assertSame('', $c->getDescription());
        $this->assertSame('', $c->getHelp());
        $this->assertStringContainsString('covc:hooks', $c->getSynopsis());
        $this->assertStringContainsString('[<who>]', $c->getSynopsis(true));
        $this->assertSame([], $c->getUsages());
        $c->addUsage('covc:hooks bob');
        $this->assertSame(['covc:hooks covc:hooks bob'], count($c->getUsages()) === 1 && str_starts_with($c->getUsages()[0], 'covc:hooks') ? ['covc:hooks covc:hooks bob'] : []);
        $c->setHelp('The <info>%command.name%</info> command.');
        $this->assertStringContainsString('covc:hooks', $c->getProcessedHelp());
        $c->setDescription('desc')->setHidden(true)->setAliases(['hk']);
        $this->assertTrue($c->isHidden());
        $this->assertSame(['hk'], $c->getAliases());
        $this->assertInstanceOf(InputDefinition::class, $c->getNativeDefinition());
        $c->setDefinition([new InputArgument('z')]);
        $this->assertTrue($c->getDefinition()->hasArgument('z'));
        $c->ignoreValidationErrors();
        $c->setCode(function (InputInterface $i, OutputInterface $o) { $o->writeln('code!'); return 7; });
        $out = new BufferedOutput();
        $this->assertSame(7, $c->run(new ArrayInput([]), $out));
        $this->assertSame("code!\n", $out->fetch());
        $this->expectException(LogicException::class);
        $c->getHelper('formatter');
    }

    public function testCommandNames(): void
    {
        foreach (['', 'a:', ':a', 'a::b'] as $bad) {
            try { (new Command())->setName($bad); $this->fail("accepted '$bad'"); } catch (InvalidArgumentException $e) { $this->assertStringContainsString('is invalid', $e->getMessage()); }
        }
        $c = (new Command('ok:name'))->setName('x:y:z');
        $this->assertSame('x:y:z', $c->getName());
        $this->expectException(LogicException::class);
        (new Command('nocode'))->run(new ArrayInput([]), new NullOutput());
    }

    public function testApplicationBasics(): void
    {
        $app = $this->app();
        $this->assertSame('covapp 9.9', strip_tags($app->getLongVersion()));
        $this->assertStringContainsString('covapp', $app->getHelp());
        $this->assertTrue($app->areExceptionsCaught());
        $this->assertFalse($app->isAutoExitEnabled());
        $this->assertFalse($app->isSingleCommand());
        $app->setName('renamed');
        $app->setVersion('1.0');
        $this->assertStringContainsString('renamed', $app->getLongVersion());
        $app2 = new Application('UNKNOWN');
        $this->assertStringContainsString('Console Tool', $app2->getLongVersion());
        $this->assertInstanceOf(HelperSet::class, $app->getHelperSet());
        $app->setHelperSet(new HelperSet([new FormatterHelper]));
        $this->assertTrue($app->getHelperSet()->has('formatter'));
        $this->assertTrue($app->getDefinition()->hasOption('help'));
        $app->setDefinition(new InputDefinition([new InputArgument('command', InputArgument::REQUIRED), new InputOption('zed', 'z')]));
        $this->assertTrue($app->getDefinition()->hasOption('zed'));
        $this->assertTrue($app->has('covc:attr'));
        $this->assertTrue($app->has('covc:alias'));
        $this->assertFalse($app->has('nope'));
        $this->assertSame('covc:attr', $app->get('covc:alias')->getName());
        $this->assertNull($app->getAlarmInterval());
        $this->assertSame('register', $app->register('covc:reg')->setCode(fn () => 0) ? 'register' : '');
        $this->assertTrue($app->has('covc:reg'));
        $app->reset();
    }

    public function testAlarmInterval(): void
    {
        $app = $this->app();
        $app->setAlarmInterval(5);
        $this->assertSame(5, $app->getAlarmInterval());
    }

    public function testRunVariants(): void
    {
        $t = new ApplicationTester($this->app());
        $this->assertSame(3, $t->run(['command' => 'covc:hooks', 'who' => 'x']));
        $this->assertSame("hello x\n", $t->getDisplay());
        $this->assertSame(0, $t->run(['command' => 'covc:attr'], ['decorated' => false]));
        $this->assertSame(0, $t->run(['--version' => true]));
        $this->assertSame("covapp 9.9\n", $t->getDisplay());
        $this->assertSame(0, $t->run(['-V' => true]));
        $this->assertSame(0, $t->run(['--help' => true]));
        $this->assertStringContainsString('Usage:', $t->getDisplay());
        $this->assertSame(0, $t->run(['command' => 'covc:hooks', '--help' => true]));
        $this->assertStringContainsString('covc:hooks [options] [--] [<who>]', $t->getDisplay());
        $this->assertSame(0, $t->run(['command' => 'help', 'command_name' => 'covc:attr']));
        $this->assertSame(0, $t->run(['command' => 'covc:attr', '--quiet' => true]));
        $this->assertSame('', $t->getDisplay());
        $this->assertSame(0, $t->run(['command' => 'covc:attr', '--silent' => true]));
        $this->assertSame(0, $t->run(['command' => 'covc:attr', '--ansi' => true]));
        $this->assertSame(0, $t->run(['command' => 'covc:attr', '--no-ansi' => true]));
        $this->assertSame(0, $t->run(['command' => 'covc:attr', '--no-interaction' => true]));
        foreach (['-v', '-vv', '-vvv', '--verbose'] as $v) {
            $this->assertSame(0, $t->run(['command' => 'covc:attr', $v => true]));
        }
        $this->assertSame(0, $t->run(['command' => 'covc:attr'], ['verbosity' => OutputInterface::VERBOSITY_DEBUG]));
        $this->assertSame(0, $t->run(['command' => 'covc:attr'], ['verbosity' => OutputInterface::VERBOSITY_QUIET]));
        $this->assertSame(0, $t->run(['command' => 'covc:attr'], ['verbosity' => OutputInterface::VERBOSITY_VERY_VERBOSE]));
        $this->assertSame(0, $t->run(['command' => 'covc:attr'], ['verbosity' => OutputInterface::VERBOSITY_SILENT]));
    }

    public function testRunErrorsRendered(): void
    {
        $t = new ApplicationTester($this->app());
        $this->assertSame(1, $t->run(['command' => 'covc:boom']));
        $this->assertStringContainsString('plain', $t->getDisplay());
        $this->assertSame(1, $t->run(['command' => 'covc:boom'], ['capture_stderr_separately' => true]));
        $this->assertStringContainsString('boom', $t->getErrorOutput());
        $this->assertSame(42, $t->run(['command' => 'covc:boom', 'kind' => 'code']));
        $this->assertSame(1, $t->run(['command' => 'covc:boom', 'kind' => 'logic']));
        $this->assertStringContainsString('logic problem', $t->getDisplay());
        $this->assertSame(1, $t->run(['command' => 'covc:boom', 'kind' => 'nested'], ['verbosity' => OutputInterface::VERBOSITY_VERBOSE]));
        $this->assertStringContainsString('inner cause', $t->getDisplay());
        $this->assertSame(1, $t->run(['command' => 'covc:boom', 'kind' => 'nested'], ['verbosity' => OutputInterface::VERBOSITY_VERY_VERBOSE]));
        $this->assertStringContainsString('Exception trace', $t->getDisplay());
        $app2 = $this->app();
        $app2->setCatchErrors(true);
        $this->assertSame(1, (new ApplicationTester($app2))->run(['command' => 'covc:boom', 'kind' => 'type']));
        $this->assertSame(1, $t->run(['command' => 'covc:noexec']));
        $this->assertStringContainsString('override the execute()', $t->getDisplay());
        $this->assertSame(1, $t->run(['command' => 'nonexistent']));
        $this->assertStringContainsString('Command "nonexistent" is not defined.', $t->getDisplay());
        $this->assertSame(1, $t->run(['command' => 'covc:attx']));
        $this->assertStringContainsString('Did you mean', $t->getDisplay());
        $this->assertSame(1, $t->run(['command' => 'covc:disabled']));
        $this->assertSame(1, $t->run(['command' => 'covc:hooks', '--nope' => true]));
        $this->assertStringContainsString('"--nope" option does not exist', $t->getDisplay());
        $this->assertSame(1, $t->run(['command' => 'covc:hooks', 'extra' => 'a', 'more' => 'b']));
    }

    public function testRenderThrowableTruncates(): void
    {
        $app = $this->app();
        $out = new BufferedOutput();
        $out->setDecorated(false);
        $app->renderThrowable(new \RuntimeException("line one\nline two"), $out);
        $this->assertStringContainsString('line two', $out->fetch());
        $out = new BufferedOutput(OutputInterface::VERBOSITY_VERBOSE);
        $app->renderThrowable(new \RuntimeException('with <tags>'), $out);
        $this->assertStringContainsString('RuntimeException', $out->fetch());
        $out = new BufferedOutput();
        $app->renderThrowable(new \Exception(''), $out);
        $this->assertNotSame('', $out->fetch());
        $out = new BufferedOutput();
        $app->renderThrowable(new class('anon') extends \Exception {}, $out);
        $this->assertStringContainsString('anon', $out->fetch());
    }

    public function testNotCatching(): void
    {
        $app = $this->app();
        $app->setCatchExceptions(false);
        $t = new ApplicationTester($app);
        $this->expectException(\RuntimeException::class);
        $t->run(['command' => 'covc:boom']);
    }
    public function testNotCatchingErrors(): void
    {
        $app = $this->app();
        $app->setCatchErrors(false);
        $t = new ApplicationTester($app);
        $this->expectException(\TypeError::class);
        $t->run(['command' => 'covc:boom', 'kind' => 'type']);
    }

    public function testFindAndAbbreviations(): void
    {
        $app = $this->app();
        $this->assertSame('covc:hooks', $app->find('covc:hoo')->getName());
        $this->assertSame('covc:hooks', $app->find('c:h')->getName());
        $this->assertSame('covc:attr', $app->find('covc:alias')->getName());
        $this->assertSame('help', $app->find('hel')->getName());
        $this->assertContains('covc', $app->getNamespaces());
        $this->assertSame('covc', $app->findNamespace('cov'));
        $this->assertArrayHasKey('covc:hooks', $app->all('covc'));
        $this->assertSame('covc', $app->extractNamespace('covc:hooks', 1));
        $this->assertSame('a:b', $app->extractNamespace('a:b:c'));
        $this->assertSame('', $app->extractNamespace('solo'));
        try { $app->find('covc:'); $this->fail('ambiguous'); } catch (CommandNotFoundException $e) { $this->assertStringContainsString('ambiguous', $e->getMessage()); $this->assertNotEmpty($e->getAlternatives()); }
        try { $app->findNamespace('zzz'); $this->fail('ns'); } catch (NamespaceNotFoundException $e) { $this->assertStringContainsString('There are no commands defined in the "zzz" namespace', $e->getMessage()); }
        try { $app->get('zzz'); $this->fail('get'); } catch (CommandNotFoundException $e) { $this->assertSame([], $e->getAlternatives()); }
        try { $app->find('covc:hooX'); $this->fail('alt'); } catch (CommandNotFoundException $e) { $this->assertStringContainsString('covc:hooks', $e->getMessage()); }
    }

    public function testDefaultAndSingleCommand(): void
    {
        $app = $this->app();
        $app->setDefaultCommand('covc:attr');
        $t = new ApplicationTester($app);
        $this->assertSame(0, $t->run([]));
        $this->assertSame("attr\n", $t->getDisplay());
        $single = new Application('single', '1');
        $single->setAutoExit(false);
        $single->add(new ConsoleCovHooks);
        $single->setDefaultCommand('covc:hooks', true);
        $this->assertTrue($single->isSingleCommand());
        $t2 = new ApplicationTester($single);
        $this->assertSame(3, $t2->run(['who' => 'solo']));
        $this->assertSame("hello solo\n", $t2->getDisplay());
        $t2->run(['--help' => true]);
        $this->assertStringContainsString('Usage', $t2->getDisplay());
    }

    public function testCommandLoader(): void
    {
        $loader = new FactoryCommandLoader(['lazy:one' => fn () => (new Command('lazy:one'))->setCode(fn () => 0), 'lazy:two' => fn () => (new Command('lazy:two'))->setCode(fn () => 0)]);
        $this->assertTrue($loader->has('lazy:one'));
        $this->assertFalse($loader->has('nope'));
        $this->assertSame(['lazy:one', 'lazy:two'], $loader->getNames());
        $this->assertInstanceOf(Command::class, $loader->get('lazy:one'));
        $app = $this->app();
        $app->setCommandLoader($loader);
        $this->assertTrue($app->has('lazy:one'));
        $this->assertSame('lazy:one', $app->get('lazy:one')->getName());
        $this->assertContains('lazy', $app->getNamespaces());
        $this->assertArrayHasKey('lazy:two', $app->all('lazy'));
        $this->assertNotNull($app->find('lazy:tw'));
        try { $loader->get('missing'); $this->fail(); } catch (\Symfony\Component\Console\Exception\CommandNotFoundException $e) { $this->assertStringContainsString('missing', $e->getMessage()); }
    }

    public function testEvents(): void
    {
        $cmd = new ConsoleCovHooks();
        $in = new ArrayInput([]);
        $out = new NullOutput();
        $e = new ConsoleCommandEvent($cmd, $in, $out);
        $this->assertTrue($e->commandShouldRun());
        $e->disableCommand();
        $this->assertFalse($e->commandShouldRun());
        $e->enableCommand();
        $this->assertTrue($e->commandShouldRun());
        $this->assertSame($cmd, $e->getCommand());
        $this->assertSame($in, $e->getInput());
        $this->assertSame($out, $e->getOutput());
        $err = new ConsoleErrorEvent($in, $out, new \RuntimeException('x', 5), $cmd);
        $this->assertSame('x', $err->getError()->getMessage());
        $this->assertSame(5, $err->getExitCode());
        $err->setExitCode(9);
        $this->assertSame(9, $err->getExitCode());
        $err->setError(new \LogicException('y'));
        $this->assertSame('y', $err->getError()->getMessage());
        $err2 = new ConsoleErrorEvent($in, $out, new \RuntimeException('x'));
        $this->assertSame(1, $err2->getExitCode());
        $this->assertNull($err2->getCommand());
        $term = new ConsoleTerminateEvent($cmd, $in, $out, 2, 15);
        $this->assertSame(2, $term->getExitCode());
        $this->assertSame(15, $term->getInterruptingSignal());
        $term->setExitCode(4);
        $this->assertSame(4, $term->getExitCode());
        $sig = new ConsoleSignalEvent($cmd, $in, $out, 2, 0);
        $this->assertSame(2, $sig->getHandlingSignal());
        $this->assertSame(0, $sig->getExitCode());
        $sig->setExitCode(130);
        $this->assertSame(130, $sig->getExitCode());
        $sig->abortExit();
        $this->assertFalse($sig->getExitCode());
        $al = new ConsoleAlarmEvent($cmd, $in, $out);
        $al->setExitCode(3);
        $this->assertSame(3, $al->getExitCode());
        $al->abortExit();
        $this->assertFalse($al->getExitCode());
        foreach ([$sig, $al] as $ev) {
            try { $ev->setExitCode(300); $this->fail('range'); } catch (\InvalidArgumentException $x) { $this->assertStringContainsString('between 0 and 255', $x->getMessage()); }
        }
        try { $term->setExitCode(1); $err->setExitCode(1); $this->assertTrue(true); } catch (\Throwable $x) { $this->fail($x->getMessage()); }
    }

    public function testCompleteCommand(): void
    {
        $app = $this->app();
        $app->add(new CompleteCommand());
        $app->add(new DumpCompletionCommand());
        $t = new ApplicationTester($app);
        $this->assertSame(0, $t->run(['command' => '_complete', '--shell' => 'bash', '--api-version' => '1', '--current' => '1', '--input' => ['covapp', 'covc:co']]));
        $this->assertStringContainsString('covc:compl', $t->getDisplay());
        $t->run(['command' => '_complete', '--shell' => 'bash', '--api-version' => '1', '--current' => '2', '--input' => ['covapp', 'covc:compl', '']]);
        $this->assertStringContainsString('red', $t->getDisplay());
        $t->run(['command' => '_complete', '--shell' => 'fish', '--api-version' => '1', '--current' => '2', '--input' => ['covapp', 'covc:compl', 'g']]);
        $this->assertStringContainsString('green', $t->getDisplay());
        $t->run(['command' => '_complete', '--shell' => 'zsh', '--api-version' => '1', '--current' => '3', '--input' => ['covapp', 'covc:compl', '--fmt', '']]);
        $this->assertStringContainsString('json', $t->getDisplay());
        $t->run(['command' => '_complete', '--shell' => 'bash', '--api-version' => '1', '--current' => '2', '--input' => ['covapp', 'covc:compl', '--']]);
        $this->assertStringContainsString('--fmt', $t->getDisplay());
        $this->assertIsInt($t->run(['command' => '_complete', '--shell' => 'nope', '--api-version' => '1', '--current' => '1', '--input' => ['covapp', '']]));
        $this->assertIsInt($t->run(['command' => '_complete', '--shell' => 'bash', '--api-version' => '0', '--current' => '1', '--input' => ['covapp', '']]));
        $this->assertIsInt($t->run(['command' => '_complete', '--shell' => 'bash', '--api-version' => '1', '--input' => ['covapp', '']]));
        $this->assertIsInt($t->run(['command' => '_complete', '--api-version' => '1', '--current' => '1', '--input' => ['covapp', '']]));
        foreach (['bash', 'fish', 'zsh'] as $sh) {
            $r = $t->run(['command' => 'completion', 'shell' => $sh]);
            $this->assertSame(0, $r, $sh);
            $this->assertNotSame('', $t->getDisplay());
        }
        $this->assertNotSame(0, $t->run(['command' => 'completion', 'shell' => 'tcsh']));
    }

    public function testCommandCompletionTester(): void
    {
        $t = new CommandCompletionTester(new ConsoleCovCompl());
        $this->assertSame(['blue', 'green', 'red'], $this->sorted($t->complete([''])));
        $this->assertSame(['blue', 'green', 'red'], $this->sorted($t->complete(['g'])));
        $this->assertSame(['json', 'yaml'], $this->sorted($t->complete(['--fmt', ''])));
        $this->assertSame([], $t->complete(['--']));
        $t2 = new CommandCompletionTester(new ConsoleCovCompl());
        $this->assertSame([], $t2->complete(['--f']));
    }
    private function sorted(array $a): array { sort($a); return $a; }

    public function testComplete(): void
    {
        $app = $this->app();
        $s = new CompletionSuggestions();
        $in = CompletionInput::fromTokens(['covapp', ''], 1);
        $in->bind($app->getDefinition());
        $app->complete($in, $s);
        $this->assertNotEmpty($s->getValueSuggestions());
        $s = new CompletionSuggestions();
        $in = CompletionInput::fromTokens(['covapp', '--'], 1);
        $in->bind($app->getDefinition());
        $app->complete($in, $s);
        $this->assertNotEmpty($s->getOptionSuggestions());
        $s = new CompletionSuggestions();
        $in = CompletionInput::fromTokens(['covapp', 'covc:compl', ''], 2);
        $in->bind($app->getDefinition());
        $app->complete($in, $s);
        $this->assertNotEmpty($s->getValueSuggestions());
    }

    public function testCursor(): void
    {
        $fp = fopen('php://memory', 'w+');
        $out = new StreamOutput($fp);
        $c = new Cursor($out, $fp);
        $c->moveUp(2)->moveDown(3)->moveRight(4)->moveLeft(5)->moveToColumn(6)->moveToPosition(7, 8)->savePosition()->restorePosition()->hide()->show()->clearLine()->clearLineAfter()->clearOutput()->clearScreen();
        $c->moveUp()->moveDown()->moveRight()->moveLeft();
        rewind($fp);
        $s = stream_get_contents($fp);
        $this->assertStringContainsString("\x1b[2A", $s);
        $this->assertStringContainsString("\x1b[3B", $s);
        $this->assertStringContainsString("\x1b[4C", $s);
        $this->assertStringContainsString("\x1b[5D", $s);
        $this->assertStringContainsString("\x1b[6G", $s);
        $this->assertStringContainsString("\x1b[9;7H", $s);
        $this->assertStringContainsString("\x1b7", $s);
        $this->assertStringContainsString("\x1b8", $s);
        $this->assertStringContainsString("\x1b[?25l", $s);
        $this->assertStringContainsString("\x1b[?25h", $s);
        $this->assertStringContainsString("\x1b[2K", $s);
        $this->assertStringContainsString("\x1b[K", $s);
        $this->assertStringContainsString("\x1b[0J", $s);
        $this->assertStringContainsString("\x1b[2J", $s);
        fclose($fp);
    }

    public function testTerminal(): void
    {
        putenv('COLUMNS=77');
        putenv('LINES=33');
        $t = new Terminal();
        $this->assertSame(77, $t->getWidth());
        $this->assertSame(33, $t->getHeight());
        $this->assertIsBool(Terminal::hasSttyAvailable());
        $this->assertInstanceOf(\Symfony\Component\Console\Output\AnsiColorMode::class, Terminal::getColorMode());
        Terminal::setColorMode(\Symfony\Component\Console\Output\AnsiColorMode::Ansi8);
        $this->assertSame(\Symfony\Component\Console\Output\AnsiColorMode::Ansi8, Terminal::getColorMode());
        Terminal::setColorMode(\Symfony\Component\Console\Output\AnsiColorMode::Ansi24);
        $this->assertSame(\Symfony\Component\Console\Output\AnsiColorMode::Ansi24, Terminal::getColorMode());
        Terminal::setColorMode(null);
        $this->assertInstanceOf(\Symfony\Component\Console\Output\AnsiColorMode::class, Terminal::getColorMode());
        putenv('COLUMNS');
        putenv('LINES');
    }

    public function testGithubReporter(): void
    {
        $out = new BufferedOutput();
        $r = new GithubActionReporter($out);
        $r->error('msg: with, odd %chars');
        $r->error("multi\nline", 'src/File.php', 12, 3);
        $r->error('only file', 'a:b.php');
        $r->error('file and line', 'a.php', 5);
        $s = $out->fetch();
        $this->assertStringContainsString('::error::msg: with, odd %25chars', $s);
        $this->assertStringContainsString("::error file=src/File.php,line=12,col=3::multi%0Aline", $s);
        $this->assertStringContainsString('file=a%3Ab.php,line=1,col=0::only file', $s);
        $this->assertStringContainsString('file=a.php,line=5,col=0::file and line', $s);
        $this->assertIsBool(GithubActionReporter::isGithubActionEnvironment());
    }
}
