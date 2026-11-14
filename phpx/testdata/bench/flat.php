<?php
// phpscript style: namespaced bags of properties and methods, no inheritance.
class Circle { public float $r; function __construct(float $r) { $this->r = $r; } function area(): float { return 3.0 * $this->r * $this->r; } function name(): string { return 'circle'; } }
class Rect { public float $w; public float $h; function __construct(float $w, float $h) { $this->w = $w; $this->h = $h; } function area(): float { return $this->w * $this->h; } function name(): string { return 'rect'; } }
$sum = 0.0; $names = 0;
for ($i = 1; $i <= 200000; $i++) {
    $s = $i % 2 ? new Circle($i % 7 + 1) : new Rect($i % 5 + 1, $i % 3 + 1);
    $sum += $s->area();
    $names += strlen($s->name());
}
echo round($sum, 2), " ", $names, "\n";
