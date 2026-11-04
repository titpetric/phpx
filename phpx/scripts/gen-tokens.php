<?php
// Writes stdlib/phpx/tokens_gen.go from the php that runs it:
//   php phpx/scripts/gen-tokens.php > stdlib/phpx/tokens_gen.go && gofmt -w stdlib/phpx/tokens_gen.go
$c = get_defined_constants(true)['tokenizer'];
$byId = [];
foreach ($c as $k => $v) {
    if (strpos($k, 'T_') === 0) {
        $byId[$v] = true;
    }
}
ksort($byId);
echo "//go:build phpx\n\npackage phpx\n\n";
echo "// Code generated from php " . PHP_VERSION . " get_defined_constants(true)[\"tokenizer\"]; DO NOT EDIT.\n\n";
echo "// phpxTokenConstants are the tokenizer extension's constants and their values.\nvar phpxTokenConstants = map[string]int64{\n";
foreach ($c as $k => $v) {
    echo "\t\"$k\": $v,\n";
}
echo "}\n\n// phpxTokenNames is what token_name() answers for an id.\nvar phpxTokenNames = map[int64]string{\n";
foreach ($byId as $id => $_) {
    echo "\t$id: \"" . token_name($id) . "\",\n";
}
echo "}\n";
