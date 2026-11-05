<?php
use PHPUnit\Framework\TestCase;
use Symfony\Component\Finder\{Glob, Gitignore};

final class FinderCov2Test extends TestCase
{
    private function g(string $glob, string $s, array $opts = []): bool
    {
        return (bool) preg_match(Glob::toRegex($glob, ...$opts), $s);
    }
    public function testGlobBasics(): void
    {
        $this->assertTrue($this->g('*.php', 'a.php'));
        $this->assertFalse($this->g('*.php', 'dir/a.php'));
        $this->assertFalse($this->g('*.php', '.hidden.php'));
        $this->assertTrue($this->g('*.php', '.hidden.php', [false]));
        $this->assertTrue($this->g('a?c', 'abc'));
        $this->assertFalse($this->g('a?c', 'a/c'));
        $this->assertTrue($this->g('a?c', 'a/c', [true, false]));
        $this->assertTrue($this->g('*.php', 'dir/a.php', [true, false]));
        $this->assertTrue($this->g('a.b', 'a.b'));
        $this->assertFalse($this->g('a.b', 'axb'));
        $this->assertTrue($this->g('a(b)+|^$', 'a(b)+|^$'));
    }
    public function testGlobBraces(): void
    {
        $this->assertTrue($this->g('*.{php,txt}', 'x.txt'));
        $this->assertTrue($this->g('*.{php,txt}', 'x.php'));
        $this->assertFalse($this->g('*.{php,txt}', 'x.md'));
        $this->assertTrue($this->g('{a,b{c,d}}', 'bd'));
        $this->assertTrue($this->g('a{b', 'a{b'));
        $this->assertTrue($this->g('a{b,c', 'a{b,c'));
        $this->assertTrue($this->g('a,b', 'a,b'));
        $this->assertTrue($this->g('a\\{b', 'a{b'));
        $this->assertTrue($this->g('a\\,b', 'a,b'));
    }
    public function testGlobEscapes(): void
    {
        $this->assertTrue($this->g('a\\*b', 'a*b'));
        $this->assertFalse($this->g('a\\*b', 'axb'));
        $this->assertTrue($this->g('a\\?b', 'a?b'));
        $this->assertTrue($this->g('a\\\\b', 'a\\b'));
        $this->assertTrue($this->g('a\\[b', 'a[b'));
        $this->assertTrue($this->g('a[b\\]c]d', 'a]d'));
    }
    public function testGlobBrackets(): void
    {
        $this->assertTrue($this->g('[abc].txt', 'b.txt'));
        $this->assertFalse($this->g('[abc].txt', 'd.txt'));
        $this->assertTrue($this->g('[a-c].txt', 'b.txt'));
        $this->assertFalse($this->g('[a-c].txt', 'z.txt'));
        $this->assertTrue($this->g('[[:digit:]].txt', '5.txt'));
        $this->assertFalse($this->g('[[:digit:]].txt', 'x.txt'));
        $this->assertTrue($this->g('a[b', 'a[b'));
        $this->assertTrue($this->g('[]a].txt', ']'.'.txt'));
        $this->assertTrue($this->g('[a-]x', '-x'));
    }
    public function testGlobGlobstar(): void
    {
        $this->assertTrue($this->g('**/*.php', 'a.php'));
        $this->assertTrue($this->g('**/*.php', 'a/b/c.php'));
        $this->assertFalse($this->g('**/*.php', 'a/.b/c.php'));
        $this->assertTrue($this->g('src/**', 'src/a/b'));
        $this->assertTrue($this->g('src/**/x', 'src/x'));
        $this->assertTrue($this->g('src/**/x', 'src/a/b/x'));
        $this->assertTrue($this->g('**/*.php', 'a/.b/c.php', [false]));
        $this->assertTrue($this->g('/etc/*', '/etc/passwd'));
        $this->assertTrue($this->g('**', 'abc'));
        $this->assertTrue($this->g('a**b', 'aXb'));
    }
    public function testGlobDelimiterAndUnanchored(): void
    {
        $this->assertSame('/^(?=[^\.])ab[^/]*$/', Glob::toRegex('ab*', true, true, '/'));
        $this->assertSame('/^(?=[^\.])a\/(?=[^\.])b$/', Glob::toRegex('a/b', true, true, '/'));
        $this->assertSame(1, preg_match(Glob::toRegex('a#b'), 'a#b'));
        $this->assertSame(1, preg_match(Glob::toRegex('*.*.*'), 'a.b.c'));
        $this->assertSame(1, preg_match(Glob::toRegex('a*', false, false), 'a/b/c'));
        $this->assertSame(1, preg_match(Glob::toRegex('a**', false, false), 'a/b/c'));
        $this->assertSame(1, preg_match(Glob::toRegex('a**b', false, false), 'a/x/b'));
    }

    private function ignored(string $content, string $path): bool
    {
        return (bool) preg_match(Gitignore::toRegex($content), $path);
    }
    public function testGitignoreToRegex(): void
    {
        $c = "*.log\n!keep.log\nbuild/\n/root.txt\n# comment\n\ndocs/**/tmp\n\\#hash\nspace\\ \ntrail   \n";
        $this->assertTrue($this->ignored($c, 'a.log'));
        $this->assertFalse($this->ignored($c, 'keep.log'));
        $this->assertTrue($this->ignored($c, 'x/y/a.log'));
        $this->assertTrue($this->ignored($c, 'build/'));
        $this->assertTrue($this->ignored($c, 'build/x.o'));
        $this->assertTrue($this->ignored($c, 'root.txt'));
        $this->assertFalse($this->ignored($c, 'sub/root.txt'));
        $this->assertTrue($this->ignored($c, 'docs/a/b/tmp'));
        $this->assertTrue($this->ignored($c, 'docs/tmp'));
        $this->assertTrue($this->ignored($c, '#hash'));
        $this->assertTrue($this->ignored($c, 'trail'));
        $this->assertFalse($this->ignored($c, 'readme.md'));
        $this->assertFalse($this->ignored('', 'anything'));
    }
    public function testGitignoreClassesAndEscapes(): void
    {
        $this->assertTrue($this->ignored("[ab].txt\n", 'a.txt'));
        $this->assertFalse($this->ignored("[ab].txt\n", 'c.txt'));
        $this->assertTrue($this->ignored("[!ab].txt\n", 'c.txt'));
        $this->assertTrue($this->ignored("[^ab].txt\n", 'c.txt'));
        $this->assertTrue($this->ignored("[a-c].txt\n", 'b.txt'));
        $this->assertTrue($this->ignored("a?c\n", 'abc'));
        $this->assertFalse($this->ignored("a?c\n", 'a/c'));
        $this->assertTrue($this->ignored("\\*star\n", '*star'));
        $this->assertTrue($this->ignored("\\!bang\n", '!bang'));
        $this->assertTrue($this->ignored("a.b\n", 'a.b'));
        $this->assertFalse($this->ignored("a.b\n", 'axb'));
        $this->assertTrue($this->ignored("**/foo\n", 'a/b/foo'));
        $this->assertTrue($this->ignored("abc/**\n", 'abc/d/e'));
        $this->assertTrue($this->ignored("a/**/b\n", 'a/b'));
        $this->assertTrue($this->ignored("a/b\n", 'a/b'));
        $this->assertFalse($this->ignored("a/b\n", 'x/a/b'));
        $this->assertTrue($this->ignored("[a\n", '[a'));
        $this->assertTrue($this->ignored("*.txt\r\n!b.txt\r\n", 'a.txt'));
        $this->assertFalse($this->ignored("*.txt\r\n!b.txt\r\n", 'b.txt'));
    }
    public function testGitignoreNegatedRegex(): void
    {
        $re = Gitignore::toRegexMatchingNegatedPatterns("*.log\n!keep.log\n");
        $this->assertSame(1, preg_match($re, 'keep.log'));
        $this->assertSame(0, preg_match($re, 'other.log'));
    }
}
