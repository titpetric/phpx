name: variable variables read and write a variable by name
description: >
  `$$name`, `$$$name` and `${expr}` name a variable at run time, on either side
  of an assignment. carbon's createSafe() reads `$$field` for each range.
  Expected output is php 8.4's.
runner:
  flatstack: false
---
<?php
function f($year = 2020, $month = 3) {
  $fields = ['year' => [1, 9999], 'month' => [1, 12]];
  foreach ($fields as $field => $range) { if ($$field !== null && $$field > $range[1]) { echo "bad $field\n"; } }
  $name = 'dyn'; $$name = 'set'; echo $dyn, "\n";
  $$name .= '!'; echo $dyn, "\n";
  ${'a' . 'b'} = 5; echo $ab, "\n";
  $ptr = 'name'; echo $$ptr, "\n";
  echo ${$name}, "\n";
  return $$name;
}
echo f(2020, 13), "\n";
$g = 'x'; $x = 9; echo $$g, "\n";
$n = 'gone'; $gone = 'v'; unset($$n); var_dump(isset($gone));
$k = 1; $v1 = 'a'; $v2 = 'b'; unset(${'v' . $k}); var_dump(isset($v1), isset($v2));
---
bad month
set
set!
5
dyn
set!
set!
9
bool(false)
bool(false)
bool(true)
