<?php
abstract class Base {
    protected static string $cls;
    protected static int $count = 0;
    public static function avail(): bool { return class_exists(static::$cls); }
    public static function make() { static::$count++; return new static::$cls(); }
    public static function count(): int { return static::$count; }
    public static function viaMember(string $kind) { return static::{"build" . $kind}(); }
    public static function buildA() { return 'A from ' . static::class; }
    abstract public function &items(): array;
}
class Impl extends Base {
    protected static string $cls = ArrayObject::class;
    private array $list = [1, 2];
    public function &items(): array { return $this->list; }
}
class Other extends Base {
    protected static string $cls = 'NoSuchClass';
    protected static int $count = 100;
    public function &items(): array { static $e = []; return $e; }
}
foreach ([Other::class, Impl::class] as $f) { var_dump($f::avail()); }
echo get_class(Impl::make()), Impl::count(), Base::count(), Other::count(), "\n";
echo Impl::viaMember('A'), "\n";
echo count((new Impl)->items()), "\n";
echo (1 / 4 + 1) * 10 + .1, " ", .5e1 + .25, " ", 'x' . 1, "\n";
