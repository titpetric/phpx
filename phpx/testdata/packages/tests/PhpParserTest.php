<?php
use PHPUnit\Framework\TestCase;
use PhpParser\{Node, NodeFinder, NodeTraverser, NodeVisitorAbstract, ParserFactory, PrettyPrinter\Standard, BuilderFactory};

final class PhpParserTest extends TestCase
{
    private function parse(string $code): array { return (new ParserFactory)->createForNewestSupportedVersion()->parse($code); }
    public function testParseClass(): void
    {
        $ast = $this->parse('<?php namespace A; class B extends C { public function f(int $x): int { return $x + 1; } }');
        $this->assertCount(1, $ast);
        $this->assertInstanceOf(Node\Stmt\Namespace_::class, $ast[0]);
        $cls = (new NodeFinder)->findFirstInstanceOf($ast, Node\Stmt\Class_::class);
        $this->assertSame('B', $cls->name->toString());
        $this->assertSame('C', $cls->extends->toString());
        $m = $cls->getMethods()[0];
        $this->assertSame('f', $m->name->toString());
        $this->assertSame('int', $m->params[0]->type->toString());
    }
    public function testFindCalls(): void
    {
        $ast = $this->parse('<?php foo(1); bar(2); $x->baz(); foo(3);');
        $calls = (new NodeFinder)->findInstanceOf($ast, Node\Expr\FuncCall::class);
        $this->assertSame(['foo', 'bar', 'foo'], array_map(fn($c) => $c->name->toString(), $calls));
    }
    public function testTraverserRewrite(): void
    {
        $ast = $this->parse('<?php $a = 1 + 2; $b = 10;');
        $t = new NodeTraverser;
        $t->addVisitor(new class extends NodeVisitorAbstract { public function leaveNode(Node $n) { if ($n instanceof Node\Scalar\Int_) { return new Node\Scalar\Int_($n->value * 2); } return null; } });
        $out = (new Standard)->prettyPrint($t->traverse($ast));
        $this->assertSame("\$a = 2 + 4;\n\$b = 20;", $out);
    }
    public function testBuilder(): void
    {
        $f = new BuilderFactory;
        $node = $f->class('Foo')->extend('Bar')->addStmt($f->method('go')->makePublic()->addParam($f->param('n'))->setReturnType('int'))->getNode();
        $code = (new Standard)->prettyPrint([$node]);
        $this->assertStringContainsString('class Foo extends Bar', $code);
        $this->assertStringContainsString('public function go($n): int', $code);
    }
    public function testSyntaxError(): void
    {
        $this->expectException(PhpParser\Error::class);
        $this->parse('<?php class {');
    }
}
