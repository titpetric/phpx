<?php
namespace N\Sub;
use function hrtime;
interface Ev {}
interface Ev2 extends Ev {}
abstract class Base implements Ev2 {}
final class HRTime extends Base { public static function make(int $a, int $b): self { return new self; } }
class P implements \Countable, \JsonSerializable {
    function count(): int { return 1; }
    function jsonSerialize(): mixed { return 1; }
}
echo get_class(HRTime::make(...[1, 2])), "\n";
$i = class_implements(HRTime::class);
ksort($i); // php orders a parent's interfaces by a rule of its own
print_r($i);
print_r(class_implements(P::class));
print_r(class_implements(Ev2::class));
print_r(class_implements(new P));
var_dump(class_implements('N\Sub\Nope', false));
var_dump(class_alias(P::class, 'N\Sub\Alias'));
echo get_class(new Alias), "\n";
