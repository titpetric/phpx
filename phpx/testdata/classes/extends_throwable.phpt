name: extends a built-in throwable
description: >
  A class extending Exception, Error or an SPL exception calls parent::__construct,
  inherits getMessage, getCode and getPrevious, and matches catch clauses and
  instanceof by php's hierarchy. The expected section is what `php` printed.
---
<?php
class AppException extends Exception {
    private $extra;
    public function __construct($m, $extra = 'x', $prev = null) { parent::__construct($m, 5, $prev); $this->extra = $extra; }
    public function getExtra() { return $this->extra; }
}
class NotFound extends AppException {}
class Bare extends RuntimeException { public function hi() { return "hi"; } }
class Fault extends Error {}
class Bad extends InvalidArgumentException {}

try { throw new NotFound("boom", "e", new Bare("inner", 9)); } catch (AppException $e) {
    echo get_class($e), ":", $e->getMessage(), ":", $e->getCode(), ":", $e->getExtra(), "\n";
    echo get_class($e->getPrevious()), ":", $e->getPrevious()->getMessage(), ":", $e->getPrevious()->getCode(), "\n";
}
$b = new Bare("nf", 3);
var_dump($b instanceof RuntimeException, $b instanceof Exception, $b instanceof Throwable, $b instanceof LogicException, $b instanceof Error);
echo $b->getMessage(), $b->getCode(), $b->hi(), "\n";
var_dump(new Bad("x") instanceof LogicException, get_parent_class(new Bad("x")));
var_dump(count(array_filter(class_parents(new Bad("x")))));
foreach ([new Bare("a"), new Fault("f"), new Bad("b")] as $t) {
    try { throw $t; }
    catch (Exception $e) { echo "Exception:", get_class($e), "\n"; }
    catch (Error $e) { echo "Error:", get_class($e), "\n"; }
}
try { throw new Fault("f"); } catch (Throwable $t) { echo get_class($t), $t->getMessage(), "\n"; }
try { throw new Bad("lg"); } catch (RuntimeException $r) { echo "wrong\n"; } catch (LogicException $l) { echo "logic ", $l->getMessage(), "\n"; }
var_dump((new Bare("m"))->getPrevious());
---
NotFound:boom:5:e
Bare:inner:9
bool(true)
bool(true)
bool(true)
bool(false)
bool(false)
nf3hi
bool(true)
string(24) "InvalidArgumentException"
int(3)
Exception:Bare
Error:Fault
Exception:Bad
Faultf
logic lg
NULL
