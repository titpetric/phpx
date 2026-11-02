name: a property type may name a fully-qualified class
description: >
  php accepts a leading backslash on a property type the same way it does on a
  parameter or return type, and framework code writes `private \stdClass
  $providers;` to say "global stdClass" inside a namespace. The class body
  parser read the backslash as a token it had no case for and refused the file.
  The expected section is what `php` printed.
---
<?php

class Holder
{
	private \stdClass $value;
	public ?\ArrayObject $list = null;
	protected \DateTimeInterface|\stdClass $either;
	public static ?\stdClass $shared = null;

	public function __construct()
	{
		$this->value = new \stdClass();
		$this->value->name = "typed";
		$this->either = new \stdClass();
		$this->either->kind = "object";
	}

	public function describe()
	{
		return $this->value->name . "," . $this->either->kind . "," . ($this->list === null ? "null" : "set");
	}
}

$h = new Holder();
echo $h->describe(), "\n";
var_dump(Holder::$shared);
---
typed,object,null
NULL
