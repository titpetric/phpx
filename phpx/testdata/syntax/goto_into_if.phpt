name: goto lands on a label inside an if
description: >
  php allows a label in a nested if block, reached from a loop further in; a
  generated LALR parser jumps back to its `recovery:` label that way. Expected
  output is php 8.4's.
---
<?php
function f() {
  $n = 0;
  do {
    if ($n > 100) {
      echo "never\n";
    } else {
      if ($n >= 0) {
        $n++;
        echo "enter $n\n";
        back:
        echo "after label $n\n";
      }
    }
    do {
      if ($n < 3) { $n++; goto back; }
    } while (false);
    echo "done $n\n";
    return $n;
  } while (true);
}
echo f(), "\n";
---
enter 1
after label 1
after label 2
after label 3
done 3
3
