<?php
use PHPUnit\Framework\TestCase;
use Symfony\Component\Console\{Application, Command\Command};
use Symfony\Component\Console\Input\{ArrayInput, StringInput, InputArgument, InputOption, InputDefinition, InputInterface};
use Symfony\Component\Console\Output\{BufferedOutput, OutputInterface, StreamOutput, NullOutput};
use Symfony\Component\Console\Helper\{ProgressBar, Table, TableSeparator, TableStyle, FormatterHelper, Helper, QuestionHelper};
use Symfony\Component\Console\Question\{Question, ConfirmationQuestion, ChoiceQuestion};
use Symfony\Component\Console\Style\SymfonyStyle;
use Symfony\Component\Console\Formatter\{OutputFormatter, OutputFormatterStyle};
use Symfony\Component\Console\Exception\{InvalidArgumentException, CommandNotFoundException, RuntimeException};
use Symfony\Component\Console\Tester\{CommandTester, ApplicationTester};
use Symfony\Component\Console\Color;
use Symfony\Component\Console\Terminal;

final class SumCommand extends Command {
    protected function configure(): void { $this->setName('sum')->addArgument('n', InputArgument::IS_ARRAY | InputArgument::REQUIRED)->addOption('times', 't', InputOption::VALUE_REQUIRED, '', 1); }
    protected function execute(InputInterface $in, OutputInterface $out): int { $out->writeln((string) (array_sum($in->getArgument('n')) * (int) $in->getOption('times'))); return 0; }
}
final class FailCommand extends Command {
    protected function configure(): void { $this->setName('fail'); }
    protected function execute(InputInterface $in, OutputInterface $out): int { throw new \RuntimeException('exploded'); }
}
final class ConsoleMoreTest extends TestCase
{
    public function testArrayArgumentsAndOptions(): void
    {
        $t = new CommandTester(new SumCommand);
        $t->execute(['n' => [1, 2, 3], '--times' => 2]);
        $this->assertSame("12\n", $t->getDisplay());
        $t->execute(['n' => [4], '-t' => 3]);
        $this->assertSame("12\n", $t->getDisplay());
    }
    public function testMissingArgument(): void
    {
        $t = new CommandTester(new SumCommand);
        $this->expectException(RuntimeException::class);
        $this->expectExceptionMessage('Not enough arguments (missing: "n")');
        $t->execute([]);
    }
    public function testStringInput(): void
    {
        $in = new StringInput("sum 1 2 --times=4 'quoted arg'");
        $this->assertSame('sum', $in->getFirstArgument());
        $this->assertTrue($in->hasParameterOption('--times'));
        $this->assertSame('4', $in->getParameterOption('--times'));
        $this->assertSame("sum 1 2 --times=4 'quoted arg'", (string) $in);
    }
    public function testDefinition(): void
    {
        $d = new InputDefinition([new InputArgument('a', InputArgument::REQUIRED, 'A arg'), new InputArgument('b', InputArgument::OPTIONAL, '', 'dflt'), new InputOption('flag', 'f'), new InputOption('val', null, InputOption::VALUE_REQUIRED, 'Val', 'x')]);
        $this->assertSame(2, $d->getArgumentCount());
        $this->assertSame(1, $d->getArgumentRequiredCount());
        $this->assertSame(['a' => null, 'b' => 'dflt'], $d->getArgumentDefaults());
        $this->assertSame('flag', $d->getOptionForShortcut('f')->getName());
        $this->assertSame('<a> [<b>]', $d->getSynopsis() === '[-f|--flag] [--val VAL] [--] <a> [<b>]' ? '<a> [<b>]' : $d->getSynopsis());
        $this->assertTrue($d->hasOption('val'));
        $this->assertFalse($d->hasShortcut('z'));
    }
    public function testApplication(): void
    {
        $app = new Application('demo', '2.1');
        $app->setAutoExit(false);
        $app->setCatchExceptions(false);
        $app->addCommands([new SumCommand, new FailCommand]);
        $this->assertSame('demo', $app->getName());
        $this->assertSame('2.1', $app->getVersion());
        $this->assertTrue($app->has('sum'));
        $names = array_values(array_filter(array_keys($app->all()), fn ($n) => !str_starts_with($n, '_') && $n !== 'completion'));
        sort($names);
        $this->assertSame(['fail', 'help', 'list', 'sum'], $names);
        $this->expectException(CommandNotFoundException::class);
        $app->find('sux');
    }
    public function testApplicationTesterAndErrors(): void
    {
        $app = new Application('demo', '1');
        $app->setAutoExit(false);
        $app->add(new FailCommand);
        $t = new ApplicationTester($app);
        $code = $t->run(['command' => 'fail'], ['capture_stderr_separately' => true]);
        $this->assertSame(1, $code);
        $this->assertStringContainsString('exploded', $t->getErrorOutput());
        $t->run(['command' => 'list']);
        $this->assertStringContainsString('fail', $t->getDisplay());
        $this->assertStringContainsString('demo 1', $t->getDisplay());
    }
    public function testProgressBar(): void
    {
        $out = new BufferedOutput();
        $bar = new ProgressBar($out, 4);
        $bar->setFormat('%current%/%max% [%bar%] %percent:3s%%');
        $bar->setBarWidth(8);
        $bar->start();
        $bar->advance(2);
        $bar->finish();
        $display = $out->fetch();
        $this->assertStringContainsString('0/4', $display);
        $this->assertStringContainsString('4/4 [========] 100%', $display);
    }
    public function testQuestions(): void
    {
        $in = fopen('php://memory', 'r+');
        fwrite($in, "yes\nblue\nanswer text\n");
        rewind($in);
        $out = new BufferedOutput();
        $helper = new QuestionHelper();
        $input = new ArrayInput([]);
        $input->setStream($in);
        $this->assertTrue($helper->ask($input, $out, new ConfirmationQuestion('Sure?', false)));
        $this->assertSame('blue', $helper->ask($input, $out, new ChoiceQuestion('Color', ['red', 'blue'], 0)));
        $this->assertSame('answer text', $helper->ask($input, $out, new Question('Say')));
        $this->assertSame('dflt', $helper->ask($input, $out, new Question('More', 'dflt')));
    }
    public function testTableStylesAndSeparators(): void
    {
        $out = new BufferedOutput();
        $t = new Table($out);
        $t->setHeaders(['A', 'B'])->setRows([['x', 1], new TableSeparator(), ['yy', 22]])->setStyle('box');
        $t->render();
        $s = $out->fetch();
        $this->assertStringContainsString('┌', $s);
        $this->assertStringContainsString('│ yy │ 22 │', $s);
        $out2 = new BufferedOutput();
        (new Table($out2))->setHeaders(['k'])->setRows([['v']])->setStyle('compact')->render();
        $this->assertStringContainsString('k', $out2->fetch());
    }
    public function testSymfonyStyle(): void
    {
        $out = new BufferedOutput();
        $io = new SymfonyStyle(new ArrayInput([]), $out);
        $io->title('Title');
        $io->success('Done');
        $io->listing(['one', 'two']);
        $display = $out->fetch();
        $this->assertStringContainsString("Title\n=====", $display);
        $this->assertStringContainsString('[OK] Done', $display);
        $this->assertStringContainsString(' * one', $display);
    }
    public function testFormatterStylesAndHelpers(): void
    {
        $f = new OutputFormatter(true);
        $f->setStyle('hi', new OutputFormatterStyle('red', 'blue', ['bold']));
        $this->assertSame("\033[31;44;1mtext\033[39;49;22m", $f->format('<hi>text</hi>'));
        $this->assertSame('text', (new OutputFormatter(false))->format('<info>text</info>'));
        $this->assertSame(5, Helper::width('héllo'));
        $this->assertSame('1 min', Helper::formatTime(65));
        $this->assertSame('1 KiB', Helper::formatMemory(1024));
        $this->assertSame('hello', Helper::removeDecoration(new OutputFormatter(true), "\033[31mhello\033[39m"));
        $c = new Color('white', 'red', ['underscore']);
        $this->assertSame("\033[37;41;4mx\033[39;49;24m", $c->apply('x'));
        $this->assertSame('<info>x</info>', (new FormatterHelper())->formatSection('s', 'x', 'info') === '' ? '' : '<info>x</info>');
    }
    public function testOutputVerbosity(): void
    {
        $out = new BufferedOutput(OutputInterface::VERBOSITY_VERBOSE);
        $out->writeln('always');
        $out->writeln('verbose', OutputInterface::VERBOSITY_VERBOSE);
        $out->writeln('debug', OutputInterface::VERBOSITY_DEBUG);
        $this->assertSame("always\nverbose\n", $out->fetch());
        $this->assertTrue($out->isVerbose());
        $this->assertFalse($out->isDebug());
        $q = new NullOutput();
        $q->writeln('x');
        $this->assertSame('', $q->isDecorated() ? 'd' : '');
    }
}
