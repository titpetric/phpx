<?php
spl_autoload_register(function ($c) {
    echo "load $c\n";
    if ($c === 'IsaLater') { require 'is_a_support.php'; }
});
interface IsaMarker {}
class IsaBase implements IsaMarker { public function count(): int { return 0; } }
var_dump(is_a('IsaLater', 'IsaBase', true));
var_dump(is_a('IsaLater', 'IsaMarker', true));
var_dump(is_a('IsaLater', 'IsaBase'));
var_dump(is_a(new IsaLater, 'IsaBase'));
var_dump(is_a('IsaMissing', 'IsaBase', true));
