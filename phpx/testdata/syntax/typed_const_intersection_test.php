<?php

interface Named { public function name(): string; }
interface Tagged { public function tag(): string; }

class Both implements Named, Tagged {
    public const string KIND = 'both';
    public const array LIST = ['a' => 1, 'b' => 2];
    final public const int|string MIXED = 7;
    protected const ?string NOTHING = null;
    const LEGACY = 'old', OTHER = 'two';
    public function name(): string { return 'n'; }
    public function tag(): string { return 't'; }
    public static function nothing(): ?string { return static::NOTHING; }
}

function both(Named&Tagged $x, ?Both $y = null, Named&Tagged ...$rest): string {
    return $x->name() . $x->tag() . count($rest);
}

function dnf((Named&Tagged)|null $x): string {
    return $x === null ? 'null' : $x->name();
}

function byref(Named &$x, array &$arr = []): void { $arr[] = 1; }

$b = new Both();
echo Both::KIND, ' ', Both::LIST['b'], ' ', Both::MIXED, ' ', Both::LEGACY, Both::OTHER, "\n";
var_dump(Both::nothing());
echo both($b, null, $b, $b), ' ', dnf($b), ' ', dnf(null), "\n";
$a = [];
byref($b, $a);
echo count($a), "\n";

try { throw new LogicException('x'); } catch (RuntimeException | LogicException) { echo "caught\n"; }
try { throw new LogicException('x'); } catch (LogicException $e) { echo get_class($e), "\n"; }
