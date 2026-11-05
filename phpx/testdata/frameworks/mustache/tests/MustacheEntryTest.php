<?php
use PHPUnit\Framework\TestCase;

final class MustacheEntryTest extends TestCase
{
    private function render(string $t, $ctx = []): string { return (new Mustache_Engine())->render($t, $ctx); }
    public function testVariable(): void
    {
        $this->assertSame('Hello world!', $this->render('Hello {{name}}!', ['name' => 'world']));
        $this->assertSame('a &lt;b&gt;', $this->render('{{s}}', ['s' => 'a <b>']));
        $this->assertSame('a <b>', $this->render('{{{s}}}', ['s' => 'a <b>']));
    }
    public function testSectionsAndLoops(): void
    {
        $this->assertSame('1,2,3,', $this->render('{{#items}}{{.}},{{/items}}', ['items' => [1, 2, 3]]));
        $this->assertSame('yes', $this->render('{{#a}}yes{{/a}}{{^a}}no{{/a}}', ['a' => true]));
        $this->assertSame('no', $this->render('{{#a}}yes{{/a}}{{^a}}no{{/a}}', ['a' => false]));
    }
    public function testNestedAndLambda(): void
    {
        $this->assertSame('x:7', $this->render('{{#o}}{{k}}:{{v}}{{/o}}', ['o' => ['k' => 'x', 'v' => 7]]));
        $this->assertSame('<b>hi</b>', $this->render('{{#bold}}hi{{/bold}}', ['bold' => fn($text) => '<b>' . $text . '</b>']));
    }
    public function testPartials(): void
    {
        $m = new Mustache_Engine(['partials' => ['p' => 'P{{x}}']]);
        $this->assertSame('[P1]', $m->render('[{{>p}}]', ['x' => 1]));
    }
}
