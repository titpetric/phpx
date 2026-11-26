name: serialize writes an object's visibility into its property keys
description: >
  serialize() writes php's object format: a public property under its name, a
  protected one under a NUL, an asterisk, a NUL and the name, and a private one
  under a NUL, the declaring class, a NUL and the name. The profile is shown as
  hex because a mangled key holds NUL bytes. unserialize() builds the instance
  again without running a constructor, a class excluded by allowed_classes
  becomes __PHP_Incomplete_Class, and an array of scalars, a string key and
  null round-trips. Expected output is php 8.4's.
---
<?php
class Point {
    public int $x = 1;
    protected string $label = 'p';
    private array $meta = ['k' => 2];
    public function show() { echo $this->x, ' ', $this->label, ' ', json_encode($this->meta), "\n"; }
}
$p = new Point;
$p->x = 7;
$s = serialize($p);
echo bin2hex($s), "\n";
$q = unserialize($s);
echo get_class($q), ' ';
$q->show();

echo serialize([1, 'two' => [true, null]]), "\n";
echo json_encode(unserialize('a:2:{i:0;i:1;s:3:"two";a:2:{i:0;b:1;i:1;N;}}')), "\n";

$incomplete = unserialize($s, ['allowed_classes' => false]);
echo $incomplete instanceof __PHP_Incomplete_Class ? 'incomplete' : 'built', "\n";
---
4f3a353a22506f696e74223a333a7b733a313a2278223b693a373b733a383a22002a006c6162656c223b733a313a2270223b733a31313a2200506f696e74006d657461223b613a313a7b733a313a226b223b693a323b7d7d
Point 7 p {"k":2}
a:2:{i:0;i:1;s:3:"two";a:2:{i:0;b:1;i:1;N;}}
{"0":1,"two":[true,null]}
incomplete
