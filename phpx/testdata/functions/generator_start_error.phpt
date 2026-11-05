name: a generator's body that throws before its first yield surfaces on valid()
description: >
  valid(), current() and key() start the generator; an exception from the body
  before the first yield reaches the caller. The bytecode engine has no
  generators, so the fixture opts it out.
runner:
  flatstack: false
---
<?php
function g() { throw new RuntimeException('boom'); yield 1; }
$x = g();
try { var_dump($x->valid()); } catch (Throwable $e) { echo "caught ", $e->getMessage(), "\n"; }
function h() { yield 1; throw new RuntimeException('late'); }
$y = h();
try { foreach ($y as $v) { echo $v; } } catch (Throwable $e) { echo " caught ", $e->getMessage(), "\n"; }
$z = h(); $z->current(); try { $z->next(); var_dump($z->valid()); } catch (Throwable $e) { echo "caught2 ", $e->getMessage(), "\n"; }
---
caught boom
1 caught late
caught2 late
