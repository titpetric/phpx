<?php
$cases = [["The quick brown fox", 10, "\n", false], ["A very long woooooooooooord.", 8, "\n", true], ["A very long woooooooooooord.", 8, "\n", false], ["short", 10, "<br>", false], ["a b c d e f", 3, "\n", false], ["line one\nline two is longer", 10, "\n", false], ["", 5, "\n", false], ["  lead  trail  ", 6, "\n", true]];
foreach ($cases as [$s, $w, $b, $c]) { echo json_encode(wordwrap($s, $w, $b, $c)), "\n"; }
echo json_encode(array_replace_recursive(['a' => [1, 2, [5]], 'b' => 'x'], ['a' => [1 => 'y', 2 => [1 => 6]], 'c' => 1], ['b' => ['z']])), "\n";
echo json_encode([grapheme_strlen("a\u{0301}b"), grapheme_substr("a\u{0301}bc", 1, 1), grapheme_strpos("héllo", "l")]), "\n";
echo json_encode([explode(',', 'a,b,c', 2), explode(',', 'a,b,c', 0), explode(',', 'a,b,c', -1), explode(',', 'a,b,c', -5), explode(',', '')]), "\n";
preg_match_all('/(a)(b)?/', 'ab a', $m, PREG_SET_ORDER);
foreach ($m as $r) { array_shift($r); echo count($r), ","; }
echo count($m[0]), "\n";
$a = fopen('php://memory', 'w+'); fwrite($a, 'hello world'); rewind($a); $b = fopen('php://memory', 'w+');
var_dump(stream_copy_to_stream($a, $b, 5)); rewind($b); var_dump(stream_get_contents($b));
echo get_debug_type(new RuntimeException), get_debug_type(new ArrayIterator([])), "\n";
