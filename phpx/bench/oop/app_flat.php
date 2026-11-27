<?php

// app_oop.php without a class: the same loop, the same formulas, the same
// output lines. Every binary here runs it, so it separates the fork's OOP
// work from the interpreter both trees share.

function circle_area(float $r): float
{
    return 3.0 * $r * $r;
}

function rect_area(float $w, float $h): float
{
    return $w * $h;
}

$total = 0.0;
$made = 0;
for ($i = 0; $i < 20000; $i++) {
    if ($i % 2 === 0) {
        $scaled = circle_area(1.0 + ($i % 7)) * 2.0;
        $label = 'Circle:circle';
    } else {
        $scaled = rect_area(2.0, 3.0 + ($i % 5)) * 1.5;
        $label = 'Rect:rect';
    }
    $made++;
    $total += $scaled;
    if ($i % 5000 === 0) {
        echo $label, "\n";
    }
}
echo "total=", (int) $total, " made=", $made, "\n";
