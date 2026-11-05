name: the built-in throwables form php's tree
description: >
  A catch of RuntimeException takes an UnexpectedValueException, instanceof, is_subclass_of and get_parent_class agree with it, for the SPL exceptions and the Error family. Expected output is php 8.4's.
runner:
  flatstack: false
---
<?php
foreach (['UnexpectedValueException' => 'RuntimeException', 'OutOfBoundsException' => 'RuntimeException', 'RangeException' => 'RuntimeException', 'OverflowException' => 'RuntimeException', 'UnderflowException' => 'RuntimeException', 'OutOfRangeException' => 'LogicException', 'LengthException' => 'LogicException', 'DomainException' => 'LogicException', 'InvalidArgumentException' => 'LogicException', 'BadMethodCallException' => 'BadFunctionCallException', 'BadFunctionCallException' => 'LogicException', 'JsonException' => 'Exception', 'ErrorException' => 'Exception', 'TypeError' => 'Error', 'ValueError' => 'Error', 'ArithmeticError' => 'Error', 'DivisionByZeroError' => 'ArithmeticError', 'ArgumentCountError' => 'TypeError', 'AssertionError' => 'Error', 'UnhandledMatchError' => 'Error'] as $child => $parent) {
  try { try { throw new $child('x'); } catch (RuntimeException $e) { echo "$child caught as Runtime\n"; } catch (LogicException $e) { echo "$child caught as Logic\n"; } catch (Error $e) { echo "$child caught as Error\n"; } catch (Exception $e) { echo "$child caught as Exception\n"; } }
  catch (Throwable $e) { echo "$child other\n"; }
  var_dump(is_subclass_of($child, $parent), (new $child('x')) instanceof $parent, get_parent_class($child));
}
---
UnexpectedValueException caught as Runtime
bool(true)
bool(true)
string(16) "RuntimeException"
OutOfBoundsException caught as Runtime
bool(true)
bool(true)
string(16) "RuntimeException"
RangeException caught as Runtime
bool(true)
bool(true)
string(16) "RuntimeException"
OverflowException caught as Runtime
bool(true)
bool(true)
string(16) "RuntimeException"
UnderflowException caught as Runtime
bool(true)
bool(true)
string(16) "RuntimeException"
OutOfRangeException caught as Logic
bool(true)
bool(true)
string(14) "LogicException"
LengthException caught as Logic
bool(true)
bool(true)
string(14) "LogicException"
DomainException caught as Logic
bool(true)
bool(true)
string(14) "LogicException"
InvalidArgumentException caught as Logic
bool(true)
bool(true)
string(14) "LogicException"
BadMethodCallException caught as Logic
bool(true)
bool(true)
string(24) "BadFunctionCallException"
BadFunctionCallException caught as Logic
bool(true)
bool(true)
string(14) "LogicException"
JsonException caught as Exception
bool(true)
bool(true)
string(9) "Exception"
ErrorException caught as Exception
bool(true)
bool(true)
string(9) "Exception"
TypeError caught as Error
bool(true)
bool(true)
string(5) "Error"
ValueError caught as Error
bool(true)
bool(true)
string(5) "Error"
ArithmeticError caught as Error
bool(true)
bool(true)
string(5) "Error"
DivisionByZeroError caught as Error
bool(true)
bool(true)
string(15) "ArithmeticError"
ArgumentCountError caught as Error
bool(true)
bool(true)
string(9) "TypeError"
AssertionError caught as Error
bool(true)
bool(true)
string(5) "Error"
UnhandledMatchError caught as Error
bool(true)
bool(true)
string(5) "Error"
