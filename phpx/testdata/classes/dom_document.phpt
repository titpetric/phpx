name: new DOMDocument takes php's version and encoding, and a third argument is refused
description: >
  Laravel's mail renderer builds the document with `new DOMDocument('1.0',
  'UTF-8')`, and the bound constructor took no arguments, so the reset mail
  answered an ArgumentCountError. Expected output is php 8.4's.
runner:
  flatstack: false
---
<?php
$a = new DOMDocument();
var_dump($a instanceof DOMDocument);
$b = new DOMDocument('1.0', 'UTF-8');
var_dump($b instanceof DOMDocument);
try {
    new DOMDocument('1.0', 'UTF-8', 'extra');
    echo "no error\n";
} catch (ArgumentCountError $e) {
    echo get_class($e), "\n";
}
---
bool(true)
bool(true)
ArgumentCountError
