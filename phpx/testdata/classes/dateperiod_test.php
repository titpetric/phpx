<?php
$s = new DateTimeImmutable("2024-01-01");
$e = new DateTimeImmutable("2024-01-04");
$i = new DateInterval("P1D");
foreach (new DatePeriod($s, $i, $e) as $k => $d) echo $k, " ", $d->format("Y-m-d"), "\n";
echo "--excl\n";
foreach (new DatePeriod($s, $i, $e, DatePeriod::EXCLUDE_START_DATE) as $k => $d) echo $k, " ", $d->format("Y-m-d"), "\n";
echo "--incl\n";
foreach (new DatePeriod($s, $i, $e, DatePeriod::INCLUDE_END_DATE) as $k => $d) echo $k, " ", $d->format("Y-m-d"), "\n";
echo "--rec\n";
foreach (new DatePeriod($s, $i, 2) as $k => $d) echo $k, " ", $d->format("Y-m-d"), "\n";
foreach (new DatePeriod($s, $i, 2, DatePeriod::EXCLUDE_START_DATE) as $k => $d) echo $k, " ", $d->format("Y-m-d"), "\n";
echo "--mut\n";
$p = new DatePeriod(new DateTime("2024-01-30"), new DateInterval("P1M"), 2);
foreach ($p as $d) echo get_class($d), " ", $d->format("Y-m-d"), "\n";
echo $p->getStartDate()->format("Y-m-d"), " ", var_export($p->getEndDate(), true), " ", $p->getDateInterval()->d, " ", $p->recurrences, "\n";
echo count(iterator_to_array(new DatePeriod($s, $i, $e))), "\n";
echo $p->include_start_date ? "y":"n", $p->include_end_date ? "y":"n", "\n";
class P extends DatePeriod {}
foreach (new P($s, $i, $e) as $d) echo $d->format("md"), " ";
echo "\n";
echo (new DatePeriod($s, $i, $e)) instanceof Traversable ? "T" : "F", "\n";
echo "--compare\n";
var_dump($s < $i2 = $e, $s > $e, $s == new DateTime("2024-01-01"), new DateTime("2024-01-02") >= $s);
