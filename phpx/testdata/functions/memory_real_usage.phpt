name: memory_get_usage(true) reports the bytes the allocator took from the system
description: >
  php splits memory_get_usage() into the live values it holds and, with
  $real_usage, the memory its allocator has taken from the system. A queue
  worker reads the real form once a job to decide whether to stop, and the two
  figures only relate: the real one is never below the live estimate and never
  below the peak. Expected output is php 8.4's.
---
<?php
$used = memory_get_usage();
$real = memory_get_usage(true);
echo "real>=used: ", $real >= $used ? "yes" : "no", "\n";
echo "real>0: ", $real > 0 ? "yes" : "no", "\n";
echo "peakreal>=real: ", memory_get_peak_usage(true) >= $real ? "yes" : "no", "\n";
echo "peak>=used: ", memory_get_peak_usage() >= $used ? "yes" : "no", "\n";
$blob = str_repeat("x", 4 * 1024 * 1024);
echo "used grew: ", memory_get_usage() > $used ? "yes" : "no", "\n";
echo "real>=used after blob: ", memory_get_usage(true) >= memory_get_usage() ? "yes" : "no", "\n";
unset($blob);
echo "real>=used after unset: ", memory_get_usage(true) >= memory_get_usage() ? "yes" : "no", "\n";
---
real>=used: yes
real>0: yes
peakreal>=real: yes
peak>=used: yes
used grew: yes
real>=used after blob: yes
real>=used after unset: yes
