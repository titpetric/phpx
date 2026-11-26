<?php
// A built-in throwable is a class php declares from the start. Its members
// are reflected by name, a subclass hides the parent's private ones, and the
// defaults are the ones a renderer reads.
foreach (['Exception', 'Error'] as $class) {
    $r = new ReflectionClass($class);
    echo $class, ' properties: ';
    foreach ($r->getProperties() as $p) {
        echo $p->getName(), ':', $p->isPrivate() ? 'private' : ($p->isProtected() ? 'protected' : 'public'), ' ';
    }
    echo "\n";
}
$r = new ReflectionClass('RuntimeException');
echo 'RuntimeException parent=', $r->getParentClass()->getName(), "\n";
echo 'RuntimeException properties: ';
foreach ($r->getProperties() as $p) { echo $p->getName(), ' '; }
echo "\n";
echo 'RuntimeException has trace=', $r->hasProperty('trace') ? 'yes' : 'no', "\n";
$trace = new ReflectionProperty('Exception', 'trace');
echo 'Exception::$trace private=', $trace->isPrivate() ? 'yes' : 'no', ' default=', var_export($trace->getDefaultValue(), true), "\n";
$previous = new ReflectionProperty('Exception', 'previous');
echo 'Exception::$previous private=', $previous->isPrivate() ? 'yes' : 'no', ' default=', var_export($previous->getDefaultValue(), true), "\n";
$e = new RuntimeException('boom', 7, new Exception('cause'));
echo 'previous=', get_class($e->getPrevious()), ' trace=', count($e->getTrace()), "\n";
