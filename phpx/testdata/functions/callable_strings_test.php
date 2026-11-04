<?php
echo implode(',', array_map('\trim', [' a ', 'b '])), "\n";
echo implode(',', array_map('\strlen', ['ab', 'c'])), "\n";
echo implode(',', array_map('trim', [' a '])), "\n";
class K { public static function up($x) { return strtoupper($x); } public function lo($x) { return strtolower($x); } }
echo implode(',', array_map('K::up', ['a'])), implode(',', array_map(['K', 'up'], ['b'])), implode(',', array_map([new K, 'lo'], ['C'])), "\n";
echo implode(',', array_map(K::up(...), ['d'])), "\n";
echo call_user_func('\strtoupper', 'z'), call_user_func_array('\strrev', ['ab']), "\n";
echo is_callable('\trim') ? 'Y' : 'N', function_exists('\trim') ? 'Y' : 'N', "\n";
