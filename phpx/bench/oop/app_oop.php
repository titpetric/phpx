<?php

// An application written in the OOP baseline the fork reverses phpscript's
// decisions for: interfaces, traits, abstract classes, extends, visibility,
// static storage, readonly promotion, __toString and late static binding.
// Nothing here reaches for an extension, so a difference is the OOP baseline
// and not a binding.
//
// app_flat.php is the same workload written without a class, so both
// binaries run it. The two print the same lines, which makes the flat app the
// cost control and the OOP app the feature measurement.

interface Shape
{
    public function area(): float;
}

trait Named
{
    protected string $name = 'shape';

    public function name(): string
    {
        return $this->name;
    }
}

abstract class Base implements Shape
{
    use Named;

    protected static int $made = 0;

    public function __construct(protected readonly float $scale = 1.0)
    {
        self::$made++;
    }

    abstract public function area(): float;

    public function scaled(): float
    {
        return $this->area() * $this->scale;
    }

    public function __toString(): string
    {
        return static::class . ':' . $this->name();
    }

    public static function made(): int
    {
        return self::$made;
    }
}

final class Circle extends Base
{
    public function __construct(private float $r = 1.0, float $scale = 1.0)
    {
        parent::__construct($scale);
        $this->name = 'circle';
    }

    public function area(): float
    {
        return 3.0 * $this->r * $this->r;
    }
}

final class Rect extends Base
{
    public function __construct(private float $w = 1.0, private float $h = 1.0, float $scale = 1.0)
    {
        parent::__construct($scale);
        $this->name = 'rect';
    }

    public function area(): float
    {
        return $this->w * $this->h;
    }
}

$total = 0.0;
for ($i = 0; $i < 20000; $i++) {
    if ($i % 2 === 0) {
        $shape = new Circle(1.0 + ($i % 7), 2.0);
    } else {
        $shape = new Rect(2.0, 3.0 + ($i % 5), 1.5);
    }
    $total += $shape->scaled();
    if ($i % 5000 === 0) {
        echo $shape, "\n";
    }
}
echo "total=", (int) $total, " made=", Base::made(), "\n";
