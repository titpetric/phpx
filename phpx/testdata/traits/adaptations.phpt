name: trait adaptations, insteadof and as
description: >
  A `use` line may carry a block: `A::f insteadof B;` picks the winner of a
  collision, `B::f as g;` adds a second name, and `f as protected;` changes
  visibility, alone or with a name. A trait composing traits takes the block
  too. The expected section is what `php` printed. Visibility is not asserted:
  the runtime enforces none on a method call (see ACCEPTANCE.md).
---
<?php

trait Hello
{
	public function say() { return "Hello"; }
	public function who() { return "hello:" . static::class; }
}

trait World
{
	public function say() { return "World"; }
	private function secret() { return "s"; }
}

class Both
{
	use Hello, World {
		Hello::say insteadof World;
		World::say as sayWorld;
		secret as public;
		who as protected whoProtected;
	}

	public function proxy() { return $this->whoProtected(); }
}

class Plain
{
	use Hello {
		say as private;
		say as greet;
	}
}

trait Nested
{
	use Hello, World {
		World::say insteadof Hello;
		Hello::say as sayHello;
	}
}

class UsesNested { use Nested; }

$b = new Both();
echo $b->say(), " ", $b->sayWorld(), " ", $b->secret(), "\n";
echo $b->proxy(), "\n";
$p = new Plain();
echo $p->greet(), "\n";
echo method_exists($p, 'say') ? "has say" : "no say", "\n";
$n = new UsesNested();
echo $n->say(), " ", $n->sayHello(), "\n";
---
Hello World s
hello:Both
Hello
has say
World Hello
