name: seeded mt_rand follows php
description: >
  mt_srand(n) gives the sequence php gives: the Mersenne Twister with php's range rejection, the legacy MT_RAND_PHP scaling, and shuffle, str_shuffle and array_rand drawing from it. Expected output is php 8.4's.
runner:
  flatstack: false
---
<?php
mt_srand(3); echo mt_rand(), ' ', mt_rand(1, 6), ' ', mt_rand(1, 100), ' ', mt_rand(0, 4294967295), ' ', mt_rand(-5, 5), "\n";
srand(7); echo rand(), ' ', rand(1, 10), ' ', mt_rand(1, 10), "\n";
mt_srand(11, MT_RAND_PHP); echo mt_rand(), ' ', mt_rand(1, 6), ' ', mt_rand(1, 1000), "\n";
mt_srand(5); $a = range(1, 10); shuffle($a); echo implode(',', $a), "\n";
mt_srand(5); echo str_shuffle('abcdefghij'), "\n";
mt_srand(9); $r = ['a' => 1, 'b' => 2, 'c' => 3, 'd' => 4, 'e' => 5];
echo array_rand($r), ' ', implode(',', array_rand($r, 2)), ' ', implode(',', array_rand($r, 4)), "\n";
mt_srand(2); echo mt_rand(1, 1 << 40), ' ', mt_getrandmax(), "\n";
mt_srand(4); echo random_int(1, 1) , ' ', lcg_value() < 1 ? 'ok' : 'bad', "\n";
---
1182829493 3 38 3607553667 1
163870807 3 2
1750656698 1 20
10,7,6,5,1,3,9,8,4,2
jgfeacihdb
a d,e a,b,d,e
66297093289 2147483647
1 ok
