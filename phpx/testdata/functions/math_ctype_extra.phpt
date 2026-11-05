name: trig, log and ctype classes
description: >
  The trigonometric and logarithmic functions, and the ctype classes the core binding left out. Expected output is php 8.4's.
runner:
  flatstack: false
---
<?php
$xs = [0.5, 1.0, 2.0];
foreach (['sin','cos','tan','asin','acos','atan','sinh','cosh','tanh','asinh','atanh','expm1','log1p','exp','log10','deg2rad','rad2deg'] as $f) {
  $r = [];
  foreach ($xs as $x) { $v = @$f($x); $r[] = is_nan($v) ? 'NAN' : (is_infinite($v) ? 'INF' : round($v, 10)); }
  echo $f, ' ', implode(' ', $r), "\n";
}
echo round(atan2(1, -1), 10), ' ', hypot(3, 4), ' ', round(pi(), 10), ' ', round(acosh(2.0), 10), "\n";
var_dump(getrandmax(), mt_getrandmax());
foreach (['ctype_cntrl' => "\t\n", 'ctype_graph' => "ab!", 'ctype_print' => "a b", 'ctype_punct' => "!?.", 'ctype_xdigit' => "dEaD09"] as $f => $s) { var_dump($f($s), $f('')); }
var_dump(ctype_punct('a!'), ctype_xdigit('xyz'), ctype_print("\x01"));
srand(5); $a = rand(); srand(5); var_dump($a === rand());
---
sin 0.4794255386 0.8414709848 0.9092974268
cos 0.8775825619 0.5403023059 -0.4161468365
tan 0.5463024898 1.5574077247 -2.1850398633
asin 0.5235987756 1.5707963268 NAN
acos 1.0471975512 0 NAN
atan 0.463647609 0.7853981634 1.1071487178
sinh 0.5210953055 1.1752011936 3.6268604078
cosh 1.1276259652 1.5430806348 3.7621956911
tanh 0.4621171573 0.761594156 0.9640275801
asinh 0.4812118251 0.881373587 1.4436354752
atanh 0.5493061443 INF NAN
expm1 0.6487212707 1.7182818285 6.3890560989
log1p 0.4054651081 0.6931471806 1.0986122887
exp 1.6487212707 2.7182818285 7.3890560989
log10 -0.3010299957 0 0.3010299957
deg2rad 0.0087266463 0.0174532925 0.034906585
rad2deg 28.6478897565 57.2957795131 114.5915590262
2.3561944902 5 3.1415926536 1.3169578969
int(2147483647)
int(2147483647)
bool(true)
bool(false)
bool(true)
bool(false)
bool(true)
bool(false)
bool(true)
bool(false)
bool(true)
bool(false)
bool(false)
bool(false)
bool(false)
bool(true)
