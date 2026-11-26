name: strncmp answers zero when the compared prefix agrees
description: >
  php compares at most $length bytes and answers 0 whenever those bytes agree,
  whatever follows them; only a string shorter than $length ends the scan
  early and lets the length break the tie. The Footnote renderer compares a
  destination against its prefix this way. Expected output is php 8.4's.
---
<?php
var_dump(strncmp('abcd', 'ab', 2));
var_dump(strncmp('ab', 'abcd', 2));
var_dump(strncmp('ab', 'abcd', 3) < 0);
var_dump(strncmp('xab', 'xac', 2) === 0);
var_dump(strncmp('abc', 'abd', 2));
var_dump(strncmp('abc', 'ABC', 3));
var_dump(strncasecmp('abc', 'ABC', 3));
var_dump(strncmp('abc', 'abc', 0));
---
int(0)
int(0)
bool(true)
bool(true)
int(0)
int(32)
int(0)
int(0)
