name: a stream is a resource with an id, and a closed one says so
description: >
  Resource id #n in a string, the id in an integer context, gettype and get_debug_type before and after fclose, is_resource false once closed. Expected output is php 8.4's.
runner:
  flatstack: false
---
<?php
$h = fopen('php://memory', 'w+'); $g = fopen('php://memory', 'r');
echo (string) $h, '|', (int) $h, '|', "$g", '|', get_resource_id($h), '|', intval($g), "\n";
var_dump(gettype($h), get_debug_type($h), is_resource($h), get_resource_type($h));
fclose($h); var_dump(gettype($h), is_resource($h), get_debug_type($h), get_resource_type($h), (string) $h);
---
Resource id #5|5|Resource id #6|5|6
string(8) "resource"
string(17) "resource (stream)"
bool(true)
string(6) "stream"
string(17) "resource (closed)"
bool(false)
string(17) "resource (closed)"
string(7) "Unknown"
string(14) "Resource id #5"
