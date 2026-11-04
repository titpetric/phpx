<?php
echo preg_replace('/(?<=[[:lower:]])(?=[[:upper:]])/u', ' ', 'SkipOnly'), "|\n";
echo preg_replace('/(?<=[[:lower:]])(?=[[:upper:]])/', ' ', 'SkipOnly'), "|\n";
echo preg_replace('/(?<=[a-z])(?=[A-Z])/u', ' ', 'SkipOnly'), "|\n";
echo preg_replace('/[[:upper:]]/', '_', 'SkipOnly'), "|\n";
echo preg_match('/[[:alpha:]]+/', 'abc1', $m), $m[0], "\n";
