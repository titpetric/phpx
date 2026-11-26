name: debug_backtrace names the running function, its class and the caller
description: >
  Laravel guesses a relation's foreign key from the caller's name
  (`Str::snake($caller['function']).'_id'`), so a frame that answered an empty
  function turned `url()` into `_id` and every factory insert failed. Expected
  output is php 8.4's. A closure is spelled `{closure}`, where php adds the
  defining file and line; the fixture keeps to functions and methods.
runner:
  flatstack: false
---
<?php
class Bag {
    public function email() { return $this->relation(); }
    public function relation() { return $this->frames(); }
    public function frames() {
        $all = debug_backtrace();
        echo count($all), "\n";
        foreach ($all as $frame) {
            echo $frame['function'];
            if (isset($frame['class'])) {
                echo ' of ', $frame['class'];
            }
            if (isset($frame['type'])) {
                echo ' ', $frame['type'];
            }
            echo "\n";
        }
        $limited = debug_backtrace(DEBUG_BACKTRACE_IGNORE_ARGS, 2);
        echo count($limited), ' ', $limited[1]['function'], "\n";
    }
}
function outer() { (new Bag)->email(); }
outer();
var_dump(debug_backtrace());
---
4
frames of Bag ->
relation of Bag ->
email of Bag ->
outer
2 relation
array(0) {
}
