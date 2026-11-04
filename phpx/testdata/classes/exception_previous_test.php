<?php
$inner = new InvalidArgumentException('inner', 3);
$outer = new RuntimeException('outer', 5, $inner);
echo $outer->getMessage(), $outer->getCode(), "\n";
echo get_class($outer->getPrevious()), ':', $outer->getPrevious()->getMessage(), "\n";
var_dump((new Exception('x'))->getPrevious());
try {
    throw new LogicException('thrown', 0, new Exception('cause'));
} catch (LogicException $e) {
    echo $e->getMessage(), ' <- ', $e->getPrevious()->getMessage(), "\n";
}
class AppError extends Exception {
    public function __construct(string $m, private string $extra = 'e', ?Throwable $prev = null) {
        parent::__construct($m, 9, $prev);
    }
    public function extra() { return $this->extra; }
}
$a = new AppError('app', 'more', new Exception('root'));
echo $a->getMessage(), $a->getCode(), $a->extra(), $a->getPrevious()->getMessage(), "\n";
echo $outer instanceof Throwable ? 'T' : 'F', "\n";
