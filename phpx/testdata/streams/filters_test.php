<?php
// The filter registry composer's autoload files reach: a name is registered
// once and stream_get_filters() reports it beside php's built-ins.
var_dump(stream_filter_register("chunk", "Chunk"));
var_dump(stream_filter_register("chunk", "Other"));
try {
    stream_filter_register("", "Chunk");
} catch (ValueError $e) {
    echo get_class($e), ": ", $e->getMessage(), "\n";
}
var_dump(in_array("chunk", stream_get_filters(), true));
var_dump(in_array("dechunk", stream_get_filters(), true));
