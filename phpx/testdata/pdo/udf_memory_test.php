<?php
// A function registered after the connection opened, on an in-memory
// database. php's PDO holds one connection, so the table created before the
// registration is the table the function query reads.
$pdo = new PDO("sqlite::memory:");
$pdo->exec("create table item (name text)");
$pdo->exec("insert into item (name) values ('alpha'), ('beta'), ('gamma')");
var_dump($pdo->sqliteCreateFunction("REGEXP", function (string $pattern, string $value) {
    return preg_match("/" . str_replace("/", "\\/", $pattern) . "/", $value) === 1;
}, 2));
var_dump($pdo->query("select name from item where name REGEXP '^a' order by name")->fetchAll(PDO::FETCH_COLUMN));
var_dump($pdo->query("select name from item where name REGEXP 'a$' order by name")->fetchAll(PDO::FETCH_COLUMN));
$pdo->exec("insert into item (name) values ('alphabet')");
var_dump($pdo->query("select count(*) from item")->fetchColumn());
