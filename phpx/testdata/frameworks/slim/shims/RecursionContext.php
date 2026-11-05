<?php
// SebastianBergmann\RecursionContext\Context without the markers it writes into
// the arrays it is given. The original appends two items to each array it
// tracks and removes them in __destruct; phpx runs no destructors, so the items
// stay and every array phpunit exports or compares carries them. Arrays are not
// tracked here, so a self-referencing array is not detected: no suite in this
// directory builds one. Objects are tracked as before.
namespace SebastianBergmann\RecursionContext;

use SplObjectStorage;
use function spl_object_id;

final class Context
{
    private SplObjectStorage $objects;

    public function __construct()
    {
        $this->objects = new SplObjectStorage;
    }

    public function add(array|object &$value): int
    {
        if (is_array($value)) {
            return 0;
        }
        if (!$this->objects->offsetExists($value)) {
            $this->objects->offsetSet($value);
        }
        return spl_object_id($value);
    }

    public function contains(array|object &$value): false|int
    {
        if (is_array($value)) {
            return false;
        }
        return $this->objects->offsetExists($value) ? spl_object_id($value) : false;
    }
}
