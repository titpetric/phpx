<?php
$utc = new DateTimeZone('UTC');
$d = new DateTimeImmutable('2020-01-02 03:04:05', $utc);
echo $d->format('Y-m-d H:i:s'), '|', $d->format(DATE_ATOM), '|', $d->format('D, d M Y jS \of F'), "\n";
echo $d->getTimestamp(), ' ', $d->format('U'), ' ', $d->format('N w z t L W'), "\n";
$e = $d->modify('+1 day'); echo $e->format('Y-m-d'), ' ', $d->format('Y-m-d'), "\n";
echo $d->add(new DateInterval('P1M2DT3H'))->format('c'), "\n";
echo $d->sub(new DateInterval('PT30M'))->format('H:i'), "\n";
$ny = new DateTimeZone('America/New_York');
echo $d->setTimezone($ny)->format('Y-m-d H:i T P e'), "\n";
echo (new DateTimeImmutable('@86400'))->format('Y-m-d\TH:i:sP'), "\n";
$m = new DateTime('2021-06-15 12:00:00', $utc); $m->modify('+2 hours'); echo $m->format('H:i'), "\n";
var_dump($d instanceof DateTimeInterface, $m instanceof DateTimeInterface, $utc->getName());
echo $d->setDate(2022, 2, 3)->setTime(4, 5)->format('Y-m-d H:i:s.u v'), "\n";
echo DateTimeInterface::RFC3339, "\n";
