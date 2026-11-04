<?php
class Agg implements IteratorAggregate { function getIterator(): Iterator { return new ArrayIterator([1]); } }
class Plain {}
$g = (function () { yield 1; })();
var_dump(is_iterable([1]), is_iterable(new ArrayIterator([])), is_iterable($g), is_iterable(new Agg), is_iterable(new Plain), is_iterable(new ArrayObject([])), is_iterable(1), is_iterable('a'), is_iterable(null), is_iterable(new stdClass));
