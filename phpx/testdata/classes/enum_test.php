<?php
enum Suit: string
{
    case Hearts = 'H';
    case Spades = 'S';
    const Wild = self::Spades;
    public function label(): string { return ucfirst(strtolower($this->name)); }
    public function color(): string
    {
        return match ($this) { self::Hearts => 'Red', self::Spades => 'Black' };
    }
    public static function fromChar(string $c): self { return self::from($c); }
}
enum Level: int implements JsonSerializable
{
    case Low = 1;
    case High = 10;
    public function jsonSerialize(): mixed { return $this->name; }
}
enum Plain { case One; case Two; }

echo Suit::Hearts->value, Suit::Hearts->name, Suit::Spades->label(), Suit::Hearts->color(), "\n";
var_dump(Suit::Hearts === Suit::Hearts, Suit::Hearts === Suit::Spades, Suit::Wild === Suit::Spades);
echo count(Suit::cases()), Suit::cases()[1]->name, "\n";
echo Suit::from('S')->name, Suit::tryFrom('S')?->name, var_export(Suit::tryFrom('Z'), true), "\n";
try { Suit::from('Z'); } catch (ValueError $e) { echo get_class($e), ': ', $e->getMessage(), "\n"; }
try { Level::from(5); } catch (ValueError $e) { echo $e->getMessage(), "\n"; }
echo Level::from(10)->name, Level::tryFrom(1)->value + 1, "\n";
var_dump(Suit::Hearts instanceof Suit, Suit::Hearts instanceof UnitEnum, Suit::Hearts instanceof BackedEnum, Plain::One instanceof BackedEnum);
echo Plain::Two->name, count(Plain::cases()), "\n";
var_dump(Suit::fromChar('H') === Suit::Hearts);
function pick(Suit $s): string { return $s->name; }
echo pick(Suit::Spades), "\n";
$map = [];
foreach (Suit::cases() as $case) { $map[$case->value] = $case->label(); }
echo json_encode($map), "\n";
$x = Suit::Hearts;
echo match (true) { $x === Suit::Hearts => 'is hearts', default => 'other' }, "\n";
try { $x->name = 'no'; } catch (Error $e) { echo get_class($e), ': ', $e->getMessage(), "\n"; }
