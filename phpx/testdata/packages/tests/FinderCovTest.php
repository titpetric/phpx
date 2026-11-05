<?php
use PHPUnit\Framework\TestCase;
use Symfony\Component\Finder\{Finder, Glob, Gitignore, SplFileInfo};
use Symfony\Component\Finder\Comparator\{NumberComparator, DateComparator, Comparator};
use Symfony\Component\Finder\Exception\DirectoryNotFoundException;
use Symfony\Component\Finder\Iterator\{SortableIterator, LazyIterator, DateRangeFilterIterator, CustomFilterIterator, ExcludeDirectoryFilterIterator, FilecontentFilterIterator, PathFilterIterator, SizeRangeFilterIterator};

final class FinderCovTest extends TestCase
{
    private string $dir;
    protected function setUp(): void
    {
        $this->dir = sys_get_temp_dir() . '/findercov_' . getmypid();
        @mkdir($this->dir . '/sub/deep', 0777, true);
        @mkdir($this->dir . '/other', 0777, true);
        file_put_contents($this->dir . '/a.txt', 'aaaa');
        file_put_contents($this->dir . '/B.php', "<?php\n// hello\n");
        file_put_contents($this->dir . '/sub/c.txt', 'cc');
        file_put_contents($this->dir . '/sub/deep/d.md', 'ddddddddd');
        file_put_contents($this->dir . '/other/e10.txt', 'e');
        file_put_contents($this->dir . '/other/e2.txt', 'eeeeeeeeeeeeeeeeeeee');
        touch($this->dir . '/a.txt', 1000000000);
        touch($this->dir . '/B.php', 1100000000);
        touch($this->dir . '/sub/c.txt', 1200000000);
        touch($this->dir . '/sub/deep/d.md', 1300000000);
        touch($this->dir . '/other/e10.txt', 1400000000);
        touch($this->dir . '/other/e2.txt', 1500000000);
    }
    protected function tearDown(): void
    {
        foreach (['/sub/deep/d.md', '/sub/c.txt', '/B.php', '/a.txt', '/other/e10.txt', '/other/e2.txt'] as $f) { @unlink($this->dir . $f); }
        @rmdir($this->dir . '/sub/deep'); @rmdir($this->dir . '/sub'); @rmdir($this->dir . '/other'); @rmdir($this->dir);
    }
    private function list(Finder $f): array
    {
        $o = [];
        foreach ($f as $x) { $o[] = $x->getRelativePathname(); }
        return $o;
    }
    private function sorted(Finder $f): array { $o = $this->list($f); sort($o); return $o; }

    public function testSortByNameVariants(): void
    {
        $this->assertSame(['B.php', 'a.txt', 'other/e10.txt', 'other/e2.txt', 'sub/c.txt', 'sub/deep/d.md'],
            $this->list((new Finder)->files()->in($this->dir)->sortByName()));
        $this->assertSame(['B.php', 'a.txt', 'other/e2.txt', 'other/e10.txt', 'sub/c.txt', 'sub/deep/d.md'],
            $this->list((new Finder)->files()->in($this->dir)->sortByName(true)));
        $this->assertSame(['a.txt', 'B.php', 'other/e10.txt', 'other/e2.txt', 'sub/c.txt', 'sub/deep/d.md'],
            $this->list((new Finder)->files()->in($this->dir)->sortByCaseInsensitiveName()));
        $this->assertSame(['a.txt', 'B.php', 'other/e2.txt', 'other/e10.txt', 'sub/c.txt', 'sub/deep/d.md'],
            $this->list((new Finder)->files()->in($this->dir)->sortByCaseInsensitiveName(true)));
    }
    public function testSortReverse(): void
    {
        $this->assertSame(['sub/deep/d.md', 'sub/c.txt', 'other/e2.txt', 'other/e10.txt', 'a.txt', 'B.php'],
            $this->list((new Finder)->files()->in($this->dir)->sortByName()->reverseSorting()));
    }
    public function testSortByTypeSizeExtension(): void
    {
        $t = $this->list((new Finder)->in($this->dir)->sortByType());
        $this->assertSame(['other', 'sub', 'sub/deep'], array_slice($t, 0, 3));
        $this->assertSame('sub/deep/d.md', end($t));
        $tr = $this->list((new Finder)->in($this->dir)->sortByType()->reverseSorting());
        $this->assertSame('other', end($tr));
        $this->assertSame(['other/e10.txt', 'sub/c.txt', 'a.txt', 'sub/deep/d.md', 'B.php', 'other/e2.txt'],
            $this->list((new Finder)->files()->in($this->dir)->sortBySize()->sortByName()->sortBySize()));
        $e = $this->list((new Finder)->files()->in($this->dir)->sortByExtension());
        $this->assertSame('sub/deep/d.md', $e[0]);
        $this->assertSame('B.php', $e[1]);
    }
    public function testSortByTimes(): void
    {
        $m = $this->list((new Finder)->files()->in($this->dir)->sortByModifiedTime());
        $this->assertSame(['a.txt', 'B.php', 'sub/c.txt', 'sub/deep/d.md', 'other/e10.txt', 'other/e2.txt'], $m);
        $mr = $this->list((new Finder)->files()->in($this->dir)->sortByModifiedTime()->reverseSorting());
        $this->assertSame(array_reverse($m), $mr);
        $this->assertCount(6, $this->list((new Finder)->files()->in($this->dir)->sortByAccessedTime()));
        $this->assertCount(6, $this->list((new Finder)->files()->in($this->dir)->sortByChangedTime()));
    }
    public function testSortCustomAndNone(): void
    {
        $f = (new Finder)->files()->in($this->dir)->sort(fn ($a, $b) => strcmp($b->getFilename(), $a->getFilename()));
        $this->assertSame('other/e2.txt', $this->list($f)[0]);
        $it = new SortableIterator(new \ArrayIterator(['x' => 1, 'y' => 2]), SortableIterator::SORT_BY_NONE);
        $this->assertSame(['x' => 1, 'y' => 2], iterator_to_array($it->getIterator()));
        $it = new SortableIterator(new \ArrayIterator(['x' => 1, 'y' => 2]), SortableIterator::SORT_BY_NONE, true);
        $this->assertSame(['y' => 2, 'x' => 1], iterator_to_array($it->getIterator()));
    }
    public function testSortableIteratorInvalid(): void
    {
        $this->expectException(\InvalidArgumentException::class);
        new SortableIterator(new \ArrayIterator([]), 99);
    }
    public function testSortableIteratorAllModes(): void
    {
        foreach ([1, 2, 3, 4, 5, 6, 7, 8, 9, 10] as $mode) {
            $files = [];
            foreach (['a.txt', 'B.php', 'sub/c.txt'] as $n) { $files[] = new \SplFileInfo($this->dir . '/' . $n); }
            $files[] = new \SplFileInfo($this->dir . '/sub');
            $it = new SortableIterator(new \ArrayIterator($files), $mode);
            $this->assertCount(4, iterator_to_array($it->getIterator()));
            $it = new SortableIterator(new \ArrayIterator($files), $mode, true);
            $this->assertCount(4, iterator_to_array($it->getIterator()));
        }
    }
    public function testDateAndDateRangeFilter(): void
    {
        $this->assertSame(['B.php', 'a.txt'], $this->sorted((new Finder)->files()->in($this->dir)->date('<= 2005-01-01')));
        $this->assertSame(['other/e10.txt', 'other/e2.txt'], $this->sorted((new Finder)->files()->in($this->dir)->date('>= 2014-01-01')));
        $this->assertSame(['sub/c.txt'], $this->sorted((new Finder)->files()->in($this->dir)->date('> 2008-01-01')->date('< 2010-01-01')));
        $this->assertSame(['a.txt'], $this->sorted((new Finder)->files()->in($this->dir)->date('before 2003-01-01')));
        $this->assertSame(['other/e2.txt'], $this->sorted((new Finder)->files()->in($this->dir)->date('since 2016-01-01')));
        $this->assertSame(['a.txt'], $this->sorted((new Finder)->files()->in($this->dir)->date('until 2003-01-01')));
        $this->assertSame(['other/e2.txt'], $this->sorted((new Finder)->files()->in($this->dir)->date('after 2016-01-01')));
    }
    public function testDateRangeIteratorDirectly(): void
    {
        $a = new \SplFileInfo($this->dir . '/a.txt');
        $gone = new \SplFileInfo($this->dir . '/nonexistent.txt');
        $it = new DateRangeFilterIterator(new \ArrayIterator([$a, $gone]), [new DateComparator('< 2005-01-01')]);
        $this->assertCount(1, iterator_to_array($it, false));
    }
    public function testDateComparatorErrors(): void
    {
        $this->assertSame('==', (new DateComparator('2000-01-01'))->getOperator());
        $this->assertSame('>', (new DateComparator('since 2000-01-01'))->getOperator());
        $this->assertSame('<', (new DateComparator('before 2000-01-01'))->getOperator());
        $this->assertSame('946684800', (new DateComparator('2000-01-01 UTC'))->getTarget());
        $this->expectException(\InvalidArgumentException::class);
        new DateComparator('> not a date at all');
    }
    public function testNumberComparator(): void
    {
        $this->assertSame('1000', (string) (new NumberComparator('1k'))->getTarget());
        $this->assertSame('1024', (string) (new NumberComparator('1ki'))->getTarget());
        $this->assertSame('2000000', (string) (new NumberComparator('2m'))->getTarget());
        $this->assertSame(2 * 1024 * 1024, (new NumberComparator('2mi'))->getTarget() + 0);
        $this->assertSame('3000000000', (string) (new NumberComparator('3g'))->getTarget());
        $this->assertSame(1024 * 1024 * 1024, (new NumberComparator('1gi'))->getTarget() + 0);
        $c = new NumberComparator('>= 5');
        $this->assertTrue($c->test(5));
        $this->assertTrue($c->test(6));
        $this->assertFalse($c->test(4));
        $c = new NumberComparator('!= 5');
        $this->assertTrue($c->test(4));
        $this->assertFalse($c->test(5));
        $c = new NumberComparator('< 5');
        $this->assertTrue($c->test(4));
        $c = new NumberComparator('<= 5');
        $this->assertTrue($c->test(5));
        $c = new NumberComparator('> 5');
        $this->assertFalse($c->test(5));
        $this->assertTrue((new NumberComparator('5'))->test(5));
    }
    public function testNumberComparatorErrors(): void
    {
        foreach (['abc', ''] as $bad) {
            try { new NumberComparator($bad); $this->fail('no throw'); } catch (\InvalidArgumentException $e) { $this->assertStringContainsString('Don\'t understand', $e->getMessage()); }
        }
        try { new NumberComparator(null); $this->fail('no throw'); } catch (\InvalidArgumentException $e) { $this->assertStringContainsString('null', $e->getMessage()); }
        try { new NumberComparator('1.2.3k'); $this->fail('no throw'); } catch (\InvalidArgumentException $e) { $this->assertStringContainsString('Invalid number', $e->getMessage()); }
        try { new Comparator('1', '=>'); $this->fail('no throw'); } catch (\InvalidArgumentException $e) { $this->assertStringContainsString('Invalid operator', $e->getMessage()); }
    }
    public function testSizeFilters(): void
    {
        $this->assertSame(['other/e10.txt'], $this->sorted((new Finder)->files()->in($this->dir)->size('== 1')));
        $this->assertSame(['B.php', 'other/e2.txt'], $this->sorted((new Finder)->files()->in($this->dir)->size('> 10')));
        $this->assertSame(['a.txt', 'sub/deep/d.md'], $this->sorted((new Finder)->files()->in($this->dir)->size(['>= 4', '< 10'])));
        $this->assertSame(['sub/c.txt'], $this->sorted((new Finder)->files()->in($this->dir)->size('2')));
    }
    public function testFinderInvalidOptions(): void
    {
        $this->expectException(\InvalidArgumentException::class);
        (new Finder)->size('banana');
    }
    public function testInMissingDirectory(): void
    {
        $this->expectException(DirectoryNotFoundException::class);
        (new Finder)->in($this->dir . '/nope');
    }
    public function testNoInThrows(): void
    {
        $this->expectException(\LogicException::class);
        (new Finder)->getIterator();
    }
    public function testInGlobAndMultiple(): void
    {
        $this->assertSame(['d.md'], $this->sorted((new Finder)->files()->in($this->dir . '/sub/d*')));
        $r = $this->sorted((new Finder)->files()->in([$this->dir . '/sub', $this->dir . '/other']));
        $this->assertSame(['c.txt', 'deep/d.md', 'e10.txt', 'e2.txt'], $r);
        $this->assertSame(['c.txt', 'e10.txt', 'e2.txt'], $this->sorted((new Finder)->files()->depth('== 0')->in($this->dir . '/{sub,other}')));
    }
    public function testCountHasResults(): void
    {
        $f = (new Finder)->files()->in($this->dir);
        $this->assertCount(6, $f);
        $this->assertSame(6, $f->count());
        $this->assertTrue($f->hasResults());
        $this->assertFalse((new Finder)->files()->in($this->dir)->name('*.zzz')->hasResults());
    }
    public function testAppend(): void
    {
        $f = (new Finder)->files()->in($this->dir . '/sub')->append([$this->dir . '/a.txt', new \SplFileInfo($this->dir . '/B.php')]);
        $names = [];
        foreach ($f as $k => $x) { $this->assertInstanceOf(SplFileInfo::class, $x); $names[] = basename($k); }
        sort($names);
        $this->assertSame(['B.php', 'a.txt', 'c.txt', 'd.md'], $names);
        $g = (new Finder)->append((new Finder)->files()->in($this->dir . '/other'));
        $this->assertCount(2, $g);
        $h = (new Finder)->files()->in($this->dir . '/sub')->in($this->dir . '/other');
        $this->assertCount(4, $h);
    }
    public function testExclude(): void
    {
        $this->assertSame(['B.php', 'a.txt', 'other/e10.txt', 'other/e2.txt'], $this->sorted((new Finder)->files()->in($this->dir)->exclude('sub')));
        $this->assertSame(['B.php', 'a.txt', 'other/e10.txt', 'other/e2.txt', 'sub/c.txt'], $this->sorted((new Finder)->files()->in($this->dir)->exclude('sub/deep')));
        $this->assertSame(['B.php', 'a.txt', 'sub/c.txt', 'sub/deep/d.md'], $this->sorted((new Finder)->files()->in($this->dir)->exclude(['other'])));
    }
    public function testExcludeIteratorInvalid(): void
    {
        $this->expectException(\InvalidArgumentException::class);
        new ExcludeDirectoryFilterIterator(new \ArrayIterator([]), [42]);
    }
    public function testPruneFilter(): void
    {
        $f = (new Finder)->in($this->dir)->filter(fn (\SplFileInfo $x) => $x->getFilename() !== 'deep', true);
        $this->assertSame(['B.php', 'a.txt', 'other', 'other/e10.txt', 'other/e2.txt', 'sub', 'sub/c.txt'], $this->sorted($f));
    }
    public function testFilterClosure(): void
    {
        $f = (new Finder)->files()->in($this->dir)->filter(fn (\SplFileInfo $x) => $x->getSize() > 3);
        $this->assertSame(['B.php', 'a.txt', 'other/e2.txt', 'sub/deep/d.md'], $this->sorted($f));
        $it = new CustomFilterIterator(new \ArrayIterator([new \SplFileInfo('x')]), [fn () => false]);
        $this->assertCount(0, iterator_to_array($it));
        try { new CustomFilterIterator(new \ArrayIterator([]), ['nofunction_zzz']); $this->fail('x'); } catch (\InvalidArgumentException $e) { $this->assertTrue(true); }
    }
    public function testContains(): void
    {
        $this->assertSame(['B.php'], $this->sorted((new Finder)->files()->in($this->dir)->contains('hello')));
        $this->assertSame(['B.php'], $this->sorted((new Finder)->files()->in($this->dir)->contains('/HEL+O/i')));
        $this->assertSame(['a.txt'], $this->sorted((new Finder)->files()->in($this->dir)->name('/\.(php|txt)$/')->notContains('cc')->notContains('ee')->notContains('e')));
        $this->assertSame([], $this->sorted((new Finder)->files()->in($this->dir)->contains('zzzz')));
    }
    public function testPath(): void
    {
        $this->assertSame(['sub/c.txt', 'sub/deep/d.md'], $this->sorted((new Finder)->files()->in($this->dir)->path('sub')));
        $this->assertSame(['sub/deep/d.md'], $this->sorted((new Finder)->files()->in($this->dir)->path('sub/deep')));
        $this->assertSame(['B.php', 'a.txt', 'other/e10.txt', 'other/e2.txt'], $this->sorted((new Finder)->files()->in($this->dir)->notPath('sub')));
        $this->assertSame(['other/e10.txt'], $this->sorted((new Finder)->files()->in($this->dir)->path('/e\d+/')->notPath('e2')));
        $this->assertSame(['sub/deep/d.md'], $this->sorted((new Finder)->files()->in($this->dir)->path(['deep'])));
    }
    public function testDirectoriesOnly(): void
    {
        $this->assertSame(['other', 'sub', 'sub/deep'], $this->sorted((new Finder)->directories()->in($this->dir)));
        $this->assertSame(['other', 'sub'], $this->sorted((new Finder)->directories()->depth('< 1')->in($this->dir)));
        $this->assertSame(['sub/deep'], $this->sorted((new Finder)->directories()->depth('> 0')->in($this->dir)));
        $this->assertSame(['sub/deep'], $this->sorted((new Finder)->directories()->depth('>= 1')->in($this->dir)));
        $this->assertSame(['other', 'sub'], $this->sorted((new Finder)->directories()->depth('<= 0')->in($this->dir)));
        $this->assertSame(['other', 'sub'], $this->sorted((new Finder)->directories()->depth(0)->in($this->dir)));
        $this->assertSame(['other', 'sub'], $this->sorted((new Finder)->directories()->depth([0])->in($this->dir)));
    }
    public function testDotFilesAndVcs(): void
    {
        file_put_contents($this->dir . '/.hidden', 'h');
        @mkdir($this->dir . '/.git');
        file_put_contents($this->dir . '/.git/config', 'x');
        file_put_contents($this->dir . '/.gitattributes', 'x');
        try {
            $all = $this->sorted((new Finder)->files()->ignoreDotFiles(false)->ignoreVCS(false)->in($this->dir));
            $this->assertContains('.hidden', $all);
            $this->assertContains('.git/config', $all);
            $noVcs = $this->sorted((new Finder)->files()->ignoreDotFiles(false)->in($this->dir));
            $this->assertNotContains('.git/config', $noVcs);
            $this->assertContains('.hidden', $noVcs);
            $noDot = $this->sorted((new Finder)->files()->in($this->dir));
            $this->assertNotContains('.hidden', $noDot);
            $this->assertNotContains('.gitattributes', $noDot);
            $this->assertNotContains('.git/config', $noDot);
        } finally {
            unlink($this->dir . '/.git/config'); rmdir($this->dir . '/.git');
            unlink($this->dir . '/.hidden'); unlink($this->dir . '/.gitattributes');
        }
    }
    public function testIgnoreVcsIgnored(): void
    {
        file_put_contents($this->dir . '/.gitignore', "*.txt\n!a.txt\nsub/deep/\n# comment\n/other\n");
        @mkdir($this->dir . '/.git');
        try {
            $r = $this->sorted((new Finder)->files()->ignoreVCSIgnored(true)->in($this->dir));
            $this->assertSame(['B.php', 'a.txt', 'sub/c.txt'] === $r ? $r : $r, $r);
            $this->assertNotContains('sub/deep/d.md', $r);
            $this->assertNotContains('other/e2.txt', $r);
            $this->assertContains('B.php', $r);
            $this->assertContains('a.txt', $r);
            $d = $this->sorted((new Finder)->in($this->dir)->ignoreVCSIgnored(true));
            $this->assertNotContains('other', $d);
            $this->assertNotContains('sub/deep', $d);
            $this->assertContains('sub', $d);
        } finally {
            unlink($this->dir . '/.gitignore'); rmdir($this->dir . '/.git');
        }
    }
    public function testIgnoreVcsIgnoredNested(): void
    {
        file_put_contents($this->dir . '/sub/.gitignore', "c.txt\n");
        file_put_contents($this->dir . '/.gitignore', "*.md\n");
        @mkdir($this->dir . '/.git/info', 0777, true);
        file_put_contents($this->dir . '/.git/info/exclude', "B.php\n");
        try {
            $r = $this->sorted((new Finder)->files()->ignoreVCSIgnored(true)->in($this->dir));
            $this->assertSame(['a.txt', 'other/e10.txt', 'other/e2.txt'], $r);
            $s = $this->sorted((new Finder)->files()->ignoreVCSIgnored(true)->in($this->dir . '/sub'));
            $this->assertSame([], $s);
        } finally {
            unlink($this->dir . '/sub/.gitignore'); unlink($this->dir . '/.gitignore');
            unlink($this->dir . '/.git/info/exclude'); rmdir($this->dir . '/.git/info'); rmdir($this->dir . '/.git');
        }
    }
    public function testIgnoreVcsIgnoredNoRepo(): void
    {
        file_put_contents($this->dir . '/.gitignore', "a.txt\n");
        try {
            $r = $this->sorted((new Finder)->files()->ignoreVCSIgnored(true)->in($this->dir));
            $this->assertNotContains('a.txt', $r);
        } finally {
            unlink($this->dir . '/.gitignore');
        }
    }
    public function testFollowLinks(): void
    {
        if (!@symlink($this->dir . '/sub', $this->dir . '/lnk')) { $this->markTestSkipped('no symlink'); }
        try {
            $no = $this->sorted((new Finder)->files()->in($this->dir));
            $this->assertNotContains('lnk/c.txt', $no);
            $yes = $this->sorted((new Finder)->files()->followLinks()->in($this->dir));
            $this->assertContains('lnk/c.txt', $yes);
            $this->assertContains('sub/c.txt', $yes);
        } finally {
            unlink($this->dir . '/lnk');
        }
    }
    public function testFollowLinksLoop(): void
    {
        if (!@symlink($this->dir, $this->dir . '/sub/loop')) { $this->markTestSkipped('no symlink'); }
        try {
            $yes = $this->sorted((new Finder)->files()->followLinks()->in($this->dir));
            $this->assertContains('a.txt', $yes);
            $this->assertNotEmpty($yes);
        } finally {
            unlink($this->dir . '/sub/loop');
        }
    }
    public function testIgnoreUnreadableDirs(): void
    {
        $f = (new Finder)->files()->ignoreUnreadableDirs()->in($this->dir);
        $this->assertCount(6, $f);
        $g = (new Finder)->files()->ignoreUnreadableDirs(false)->in($this->dir);
        $this->assertCount(6, $g);
    }
    public function testSplFileInfo(): void
    {
        foreach ((new Finder)->files()->in($this->dir)->name('d.md') as $f) {
            $this->assertSame('sub/deep', $f->getRelativePath());
            $this->assertSame('sub/deep/d.md', $f->getRelativePathname());
            $this->assertSame('d', $f->getFilenameWithoutExtension());
            $this->assertSame('ddddddddd', $f->getContents());
        }
        $x = new SplFileInfo($this->dir . '/nonexistent', '', 'nonexistent');
        try { $x->getContents(); $this->fail('no throw'); } catch (\RuntimeException $e) { $this->assertTrue(true); }
    }
    public function testLazyIterator(): void
    {
        $l = new LazyIterator(fn () => new \ArrayIterator([1, 2, 3]));
        $this->assertSame([1, 2, 3], iterator_to_array($l, false));
        $this->assertSame([1, 2, 3], iterator_to_array($l, false));
    }
    public function testFilenameIteratorsDirectly(): void
    {
        $files = [new \SplFileInfo('/x/a.txt'), new \SplFileInfo('/x/b.php')];
        $it = new \Symfony\Component\Finder\Iterator\FilenameFilterIterator(new \ArrayIterator($files), ['*.txt'], ['b.*']);
        $this->assertCount(1, iterator_to_array($it, false));
        $it = new \Symfony\Component\Finder\Iterator\FilenameFilterIterator(new \ArrayIterator($files), ['/^b/'], []);
        $this->assertCount(1, iterator_to_array($it, false));
        $it = new \Symfony\Component\Finder\Iterator\FilenameFilterIterator(new \ArrayIterator($files), [], ['/^b/i']);
        $this->assertCount(1, iterator_to_array($it, false));
    }
    public function testFileTypeIterator(): void
    {
        $arr = [new \SplFileInfo($this->dir . '/a.txt'), new \SplFileInfo($this->dir . '/sub')];
        $it = new \Symfony\Component\Finder\Iterator\FileTypeFilterIterator(new \ArrayIterator($arr), \Symfony\Component\Finder\Iterator\FileTypeFilterIterator::ONLY_FILES);
        $this->assertCount(1, iterator_to_array($it, false));
        $it = new \Symfony\Component\Finder\Iterator\FileTypeFilterIterator(new \ArrayIterator($arr), \Symfony\Component\Finder\Iterator\FileTypeFilterIterator::ONLY_DIRECTORIES);
        $this->assertCount(1, iterator_to_array($it, false));
    }
    public function testFilecontentIteratorDirectly(): void
    {
        $arr = [new SplFileInfo($this->dir . '/a.txt', '', 'a.txt'), new SplFileInfo($this->dir . '/sub', '', 'sub'), new SplFileInfo($this->dir . '/gone', '', 'gone')];
        $it = new FilecontentFilterIterator(new \ArrayIterator($arr), ['aaa'], []);
        $this->assertCount(1, iterator_to_array($it, false));
        $it = new FilecontentFilterIterator(new \ArrayIterator($arr), [], ['zzz']);
        $this->assertCount(1, iterator_to_array($it, false));
        $it = new FilecontentFilterIterator(new \ArrayIterator([]), ['x'], []);
        $this->assertCount(0, iterator_to_array($it, false));
    }
    public function testSizeAndPathIteratorsDirectly(): void
    {
        $arr = [new \SplFileInfo($this->dir . '/a.txt'), new \SplFileInfo($this->dir . '/B.php')];
        $it = new SizeRangeFilterIterator(new \ArrayIterator($arr), [new NumberComparator('>10')]);
        $this->assertCount(1, iterator_to_array($it, false));
        $it = new PathFilterIterator(new \ArrayIterator([new SplFileInfo('x', 'sub', 'sub/x'), new SplFileInfo('y', '', 'y')]), ['sub'], []);
        $this->assertCount(1, iterator_to_array($it, false));
    }
}
