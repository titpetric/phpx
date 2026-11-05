name: a global is read in the global scope, and in a frame that asks for it
phpx: true
runner:
  php: false
description: >
  $argv and $argc are globals, and the file body reads them where a function
  and a closure do not. The method is the difference phpx makes: `global $argv`
  binds the name to the global frame and the read answers true, where upstream
  phpscript parses the statement and binds nothing. The bytecode engine
  declines a program that uses the statement and the column skips. A
  superglobal is the other rule and is read from every frame. The harness
  supplies the request, which the PHP CLI SAPI has no equivalent of.

request:
  args:
    name: Alice
---
<?php

function reader()
{
	return "argv=" . var_export(isset($argv), true)
		. " argc=" . var_export(isset($argc), true)
		. " server=" . var_export(isset($_SERVER), true);
}

class Holder
{
	public function read()
	{
		global $argv;
		return "argv=" . var_export(isset($argv), true);
	}
}

echo "top: argv=", var_export(isset($argv), true),
	" argc=", var_export(isset($argc), true),
	" name=", $argv[1], "\n";
echo "function: ", reader(), "\n";

$closure = function () {
	return "argv=" . var_export(isset($argv), true)
		. " server=" . var_export(isset($_SERVER), true);
};
echo "closure: ", $closure(), "\n";

$capturing = function () use ($argv) {
	return "argv=" . var_export(isset($argv), true) . " name=" . $argv[1];
};
echo "capture: ", $capturing(), "\n";
echo "method: ", (new Holder)->read(), "\n";
?>
---
top: argv=true argc=true name=Alice
function: argv=false argc=false server=true
closure: argv=false server=true
capture: argv=true name=Alice
method: argv=true
