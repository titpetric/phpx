<?php
use PHPUnit\Framework\TestCase;
use Latte\Engine;
use Latte\Loaders\StringLoader;

final class LatteEntryTest extends TestCase
{
    private function render(string $template, array $params = []): string
    {
        $latte = new Engine();
        $latte->setLoader(new StringLoader());
        $latte->setTempDirectory(null);
        return $latte->renderToString($template, $params);
    }
    public function testVariable(): void
    {
        $this->assertSame('Hello world!', $this->render('Hello {$name}!', ['name' => 'world']));
    }
    public function testEscaping(): void
    {
        $this->assertSame('a &lt;b&gt; &amp; c', $this->render('{$s}', ['s' => 'a <b> & c']));
        $this->assertSame('a <b>', $this->render('{$s|noescape}', ['s' => 'a <b>']));
    }
    public function testConditionsAndLoops(): void
    {
        $this->assertSame('yes', $this->render('{if $a > 1}yes{else}no{/if}', ['a' => 2]));
        $this->assertSame('1,2,3,', $this->render('{foreach $items as $i}{$i},{/foreach}', ['items' => [1, 2, 3]]));
    }
    public function testFilters(): void
    {
        $this->assertSame('HELLO', $this->render('{$s|upper}', ['s' => 'hello']));
        $this->assertSame('Hello', $this->render('{$s|trim|firstUpper}', ['s' => ' hello ']));
        $this->assertSame('3', $this->render('{$a|length}', ['a' => [1, 2, 3]]));
    }
    public function testNAttributes(): void
    {
        $this->assertSame('<p class="x">hi</p>', $this->render('<p n:class="$on ? x">hi</p>', ['on' => true]));
    }
}
