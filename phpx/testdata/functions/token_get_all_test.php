<?php
$samples = [
    '<?php echo 1 + 2.5e3, 0x1F, 0b101, 0o17, 1_000; ?>tail',
    '<?php $a = "x $b y {$c->d} ${e} $f[0] $g[key] $h->i"; $z = \'raw $x\';',
    "<?php \$s = <<<EOT\n  hi \$name {\$a['k']}\n  EOT;\n\$n = <<<'RAW'\nraw \$x\nRAW;\n",
    '<?php $x = (int) $y; $z = ( string )$w; (object)[]; (unset) $q;',
    '<?php namespace A\B; use \C\D; new namespace\E; $o?->p->class; Foo::class; A\b();',
    '<?php #[Attr(1)] final class X extends Y implements Z { public readonly int $a; use T; }',
    '<?php enum S: string { case A = "a"; } $e = enum_exists("S"); match ($x) { 1, 2 => fn($v) => $v ?? throw new E, default => yield from f() };',
    '<?php $a &= $b & $c; foo(...$args); function &r(array &$x, &...$y) {} $a <=> $b; $c ??= 1; $d **= 2;',
    '<?php /** doc */ /* c */ // line' . "\n" . '# hash' . "\n" . '__CLASS__ . __LINE__ . __DIR__; exit; die; __halt_compiler();',
    "<html><?= \$x ?>\n<?php if (\$a): ?>b<?php endif; ?>\n",
];
foreach ($samples as $code) {
    foreach (token_get_all($code) as $t) {
        echo is_array($t) ? token_name($t[0]) . ' ' . $t[2] . ' ' . json_encode($t[1], JSON_UNESCAPED_SLASHES) : "'" . $t . "'", "\n";
    }
    echo "--\n";
}
var_dump(T_THROW, T_ENUM, T_MATCH, T_NAME_QUALIFIED, T_OPEN_TAG, T_STRING, TOKEN_PARSE, defined('T_PUBLIC_SET'));
