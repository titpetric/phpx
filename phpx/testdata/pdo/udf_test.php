<?php
$path = tempnam(sys_get_temp_dir(), "phpx_pdo_");
$pdo = new PDO("sqlite:" . $path);
$pdo->exec("create table item (name text)");
$pdo->exec("insert into item (name) values ('alpha'), ('beta'), ('gamma')");
var_dump($pdo->sqliteCreateFunction("REGEXP", function (string $pattern, string $value) {
    return preg_match("/" . str_replace("/", "\\/", $pattern) . "/", $value) === 1;
}, 2));
var_dump($pdo->query("select name from item where name REGEXP '^a' order by name")->fetchAll(PDO::FETCH_COLUMN));
var_dump($pdo->query("select name from item where name REGEXP 'a$' order by name")->fetchAll(PDO::FETCH_COLUMN));
unlink($path);
