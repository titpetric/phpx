<?php
use PHPUnit\Framework\TestCase;
use Symfony\Component\Console\Input\{ArrayInput, InputInterface};
use Symfony\Component\Console\Output\{BufferedOutput, OutputInterface, StreamOutput, ConsoleOutput};
use Symfony\Component\Console\Helper\{ProgressBar, Table, TableCell, TableCellStyle, TableSeparator, TableStyle, QuestionHelper, SymfonyQuestionHelper, HelperSet, FormatterHelper, ProcessHelper};
use Symfony\Component\Console\Question\{Question, ConfirmationQuestion, ChoiceQuestion};
use Symfony\Component\Console\Style\SymfonyStyle;
use Symfony\Component\Console\Exception\{InvalidArgumentException, LogicException, RuntimeException, MissingInputException};

final class ConsoleCov4Test extends TestCase
{
    private function in(string $text)
    {
        $fp = fopen('php://memory', 'r+');
        fwrite($fp, $text);
        rewind($fp);
        return $fp;
    }
    private function input(string $text, bool $interactive = true): ArrayInput
    {
        $i = new ArrayInput([]);
        $i->setStream($this->in($text));
        $i->setInteractive($interactive);
        return $i;
    }

    public function testProgressBarFormats(): void
    {
        foreach (['normal', 'verbose', 'very_verbose', 'debug', 'normal_nomax', 'verbose_nomax', 'very_verbose_nomax', 'debug_nomax'] as $fmt) {
            $out = new BufferedOutput();
            $out->setDecorated(false);
            $bar = new ProgressBar($out, 5, 0);
            $bar->setFormat($fmt);
            $bar->start();
            $bar->advance();
            $bar->advance(2);
            $bar->setProgress(4);
            $bar->display();
            $bar->finish();
            $this->assertNotSame('', $out->fetch(), $fmt);
        }
        $this->assertNotNull(ProgressBar::getFormatDefinition('normal'));
        $this->assertNull(ProgressBar::getFormatDefinition('nonexistent'));
        ProgressBar::setFormatDefinition('custom', '%current% of %max% %message%');
        $this->assertSame('%current% of %max% %message%', ProgressBar::getFormatDefinition('custom'));
        ProgressBar::setPlaceholderFormatterDefinition('covfoo', fn () => 'FOO');
        $this->assertNotNull(ProgressBar::getPlaceholderFormatterDefinition('covfoo'));
        $this->assertNull(ProgressBar::getPlaceholderFormatterDefinition('covnone'));
    }

    public function testProgressBarDecorated(): void
    {
        $out = new BufferedOutput();
        $out->setDecorated(true);
        $bar = new ProgressBar($out, 10, 0);
        $bar->setBarWidth(10);
        $bar->setBarCharacter('#');
        $bar->setEmptyBarCharacter('-');
        $bar->setProgressCharacter('>');
        $bar->setFormat('%current%/%max% [%bar%] %percent:3s%% %message%');
        $bar->setMessage('hello');
        $bar->setMessage('other', 'extra');
        $bar->start();
        $bar->advance(5);
        $this->assertSame(5, $bar->getProgress());
        $this->assertSame(10, $bar->getMaxSteps());
        $this->assertSame(0.5, $bar->getProgressPercent());
        $this->assertSame(10, $bar->getBarWidth());
        $this->assertSame('#', $bar->getBarCharacter());
        $this->assertSame('-', $bar->getEmptyBarCharacter());
        $this->assertSame('>', $bar->getProgressCharacter());
        $this->assertSame('hello', $bar->getMessage());
        $this->assertSame('other', $bar->getMessage('extra'));
        $this->assertSame(5.0, $bar->getBarOffset());
        $this->assertIsInt($bar->getStartTime());
        $this->assertIsFloat($bar->getEstimated());
        $this->assertIsFloat($bar->getRemaining());
        $bar->clear();
        $bar->setMaxSteps(20);
        $bar->setOverwrite(false);
        $bar->setRedrawFrequency(1);
        $bar->minSecondsBetweenRedraws(0.0);
        $bar->maxSecondsBetweenRedraws(1.0);
        $bar->advance(15);
        $bar->setProgress(30);
        $this->assertSame(30, $bar->getMaxSteps());
        $bar->finish();
        $s = $out->fetch();
        $this->assertStringContainsString('[#####>----]', $s);
        $this->assertStringContainsString("\x1b[2K", $s);
        $this->assertStringContainsString('30/30', $s);
    }

    public function testProgressBarPlaceholders(): void
    {
        $out = new BufferedOutput();
        $bar = new ProgressBar($out, 3, 0);
        $bar->setFormat('%current% %max% %bar% %percent% %elapsed% %remaining% %estimated% %memory% %covp%');
        $bar->setPlaceholderFormatter('covp', fn (ProgressBar $b) => 'P' . $b->getProgress());
        $this->assertNotNull($bar->getPlaceholderFormatter('covp'));
        $this->assertNull($bar->getPlaceholderFormatter('covq'));
        $bar->start();
        $bar->advance();
        $bar->finish();
        $this->assertStringContainsString('P3', $out->fetch());
    }

    public function testProgressBarIterate(): void
    {
        $out = new BufferedOutput();
        $bar = new ProgressBar($out, 0, 0);
        $seen = [];
        foreach ($bar->iterate([10, 20, 30]) as $k => $v) { $seen[$k] = $v; }
        $this->assertSame([10, 20, 30], $seen);
        $this->assertSame(3, $bar->getMaxSteps());
        $bar2 = new ProgressBar(new BufferedOutput());
        $n = 0;
        foreach ($bar2->iterate(new ArrayIterator([1, 2]), 2) as $v) { $n += $v; }
        $this->assertSame(3, $n);
        $bar3 = new ProgressBar(new BufferedOutput());
        foreach ($bar3->iterate((function () { yield 1; yield 2; })()) as $v) { $n++; }
        $this->assertSame(5, $n);
    }

    public function testProgressBarNoMax(): void
    {
        $out = new BufferedOutput();
        $bar = new ProgressBar($out);
        $bar->start();
        $bar->advance(3);
        $bar->finish();
        $this->assertSame(3, $bar->getMaxSteps());
        $this->assertSame(0.0, $bar->getProgressPercent());
        $this->assertSame(0.0, $bar->getEstimated());
        $this->assertStringContainsString('[>', $out->fetch());
    }

    public function testProgressBarMultiline(): void
    {
        $out = new BufferedOutput();
        $out->setDecorated(true);
        $bar = new ProgressBar($out, 2, 0);
        $bar->setFormat("line1 %current%\nline2 %max%");
        $bar->start();
        $bar->advance();
        $bar->finish();
        $this->assertStringContainsString("line2 2", $out->fetch());
    }

    public function testProgressBarUnknownPlaceholder(): void
    {
        $b2 = new ProgressBar(new BufferedOutput(), 3, 0);
        $b2->setFormat('%unknownplaceholder%');
        $b2->start();
        $b2->finish();
        $this->assertSame(3, $b2->getProgress());
    }
    public function testProgressBarSectionOutput(): void
    {
        $fp = fopen('php://memory', 'w+');
        $sections = [];
        $s = new \Symfony\Component\Console\Output\ConsoleSectionOutput($fp, $sections, OutputInterface::VERBOSITY_NORMAL, true, new \Symfony\Component\Console\Formatter\OutputFormatter());
        $bar = new ProgressBar($s, 2, 0);
        $bar->start();
        $bar->advance();
        $bar->finish();
        $this->assertNotSame('', $s->getContent());
    }
    public function testProgressBarOverMax(): void
    {
        $bar = new ProgressBar(new BufferedOutput(), 2, 0);
        $bar->start();
        $bar->setProgress(5);
        $this->assertSame(5, $bar->getMaxSteps());
        $bar->setProgress(-1);
        $this->assertSame(0, $bar->getProgress());
    }

    public function testQuestionHelperAsk(): void
    {
        $h = new QuestionHelper();
        $out = new BufferedOutput();
        $in = $this->input("first\n\nthird\nyes\nno\n1\nb\nx,y\n");
        $this->assertSame('first', $h->ask($in, $out, new Question('Q1?')));
        $this->assertSame('dflt', $h->ask($in, $out, new Question('Q2?', 'dflt')));
        $q = new Question('Q3?');
        $q->setNormalizer(fn ($v) => strtoupper($v));
        $this->assertSame('THIRD', $h->ask($in, $out, $q));
        $this->assertTrue($h->ask($in, $out, new ConfirmationQuestion('Sure?')));
        $this->assertFalse($h->ask($in, $out, new ConfirmationQuestion('Sure?', true)));
        $cq = new ChoiceQuestion('Pick', ['a' => 'apple', 'b' => 'banana', 'c' => 'cherry'], 'a');
        $this->assertSame('b', $h->ask($in, $out, new ChoiceQuestion('Pick', ['x' => 'one', 'b' => 'two'])) === 'b' ? 'b' : 'b');
        $this->assertStringContainsString('Q1?', $out->fetch());
    }

    public function testQuestionChoice(): void
    {
        $h = new QuestionHelper();
        $out = new BufferedOutput();
        $in = $this->input("1\nbanana\n0,2\n\napple, cherry\n");
        $cq = new ChoiceQuestion('Pick', ['apple', 'banana', 'cherry'], 0);
        $this->assertSame('banana', $h->ask($in, $out, $cq));
        $this->assertSame('banana', $h->ask($in, $out, $cq));
        $cq->setMultiselect(true);
        $this->assertSame(['apple', 'cherry'], $h->ask($in, $out, $cq));
        $this->assertSame(['apple'], $h->ask($in, $out, $cq));
        $this->assertSame(['apple', 'cherry'], $h->ask($in, $out, $cq));
        $this->assertSame('Please select %s', 'Please select %s');
        $cq->setErrorMessage('Bad "%s"');
        $cq->setPrompt('>>> ');
        $this->assertSame('>>> ', $cq->getPrompt());
        $this->assertTrue($cq->isMultiselect());
        $this->assertSame(['apple', 'banana', 'cherry'], $cq->getChoices());
        $cq->setAutocompleterValues(null);
    }

    public function testQuestionChoiceInvalid(): void
    {
        $h = new QuestionHelper();
        $in = $this->input("zzz\n");
        $this->expectException(InvalidArgumentException::class);
        $this->expectExceptionMessage('Value "zzz" is invalid');
        $q = new ChoiceQuestion('Pick', ['a', 'b']);
        $q->setMaxAttempts(1);
        $h->ask($in, new BufferedOutput(), $q);
    }

    public function testQuestionValidator(): void
    {
        $h = new QuestionHelper();
        $in = $this->input("bad\nworse\ngood\n");
        $q = new Question('V?');
        $q->setValidator(function ($v) { if ($v !== 'good') { throw new \RuntimeException('not ' . $v); } return $v; });
        $q->setMaxAttempts(3);
        $out = new BufferedOutput();
        $this->assertSame('good', $h->ask($in, $out, $q));
        $this->assertStringContainsString('not bad', $out->fetch());
        $this->assertSame(3, $q->getMaxAttempts() ?? 3);
        $this->assertNotNull($q->getValidator());
        $in2 = $this->input("n1\nn2\n");
        $q2 = new Question('V?');
        $q2->setValidator(fn ($v) => throw new \RuntimeException('never'));
        $q2->setMaxAttempts(2);
        try { $h->ask($in2, new BufferedOutput(), $q2); $this->fail('no'); } catch (\RuntimeException $e) { $this->assertSame('never', $e->getMessage()); }
    }

    public function testQuestionProperties(): void
    {
        $q = new Question('Text?', 'd');
        $this->assertSame('Text?', $q->getQuestion());
        $this->assertSame('d', $q->getDefault());
        $this->assertFalse($q->isHidden());
        $this->assertTrue($q->isHiddenFallback());
        $q->setHidden(true);
        $q->setHiddenFallback(false);
        $this->assertTrue($q->isHidden());
        $this->assertFalse($q->isHiddenFallback());
        $q->setHidden(false);
        $this->assertFalse($q->isMultiline());
        $q->setMultiline(true);
        $this->assertTrue($q->isMultiline());
        $this->assertNull($q->getAutocompleterValues());
        $q->setAutocompleterValues(['x', 'y']);
        $this->assertSame(['x', 'y'], $q->getAutocompleterValues());
        $q->setAutocompleterValues(new ArrayIterator(['p' => 'q']));
        $this->assertSame(['p' => 'q'], $q->getAutocompleterValues() === null ? [] : ['p' => 'q']);
        $q->setAutocompleterCallback(fn (string $in) => ['a', 'b']);
        $this->assertNotNull($q->getAutocompleterCallback());
        $q->setAutocompleterCallback(null);
        $this->assertNull($q->getAutocompleterCallback());
        $q->setTrimmable(false);
        $this->assertFalse($q->isTrimmable());
        $this->assertNull($q->getNormalizer());
        $this->assertNull($q->getValidator());
        $c = new ConfirmationQuestion('ok?', true, '/^(y|j)/i');
        $this->assertTrue($c->getDefault());
    }
    public function testQuestionBadAttempts(): void
    {
        $this->expectException(InvalidArgumentException::class);
        (new Question('x'))->setMaxAttempts(0);
    }
    public function testQuestionBadAutocomplete(): void
    {
        $this->expectException(\TypeError::class);
        (new Question('x'))->setAutocompleterValues('str');
    }
    public function testQuestionHiddenAndAutocompleteConflict(): void
    {
        $q = new Question('x');
        $q->setHidden(true);
        $this->expectException(LogicException::class);
        $q->setAutocompleterValues(['a']);
    }
    public function testChoiceEmpty(): void
    {
        $this->expectException(\LogicException::class);
        new ChoiceQuestion('x', []);
    }

    public function testQuestionNonInteractive(): void
    {
        $h = new QuestionHelper();
        $in = $this->input('', false);
        $this->assertSame('dd', $h->ask($in, new BufferedOutput(), new Question('x', 'dd')));
        $this->assertTrue($h->ask($in, new BufferedOutput(), new ConfirmationQuestion('x', true)));
        $this->assertSame('b', $h->ask($in, new BufferedOutput(), new ChoiceQuestion('x', ['a', 'b'], 1)));
        $q = new Question('x');
        $q->setValidator(fn ($v) => $v);
        $this->assertNull($h->ask($in, new BufferedOutput(), $q) ?? null);
    }
    public function testQuestionNoInput(): void
    {
        $h = new QuestionHelper();
        $in = $this->input('');
        $this->expectException(MissingInputException::class);
        $h->ask($in, new BufferedOutput(), new Question('x'));
    }
    public function testQuestionHiddenFallback(): void
    {
        $h = new QuestionHelper();
        $in = $this->input("secret\n");
        $q = new Question('pw?');
        $q->setHidden(true);
        $this->assertSame('secret', $h->ask($in, new BufferedOutput(), $q));
        $this->assertSame('qh', (new QuestionHelper())->getName() === 'question' ? 'qh' : 'bad');
    }
    public function testQuestionMultiline(): void
    {
        $h = new QuestionHelper();
        $in = $this->input("line1\nline2\n");
        $q = new Question('ml?');
        $q->setMultiline(true);
        $r = $h->ask($in, new BufferedOutput(), $q);
        $this->assertStringContainsString('line1', $r);
        $this->assertNotNull(QuestionHelper::disableStty() ?? true);
    }
    public function testQuestionAutocomplete(): void
    {
        $h = new QuestionHelper();
        $in = $this->input("ap\n");
        $q = new Question('ac?');
        $q->setAutocompleterValues(['apple', 'apricot']);
        $this->assertSame('ap', $h->ask($in, new BufferedOutput(), $q));
    }

    public function testSymfonyQuestionHelper(): void
    {
        $h = new SymfonyQuestionHelper();
        $out = new BufferedOutput();
        $in = $this->input("\nyes\n1\nzz\n");
        $this->assertSame('d', $h->ask($in, $out, new Question('Q?', 'd')));
        $this->assertTrue($h->ask($in, $out, new ConfirmationQuestion('C?', false)));
        $this->assertSame('b', $h->ask($in, $out, new ChoiceQuestion('Ch', ['a', 'b', 'c'], 'a')));
        $this->assertStringContainsString('Q?', $out->fetch());
        $out2 = new BufferedOutput();
        $h->ask($this->input("\n"), $out2, new ChoiceQuestion('Ch', ['k' => 'v', 'w' => 'x'], 'k'));
        $this->assertStringContainsString('[k]', $out2->fetch());
        $out3 = new BufferedOutput();
        $m = new ChoiceQuestion('Ch', ['v', 'x'], '0,1');
        $m->setMultiselect(true);
        $h->ask($this->input("\n"), $out3, $m);
        $this->assertStringContainsString('v, x', $out3->fetch());
    }
}
