---
name: vendor-code predicates and string readers
description: >
  Bindings the top composer packages call with one or two arguments. The
  expected section is what `php` printed.
---
<?php
class P {}
class C extends P {}
class D extends C {}
$f = fopen('php://output', 'w');
foreach ([null, true, 1, 1.5, "s", [], new C()] as $v) { echo get_debug_type($v), " ", var_export(is_scalar($v), true), " ", var_export(is_resource($v), true), "\n"; }
var_dump(is_resource($f), is_iterable([1]), is_iterable(1), is_countable([]), is_countable("x"));
var_dump(boolval("0"), boolval("a"), strval(12), strval(1.5), strval(true));
var_dump(is_nan(NAN), is_nan(1.0), is_finite(INF), is_finite(2.0));
foreach (["123", "12a", "", "abc", "ABC", "aB1", " \t\n", "x y"] as $s) {
	echo "[$s] ", (int)ctype_digit($s), (int)ctype_alpha($s), (int)ctype_alnum($s), (int)ctype_upper($s), (int)ctype_lower($s), (int)ctype_space($s), "\n";
}
var_dump(strcspn("abcd", "cd"), strcspn("abcd", "x"), octdec("777"), fmod(10, 3), mb_strwidth("abc"), nl2br("a\nb"));
var_dump(json_last_error(), json_last_error_msg(), preg_last_error(), preg_last_error_msg());
var_dump(is_string(gethostname()), is_string(sys_get_temp_dir()), date_default_timezone_get() !== '');
print_r(class_parents(new D()));
print_r(class_parents('C'));
interface I {}
var_dump(interface_exists('I'), interface_exists('P'), interface_exists('Nope'));
---
null false false
bool true false
int true false
float true false
string true false
array false false
C false false
bool(true)
bool(true)
bool(false)
bool(true)
bool(false)
bool(false)
bool(true)
string(2) "12"
string(3) "1.5"
string(1) "1"
bool(true)
bool(false)
bool(false)
bool(true)
[123] 101000
[12a] 001000
[] 000000
[abc] 011010
[ABC] 011100
[aB1] 001000
[ 	
] 000001
[x y] 000000
int(2)
int(4)
int(511)
float(1)
int(3)
string(9) "a<br />
b"
int(0)
string(8) "No error"
int(0)
string(8) "No error"
bool(true)
bool(true)
bool(true)
Array
(
    [C] => C
    [P] => P
)
Array
(
    [P] => P
)
bool(true)
bool(false)
bool(false)
