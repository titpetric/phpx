<?php

// An assignment as the value of a yield. symfony's AsyncResponse streams a
// chunk it also keeps: `yield $r => $chunk = new ErrorChunk(...)`. The
// generator binds the chunk and yields it, so the reader gets the object and
// the body still reads it after the yield.

class Chunk
{
    public $v = 1;
}

function source()
{
    $r = 1;
    yield $r => $chunk = new Chunk();
    echo "chunk: ", $chunk->v, "\n";
}

foreach (source() as $key => $value) {
    echo "key: ", $key, "\n";
    echo "value: ", get_class($value), "\n";
}
