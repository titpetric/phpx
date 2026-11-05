<?php
// extension-tally.php DIR...  Prints, per directory, the calls to functions and
// uses of classes that belong to a PHP extension other than Core, standard,
// SPL, date, pcre, hash, json, ctype, random and Reflection, counted by token.
// Run it with php: the extension of a name is php's own answer.
// Skips vendor/phpunit and its siblings, the dev dependencies a framework tree carries, so a
// count reads the framework and not its test runner. PHPX_TALLY_SKIP=a,b names others; empty skips none.
$env = getenv('PHPX_TALLY_SKIP');
$skip = array_filter(explode(',', $env === false ? 'phpunit,sebastian,phar-io,theseer' : $env));
$base = ['Core', 'standard', 'SPL', 'date', 'pcre', 'hash', 'random', 'Reflection', 'ctype', 'json'];
foreach (array_slice($argv, 1) as $dir) {
    $fn = [];
    $cls = [];
    $files = 0;
    $it = new RecursiveIteratorIterator(new RecursiveDirectoryIterator($dir, FilesystemIterator::SKIP_DOTS));
    foreach ($it as $f) {
        if (substr($f->getFilename(), -4) !== '.php' || strpos($f->getPathname(), '/tests/') !== false) {
            continue;
        }
        foreach ($skip as $s) {
            if (strpos($f->getPathname(), "/vendor/$s/") !== false) continue 2;
        }
        $files++;
        $t = token_get_all(file_get_contents($f->getPathname()));
        $n = count($t);
        for ($i = 0; $i < $n; $i++) {
            if (!is_array($t[$i])) continue;
            if ($t[$i][0] === T_STRING) {
                $j = $i + 1;
                while ($j < $n && is_array($t[$j]) && $t[$j][0] === T_WHITESPACE) $j++;
                $p = $i - 1;
                while ($p >= 0 && is_array($t[$p]) && $t[$p][0] === T_WHITESPACE) $p--;
                $prev = $t[$p] ?? null;
                $isMember = is_array($prev) && in_array($prev[0], [T_OBJECT_OPERATOR, T_DOUBLE_COLON, T_NULLSAFE_OBJECT_OPERATOR, T_FUNCTION, T_NEW, T_CONST], true);
                if ($t[$j] === '(' && !$isMember) {
                    $name = $t[$i][1];
                    if (function_exists($name)) {
                        $e = (new ReflectionFunction($name))->getExtensionName();
                        if ($e !== false && !in_array($e, $base, true)) $fn[$e][$name] = ($fn[$e][$name] ?? 0) + 1;
                    }
                }
            }
            if ($t[$i][0] === T_NAME_FULLY_QUALIFIED || $t[$i][0] === T_STRING) {
                $name = ltrim($t[$i][1], '\\');
                if (strpos($name, '\\') === false && (class_exists($name, false) || interface_exists($name, false))) {
                    $e = (new ReflectionClass($name))->getExtensionName();
                    if ($e !== false && !in_array($e, $base, true)) $cls[$e][$name] = ($cls[$e][$name] ?? 0) + 1;
                }
            }
        }
    }
    echo "## $dir ($files files)\n";
    $all = array_unique(array_merge(array_keys($fn), array_keys($cls)));
    $rows = [];
    foreach ($all as $e) $rows[$e] = [array_sum($fn[$e] ?? []), array_sum($cls[$e] ?? [])];
    uasort($rows, fn($a, $b) => ($b[0] + $b[1]) <=> ($a[0] + $a[1]));
    foreach ($rows as $e => [$c, $k]) {
        $top = $fn[$e] ?? []; arsort($top);
        $topc = $cls[$e] ?? []; arsort($topc);
        echo "$e\t$c calls\t$k class uses\t" . implode(',', array_slice(array_keys($top + $topc), 0, 4)) . "\n";
    }
}
