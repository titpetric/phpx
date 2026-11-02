name: an abstract class leaves its methods to the child
description: >
  One hierarchy settles what php gives `abstract`: a class that declares the
  modifier is never instantiated itself, an abstract method has no body and
  the child supplies one, a child that implements every method instantiates
  and runs the concrete methods it inherited, an instance answers instanceof
  for the abstract ancestor, a class that declares no abstract method may
  still be abstract and a child of it instantiates, and a trait's abstract
  method is satisfied by the class that uses the trait. The expected section
  is what `php` printed. The bytecode engine declines a program that declares
  an abstract class or an abstract method, so its column skips and the whole
  program runs on the interpreter.
---
<?php

abstract class Shape
{
	abstract public function area();
	abstract public function name();

	public function describe()
	{
		return $this->name() . ":" . $this->area();
	}
}

class Square extends Shape
{
	public function area()
	{
		return 4;
	}

	public function name()
	{
		return "square";
	}
}

class Circle extends Shape
{
	public function area()
	{
		return 3;
	}

	public function name()
	{
		return "circle";
	}
}

$s = new Square();
echo $s->describe(), "\n";
echo $s instanceof Shape ? "shape\n" : "no\n";
echo (new Circle())->describe(), "\n";

abstract class Base
{
	public function hello()
	{
		return "hello";
	}
}

class Child extends Base
{
}

echo (new Child())->hello(), "\n";

trait Labelled
{
	abstract public function label();
}

class Tag
{
	use Labelled;

	public function label()
	{
		return "tag";
	}
}

echo (new Tag())->label(), "\n";
---
square:4
shape
circle:3
hello
tag
