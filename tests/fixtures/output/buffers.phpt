name: buffers
description: >
  An output buffer captures echo, print and inline output. ob_get_contents
  answers a copy, so a level closed and opened again leaves what the script
  holds alone; a nested level flushes into the one below it; print writes into
  the active level; and a buffer the program leaves open is flushed at
  shutdown.
---
<?php
ob_start();
echo "a";
$first = ob_get_contents();
print "b";
$second = ob_get_contents();
ob_end_clean();
ob_start();
echo "c";
ob_end_flush();
echo "[", $first, "][", $second, "]\n";

ob_start();
echo "outer-";
ob_start();
echo "inner";
ob_end_flush();
echo "after";
ob_end_flush();
echo "\n";

ob_start();
echo "left open";
---
c[a][ab]
outer-innerafter
left open
