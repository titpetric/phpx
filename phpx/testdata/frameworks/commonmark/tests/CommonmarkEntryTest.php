<?php
use PHPUnit\Framework\TestCase;
use League\CommonMark\CommonMarkConverter;
use League\CommonMark\GithubFlavoredMarkdownConverter;

final class CommonmarkEntryTest extends TestCase
{
    public function testHeadingAndEmphasis(): void
    {
        $c = new CommonMarkConverter();
        $this->assertSame("<h1>Title</h1>\n<p>a <em>b</em> <strong>c</strong></p>\n", (string) $c->convert("# Title\n\na *b* **c**"));
    }
    public function testListsAndCode(): void
    {
        $c = new CommonMarkConverter();
        $this->assertSame("<ul>\n<li>a</li>\n<li>b</li>\n</ul>\n", (string) $c->convert("- a\n- b"));
        $this->assertSame("<pre><code class=\"language-php\">echo 1;\n</code></pre>\n", (string) $c->convert("```php\necho 1;\n```"));
    }
    public function testLinksAndSafety(): void
    {
        $c = new CommonMarkConverter(['html_input' => 'strip', 'allow_unsafe_links' => false]);
        $this->assertSame("<p><a href=\"https://x.test\">x</a></p>\n", (string) $c->convert('[x](https://x.test)'));
        $this->assertSame("<p><a>bad</a></p>\n", (string) $c->convert('[bad](javascript:alert(1))'));
    }
    public function testGithubFlavoured(): void
    {
        $c = new GithubFlavoredMarkdownConverter();
        $this->assertStringContainsString('<del>gone</del>', (string) $c->convert('~~gone~~'));
        $this->assertStringContainsString('<table>', (string) $c->convert("| a | b |\n|---|---|\n| 1 | 2 |"));
    }
}
