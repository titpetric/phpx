<?php
namespace App\Tests;

use Attribute;
use ReflectionClass;
use ReflectionMethod;

#[Attribute(Attribute::TARGET_CLASS | Attribute::TARGET_METHOD)]
final class Route
{
    public function __construct(public string $path = '/', public array $methods = ['GET']) {}
}

#[Attribute]
final class Tag
{
    public function __construct(public string $name) {}
}

/**
 * A documented controller.
 */
#[Route('/home', methods: ['GET', 'POST']), Tag('web')]
final class Home
{
    /** The index action. */
    #[Route('/index')]
    #[Tag(name: 'idx')]
    public function index(): void {}

    public function plain(): void {}

    #[Tag('prop')]
    public int $count = 0;
}

$rc = new ReflectionClass(Home::class);
foreach ($rc->getAttributes() as $a) {
    echo $a->getName(), ' ', json_encode($a->getArguments(), JSON_UNESCAPED_SLASHES), "\n";
}
echo count($rc->getAttributes(Route::class)), count($rc->getAttributes(Tag::class)), count($rc->getAttributes('Nope')), "\n";
$route = $rc->getAttributes(Route::class)[0]->newInstance();
echo $route->path, ' ', implode(',', $route->methods), "\n";
$m = new ReflectionMethod(Home::class, 'index');
echo count($m->getAttributes()), ' ', $m->getAttributes()[0]->newInstance()->path, "\n";
echo trim($m->getDocComment()), "\n";
var_dump((new ReflectionMethod(Home::class, 'plain'))->getDocComment());
echo count((new ReflectionMethod(Home::class, 'plain'))->getAttributes()), "\n";
echo $rc->getProperty('count')->getAttributes()[0]->getName(), "\n";
echo preg_replace('/\s+/', ' ', trim(trim($rc->getDocComment(), '/*'))), "\n";
