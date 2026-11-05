<?php
foreach ([1, 2] as $i) {
    foreach ([1, 2, 3] as $j) {
        if ($j === 2) { continue 2; }
        echo "$i$j ";
    }
    echo "never\n";
}
echo "\n";
for ($i = 0; $i < 3; $i++) {
    for ($j = 0; $j < 3; $j++) {
        if ($i === 1) { break 2; }
        echo "$i$j ";
    }
}
echo "\n";
$n = 0;
while (true) {
    while (true) {
        do { $n++; if ($n > 2) { break 3; } continue 2; } while (false);
    }
}
echo $n, "\n";
foreach ([1, 2, 3] as $x) {
    switch ($x) {
        case 2: continue 2;
        case 3: break 2;
        default: echo "d$x ";
    }
    echo "after$x ";
}
echo "\n";
