<?php
use PHPUnit\Framework\TestCase;
use Symfony\Component\Finder\{Finder, Glob, Gitignore};
use Symfony\Component\Finder\Comparator\{NumberComparator, DateComparator};

final class FinderMoreTest extends TestCase
{
    private string $dir;
    protected function setUp(): void
    {
        $this->dir = sys_get_temp_dir() . '/finder_more_' . getmypid();
        @mkdir($this->dir . '/sub/deep', 0777, true);
        file_put_contents($this->dir . '/a.txt', 'aaaa');
        file_put_contents($this->dir . '/b.php', "<?php\n");
        file_put_contents($this->dir . '/sub/c.txt', 'cc');
        file_put_contents($this->dir . '/sub/deep/d.md', 'd');
    }
    protected function tearDown(): void
    {
        foreach (['/sub/deep/d.md', '/sub/c.txt', '/b.php', '/a.txt'] as $f) { @unlink($this->dir . $f); }
        @rmdir($this->dir . '/sub/deep'); @rmdir($this->dir . '/sub'); @rmdir($this->dir);
    }
    private function names(Finder $f): array
    {
        $out = [];
        foreach ($f as $file) { $out[] = $file->getRelativePathname(); }
        sort($out);
        return $out;
    }
    public function testNameAndNotName(): void
    {
        $this->assertSame(['a.txt', 'sub/c.txt'], $this->names((new Finder)->files()->in($this->dir)->name('*.txt')));
        $this->assertSame(['b.php', 'sub/deep/d.md'], $this->names((new Finder)->files()->in($this->dir)->notName('*.txt')));
        $this->assertSame(['b.php'], $this->names((new Finder)->files()->in($this->dir)->name('/\.php$/')));
    }
    public function testDepth(): void
    {
        $this->assertSame(['a.txt', 'b.php', 'sub'], $this->names((new Finder)->in($this->dir)->depth('== 0')));
        $this->assertSame(['sub/c.txt', 'sub/deep'], $this->names((new Finder)->in($this->dir)->depth('== 1')));
    }
    public function testSizeAndContains(): void
    {
        $this->assertSame(['a.txt', 'b.php'], $this->names((new Finder)->files()->in($this->dir)->size('> 3')));
        $this->assertSame(['sub/c.txt'], $this->names((new Finder)->files()->in($this->dir)->contains('cc')));
        $this->assertSame(['a.txt', 'b.php', 'sub/deep/d.md'], $this->names((new Finder)->files()->in($this->dir)->notContains('cc')));
    }
    public function testExcludeAndDirectories(): void
    {
        $this->assertSame(['a.txt', 'b.php'], $this->names((new Finder)->files()->in($this->dir)->exclude('sub')));
        $this->assertSame(['sub', 'sub/deep'], $this->names((new Finder)->directories()->in($this->dir)));
    }
    public function testSortAndCount(): void
    {
        $f = (new Finder)->files()->in($this->dir)->sortByName();
        $names = array_map(fn ($x) => $x->getFilename(), iterator_to_array($f, false));
        $this->assertSame(['a.txt', 'b.php', 'c.txt', 'd.md'], $names);
        $this->assertSame(4, count($f));
        $this->assertTrue($f->hasResults());
        $this->assertSame('aaaa', array_values(iterator_to_array((new Finder)->files()->in($this->dir)->name('a.txt')))[0]->getContents());
    }
    public function testFilterAndPath(): void
    {
        $f = (new Finder)->files()->in($this->dir)->filter(fn (SplFileInfo $i) => $i->getSize() === 1);
        $this->assertSame(['sub/deep/d.md'], $this->names($f));
        $this->assertSame(['sub/c.txt', 'sub/deep/d.md'], $this->names((new Finder)->files()->in($this->dir)->path('sub')));
        $this->assertSame(['b.php'], $this->names((new Finder)->files()->in($this->dir)->name('b.*')->ignoreDotFiles(true)));
    }
    public function testGlobAndComparators(): void
    {
        $this->assertSame(1, preg_match(Glob::toRegex('*.txt'), 'x.txt'));
        $this->assertSame(0, preg_match(Glob::toRegex('*.txt'), 'x/y.txt'));
        $c = new NumberComparator('<= 2k');
        $this->assertTrue($c->test(1999));
        $this->assertFalse($c->test(2001));
        $d = new DateComparator('> 2000-01-01');
        $this->assertSame('>', $d->getOperator());
        $this->assertSame(1, preg_match(Gitignore::toRegex('*.txt'), 'a.txt'));
    }
}
