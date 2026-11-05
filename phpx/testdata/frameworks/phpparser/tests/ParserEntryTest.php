<?php
use PHPUnit\Framework\TestCase;
use PhpParser\ParserFactory;
use PhpParser\PrettyPrinter\Standard;
use PhpParser\NodeTraverser;
use PhpParser\NodeVisitorAbstract;
use PhpParser\Node;

final class ParserEntryTest extends TestCase
{
    public function testParseAndPrint(): void
    {
        $parser = (new ParserFactory())->createForNewestSupportedVersion();
        $ast = $parser->parse('<?php $a = 1 + 2; echo $a;');
        $this->assertCount(2, $ast);
        $this->assertSame("\$a = 1 + 2;\necho \$a;", (new Standard())->prettyPrint($ast));
    }
    public function testTraverse(): void
    {
        $parser = (new ParserFactory())->createForNewestSupportedVersion();
        $ast = $parser->parse('<?php function f($x) { return $x * 2; } class A { function m() {} }');
        $names = [];
        $t = new NodeTraverser();
        $t->addVisitor(new class($names) extends NodeVisitorAbstract {
            public function __construct(private array &$names) {}
            public function enterNode(Node $n) {
                if ($n instanceof Node\Stmt\Function_ || $n instanceof Node\Stmt\ClassMethod || $n instanceof Node\Stmt\Class_) { $this->names[] = (string) $n->name; }
                return null;
            }
        });
        $t->traverse($ast);
        $this->assertSame(['f', 'A', 'm'], $names);
    }
    public function testRewrite(): void
    {
        $parser = (new ParserFactory())->createForNewestSupportedVersion();
        $ast = $parser->parse('<?php echo 1 + 2;');
        $t = new NodeTraverser();
        $t->addVisitor(new class extends NodeVisitorAbstract {
            public function leaveNode(Node $n) { if ($n instanceof Node\Expr\BinaryOp\Plus) { return new Node\Expr\BinaryOp\Mul($n->left, $n->right); } return null; }
        });
        $this->assertSame('echo 1 * 2;', (new Standard())->prettyPrint($t->traverse($ast)));
    }
}
