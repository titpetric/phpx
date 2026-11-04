<?php
spl_autoload_register(function ($c) {
    if ($c === 'LateBase') {
        eval('abstract class LateBase { public function hi() { return "hi from " . static::class; } }');
    }
});
class Child extends LateBase {}
echo (new Child)->hi(), "\n";
