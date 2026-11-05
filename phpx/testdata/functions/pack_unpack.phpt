name: pack builds binary strings and unpack reads them back
description: >
  The format codes libraries use: big and little endian integers, strings
  padded with NUL or spaces, hex, doubles, and a repeat count. Expected output
  is php 8.4's.
---
<?php
echo bin2hex(pack('n', 0x1234)), ' ', bin2hex(pack('v', 0x1234)), ' ', bin2hex(pack('N', 1)), ' ', bin2hex(pack('V', 1)), "\n";
echo bin2hex(pack('n*', 1, 2, 3)), ' ', bin2hex(pack('J', 258)), ' ', bin2hex(pack('P', 258)), "\n";
echo bin2hex(pack('a4', 'ab')), ' ', bin2hex(pack('A4', 'ab')), ' ', bin2hex(pack('Z*', 'ab')), ' ', bin2hex(pack('H*', 'deadbeef')), "\n";
echo bin2hex(pack('CxC', 1, 2)), ' ', bin2hex(pack('E', 1.0)), ' ', bin2hex(pack('e', 1.0)), ' ', bin2hex(pack('c', -1)), "\n";
print_r(unpack('n2a/Cb', "\x00\x01\x00\x02\x07"));
print_r(unpack('l*', pack('l*', -5, 6)));
print_r(unpack('H*hex', "\xde\xad"));
print_r(unpack('E', pack('E', 2.5)));
---
1234 3412 00000001 01000000
000100020003 0000000000000102 0201000000000000
61620000 61622020 616200 deadbeef
010002 3ff0000000000000 000000000000f03f ff
Array
(
    [a1] => 1
    [a2] => 2
    [b] => 7
)
Array
(
    [1] => -5
    [2] => 6
)
Array
(
    [hex] => dead
)
Array
(
    [1] => 2.5
)
