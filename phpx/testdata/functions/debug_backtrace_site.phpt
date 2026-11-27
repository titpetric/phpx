name: debug_backtrace names each frame's call site
description: >
  php reports the file and line a frame was called from for every entry, and
  Pest's Support\Backtrace reads the file to find the test file it loaded. A
  trace with no file raised `Backtrace not found.` The file prints whether the
  entry names a file and the line the call is on. Expected output is php
  8.4's.
runner:
  flatstack: false
---
<?php
function c() {
    foreach (debug_backtrace(DEBUG_BACKTRACE_IGNORE_ARGS) as $frame) {
        $file = $frame['file'] ?? '';
        echo $frame['function'], ' ', $file !== '' ? 'file' : 'no-file', ':', $frame['line'], "\n";
    }
}
function b() { c(); }
function a() { b(); }
a();
---
c file:8
b file:9
a file:10
