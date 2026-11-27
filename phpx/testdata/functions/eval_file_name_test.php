<?php

// A class declared by eval() belongs to php's "<caller>(<line>) : eval()'d
// code" file: __FILE__ inside the eval, and the class reflection, agree.
eval('class EvalName {}');
$file = (new ReflectionClass('EvalName'))->getFileName();
echo basename($file), "\n";

eval('$class = new ReflectionClass("EvalName");');
echo substr($file, -13), "\n";
var_dump(str_starts_with($file, __FILE__ . '('));
