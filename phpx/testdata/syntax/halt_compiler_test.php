<?php
$data = file_get_contents(__FILE__, false, null, __COMPILER_HALT_OFFSET__);
echo trim($data), "|", strlen($data), "\n";
function again() { return __COMPILER_HALT_OFFSET__ > 0 ? 'offset' : 'none'; }
echo again(), "\n";
__halt_compiler();
@@ foo
{{ foo }} class { not php
