<?php
// phpx style: abstract base, inheritance, trait, interface, parent::, static::.
interface Shape { function area(): float; }
trait Named { function name(): string { return static::LABEL; } }
abstract class Base implements Shape { use Named; const LABEL = 'base'; }
class Circle extends Base { const LABEL = 'circle'; public float $r; function __construct(float $r) { $this->r = $r; } function area(): float { return 3.0 * $this->r * $this->r; } }
class Rect extends Base { const LABEL = 'rect'; public float $w; public float $h; function __construct(float $w, float $h) { $this->w = $w; $this->h = $h; } function area(): float { return $this->w * $this->h; } }
class Square extends Rect { const LABEL = 'square'; function __construct(float $s) { parent::__construct($s, $s); } }
$sum = 0.0; $names = 0;
for ($i = 1; $i <= 200000; $i++) {
    $s = $i % 2 ? new Circle($i % 7 + 1) : new Rect($i % 5 + 1, $i % 3 + 1);
    $sum += $s->area();
    $names += strlen($s->name());
}
echo round($sum, 2), " ", $names, "\n";
