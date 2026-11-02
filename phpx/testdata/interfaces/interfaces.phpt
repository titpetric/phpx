name: an interface is satisfied by an inherited or a trait method, and it carries its constants
description: >
  php reads `implements` against the finished class: a method a parent carried
  and a method a trait contributed both meet the contract, and only a method
  the class has nowhere is missing. An interface contributes its constants to
  the class that names it, an interface that extends another answers the
  inherited constant through its own name, and a class constant of the same
  name wins. The expected section is what `php` printed. The bytecode engine
  declines a program that declares `extends` or a trait, so its column skips
  and the whole program runs on the interpreter.
---
<?php

interface Quota
{
	const LIMIT = 5;
	public function limit();
}

interface Named extends Quota
{
	const PREFIX = "quota:";
	public function name();
}

class Base
{
	public function limit()
	{
		return Quota::LIMIT;
	}
}

class Store extends Base implements Named
{
	public function name()
	{
		return self::PREFIX . "store";
	}
}

trait Named2
{
	public function name()
	{
		return "trait";
	}
}

class Other implements Quota
{
	use Named2;

	public function limit()
	{
		return 2;
	}
}

class Limited implements Quota
{
	const LIMIT = 9;

	public function limit()
	{
		return self::LIMIT;
	}
}

$store = new Store;
echo $store->limit(), " ", $store->name(), "\n";
echo Store::LIMIT, " ", Named::LIMIT, " ", Store::PREFIX, "\n";
echo $store instanceof Named ? "a Named" : "not a Named", "\n";

$other = new Other;
echo $other->name(), " ", Other::LIMIT, " ", $other->limit(), "\n";

$limited = new Limited;
echo $limited->limit(), " ", Limited::LIMIT, " ", Quota::LIMIT, "\n";
---
5 quota:store
5 5 quota:
a Named
trait 5 2
9 9 5
