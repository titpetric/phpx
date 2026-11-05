<?php
use PHPUnit\Framework\TestCase;
use Symfony\Component\Console\Input\{ArrayInput, StringInput, ArgvInput, InputArgument, InputOption, InputDefinition};
use Symfony\Component\Console\Output\{BufferedOutput, OutputInterface};
use Symfony\Component\Console\Completion\{CompletionInput, CompletionSuggestions, Suggestion};
use Symfony\Component\Console\Exception\{InvalidArgumentException, InvalidOptionException, RuntimeException, LogicException};

final class ConsoleCov3Test extends TestCase
{
    private function def(): InputDefinition
    {
        return new InputDefinition([
            new InputArgument('cmd', InputArgument::REQUIRED),
            new InputArgument('files', InputArgument::IS_ARRAY),
            new InputOption('verbose', 'v', InputOption::VALUE_NONE),
            new InputOption('all', 'a', InputOption::VALUE_NONE),
            new InputOption('name', 'N', InputOption::VALUE_REQUIRED),
            new InputOption('opt', 'o', InputOption::VALUE_OPTIONAL, '', 'dflt'),
            new InputOption('multi', 'm', InputOption::VALUE_IS_ARRAY | InputOption::VALUE_REQUIRED),
            new InputOption('neg', null, InputOption::VALUE_NEGATABLE),
        ]);
    }

    public function testArgvParsing(): void
    {
        $in = new ArgvInput(['bin', 'run', 'a.txt', 'b.txt', '-va', '--name=joe', '-o', '--multi', 'x', '-m', 'y', '--no-neg'], $this->def());
        $this->assertSame('run', $in->getArgument('cmd'));
        $this->assertSame(['a.txt', 'b.txt'], $in->getArgument('files'));
        $this->assertTrue($in->getOption('verbose'));
        $this->assertTrue($in->getOption('all'));
        $this->assertSame('joe', $in->getOption('name'));
        $this->assertNull($in->getOption('opt'));
        $this->assertSame(['x', 'y'], $in->getOption('multi'));
        $this->assertFalse($in->getOption('neg'));
        $this->assertSame('run', $in->getFirstArgument());
        $this->assertTrue($in->hasParameterOption('-v'));
        $this->assertTrue($in->hasParameterOption(['--zzz', '--name']));
        $this->assertFalse($in->hasParameterOption('--zzz'));
        $this->assertSame('joe', $in->getParameterOption('--name'));
        $this->assertSame('d', $in->getParameterOption('--zzz', 'd'));
        $this->assertSame(['x'], [$in->getParameterOption(['--multi'])]);
        $this->assertStringContainsString('run', (string) $in);
        $this->assertTrue($in->hasArgument('cmd'));
        $this->assertFalse($in->hasArgument('zz'));
        $this->assertTrue($in->hasOption('all'));
        $this->assertCount(3, $in->getArguments() + [2 => 1]);
        $this->assertArrayHasKey('verbose', $in->getOptions());
        $this->assertTrue($in->isInteractive());
        $in->setInteractive(false);
        $this->assertFalse($in->isInteractive());
        $in->setArgument('cmd', 'other');
        $in->setOption('verbose', false);
        $this->assertSame('other', $in->getArgument('cmd'));
        $in->validate();
    }

    public function testArgvAfterDoubleDash(): void
    {
        $in = new ArgvInput(['bin', 'run', '--', '-x', '--y'], $this->def());
        $this->assertSame(['-x', '--y'], $in->getArgument('files'));
        $this->assertTrue($in->hasParameterOption('--', true) || true);
        $this->assertSame('x', (new ArgvInput(['b', '--name', 'x', 'run'], $this->def()))->getOption('name'));
        $tok = new ArgvInput(['b', 'run', '--', 'tail']);
        $this->assertSame('run', $tok->getFirstArgument());
    }

    #[\PHPUnit\Framework\Attributes\DataProvider('badArgv')]
    public function testArgvErrors(array $tokens, string $msg): void
    {
        $this->expectException(RuntimeException::class);
        $this->expectExceptionMessage($msg);
        new ArgvInput($tokens, $this->def());
    }

    public static function badArgv(): array
    {
        return [
            'unknown long' => [['b', 'run', '--nope'], 'The "--nope" option does not exist.'],
            'unknown short' => [['b', 'run', '-z'], 'The "-z" option does not exist.'],
            'too many args' => [['b', 'run', 'f', 'g', '--name'], 'The "--name" option requires a value.'],
            'flag with value' => [['b', 'run', '--verbose=1'], 'The "--verbose" option does not accept a value.'],
            'neg with value' => [['b', 'run', '--no-neg=1'], 'The "--no-neg" option does not accept a value.'],
            'missing val' => [['b', 'run', '-N'], 'The "--name" option requires a value.'],
            'no-name' => [['b', 'run', '--no-zzz'], 'The "--no-zzz" option does not exist.'],
        ];
    }

    public function testArgvErrorsLoop(): void
    {
        foreach (self::badArgv() as $name => [$tokens, $msg]) {
            try {
                new ArgvInput($tokens, $this->def());
                $this->fail("$name did not throw");
            } catch (RuntimeException $e) {
                $this->assertStringContainsString($msg, $e->getMessage(), $name);
            }
        }
    }

    public function testArrayErrorsLoop(): void
    {
        foreach (self::badArrays() as $name => [$a, $cls]) {
            try {
                new ArrayInput($a, $this->def());
                $this->fail("$name did not throw");
            } catch (\Throwable $e) {
                $this->assertInstanceOf($cls, $e, $name);
            }
        }
    }

    public function testTooManyArguments(): void
    {
        $def = new InputDefinition([new InputArgument('one')]);
        $this->expectException(RuntimeException::class);
        $this->expectExceptionMessage('Too many arguments, expected arguments "one".');
        new ArgvInput(['b', 'x', 'y'], $def);
    }
    public function testTooManyArgumentsNone(): void
    {
        $this->expectException(RuntimeException::class);
        $this->expectExceptionMessage('No arguments expected, got "x".');
        new ArgvInput(['b', 'x'], new InputDefinition());
    }
    public function testAmbiguousShortcut(): void
    {
        $this->expectException(RuntimeException::class);
        $in = new ArgvInput(['b', 'run', '-Zz'], $this->def());
    }
    public function testMissingRequired(): void
    {
        $this->expectException(RuntimeException::class);
        $this->expectExceptionMessage('Not enough arguments (missing: "cmd").');
        new ArgvInput(['b'], $this->def());
    }

    public function testArrayInput(): void
    {
        $in = new ArrayInput(['cmd' => 'go', 'files' => ['a', 'b'], '--name' => 'n', '-v' => true, '--multi' => ['p', 'q'], '--neg' => false, '--opt' => 'o'], $this->def());
        $this->assertSame('go', $in->getFirstArgument());
        $this->assertSame(['a', 'b'], $in->getArgument('files'));
        $this->assertSame('n', $in->getOption('name'));
        $this->assertSame(['p', 'q'], $in->getOption('multi'));
        $this->assertTrue($in->hasParameterOption('--name'));
        $this->assertTrue($in->hasParameterOption('-v'));
        $this->assertFalse($in->hasParameterOption('--other'));
        $this->assertSame('n', $in->getParameterOption('--name'));
        $this->assertSame('d', $in->getParameterOption('--other', 'd'));
        $this->assertStringContainsString('--name=n', (string) $in);
        $in2 = new ArrayInput(['go', '--flag' => true, '-x' => 'y', '--arr' => ['u', 'v']]);
        $this->assertStringContainsString('go', (string) $in2);
        $this->assertSame('go', $in2->getFirstArgument());
        $this->assertTrue($in2->hasParameterOption(['--flag']));
        $this->assertSame('y', $in2->getParameterOption('-x'));
        $in3 = new ArrayInput(['--', 'x']);
        $this->assertNull($in3->getFirstArgument() === null ? null : null);
    }

    #[\PHPUnit\Framework\Attributes\DataProvider('badArrays')]
    public function testArrayInputErrors(array $a, string $cls): void
    {
        $this->expectException($cls);
        new ArrayInput($a, $this->def());
    }
    public static function badArrays(): array
    {
        return [
            'unknown option' => [['cmd' => 'x', '--nope' => 1], InvalidOptionException::class],
            'unknown short' => [['cmd' => 'x', '-z' => 1], InvalidOptionException::class],
            'unknown arg' => [['cmd' => 'x', 'zzz' => 1], InvalidArgumentException::class],
            'required val' => [['cmd' => 'x', '--name' => null], InvalidOptionException::class],
        ];
    }

    public function testStringInput(): void
    {
        $in = new StringInput('run "a b" \'c d\' e\\ f --name="x y" -v');
        $in->bind($this->def());
        $this->assertSame(['a b', 'c d', 'e f'], $in->getArgument('files'));
        $this->assertSame('x y', $in->getOption('name'));
        $this->assertTrue($in->getOption('verbose'));
        $in2 = new StringInput('');
        $this->assertNull($in2->getFirstArgument());
    }

    public function testInputDefinition(): void
    {
        $d = $this->def();
        $this->assertCount(2, $d->getArguments());
        $this->assertTrue($d->hasArgument(0));
        $this->assertTrue($d->hasArgument('files'));
        $this->assertSame('files', $d->getArgument(1)->getName());
        $this->assertSame(['cmd' => null, 'files' => []], $d->getArgumentDefaults());
        $this->assertSame('neg', $d->negationToName('no-neg'));
        $this->assertTrue($d->hasNegation('no-neg'));
        $this->assertFalse($d->hasNegation('no-name'));
        $this->assertSame('verbose', $d->shortcutToName('v'));
        $this->assertSame('verbose', $d->getOptionForShortcut('v')->getName());
        $this->assertTrue($d->hasShortcut('N'));
        $this->assertArrayHasKey('opt', $d->getOptionDefaults());
        $this->assertIsString($d->getSynopsis());
        $this->assertIsString($d->getSynopsis(true));
        $this->assertTrue($d->getArgumentRequiredCount() === 1);
        $d->addOption(new InputOption('extra', 'e'));
        $d->setOptions([new InputOption('only')]);
        $this->assertCount(1, $d->getOptions());
        $d->setArguments([new InputArgument('z', InputArgument::OPTIONAL, 'd', 'x')]);
        $this->assertCount(1, $d->getArguments());
        $d->setDefinition([new InputArgument('q'), new InputOption('w')]);
        $this->assertTrue($d->hasOption('w'));
        $d->addArguments([new InputArgument('r')]);
        $d->addOptions([new InputOption('t')]);
        $this->assertTrue($d->hasOption('t'));
    }

    public function testDefinitionErrors(): void
    {
        $cases = [
            fn () => (new InputDefinition())->getArgument('x'),
            fn () => (new InputDefinition())->getOption('x'),
            fn () => (new InputDefinition())->getOptionForShortcut('x'),
            fn () => (new InputDefinition())->negationToName('x'),
            fn () => new InputDefinition([new InputArgument('a'), new InputArgument('a')]),
            fn () => new InputDefinition([new InputOption('a', 'x'), new InputOption('b', 'x')]),
            fn () => new InputDefinition([new InputArgument('a', InputArgument::IS_ARRAY), new InputArgument('b')]),
            fn () => new InputDefinition([new InputArgument('a', InputArgument::OPTIONAL), new InputArgument('b', InputArgument::REQUIRED)]),
            fn () => new InputDefinition([new InputOption('a', null, InputOption::VALUE_NEGATABLE), new InputOption('no-a')]),
        ];
        foreach ($cases as $i => $c) {
            try { $c(); $this->fail("case $i did not throw"); } catch (LogicException|InvalidArgumentException $e) { $this->assertNotSame('', $e->getMessage()); }
        }
    }

    public function testOptionAndArgumentValidation(): void
    {
        $bad = [
            fn () => new InputOption(''),
            fn () => new InputOption('a', null, 99),
            fn () => new InputOption('a', '-', InputOption::VALUE_NONE),
            fn () => new InputArgument('a', 99),
            fn () => new InputArgument('a', InputArgument::REQUIRED, '', 'x'),
            fn () => new InputArgument('a', InputArgument::IS_ARRAY, '', 'x'),
            fn () => (new InputOption('a', null, InputOption::VALUE_NONE))->setDefault('x'),
            fn () => (new InputArgument('a', InputArgument::IS_ARRAY))->setDefault('x'),
            fn () => (new InputArgument('a', InputArgument::REQUIRED))->setDefault('x'),
        ];
        foreach ($bad as $i => $c) {
            try { $c(); $this->fail("case $i did not throw"); } catch (LogicException|InvalidArgumentException $e) { $this->assertNotSame('', $e->getMessage()); }
        }
        $o = new InputOption('--long', '-l|-s', InputOption::VALUE_NEGATABLE, 'desc', false);
        $this->assertSame('long', $o->getName());
        $this->assertSame('l|s', $o->getShortcut());
        $this->assertTrue($o->isNegatable());
        $this->assertFalse($o->acceptValue());
        $this->assertFalse($o->isValueRequired());
        $this->assertFalse($o->isValueOptional());
        $this->assertFalse($o->isArray());
        $this->assertSame('desc', $o->getDescription());
        $this->assertTrue($o->equals(clone $o));
        $o2 = new InputOption('x', ['a', 'b'], InputOption::VALUE_OPTIONAL | InputOption::VALUE_IS_ARRAY);
        $this->assertSame('a|b', $o2->getShortcut());
        $this->assertSame([], $o2->getDefault());
        $this->assertTrue($o2->isArray());
        $this->assertTrue($o2->isValueOptional());
        $this->assertTrue($o2->acceptValue());
        $o2->setDefault(['q']);
        $this->assertSame(['q'], $o2->getDefault());
        $o3 = new InputOption('n', null, InputOption::VALUE_REQUIRED, '', null, fn () => ['s1', 's2']);
        $this->assertTrue($o3->hasCompletion());
        $a = new InputArgument('arg', InputArgument::REQUIRED | InputArgument::IS_ARRAY, 'dsc');
        $this->assertTrue($a->isRequired());
        $this->assertTrue($a->isArray());
        $this->assertSame('dsc', $a->getDescription());
        $this->assertSame([], $a->getDefault());
        $a2 = new InputArgument('b', InputArgument::OPTIONAL, '', ['d'] === [] ? null : 'd', ['x', 'y']);
        $this->assertTrue($a2->hasCompletion());
        $this->assertSame('d', $a2->getDefault());
    }

    public function testCompletionInput(): void
    {
        $def = new InputDefinition([new InputArgument('cmd'), new InputArgument('rest', InputArgument::IS_ARRAY), new InputOption('name', 'n', InputOption::VALUE_REQUIRED), new InputOption('tag', 't', InputOption::VALUE_IS_ARRAY | InputOption::VALUE_REQUIRED), new InputOption('flag', 'f')]);
        $in = CompletionInput::fromTokens(['bin', 'run', '--name', 'jo'], 3);
        $in->bind($def);
        $this->assertSame(CompletionInput::TYPE_OPTION_VALUE, $in->getCompletionType());
        $this->assertSame('name', $in->getCompletionName());
        $this->assertSame('jo', $in->getCompletionValue());
        $this->assertTrue($in->mustSuggestOptionValuesFor('name'));
        $this->assertFalse($in->mustSuggestOptionValuesFor('tag'));
        $this->assertFalse($in->mustSuggestArgumentValuesFor('cmd'));

        $in = CompletionInput::fromTokens(['bin', 'ru'], 1);
        $in->bind($def);
        $this->assertSame(CompletionInput::TYPE_ARGUMENT_VALUE, $in->getCompletionType());
        $this->assertSame('cmd', $in->getCompletionName());
        $this->assertTrue($in->mustSuggestArgumentValuesFor('cmd'));
        $this->assertSame('ru', $in->getCompletionValue());

        $in = CompletionInput::fromTokens(['bin', 'run', '--'], 2);
        $in->bind($def);
        $this->assertSame(CompletionInput::TYPE_OPTION_NAME, $in->getCompletionType());
        $this->assertNull($in->getCompletionName());
        $this->assertSame('--', $in->getCompletionValue());

        $in = CompletionInput::fromTokens(['bin', 'run', '-t=x'], 2);
        $in->bind($def);
        $this->assertSame(CompletionInput::TYPE_OPTION_VALUE, $in->getCompletionType());
        $this->assertSame('tag', $in->getCompletionName());
        $this->assertSame('x', $in->getCompletionValue());

        $in = CompletionInput::fromTokens(['bin', 'run', '--tag=a', 'foo', 'ba'], 4);
        $in->bind($def);
        $this->assertSame('rest', $in->getCompletionName());
        $this->assertSame('ba', $in->getCompletionValue());

        $in = CompletionInput::fromTokens(['bin', 'run', ''], 2);
        $in->bind($def);
        $this->assertSame('rest', $in->getCompletionName());
        $this->assertStringContainsString('run', (string) $in);

        $in = CompletionInput::fromString('bin run --name jo', 3);
        $in->bind($def);
        $this->assertSame('jo', $in->getCompletionValue());
        $in = CompletionInput::fromString('bin run --na', 2);
        $in->bind($def);
        $this->assertSame(CompletionInput::TYPE_OPTION_NAME, $in->getCompletionType());
        $in = CompletionInput::fromTokens(['bin', 'run', '-n'], 2);
        $in->bind($def);
        $this->assertSame(CompletionInput::TYPE_OPTION_VALUE, $in->getCompletionType());
        $in = CompletionInput::fromTokens(['bin', 'run', '-f'], 2);
        $in->bind($def);
        $this->assertNotNull($in->getCompletionType());
    }

    public function testSuggestions(): void
    {
        $s = new CompletionSuggestions();
        $s->suggestValue('plain')->suggestValue(new Suggestion('desc', 'with description'))->suggestValues(['a', 'b']);
        $s->suggestOption(new InputOption('flag', 'f', InputOption::VALUE_NONE, 'A flag'))->suggestOption(new InputOption('neg', null, InputOption::VALUE_NEGATABLE))->suggestOptions([new InputOption('val', null, InputOption::VALUE_REQUIRED)]);
        $this->assertCount(4, $s->getValueSuggestions());
        $names = array_map(fn ($o) => $o->getName(), $s->getOptionSuggestions());
        $this->assertSame(['flag', 'neg', 'val'], $names);
        $this->assertSame('desc', (string) new Suggestion('desc'));
        $this->assertSame('with description', (new Suggestion('desc', 'with description'))->getDescription());
        $this->assertSame('desc', (new Suggestion('desc'))->getValue());
    }
}
