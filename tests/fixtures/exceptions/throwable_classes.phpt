name: a catch clause filters on the class hierarchy
phpx: true
description: >
  phpx reverses the "no hierarchy" decision of docs/design.md for the built-in
  throwables: a clause naming a parent takes a child, as in php, so
  InvalidArgumentException is a LogicException and ArgumentCountError a
  TypeError. Expected output is php 8.4's.
---
<?php

// One line per thrown class: which of the clause types below take it.
// A clause that does not match falls through to the Throwable clause.

$row = "Exception:";
try { throw new Exception("x"); } catch (Exception $e) { $row = $row . " Exception"; } catch (Throwable $e) { }
try { throw new Exception("x"); } catch (Error $e) { $row = $row . " Error"; } catch (Throwable $e) { }
try { throw new Exception("x"); } catch (RuntimeException $e) { $row = $row . " RuntimeException"; } catch (Throwable $e) { }
echo $row, "\n";

$row = "RuntimeException:";
try { throw new RuntimeException("x"); } catch (Exception $e) { $row = $row . " Exception"; } catch (Throwable $e) { }
try { throw new RuntimeException("x"); } catch (Error $e) { $row = $row . " Error"; } catch (Throwable $e) { }
try { throw new RuntimeException("x"); } catch (RuntimeException $e) { $row = $row . " RuntimeException"; } catch (Throwable $e) { }
try { throw new RuntimeException("x"); } catch (LogicException $e) { $row = $row . " LogicException"; } catch (Throwable $e) { }
echo $row, "\n";

$row = "InvalidArgumentException:";
try { throw new InvalidArgumentException("x"); } catch (Exception $e) { $row = $row . " Exception"; } catch (Throwable $e) { }
try { throw new InvalidArgumentException("x"); } catch (InvalidArgumentException $e) { $row = $row . " InvalidArgumentException"; } catch (Throwable $e) { }
try { throw new InvalidArgumentException("x"); } catch (LogicException $e) { $row = $row . " LogicException"; } catch (Throwable $e) { }
echo $row, "\n";

$row = "Error:";
try { throw new Error("x"); } catch (Exception $e) { $row = $row . " Exception"; } catch (Throwable $e) { }
try { throw new Error("x"); } catch (Error $e) { $row = $row . " Error"; } catch (Throwable $e) { }
echo $row, "\n";

$row = "TypeError:";
try { throw new TypeError("x"); } catch (Exception $e) { $row = $row . " Exception"; } catch (Throwable $e) { }
try { throw new TypeError("x"); } catch (Error $e) { $row = $row . " Error"; } catch (Throwable $e) { }
try { throw new TypeError("x"); } catch (TypeError $e) { $row = $row . " TypeError"; } catch (Throwable $e) { }
echo $row, "\n";

$row = "ArgumentCountError:";
try { throw new ArgumentCountError("x"); } catch (Error $e) { $row = $row . " Error"; } catch (Throwable $e) { }
try { throw new ArgumentCountError("x"); } catch (TypeError $e) { $row = $row . " TypeError"; } catch (Throwable $e) { }
echo $row, "\n";

echo "class=", get_class(new InvalidArgumentException("x")), "\n";
echo "code=", (new InvalidArgumentException("x", 7))->getCode(), "\n";
echo "message=", (new InvalidArgumentException("boom"))->getMessage(), "\n";
echo "parent=", var_export(get_parent_class(new InvalidArgumentException("x")), true), "\n";
---
Exception: Exception
RuntimeException: Exception RuntimeException
InvalidArgumentException: Exception InvalidArgumentException LogicException
Error: Error
TypeError: Error TypeError
ArgumentCountError: Error TypeError
class=InvalidArgumentException
code=7
message=boom
parent='LogicException'
