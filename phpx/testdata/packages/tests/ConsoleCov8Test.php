<?php
use PHPUnit\Framework\TestCase;
use Symfony\Component\Console\Attribute\{Argument, Ask, Interact, MapInput, Option};
use Symfony\Component\Console\Input\{ArrayInput, InputDefinition, InputInterface, InputOption, InputArgument};
use Symfony\Component\Console\Output\BufferedOutput;
use Symfony\Component\Console\Style\SymfonyStyle;
use Symfony\Component\Console\Exception\LogicException;

enum ConsoleCovColor: string { case Red = 'red'; case Blue = 'blue'; }

final class ConsoleCovProps
{
    #[Argument(description: 'The user')]
    public string $user;
    #[Argument(name: 'cnt')]
    public int $count = 2;
    #[Argument]
    public array $extras = [];
    #[Option(description: 'Loud', shortcut: 'l')]
    public bool $loud = false;
    #[Option]
    public bool $colorful = true;
    #[Option]
    public ?string $note = null;
    #[Option(suggestedValues: ['a', 'b'])]
    public string $mode = 'a';
    #[Option]
    public ?ConsoleCovColor $color = null;
    #[Option]
    public float $ratio = 1.5;
    #[Option]
    public array $tags = [];
    #[Option]
    public bool|string $either = false;
    public string $plain = 'p';
}
final class ConsoleCovHolder
{
    #[MapInput]
    public ConsoleCovProps $props;
    #[Argument]
    public ConsoleCovColor $enumArg;
    #[Argument]
    public float $fl = 0.0;
    #[Argument]
    public bool $bl = false;
    #[Argument(suggestedValues: ['x', 'y'])]
    public ?string $sv = null;
}
final class ConsoleCovBadProps
{
    #[Argument]
    public $untyped;
    #[Argument]
    public object $badType;
    #[Option]
    public $untypedOpt = null;
    #[Option]
    public string|int|null $union = null;
    #[Option]
    public stdClass $badOptType;
    #[Option]
    public ?bool $nullableBool = false;
    #[Option]
    public ?string $nullableDefault = 'x';
    #[Option]
    public string $nodefault;
    #[MapInput]
    public $untypedMap;
    #[MapInput]
    public NoSuchClassCov $missing;
    #[MapInput]
    public ConsoleCovPlainDto $noSpecs;
}
final class ConsoleCovPlainDto { public string $a = 'a'; }
final class ConsoleCovInteractive
{
    #[Argument, Ask('Who are you?', 'nobody')]
    public string $who;
    #[Argument, Ask('Secret?', hidden: true)]
    public string $secret;
    #[Option, Ask('Sure?')]
    public bool $sure = false;
    #[Argument, Ask('Items', maxAttempts: 2)]
    public array $items;
    public bool $interacted = false;
    #[Interact]
    public function ask(SymfonyStyle $io): void { $this->interacted = true; }
}

final class ConsoleCovInteractiveHolder
{
    #[MapInput]
    public ConsoleCovInteractive $in;
}

final class ConsoleCov8Test extends TestCase
{
    private function prop(string $class, string $name): ReflectionProperty { return new ReflectionProperty($class, $name); }

    public function testArgumentFromProperty(): void
    {
        $a = Argument::tryFrom($this->prop(ConsoleCovProps::class, 'user'));
        $this->assertSame('user', $a->name);
        $ia = $a->toInputArgument();
        $this->assertTrue($ia->isRequired());
        $this->assertSame('The user', $ia->getDescription());
        $this->assertTrue($a->isRequired());
        $b = Argument::tryFrom($this->prop(ConsoleCovProps::class, 'count'));
        $this->assertSame('cnt', $b->name);
        $this->assertFalse($b->toInputArgument()->isRequired());
        $this->assertSame(2, $b->toInputArgument()->getDefault());
        $c = Argument::tryFrom($this->prop(ConsoleCovProps::class, 'extras'));
        $this->assertTrue($c->toInputArgument()->isArray());
        $this->assertNull(Argument::tryFrom($this->prop(ConsoleCovProps::class, 'plain')));
        $in = new ArrayInput(['user' => 'ann', 'cnt' => '9', 'extras' => ['x', 'y']], new InputDefinition([$a->toInputArgument(), $b->toInputArgument(), $c->toInputArgument()]));
        $this->assertSame('ann', $a->resolveValue($in));
        $this->assertEquals(9, $b->resolveValue($in));
        $this->assertSame(['x', 'y'], $c->resolveValue($in));
        $h = [];
        foreach (['enumArg', 'fl', 'bl', 'sv'] as $p) { $h[$p] = Argument::tryFrom($this->prop(ConsoleCovHolder::class, $p)); }
        $this->assertNotNull($h['enumArg']->toInputArgument());
        $def = new InputDefinition([$h['enumArg']->toInputArgument(), $h['fl']->toInputArgument(), $h['bl']->toInputArgument(), $h['sv']->toInputArgument()]);
        $in2 = new ArrayInput(['enum-arg' => 'blue', 'fl' => '2.5', 'bl' => '1', 'sv' => 'x'], $def);
        $this->assertSame(ConsoleCovColor::Blue, $h['enumArg']->resolveValue($in2));
        $this->assertEquals(2.5, $h['fl']->resolveValue($in2));
        $this->assertEquals(true, $h['bl']->resolveValue($in2));
        $this->assertSame('x', $h['sv']->resolveValue($in2));
        $in3 = new ArrayInput(['enum-arg' => 'bogus'], $def);
        $this->expectException(\Symfony\Component\Console\Exception\InvalidArgumentException::class);
        $h['enumArg']->resolveValue($in3);
    }

    public function testOptionFromProperty(): void
    {
        $opts = [];
        foreach (['loud', 'colorful', 'note', 'mode', 'color', 'ratio', 'tags', 'either'] as $p) {
            $opts[$p] = Option::tryFrom($this->prop(ConsoleCovProps::class, $p));
            $this->assertNotNull($opts[$p], $p);
        }
        $this->assertNull(Option::tryFrom($this->prop(ConsoleCovProps::class, 'plain')));
        $io = array_map(fn ($o) => $o->toInputOption(), $opts);
        $this->assertSame('l', $io['loud']->getShortcut());
        $this->assertFalse($io['loud']->acceptValue());
        $this->assertTrue($io['colorful']->isNegatable());
        $this->assertTrue($io['note']->isValueRequired());
        $this->assertTrue($io['tags']->isArray());
        $this->assertTrue($io['mode']->hasCompletion());
        $this->assertTrue($io['color']->hasCompletion());
        $def = new InputDefinition(array_values($io));
        $in = new ArrayInput(['--loud' => true, '--no-colorful' => true, '--note' => 'n', '--mode' => 'b', '--color' => 'red', '--ratio' => '2.25', '--tags' => ['t1', 't2'], '--either' => 'str'], $def);
        $this->assertTrue($opts['loud']->resolveValue($in));
        $this->assertFalse($opts['colorful']->resolveValue($in));
        $this->assertSame('n', $opts['note']->resolveValue($in));
        $this->assertSame('b', $opts['mode']->resolveValue($in));
        $this->assertSame(ConsoleCovColor::Red, $opts['color']->resolveValue($in));
        $this->assertEquals(2.25, $opts['ratio']->resolveValue($in));
        $this->assertSame(['t1', 't2'], $opts['tags']->resolveValue($in));
        $this->assertSame('str', $opts['either']->resolveValue($in));
        $in2 = new ArrayInput([], $def);
        $this->assertFalse($opts['loud']->resolveValue($in2));
        $this->assertNull($opts['note']->resolveValue($in2));
        $this->assertNull($opts['color']->resolveValue($in2));
        $this->assertFalse($opts['either']->resolveValue($in2));
        $in3 = new ArrayInput(['--either' => null], $def);
        $this->assertTrue($opts['either']->resolveValue($in3));
    }

    public function testBadProperties(): void
    {
        $n = 0;
        foreach (['untyped' => Argument::class, 'badType' => Argument::class, 'untypedOpt' => Option::class, 'union' => Option::class, 'badOptType' => Option::class, 'nullableBool' => Option::class, 'nullableDefault' => Option::class, 'nodefault' => Option::class, 'untypedMap' => MapInput::class, 'missing' => MapInput::class, 'noSpecs' => MapInput::class] as $p => $cls) {
            try { $cls::tryFrom($this->prop(ConsoleCovBadProps::class, $p)); $this->fail("$p accepted"); } catch (LogicException $e) { $n++; $this->assertNotSame('', $e->getMessage()); }
        }
        $this->assertSame(11, $n);
    }

    public function testMapInput(): void
    {
        $m = MapInput::tryFrom($this->prop(ConsoleCovHolder::class, 'props'));
        $args = iterator_to_array($m->getArguments(), false);
        $opts = iterator_to_array($m->getOptions(), false);
        $this->assertCount(3, $args);
        $this->assertCount(8, $opts);
        $def = new InputDefinition();
        foreach ($args as $a) { $def->addArgument($a->toInputArgument()); }
        foreach ($opts as $o) { $def->addOption($o->toInputOption()); }
        $in = new ArrayInput(['user' => 'ann', '--loud' => true, '--tags' => ['q']], $def);
        $obj = $m->resolveValue($in);
        $this->assertInstanceOf(ConsoleCovProps::class, $obj);
        $this->assertSame('ann', $obj->user);
        $this->assertSame(2, $obj->count);
        $this->assertTrue($obj->loud);
        $this->assertSame(['q'], $obj->tags);
        $this->assertSame('a', $obj->mode);
        $obj->user = 'changed';
        $obj->note = 'set';
        $m->setValue($in, $obj);
        $this->assertSame('changed', $in->getArgument('user'));
        $this->assertSame('set', $in->getOption('note'));
        $this->assertSame([], iterator_to_array($m->getPropertyInteractions(), false));
        $this->assertSame([], iterator_to_array($m->getMethodInteractions(), false));
    }

    public function testInteractive(): void
    {
        $m = new ConsoleCovInteractiveHolder();
        $mi = MapInput::tryFrom(new ReflectionProperty($m, 'in'));
        $props = iterator_to_array($mi->getPropertyInteractions(), false);
        $methods = iterator_to_array($mi->getMethodInteractions(), false);
        $this->assertGreaterThanOrEqual(2, count($props));
        $this->assertCount(1, $methods);
        $def = new InputDefinition();
        foreach ($mi->getArguments() as $a) { $def->addArgument($a->toInputArgument()); }
        foreach ($mi->getOptions() as $o) { $def->addOption($o->toInputOption()); }
        $fp = fopen('php://memory', 'r+');
        fwrite($fp, "\nsecretval\nitem1\nitem2\n\n");
        rewind($fp);
        $input = new ArrayInput([]);
        $input->bind($def);
        $input->setStream($fp);
        $out = new BufferedOutput();
        $io = new SymfonyStyle($input, $out);
        $resolver = function (ReflectionFunction $f, InputInterface $i, $o) use ($io): array { return array_map(fn ($p) => match ($p->getType()?->getName()) { SymfonyStyle::class => $io, InputInterface::class => $i, default => $o }, $f->getParameters()); };
        foreach ($props as $p) { $p->interact($input, $out, $resolver); }
        foreach ($methods as $p) { $p->interact($input, $out, $resolver); }
        $this->assertStringContainsString('Who are you?', $out->fetch());
        $this->assertSame('nobody', $input->getArgument('who'));
    }
}
