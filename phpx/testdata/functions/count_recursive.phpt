name: count with COUNT_RECURSIVE counts every element at every depth
description: >
  laravel's MessageBag reads count($messages, COUNT_RECURSIVE) - count($messages)
  as the number of messages across the groups, so a validator that answered 0
  passed every rule. Expected output is php 8.4's.
runner:
  flatstack: false
---
<?php
$a = ['a' => ['b', 'c'], 'd' => 'e'];
var_dump(count($a), count($a, COUNT_RECURSIVE));
$bag = ['name' => ['required'], 'email' => ['required', 'email'], 'password' => []];
var_dump(count($bag), count($bag, COUNT_RECURSIVE), count($bag, COUNT_RECURSIVE) - count($bag));
var_dump(count([], COUNT_RECURSIVE));
---
int(2)
int(4)
int(3)
int(6)
int(3)
int(0)
