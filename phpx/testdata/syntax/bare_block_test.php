<?php
$t = 2;
switch (true) {
    case $t === 1: {
        echo "one\n";
        break;
    }
    case $t === 2: {
        echo "two\n";
        break;
    }
    default: {
        echo "other\n";
    }
}
{
    $inner = 5;
    echo $inner, "\n";
}
echo $inner + 1, "\n";
foreach ([1, 2] as $v) {
    {
        if ($v === 1) { continue; }
        echo "v=$v\n";
    }
}
