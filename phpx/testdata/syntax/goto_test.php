<?php
function retry(int $times, callable $f) {
    $attempts = 0;
    beginning:
    $attempts++;
    try {
        return $f($attempts);
    } catch (RuntimeException $e) {
        if ($attempts < $times) {
            goto beginning;
        }
        throw $e;
    }
}
echo retry(3, function ($n) { if ($n < 3) { throw new RuntimeException("fail $n"); } return "ok after $n"; }), "\n";

function search(array $grid, $needle) {
    foreach ($grid as $y => $row) {
        foreach ($row as $x => $cell) {
            if ($cell === $needle) {
                goto found;
            }
        }
    }
    return "none";
    found:
    return "found at $x,$y";
}
echo search([[1, 2], [3, 4]], 3), " ", search([[1]], 9), "\n";

$i = 0;
loop:
$i++;
if ($i < 5) { goto loop; }
echo "i=$i\n";

function skip() {
    echo "a ";
    goto end;
    echo "never ";
    end:
    echo "b\n";
}
skip();

function inSwitch($v) {
    switch ($v) {
        case 1: goto one;
        case 2: echo "two "; break;
    }
    echo "after-switch ";
    goto done;
    one:
    echo "one ";
    done:
    echo "done\n";
}
inSwitch(1); inSwitch(2);

for ($a = 0; $a < 3; $a++) {
    for ($b = 0; $b < 3; $b++) {
        if ($a * $b == 2) goto out;
    }
}
out:
echo "a=$a b=$b\n";
$parser = ['x' => 1];
function lex(string $s) {
    $pos = 0; $n = 0;
    retry:
    if ($pos < strlen($s)) {
        $pos += 2; $n++;
        goto retry;
    }
    return $n;
}
echo lex("abcdefg"), "\n";
