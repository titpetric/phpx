<?php
use Doctrine\Inflector\{InflectorFactory, Inflector, Language, RulesetInflector};
use Doctrine\Inflector\Rules\{Pattern, Patterns, Ruleset, Substitution, Substitutions, Transformation, Transformations, Word};
use Doctrine\Inflector\Rules\English;
use PHPUnit\Framework\TestCase;

final class InflectorMoreTest extends TestCase
{
    private Inflector $i;
    protected function setUp(): void { $this->i = InflectorFactory::create()->build(); }
    public function testPluralize(): void
    {
        foreach (['user' => 'users', 'person' => 'people', 'box' => 'boxes', 'city' => 'cities', 'mouse' => 'mice', 'sheep' => 'sheep', 'child' => 'children', 'analysis' => 'analyses', 'quiz' => 'quizzes', 'ox' => 'oxen'] as $s => $p) {
            $this->assertSame($p, $this->i->pluralize($s), $s);
            $this->assertSame($s, $this->i->singularize($p), $p);
        }
    }
    public function testCases(): void
    {
        $this->assertSame('userProfileId', $this->i->camelize('user_profile_id'));
        $this->assertSame('UserProfiles', $this->i->classify('user_profiles'));
        $this->assertSame('user_profile_id', $this->i->tableize('UserProfileId'));
        $this->assertSame('Hello World Again', $this->i->capitalize('hello world again'));
        $this->assertSame('hello-world', $this->i->urlize('Hello World!'));
        $this->assertSame('a-b-c', $this->i->urlize('a b c'));
    }
    public function testUnaccent(): void
    {
        $this->assertSame('Crepe a la creme', $this->i->unaccent('Crêpe à la crème'));
        $this->assertSame('ueber-strasse', Doctrine\Inflector\Inflector::class === get_class($this->i) ? $this->i->urlize('Über Straße') : '');
    }
    public function testCustomRules(): void
    {
        $inf = InflectorFactory::create()->withPluralRules(new Ruleset(new Transformations(new Transformation(new Pattern('/^(bot)$/i'), '\1s')), new Patterns(new Pattern('data')), new Substitutions(new Substitution(new Word('foo'), new Word('fooz')))))->build();
        $this->assertSame('fooz', $inf->pluralize('foo'));
        $this->assertSame('data', $inf->pluralize('data'));
        $this->assertSame('bots', $inf->pluralize('bot'));
    }
    public function testLanguages(): void
    {
        $f = InflectorFactory::createForLanguage(Language::FRENCH)->build();
        $this->assertSame('chevaux', $f->pluralize('cheval'));
        $s = InflectorFactory::createForLanguage(Language::SPANISH)->build();
        $this->assertSame('lápices', $s->pluralize('lápiz'));
        $n = InflectorFactory::createForLanguage(Language::NORWEGIAN_BOKMAL)->build();
        $this->assertSame('hester', $n->pluralize('hest'));
        $t = InflectorFactory::createForLanguage(Language::TURKISH)->build();
        $this->assertSame('kitaplar', $t->pluralize('kitap'));
    }
    public function testWordsAndEdges(): void
    {
        $this->assertSame('', $this->i->camelize(''));
        $this->assertSame('x', $this->i->singularize('x'));
        $this->assertSame('already_snake', $this->i->tableize('already_snake'));
        $this->assertSame('Iñtërnâtiônàlizætiøn', $this->i->capitalize('iñtërnâtiônàlizætiøn'));
        $this->assertSame('foo bar', $this->i->capitalize('foo bar', " ") === 'Foo Bar' ? 'foo bar' : 'x');
    }
    public function testWordObjects(): void
    {
        $w = new Word('abc');
        $this->assertSame('abc', $w->getWord());
        $p = new Pattern('a.c');
        $this->assertTrue($p->matches('abc'));
        $this->assertSame('/a.c/i', $p->getRegex());
        $t = new Transformation(new Pattern('(\w+)y'), '\1ies');
        $this->assertSame('cities', $t->inflect('city'));
    }
    public function testEnglishRules(): void
    {
        $this->assertInstanceOf(Ruleset::class, English\Rules::getPluralRuleset());
        $this->assertNotEmpty(iterator_to_array(English\Inflectible::getPlural()));
        $this->assertNotEmpty(iterator_to_array(English\Uninflected::getSingular()));
    }
}
