name: include of a file that is not there warns and answers false
description: >
  php's include and include_once return false when the file cannot be opened,
  where require is fatal, and a suppressed include is how carbon's translator
  walks its locale directories. Expected output is php 8.4's.
runner:
  flatstack: false
---
<?php
var_dump(@include 'no_such_file.php');
var_dump(file_exists('no_such_file.php'));
var_dump(@include_once 'no_such_file.php');
---
bool(false)
bool(false)
bool(false)
