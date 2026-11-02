name: a class property that repeats a trait property must agree
runner:
  php: false
description: >
  A method the class declares wins over the trait's, but a property does not:
  php composes the class's own declaration with the trait's and both must
  agree on visibility, type and initial value. Two different defaults are a
  fatal, and the message names the class first. The expected section is what
  `php` printed, which is nothing. The expected error names the message php
  printed. The bytecode engine declines a program that declares or uses a
  trait, so its column skips and the whole program runs on the interpreter.
error: 'Owner and Labelled define the same property ($name) in the composition of Owner.'
---
<?php

trait Labelled
{
	public $name = "trait name";
}

class Owner
{
	use Labelled;

	public $name = "own name";
}

echo "unreachable\n";
---
Internal Server Error
