<?php
use PHPUnit\Framework\TestCase;
use Symfony\Component\Console\{Application, Command\Command};
use Symfony\Component\Console\Input\{ArrayInput, InputArgument, InputOption, InputInterface};
use Symfony\Component\Console\Output\{BufferedOutput, OutputInterface};
use Symfony\Component\Console\Helper\Table;
use Symfony\Component\Console\Formatter\OutputFormatter;
use Symfony\Component\Console\Tester\CommandTester;

final class GreetCommand extends Command {
    protected function configure(): void { $this->setName('greet')->setDescription('Greets')->addArgument('name', InputArgument::OPTIONAL, 'Who', 'world')->addOption('shout', 's', InputOption::VALUE_NONE); }
    protected function execute(InputInterface $in, OutputInterface $out): int { $m = 'Hello ' . $in->getArgument('name'); $out->writeln($in->getOption('shout') ? strtoupper($m) : $m); return Command::SUCCESS; }
}
final class ConsoleTest extends TestCase
{
    public function testCommandTester(): void
    {
        $t = new CommandTester(new GreetCommand);
        $this->assertSame(0, $t->execute(['name' => 'bob']));
        $this->assertSame("Hello bob\n", $t->getDisplay());
        $t->execute(['--shout' => true]);
        $this->assertSame("HELLO WORLD\n", $t->getDisplay());
    }
    public function testApplicationRun(): void
    {
        $app = new Application('t', '1.0'); $app->setAutoExit(false); $app->add(new GreetCommand);
        $out = new BufferedOutput;
        $code = $app->run(new ArrayInput(['command' => 'greet', 'name' => 'x']), $out);
        $this->assertSame(0, $code);
        $this->assertSame("Hello x\n", $out->fetch());
        $this->assertSame('greet', $app->find('gre')->getName());
    }
    public function testTable(): void
    {
        $out = new BufferedOutput;
        (new Table($out))->setHeaders(['id', 'name'])->setRows([[1, 'a'], [2, 'bb']])->render();
        $this->assertStringContainsString('| id | name |', $out->fetch());
    }
    public function testFormatter(): void
    {
        $this->assertSame('plain \\<b\\>', OutputFormatter::escape('plain <b>'));
        $f = new OutputFormatter(false);
        $this->assertSame('text', $f->format('<info>text</info>'));
    }
    public function testInputParsing(): void
    {
        $def = (new GreetCommand)->getDefinition();
        $in = new ArrayInput(['name' => 'z', '-s' => true], $def);
        $this->assertSame('z', $in->getArgument('name'));
        $this->assertTrue($in->getOption('shout'));
    }
}
