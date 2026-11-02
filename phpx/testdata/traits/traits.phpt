name: a trait's members land in every class that uses it
description: >
  What php gives `use`: a trait's properties, constants, statics and methods
  arrive in each class that names it, a class overrides a trait method with its
  own body, `self::` inside a trait method resolves to the class that used the
  trait, and a trait composes another trait through its own `use` line. Two
  traits declaring one property with the same definition compose, and two
  classes using one trait take independent storage. A trait is a bag of members
  and not a type: `instanceof` against its name is false, while method_exists
  finds the method it installed. The expected section is what `php` printed.
  The bytecode engine declines a program that declares or uses a trait, so its
  column skips and the whole program runs on the interpreter.
---
<?php

trait Greets
{
	public $greeting = "hello";
	const SUFFIX = "!";

	public function greet()
	{
		return $this->greeting . " " . $this->name . self::SUFFIX;
	}
}

trait Counts
{
	public static $instances = 0;

	public function count()
	{
		self::$instances = self::$instances + 1;
		return self::$instances;
	}
}

trait Loud
{
	public function shout()
	{
		return strtoupper($this->name) . "!";
	}
}

trait Named
{
	use Loud;

	public $name = "trait name";
	public $greeting = "hello";

	public function labels()
	{
		return $this->shout() . " " . $this->name;
	}
}

class Base
{
	use Greets;
}

class Person
{
	use Greets, Counts;

	public $name;

	public function __construct($name)
	{
		$this->name = $name;
	}

	public function greet()
	{
		return "override: " . $this->name;
	}
}

class Box
{
	use Named;
}

class Sign
{
	use Named;

	public $name = "trait name";
}

class Card
{
	use Greets, Named;
}

$base = new Base();
$base->name = "base";
$person = new Person("Ann");
$card = new Card();
$card->name = "card";

echo $base->greet(), "\n";
echo $person->greet(), "\n";
echo $person->count(), "\n";
echo $person->count(), "\n";
echo "static: ", Person::$instances, "\n";
echo "instanceof trait: ", var_export($person instanceof Greets, true), "\n";
echo "method_exists: ", var_export(method_exists($person, "count"), true), "\n";
echo "property: ", $person->greeting, "\n";
echo (new Box())->labels(), "\n";
echo (new Sign())->labels(), "\n";
echo $card->greet(), "\n";
echo $card->labels(), "\n";
---
hello base!
override: Ann
1
2
static: 2
instanceof trait: false
method_exists: true
property: hello
TRAIT NAME! trait name
TRAIT NAME! trait name
hello card!
CARD! card
