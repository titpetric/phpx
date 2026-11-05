<?php
use PHPUnit\Framework\TestCase;
use Twig\Environment;
use Twig\Loader\ArrayLoader;

final class TwigEntryTest extends TestCase
{
    private function render(string $tpl, array $ctx = []): string
    {
        $twig = new Environment(new ArrayLoader(['t' => $tpl]), ['cache' => false]);
        return $twig->render('t', $ctx);
    }
    public function testVariable(): void
    {
        $this->assertSame('Hello world!', $this->render('Hello {{ name }}!', ['name' => 'world']));
    }
    public function testEscaping(): void
    {
        $this->assertSame('a &lt;b&gt; &amp; c', $this->render('{{ s }}', ['s' => 'a <b> & c']));
        $this->assertSame('a <b>', $this->render('{{ s|raw }}', ['s' => 'a <b>']));
    }
    public function testControlFlow(): void
    {
        $this->assertSame('yes', $this->render('{% if a > 1 %}yes{% else %}no{% endif %}', ['a' => 2]));
        $this->assertSame('1,2,3,', $this->render('{% for i in items %}{{ i }},{% endfor %}', ['items' => [1, 2, 3]]));
    }
    public function testFilters(): void
    {
        $this->assertSame('HELLO', $this->render('{{ s|upper }}', ['s' => 'hello']));
        $this->assertSame('Hello', $this->render('{{ s|trim|capitalize }}', ['s' => ' hello ']));
        $this->assertSame('3', $this->render('{{ a|length }}', ['a' => [1, 2, 3]]));
        $this->assertSame('a,b', $this->render('{{ a|join(",") }}', ['a' => ['a', 'b']]));
    }
    public function testInheritanceAndMacros(): void
    {
        $twig = new Environment(new ArrayLoader([
            'base' => '[{% block body %}base{% endblock %}]',
            'child' => '{% extends "base" %}{% block body %}child {{ x }}{% endblock %}',
            'm' => '{% macro hi(n) %}hi {{ n }}{% endmacro %}{{ _self.hi("z") }}',
        ]), ['cache' => false]);
        $this->assertSame('[child 7]', $twig->render('child', ['x' => 7]));
        $this->assertSame('hi z', $twig->render('m'));
    }
}
