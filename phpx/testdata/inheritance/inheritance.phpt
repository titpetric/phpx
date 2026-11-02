name: extends carries the parent's members to the child
description: >
  One hierarchy settles what php gives `extends`: fields keep the parent's
  default where the child does not redeclare and take the child's where it
  does, methods are inherited and an override wins, a grandchild reaches two
  levels, parent:: runs the parent's body with $this still bound, a constant
  is inherited and overridden, an instance answers instanceof for every
  ancestor, get_parent_class and is_subclass_of read the chain, and a catch
  clause naming a grandparent takes a thrown grandchild. `self::` inside an
  inherited body resolves to the class that declared it, which is why the
  inherited describe() reports "animal" on a Dog. The expected section is what
  `php` printed. The bytecode engine declines a program that declares
  `extends`, so its column skips and the whole program runs on the
  interpreter.
---
<?php

class Animal
{
	public $name = "animal";
	public $legs = 4;
	const KIND = "animal";

	public function __construct($name)
	{
		$this->name = $name;
	}

	public function describe()
	{
		return $this->name . " is a " . self::KIND . " with " . $this->legs . " legs";
	}

	public function speak()
	{
		return $this->name . " makes a sound";
	}
}

class Dog extends Animal
{
	public $legs = 3;
	public $breed = "mixed";
	const KIND = "dog";

	public function __construct($name, $breed)
	{
		parent::__construct($name);
		$this->breed = $breed;
	}

	public function speak()
	{
		return $this->name . " barks";
	}
}

class Puppy extends Dog
{
	public function speak()
	{
		return parent::speak() . " softly";
	}
}

class MyException extends Exception
{
}

class NotFound extends MyException
{
}

$animal = new Animal("cat");
$dog = new Dog("Rex", "labrador");
$puppy = new Puppy("Tiny", "poodle");

echo "inherited: ", $dog->describe(), "\n";
echo "two levels: ", $puppy->describe(), "\n";
echo "override: ", $dog->speak(), "\n";
echo "parent:: : ", $puppy->speak(), "\n";
echo "base: ", $animal->speak(), "\n";
echo "breed: ", $dog->breed, "\n";
echo "constants: ", Dog::KIND, " ", Animal::KIND, "\n";
echo "instanceof Animal: ", var_export($dog instanceof Animal, true), "\n";
echo "instanceof Dog: ", var_export($dog instanceof Dog, true), "\n";
echo "instanceof Puppy: ", var_export($dog instanceof Puppy, true), "\n";
echo "instanceof Animal: ", var_export($puppy instanceof Animal, true), "\n";
echo "get_class: ", get_class($puppy), "\n";
echo "get_parent_class: ", get_parent_class($puppy), "\n";
echo "is_subclass_of: ", var_export(is_subclass_of($puppy, "Animal"), true), "\n";
echo "is_subclass_of self: ", var_export(is_subclass_of($puppy, "Puppy"), true), "\n";

try {
	throw new NotFound();
} catch (MyException $e) {
	echo "caught ancestor: ", get_class($e), "\n";
}
---
inherited: Rex is a animal with 3 legs
two levels: Tiny is a animal with 3 legs
override: Rex barks
parent:: : Tiny barks softly
base: cat makes a sound
breed: labrador
constants: dog animal
instanceof Animal: true
instanceof Dog: true
instanceof Puppy: false
instanceof Animal: true
get_class: Puppy
get_parent_class: Dog
is_subclass_of: true
is_subclass_of self: false
caught ancestor: NotFound
