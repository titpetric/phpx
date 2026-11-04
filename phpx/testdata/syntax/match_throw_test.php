<?php
function kind($x) {
    return match (true) {
        $x < 0 => 'negative',
        $x === 0 => 'zero',
        $x < 10, $x < 100 => 'small',
        default => 'large',
    };
}
foreach ([-5, 0, 7, 50, 500] as $n) { echo kind($n), ' '; }
echo "\n";
echo match ('b') { 'a' => 1, 'b', 'c' => 2, default => 3 }, "\n";
$v = 3;
echo match ($v) { 1, 2 => 'low', 3, 4 => 'mid', }, "\n";
try {
    echo match ($v) { 1 => 'one' };
} catch (\UnhandledMatchError $e) {
    echo get_class($e), ': ', $e->getMessage(), "\n";
}
try {
    echo match ('zz') { 'a' => 1 };
} catch (\Error $e) {
    echo $e->getMessage(), "\n";
}
$calls = 0;
function subject() { global $calls; $calls++; return 2; }
echo match (subject()) { 1 => 'a', 2 => 'b', 3 => 'c' }, $calls, "\n";
class Status {
    public function label(int $s): string {
        return match ($s) { 200 => 'ok', 404 => 'missing', default => throw new \RuntimeException("bad status $s") };
    }
}
echo (new Status)->label(404), "\n";
try { (new Status)->label(500); } catch (\RuntimeException $e) { echo $e->getMessage(), "\n"; }
$name = null;
try { $x = $name ?? throw new \LogicException('no name'); } catch (\LogicException $e) { echo $e->getMessage(), "\n"; }
$f = fn($a) => $a > 0 ? $a : throw new \DomainException('non-positive');
echo $f(3), "\n";
try { $f(-1); } catch (\DomainException $e) { echo get_class($e), "\n"; }
echo match (1) { default => 'only default' }, "\n";
