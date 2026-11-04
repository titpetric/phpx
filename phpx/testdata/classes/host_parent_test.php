<?php
class Stamp extends DateTimeImmutable {
    public function __construct(string $t = 'now', ?DateTimeZone $z = null) { parent::__construct($t, $z ?? new DateTimeZone('UTC')); }
    public function label(): string { return 'at ' . $this->format('Y-m-d'); }
}
$s = new Stamp('2020-05-06 07:08:09');
echo $s->label(), ' ', $s->format('H:i'), ' ', $s->getTimestamp(), "\n";
var_dump($s instanceof DateTimeImmutable);
