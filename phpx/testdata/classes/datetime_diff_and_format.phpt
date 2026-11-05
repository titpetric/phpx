name: DateTime diff, createFromFormat and the interval
description: >
  diff counts the calendar the way php 8.1 does in both directions, createFromFormat reads its format characters, DateTime setters mutate and answer $this, DateInterval formats and reads its properties. Expected output is php 8.4's.
runner:
  flatstack: false
---
<?php
date_default_timezone_set('UTC');
$a = new DateTimeImmutable('2024-01-31 10:00:00'); $b = new DateTimeImmutable('2024-03-01 09:30:15');
$d = $a->diff($b); echo $d->format('%R %yy %mm %dd %hh %im %ss days=%a'), ' ', $d->y, $d->m, $d->d, $d->h, $d->i, $d->s, ' ', var_export($d->days, true), ' ', $d->invert, "\n";
$d = $b->diff($a); echo $d->format('%R%a %m %d'), ' ', $d->invert, "\n";
echo (new DateTime('2000-02-29'))->diff(new DateTime('2024-02-28'))->format('%y-%m-%d %a'), "\n";
$x = DateTime::createFromFormat('Y-m-d H:i:s', '2021-05-06 07:08:09'); echo $x->format('c'), get_class($x), "\n";
$x = DateTimeImmutable::createFromFormat('d/m/Y', '15/08/2022'); echo $x->format('Y-m-d'), ' ', get_class($x), ' ', $x->format('H') !== '' ? 'time' : '', "\n";
$x = DateTime::createFromFormat('!d/m/Y', '15/08/2022'); echo $x->format('c'), "\n";
$x = DateTime::createFromFormat('j-M-Y g:i A', '5-Mar-2020 3:30 PM'); echo $x->format('Y-m-d H:i'), "\n";
$x = DateTime::createFromFormat('U', '86400'); echo $x->format('c'), "\n";
$x = DateTime::createFromFormat('Y-m-d H:i:s T', '2020-01-02 03:04:05 EST'); echo $x->format('c T'), "\n";
var_dump(DateTime::createFromFormat("Y-m-d", "nope"));
$m = new DateTime('2024-05-06'); $r = $m->setDate(2020, 2, 30)->setTime(25, 61); echo $r === $m ? 'same ' : 'other ', $m->format('c'), "\n";
$m = new DateTime('2024-05-06'); $m->setISODate(2024, 1, 1); echo $m->format('Y-m-d'), "\n";
$i = new DateInterval('P1Y2M3DT4H5M6S'); echo $i->format('%y %m %d %h %i %s %R %a'), ' ', var_export($i->days, true), "\n";
$i2 = DateInterval::createFromDateString('3 days ago'); echo $i2->d, ' ', $i2->invert, "\n";
echo $a->add(new DateInterval('P1M'))->format('Y-m-d'), ' ', $a->sub(new DateInterval('PT36H'))->format('Y-m-d H:i'), "\n";
var_dump(checkdate(2, 29, 2023), checkdate(2, 29, 2024), mktime(12, 0, 0, 6, 15, 2020), gmmktime(0, 0, 0, 1, 1, 1970), idate('Y', 86400 * 366));
print_r(date_parse('2006-12-12 10:00:00.5')); 
foreach ([['2024-03-31', '2024-02-29'], ['2023-12-31 23:59:59', '2024-01-01 00:00:00'], ['2020-02-29', '2021-02-28'], ['2024-01-15', '2024-01-15 12:00'], ['2024-05-31', '2024-06-30'], ['2022-10-31 22:00', '2022-11-01 01:00']] as [$p, $q]) { $d = (new DateTime($p))->diff(new DateTime($q)); echo $p, '→', $q, ' ', $d->format('%R %yy %mm %dd %hh %im %ss a=%a'), "\n"; $d = (new DateTime($q))->diff(new DateTime($p)); echo '  rev ', $d->format('%R %yy %mm %dd %hh %im %ss a=%a'), "\n"; }
---
+ 0y 0m 29d 23h 30m 15s days=29 0029233015 29 0
-29 1 0 1
23-11-30 8765
2021-05-06T07:08:09+00:00DateTime
2022-08-15 DateTimeImmutable time
2022-08-15T00:00:00+00:00
2020-03-05 15:30
1970-01-02T00:00:00+00:00
2020-01-02T03:04:05-05:00 EST
bool(false)
same 2020-03-02T02:01:00+00:00
2024-01-01
1 2 3 4 5 6 + (unknown) false
-3 0
2024-03-02 2024-01-29 22:00
bool(false)
bool(true)
int(1592222400)
int(0)
int(1971)
Array
(
    [year] => 2006
    [month] => 12
    [day] => 12
    [hour] => 10
    [minute] => 0
    [second] => 0
    [fraction] => 0.5
    [warning_count] => 0
    [warnings] => Array
        (
        )

    [error_count] => 0
    [errors] => Array
        (
        )

    [is_localtime] => 
)
2024-03-31→2024-02-29 - 0y 1m 2d 0h 0m 0s a=31
  rev + 0y 1m 2d 0h 0m 0s a=31
2023-12-31 23:59:59→2024-01-01 00:00:00 + 0y 0m 0d 0h 0m 1s a=0
  rev - 0y 0m 0d 0h 0m 1s a=0
2020-02-29→2021-02-28 + 0y 11m 30d 0h 0m 0s a=365
  rev - 0y 11m 28d 0h 0m 0s a=365
2024-01-15→2024-01-15 12:00 + 0y 0m 0d 12h 0m 0s a=0
  rev - 0y 0m 0d 12h 0m 0s a=0
2024-05-31→2024-06-30 + 0y 0m 30d 0h 0m 0s a=30
  rev - 0y 0m 30d 0h 0m 0s a=30
2022-10-31 22:00→2022-11-01 01:00 + 0y 0m 0d 3h 0m 0s a=0
  rev - 0y 0m 0d 3h 0m 0s a=0
