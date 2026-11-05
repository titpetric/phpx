name: the string functions the core binding left out
description: >
  strtok, similar_text, str_increment, stristr, chunk_split, substr_compare, str_word_count, str_ireplace, the html entity functions with their flags, sscanf, soundex, quoted-printable, array_multisort and the printf family over streams. Expected output is php 8.4's.
runner:
  flatstack: false
---
<?php
$t = strtok("a,,b;c d", ",; "); $r = []; while ($t !== false) { $r[] = $t; $t = strtok(",; "); } echo implode('|', $r), "\n";
var_dump(strtok("/x/y", "/"), strtok("/"), strtok("/"));
echo similar_text("World", "Word", $p), ' ', round($p, 2), ' ', similar_text("Hello", "World"), "\n";
echo str_increment("a"), str_increment("Az"), str_increment("zz"), str_increment("a9"), ' ', str_decrement("b"), str_decrement("aa"), str_decrement("Ba"), str_decrement("10"), "\n";
echo str_rot13("Hello, World!"), ' ', stristr("HayStack", "st"), ' ', stristr("HayStack", "st", true), ' ', strrchr("a/b/c", "/"), ' ', var_export(strrchr("abc", "x"), true), "\n";
echo chunk_split("abcdefg", 3, "-"), ' ', quotemeta("1+1=2?"), ' ', count_chars("abca", 3), ' ', json_encode(count_chars("aab", 1)), "\n";
echo substr_compare("Hello", "ello", 1), substr_compare("Hello", "ELLO", 1, 4, true), substr_compare("abcde", "cd", -3, 2), "\n";
echo str_word_count("Hello fri3nd, you're looking good today!"), ' ', json_encode(str_word_count("Hello World again", 1)), ' ', json_encode(str_word_count("Hello World", 2)), "\n";
echo str_ireplace("WORLD", "There", "Hello world, WoRlD", $c), $c, ' ', json_encode(str_ireplace(['A', 'b'], ['1', '2'], ['aAbB', 'xyz'])), "\n";
echo htmlspecialchars("<a href=\"x\">T'est</a> &amp;"), ' ', htmlspecialchars("<'\">", ENT_NOQUOTES), ' ', htmlspecialchars("&amp; &", ENT_QUOTES, 'UTF-8', false), "\n";
echo htmlspecialchars_decode("&lt;b&gt; &quot;q&quot; &#039;s&#039; &amp;amp;"), ' ', htmlspecialchars_decode("&quot;&#039;", ENT_NOQUOTES), ' ', html_entity_decode("caf&eacute; &euro; &#8364; &hellip; &nbsp;|"), ' ', htmlentities("café <b> €"), "\n";
[$a, $b] = sscanf("age: 25 name: Bob", "age: %d name: %s"); echo $a, $b, ' ', sscanf("12:34", "%d:%d", $h, $m), $h, $m, ' ', json_encode(sscanf("id-7 x", "id-%d %c")), "\n";
echo soundex("Robert"), soundex("Tymczak"), ' ', quoted_printable_encode("héllo=wörld"), ' ', quoted_printable_decode("h=C3=A9llo=3D"), "\n";
$data = [3, 1, 2]; $names = ['c', 'a', 'b']; array_multisort($data, $names); echo implode($data), implode($names), ' ';
$d1 = [3, 1, 3, 2]; $d2 = ['x', 'y', 'a', 'z']; array_multisort($d1, SORT_DESC, $d2, SORT_ASC); echo implode($d1), implode($d2), "\n";
ob_start(); vprintf("%s-%03d\n", ['x', 7]); $o = ob_get_clean(); echo $o; $h = fopen('php://memory', 'w+'); fprintf($h, "%05.1f|%s", 3.14159, 'z'); vfprintf($h, "[%d]", [5]); rewind($h); echo stream_get_contents($h), "\n";
---
a|b|c|d
string(1) "x"
string(1) "y"
bool(false)
4 88.89 1
bBaaaab0 azAz9
Uryyb, Jbeyq! Stack Hay /c false
abc-def-g- 1\+1=2\? abc {"97":2,"98":1}
000
7 ["Hello","World","again"] {"0":"Hello","6":"World"}
Hello There, There2 ["1122","xyz"]
&lt;a href=&quot;x&quot;&gt;T&#039;est&lt;/a&gt; &amp;amp; &lt;'"&gt; &amp; &amp;
<b> "q" 's' &amp; &quot;&#039; café € € …  | caf&eacute; &lt;b&gt; &euro;
25Bob 21234 [7,"x"]
R163T522 h=C3=A9llo=3Dw=C3=B6rld héllo=
123abc 3321axzy
x-007
003.1|z[5]
