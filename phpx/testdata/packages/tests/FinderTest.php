<?php
use PHPUnit\Framework\TestCase;
use Symfony\Component\Finder\{Finder, Glob, Comparator\NumberComparator, Comparator\DateComparator};

final class FinderTest extends TestCase
{
    private string $dir;
    protected function setUp(): void
    {
        $this->dir = 'tests/.tmp-finder-' . mt_rand();
        @mkdir($this->dir . '/sub/deep', 0777, true);
        file_put_contents($this->dir . '/a.txt', 'aaa');
        file_put_contents($this->dir . '/b.php', '<?php');
        file_put_contents($this->dir . '/sub/c.txt', 'c');
        file_put_contents($this->dir . '/sub/deep/d.md', 'dddd');
    }
    protected function tearDown(): void
    {
        foreach (['/sub/deep/d.md', '/sub/c.txt', '/b.php', '/a.txt'] as $f) { @unlink($this->dir . $f); }
        @rmdir($this->dir . '/sub/deep'); @rmdir($this->dir . '/sub'); @rmdir($this->dir);
    }
    private function names(Finder $f): array { $n = array_map(fn($x) => $x->getRelativePathname(), iterator_to_array($f, false)); sort($n); return $n; }
    public function testFilesByName(): void
    {
        $this->assertSame(['a.txt', 'sub/c.txt'], $this->names((new Finder)->files()->in($this->dir)->name('*.txt')));
    }
    public function testDepthAndExclude(): void
    {
        $this->assertSame(['a.txt', 'b.php'], $this->names((new Finder)->files()->in($this->dir)->depth('== 0')));
        $this->assertSame(['a.txt', 'b.php', 'sub/c.txt'], $this->names((new Finder)->files()->in($this->dir)->exclude('deep')));
    }
    public function testSizeAndNotName(): void
    {
        $this->assertSame(['b.php', 'sub/deep/d.md'], $this->names((new Finder)->files()->in($this->dir)->size('> 3')));
        $this->assertSame(['b.php'], $this->names((new Finder)->files()->in($this->dir)->notName('*.txt')->notName('*.md')));
    }
    public function testDirectoriesAndCount(): void
    {
        $f = (new Finder)->directories()->in($this->dir);
        $this->assertCount(2, $f);
        $this->assertTrue((new Finder)->files()->in($this->dir)->hasResults());
    }
    public function testComparators(): void
    {
        $c = new NumberComparator('>= 2k');
        $this->assertTrue($c->test(2048));
        $this->assertFalse($c->test(100));
        $this->assertSame('>=', $c->getOperator());
    }
}
