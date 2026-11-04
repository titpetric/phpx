<?php
echo json_encode(unpack('C*', "ABC")), "\n";
echo json_encode(unpack('C2a/Cb', "\x01\x02\x03")), "\n";
echo json_encode(unpack('nlen/Nbig/vlittle', "\x01\x02\x00\x00\x01\x00\x03\x00")), "\n";
echo json_encode(unpack('a3x/A5y', "abcde   ")), "\n";
echo json_encode(unpack('H*', "\xde\xad")), "\n";
echo json_encode(unpack('c', "\xff")), "\n";
echo json_encode(unpack('Cfirst/a*rest', "Zhello")), "\n";
echo json_encode(unpack('C', "A", 0)), "\n";
